package handlers

import (
	"context"
	"fmt"
	"net/http"

	"scrapeit/internal/export"
	"scrapeit/internal/models"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ExportGroupResultsHandlerRequest struct {
	Type              models.ExportType `json:"type"`
	FileName          models.ExportType `json:"fileName"`
	DeleteAfterExport bool              `json:"deleteAfterExport"`
}

func ExportGroupResultsHandler(c echo.Context) error {
	dbClient, ok := c.Get("db").(*mongo.Client)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get database client")
	}
	groupIdString := c.Param("groupId")
	groupId, err := primitive.ObjectIDFromHex(groupIdString)
	if err != nil {
		fmt.Printf("ExportGroupResultsHandler: error parsing objectId: %s ", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid objectId",
		})
	}

	var tmp map[string]interface{}
	var group models.ScrapeGroup
	var archivedGroup models.ArchivedScrapeGroup
	isArchived := false
	groupResult := dbClient.Database("scrapeit").Collection("scrape_groups").FindOne(context.Background(), bson.M{"_id": groupId})
	if groupResult.Err() == mongo.ErrNoDocuments {
		fmt.Printf("ExportGroupResultsHandler: requested group not found: %s", groupResult.Err())
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "requested group not found",
		})
	}
	if err := groupResult.Decode(&tmp); err != nil {
		fmt.Printf("ExportGroupResultsHandler: could not parse mongo document to interface{}: %s", err.Error())
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "could not parse mongo document to interface{}",
		})
	}

	if _, ok := tmp["originalId"]; ok {
		isArchived = true
		if err = groupResult.Decode(&archivedGroup); err != nil {
			fmt.Printf("ExportGroupResultsHandler: failed to decode groupResult in archivedGroup: %s", err.Error())
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "failed to decode groupResult in archivedGroup",
			})
		}
	} else {
		if err = groupResult.Decode(&group); err != nil {
			fmt.Printf("ExportGroupResultsHandler: failed to decode groupResult in group: %s", err.Error())
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "failed to decode groupResult in group",
			})
		}
	}

	var collectionName string
	var groupIdForResults primitive.ObjectID
	var versionTag string = ""

	if isArchived {
		collectionName = "archived_scrape_results"
		groupIdForResults = archivedGroup.ID
		versionTag = archivedGroup.VersionTag
	} else {
		collectionName = "scrape_results"
		groupIdForResults = group.ID
	}

	var results []models.ScrapeResult
	sortOpt := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}})

	cursor, err := dbClient.Database("scrapeit").Collection(collectionName).Find(context.Background(), bson.M{
		"groupId":         groupIdForResults,
		"groupVersionTag": versionTag,
	}, sortOpt)

	if err != nil {
		fmt.Printf("ExportGroupResultsHandler: failed to get cursor for results: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to get cursor for results",
		})
	}

	for cursor.Next(context.Background()) {
		var result models.ScrapeResult
		if err := cursor.Decode(&result); err != nil {
			fmt.Printf("ExportGroupResultsHandler: failed to decode cursor.Next into result: %s", err.Error())
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "failed to decode cursor.Next into result",
			})
		}
		results = append(results, result)
	}

	var body ExportGroupResultsHandlerRequest
	if err = c.Bind(&body); err != nil {
		fmt.Printf("ExportGroupResultsHandler: invalid request body: %s", err.Error())
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	bytes, err := export.CreateResultsExportFile(results, group, body.Type)

	if err != nil {
		fmt.Printf("ExportGroupResultsHandler: failed to create export file %s", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to create export file",
		})
	}

	extension, err := export.GetFileExtension(body.Type)
	if err != nil {
		fmt.Printf("ExportGroupResultsHandler: failed to get export type extension: %s", err.Error())
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "failed to get export type extension",
		})
	}
	c.Response().Header().Set(echo.HeaderContentType, "application/octet-stream")
	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%s%s", body.FileName, extension))
	c.Response().WriteHeader(http.StatusOK)

	_, err = c.Response().Write(bytes)
	if err != nil {
		fmt.Printf("ExportGroupResultsHandler: failed to write bytes to response: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to write bytes to response",
		})
	}
	return nil
}
