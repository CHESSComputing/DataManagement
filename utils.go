package main

import (
	"errors"
	"fmt"

	services "github.com/CHESSComputing/golib/services"
)

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
