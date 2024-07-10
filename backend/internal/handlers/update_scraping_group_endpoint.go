package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func UpdateScrapingGroupEndpoint(c echo.Context) error {
	return c.String(http.StatusOK, "UpdateScrapingGroupEndpoint")
}
