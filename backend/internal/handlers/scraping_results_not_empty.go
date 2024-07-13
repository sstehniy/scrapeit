package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type GetScrapingResultsNotEmptyResponse struct {
	ResultsNotEmpty bool `json:"resultsNotEmpty"`
}

func GetScrapingResultsNotEmpty(c echo.Context) error {
	dbClient, ok := c.Get("db").(*mongo.Client)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get database client")
	}
	groupId := c.Param("groupId")
	groupObjId, err := primitive.ObjectIDFromHex(groupId)
	if err != nil {
		fmt.Println("error", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid group ID")
	}

	scrapeResultsCollection := dbClient.Database("scrapeit").Collection("scrape_results")
	scrapeResultsCount, err := scrapeResultsCollection.CountDocuments(c.Request().Context(), bson.M{"groupId": groupObjId})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get scrape results count")
	}

	return c.JSON(http.StatusOK, GetScrapingResultsNotEmptyResponse{ResultsNotEmpty: scrapeResultsCount > 0})

}
