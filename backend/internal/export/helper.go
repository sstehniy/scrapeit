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
