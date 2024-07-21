package handlers

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func DeleteScrapingGroupNotificationConfigByGroupId(ctx context.Context, groupId string, client *mongo.Client) error {

	groupNotificationConfigCollection := client.Database("scrapeit").Collection("notification_configs")

	groupNotificationConfigQuery := bson.M{"groupId": groupId}

	_, err := groupNotificationConfigCollection.DeleteOne(ctx, groupNotificationConfigQuery)
	if err != nil {
		return err
	}

	return nil
}

func DeleteScrapingGroupNotificationConfig(c echo.Context) error {
	groupId := c.Param("id")
	dbClient, ok := c.Get("db").(*mongo.Client)
	if !ok {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get database client"})

	}

	err := DeleteScrapingGroupNotificationConfigByGroupId(c.Request().Context(), groupId, dbClient)

	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Notification config deleted"})
}
