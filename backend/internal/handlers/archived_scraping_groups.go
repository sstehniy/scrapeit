package handlers

import (
	"fmt"
	"net/http"
	"scrapeit/internal/models"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetArchivedScrapingGroups(c echo.Context) error {
	dbClient, ok := c.Get("db").(*mongo.Client)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get database client")
	}

	allGroups := []models.ArchivedScrapeGroup{}

	result, err := dbClient.Database("scrapeit").Collection("scrape_groups").Find(c.Request().Context(), bson.M{
		// not equal to empty string
		"versionTag": bson.M{"$ne": ""},
	})
	if err != nil {
		fmt.Println("error in finding groups")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

	}
	defer result.Close(c.Request().Context())
	err = result.All(c.Request().Context(), &allGroups)
	if err != nil {
		fmt.Println("error in getting all archived groups")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, allGroups)
}
