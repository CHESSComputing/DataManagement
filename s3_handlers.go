package main

// handlers module
//
// Copyright (c) 2023 - Valentin Kuznetsov <vkuznet@gmail.com>
//
import (
	"fmt"
	"log"
	"net/http"

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

// S3StorageHandler provides access to GET /storage/:bucket/:object end-point
/*
```
# get list of storage
curl http://localhost:8340/storage
# get list of specific bucket in a storage
curl http://localhost:8340/storage/s3-bucket
# get concrete object from storage bucket
curl http://localhost:8340/storage/s3-bucket/archive.zip
```
*/
func S3StorageHandler(c *gin.Context) {
	var objectParams ObjectParams
	var bucketParams BucketParams
	if err := c.ShouldBindUri(&objectParams); err == nil {
		if data, err := s3Client.GetObject(objectParams.Bucket, objectParams.Object); err == nil {
			header := fmt.Sprintf("attachment; filename=%s", objectParams.Object)
			c.Header("Content-Disposition", header)
			c.Data(http.StatusOK, "application/octet-stream", data)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "error": err.Error()})
		}
	} else if err := c.ShouldBindUri(&bucketParams); err == nil {
		if data, err := s3Client.BucketContent(bucketParams.Bucket); err == nil {
			c.JSON(http.StatusOK, gin.H{"status": "ok", "data": data})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "error": err.Error()})
		}
	}
	// get list of buckets
	buckets, err := s3Client.ListBuckets()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "data": buckets})

}

// POST handlers

// S3PostHandler provides access to POST /storate/:bucket end-point
/*
```
curl -X POST http://localhost:8340/storage/s3-bucket
curl -X POST http://localhost:8340/storage/s3-bucket/archive.zip \
     -F "file=@/path/test.zip" \
     -H "Content-Type: multipart/form-data"
 ```
*/
func S3PostHandler(c *gin.Context) {
	var bucketParams BucketParams
	var objectParams ObjectParams
	if err := c.ShouldBindUri(&bucketParams); err == nil {
		if err := s3Client.CreateBucket(bucketParams.Bucket); err == nil {
			msg := fmt.Sprintf("Bucket %s created successfully", bucketParams.Bucket)
			c.JSON(http.StatusOK, gin.H{"status": "ok", "msg": msg})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "error": err.Error()})
		}
	} else if err := c.ShouldBindUri(&objectParams); err == nil {
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

		if err := s3Client.UploadObject(objectParams.Bucket, objectParams.Object, ctype, reader, size); err == nil {
			msg := fmt.Sprintf("File %s/%s uploaded successfully", objectParams.Bucket, objectParams.Object)
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

// S3DeleteHandler provides access to DELETE /storate/:bucket end-point
/*
```
curl -X DELETE http://localhost:8340/storage/s3-bucket
curl -X DELETE http://localhost:8340/storage/s3-bucket/archive.zip
```
*/
func S3DeleteHandler(c *gin.Context) {
	var bucketParams BucketParams
	var objectParams ObjectParams
	if err := c.ShouldBindUri(&bucketParams); err == nil {
		if err := s3Client.DeleteBucket(bucketParams.Bucket); err == nil {
			msg := fmt.Sprintf("Bucket %s deleted successfully", bucketParams.Bucket)
			c.JSON(http.StatusOK, gin.H{"status": "ok", "msg": msg})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "error": err.Error()})
		}
	} else if err := c.ShouldBindUri(&objectParams); err == nil {
		var versionId string // TODO: in a future we may need to handle different version of objects
		if err := s3Client.DeleteObject(objectParams.Bucket, objectParams.Object, versionId); err == nil {
			msg := fmt.Sprintf("File %s/%s deleted successfully", objectParams.Bucket, objectParams.Object)
			c.JSON(http.StatusOK, gin.H{"status": "ok", "msg": msg})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "error": err.Error()})
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "error": err.Error()})
	}
}
