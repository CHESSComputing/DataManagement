package main

import (
	"log"
	"net/http"
	"net/url"
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
				tmpl["FileExtensions"] = fileExtensions()
				content := server.TmplPage(StaticFs, "fs.tmpl", tmpl)
				page := server.Header(StaticFs, base) + content + server.FooterEmpty(StaticFs, base)
				c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(page))
				return
			}

			// Serve file content if it's a file
			http.ServeFile(c.Writer, c.Request, path)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "data location not found in metadata"})
}

// DataFilesHandler provides access to data files
func DataFilesHandler(c *gin.Context) {
	// Get DID from HTTP request
	did := c.Query("did")
	if did == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing did parameter"})
		return
	}
	pattern := c.Query("pattern")
	if pattern == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "no files pattern is provided"})
		return
	}
	if val, err := url.QueryUnescape(did); err == nil {
		did = val
	}
	if val, err := url.QueryUnescape(pattern); err == nil {
		pattern = val
	}

	// Find metadata record for given DID
	meta, err := findMetaDataRecord(did)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "metadata record not found"})
		return
	}

	// Extract data location from metadata record
	for _, attr := range srvConfig.Config.CHESSMetaData.DataLocationAttributes {
		if val, ok := meta[attr]; ok {
			// take data location
			path := val.(string)

			// find all files in that location using our pattern
			files, err := findFiles(path, pattern)
			if err != nil {
				log.Println("WARNING: findFiles error", err)
				//                 c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				//                 return
			}
			c.JSON(http.StatusOK, files)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "data files not found"})
}
