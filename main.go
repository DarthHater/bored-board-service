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

	"github.com/DarthHater/bored-board-service/model"
	"github.com/garyburd/redigo/redis"

	"github.com/DarthHater/bored-board-service/database"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	ginlogrus "github.com/toorop/gin-logrus"
)

var (
	db          database.IDatabase
	gPubSubConn *redis.PubSubConn
	gRedisConn  = func() (redis.Conn, error) {
		viper.BindEnv("REDIS_URL")
		redisURL := viper.GetString("REDIS_URL")
		log.Print(redisURL)
		if redisURL != "" {
			return redis.DialURL(redisURL)
		} else {
			return redis.Dial("tcp", "redis_db:6379")
		}
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

	// TODO: We will need to set this to something sane
	r.Use(cors.Default())
	r.Run(":8000")

}

func setupRouter(d database.IDatabase) *gin.Engine {
	log := log.New()
	r := gin.New()
	r.Use(ginlogrus.Logger(log), gin.Recovery())

	err := d.InitDb("development", "./.environment")
	if err != nil {
		log.Fatal(err)
	}

	r.GET("/ws", func(c *gin.Context) {
		webSocketHandler(c.Writer, c.Request)
	})

	r.GET("/thread/:threadid", func(c *gin.Context) {
		threadId := c.Param("threadid")
		getThread(c, d, threadId)
	})

	r.GET("/post/:postid", func(c *gin.Context) {
		postId := c.Param("postid")
		getPost(c, d, postId)
	})

	r.GET("/posts/:threadid", func(c *gin.Context) {
		threadId := c.Param("threadid")
		getPosts(c, d, threadId)
	})

	r.GET("/threads", func(c *gin.Context) {
		getThreads(c, d, 20)
	})

	r.POST("/thread", func(c *gin.Context) {
		postThread(c, d)
	})

	r.POST("/post", func(c *gin.Context) {
		postPost(c, d)
	})

	return r
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
func getThread(c *gin.Context, d database.IDatabase, threadId string) {
	thread, err := d.GetThread(threadId)
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusBadRequest, "Uh oh")
	}
	c.JSON(http.StatusOK, thread)
}

func getPost(c *gin.Context, d database.IDatabase, postId string) {
	post, err := d.GetPost(postId)
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusBadRequest, "Uh oh")
	}
	c.JSON(http.StatusOK, post)
}

func getThreads(c *gin.Context, d database.IDatabase, num int) {
	threads, err := d.GetThreads(num)
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusBadRequest, "Uh oh")
	} else {
		c.JSON(http.StatusOK, threads)
	}
}

func getPosts(c *gin.Context, d database.IDatabase, threadId string) {
	posts, err := d.GetPosts(threadId)
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusBadRequest, "Uh oh")
	} else {
		c.JSON(http.StatusOK, posts)
	}
}

func postThread(c *gin.Context, d database.IDatabase) {
	var newThread model.NewThread
	c.BindJSON(&newThread)
	id, err := d.PostThread(&newThread)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	} else {
		c.JSON(http.StatusCreated, gin.H{"id": id})
	}
}

func postPost(c *gin.Context, d database.IDatabase) {
	var post model.Post
	c.BindJSON(&post)
	id, err := d.PostPost(&post)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	} else {
		post, err = d.GetPost(id)
		if err != nil {
			log.Error("Cannot get post")
		}
		if bytes, err := json.Marshal(&post); err != nil {
			return
		} else {
			c.JSON(http.StatusCreated, gin.H{"id": id})
			if c, err := gRedisConn(); err != nil {
				log.Printf("Error on redis conn. %s", err)
			} else {
				c.Do("PUBLISH", "posts", bytes)
			}
		}
	}
}
