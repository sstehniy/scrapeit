package export

import (
	"fmt"
	"scrapeit/internal/models"

	"github.com/xuri/excelize/v2"
)

func getExcelBytes(input []models.ExportScrapeResult) ([]byte, error) {
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
	rows := [][]interface{}{}
	for _, record := range records {
		row := []interface{}{}
		for _, key := range headers {
			row = append(row, fmt.Sprintf("%v", record[key]))
		}
		rows = append(rows, row)
	}
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	// Create a new sheet.
	_, err := f.NewSheet("Sheet1")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	for i, header := range headers {
		cell, err := excelize.CoordinatesToCellName(i+1, 1)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		fmt.Println("Cell name", cell)
		f.SetCellValue("Sheet1", cell, header)
	}
	for i, row := range rows {
		for j, value := range row {
			cell, err := excelize.CoordinatesToCellName(j+1, i+2)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			f.SetCellValue("Sheet1", cell, value)
		}
	}

	buffer, err := f.WriteToBuffer()
	if err != nil {
	}

	defer f.Close()

	return buffer.Bytes(), nil

}
