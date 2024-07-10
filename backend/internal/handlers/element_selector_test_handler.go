package handlers

import (
	"net/http"
	"scrapeit/internal/scraper"

	"github.com/labstack/echo/v4"
)

type ElementSelectorTestHandlerRequest struct {
	Url                 string `json:"url"`
	MainElementSelector string `json:"mainElementSelector"`
}

func ElementSelectorTestHandler(c echo.Context) error {
	var body ElementSelectorTestHandlerRequest
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	html, err := scraper.GetMainElementHTMLContent(body.Url, body.MainElementSelector, 1)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"html": html})
}
