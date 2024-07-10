package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func CreateScrapingGroupEndpoint(c echo.Context) error {
	return c.String(http.StatusOK, "CreateScrapingGroupEndpoint")
}
