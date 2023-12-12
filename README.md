# DataManagement Service
Data Management for CHESSComputing provides RESTfull access to site's S3 storage.
It supports the following (protected) APIs:
```
# create bucket (s3-bucket) at site Cornell
curl -v -X POST -H "Content-type: application/json" \
    -H "Authorization: Bearer $token" \
    http://localhost:8340/storage/Cornell/s3-bucket

# delete bucket
curl -v -H "Authorization: Bearer $token" \
    -X DELETE http://localhost:8340/storage/Cornell/s3-bucket

# upload file:
# take local file at /path/test.zip and upload it to
# S3 object: Cornell/s3-bucket/archive.zip
curl -v -H "Authorization: Bearer $token" \
    -H "content-type: multipart/form-data" \
    -X POST http://localhost:8340/storage/Cornell/s3-bucket/archive.zip \
    -F "file=@/path/test.zip"

# get file
curl http://localhost:8340/storage/Cornell/s3-bucket/archive.zip > archive.zip

# delete file
curl -v -H "Authorization: Bearer $token" \
    -X DELETE http://localhost:8340/storage/Cornell/s3-bucket/archive.zip
```
