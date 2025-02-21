package handlers

import (
	"context"
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
	dbClient, _ := models.GetDbClient()
	cronManager := cron.GetCronManager()

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

	go func() {
		for _, endpoint := range group.Endpoints {
			cronManager.DestroyJob(group.ID.Hex(), endpoint.ID)
		}
	}()

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

		defer allScrapeResultsResult.Close(context.Background())

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

	foundNotificationConfigs := []*models.NotificationConfig{}
	collection := dbClient.Database("scrapeit").Collection("notification_configs")

	cursor, err := collection.Find(c.Request().Context(), bson.M{"groupId": groupId})

	if err != nil {
		fmt.Println("Failed to get notification configs")
	} else {
		for cursor.Next(c.Request().Context()) {
			var config models.NotificationConfig
			if err := cursor.Decode(&config); err != nil {
				fmt.Println("Failed to decode config to struct")
			}
			foundNotificationConfigs = append(foundNotificationConfigs, &config)
		}

		if err := cursor.Err(); err != nil {
			fmt.Println("Failed to get Notification Config cursor")
		}
	}

	defer cursor.Close(c.Request().Context())

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
				// check if field is in fieldIdsToNotify or conditions, if so, remove it
				for _, config := range foundNotificationConfigs {
					if config != nil {
						for i, fieldId := range config.FieldIdsToNotify {
							if fieldId == change.FieldID {
								config.FieldIdsToNotify = append(config.FieldIdsToNotify[:i], config.FieldIdsToNotify[i+1:]...)
							}
						}
						for i, condition := range config.Conditions {
							if condition.FieldId == change.FieldID {
								config.Conditions = append(config.Conditions[:i], config.Conditions[i+1:]...)
							}
						}
					}
				}

				for idx, endpoint := range group.Endpoints {
					for i, fieldSelector := range endpoint.DetailFieldSelectors {
						if fieldSelector.FieldID == change.FieldID {
							endpoint.DetailFieldSelectors = append(endpoint.DetailFieldSelectors[:i], endpoint.DetailFieldSelectors[i+1:]...)
							group.Endpoints[idx] = endpoint
						}
					}
				}

			case models.ChangeFieldKey, models.ChangeFieldName, models.ChangeFieldType:
				// check if field is in fieldIdsToNotify or conditions, if so,
				// check if changed field is FieldTypeNumber, otherwise remove it from conditions
				for _, config := range foundNotificationConfigs {
					if config != nil {
						for i, condition := range config.Conditions {
							if condition.FieldId == change.FieldID && !change.FieldIsNewSinceLastSave {
								foundField := group.GetFieldById(change.FieldID)
								if foundField == nil {
									continue
								}
								if foundField.Type != models.FieldTypeNumber {
									config.Conditions = append(config.Conditions[:i], config.Conditions[i+1:]...)
								}
							}
						}
					}
				}
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

	// update notification config
	for _, config := range foundNotificationConfigs {
		if config != nil {
			notificationConfigCollection := dbClient.Database("scrapeit").Collection("notification_configs")
			_, err = notificationConfigCollection.UpdateOne(c.Request().Context(), bson.M{"groupId": groupId}, bson.M{"$set": config})
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update notification config")
			}
			fmt.Println("Notification config updated")
		}
	}

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
