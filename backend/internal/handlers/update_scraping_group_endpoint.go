package handlers

import (
	"fmt"
	"net/http"
	"scrapeit/internal/cron"
	"scrapeit/internal/models"
	"strings"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UpdateScrapingGroupEndpointRequest struct {
	Endpoint models.Endpoint `json:"endpoint" bson:"endpoint"`
}

func UpdateScrapingGroupEndpoint(c echo.Context) error {
	dbClient, ok := c.Get("db").(*mongo.Client)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get database client")
	}
	cronManager, ok := c.Get("cron").(*cron.CronManager)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get cron manager")
	}
	groupIdString := c.Param("groupId")
	endpointId := c.Param("endpointId")

	var req UpdateScrapingGroupEndpointRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body in UpdateScrapingGroupEndpointRequest")
	}

	groupId, err := primitive.ObjectIDFromHex(groupIdString)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid group ID")
	}

	groupResult := dbClient.Database("scrapeit").Collection("scrape_groups").FindOne(c.Request().Context(), bson.M{"_id": groupId})

	if groupResult.Err() == mongo.ErrNoDocuments {
		return echo.NewHTTPError(http.StatusNotFound, "Group not found")
	}

	var group models.ScrapeGroup

	err = groupResult.Decode(&group)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to parse found group")
	}

	newEndpoint := req.Endpoint

	foundJob := cronManager.GetJob(groupIdString, endpointId)

	if foundJob != nil && !newEndpoint.Active {
		cronManager.StopJob(groupIdString, endpointId)
	}

	if foundJob != nil && !foundJob.Active && newEndpoint.Active {
		cronManager.StartJob(groupIdString, endpointId)
	}

	oldEndpoint := group.GetEndpointById(endpointId)
	if oldEndpoint == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Endpoint Id")
	}

	if newEndpoint.Interval != oldEndpoint.Interval {
		cronManager, ok := c.Get("cron").(*cron.CronManager)
		if !ok {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get cron manager")
		}
		fmt.Println("Updating job interval")
		cronManager.UpdateJobInterval(groupIdString, endpointId, newEndpoint.Interval)
	}

	for _, oldSelector := range oldEndpoint.DetailFieldSelectors {
		var foundNewSelector *models.FieldSelector
		var foundNewSelectorIdx int
		for idx, sc := range newEndpoint.DetailFieldSelectors {
			if sc.ID == oldSelector.ID {
				foundNewSelector = &sc
				foundNewSelectorIdx = idx
				break
			}
		}

		if foundNewSelector != nil && strings.Trim(foundNewSelector.Selector, " ") != "" {
			foundNewSelector.SelectorStatus = models.SelectorStatusOk
			newEndpoint.DetailFieldSelectors[foundNewSelectorIdx] = *foundNewSelector
		}
	}
	updated := false
	for idx, ep := range group.Endpoints {
		if ep.ID == newEndpoint.ID {
			group.Endpoints[idx] = newEndpoint
			updated = true
			break
		}
	}

	if !updated {
		return echo.NewHTTPError(http.StatusBadRequest, "Endpoint ID does not match any existing endpoints in the group")
	}

	_, err = dbClient.Database("scrapeit").Collection("scrape_groups").UpdateOne(c.Request().Context(), bson.M{"_id": groupId}, bson.M{"$set": bson.M{"endpoints": group.Endpoints}})

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update group")
	}

	return c.String(http.StatusOK, "UpdateScrapingGroupEndpoint")
}
