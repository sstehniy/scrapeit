package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"scrapeit/internal/models"

	"go.mongodb.org/mongo-driver/mongo"
)

func HandleNotifyResults(res *mongo.SingleResult, group models.ScrapeGroup, results []models.ScrapeResult, toReplace []models.ScrapeResult) {
	var relevantNotificationConfig models.NotificationConfig
	err := res.Decode(&relevantNotificationConfig)
	if err != nil {
		fmt.Println("Error decoding notification config:", err)
		return
	}

	type ScrapeResultWithStatus struct {
		Status string              `json:"status"`
		Result models.ScrapeResult `json:"result"`
	}

	allResults := []ScrapeResultWithStatus{}
	for _, r := range results {
		allResults = append(allResults, ScrapeResultWithStatus{
			Status: "new",
			Result: r,
		})
	}
	for _, r := range toReplace {
		allResults = append(allResults, ScrapeResultWithStatus{
			Status: "updated",
			Result: r,
		})
	}
	fmt.Println("All results:", len(allResults))
	resultsToNotify := []ScrapeResultWithStatus{}
	for _, result := range allResults {
		mustBeNotified := true
	OUTER:
		for conditionIdx, condition := range relevantNotificationConfig.Conditions {
			var foundValueByField interface{}
			for _, value := range result.Result.Fields {
				if value.FieldID == condition.FieldId {
					foundValueByField = value.Value
					foundSchemaField := group.GetFieldById(value.FieldID)
					relevantNotificationConfig.Conditions[conditionIdx].FieldName = foundSchemaField.Name
					break
				}
			}
			if _, ok := foundValueByField.(float64); ok {
				value := foundValueByField.(float64)
				switch condition.Operator {
				case "=":
					if value != condition.Value {
						mustBeNotified = false
						break OUTER
					}
				case "!=":
					if value == condition.Value {
						mustBeNotified = false
						break OUTER
					}
				case ">":
					if value <= condition.Value {
						mustBeNotified = false
						break OUTER
					}
				case "<":
					if value >= condition.Value {
						mustBeNotified = false
						break OUTER
					}
				}
			}
		}
		if mustBeNotified {
			resultsToNotify = append(resultsToNotify, result)
		}
	}

	if len(resultsToNotify) > 0 {
		requestBody := models.NotificationSearchResultRequestBody{
			GroupName: group.Name,
			Filters:   relevantNotificationConfig.Conditions,
			Results:   []models.NotificationResult{},
		}
		for _, result := range resultsToNotify {
			notificationResult := models.NotificationResult{
				Status:       result.Status,
				EndpointName: group.GetEndpointById(result.Result.EndpointID).Name,
				Fields:       []models.NotificationResultField{},
				URL:          "",
			}
			for _, field := range result.Result.Fields {
				shouldAddField := false
				for _, fieldIdToNotify := range relevantNotificationConfig.FieldIdsToNotify {
					if field.FieldID == fieldIdToNotify {
						shouldAddField = true
						break
					}
				}
				if shouldAddField {
					foundFieldName := ""
					for _, fieldConfig := range group.Fields {
						if fieldConfig.ID == field.FieldID {
							foundFieldName = fieldConfig.Name
							break
						}
					}
					notificationResult.Fields = append(notificationResult.Fields, models.NotificationResultField{
						FieldName: foundFieldName,
						Value:     field.Value,
					})
				}
			}
			foundUrlFieldId := ""
			for _, fieldConfig := range group.Fields {
				if fieldConfig.Type == models.FieldTypeLink {
					foundUrlFieldId = fieldConfig.ID
					break
				}

			}
			foundEndpoint := group.GetEndpointById(result.Result.EndpointID)
			if foundEndpoint != nil {
				urlValue := ""
				for _, field := range result.Result.Fields {
					if field.FieldID == foundUrlFieldId {
						urlValue = field.Value.(string)
						break
					}
				}
				notificationResult.URL = GetFullUrl(foundEndpoint.URL, urlValue)
			}

			imageFieldId := ""
			for _, fieldConfig := range group.Fields {
				if fieldConfig.Type == models.FieldTypeImage {
					imageFieldId = fieldConfig.ID
					break
				}
			}
			imageValue := ""
			if imageFieldId != "" {
				for _, field := range result.Result.Fields {
					if field.FieldID == imageFieldId {
						imageValue = field.Value.(string)
						break
					}
				}
			}

			notificationResult.ImageUrl = imageValue
			requestBody.Results = append(requestBody.Results, notificationResult)
		}
		sendNotification(
			requestBody,
		)

	}
}

func sendNotification(
	requestBody models.NotificationSearchResultRequestBody,
) {
	// Send notification
	fmt.Println("Sending notification")
	marschaledBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Println("Error marshaling notification request body:", err)
		return
	}
	// os.WriteFile("notification_request.json", marschaledBody, 0644)
	client := &http.Client{}
	req, err := http.NewRequest("POST", "http://bot:5005/send-notification", bytes.NewBuffer(marschaledBody))
	if err != nil {
		fmt.Println("Error creating notification request:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending notification request:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Notification response Status:", resp.Status)
}
