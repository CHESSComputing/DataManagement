package main

import (
	"context"
	"io"
	"log"

	srvConfig "github.com/CHESSComputing/golib/config"
	minio "github.com/minio/minio-go/v7"
	credentials "github.com/minio/minio-go/v7/pkg/credentials"
)

// S3 represent S3 storage record
type S3 struct {
	Endpoint     string
	AccessKey    string
	AccessSecret string
	UseSSL       bool
}

// helper function to get s3 minio client for given site
func s3client(site string) (*minio.Client, error) {
	// get s3 site object without any buckets info
	s3, err := site2s3(site)
	if srvConfig.Config.DataManagement.WebServer.Verbose > 1 {
		log.Println("INFO: s3 object %+v", s3)
	}
	if err != nil {
		log.Printf("ERROR: unable to get S3 object for site %s, error %v", site, err)
		return nil, err
	}

	// Initialize minio client object.
	minioClient, err := minio.New(s3.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s3.AccessKey, s3.AccessSecret, ""),
		Secure: s3.UseSSL,
	})
	if err != nil {
		log.Printf("ERROR: unable to initialize s3 endpoint %s, error %v", s3.Endpoint, err)
	}
	return minioClient, err
}

// helper function to provide list of buckets in S3 store
func listBuckets(s3 S3) ([]minio.BucketInfo, error) {
	var out []minio.BucketInfo
	ctx := context.Background()
	// Initialize minio client object.
	minioClient, err := minio.New(s3.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s3.AccessKey, s3.AccessSecret, ""),
		Secure: s3.UseSSL,
	})
	if err != nil {
		log.Println("ERROR", err)
		return out, err
	}

	buckets, err := minioClient.ListBuckets(ctx)
	return buckets, err
}

// helper function to provide list of buckets in S3 store
func listObjects(s3 S3, bucket string) ([]minio.ObjectInfo, error) {
	var out []minio.ObjectInfo
	ctx := context.Background()
	// Initialize minio client object.
	minioClient, err := minio.New(s3.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s3.AccessKey, s3.AccessSecret, ""),
		Secure: s3.UseSSL,
	})
	if err != nil {
		log.Println("ERROR", err)
		return out, err
	}

	// list individual buckets
	objectCh := minioClient.ListObjects(ctx, bucket, minio.ListObjectsOptions{
		Recursive: true,
	})
	for object := range objectCh {
		if object.Err != nil {
			log.Println("ERROR: unable to list objects in a bucket, error %v", object.Err)
			return out, err
		}
		//         obj := fmt.Sprintf("%v %s %10d %s\n", object.LastModified, object.ETag, object.Size, object.Key)
		out = append(out, object)
	}
	return out, nil
}

// helper function to create new bucket in site's S3 store
func createBucket(site, bucket string) error {
	// get s3 site object without any buckets info
	minioClient, err := s3client(site)
	if err != nil {
		log.Printf("ERROR: unable to initialize minio client for site %s, error %v", site, err)
		return err
	}
	ctx := context.Background()

	// create new bucket on site s3 storage
	//     err = minioClient.MakeBucket(ctx, bucket, minio.MakeBucketOptions{Region: location})
	err = minioClient.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := minioClient.BucketExists(ctx, bucket)
		if errBucketExists == nil && exists {
			if srvConfig.Config.DataManagement.WebServer.Verbose > 0 {
				log.Printf("WARNING: we already own %s\n", bucket)
			}
			return nil
		} else {
			log.Printf("ERROR: unable to create bucket, error %v", err)
		}
	} else {
		if srvConfig.Config.DataManagement.WebServer.Verbose > 0 {
			log.Printf("Successfully created %s\n", bucket)
		}
	}
	return err
}
func deleteBucket(site, bucket string) error {
	minioClient, err := s3client(site)
	if err != nil {
		log.Printf("ERROR: unable to initialize minio client for site %s, error", site, err)
		return err
	}
	ctx := context.Background()
	err = minioClient.RemoveBucket(ctx, bucket)
	if err != nil {
		log.Printf("ERROR: unable to remove bucket %s, error, %v", bucket, err)
	}
	return err
}

// helper function to upload given object to site's S3 store
func uploadObject(site, bucket, objectName, contentType string, reader io.Reader, size int64) (minio.UploadInfo, error) {
	minioClient, err := s3client(site)
	if err != nil {
		log.Printf("ERROR: unable to initialize minio client for site %s, error", site, err)
		return minio.UploadInfo{}, err
	}
	ctx := context.Background()

	// Upload the zip file with PutObject
	options := minio.PutObjectOptions{}
	if contentType != "" {
		options = minio.PutObjectOptions{ContentType: contentType}
	}
	info, err := minioClient.PutObject(
		ctx,
		bucket,
		objectName,
		reader,
		size,
		options)
	if err != nil {
		log.Printf("ERROR: fail to upload file object, error %v", err)
	} else {
		if srvConfig.Config.DataManagement.WebServer.Verbose > 0 {
			log.Println("INFO: upload file", info)
		}
	}
	return info, err
}

// helper function to delete object from site S3 storage
func deleteObject(site, bucket, objectName, versionId string) error {
	minioClient, err := s3client(site)
	if err != nil {
		log.Printf("ERROR: unable to initialize minio client for site %s, error", site, err)
		return err
	}
	ctx := context.Background()

	// remove given object from our s3 store
	options := minio.RemoveObjectOptions{
		// Set the bypass governance header to delete an object locked with GOVERNANCE mode
		GovernanceBypass: true,
	}
	if versionId != "" {
		options.VersionID = versionId
	}
	err = minioClient.RemoveObject(
		ctx,
		bucket,
		objectName,
		options)
	if err != nil {
		log.Printf("ERROR: fail to delete file object, error %v", err)
	}
	return err
}

// helper function to get given object from site's S3 storage
func getObject(site, bucket, objectName string) ([]byte, error) {
	minioClient, err := s3client(site)
	if err != nil {
		log.Printf("ERROR: unable to initialize minio client for site %s, error", site, err)
		return []byte{}, err
	}
	ctx := context.Background()

	// Upload the zip file with PutObject
	options := minio.GetObjectOptions{}
	object, err := minioClient.GetObject(
		ctx,
		bucket,
		objectName,
		options)
	if err != nil {
		log.Printf("ERROR: fail to download file object, error %v", err)
	}
	data, err := io.ReadAll(object)
	return data, err
}
