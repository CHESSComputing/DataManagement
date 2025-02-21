package main

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	srvConfig "github.com/CHESSComputing/golib/config"
	services "github.com/CHESSComputing/golib/services"
)

// FileEntry represents a directory entry
type FileEntry struct {
	Did    string `json:"did"`
	EscDid string `json:"esc_did"`
	Name   string `json:"name"`
	IsDir  bool   `json:"is_dir"`
	Path   string `json:"path"` // path here correspond to sub-path within raw location area
}

// getFileList returns a list of files and directories in the given path
func getFileList(did, path, spath string) ([]FileEntry, error) {
	var entries []FileEntry

	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		entry := FileEntry{
			Did:    did,
			EscDid: url.QueryEscape(did),
			Name:   file.Name(),
			IsDir:  file.IsDir(),
			Path:   filepath.Join(spath, file.Name()),
			//             Path:   filepath.Join(path, file.Name()),
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// helper function to find meta-data record for given did
func findMetaDataRecord(did string) (map[string]any, error) {
	var rec map[string]any
	query := fmt.Sprintf("{\"did\":\"%s\"}", did)
	var skeys []string
	var sorder, idx int
	limit := 1
	records, err := services.MetaDataRecords(query, skeys, sorder, idx, limit)
	if err != nil {
		return rec, err
	}
	if len(records) != 1 {
		msg := fmt.Sprintf("multiple records found for did=%s, records=%v", did, records)
		return rec, errors.New(msg)
	}
	return records[0], nil
}

// findFiles recursively finds all files in idir matching the given pattern pat.
func findFiles(idir string, pat string) ([]string, error) {
	if !strings.HasSuffix(idir, "/") {
		idir += "/"
	}
	var files []string

	// Compile the regex pattern
	re, err := regexp.Compile(pat)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %v", err)
	}

	// Walk through the directory
	err = filepath.Walk(idir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println("WARNING:", err)
		}

		if pat == "all" {
			// get all files
			if !info.IsDir() {
				files = append(files, path)
			}
		} else {
			// Check if it's a regular file and matches the pattern
			if !info.IsDir() && re.MatchString(info.Name()) {
				files = append(files, path)
			}
		}
		return nil
	})

	return files, err
}

// fileExtensions finds all unique file extensions in the given directory and subdirectories.
func fileExtensions(idir string) []string {
	if !strings.HasSuffix(idir, "/") {
		idir += "/"
	}
	extMap := make(map[string]bool)

	// Walk through the directory
	err := filepath.Walk(idir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println("WARNING: filepath.Walk", err.Error())
		}

		// Check if it's a file
		if !info.IsDir() {
			ext := filepath.Ext(info.Name()) // Extract file extension
			if ext != "" {
				extMap[ext] = true // Store unique extensions
			}
		}
		return nil
	})

	if err != nil {
		log.Println("WARNING: filepath.Walk", err.Error())
		// return default list of file extensions
		if len(srvConfig.Config.DataManagement.FileExtensions) > 0 {
			return srvConfig.Config.DataManagement.FileExtensions
		}
		return []string{"png", "jpg", "tiff", "mcs"}
	}

	// Convert map keys to a slice
	var extensions []string
	for ext := range extMap {
		extensions = append(extensions, ext)
	}

	return extensions
}
