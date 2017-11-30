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
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/toorop/gin-logrus"
)

type Thread struct {
	ID    int32
	Title string
}

func main() {
	r := setupRouter()
	r.Use(gin.Logger())
	r.Run(":8000")
}

func setupRouter() *gin.Engine {
	log := log.New()
	r := gin.New()
	r.Use(ginlogrus.Logger(log), gin.Recovery())

	r.GET("/thread/:id", getThread)
	r.GET("/thread", getThreads)
	r.GET("/user/:id", getUser)

	return r
}

// Handlers
func getThread(c *gin.Context) {
	/* id := c.Param("id") */
	c.JSON(http.StatusOK, gin.H{"Test": "Test Response"})
}

func getThreads(c *gin.Context) {
	threads := []Thread{
		Thread{ID: 1, Title: "Ok"},
		Thread{ID: 2, Title: "K"},
	}
	c.JSON(http.StatusOK, threads)
}

func getUser(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"User": id})
}
