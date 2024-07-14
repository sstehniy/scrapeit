package handlers

import (
	"context"
	"fmt"
	"net/http"
	"scrapeit/internal/models"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type GetScrapingResultsRespones struct {
	Results []models.ScrapeResult `json:"results"`
	HasMore bool                  `json:"hasMore"`
}

func getScrapeResults(
	ctx context.Context, groupId string, offset, limit int, endpointIds []string,
	isArchive bool,
	client *mongo.Client,
) ([]models.ScrapeResult, bool, error) {

	findOptions := options.FindOptions{}
	findOptions.SetSkip(int64(offset))
	findOptions.SetLimit(int64(limit))
	findOptions.SetSort(bson.D{{Key: "timestamp", Value: -1}})

	groupObjId, err := primitive.ObjectIDFromHex(groupId)
	if err != nil {
		fmt.Println("error", err)
		return nil, false, err
	}
	collectionName := "scrape_results"

	if isArchive {
		collectionName = "archived_scrape_results"
	}

	cursor, err := client.Database("scrapeit").Collection(collectionName).Find(ctx, bson.M{
		"groupId":    groupObjId,
		"endpointId": bson.M{"$in": endpointIds},
	}, &findOptions)
	if err != nil {
		return nil, false, err
	}
	defer cursor.Close(ctx)

	var endpointResults []models.ScrapeResult
	for cursor.Next(ctx) {
		var result models.ScrapeResult
		if err := cursor.Decode(&result); err != nil {
			return nil, false, err
		}
		endpointResults = append(endpointResults, result)
	}

	hasMore := false
	limit = limit + 1

	newFindOptions := options.FindOptions{}
	newFindOptions.SetSkip(int64(offset + limit))
	newFindOptions.SetLimit(1)

	hasMoreCursor, err := client.Database("scrapeit").Collection("scrape_results").Find(ctx, bson.M{
		"groupId":    groupObjId,
		"endpointId": bson.M{"$in": endpointIds},
	}, &newFindOptions)
	if err != nil {
		fmt.Println("error", err)
		return nil, false, err

	}
	defer hasMoreCursor.Close(ctx)

	for hasMoreCursor.Next(ctx) {
		hasMore = true
	}

	fmt.Println("endpointResults", endpointResults)
	return endpointResults, hasMore, nil
}

func GetScrapingResults(c echo.Context) error {
	dbClient, ok := c.Get("db").(*mongo.Client)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get database client")
	}
	groupId := c.Param("groupId")
	offset := c.QueryParam("offset")
	if offset == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Missing parameter 'offset'",
		})
	}
	offsetNumber, err := strconv.Atoi(offset)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid parameter type for 'offset': must be a number",
		})
	}
	if offsetNumber < 0 {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "invalid value for parameter 'offset': must be a positive number",
		})
	}
	limit := c.QueryParam("limit")
	if limit == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Missing parameter 'limit'",
		})
	}
	limitNumber, err := strconv.Atoi(limit)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid parameter type for 'limit': must be a number",
		})
	}
	if limitNumber < 0 {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "invalid value for parameter 'limitNumber': must be a positive number",
		})
	}
	// should be an array of endpoint ids
	endpointIdsParam := c.QueryParam("endpointIds")

	isArchive := false
	// query params
	isArchiveParam := c.QueryParam("isArchive")
	if isArchiveParam == "true" {
		isArchive = true
	}

	endpointIds := strings.Split(endpointIdsParam, ",")
	if len(endpointIds) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Missing parameter 'endpointIds'",
		})
	}

	fmt.Println("endpointIds", endpointIds)

	results, hasMore, err := getScrapeResults(c.Request().Context(), groupId, offsetNumber, limitNumber, endpointIds, isArchive, dbClient)
	if err != nil {
		fmt.Println("error", err)
		return c.JSON(http.StatusOK, map[string]interface{}{})
	}
	return c.JSON(http.StatusOK, GetScrapingResultsRespones{
		Results: results,
		HasMore: hasMore,
	})

}
