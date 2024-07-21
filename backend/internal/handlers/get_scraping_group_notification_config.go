package handlers

import (
	"context"
	"fmt"
	"net/http"
	"scrapeit/internal/models"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func getGroupNotificationConfigById(ctx context.Context, groupId string, client *mongo.Client) (models.NotificationConfig, error) {

	var groupNotificationConfig models.NotificationConfig
	groupCollection := client.Database("scrapeit").Collection("notification_configs")

	groupNotificationConfigQuery := bson.M{"groupId": groupId}
	groupNotificationConfigResult := groupCollection.FindOne(ctx, groupNotificationConfigQuery)
	if groupNotificationConfigResult.Err() == mongo.ErrNoDocuments {
		return models.NotificationConfig{}, nil
	}
	err := groupNotificationConfigResult.Decode(&groupNotificationConfig)
	if err != nil {
		return models.NotificationConfig{}, ScrapeGroupError{Message: "Failed to decode groupNotificationConfigQuery"}
	}

	return groupNotificationConfig, nil

}

func GetScrapingGroupNotificationConfig(c echo.Context) error {
	groupId := c.Param("id")
	dbClient, ok := c.Get("db").(*mongo.Client)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get database client")
	}
	groupNotificationConfigResult, err := getGroupNotificationConfigById(c.Request().Context(), groupId, dbClient)

	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
	}

	fmt.Println(groupNotificationConfigResult)

	if groupNotificationConfigResult.GroupId.IsZero() {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Group not found"})
	}

	return c.JSON(http.StatusOK, groupNotificationConfigResult)
}
