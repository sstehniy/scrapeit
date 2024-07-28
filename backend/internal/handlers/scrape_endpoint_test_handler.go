package handlers

import (
	"net/http"
	"scrapeit/internal/models"
	"scrapeit/internal/scraper"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ScrapeEndpointTestHandlerRequest struct {
	Group TestScrapeGroup `json:"group"`
}

type TestScrapeGroup struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Fields        []models.Field    `json:"fields"`
	Endpoints     []models.Endpoint `json:"endpoints"`
	WithThumbnail bool              `json:"withThumbnail"`
}

func ScrapeEndpointTestHandler(c echo.Context) error {
	var body ScrapeEndpointTestHandlerRequest
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	dbClient, ok := c.Get("db").(*mongo.Client)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get database client")
	}

	group := models.ScrapeGroup{
		ID:            primitive.NewObjectID(),
		Name:          body.Group.Name,
		Fields:        body.Group.Fields,
		Endpoints:     body.Group.Endpoints,
		WithThumbnail: body.Group.WithThumbnail,
	}

	browser := scraper.GetBrowser()

	results, _, err := scraper.ScrapeEndpointTest(body.Group.Endpoints[0], group, dbClient, browser)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, results)
}
