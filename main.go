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
	"fmt"
	"github.com/darthhater/bored-board-service/model"
	"net/http"
	
	"github.com/darthhater/bored-board-service/database"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/toorop/gin-logrus"
)

var db database.IDatabase

func main() {
	d := database.Database{}
	db = &d
	r := setupRouter(db)
	r.Use(gin.Logger())
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

	r.GET("/thread/:threadid", func (c *gin.Context) {
		threadId := c.Param("threadid")
		getThread(c, d, threadId)
	})

	r.POST("/thread", func(c *gin.Context) {
		postThread(c, d)
	})

	return r
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

func postThread(c *gin.Context, d database.IDatabase) {
	var newThread model.NewThread
	c.BindJSON(&newThread)
	fmt.Println(&newThread.P)
	id, err := d.PostThread(&newThread)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
	} else {
		c.JSON(201, gin.H{"id": id})
	}
}
