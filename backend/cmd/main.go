package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"scrapeit/internal/cron"
	"scrapeit/internal/handlers"
	"scrapeit/internal/models"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.mongodb.org/mongo-driver/bson"
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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
	}, 4, 4)

	e := echo.New()

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:3456"},
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

	internal := e.Group("/internal")
	internal.POST("/scrape/endpoint", handlers.ScrapeEndpointHandler)

	// Home route
	api.GET("/", handlers.Home)

	// Scrape routes
	scrape := api.Group("/scrape")
	scrape.GET("/test", handlers.Scrape)
	scrape.POST("/results", handlers.GetScrapingResults)
	scrape.GET("/results/not-empty/:groupId", handlers.GetScrapingResultsNotEmpty)
	scrape.POST("/results/export/:groupId", handlers.ExportGroupResultsHandler)
	scrape.POST("/endpoints", handlers.ScrapeEndpointsHandler)
	scrape.POST("/endpoint-test", handlers.ScrapeEndpointTestHandler)

	// Scrape groups routes
	groups := api.Group("/scrape-groups")
	groups.GET("", handlers.GetScrapingGroups)
	groups.POST("/import", handlers.ImportScrapingGroup)
	groups.GET("/archived", handlers.GetArchivedScrapingGroups)
	groups.POST("", handlers.CreateScrapingGroup)
	groups.GET("/:id", handlers.GetScrapingGroup)
	groups.DELETE("/:id", handlers.DeleteScrapingGroup)
	groups.PUT("/:groupId/schema", handlers.UpdateScrapingGroupSchema)
	groups.GET("/version-tag-exists/:versionTag", handlers.VersionTagExists)
	groups.GET("/:id/notification-config", handlers.GetScrapingGroupNotificationConfig)
	groups.PUT("/:id/notification-config", handlers.ChangeScrapingGroupNotificationConfig)

	// Endpoints within scrape groups
	groups.POST("/:groupId/endpoints", handlers.CreateScrapingGroupEndpoint)
	groups.DELETE("/:groupId/endpoints/:endpointId", handlers.DeleteScrapingGroupEndpoint)
	groups.DELETE("/:groupId/endpoints/results/:endpointId", handlers.DeleteScrapingGroupEndpointResults)
	groups.PUT("/:groupId/endpoints/:endpointId", handlers.UpdateScrapingGroupEndpoint)

	// Selector routes
	selectors := api.Group("/selectors")
	selectors.POST("/extract", handlers.ExtractSelectorsHandler)
	selectors.POST("/test", handlers.ElementSelectorTestHandler)

	ai := api.Group("/ai")
	ai.POST("/completion", handlers.CompletionHandler)
	fmt.Println("Starting server on port 3457")
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
	if err := e.Start(":3457"); err != nil && err != http.ErrServerClosed {
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

	defer cursor.Close(context.Background())

	if err := cursor.All(context.Background(), &groups); err != nil {
		fmt.Println("Error getting groups:", err)
		return
	}
	for _, group := range groups {
		for _, endpoint := range group.Endpoints {
			if endpoint.Active {
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
}
