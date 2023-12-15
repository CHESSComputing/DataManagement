package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	srvConfig "github.com/CHESSComputing/golib/config"
	minio "github.com/minio/minio-go/v7"
	cryptoutils "github.com/vkuznet/cryptoutils"
)

// Site represents Site object returned from discovery service
type Site struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Description string `json:"description"`
}

// SiteObject represents site object
type SiteObject struct {
	Site    string             `json:"site"`
	Buckets []minio.BucketInfo `json:"buckets"`
}

// BucketObject represents site object
type BucketObject struct {
	Site    string             `json:"site"`
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

// helper function to fetch sites info from discovery service
func sites() []Site {
	var out []Site
	rurl := fmt.Sprintf("%s/sites", srvConfig.Config.Services.DiscoveryURL)
	if srvConfig.Config.DataManagement.WebServer.Verbose > 0 {
		log.Println("query Discovery service", rurl)
	}
	resp, err := http.Get(rurl)
	if err != nil {
		log.Println("ERROR:", err)
		return out
	}
	defer resp.Body.Close()
	var results []Site
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&results); err != nil {
		log.Println("ERROR:", err)
		return out
	}
	return results
}

// helper function to return S3 object for given site
func site2s3(site string) (S3, error) {
	var s3 S3
	if srvConfig.Config.DataManagement.WebServer.Verbose > 0 {
		log.Printf("looking for site:%s", site)
	}
	surl := fmt.Sprintf("%s/sites", srvConfig.Config.Services.DiscoveryURL)
	if srvConfig.Config.DataManagement.WebServer.Verbose > 0 {
		log.Println("query Discovery service", surl)
	}
	resp, err := http.Get(surl)
	if err != nil {
		log.Printf("ERROR: unable to contact DataDiscovery service %s, error %v", surl, err)
		return s3, err
	}
	// read data discovery content
	var records []DiscoveryRecord
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("ERROR: unable to read DataDiscovery response, error %v", err)
		return s3, err
	}
	err = json.Unmarshal(body, &records)
	if err != nil {
		log.Printf("ERROR: unable to unmarshal DataDiscovery response, error %v", err)
		return s3, err
	}
	if srvConfig.Config.DataManagement.WebServer.Verbose > 0 {
		log.Printf("site records %+v", records)
	}

	for _, rec := range records {
		if rec.Name == site {
			log.Printf("INFO: found %s in DataDiscovery records, will access its s3 via %s", rec.Name, rec.URL)
			akey, err := cryptoutils.HexDecrypt(rec.AccessKey, srvConfig.Config.Encryption.Secret, srvConfig.Config.Encryption.Cipher)
			if err != nil {
				log.Printf("ERROR: unable to decrypt data discovery access key, error %v", err)
				return s3, nil

			}
			apwd, err := cryptoutils.HexDecrypt(rec.AccessSecret, srvConfig.Config.Encryption.Secret, srvConfig.Config.Encryption.Cipher)
			if err != nil {
				log.Printf("ERROR: unable to decrypt data discovery acess secret, error %v", err)
				return s3, nil

			}
			s3 = S3{
				Endpoint:     rec.Endpoint,
				AccessKey:    string(akey),
				AccessSecret: string(apwd),
				UseSSL:       rec.UseSSL,
			}
			return s3, nil
		}
	}
	return s3, errors.New("No matching site found")
}

// siteContent provides content on given site
func siteContent(site string) (SiteObject, error) {
	var siteObj SiteObject
	s3, err := site2s3(site)
	if err != nil {
		return siteObj, err
	}
	if srvConfig.Config.DataManagement.WebServer.Verbose > 0 {
		log.Printf("Use %v", s3)
	}
	log.Printf("INFO: found %s in DataDiscovery s3ords, will access its s3", site)
	if srvConfig.Config.DataManagement.WebServer.Verbose > 0 {
		log.Printf("INFO: accessing %+v", s3)
	}
	buckets, err := listBuckets(s3)
	if err != nil {
		log.Printf("ERROR: unabel to list buckets at %s, error %v", site, err)
	}
	obj := SiteObject{
		Site:    site,
		Buckets: buckets,
	}
	return obj, nil
}

// bucketContent provides content on given site and bucket
func bucketContent(site, bucket string) (BucketObject, error) {
	var bucketObj BucketObject
	s3, err := site2s3(site)
	if err != nil {
		return bucketObj, err
	}
	if srvConfig.Config.DataManagement.WebServer.Verbose > 0 {
		log.Printf("Use %v", s3)
	}
	if srvConfig.Config.DataManagement.WebServer.Verbose > 0 {
		log.Printf("looking for site:%s bucket:%s", site, bucket)
	}
	log.Printf("INFO: found %s in DataDiscovery records, will access its s3", site)
	if srvConfig.Config.DataManagement.WebServer.Verbose > 0 {
		log.Printf("INFO: accessing %+v", s3)
	}
	objects, err := listObjects(s3, bucket)
	if err != nil {
		log.Printf("ERROR: unabel to list objects at %s/%s, error %v", site, bucket, err)
	}
	obj := BucketObject{
		Site:    site,
		Bucket:  bucket,
		Objects: objects,
	}
	return obj, nil
}
