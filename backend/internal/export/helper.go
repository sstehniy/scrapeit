package export

import (
	"fmt"
	"scrapeit/internal/models"
)

func GetFileExtension(exportType models.ExportType) (string, error) {
	switch exportType {
	case models.ExportTypeXML:
		return ".xml", nil
	case models.ExportTypeCSV:
		return ".csv", nil
	case models.ExportTypeEXCEL:
		return ".xlsx", nil
	case models.ExportTypeJSON:
		return ".json", nil
	case models.ExportTypePDF:
		return ".pdf", nil
	default:
		return "", fmt.Errorf("no matched format found for input %s", exportType)
	}

}

func inputToRecords(input []models.ExportScrapeResult) ([]map[string]interface{}, []string) {
	records := []map[string]interface{}{}
	valueHeaderKeys := []string{}
	for _, scrapeResult := range input {
		record := map[string]interface{}{}
		record["id"] = scrapeResult.ID
		record["endpointName"] = scrapeResult.EndpointName
		record["endpointId"] = scrapeResult.EndpointID
		record["groupName"] = scrapeResult.GroupName
		record["groupId"] = scrapeResult.GroupId
		record["timestampInitial"] = scrapeResult.TimestampInitial
		record["timestampLastUpdate"] = scrapeResult.TimestampLastUpdate
		record["groupVersionTag"] = scrapeResult.GroupVersionTag

		for _, value := range scrapeResult.Fields {
			record[value.FieldName] = value.Value
			if !contains(valueHeaderKeys, value.FieldName) {
				valueHeaderKeys = append(valueHeaderKeys, value.FieldName)
			}
		}
		records = append(records, record)
	}
	return records, valueHeaderKeys
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
