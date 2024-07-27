package handlers

import (
	"fmt"
	"net/http"
	"scrapeit/internal/models"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ChangeScrapingGroupNotificationConfigRequest []models.NotificationConfig

func ChangeScrapingGroupNotificationConfig(c echo.Context) error {
	var configs ChangeScrapingGroupNotificationConfigRequest
	if err := c.Bind(&configs); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	groupId := c.Param("id")
	dbClient, ok := c.Get("db").(*mongo.Client)
	if !ok {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get database client"})
	}

	groupNotificationConfigCollection := dbClient.Database("scrapeit").Collection("notification_configs")

	groupIdObj, err := primitive.ObjectIDFromHex(groupId)

	if err != nil {
		fmt.Println("failed to parse groupId")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Delete existing configurations for the group
	_, err = groupNotificationConfigCollection.DeleteMany(
		c.Request().Context(),
		bson.M{"groupId": groupIdObj},
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete existing configs"})
	}

	// Prepare new configurations to insert
	var docs []interface{}
	for _, config := range configs {
		config.ID = primitive.NewObjectID()
		docs = append(docs, config)
	}

	if len(docs) == 0 {
		return c.JSON(http.StatusOK, map[string]string{"message": "Configs deleted successfully"})
	}

	// Insert new configurations
	_, err = groupNotificationConfigCollection.InsertMany(
		c.Request().Context(),
		docs,
	)
	if err != nil {
		fmt.Printf("Failed to insert notification configs\n")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Configs updated successfully"})
}
