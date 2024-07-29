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

func getGroupNotificationConfigById(ctx context.Context, groupId string, client *mongo.Client) ([]models.NotificationConfig, error) {

	groupCollection := client.Database("scrapeit").Collection("notification_configs")

	groupIdObj, err := primitive.ObjectIDFromHex(groupId)
	if err != nil {
		fmt.Println("Failed to convert groupId to ObjectId: ", err)
		return []models.NotificationConfig{}, err
	}

	groupNotificationConfigQuery := bson.M{"groupId": groupIdObj}
	groupNotificationConfigResult, err := groupCollection.Find(ctx, groupNotificationConfigQuery)

	notificationConfigs := []models.NotificationConfig{}

	if err != nil {
		return notificationConfigs, nil
	}

	defer groupNotificationConfigResult.Close(context.Background())

	for groupNotificationConfigResult.Next(ctx) {
		var notConfig models.NotificationConfig
		err := groupNotificationConfigResult.Decode(&notConfig)
		if err != nil {
			fmt.Println("error decoding group Not Config")
		} else {
			notificationConfigs = append(notificationConfigs, notConfig)
		}
	}

	return notificationConfigs, nil

}

func GetScrapingGroupNotificationConfig(c echo.Context) error {
	groupId := c.Param("id")
	dbClient, _ := models.GetDbClient()
	groupNotificationConfigResult, err := getGroupNotificationConfigById(c.Request().Context(), groupId, dbClient)

	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
	}

	fmt.Println(groupNotificationConfigResult)

	return c.JSON(http.StatusOK, groupNotificationConfigResult)
}
