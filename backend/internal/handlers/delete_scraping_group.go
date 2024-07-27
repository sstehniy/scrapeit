package handlers

import (
	"context"
	"fmt"
	"net/http"
	"scrapeit/internal/cron"
	"scrapeit/internal/models"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func deleteScrapingGroup(ctx context.Context, groupId primitive.ObjectID, client *mongo.Client) error {

	groupCollection := client.Database("scrapeit").Collection("scrape_groups")
	fmt.Println("Deleting group with ID: ", groupId)
	_, err := groupCollection.DeleteOne(ctx, bson.M{"_id": groupId})
	if err != nil {
		return err
	}

	return nil
}

func DeleteScrapingGroup(c echo.Context) error {
	groupId := c.Param("id")
	dbClient, ok := c.Get("db").(*mongo.Client)
	if !ok {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get database client"})
	}

	cronManager, ok := c.Get("cron").(*cron.CronManager)
	if !ok {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get cron manager"})
	}

	groupIdObj, err := primitive.ObjectIDFromHex(groupId)

	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid group ID"})
	}

	var group models.ScrapeGroup
	groupFilter := bson.M{"_id": groupIdObj}
	err = dbClient.Database("scrapeit").Collection("scrape_groups").FindOne(c.Request().Context(), groupFilter).Decode(&group)

	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Group not found"})
	}

	for _, endpoint := range group.Endpoints {
		if endpoint.Active {
			cronManager.DestroyJob(group.ID.Hex(), endpoint.ID)
		}
	}

	err = deleteScrapingGroup(c.Request().Context(), groupIdObj, dbClient)

	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
	}

	allEndpointsIds := []string{}
	for _, endpoint := range group.Endpoints {
		allEndpointsIds = append(allEndpointsIds, endpoint.ID)
	}

	endpointsFilter := bson.M{"endpointId": bson.M{"$in": allEndpointsIds}, "groupId": groupIdObj}

	_, err = dbClient.Database("scrapeit").Collection("scrape_results").DeleteMany(c.Request().Context(), endpointsFilter)

	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
	}

	archivedFilter := bson.M{"originalId": groupIdObj}
	cursor, err := dbClient.Database("scrapeit").Collection("scrape_groups").Find(c.Request().Context(), archivedFilter)

	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
	}
	defer cursor.Close(context.Background())

	allArchivedGroupIds := []primitive.ObjectID{}

	for cursor.Next(c.Request().Context()) {
		var result bson.M
		err := cursor.Decode(&result)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"error": err.Error(),
			})
		}
		archivedGroupId := result["_id"].(primitive.ObjectID)
		allArchivedGroupIds = append(allArchivedGroupIds, archivedGroupId)
	}

	archived_results_filter := bson.M{"groupId": bson.M{"$in": allArchivedGroupIds}}
	_, err = dbClient.Database("scrapeit").Collection("archived_scrape_results").DeleteMany(c.Request().Context(), archived_results_filter)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
	}

	_, err = dbClient.Database("scrapeit").Collection("scrape_groups").DeleteMany(c.Request().Context(), archivedFilter)

	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
	}

	// remove notification configs
	notificationConfigsFilter := bson.M{"groupId": group.ID}

	_, err = dbClient.Database("scrapeit").Collection("notification_configs").DeleteMany(context.Background(), notificationConfigsFilter)

	if err != nil {
		fmt.Printf("Failed to delete notification configs for group: %v", err)
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Group deleted successfully"})

}
