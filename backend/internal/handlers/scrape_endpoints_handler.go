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

type ScraperEndpointsHandlerRequest struct {
	EndpointIds []string `json:"endpointIds"`
	GroupId     string   `json:"groupId"`
}

type ScraperEndpointsHandlerResponse struct {
	NewResults      []models.ScrapeResult `json:"newResults"`
	ReplacedResults []models.ScrapeResult `json:"replacedResults"`
}

func ScrapeEndpointsHandler(c echo.Context) error {
	var body ScraperEndpointsHandlerRequest
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
	var endpointsToScrape []models.Endpoint

	for _, endpointId := range body.EndpointIds {
		endpointToScrape := relevantGroup.GetEndpointById(endpointId)
		if endpointToScrape != nil {
			endpointsToScrape = append(endpointsToScrape, *endpointToScrape)
		}
	}

	resultsChan := make(chan []models.ScrapeResult)
	toReplaceChan := make(chan []models.ScrapeResult)

	for _, endpointToScrape := range endpointsToScrape {
		fmt.Println("Scraping endpoints:", endpointToScrape.ID)

		go func(endpoint models.Endpoint) {
			results, toReplace, err := scraper.ScrapeEndpoint(endpoint, *relevantGroup, dbClient)
			if err != nil {
				fmt.Println("Error scraping endpoint:", err)
				resultsChan <- []models.ScrapeResult{}
				toReplaceChan <- []models.ScrapeResult{}
			} else {
				resultsChan <- results
				toReplaceChan <- toReplace
			}
		}(endpointToScrape)
	}

	var results []models.ScrapeResult
	var toReplaceResults []models.ScrapeResult
	for range endpointsToScrape {
		results = append(results, <-resultsChan...)
		toReplaceResults = append(toReplaceResults, <-toReplaceChan...)
	}

	allResultsCollection := dbClient.Database("scrapeit").Collection("scrape_results")
	var interfaceResults []interface{}
	for _, r := range results {
		r.ID = primitive.NewObjectID()
		interfaceResults = append(interfaceResults, r)
	}

	if len(interfaceResults) > 0 {
		_, err = allResultsCollection.InsertMany(c.Request().Context(), interfaceResults)
		if err != nil {
			fmt.Println("Error inserting results:", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
	}

	// Update existing results
	if len(toReplaceResults) > 0 {
		var bulkWrites []mongo.WriteModel
		for _, r := range toReplaceResults {
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

	return c.JSON(http.StatusOK, ScraperEndpointsHandlerResponse{
		NewResults:      results,
		ReplacedResults: toReplaceResults,
	})
}
