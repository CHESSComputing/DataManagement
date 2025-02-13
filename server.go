package main

// server module
//
// Copyright (c) 2023 - Valentin Kuznetsov <vkuznet@gmail.com>
//
import (
	"log"
	"strings"

	srvConfig "github.com/CHESSComputing/golib/config"
	s3 "github.com/CHESSComputing/golib/s3"
	server "github.com/CHESSComputing/golib/server"
	"github.com/gin-gonic/gin"
)

var s3Client s3.S3Client

// helper function to setup our server router
func setupRouter() *gin.Engine {
	if srvConfig.Config.DataManagement.S3.Name != "" {
		return setupS3Router()
	}
	return setupFSRouter()
}

// helper function to setup server router for S3 backend
func setupS3Router() *gin.Engine {
	routes := []server.Route{
		server.Route{Method: "GET", Path: "/data", Handler: DataLocationHandler, Authorized: true},
		server.Route{Method: "GET", Path: "/storage", Handler: S3StorageHandler, Authorized: true},
		server.Route{Method: "GET", Path: "/storage/:bucket/:object", Handler: S3StorageHandler, Authorized: true},

		server.Route{Method: "POST", Path: "/storage/:bucket", Handler: S3PostHandler, Authorized: true, Scope: "write"},
		server.Route{Method: "POST", Path: "/storage/:bucket/:object", Handler: S3PostHandler, Authorized: true, Scope: "write"},

		server.Route{Method: "DELETE", Path: "/storage/:bucket", Handler: S3DeleteHandler, Authorized: true, Scope: "delete"},
		server.Route{Method: "DELETE", Path: "/storage/:bucket/:object", Handler: S3DeleteHandler, Authorized: true, Scope: "delete"},
	}
	r := server.Router(routes, nil, "static", srvConfig.Config.DataManagement.WebServer)
	return r
}

// helper function to setup our server router for file-system backend
func setupFSRouter() *gin.Engine {
	routes := []server.Route{
		server.Route{Method: "GET", Path: "/data", Handler: DataLocationHandler, Authorized: true},
		server.Route{Method: "GET", Path: "/storage", Handler: FsStorageHandler, Authorized: true},
		server.Route{Method: "GET", Path: "/storage/:dir/:file", Handler: FsStorageHandler, Authorized: true},

		server.Route{Method: "POST", Path: "/storage/:dir", Handler: FsPostHandler, Authorized: true, Scope: "write"},
		server.Route{Method: "POST", Path: "/storage/:dir/:file", Handler: FsPostHandler, Authorized: true, Scope: "write"},

		server.Route{Method: "DELETE", Path: "/storage/:dir", Handler: FsDeleteHandler, Authorized: true, Scope: "delete"},
		server.Route{Method: "DELETE", Path: "/storage/:dir/:file", Handler: FsDeleteHandler, Authorized: true, Scope: "delete"},
	}
	r := server.Router(routes, nil, "static", srvConfig.Config.DataManagement.WebServer)
	return r
}

// Server defines our HTTP server
func Server() {
	// Initialize the appropriate S3 client.
	var err error
	s3Client, err = s3.InitializeS3Client(strings.ToLower(srvConfig.Config.DataManagement.S3.Name))
	if err != nil {
		log.Fatalf("Failed to initialize S3 client %s, error %v", srvConfig.Config.DataManagement.S3.Name, err)
	}

	// setup web router and start the service
	r := setupRouter()
	webServer := srvConfig.Config.DataManagement.WebServer
	server.StartServer(r, webServer)
}
