package handlers

import (
	"net/http"
	"scrapeit/internal/models"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ChangeScrapingGroupNotificationConfig(c echo.Context) error {
	var body models.NotificationConfig
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	groupId := c.Param("id")
	dbClient, ok := c.Get("db").(*mongo.Client)
	if !ok {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get database client"})
	}

	groupNotificationConfigCollection := dbClient.Database("scrapeit").Collection("notification_configs")

	groupNotificationConfigQuery := bson.M{"groupId": groupId}
	groupNotificationConfigUpdate := bson.M{
		"$set": bson.M{
			"groupId":          groupId,
			"fieldIdsToNotify": body.FieldIdsToNotify,
			"conditions":       body.Conditions,
		},
	}

	opts := options.Update().SetUpsert(true)
	result, err := groupNotificationConfigCollection.UpdateOne(
		c.Request().Context(),
		groupNotificationConfigQuery,
		groupNotificationConfigUpdate,
		opts,
	)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	if result.MatchedCount > 0 {
		return c.JSON(http.StatusOK, map[string]string{"message": "Notification config updated"})
	} else if result.UpsertedCount > 0 {
		return c.JSON(http.StatusCreated, map[string]string{"message": "Notification config created"})
	}

	return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Unexpected result"})
}
