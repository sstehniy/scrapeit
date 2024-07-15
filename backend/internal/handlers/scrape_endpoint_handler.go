package handlers

import (
	"fmt"
	"net/http"
	"scrapeit/internal/models"
	"scrapeit/internal/scraper"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ScraperEndpointHandlerRequest struct {
	EndpointId string `json:"endpointId"`
	GroupId    string `json:"groupId"`
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
	dbClient, ok := c.Get("db").(*mongo.Client)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get database client")
	}
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

	results, toReplace, err := scraper.ScrapeEndpoint(*endpointToScrape, *relevantGroup, dbClient)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	// allResultsBytes, err := utils.ReadJson("/internal/data/scrape_results.json")
	// if err != nil {
	// 	return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	// }
	// var allResults []models.ScrapeResult
	// err = json.Unmarshal(allResultsBytes, &allResults)
	// if err != nil {
	// 	return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	// }
	// allResults = append(allResults, results...)
	// utils.WriteJson("/internal/data/scrape_results.json", results)

	allResultsCollection := dbClient.Database("scrapeit").Collection("scrape_results")
	var interfaceResults []interface{}
	for _, r := range results {
		r.ID = primitive.NewObjectID()
		interfaceResults = append(interfaceResults, r)
	}
	_, err = allResultsCollection.InsertMany(c.Request().Context(), interfaceResults)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Update existing results
	if len(toReplace) > 0 {
		var bulkWrites []mongo.WriteModel
		for _, r := range toReplace {
			update := mongo.NewUpdateOneModel().
				SetFilter(bson.M{"uniqueHash": r.UniqueHash, "groupId": r.GroupId}).
				SetUpdate(bson.M{"$set": bson.M{
					"fields":    r.Fields,
					"timestamp": r.Timestamp,
				}})
			bulkWrites = append(bulkWrites, update)
		}

		_, err = allResultsCollection.BulkWrite(c.Request().Context(), bulkWrites)
		if err != nil {
			fmt.Println("Error updating existing results:", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
	}

	return c.JSON(http.StatusOK, results)
}
