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
	"net/http"
	"os"
	"time"
	"strconv"

	"github.com/DarthHater/bored-board-service/auth"
	"github.com/DarthHater/bored-board-service/constants"
	"github.com/DarthHater/bored-board-service/model"
	"github.com/garyburd/redigo/redis"
	"golang.org/x/crypto/bcrypt"
	"github.com/DarthHater/bored-board-service/database"
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
	db         database.IDatabase
	a		auth.IAuth
	gPubSubConn *redis.PubSubConn
	gRedisConn  = func() (redis.Conn, error) {
		redisURL := os.Getenv("REDIS_URL")
		if redisURL != "" {
			return redis.DialURL(redisURL)
		}

		return redis.Dial("tcp", "redis_db:6379")
	}
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

func init() {
	au := auth.Auth{}
	a = &au

	a.ReadAndSetKeys();
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
	gPubSubConn.Subscribe("message_posts")
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

	authGroup := r.Group("/")

	authGroup.Use(a.UserIsLoggedIn())
	{
		authGroup.GET("/thread/:threadid", func(c *gin.Context) {
			threadID := c.Param("threadid")
			getThread(c, d, threadID)
		})

		authGroup.GET("/post/:postid", func(c *gin.Context) {
			postID := c.Param("postid")
			getPost(c, d, postID)
		})

		authGroup.GET("/posts/:threadid", func(c *gin.Context) {
			threadID := c.Param("threadid")
			getPosts(c, d, threadID, 20, "", "")
		})

		authGroup.GET("/posts/:threadid/:since", func(c *gin.Context) {
			threadID := c.Param("threadid")
			since := c.Param("since")
			direction := c.Query("direction")
			getPosts(c, d, threadID, 20, since, direction)
		})

		authGroup.GET("/threads/:since", func(c *gin.Context) {
			since := c.Param("since")
			getThreads(c, d, 20, since)
		})

		authGroup.GET("/message/:messageid", func(c *gin.Context) {
			messageID := c.Param("messageid")
			getMessage(c, d, messageID)
		})

		authGroup.GET("/messages/:userid", func(c *gin.Context) {
			userID := c.Param("userid")
			getMessages(c, d, 20, userID)
		})

		authGroup.GET("/messageposts/:messageid", func(c *gin.Context) {
			messageID := c.Param("messageid")
			getMessagePosts(c, d, messageID)
		})

		authGroup.POST("/thread", func(c *gin.Context) {
			postThread(c, d)
		})

		authGroup.POST("/post", func(c *gin.Context) {
			postPost(c, d)
		})

		authGroup.POST("/newmessage", func(c *gin.Context) {
			postMessage(c, d)
		})

		authGroup.POST("/message", func(c *gin.Context) {
			postMessagePost(c, d)
		})

		authGroup.PATCH("/posts/:postid", func(c *gin.Context) {
			postID := c.Param("postid")
			editPost(c, d, postID)
		})

		authGroup.GET("/user/:userid", func(c *gin.Context) {
			userID := c.Param("userid")
			getUserInfo(c, d, userID)
		})

		authGroup.GET("/users", func(c *gin.Context) {
			search := c.Query("search")
			getUsers(c, d, search)
		})

		authGroup.Use(a.UserIsInRole(d, []constants.Role{constants.Admin, constants.Mod}))
		{
			authGroup.DELETE("/thread/:threadid", func(c *gin.Context) {
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

func getThreads(c *gin.Context, d database.IDatabase, num int, since string) {
	threads, err := d.GetThreads(num, since)
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

func getPosts(c *gin.Context, d database.IDatabase, threadID string, num int, since string, direction string) {
	value, _ := a.GetTokenValue(c, constants.UserID)
	userID := value.(string)
	dir, _ := strconv.Atoi(direction)
	posts, err := d.GetPosts(threadID, 20, since, userID, constants.Direction(dir))

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
		bytes, err := json.Marshal(&newMessage)
		if err != nil {
			return
		}

		c.JSON(http.StatusCreated, newMessage)

		if c, err := gRedisConn(); err != nil {
			log.Printf("Error on redis conn. %s", err)
		} else {
			c.Do("PUBLISH", "message_posts", bytes)
		}
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

	tokenString, err := a.CreateToken(user);

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
