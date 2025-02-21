package handlers

import (
	"context"
	"fmt"
	"net/http"
	"scrapeit/internal/helpers"
	"scrapeit/internal/models"
	"scrapeit/internal/scraper"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ScraperEndpointHandlerRequest struct {
	EndpointId string `json:"endpointId"`
	GroupId    string `json:"groupId"`
	Internal   bool   `json:"internal"`
}

type ScraperEndpointHandlerResponse struct {
	NewResults      []models.ScrapeResult `json:"newResults"`
	ReplacedResults []models.ScrapeResult `json:"replacedResults"`
}

func ScrapeEndpointHandler(c echo.Context) error {
	var body ScraperEndpointHandlerRequest
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	var relevantGroup *models.ScrapeGroup

	groupId, err := primitive.ObjectIDFromHex(body.GroupId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	groupQuery := bson.M{"_id": groupId}
	dbClient, _ := models.GetDbClient()
	groupCollection := dbClient.Database("scrapeit").Collection("scrape_groups")
	groupResult := groupCollection.FindOne(c.Request().Context(), groupQuery)
	if groupResult.Err() != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": groupResult.Err().Error()})
	}
	err = groupResult.Decode(&relevantGroup)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	endpointToScrape := relevantGroup.GetEndpointById(body.EndpointId)

	if endpointToScrape == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Endpoint not found"})
	}

	browser := scraper.GetBrowser()

	results, toReplace, err := scraper.ScrapeEndpoint(*endpointToScrape, *relevantGroup, dbClient, browser)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	seenUniqueHashes := make(map[string]bool)
	allResultsCollection := dbClient.Database("scrapeit").Collection("scrape_results")

	toInsert := []interface{}{}
	for _, r := range results {
		if seenUniqueHashes[r.UniqueHash] {
			continue
		}
		seenUniqueHashes[r.UniqueHash] = true
		r.ID = primitive.NewObjectID()
		toInsert = append(toInsert, r)
	}

	_, err = allResultsCollection.InsertMany(context.TODO(), toInsert, &options.InsertManyOptions{})
	if err != nil {
		fmt.Println("Error inserting new results:", err)
	}

	fmt.Println("Here go the to replace", toReplace)

	if len(toReplace) > 0 {
		var bulkWrites []mongo.WriteModel
		for _, r := range toReplace {
			update := mongo.NewUpdateOneModel().
				SetFilter(bson.M{"_id": r.ID}).
				SetUpdate(bson.M{"$set": bson.M{
					"fields":              r.Fields,
					"timestampLastUpdate": r.TimestampLastUpdate,
				}})
			bulkWrites = append(bulkWrites, update)
		}

		fmt.Println("Len writes", len(bulkWrites))

		_, err = allResultsCollection.BulkWrite(c.Request().Context(), bulkWrites)
		if err != nil {
			fmt.Println("Error updating existing results:", err)
		}
	}

	// update group and set endpoint status to idle
	endpointToScrape.Status = models.ScrapeStatusIdle
	_, err = groupCollection.UpdateOne(c.Request().Context(), groupQuery, bson.M{"$set": bson.M{"endpoints": relevantGroup.Endpoints}})
	if err != nil {
		fmt.Println("Error updating group:", err)
	}

	notificationConfigs := []models.NotificationConfig{}

	notificationConfigResult, err := dbClient.Database("scrapeit").Collection("notification_configs").Find(c.Request().Context(), bson.M{"groupId": relevantGroup.ID})
	if err != nil {
		fmt.Println("Failed to get notification configs")
	} else {
		for notificationConfigResult.Next(context.Background()) {
			var config models.NotificationConfig
			if err := notificationConfigResult.Decode(&config); err != nil {
				fmt.Println("failed to decode config into struct")
			} else {
				notificationConfigs = append(notificationConfigs, config)
			}
		}
		fmt.Println("Here is the result", notificationConfigs)
		if len(notificationConfigs) > 0 {
			go helpers.HandleNotifyResults(notificationConfigs, *relevantGroup, results, toReplace)
		}
	}

	defer notificationConfigResult.Close(context.Background())

	return c.JSON(http.StatusOK, ScraperEndpointHandlerResponse{
		NewResults:      results,
		ReplacedResults: toReplace,
	})
}
