package handlers

import (
	"fmt"
	"net/http"
	"scrapeit/internal/ai"
	"scrapeit/internal/models"
	"scrapeit/internal/scraper"

	"github.com/labstack/echo/v4"
)

type ExtractSelectorsResponse struct {
	Fields    []models.FieldSelectorsResponse `json:"fields"`
	TotalCost float32                         `json:"totalCost"`
}

func ExtractSelectorsHandler(c echo.Context) error {
	requestData := models.FieldSelectorsRequest{}
	if err := c.Bind(&requestData); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	scrapeType := scraper.GetScrapeType(requestData.Endpoint)
	maxElements := 4
	if scrapeType != scraper.Previews {
		maxElements = 1
	}
	html, err := scraper.GetMainElementHTMLContent(requestData.Endpoint, maxElements)
	// write html to a file for debugging
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	response, totalCost, err := ai.ExtractSelectors(html, requestData.FieldsToExtractSelectorsFor)
	fmt.Println("Total cost: ", totalCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, ExtractSelectorsResponse{
		Fields:    response.Fields,
		TotalCost: totalCost,
	})
}
