package models

import (
	"context"
	"fmt"
	"os"
	"time"

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
		SetServerSelectionTimeout(5 * time.Second).
		SetMonitor(nil)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %v", err)
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
