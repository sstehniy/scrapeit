package handlers

import (
	"context"
	"fmt"
	"net/http"
	"scrapeit/internal/models"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ScrapeGroupError struct {
	Message string
}

func (e ScrapeGroupError) Error() string {
	return fmt.Sprintf("scrape group error: %s", e.Message)
}

func getGroupById(ctx context.Context, groupId string, client *mongo.Client) (models.ScrapeGroup, error) {

	var group models.ScrapeGroup
	groupCollection := client.Database("scrapeit").Collection("scrape_groups")
	groupIdPrimitive, err := primitive.ObjectIDFromHex(groupId)
	if err != nil {
		return models.ScrapeGroup{}, ScrapeGroupError{Message: "Invalid group ID"}

	}
	groupQuery := bson.M{"_id": groupIdPrimitive}
	groupResult := groupCollection.FindOne(ctx, groupQuery)
	if groupResult.Err() != nil {
		return models.ScrapeGroup{}, ScrapeGroupError{Message: "Group not found"}
	}
	err = groupResult.Decode(&group)
	if err != nil {
		return models.ScrapeGroup{}, ScrapeGroupError{Message: "Failed to decode group"}
	}

	return group, nil

}

func GetScrapingGroup(c echo.Context) error {
	groupId := c.Param("id")
	dbClient, ok := c.Get("db").(*mongo.Client)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get database client")
	}
	group, err := getGroupById(c.Request().Context(), groupId, dbClient)

	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, group)
}
