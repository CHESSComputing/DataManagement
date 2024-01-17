# DataManagement Service

![build status](https://github.com/CHESSComputing/DataManagement/actions/workflows/go.yml/badge.svg)
[![go report card](https://goreportcard.com/badge/github.com/CHESSComputing/DataManagement)](https://goreportcard.com/report/github.com/CHESSComputing/DataManagement)
[![godoc](https://godoc.org/github.com/CHESSComputing/DataManagement?status.svg)](https://godoc.org/github.com/CHESSComputing/DataManagement)

Data Management for CHESSComputing provides RESTfull access to site's S3 storage.

### Prerequisites
- all operations requires token with appropriate scope, e.g. to get data we can
  use read scope token, for posting (writing) the data we use token with write
  scope, and for deletion token with delete scope
- in order to access DataManagement service the MIN.IO server sohuld be up and runnig
  - the MIN.IO server should be configured with proper user account(s)
  - the access token and access secret should be created and placed under DataManagement.S3 section into foxden  configuration, e.g.
```
DataManagement:
  S3:
    AccessKey: <access_key>
    AccessSecret: <access_secret>
    Endpoint: localhost:8330
    UseSSL: false
```
For further details of managing users in MIN.IO please consult 
[User Management](https://min.io/docs/minio/linux/administration/identity-access-management/minio-user-management.html)
documentation.

### DataManagement service APIs
DataManagement service supports the following (protected) APIs:
```
# list all buckets in current storage
curl -v -H "Authorization: Bearer $token" \
    http://localhost:8340/storage

# create bucket (s3-bucket)
curl -v -X POST -H "Content-type: application/json" \
    -H "Authorization: Bearer $token" \
    http://localhost:8340/storage/s3-bucket

# delete bucket
curl -v -H "Authorization: Bearer $token" \
    -X DELETE http://localhost:8340/storage/s3-bucket

# upload file:
# take local file at /path/test.zip and upload it to
# S3 object: s3-bucket/archive.zip
curl -v -H "Authorization: Bearer $token" \
    -H "content-type: multipart/form-data" \
    -X POST http://localhost:8340/storage/s3-bucket/archive.zip \
    -F "file=@/path/test.zip"

# get file
curl -H "Authorization: Bearer $token" \
    http://localhost:8340/storage/s3-bucket/archive.zip > archive.zip

# delete file
curl -v -H "Authorization: Bearer $token" \
    -X DELETE http://localhost:8340/storage/s3-bucket/archive.zip
```
