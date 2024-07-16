package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"scrapeit/internal/cron"
	"scrapeit/internal/handlers"
	"scrapeit/internal/helpers"
	"scrapeit/internal/models"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// SimpleLogger is a basic implementation of the Logger interface
type SimpleLogger struct {
	logger *log.Logger
}

func (l *SimpleLogger) Info(msg string, keysAndValues ...interface{}) {
	l.logger.Printf("INFO: %s %v", msg, keysAndValues)
}

func (l *SimpleLogger) Error(msg string, keysAndValues ...interface{}) {
	l.logger.Printf("ERROR: %s %v", msg, keysAndValues)
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5)
	defer cancel()
	DbClient, err := models.GetDbClient()
	fmt.Println("Connected to MongoDB")
	if err != nil {
		panic(err)
	}

	defer func() {
		fmt.Println("Disconnecting from MongoDB")
		if err := DbClient.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	// remove the log file if it exists
	if _, err := os.Stat("/app/logs/cron.log"); err == nil {
		if err := os.Remove("/app/logs/cron.log"); err != nil {
			fmt.Println("Error removing log file:", err)
		}
	}

	logFile, err := os.OpenFile("/app/logs/cron.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	if err != nil {
		fmt.Println("Error opening log file:", err)
		return
	}

	cronManager := cron.NewCronManager(&SimpleLogger{
		logger: log.New(logFile, "", log.LstdFlags),
	})

	e := echo.New()

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:5173"},
	}))
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}\n",
	}))
	e.Use(middleware.Recover())
	// pass the db client to the context
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("db", DbClient)
			c.Set("cron", cronManager)
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
	scrape.GET("/results/not-empty/:groupId", handlers.GetScrapingResultsNotEmpty)
	scrape.POST("/endpoints", handlers.ScrapeEndpointsHandler)
	scrape.POST("/endpoint", handlers.ScrapeEndpointHandler)
	scrape.POST("/endpoint-test", handlers.ScrapeEndpointTestHandler)

	// Scrape groups routes
	groups := api.Group("/scrape-groups")
	groups.GET("", handlers.GetScrapingGroups)
	groups.POST("", handlers.CreateScrapingGroup)
	groups.GET("/:id", handlers.GetScrapingGroup)
	groups.PUT("/:groupId/schema", handlers.UpdateScrapingGroupSchema)
	groups.GET("/version-tag-exists/:versionTag", handlers.VersionTagExists)

	// Endpoints within scrape groups
	groups.POST("/:groupId/endpoints", handlers.CreateScrapingGroupEndpoint)
	groups.DELETE("/:groupId/endpoints/:endpointId", handlers.DeleteScrapingGroupEndpoint)
	groups.PUT("/:groupId/endpoints/:endpointId", handlers.UpdateScrapingGroupEndpoint)

	// Selector routes
	selectors := api.Group("/selectors")
	selectors.POST("/extract", handlers.ExtractSelectorsHandler)
	selectors.POST("/test", handlers.ElementSelectorTestHandler)

	ai := api.Group("/ai")
	ai.POST("/completion", handlers.CompletionHandler)
	fmt.Println("Starting server on port 8080")
	fmt.Println("Setting up cron jobs")
	setupCronJobs(e, cronManager, DbClient)
	fmt.Println("Cron jobs set up")

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
		<-quit
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := e.Shutdown(ctx); err != nil {
			e.Logger.Fatal(err)
		}
	}()

	// Start server
	if err := e.Start(":8080"); err != nil && err != http.ErrServerClosed {
		e.Logger.Fatal(err)
	}

	// Perform cleanup actions here
	fmt.Println("Server is shutting down...")

	cronManager.Stop()

	if err := DbClient.Disconnect(context.Background()); err != nil {
		e.Logger.Error(err)
	}
	fmt.Println("Cleanup completed. Goodbye!")
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

	archivedResultsCollection := client.Database("scrapeit").Collection("archived_scrape_results")
	if err := archivedResultsCollection.Drop(context.Background()); err != nil {
		fmt.Println("Error dropping archived results collection:", err)
		return err
	}

	byteValue, err := helpers.ReadJson("/internal/data/scraping_groups.json")
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

func setupCronJobs(e *echo.Echo, cronManager *cron.CronManager, client *mongo.Client) {
	collection := client.Database("scrapeit").Collection("scrape_groups")
	groups := []models.ScrapeGroup{}
	cursor, err := collection.Find(context.Background(), bson.M{
		"versionTag": "",
	})
	if err != nil {
		fmt.Println("Error getting groups:", err)
		return
	}
	if err := cursor.All(context.Background(), &groups); err != nil {
		fmt.Println("Error getting groups:", err)
		return
	}
	for _, group := range groups {
		for _, endpoint := range group.Endpoints {

			cronManager.AddJob(cron.CronManagerJob{
				GroupID:    group.ID.Hex(),
				EndpointID: endpoint.ID,
				Active:     true,
				Interval:   endpoint.Interval,

				Job: func() error {
					fmt.Println("Running job for", group.ID.Hex(), endpoint.ID)
					return handlers.HandleCallInternalScrapeEndpoint(e, group.ID.Hex(), endpoint.ID, client, cronManager)
				},
			})
		}
	}
}
