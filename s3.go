package main

// s3 module
//
// Copyright (c) 2023 - Valentin Kuznetsov <vkuznet@gmail.com>
//
import (
	"log"

	srvConfig "github.com/CHESSComputing/golib/config"
	minio "github.com/minio/minio-go/v7"
)

// BucketObject represents s3 object
type BucketObject struct {
	Bucket  string             `json:"bucket"`
	Objects []minio.ObjectInfo `json:"objects"`
}

// DiscoveryRecord represents structure of data discovery record
type DiscoveryRecord struct {
	Name         string `json:"name"`
	URL          string `json:"url"`
	Endpoint     string `json:"endpoint""`
	AccessKey    string `json:"access_key"`
	AccessSecret string `json:"access_secret"`
	UseSSL       bool   `json:"use_ssl"`
}

// bucketContent provides content on given bucket
func bucketContent(bucket string) (BucketObject, error) {
	s3 := S3{
		Endpoint:     srvConfig.Config.DataManagement.S3.Endpoint,
		AccessKey:    srvConfig.Config.DataManagement.S3.AccessKey,
		AccessSecret: srvConfig.Config.DataManagement.S3.AccessSecret,
		UseSSL:       srvConfig.Config.DataManagement.S3.UseSSL,
	}
	if srvConfig.Config.DataManagement.WebServer.Verbose > 0 {
		log.Printf("Use %v", s3)
	}
	if srvConfig.Config.DataManagement.WebServer.Verbose > 0 {
		log.Printf("looking for bucket:'%s'", bucket)
	}
	objects, err := listObjects(s3, bucket)
	if err != nil {
		log.Printf("ERROR: unabel to list bucket '%s', error %v", bucket, err)
	}
	obj := BucketObject{
		Bucket:  bucket,
		Objects: objects,
	}
	return obj, nil
}
