package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Metadata represents file metadata information
type Metadata struct {
	Name        string    `json:"name"`
	Size        int64     `json:"size"`
	ModTime     time.Time `json:"mod_time"`
	IsDirectory bool      `json:"is_directory"`
}

// FsClient represents generic interface to communicate with FileSystem instances
type FsClient interface {
	Get(dir, file string) ([]byte, error)
	List(dir string) ([]Metadata, error)
	Create(dir string) error
	Upload(dir, file, ctype string, reader io.Reader, size int64) error
	Delete(dir, file string) error
}

// LocalFsClient provides local file system implementation of FsClient
type LocalFsClient struct {
	Storage string
	Logger  *log.Logger
}

// NewLocalFsClient creates a new LocalFsClient with logging enabled
func NewLocalFsClient(storage string) *LocalFsClient {
	return &LocalFsClient{
		Storage: storage,
		Logger:  log.New(os.Stdout, "FsClient: ", log.LstdFlags),
	}
}

// Get retrieves a file's content or lists directory contents if file is empty
func (l *LocalFsClient) Get(dir, file string) ([]byte, error) {
	path := filepath.Join(l.Storage, dir)

	// If file is empty, return directory metadata
	if file == "" {
		files, err := l.List(dir)
		if err != nil {
			l.Logger.Printf("Error listing directory %s: %v", dir, err)
			return nil, err
		}
		data, _ := json.MarshalIndent(files, "", "  ")
		return data, nil
	}

	// Otherwise, read the file
	path = filepath.Join(path, file)
	data, err := os.ReadFile(path)
	if err != nil {
		l.Logger.Printf("Error reading file %s: %v", path, err)
		return nil, err
	}
	l.Logger.Printf("Read file %s successfully", path)
	return data, nil
}

// List retrieves metadata for all files in a given directory
func (l *LocalFsClient) List(dir string) ([]Metadata, error) {
	path := filepath.Join(l.Storage, dir)
	files, err := ioutil.ReadDir(path)
	if err != nil {
		l.Logger.Printf("Failed to list directory %s: %v", path, err)
		return nil, err
	}

	var metadataList []Metadata
	for _, file := range files {
		metadataList = append(metadataList, Metadata{
			Name:        file.Name(),
			Size:        file.Size(),
			ModTime:     file.ModTime(),
			IsDirectory: file.IsDir(),
		})
	}
	l.Logger.Printf("Listed directory %s", path)
	return metadataList, nil
}

// Create creates a new directory
func (l *LocalFsClient) Create(dir string) error {
	path := filepath.Join(l.Storage, dir)
	err := os.MkdirAll(path, os.ModePerm)
	if err == nil {
		l.Logger.Printf("Created directory %s", path)
	} else {
		l.Logger.Printf("Failed to create directory %s: %v", path, err)
	}
	return err
}

// Upload writes data to a file in chunks to handle large files efficiently
func (l *LocalFsClient) Upload(dir, file, ctype string, reader io.Reader, size int64) error {
	path := filepath.Join(l.Storage, dir, file)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		l.Logger.Printf("Failed to create directory for file %s: %v", path, err)
		return err
	}

	// Open the file for writing
	outFile, err := os.Create(path)
	if err != nil {
		l.Logger.Printf("Failed to create file %s: %v", path, err)
		return err
	}
	defer outFile.Close()

	// Copy data from reader to file using buffer
	buffer := make([]byte, 4096) // 4KB buffer
	_, err = io.CopyBuffer(outFile, reader, buffer)
	if err == nil {
		l.Logger.Printf("Uploaded file %s successfully", path)
	} else {
		l.Logger.Printf("Failed to upload file %s: %v", path, err)
	}
	return err
}

// Delete removes a file or an entire directory if file is empty
func (l *LocalFsClient) Delete(dir, file string) error {
	path := filepath.Join(l.Storage, dir)

	// If file is empty, delete the entire directory
	if file == "" {
		err := os.RemoveAll(path)
		if err == nil {
			l.Logger.Printf("Deleted directory %s", path)
		} else {
			l.Logger.Printf("Failed to delete directory %s: %v", path, err)
		}
		return err
	}

	// Otherwise, delete the specific file
	path = filepath.Join(path, file)
	err := os.Remove(path)
	if err == nil {
		l.Logger.Printf("Deleted file %s", path)
	} else {
		l.Logger.Printf("Failed to delete file %s: %v", path, err)
	}
	return err
}

/*
// Example usage
func main() {
	client := NewLocalFsClient("./data")

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

	// Get directory metadata
	data, err := client.Get("testdir", "")
	if err != nil {
		fmt.Println("Error getting directory metadata:", err)
	} else {
		fmt.Println(string(data))
	}

	// Delete directory
	err = client.Delete("testdir", "")
	if err != nil {
		fmt.Println("Error deleting directory:", err)
	}
}
*/
