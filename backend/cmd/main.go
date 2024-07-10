package main

import (
	"context"
	"encoding/json"
	"fmt"
	"scrapeit/internal/handlers"
	"scrapeit/internal/models"
	"scrapeit/internal/utils"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5)
	defer cancel()
	DbClient, err := models.GetDbClient()
	if err != nil {
		panic(err)
	}

	defer func() {
		fmt.Println("Disconnecting from MongoDB")
		if err := DbClient.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:5173"},
	}))
	e.Use(middleware.Recover())
	// pass the db client to the context
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("db", DbClient)
			return next(c)
		}
	})

	// prepopulateScrapeGroups(DbClient)

	api := e.Group("")

	// Home route
	e.GET("/", handlers.Home)

	// Scrape routes
	scrape := api.Group("/scrape")
	scrape.GET("/results/:groupId", handlers.GetScrapingResults)
	scrape.POST("/endpoints", handlers.ScrapeEndpointsHandler)
	scrape.POST("/endpoint", handlers.ScrapeEndpointHandler)
	scrape.POST("/endpoint-test", handlers.ScrapeEndpointTestHandler)

	// Scrape groups routes
	groups := api.Group("/scrape-groups")
	groups.GET("", handlers.GetScrapingGroups)
	groups.POST("", handlers.CreateScrapingGroup)
	groups.GET("/:id", handlers.GetScrapingGroup)
	groups.POST("/:id/fields", handlers.UpdateScrapingGroupSchema)

	// Endpoints within scrape groups
	groups.POST("/:groupId/endpoints", handlers.CreateScrapingGroupEndpoint)
	groups.PUT("/:groupId/endpoints/:endpointId", handlers.UpdateScrapingGroupEndpoint)

	// Selector routes
	selectors := api.Group("/selectors")
	selectors.POST("/extract", handlers.ExtractSelectorsHandler)
	selectors.POST("/test", handlers.ElementSelectorTestHandler)

	e.Logger.Fatal(e.Start(":8080"))
}

type ScrapeGroupLocal struct {
}

func prepopulateScrapeGroups(client *mongo.Client) error {
	// drop the collection
	collection := client.Database("scrapeit").Collection("scrape_groups")
	if err := collection.Drop(context.Background()); err != nil {
		fmt.Println("Error dropping collection:", err)
		return err
	}

	resultsCollection := client.Database("scrapeit").Collection("scrape_results")
	if err := resultsCollection.Drop(context.Background()); err != nil {
		fmt.Println("Error dropping results collection:", err)
		return err
	}

	byteValue, err := utils.ReadJson("/internal/data/scraping_groups.json")
	if err != nil {
		fmt.Println("Error reading JSON groups:", err)
		return err
	}
	var groups []models.ScrapeGroupLocal
	if err := json.Unmarshal(byteValue, &groups); err != nil {
		fmt.Println("Error unmarshaling JSON bytes:", err)
		return err
	}
	collection = client.Database("scrapeit").Collection("scrape_groups")
	for _, group := range groups {
		groupDb := models.ScrapeGroup{
			ID:        primitive.NewObjectID(),
			Name:      group.Name,
			Fields:    group.Fields,
			Endpoints: group.Endpoints,
		}
		_, err := collection.InsertOne(context.Background(), groupDb)
		if err != nil {
			fmt.Println("Error inserting group:", err)
			return err
		} else {
			fmt.Println("Group inserted successfully", group.Name)
		}

	}
	return nil
}
