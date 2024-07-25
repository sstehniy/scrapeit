package handlers

import (
	"net/http"
	"scrapeit/internal/models"
	"scrapeit/internal/scraper"

	"github.com/labstack/echo/v4"
)

type ElementSelectorTestHandlerRequest struct {
	Endpoint models.Endpoint `json:"endpoint"`
}

func ElementSelectorTestHandler(c echo.Context) error {
	var body ElementSelectorTestHandlerRequest
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body: " + err.Error()})
	}

	html, err := scraper.GetMainElementHTMLContent(body.Endpoint, 1)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"html": html})
}
