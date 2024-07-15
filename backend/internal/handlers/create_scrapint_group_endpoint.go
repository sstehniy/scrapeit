package handlers

import (
	"fmt"
	"net/http"
	"scrapeit/internal/cron"
	"scrapeit/internal/models"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CreateScrapingGroupRequest struct {
	NewEndpoint models.Endpoint `json:"endpoint"`
}

type CreateScrapingGroupResponse struct {
	Success bool `json:"success"`
}

func CreateScrapingGroupEndpoint(c echo.Context) error {
	var body CreateScrapingGroupRequest
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	var relevantGroup *models.ScrapeGroup

	groupIdParam := c.Param("groupId")
	groupId, err := primitive.ObjectIDFromHex(groupIdParam)
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
		fmt.Println(groupResult.Err())
		return c.JSON(http.StatusBadRequest, map[string]string{"error": groupResult.Err().Error()})
	}
	err = groupResult.Decode(&relevantGroup)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	relevantGroup.Endpoints = append(relevantGroup.Endpoints, body.NewEndpoint)

	updateQuery := bson.M{"$set": bson.M{"endpoints": relevantGroup.Endpoints}}
	_, err = groupCollection.UpdateOne(c.Request().Context(), groupQuery, updateQuery)
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	cronManager, ok := c.Get("cron").(*cron.CronManager)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get cron manager")
	}

	cronManager.AddJob(cron.CronManagerJob{
		GroupID:    groupIdParam,
		EndpointID: body.NewEndpoint.ID,
		Interval:   body.NewEndpoint.Interval,
		Active:     true,
		Job: func() error {
			fmt.Println("Running job for", groupIdParam, body.NewEndpoint.ID)
			return HandleCallInternalScrapeEndpoint(c.Echo(), groupIdParam, body.NewEndpoint.ID, dbClient, cronManager)
		},
	})

	return c.JSON(http.StatusOK, CreateScrapingGroupResponse{Success: true})
}
