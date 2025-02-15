package main

import (
	"net/http"
	"os"
	"path/filepath"

	srvConfig "github.com/CHESSComputing/golib/config"
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
	sdir := c.Query("sdir")

	// Find metadata record for given DID
	meta, err := findMetaDataRecord(did)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "metadata record not found"})
		return
	}

	// Extract data location from metadata record
	for _, attr := range srvConfig.Config.CHESSMetaData.DataLocationAttributes {
		if val, ok := meta[attr]; ok {
			// if location attribute is found redirect to it
			path := val.(string)
			// join path from meta-data record with possible sdir
			if sdir != "" {
				path = filepath.Join(path, sdir)
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
				entries, err := getFileList(path)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot read directory"})
					return
				}

				if acceptHeader == "application/json" {
					c.JSON(http.StatusOK, entries)
					return
				}

				// Render HTML template
				c.HTML(http.StatusOK, "fs.tmpl", gin.H{
					"Path":    path,
					"Entries": entries,
				})
				return
			}

			// Serve file content if it's a file
			http.ServeFile(c.Writer, c.Request, path)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "data location not found in metadata"})
}
