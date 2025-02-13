package main

import (
	"net/http"
	"os"

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
			dataLocation := val.(string)
			// if location is individual file
			//             c.File(dataLocation)

			// If location is a directory, list its contents
			files, err := os.ReadDir(dataLocation)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to list directory contents"})
				return
			}

			// Collect file names
			var fileNames []string
			for _, file := range files {
				fileNames = append(fileNames, file.Name())
			}

			c.JSON(http.StatusOK, gin.H{"directory": dataLocation, "files": fileNames})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "data location not found in metadata"})
}
