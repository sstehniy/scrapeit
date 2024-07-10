package utils

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func FindScrapeResultExists(potentialResultUrl string, endpointId string, groupId primitive.ObjectID, client *mongo.Client) bool {
	potentialResultHash := GenerateScrapeResultHash(potentialResultUrl)

	fmt.Println("+++++++++++++++++")
	fmt.Println(fmt.Sprintf("Url: %s, Checking if scrape result exists for hash: %s", potentialResultUrl, potentialResultHash))
	fmt.Println("+++++++++++++++++")
	query := bson.M{
		"uniqueHash": potentialResultHash,
		"endpointId": endpointId,
		"groupId":    groupId,
	}
	collection := client.Database("scrapeit").Collection("scrape_results")

	count, err := collection.CountDocuments(context.Background(), query)
	if err != nil {
		return false
	}
	exists := count > 0

	return exists
}
