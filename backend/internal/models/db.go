package models

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetDbClient() (*mongo.Client, error) {
	// Get MongoDB URI from environment variable
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		return nil, fmt.Errorf("MONGO_URI environment variable is not set")
	}

	fmt.Printf("Attempting to connect to MongoDB at %s...\n", mongoURI)

	// Configure the client to use a longer timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Configure the client options and disable logging
	clientOptions := options.Client().
		ApplyURI(mongoURI).
		SetServerSelectionTimeout(5 * time.Second)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %v", err)
	}

	scrapeResultsCollection := client.Database("scrapeit").Collection("scrape_results")
	err = createFullTextSearchIndex(scrapeResultsCollection, "scrape_results_fulltext")
	if err != nil {
		log.Fatalf("Failed to create full-text search index: %v", err)
	} else {
		fmt.Println("Full-text search index created successfully.")
	}

	archivedScrapeResultsCollection := client.Database("scraper").Collection("archived_scrape_results")
	err = createFullTextSearchIndex(archivedScrapeResultsCollection, "archived_scrape_results_fulltext")

	if err != nil {
		log.Fatalf("Failed to create full-text search index: %v", err)
	} else {
		fmt.Println("Full-text search index created successfully.")
	}

	fmt.Println("Connection established, attempting to ping...")

	// Ping the database
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %v", err)
	} else {
		fmt.Println("Ping successful")
	}

	fmt.Println("Connected to MongoDB!")
	return client, nil
}

func createFullTextSearchIndex(collection *mongo.Collection, indexName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Remove any existing full-text search index
	_, err := collection.Indexes().DropOne(ctx, indexName)
	if err != nil {
		log.Printf("Error dropping existing full-text search index: %v", err)
		// Don't return here, as the index might not exist
	}

	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "fields.value", Value: "text"},
		},
		Options: options.Index().
			SetName(indexName).
			SetDefaultLanguage("english").
			SetLanguageOverride("language"),
	}

	name, err := collection.Indexes().CreateOne(ctx, indexModel, &options.CreateIndexesOptions{})
	if err != nil {
		log.Printf("Error creating full-text search index: %v", err)
		return err
	}

	fmt.Println("Full-text search index created successfully. Bame:", name)

	return nil
}
