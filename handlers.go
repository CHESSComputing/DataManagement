package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SiteParams represents site URI parameter for /storage/:site end-point
type SiteParams struct {
	Site string `uri:"site" binding:"required"`
}

// BucketParams represents site URI parameter for /storage/:site/:bucket end-point
type BucketParams struct {
	SiteParams
	Bucket string `uri:"bucket" binding:"required"`
}

// ObjectParams represents site URI parameter for /storage/:site/:bucket/:object end-point
type ObjectParams struct {
	BucketParams
	Object string `uri:"object" binding:"required"`
}

// GET handlers

// StorageHandler provides access to GET /storage end-point
/*
```
curl http://localhost:8340/storage
```
*/
func StorageHandler(c *gin.Context) {
	data := sites()
	c.JSON(200, gin.H{"status": "ok", "data": data})
}

// SiteHandler provides access to GET /storage/:site end-point
/*
```
curl http://localhost:8340/storage/cornell/s3-bucket
```
*/
func SiteHandler(c *gin.Context) {
	var params SiteParams
	if err := c.ShouldBindUri(&params); err == nil {
		if data, err := siteContent(params.Site); err == nil {
			c.JSON(200, gin.H{"status": "ok", "data": data})
		} else {
			c.JSON(400, gin.H{"status": "fail", "error": err.Error()})
		}
	} else {
		c.JSON(400, gin.H{"status": "fail", "error": err.Error()})
	}
}

// BucketHandler provides access to GET /storate/:site/:bucket end-point
/*
```
curl http://localhost:8340/storage/cornell/s3-bucket/archive.zip > archive.zip
```
*/
func BucketHandler(c *gin.Context) {
	var params BucketParams
	if err := c.ShouldBindUri(&params); err == nil {
		if data, err := bucketContent(params.Site, params.Bucket); err == nil {
			c.JSON(200, gin.H{"status": "ok", "data": data})
		} else {
			c.JSON(400, gin.H{"status": "fail", "error": err.Error()})
		}
	} else {
		c.JSON(400, gin.H{"status": "fail", "error": err.Error()})
	}
}

// FileHandler provides access to GET /storage/:site/:bucket/:object end-point
func FileHandler(c *gin.Context) {
	var params ObjectParams
	if err := c.ShouldBindUri(&params); err == nil {
		if data, err := getObject(params.Site, params.Bucket, params.Object); err == nil {
			header := fmt.Sprintf("attachment; filename=%s", params.Object)
			c.Header("Content-Disposition", header)
			c.Data(http.StatusOK, "application/octet-stream", data)
		} else {
			c.JSON(400, gin.H{"status": "fail", "error": err.Error()})
		}
	} else {
		c.JSON(400, gin.H{"status": "fail", "error": err.Error()})
	}
}

// POST handlers

// BucketPostHandler provides access to POST /storate/:site/:bucket end-point
/*
```
curl -X POST http://localhost:8340/storage/cornell/s3-bucket
 ```
*/
func BucketPostHandler(c *gin.Context) {
	var params BucketParams
	if err := c.ShouldBindUri(&params); err == nil {
		if err := createBucket(params.Site, params.Bucket); err == nil {
			msg := fmt.Sprintf("Bucket %s/%s created successfully", params.Site, params.Bucket)
			c.JSON(200, gin.H{"status": "ok", "msg": msg})
		} else {
			c.JSON(400, gin.H{"status": "fail", "error": err.Error()})
		}
	} else {
		c.JSON(400, gin.H{"status": "fail", "error": err.Error()})
	}
}

// FilePostHandler provides access to POST /storate/:site/:bucket/:object end-point
/*
```
 curl -X POST http://localhost:8340/storage/cornell/s3-bucket/archive.zip \
  -F "file=@/path/test.zip" \
  -H "Content-Type: multipart/form-data"
```
*/
func FilePostHandler(c *gin.Context) {
	var params ObjectParams
	if err := c.ShouldBindUri(&params); err == nil {
		// single file
		file, err := c.FormFile("file")
		if err != nil {
			log.Println("ERROR: fail to get file from HTTP form", err)
			c.JSON(400, gin.H{"status": "fail", "error": err.Error()})
			return
		}
		log.Printf("INFO: uploading file %s", file.Filename)

		// Upload the file to specific dst.
		reader, err := file.Open()
		if err != nil {
			log.Println("ERROR: fail to open file", err)
			c.JSON(400, gin.H{"status": "fail", "error": err.Error()})
			return
		}
		defer reader.Close()
		size := file.Size
		ctype := "" // TODO: decide on how to read content-type

		if info, err := uploadObject(
			params.Site,
			params.Bucket,
			params.Object,
			ctype,
			reader,
			size); err == nil {
			msg := fmt.Sprintf("File %s/%s/%s uploaded successfully", params.Site, params.Bucket, params.Object)
			c.JSON(200, gin.H{"status": "ok", "msg": msg, "object": info})
		} else {
			log.Println("ERROR: fail to upload object", err)
			c.JSON(400, gin.H{"status": "fail", "error": err.Error()})
		}
	} else {
		log.Println("ERROR: fail to bind HTTP parameters", err)
		c.JSON(400, gin.H{"status": "fail", "error": err.Error()})
	}
}

// DELETE handlers

// BucketDeleteHandler provides access to DELETE /storate/:site/:bucket end-point
/*
```
curl -X DELETE http://localhost:8340/storage/cornell/s3-bucket
```
*/
func BucketDeleteHandler(c *gin.Context) {
	var params BucketParams
	if err := c.ShouldBindUri(&params); err == nil {
		if err := deleteBucket(params.Site, params.Bucket); err == nil {
			msg := fmt.Sprintf("Bucket %s/%s deleted successfully", params.Site, params.Bucket)
			c.JSON(200, gin.H{"status": "ok", "msg": msg})
		} else {
			c.JSON(400, gin.H{"status": "fail", "error": err.Error()})
		}
	} else {
		c.JSON(400, gin.H{"status": "fail", "error": err.Error()})
	}
}

// BucketDeleteHandler provides access to DELETE /storate/:site/:bucket end-point
/*
```
curl -X DELETE http://localhost:8340/storage/cornell/s3-bucket/archive.zip
```
*/
func FileDeleteHandler(c *gin.Context) {
	var params ObjectParams
	if err := c.ShouldBindUri(&params); err == nil {
		var versionId string // TODO: in a future we may need to handle different version of objects
		if err := deleteObject(params.Site, params.Bucket, params.Object, versionId); err == nil {
			msg := fmt.Sprintf("File %s/%s/%s deleted successfully", params.Site, params.Bucket, params.Object)
			c.JSON(200, gin.H{"status": "ok", "msg": msg})
		} else {
			c.JSON(400, gin.H{"status": "fail", "error": err.Error()})
		}
	} else {
		c.JSON(400, gin.H{"status": "fail", "error": err.Error()})
	}
}
