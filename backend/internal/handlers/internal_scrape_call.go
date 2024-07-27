package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"scrapeit/internal/cron"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
)

func HandleCallInternalScrapeEndpoint(e *echo.Echo, groupId, endpointId string, db *mongo.Client, cronmanager *cron.CronManager) error {
	// call the internal scrape endpoint
	body := map[string]interface{}{
		"groupId":    groupId,
		"endpointId": endpointId,
		"internal":   true,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		fmt.Println("Error marshaling body:", err)
		return err
	}

	req, err := http.NewRequest(http.MethodPost, "http://localhost:3457/internal/scrape/endpoint", bytes.NewBuffer(bodyBytes))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return err
	}

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	customContext := e.NewContext(req, rec)
	customContext.Set("db", db)
	customContext.Set("cron", cronmanager)
	if err := ScrapeEndpointHandler(customContext); err != nil {
		fmt.Println("Error handling request in internal scrape endpoint:", err)
		return err
	}
	return nil
}
