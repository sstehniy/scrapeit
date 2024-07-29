package handlers

import (
	"context"
	"fmt"
	"net/http"
	"scrapeit/internal/cron"

	"github.com/labstack/echo/v4"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func deleteScrapingGroupEndpointResults(dbClient *mongo.Client, groupId, endpointId string) error {
	// print params
	fmt.Printf("groupId: %v\n", groupId)
	fmt.Printf("endpointId: %v\n", endpointId)
	// remove all scrape results for this endpoint
	scrapeResultsCollection := dbClient.Database("scrapeit").Collection("scrape_results")
	groupIdObj, err := primitive.ObjectIDFromHex(groupId)
	if err != nil {
		return err
	}
	result, err := scrapeResultsCollection.DeleteMany(context.TODO(), bson.M{"endpointId": endpointId, "groupId": groupIdObj})

	fmt.Printf("Delete count: %v\n", result.DeletedCount)

	return err
}

func DeleteScrapingGroupEndpointResults(c echo.Context) error {
	dbClient, ok := c.Get("db").(*mongo.Client)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get database client")
	}
	cronManager, ok := c.Get("cron").(*cron.CronManager)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get cron manager")
	}

	groupId := c.Param("groupId")
	if groupId == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Missing parameter 'groupId'",
		})
	}
	endpointId := c.Param("endpointId")
	if endpointId == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Missing parameter 'endpointId'",
		})
	}

	go cronManager.DestroyJob(groupId, endpointId)

	err := deleteScrapingGroupEndpointResults(dbClient, groupId, endpointId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Endpoint Results deleted successfully"})
}
