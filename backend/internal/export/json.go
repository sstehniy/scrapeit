package export

import (
	"encoding/json"
	"fmt"
	"scrapeit/internal/models"
)

func getJsonBytes(input []models.ExportScrapeResult) ([]byte, error) {
	output, err := json.Marshal(input)
	if err != nil {
		fmt.Println("JSON: failed to convert input to json")
		return nil, err
	}

	return output, err
}
