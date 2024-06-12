package handlers

import (
	"net/http"

	"scrapeit/internal/models"

	"github.com/labstack/echo/v4"
)

var dummyScrapeGroup = []models.ScrapeGroup{
	{
		ID:        "1",
		Name:      "Example Group 1",
		URL:       "https://example1.com",
		Fields:    []models.Field{},
		Endpoints: []models.Endpoint{},
	},
	{
		ID:        "2",
		Name:      "Example Group 2",
		URL:       "https://example2.com",
		Fields:    []models.Field{},
		Endpoints: []models.Endpoint{},
	},
	{
		ID:        "3",
		Name:      "Example Group 3",
		URL:       "https://example3.com",
		Fields:    []models.Field{},
		Endpoints: []models.Endpoint{},
	},
	{
		ID:        "4",
		Name:      "Example Group 4",
		URL:       "https://example4.com",
		Fields:    []models.Field{},
		Endpoints: []models.Endpoint{},
	},
	{
		ID:        "5",
		Name:      "Example Group 5",
		URL:       "https://example5.com",
		Fields:    []models.Field{},
		Endpoints: []models.Endpoint{},
	},
}

func GetScrapingGroups(c echo.Context) error {

	return c.JSON(http.StatusOK, dummyScrapeGroup)
}
