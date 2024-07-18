package export

import (
	"encoding/xml"
	"fmt"
	"scrapeit/internal/models"
)

type ExportScrapeResults struct {
	XMLName xml.Name                    `xml:"ScrapeResults"`
	Results []models.ExportScrapeResult `xml:"ScrapeResult"`
}

func getXmlBytes(input []models.ExportScrapeResult) ([]byte, error) {
	wrapped := ExportScrapeResults{Results: input}
	output, err := xml.MarshalIndent(wrapped, "", "  ")
	if err != nil {
		fmt.Println("XML: failed to convert input to xml")
		return nil, err
	}
	xmlHeader := []byte(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	output = append(xmlHeader, output...)
	return output, err
}
