package export

import (
	"fmt"
	"scrapeit/internal/models"
)

func CreateResultsExportFile(inputResults []models.ScrapeResult, group models.ScrapeGroup, fileType models.ExportType) ([]byte, error) {
	fmt.Printf("first input %v", inputResults[0])
	transformed := transformResultsForExport(inputResults, group)
	fmt.Printf("first val %v", transformed[0])
	switch fileType {
	case models.ExportTypeXML:
		return getXmlBytes(transformed)
	}

	return nil, nil

}

func transformResultsForExport(inputResults []models.ScrapeResult, group models.ScrapeGroup) []models.ExportScrapeResult {
	data := make([]models.ExportScrapeResult, len(inputResults))
	for idx, res := range inputResults {

		exportDetails := make([]models.ExportScrapeResultDetail, len(res.Fields))

		for didx, df := range res.Fields {
			exportDetails[didx] =
				models.ExportScrapeResultDetail{
					ID:        df.ID,
					FieldName: getFieldNameById(&group.Fields, df.FieldID),
					FieldID:   df.FieldID,
					Value:     df.Value,
				}

		}
		data[idx] = models.ExportScrapeResult{
			ID:                  res.ID.Hex(),
			EndpointName:        getEndpointNameById(&group.Endpoints, res.EndpointID),
			EndpointID:          res.EndpointID,
			GroupName:           group.Name,
			GroupId:             group.ID.Hex(),
			Fields:              exportDetails,
			TimestampInitial:    res.TimestampInitial,
			TimestampLastUpdate: res.TimestampLastUpdate,
			GroupVersionTag:     res.GroupVersionTag,
		}
	}

	return data
}

func getFieldNameById(fields *[]models.Field, fieldId string) string {

	for _, f := range *fields {
		if f.ID == fieldId {
			return f.Name
		}
	}
	return ""
}

func getEndpointNameById(eps *[]models.Endpoint, epId string) string {
	for _, ep := range *eps {
		if ep.ID == epId {
			return ep.Name
		}
	}
	return ""
}
