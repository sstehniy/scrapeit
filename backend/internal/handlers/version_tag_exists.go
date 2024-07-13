package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func VersionTagExists(c echo.Context) error {
	versionTag := c.Param("versionTag")

	dbClient, ok := c.Get("db").(*mongo.Client)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get database client")
	}

	groupCollection := dbClient.Database("scrapeit").Collection("scrape_groups")
	groupResult := groupCollection.FindOne(c.Request().Context(), bson.M{"versionTag": versionTag})

	if groupResult.Err() == mongo.ErrNoDocuments {
		return c.JSON(http.StatusOK, map[string]bool{"exists": false})
	}

	if groupResult.Err() != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": groupResult.Err().Error()})
	}

	return c.JSON(http.StatusOK, map[string]bool{"exists": true})
}
