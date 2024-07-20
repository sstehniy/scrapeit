package export

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"scrapeit/internal/models"
)

func getCsvBytes(input []models.ExportScrapeResult) ([]byte, error) {
	records, valueHeaderKeys := inputToRecords(input)
	headers := []string{
		"id",
		"endpointName",
		"endpointId",
		"groupName",
		"groupId",
		"timestampInitial",
		"timestampLastUpdate",
		"groupVersionTag",
	}
	headers = append(headers, valueHeaderKeys...)
	buffer := &bytes.Buffer{}
	writer := csv.NewWriter(buffer)
	err := writer.Write(headers)
	if err != nil {
		return nil, err
	}
	for _, record := range records {
		row := make([]string, len(headers))
		for i, header := range headers {
			if value, ok := record[header]; ok {
				row[i] = fmt.Sprintf("%v", value)
			} else {
				row[i] = ""
			}
		}
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}

	writer.Flush()

	if err := writer.Error(); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil

}
