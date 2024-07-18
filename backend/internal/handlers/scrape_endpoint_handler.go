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
	Internal   bool   `json:"internal"`
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

	browser := scraper.GetBrowser()
	defer browser.Close()

	results, toReplace, err := scraper.ScrapeEndpoint(*endpointToScrape, *relevantGroup, dbClient, browser)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
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
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
	}

	// update group and set endpoint status to idle
	endpointToScrape.Status = models.ScrapeStatusIdle
	_, err = groupCollection.UpdateOne(c.Request().Context(), groupQuery, bson.M{"$set": bson.M{"endpoints": relevantGroup.Endpoints}})
	if err != nil {
		fmt.Println("Error updating group:", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, results)
}
