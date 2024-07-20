package handlers

import (
	"fmt"
	"net/http"
	"scrapeit/internal/cron"
	"scrapeit/internal/models"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UpdateScrapingGroupSchemaRequest struct {
	Schema        []models.Field       `json:"fields" bson:"fields"`
	Changes       []models.FieldChange `json:"changes"`
	VersionTag    string               `json:"versionTag"`
	ShouldArchive bool                 `json:"shouldArchive"`
}

func UpdateScrapingGroupSchema(c echo.Context) error {
	dbClient, ok := c.Get("db").(*mongo.Client)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get database client")
	}
	cronManager, ok := c.Get("cron").(*cron.CronManager)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get cron manager")
	}

	var req UpdateScrapingGroupSchemaRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	groupIdString := c.Param("groupId")
	groupId, err := primitive.ObjectIDFromHex(groupIdString)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid group ID")
	}

	groupResult := dbClient.Database("scrapeit").Collection("scrape_groups").FindOne(c.Request().Context(), bson.M{"_id": groupId})

	if groupResult.Err() == mongo.ErrNoDocuments {
		return echo.NewHTTPError(http.StatusNotFound, "Group not found")
	}

	group := new(models.ScrapeGroup)

	group.Updated = primitive.NewDateTimeFromTime(time.Now())

	err = groupResult.Decode(group)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get updated group")
	}

	for _, endpoint := range group.Endpoints {
		cronManager.DestroyJob(group.ID.Hex(), endpoint.ID)
	}

	if req.ShouldArchive {

		if req.VersionTag == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "Version tag is required to archive group")
		}
		groupCopy := models.ArchivedScrapeGroup{}
		newGroupId := primitive.NewObjectID()
		groupCopy.OriginalID = group.ID
		groupCopy.Name = group.Name
		groupCopy.Fields = group.Fields
		groupCopy.Endpoints = group.Endpoints
		groupCopy.WithThumbnail = group.WithThumbnail
		groupCopy.VersionTag = req.VersionTag
		groupCopy.ID = newGroupId
		groupCopy.Created = group.Created
		groupCopy.Updated = group.Updated
		_, err = dbClient.Database("scrapeit").Collection("scrape_groups").InsertOne(c.Request().Context(), groupCopy)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to archive group")
		}

		var allScrapeResults []models.ScrapeResult
		// move all scrape results to archived scrape results
		allScrapeResultsResult, err := dbClient.Database("scrapeit").Collection("scrape_results").Find(c.Request().Context(), bson.M{"groupId": groupId})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get scrape results")
		}

		for allScrapeResultsResult.Next(c.Request().Context()) {
			var result models.ScrapeResult
			err = allScrapeResultsResult.Decode(&result)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to decode scrape result")
			}
			allScrapeResults = append(allScrapeResults, result)
		}
		var archiveScrapeResults []interface{}
		for _, result := range allScrapeResults {
			result.GroupVersionTag = req.VersionTag
			result.GroupId = newGroupId
			archiveScrapeResults = append(archiveScrapeResults, result)
		}

		_, err = dbClient.Database("scrapeit").Collection("archived_scrape_results").InsertMany(c.Request().Context(), archiveScrapeResults)

		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to archive scrape results")
		}

		_, err = dbClient.Database("scrapeit").Collection("scrape_results").DeleteMany(c.Request().Context(), bson.M{"groupId": groupId})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete scrape results")
		}
	}

	group.Fields = req.Schema

	if len(group.Endpoints) > 0 {
		for _, change := range req.Changes {
			switch change.ChangeType {
			case models.AddField:
				for idx, endpoint := range group.Endpoints {
					addedField := group.GetFieldById(change.FieldID)
					if addedField == nil {
						return echo.NewHTTPError(http.StatusNotFound, "Field not found")
					}
					endpoint.DetailFieldSelectors = append(endpoint.DetailFieldSelectors, models.FieldSelector{
						ID:                   uuid.New().String(),
						FieldID:              addedField.ID,
						Selector:             "",
						Regex:                "",
						AttributeToGet:       "",
						LockedForEdit:        false,
						RegexMatchIndexToUse: 0,
						SelectorStatus:       models.SelectorStatusNew,
					})
					group.Endpoints[idx] = endpoint
					fmt.Println("Endpoint after adding field", endpoint)
				}
			case models.DeleteField:
				for idx, endpoint := range group.Endpoints {
					for i, fieldSelector := range endpoint.DetailFieldSelectors {
						if fieldSelector.FieldID == change.FieldID {
							endpoint.DetailFieldSelectors = append(endpoint.DetailFieldSelectors[:i], endpoint.DetailFieldSelectors[i+1:]...)
							group.Endpoints[idx] = endpoint
						}
					}
				}

			case models.ChangeFieldKey, models.ChangeFieldName, models.ChangeFieldType:
				if !change.FieldIsNewSinceLastSave {
					fmt.Printf("Field is not new since last save: %+v\n", change)
					for idx, endpoint := range group.Endpoints {
						for i, fieldSelector := range endpoint.DetailFieldSelectors {
							if fieldSelector.FieldID == change.FieldID {
								fieldSelector.SelectorStatus = models.SelectorStatusNeedsUpdate
								endpoint.DetailFieldSelectors[i] = fieldSelector
								group.Endpoints[idx] = endpoint
							}
						}
					}
				}

			}
		}
	}

	fmt.Printf("Group after changes: %+v\n", group)

	result, err := dbClient.Database("scrapeit").Collection("scrape_groups").UpdateOne(
		c.Request().Context(),
		bson.M{"_id": groupId},
		bson.M{"$set": bson.M{"fields": group.Fields, "endpoints": group.Endpoints}},
	)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update group schema and endpoints")
	}

	if result.MatchedCount == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "Group not found")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Group schema updated"})
}
