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

	for _, endpointToScrape := range endpointsToScrape {
		fmt.Println("Scraping endpoints:", endpointToScrape.ID)

		go func(endpoint models.Endpoint) {
			results, err := scraper.ScrapeEndpoint(endpoint, *relevantGroup, true, dbClient)
			if err != nil {
				fmt.Println("Error scraping endpoint:", err)
				resultsChan <- []models.ScrapeResult{}
			}
			if endpoint.ID == "713d796b-246c-409c-8031-b4e467eaaaee" {
				fmt.Println("Scraped endpoint:", len(results))
			}
			resultsChan <- results
		}(endpointToScrape)
	}

	var results []models.ScrapeResult
	for range endpointsToScrape {
		results = append(results, <-resultsChan...)
	}

	allResultsCollection := dbClient.Database("scrapeit").Collection("scrape_results")
	var interfaceResults []interface{}
	for _, r := range results {
		r.ID = primitive.NewObjectID()
		interfaceResults = append(interfaceResults, r)
	}
	if len(interfaceResults) == 0 {
		return c.JSON(http.StatusOK, results)
	}
	_, err = allResultsCollection.InsertMany(c.Request().Context(), interfaceResults)

	if err != nil {
		fmt.Println("Error inserting results:", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, results)
}
