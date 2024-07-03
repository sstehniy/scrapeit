package main

import (
	"scrapeit/internal/handlers"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:5173"},
	}))
	e.Use(middleware.Recover())

	// Routes
	e.GET("/", handlers.Home)
	e.GET("/scrape", handlers.Scrape)
	e.GET("/scrape-groups", handlers.GetScrapingGroups)
	e.POST("/scrape-groups", handlers.CreateScrapingGroup)

	e.Logger.Fatal(e.Start(":8080"))
}

func connectToDb() error
