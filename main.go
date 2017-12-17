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
	"github.com/darthhater/bored-board-service/model"
	"net/http"

	"github.com/darthhater/bored-board-service/database"
	"github.com/darthhater/bored-board-service/redis"
	redigo "github.com/garyburd/redigo/redis"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"github.com/toorop/gin-logrus"
)

var (
	db           database.IDatabase
	rr           redis.RedisReciever
	rw           redis.RedisWriter
	redisAddress string
)

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

	// Redis

	redisAddress = "redis_db:6379"

	redisPool := redigo.NewPool(func() (redigo.Conn, error) {
		c, err := redigo.Dial("tcp", redisAddress)

		if err != nil {
			return nil, err
		}

		return c, nil
	}, 10)

	defer redisPool.Close()

	rr = redis.NewRedisReciever(redisPool)
	rw = redis.NewRedisWriter(redisPool)

	go func() {
		for {
			err := rr.Run("posts")
			if err == nil {
				break
			}
			log.Print(err)
		}
	}()

	go func() {
		for {
			err := rw.Run("posts")
			if err == nil {
				break
			}
			log.Print(err)
		}
	}()

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
		log.Infoln("Failed to set websocket upgrade: %+v", err)
		return
	}

	rr.Register(conn)

	for {
		t, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		switch t {
		case websocket.TextMessage:
			log.Printf("Made it here: %s", msg)
			rw.Publish(msg)
		default:
			log.Warning("Unknown message")
		}
	}

	rr.DeRegister(conn)

	conn.WriteMessage(websocket.CloseMessage, []byte{})
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
		c.JSON(http.StatusCreated, gin.H{"id": id})
	}
}
