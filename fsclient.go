package main

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// FsClient represents generic interface to communicate with FileSystem instances
type FsClient interface {
	Get(dir, file string) ([]byte, error)
	List(dir string) ([]string, error)
	Create(dir string) error
	Upload(dir, file, ctype string, reader io.Reader, size int64) error
	Delete(dir, file string) error
}

// LocalFsClient provides local file system implementation of FsClient
type LocalFsClient struct {
	Storage string
}

// Get retrieves a file from the given directory
func (l *LocalFsClient) Get(dir, file string) ([]byte, error) {
	path := filepath.Join(l.Storage, dir, file)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// List retrieves the list of files in the given directory
func (l *LocalFsClient) List(dir string) ([]string, error) {
	path := filepath.Join(l.Storage, dir)
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var fileNames []string
	for _, file := range files {
		if !file.IsDir() {
			fileNames = append(fileNames, file.Name())
		}
	}
	return fileNames, nil
}

// Create creates a new directory
func (l *LocalFsClient) Create(dir string) error {
	path := filepath.Join(l.Storage, dir)
	return os.MkdirAll(path, os.ModePerm)
}

// Upload writes data to a file in chunks to handle large files efficiently
func (l *LocalFsClient) Upload(dir, file, ctype string, reader io.Reader, size int64) error {
	path := filepath.Join(l.Storage, dir, file)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return err
	}

	// Open the file for writing
	outFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Copy data from reader to file using buffer
	buffer := make([]byte, 4096) // 4KB buffer
	_, err = io.CopyBuffer(outFile, reader, buffer)
	return err
}

// Delete removes a file from the given directory
func (l *LocalFsClient) Delete(dir, file string) error {
	path := filepath.Join(l.Storage, dir, file)
	return os.Remove(path)
}

/*
// Example usage
func main() {
	client := &LocalFsClient{Storage: "./data"}

	// Create directory
	err := client.Create("testdir")
	if err != nil {
		fmt.Println("Error creating directory:", err)
	}

	// Upload file
	content := bytes.NewReader([]byte("Hello, World!"))
	err = client.Upload("testdir", "hello.txt", "text/plain", content, int64(content.Len()))
	if err != nil {
		fmt.Println("Error uploading file:", err)
	}

	// List files
	files, err := client.List("testdir")
	if err != nil {
		fmt.Println("Error listing files:", err)
	} else {
		fmt.Println("Files:", files)
	}

	// Get file
	data, err := client.Get("testdir", "hello.txt")
	if err != nil {
		fmt.Println("Error getting file:", err)
	} else {
		fmt.Println("File content:", string(data))
	}

	// Delete file
	err = client.Delete("testdir", "hello.txt")
	if err != nil {
		fmt.Println("Error deleting file:", err)
	}
}
*/
