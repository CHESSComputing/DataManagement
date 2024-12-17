package main

// handlers module
//
// Copyright (c) 2023 - Valentin Kuznetsov <vkuznet@gmail.com>
//
import (
	"fmt"
	"log"
	"net/http"

	s3 "github.com/CHESSComputing/golib/s3"
	"github.com/gin-gonic/gin"
)

// BucketParams represents site URI parameter for /storage/:bucket end-point
type BucketParams struct {
	Bucket string `uri:"bucket" binding:"required"`
}

// ObjectParams represents site URI parameter for /storage/:bucket/:object end-point
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
	// get list of buckets
	buckets, err := s3.ListBuckets()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "data": buckets})
}

// BucketHandler provides access to GET /storate/:bucket end-point
/*
```
curl http://localhost:8340/storage/s3-bucket/archive.zip > archive.zip
```
*/
func BucketHandler(c *gin.Context) {
	var params BucketParams
	if err := c.ShouldBindUri(&params); err == nil {
		if data, err := s3.BucketContent(params.Bucket); err == nil {
			c.JSON(http.StatusOK, gin.H{"status": "ok", "data": data})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "error": err.Error()})
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "error": err.Error()})
	}
}

// FileHandler provides access to GET /storage/:bucket/:object end-point
func FileHandler(c *gin.Context) {
	var params ObjectParams
	if err := c.ShouldBindUri(&params); err == nil {
		if data, err := s3.GetObject(params.Bucket, params.Object); err == nil {
			header := fmt.Sprintf("attachment; filename=%s", params.Object)
			c.Header("Content-Disposition", header)
			c.Data(http.StatusOK, "application/octet-stream", data)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "error": err.Error()})
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "error": err.Error()})
	}
}

// POST handlers

// BucketPostHandler provides access to POST /storate/:bucket end-point
/*
```
curl -X POST http://localhost:8340/storage/s3-bucket
 ```
*/
func BucketPostHandler(c *gin.Context) {
	var params BucketParams
	if err := c.ShouldBindUri(&params); err == nil {
		if err := s3.CreateBucket(params.Bucket); err == nil {
			msg := fmt.Sprintf("Bucket %s created successfully", params.Bucket)
			c.JSON(http.StatusOK, gin.H{"status": "ok", "msg": msg})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "error": err.Error()})
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "error": err.Error()})
	}
}

// FilePostHandler provides access to POST /storate/:bucket/:object end-point
/*
```
 curl -X POST http://localhost:8340/storage/s3-bucket/archive.zip \
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
			c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "error": err.Error()})
			return
		}
		log.Printf("INFO: uploading file %s", file.Filename)

		// Upload the file to specific dst.
		reader, err := file.Open()
		if err != nil {
			log.Println("ERROR: fail to open file", err)
			c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "error": err.Error()})
			return
		}
		defer reader.Close()
		size := file.Size
		ctype := "" // TODO: decide on how to read content-type

		if err := s3.UploadObject(params.Bucket, params.Object, ctype, reader, size); err == nil {
			msg := fmt.Sprintf("File %s/%s uploaded successfully", params.Bucket, params.Object)
			c.JSON(http.StatusOK, gin.H{"status": "ok", "msg": msg})
		} else {
			log.Println("ERROR: fail to upload object", err)
			c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "error": err.Error()})
		}
	} else {
		log.Println("ERROR: fail to bind HTTP parameters", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "error": err.Error()})
	}
}

// DELETE handlers

// BucketDeleteHandler provides access to DELETE /storate/:bucket end-point
/*
```
curl -X DELETE http://localhost:8340/storage/s3-bucket
```
*/
func BucketDeleteHandler(c *gin.Context) {
	var params BucketParams
	if err := c.ShouldBindUri(&params); err == nil {
		if err := s3.DeleteBucket(params.Bucket); err == nil {
			msg := fmt.Sprintf("Bucket %s deleted successfully", params.Bucket)
			c.JSON(http.StatusOK, gin.H{"status": "ok", "msg": msg})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "error": err.Error()})
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "error": err.Error()})
	}
}

// BucketDeleteHandler provides access to DELETE /storate/:bucket end-point
/*
```
curl -X DELETE http://localhost:8340/storage/s3-bucket/archive.zip
```
*/
func FileDeleteHandler(c *gin.Context) {
	var params ObjectParams
	if err := c.ShouldBindUri(&params); err == nil {
		var versionId string // TODO: in a future we may need to handle different version of objects
		if err := s3.DeleteObject(params.Bucket, params.Object, versionId); err == nil {
			msg := fmt.Sprintf("File %s/%s deleted successfully", params.Bucket, params.Object)
			c.JSON(http.StatusOK, gin.H{"status": "ok", "msg": msg})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "error": err.Error()})
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "error": err.Error()})
	}
}
