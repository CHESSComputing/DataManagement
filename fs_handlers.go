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

var fsClient *LocalFsClient

// StorageParams represents site URI parameter for /storage/:dir end-point
type StorageParams struct {
	Dir string `uri:"dir" binding:"required"`
}

// FileStorageParams represents site URI parameter for /storage/:dir/:file end-point
type FileStorageParams struct {
	StorageParams
	File string `uri:"file" binding:"required"`
}

// GET handlers

// FsStorageHandler provides access to GET /storage/:dir/:file end-point
/*
```
# get list of storage
curl http://localhost:8340/storage
# get list of specific dir in a storage
curl http://localhost:8340/storage/dir
# get concrete file from storage dir
curl http://localhost:8340/storage/dir/archive.zip
```
*/
func FsStorageHandler(c *gin.Context) {
	var fParams FileStorageParams
	var dParams StorageParams
	if err := c.ShouldBindUri(&fParams); err == nil {
		if data, err := fsClient.Get(fParams.Dir, fParams.File); err == nil {
			header := fmt.Sprintf("attachment; filename=%s", fParams.File)
			c.Header("Content-Disposition", header)
			c.Data(http.StatusOK, "application/octet-stream", data)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "error": err.Error()})
		}
	} else if err := c.ShouldBindUri(&dParams); err == nil {
		if data, err := fsClient.List(dParams.Dir); err == nil {
			c.JSON(http.StatusOK, gin.H{"status": "ok", "data": data})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "error": err.Error()})
		}
	}
	// get list of dirs
	data, err := fsClient.List("")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "data": data})

}

// POST handlers

// FsPostHandler provides access to POST /storate/:dir end-point
/*
```
curl -X POST http://localhost:8340/storage/dir
curl -X POST http://localhost:8340/storage/dir/archive.zip \
     -F "file=@/path/test.zip" \
     -H "Content-Type: multipart/form-data"
 ```
*/
func FsPostHandler(c *gin.Context) {
	var dParams StorageParams
	var fParams FileStorageParams
	if err := c.ShouldBindUri(&dParams); err == nil {
		if err := fsClient.Create(dParams.Dir); err == nil {
			msg := fmt.Sprintf("Dir %s created successfully", dParams.Dir)
			c.JSON(http.StatusOK, gin.H{"status": "ok", "msg": msg})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "error": err.Error()})
		}
	} else if err := c.ShouldBindUri(&fParams); err == nil {
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

		if err := fsClient.Upload(fParams.Dir, fParams.File, ctype, reader, size); err == nil {
			msg := fmt.Sprintf("File %s/%s uploaded successfully", fParams.Dir, fParams.File)
			c.JSON(http.StatusOK, gin.H{"status": "ok", "msg": msg})
		} else {
			log.Println("ERROR: fail to upload file", err)
			c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "error": err.Error()})
		}
	} else {
		log.Println("ERROR: fail to bind HTTP parameters", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "error": err.Error()})
	}
}

// DELETE handlers

// FsDeleteHandler provides access to DELETE /storate/:dir end-point
/*
```
curl -X DELETE http://localhost:8340/storage/dir
curl -X DELETE http://localhost:8340/storage/dir/archive.zip
```
*/
func FsDeleteHandler(c *gin.Context) {
	var dParams StorageParams
	var fParams FileStorageParams
	if err := c.ShouldBindUri(&dParams); err == nil {
		if err := fsClient.Delete(dParams.Dir, ""); err == nil {
			msg := fmt.Sprintf("Dir %s deleted successfully", dParams.Dir)
			c.JSON(http.StatusOK, gin.H{"status": "ok", "msg": msg})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "error": err.Error()})
		}
	} else if err := c.ShouldBindUri(&fParams); err == nil {
		if err := fsClient.Delete(fParams.Dir, fParams.File); err == nil {
			msg := fmt.Sprintf("File %s/%s deleted successfully", fParams.Dir, fParams.File)
			c.JSON(http.StatusOK, gin.H{"status": "ok", "msg": msg})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "error": err.Error()})
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"status": "fail", "error": err.Error()})
	}
}
