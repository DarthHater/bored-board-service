package main

import (
	"net/http"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/toorop/gin-logrus"
)

func setupRouter() *gin.Engine {
	log := log.New()
	r := gin.New()
	r.Use(ginlogrus.Logger(log), gin.Recovery())

	r.GET("/thread", getThread)

	return r
}

func main() {
	r := setupRouter()
	r.Use(gin.Logger())
	r.Run(":8000")
}

// Handlers
func getThread(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"Test": "Test Response"})
}
