package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	srvConfig "github.com/CHESSComputing/golib/config"
	server "github.com/CHESSComputing/golib/server"
	"github.com/gin-gonic/gin"
)

// GET handlers
func DataLocationHandler(c *gin.Context) {
	// Get DID from HTTP request
	did := c.Query("did")
	if did == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing did parameter"})
		return
	}
	// the /data URL may contain additional path parameter
	// it refers to sub-path within raw data location
	spath := c.Query("path")

	// if we got file parameter we know which file to fetch
	fileName := c.Query("file")

	// Find metadata record for given DID
	meta, err := findMetaDataRecord(did)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "metadata record not found"})
		return
	}

	// Extract data location from metadata record
	for _, attr := range srvConfig.Config.CHESSMetaData.DataLocationAttributes {
		if val, ok := meta[attr]; ok {
			// if location attribute (raw data location) is found redirect to it
			path := val.(string)
			// join path from meta-data record with possible spath
			if spath != "" {
				path = filepath.Join(path, spath)
			}

			// if we have file name we should present it back to upstream caller
			if fileName != "" {
				fname := filepath.Join(path, fileName)
				log.Println("### serve file", fname)
				// Serve file content if it's a file
				http.ServeFile(c.Writer, c.Request, fname)
				return
			}

			// get info about our path
			info, err := os.Stat(path)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "path not found"})
				return
			}

			// If requesting JSON, return directory listing in JSON format
			acceptHeader := c.GetHeader("Accept")
			if info.IsDir() {
				entries, err := getFileList(did, path, spath)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot read directory"})
					return
				}

				if acceptHeader == "application/json" {
					c.JSON(http.StatusOK, entries)
					return
				}

				// Render HTML template
				tmpl := server.MakeTmpl(StaticFs, "DataManagement")
				base := srvConfig.Config.Frontend.WebServer.Base
				tmpl["Base"] = base
				tmpl["Area"] = path
				tmpl["Entries"] = entries
				tmpl["Did"] = did
				content := server.TmplPage(StaticFs, "fs.tmpl", tmpl)
				c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(content))
				return
			}

			// Serve file content if it's a file
			http.ServeFile(c.Writer, c.Request, path)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "data location not found in metadata"})
}
