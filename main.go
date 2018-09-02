/* Copyright 2017 Jeffry Hesse

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License. */

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/DarthHater/bored-board-service/constants"
	"github.com/DarthHater/bored-board-service/model"
	"github.com/garyburd/redigo/redis"
	"golang.org/x/crypto/bcrypt"

	"crypto/rsa"

	"github.com/DarthHater/bored-board-service/database"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	ginlogrus "github.com/toorop/gin-logrus"
)

const (
	redisURL = "redis_db:6379"
)

var (
	db          database.IDatabase
	gPubSubConn *redis.PubSubConn
	gRedisConn  = func() (redis.Conn, error) {
		redisURL := os.Getenv("REDIS_URL")
		if redisURL != "" {
			return redis.DialURL(redisURL)
		}

		return redis.Dial("tcp", "redis_db:6379")
	}
	signKey   *rsa.PrivateKey
	verifyKey *rsa.PublicKey
)

type clientManager struct {
	clients    map[*client]bool
	broadcast  chan []byte
	register   chan *client
	unregister chan *client
}

type client struct {
	id     string
	socket *websocket.Conn
	send   chan []byte
}

var manager = clientManager{
	broadcast:  make(chan []byte),
	register:   make(chan *client),
	unregister: make(chan *client),
	clients:    make(map[*client]bool),
}

func (manager *clientManager) start() {
	for {
		select {
		case conn := <-manager.register:
			manager.clients[conn] = true
			log.Print("New connection registered")
		case conn := <-manager.unregister:
			if _, ok := manager.clients[conn]; ok {
				close(conn.send)
				delete(manager.clients, conn)
			}
		case message := <-manager.broadcast:
			for conn := range manager.clients {
				select {
				case conn.send <- message:
				default:
					close(conn.send)
					delete(manager.clients, conn)
				}
			}
		}
	}
}

func (manager *clientManager) send(message []byte) {
	for conn := range manager.clients {
		conn.send <- message
	}
}

func (c *client) read() {
	defer func() {
		manager.unregister <- c
		c.socket.Close()
	}()

	for {
		switch v := gPubSubConn.Receive().(type) {
		case redis.Message:
			manager.broadcast <- v.Data
		case error:
			manager.unregister <- c
			c.socket.Close()
			break
		}
	}
}

func (c *client) write() {
	defer func() {
		c.socket.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.socket.WriteMessage(websocket.TextMessage, message)
		}
	}
}

var webSocketUpgrade = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func getToken(c *gin.Context) (*jwt.Token, error) {
	tokenString := c.GetHeader("Authorization")

	if tokenString == "" {
		origin := c.GetHeader("Origin")
		c.JSON(http.StatusForbidden, gin.H{"err": "token is required"})
		log.Error(fmt.Sprintf("Attempt to access resources without token from origin: %s", origin))
		return nil, errors.New("Error")
	}

	tokenString = tokenString[len("Bearer "):]

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return verifyKey, nil
	})

	return token, err
}

func userIsLoggedIn() gin.HandlerFunc {
	return func(c *gin.Context) {

		token, err := getToken(c)

		if err != nil {
			c.Abort()
			return
		}

		if token.Valid {
			// save token in context for use in other middleware
			c.Set("token", token)
		} else {
			origin := c.GetHeader("Origin")
			if ve, ok := err.(*jwt.ValidationError); ok {
				if ve.Errors&jwt.ValidationErrorMalformed != 0 {
					c.JSON(http.StatusForbidden, gin.H{"err": "token is malformed"})
					log.Error(fmt.Sprintf("Attempt to access resources with malformed token from origin: %s", origin))
				} else if ve.Errors&(jwt.ValidationErrorExpired) != 0 {
					// Token is expired
					c.JSON(http.StatusForbidden, gin.H{"err": "token is expired"})
				} else {
					c.JSON(http.StatusForbidden, gin.H{"err": "error reading token"})
					log.Error(fmt.Sprintf("Error accessing resources with token from origin %s", origin))
				}
			} else {
				c.JSON(http.StatusForbidden, gin.H{"err": "error reading token"})
				log.Error(fmt.Sprintf("Error accessing resources with token from origin %s", origin))
			}
			c.Abort()
		}
	}
}

func userIsInRole(d database.IDatabase, roles []constants.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		var token interface{}
		var ok bool

		if token, ok = c.Get("token"); !ok {
			c.JSON(http.StatusForbidden, gin.H{"err": "Error accessing token"})
			c.Abort()
			return
		}

		var userRole constants.Role
		if claims, ok := token.(*jwt.Token).Claims.(jwt.MapClaims); ok {
			userRole = constants.Role(claims["role"].(float64))
		} else {
			c.Abort()
			return
		}

		for _, role := range roles {
			if userRole == role {
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"err": "User doesn't have access"})
		c.Abort()
	}
}

func init() {
	privKeyPath := os.Getenv("PRIVATE_KEY_PATH")
	pubKeyPath := os.Getenv("PUBLIC_KEY_PATH")

	signBytes, err := ioutil.ReadFile(privKeyPath)
	if err != nil {
		log.Fatal(err)
	}

	signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		log.Fatal(err)
	}

	verifyBytes, err := ioutil.ReadFile(pubKeyPath)
	if err != nil {
		log.Fatal(err)
	}

	verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	d := database.Database{}
	db = &d
	r := setupRouter(db)
	r.Use(gin.Logger())

	gRedisConn, err := gRedisConn()
	if err != nil {
		panic(err)
	}
	defer gRedisConn.Close()

	gPubSubConn = &redis.PubSubConn{Conn: gRedisConn}
	gPubSubConn.Subscribe("posts")
	defer gPubSubConn.Close()

	go manager.start()

	port := os.Getenv("PORT")
	if port == "" {
		r.Run(":8000")
	} else {
		r.Run(":" + port)
	}
}

func setupRouter(d database.IDatabase) *gin.Engine {
	log := log.New()
	r := gin.New()
	r.Use(cors.New(cors.Config{
		AllowOrigins: allowedCorsOrigins(),
		AllowMethods: []string{"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"HEAD"},
		AllowHeaders: []string{"Origin",
			"Content-Length",
			"Content-Type",
			"Accept-Encoding",
			"Authorization",
			"Cache-Control"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.Use(ginlogrus.Logger(log), gin.Recovery())

	err := d.InitDb("development", "./.environment")
	if err != nil {
		log.Fatal(err)
	}

	r.POST("/login", func(c *gin.Context) {
		checkCredentials(c, d)
	})

	r.POST("/register", func(c *gin.Context) {
		createUser(c, d)
	})

	r.GET("/ws", func(c *gin.Context) {
		webSocketHandler(c.Writer, c.Request)
	})

	auth := r.Group("/")

	auth.Use(userIsLoggedIn())
	{
		auth.GET("/thread/:threadid", func(c *gin.Context) {
			threadID := c.Param("threadid")
			getThread(c, d, threadID)
		})

		auth.GET("/post/:postid", func(c *gin.Context) {
			postID := c.Param("postid")
			getPost(c, d, postID)
		})

		auth.GET("/posts/:threadid", func(c *gin.Context) {
			threadID := c.Param("threadid")
			getPosts(c, d, threadID)
		})

		auth.GET("/threads", func(c *gin.Context) {
			userID := c.Query("userid")
			getThreads(c, d, 20, userID)
		})

		auth.GET("/message/:messageid", func(c *gin.Context) {
			messageID := c.Param("messageid")
			getMessage(c, d, messageID)
		})

		auth.GET("/messages/:userid", func(c *gin.Context) {
			userID := c.Param("userid")
			getMessages(c, d, 20, userID)
		})

		auth.GET("/messageposts/:messageid", func(c *gin.Context) {
			messageID := c.Param("messageid")
			getMessagePosts(c, d, messageID)
		})

		auth.POST("/thread", func(c *gin.Context) {
			postThread(c, d)
		})

		auth.POST("/post", func(c *gin.Context) {
			postPost(c, d)
		})

		auth.POST("/newmessage", func(c *gin.Context) {
			postMessage(c, d)
		})

		auth.POST("/message", func(c *gin.Context) {
			postMessagePost(c, d)
		})

		auth.PATCH("/posts/:postid", func(c *gin.Context) {
			postID := c.Param("postid")
			editPost(c, d, postID)
		})

		auth.GET("/user/:userid", func(c *gin.Context) {
			userID := c.Param("userid")
			getUserInfo(c, d, userID)
		})

		auth.GET("/users", func(c *gin.Context) {
			search := c.Query("search")
			getUsers(c, d, search)
		})

		auth.Use(userIsInRole(d, []constants.Role{constants.Admin, constants.Mod}))
		{
			auth.DELETE("/thread/:threadid", func(c *gin.Context) {
				threadID := c.Param("threadid")
				deleteThread(c, d, threadID)
			})
		}
	}

	return r
}

func allowedCorsOrigins() []string {
	var environment = os.Getenv("ENVIRONMENT")
	if environment == "development" {
		return []string{"http://localhost:8090",
			"http://127.0.0.1:8090",
			"http://0.0.0.0:8090",
			"https://vivalavinyl-webapp.herokuapp.com"}
	}
	return []string{
		"https://vivalavinyl-webapp.herokuapp.com"}
}

// Websocket Handler
func webSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := webSocketUpgrade.Upgrade(w, r, nil)
	if err != nil {
		log.Error(err)
		return
	}

	client := &client{id: uuid.NewV4().String(), socket: conn, send: make(chan []byte)}

	manager.register <- client

	go client.read()
	go client.write()
}

// Handlers
func getThread(c *gin.Context, d database.IDatabase, threadID string) {
	thread, err := d.GetThread(threadID)
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusBadRequest, "Uh oh")
	}
	c.JSON(http.StatusOK, thread)
}

func getPost(c *gin.Context, d database.IDatabase, postID string) {
	post, err := d.GetPost(postID)
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusBadRequest, "Uh oh")
	}
	c.JSON(http.StatusOK, post)
}

func getThreads(c *gin.Context, d database.IDatabase, num int, userID string) {
	threads, err := d.GetThreads(num, userID)
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusBadRequest, "Uh oh")
	} else {
		c.JSON(http.StatusOK, threads)
	}
}

func getMessages(c *gin.Context, d database.IDatabase, num int, userID string) {
	messages, err := d.GetMessages(num, userID)
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusBadRequest, "Uh oh")
	} else {
		c.JSON(http.StatusOK, messages)
	}
}

func getUserInfo(c *gin.Context, d database.IDatabase, userID string) {
	userInfo, err := d.GetUserInfo(userID)
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusBadRequest, "Uh oh")
	} else {
		c.JSON(http.StatusOK, userInfo)
	}
}


func getUsers(c *gin.Context, d database.IDatabase, search string) {
	userInfo, err := d.GetUsers(search)
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusBadRequest, "Uh oh")
	} else {
		c.JSON(http.StatusOK, userInfo)
	}
}

func getMessagePosts(c *gin.Context, d database.IDatabase, messageID string) {
	messages, err := d.GetMessagePosts(messageID)
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusBadRequest, "Uh oh")
	} else {
		c.JSON(http.StatusOK, messages)
	}
}

func getPosts(c *gin.Context, d database.IDatabase, threadID string) {
	posts, err := d.GetPosts(threadID)
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusBadRequest, "Uh oh")
	} else {
		c.JSON(http.StatusOK, posts)
	}
}

func getMessage(c *gin.Context, d database.IDatabase, messageID string) {
	message, err := d.GetMessage(messageID)
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusBadRequest, "Uh oh")
	}
	c.JSON(http.StatusOK, message)
}

func postMessage(c *gin.Context, d database.IDatabase) {
	var newMessage model.NewMessage
	c.BindJSON(&newMessage)
	log.Info(newMessage)
	message, err := d.PostMessage(&newMessage)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, message)
}

func postMessagePost(c *gin.Context, d database.IDatabase) {
	var message model.MessagePost
	c.BindJSON(&message)
	newMessage, err := d.PostMessagePost(&message)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	} else {
		_, err := json.Marshal(&newMessage)
		if err != nil {
			return
		}

		c.JSON(http.StatusCreated, newMessage)
	}
}

func postThread(c *gin.Context, d database.IDatabase) {
	var newThread model.NewThread
	c.BindJSON(&newThread)
	thread, err := d.PostThread(&newThread)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, thread)
}

func postPost(c *gin.Context, d database.IDatabase) {
	var post model.Post
	c.BindJSON(&post)
	newPost, err := d.PostPost(&post)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	} else {
		bytes, err := json.Marshal(&newPost)
		if err != nil {
			return
		}

		c.JSON(http.StatusCreated, newPost)
		if c, err := gRedisConn(); err != nil {
			log.Printf("Error on redis conn. %s", err)
		} else {
			c.Do("PUBLISH", "posts", bytes)
		}
	}
}

func editPost(c *gin.Context, d database.IDatabase, postID string) {
	var post model.Post
	c.BindJSON(&post)
	post, err := d.EditPost(postID, post.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	} else {

		bytes, err := json.Marshal(&post)
		if err != nil {
			return
		}

		c.JSON(http.StatusOK, post)
		if c, err := gRedisConn(); err != nil {
			log.Printf("Error on redis conn. %s", err)
		} else {
			c.Do("PUBLISH", "posts", bytes)
		}
	}
}

func deleteThread(c *gin.Context, d database.IDatabase, threadID string) {
	err := d.DeleteThread(threadID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	} else {
		c.Status(http.StatusOK)
	}
}

func checkCredentials(c *gin.Context, d database.IDatabase) {
	var credentials model.Credentials
	c.BindJSON(&credentials)
	var user model.User

	user, err := d.GetUser(credentials.Username)
	if err != nil {
		log.WithFields(log.Fields{"username": credentials.Username}).Error("Can't find that username")
		c.JSON(http.StatusBadRequest, gin.H{"err": "Can't find that username"})
		return
	}

	err = bcrypt.CompareHashAndPassword(user.Password, []byte(credentials.Password))
	if err != nil {
		log.WithFields(log.Fields{"username": user.Username}).Error("Wrong password")
		c.JSON(http.StatusUnauthorized, gin.H{"err": "Wrong password"})
		return
	}

	token := jwt.New(jwt.SigningMethodRS256)
	claims := make(jwt.MapClaims)
	claims["exp"] = time.Now().Add(time.Hour * 24 * 7).Unix() // expires in one week
	claims["iat"] = time.Now().Unix()
	claims["user"] = user.Username
	claims["id"] = user.ID
	claims["role"] = user.UserRole
	token.Claims = claims

	tokenString, err := token.SignedString(signKey)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
		log.Error("Error signing the token")
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

func createUser(c *gin.Context, d database.IDatabase) {
	var registration model.Registration
	c.BindJSON(&registration)

	user := model.User{Username: registration.Username, EmailAddress: registration.EmailAddress}
	_ = user.HashPassword(registration.Password)
	id, err := d.CreateUser(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
	} else {
		c.JSON(http.StatusCreated, id)
	}
}
