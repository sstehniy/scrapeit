package handlers

import (
	"net/http"
	"scrapeit/internal/ai"
	"scrapeit/internal/models"
	"scrapeit/internal/scraper"

	"github.com/labstack/echo/v4"
)

func ExtractSelectorsHandler(c echo.Context) error {
	requestData := models.FieldSelectorsRequest{}
	if err := c.Bind(&requestData); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	html, err := scraper.GetMainElementHTMLContent(requestData.URL, requestData.MainElementSelector, 2)
	// write html to a file for debugging
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	response, err := ai.ExtractSelectors(html, requestData.FieldsToExtractSelectorsFor)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, response)
}
