package main

import (
	"fmt"
	"log"

	srvConfig "github.com/CHESSComputing/golib/config"
	"github.com/gin-gonic/gin"
)

// examples: https://go.dev/doc/tutorial/web-service-gin

// helper function to setup our server router
func setupRouter() *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()

	// GET routes
	r.GET("/storage", StorageHandler)
	r.GET("/storage/:site", SiteHandler)
	r.GET("/storage/:site/:bucket", BucketHandler)
	r.GET("/storage/:site/:bucket/:object", FileHandler)

	// POST routes
	r.POST("/storage/:site/:bucket", BucketPostHandler)
	r.POST("/storage/:site/:bucket/:object", FilePostHandler)

	// DELETE routes
	r.DELETE("/storage/:site/:bucket", BucketDeleteHandler)
	r.DELETE("/storage/:site/:bucket/:object", FileDeleteHandler)
	return r
}

func Server() {
	r := setupRouter()
	sport := fmt.Sprintf(":%d", srvConfig.Config.DataManagement.WebServer.Port)
	log.Printf("Start HTTP server %s", sport)
	r.Run(sport)
}
