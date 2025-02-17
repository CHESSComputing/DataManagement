package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

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
