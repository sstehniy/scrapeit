package handlers

import (
	"fmt"
	"net/http"

	"scrapeit/internal/scraper"

	"github.com/labstack/echo/v4"
)

func Home(c echo.Context) error {
	return c.String(http.StatusOK, "Welcome to the Golang Scraper API!")
}

func Scrape(c echo.Context) error {
	result, err := scraper.Scrape()
	if err != nil {
		fmt.Println("error scraping:", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, result)
}
