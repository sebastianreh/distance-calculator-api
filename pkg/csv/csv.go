package csv

import (
	"bytes"
	"encoding/csv"
	"errors"
)

const (
	minRecords = 2
)

func CsvBytesToRecords(csvBytes []byte) ([][]string, error) {
	var records [][]string
	reader := csv.NewReader(bytes.NewReader(csvBytes))
	for {
		record, err := reader.Read()
		if err != nil {
			if err.Error() != "EOF" {
				return records, err
			}
			break
		}
		records = append(records, record)
	}
	if len(records) < minRecords {
		return records, errors.New("error reading csv bytes - records < 2")
	}
	return records, nil
}
