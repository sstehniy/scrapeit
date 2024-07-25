package handlers

import (
	"fmt"
	"net/http"
	"scrapeit/internal/models"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ImportScrapingGroupBody struct {
	Group models.ScrapeGroup `json:"group"`
}

func ImportScrapingGroup(e echo.Context) error {
	var body ImportScrapingGroupBody
	if err := e.Bind(&body); err != nil {
		fmt.Println("Invalid request body")
		return e.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	body.Group.ID = primitive.NewObjectID()

	for _, endpoint := range body.Group.Endpoints {
		endpoint.ID = uuid.New().String()
	}

	dbClient, ok := e.Get("db").(*mongo.Client)
	if !ok {
		return e.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get database client"})
	}

	groupCollection := dbClient.Database("scrapeit").Collection("scrape_groups")

	_, err := groupCollection.InsertOne(e.Request().Context(), body.Group)
	if err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return e.JSON(http.StatusOK, map[string]string{"message": "Group imported"})
}
