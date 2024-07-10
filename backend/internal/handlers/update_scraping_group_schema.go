package handlers

import (
	"net/http"
	"scrapeit/internal/models"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UpdateScrapingGroupSchemaRequest struct {
	GroupId string         `json:"groupId"`
	Schema  []models.Field `json:"fields"`
}

func UpdateScrapingGroupSchema(c echo.Context) error {
	dbClient, ok := c.Get("db").(*mongo.Client)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get database client")
	}

	var req UpdateScrapingGroupSchemaRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	groupId, err := primitive.ObjectIDFromHex(req.GroupId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid group ID")
	}

	result, err := dbClient.Database("scrapeit").Collection("scrape_groups").UpdateOne(
		c.Request().Context(),
		bson.M{"_id": groupId},
		bson.M{"$set": bson.M{"fields": req.Schema}},
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update group schema")
	}

	if result.MatchedCount == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "Group not found")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Group schema updated"})
}
