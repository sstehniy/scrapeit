package handlers

import (
	"context"
	"fmt"
	"net/http"
	"scrapeit/internal/models"

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

type FilterOperator string

const (
	FilterOperatorEqual       FilterOperator = "="
	FilterOperatorNotEqual    FilterOperator = "!="
	FilterOperatorGreaterThan FilterOperator = ">"
	FilterOperatorLessThan    FilterOperator = "<"
)

type SearchFilter struct {
	FieldId  string         `json:"fieldId"`
	Value    interface{}    `json:"value"`
	Operator FilterOperator `json:"operator"`
}

type Sort int

const (
	SortAsc  Sort = 1
	SortDesc Sort = -1
)

type SearchSort struct {
	FieldId string `json:"fieldId"`
	Order   Sort   `json:"order"`
}

type GetScrapingResultsRequest struct {
	Offset      int64          `query:"offset"`
	Limit       int64          `query:"limit"`
	EndpointIds []string       `query:"endpointIds"`
	GroupId     string         `query:"groupId"`
	Q           string         `query:"q"`
	IsArchive   bool           `query:"isArchive"`
	Filters     []SearchFilter `query:"filters"`
	Sort        SearchSort     `query:"sort"`
}

func getScrapeResults(
	ctx context.Context, params GetScrapingResultsRequest,
	client *mongo.Client,
) ([]models.ScrapeResult, bool, error) {

	findOptions := options.Find()
	findOptions.SetSkip(params.Offset)
	findOptions.SetLimit(params.Limit)

	groupObjId, err := primitive.ObjectIDFromHex(params.GroupId)
	if err != nil {
		fmt.Println("error", err)
		return nil, false, err
	}
	collectionName := "scrape_results"

	if params.IsArchive {
		collectionName = "archived_scrape_results"
	}

	filter := bson.M{
		"groupId":    groupObjId,
		"endpointId": bson.M{"$in": params.EndpointIds},
	}

	if len(params.Filters) > 0 {
		// Initialize the filter structure for fields
		var fieldConditions []bson.M

		for _, requestFilter := range params.Filters {
			var condition bson.M
			switch requestFilter.Operator {
			case FilterOperatorEqual:
				condition = bson.M{
					"fields": bson.M{
						"$elemMatch": bson.M{
							"fieldId": requestFilter.FieldId,
							"value":   requestFilter.Value,
						},
					},
				}
			case FilterOperatorNotEqual:
				condition = bson.M{
					"fields": bson.M{
						"$elemMatch": bson.M{
							"fieldId": requestFilter.FieldId,
							"value":   bson.M{"$ne": requestFilter.Value},
						},
					},
				}
			case FilterOperatorGreaterThan:
				condition = bson.M{
					"fields": bson.M{
						"$elemMatch": bson.M{
							"fieldId": requestFilter.FieldId,
							"value":   bson.M{"$gt": requestFilter.Value},
						},
					},
				}
			case FilterOperatorLessThan:
				condition = bson.M{
					"fields": bson.M{
						"$elemMatch": bson.M{
							"fieldId": requestFilter.FieldId,
							"value":   bson.M{"$lt": requestFilter.Value},
						},
					},
				}
			default:
				fmt.Println("Unknown filter operator:", requestFilter.Operator)
				continue
			}

			fieldConditions = append(fieldConditions, condition)
		}

		if len(fieldConditions) > 0 {
			filter["$and"] = fieldConditions
		}
	}

	var pipeline mongo.Pipeline

	// Default sort by timestampLastUpdate
	if params.Sort.FieldId == "" {
		pipeline = mongo.Pipeline{
			{{"$match", filter}},
			{{"$sort", bson.D{{"timestampLastUpdate", -1}}}},
			{{"$skip", params.Offset}},
			{{"$limit", params.Limit}},
		}
	} else {
		// Sort by specific fieldId's value
		pipeline = mongo.Pipeline{
			{{"$match", filter}},
			{{"$addFields", bson.M{
				"sortField": bson.M{
					"$arrayElemAt": bson.A{
						bson.M{"$filter": bson.M{
							"input": "$fields",
							"cond":  bson.M{"$eq": bson.A{"$$this.fieldId", params.Sort.FieldId}},
						}},
						0,
					},
				},
			}}},
			{{"$sort", bson.D{{"sortField.value", params.Sort.Order}}}},
			{{"$skip", params.Offset}},
			{{"$limit", params.Limit}},
			{{"$unset", "sortField"}},
		}
	}
	cursor, err := client.Database("scrapeit").Collection(collectionName).Aggregate(ctx, pipeline)
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

	if params.Sort.FieldId == "" {
		pipeline = mongo.Pipeline{
			{{"$match", filter}},
			{{"$sort", bson.D{{"timestampLastUpdate", -1}}}},
			{{"$skip", params.Offset + limit}},
			{{"$limit", 1}},
		}
	} else {
		pipeline = mongo.Pipeline{
			{{"$match", filter}},
			{{"$unwind", "$fields"}},
			{{"$match", bson.M{"fields.fieldId": params.Sort.FieldId}}},
			{{"$sort", bson.D{{"fields.value", params.Sort.Order}}}},
			{{"$group", bson.M{
				"_id": "$_id",
				"doc": bson.M{"$first": "$$ROOT"},
			}}},
			{{"$replaceRoot", bson.M{"newRoot": "$doc"}}},
			{{"$sort", bson.D{{"timestampLastUpdate", -1}}}}, // Apply default sort after grouping
			{{"$skip", params.Offset + limit}},
			{{"$limit", 1}},
		}
	}

	hasMoreCursor, err := client.Database("scrapeit").Collection(collectionName).Aggregate(ctx, pipeline)
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
	dbClient, _ := models.GetDbClient()

	var body GetScrapingResultsRequest
	if err := c.Bind(&body); err != nil {
		fmt.Printf("error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Failed to parse request body",
		})
	}

	if body.GroupId == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Missing 'groupId' in request body",
		})
	}

	if body.Offset < 0 {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid value for 'offset': must be a positive number",
		})
	}
	if body.Limit <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid value for 'limit': must be a positive number",
		})
	}

	if len(body.EndpointIds) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Missing 'endpointIds' in request body",
		})
	}

	results, hasMore, err := getScrapeResults(c.Request().Context(), body, dbClient)
	if err != nil {
		fmt.Println("error", err)
		return c.JSON(http.StatusOK, map[string]interface{}{})
	}

	return c.JSON(http.StatusOK, GetScrapingResultsRespones{
		Results: results,
		HasMore: hasMore,
	})
}
