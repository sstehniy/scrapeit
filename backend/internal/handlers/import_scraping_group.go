package handlers

import (
	"fmt"
	"net/http"
	"scrapeit/internal/models"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

	for idx := range body.Group.Endpoints {
		body.Group.Endpoints[idx].ID = uuid.New().String()
	}

	dbClient, _ := models.GetDbClient()

	groupCollection := dbClient.Database("scrapeit").Collection("scrape_groups")

	_, err := groupCollection.InsertOne(e.Request().Context(), body.Group)
	if err != nil {
		return e.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return e.JSON(http.StatusOK, map[string]string{"message": "Group imported"})
}
