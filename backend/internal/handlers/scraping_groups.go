package handlers

import (
	"fmt"
	"net/http"
	"scrapeit/internal/models"
	"scrapeit/internal/utils"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func writeAllGroups(groups *[]models.ScrapeGroup) error {
	return utils.WriteJson("/internal/data/scraping_groups.json", groups)
}

func GetScrapingGroups(c echo.Context) error {
	dbClient, ok := c.Get("db").(*mongo.Client)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get database client")
	}
	var allGroups []models.ScrapeGroup

	result, err := dbClient.Database("scrapeit").Collection("scrape_groups").Find(c.Request().Context(), bson.M{})
	if err != nil {
		fmt.Println("error in finding groups")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})

	}
	defer result.Close(c.Request().Context())
	err = result.All(c.Request().Context(), &allGroups)
	if err != nil {
		fmt.Println("error in getting all groups")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, allGroups)
}

func CreateScrapingGroup(c echo.Context) error {
	dbClient, ok := c.Get("db").(*mongo.Client)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get database client")
	}
	body := map[string]interface{}{}
	if err := c.Bind(&body); err != nil {
		return err
	}

	var newGroup *models.ScrapeGroup

	if name, ok := body["name"].(string); ok {

		newGroup = &models.ScrapeGroup{
			ID:            primitive.NewObjectID(),
			Name:          name,
			Fields:        []models.Field{},
			Endpoints:     []models.Endpoint{},
			WithThumbnail: false,
		}
		result, err := dbClient.Database("scrapeit").Collection("scrape_groups").InsertOne(c.Request().Context(), newGroup)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		fmt.Println(result.InsertedID)
	} else {
		fmt.Println("name is not a string")
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid name type in body",
		})
	}

	return c.JSON(http.StatusOK, newGroup)
}
