package main

import "io"

// FsClient represents generic interface to communicate with FileSystem instances
type FsClient interface {
	Get(dir, file string) ([]byte, error)
	List(dir string) ([]byte, error)
	Create(dir string) error
	Upload(dir, file, ctype string, reader io.Reader, size int64) error
	Delete(dir, file string) error
}

type LocalFsClient struct {
	Storage string
}

func (l *LocalFsClient) Get(dir, file string) ([]byte, error) {
	var data []byte
	return data, nil
}
func (l *LocalFsClient) List(dir string) ([]byte, error) {
	var data []byte
	return data, nil
}
func (l *LocalFsClient) Create(dir string) error {
	return nil
}
func (l *LocalFsClient) Upload(dir, file, ctype string, reader io.Reader, size int64) error {
	return nil
}
func (l *LocalFsClient) Delete(dir, file string) error {
	return nil
}
