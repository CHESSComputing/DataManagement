package main

import (
	"fmt"
	"log"

	srvConfig "github.com/CHESSComputing/golib/config"
	server "github.com/CHESSComputing/golib/server"
	"github.com/gin-gonic/gin"
)

// helper function to setup our server router
func setupRouter() *gin.Engine {
	routes := []server.Route{
		server.Route{Method: "GET", Path: "/storage", Handler: StorageHandler, Authorized: true},
		server.Route{Method: "GET", Path: "/storage/:site", Handler: SiteHandler, Authorized: true},
		server.Route{Method: "GET", Path: "/storage/:site/:bucket", Handler: BucketHandler, Authorized: true},
		server.Route{Method: "GET", Path: "/storage/:site/:bucket/:object", Handler: FileHandler, Authorized: true},

		server.Route{Method: "POST", Path: "/storage/:site/:bucket", Handler: BucketPostHandler, Authorized: true},
		server.Route{Method: "POST", Path: "/storage/:site/:bucket/:object", Handler: FilePostHandler, Authorized: true},

		server.Route{Method: "DELETE", Path: "/storage/:site/:bucket", Handler: BucketDeleteHandler, Authorized: true},
		server.Route{Method: "DELETE", Path: "/storage/:site/:bucket/:object", Handler: FileDeleteHandler, Authorized: true},
	}
	r := server.Router(routes, nil, "static",
		srvConfig.Config.DataManagement.WebServer.Base,
		srvConfig.Config.DataManagement.WebServer.Verbose,
	)
	return r
}

// Server defines our HTTP server
func Server() {
	r := setupRouter()
	sport := fmt.Sprintf(":%d", srvConfig.Config.DataManagement.WebServer.Port)
	log.Printf("Start HTTP server %s", sport)
	r.Run(sport)
}
