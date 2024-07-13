package handlers

import (
	"context"
	"net/http"
	"scrapeit/internal/models"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func deleteScrapingGroupEndpoint(dbClient *mongo.Client, groupId, endpointId string) error {
	group := models.ScrapeGroup{}
	groupIdObj, err := primitive.ObjectIDFromHex(groupId)
	if err != nil {
		return err
	}
	groupCollection := dbClient.Database("scrapeit").Collection("scrape_groups")
	groupResult := groupCollection.FindOne(context.TODO(), bson.M{"_id": groupIdObj})
	if groupResult.Err() != nil {
		return groupResult.Err()
	}
	err = groupResult.Decode(&group)
	if err != nil {
		return err
	}

	group.DeleteEndpoint(endpointId)

	_, err = groupCollection.UpdateOne(context.TODO(), bson.M{"_id": groupIdObj}, bson.M{"$set": group})

	if err != nil {
		return err
	}

	// remove all scrape results for this endpoint
	scrapeResultsCollection := dbClient.Database("scrapeit").Collection("scrape_results")
	_, err = scrapeResultsCollection.DeleteMany(context.TODO(), bson.M{"endpointId": endpointId, "groupId": groupId})

	return err
}

func DeleteScrapingGroupEndpoint(c echo.Context) error {
	dbClient, ok := c.Get("db").(*mongo.Client)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get database client")
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

	err := deleteScrapingGroupEndpoint(dbClient, groupId, endpointId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Endpoint deleted successfully"})
}
