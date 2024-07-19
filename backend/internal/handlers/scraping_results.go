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

type GetScrapingResultsQueryParams struct {
	Offset      int32    `query:"offset"`
	Limit       int32    `query:"limit"`
	EndpointIds []string `query:"endpointIds"`
	GroupId     string   `query:"groupId"`
	Q           string   `query:"q"`
	IsArchive   bool     `query:"isArchive"`
}

func getScrapeResults(
	ctx context.Context, params GetScrapingResultsQueryParams,
	client *mongo.Client,
) ([]models.ScrapeResult, bool, error) {

	findOptions := options.FindOptions{}
	findOptions.SetSkip(int64(params.Offset))
	findOptions.SetLimit(int64(params.Limit))

	groupObjId, err := primitive.ObjectIDFromHex(params.GroupId)
	if err != nil {
		fmt.Println("error", err)
		return nil, false, err
	}
	collectionName := "scrape_results"

	if params.IsArchive {
		collectionName = "archived_scrape_results"
	}

	filter := bson.M{}

	if params.Q != "" {
		fmt.Println("searching for", params.Q)
		filter["$text"] = bson.M{"$search": "\"" + params.Q + "\""}
		findOptions.SetSort(bson.M{"score": bson.M{"$meta": "textScore"}})
		findOptions.SetProjection(bson.M{"score": bson.M{"$meta": "textScore"}})
	} else {
		filter = bson.M{
			"groupId":    groupObjId,
			"endpointId": bson.M{"$in": params.EndpointIds},
		}
		findOptions.SetSort(bson.D{{Key: "timestampLastUpdate", Value: -1}})
	}

	fmt.Println("filter", filter)
	fmt.Println("collection", collectionName)

	cursor, err := client.Database("scrapeit").Collection(collectionName).Find(ctx, filter, &findOptions)
	if err != nil {
		return nil, false, err
	}
	defer cursor.Close(ctx)

	endpointResults := []models.ScrapeResult{}
	for cursor.Next(ctx) {
		var result models.ScrapeResult
		if err := cursor.Decode(&result); err != nil {
			return nil, false, err
		}
		endpointResults = append(endpointResults, result)
	}

	hasMore := false
	limit := params.Limit + 1

	findOptions.SetSkip(int64(params.Offset + limit))
	findOptions.SetLimit(1)

	hasMoreCursor, err := client.Database("scrapeit").Collection("scrape_results").Find(ctx, filter, &findOptions)
	if err != nil {
		fmt.Println("error", err)
		return nil, false, err

	}
	defer hasMoreCursor.Close(ctx)

	for hasMoreCursor.Next(ctx) {
		hasMore = true
	}

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

	q := c.QueryParam("q")

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

	params := GetScrapingResultsQueryParams{
		Offset:      int32(offsetNumber),
		Limit:       int32(limitNumber),
		IsArchive:   isArchive,
		GroupId:     groupId,
		Q:           q,
		EndpointIds: endpointIds,
	}

	fmt.Println("endpointIds", endpointIds)

	results, hasMore, err := getScrapeResults(c.Request().Context(), params, dbClient)
	if err != nil {
		fmt.Println("error", err)
		return c.JSON(http.StatusOK, map[string]interface{}{})
	}
	return c.JSON(http.StatusOK, GetScrapingResultsRespones{
		Results: results,
		HasMore: hasMore,
	})

}
