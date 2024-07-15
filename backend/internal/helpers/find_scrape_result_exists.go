package helpers

import (
	"context"
	"fmt"
	"scrapeit/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FindScrapeResultExistsResult struct {
	Exists       bool
	NeedsReplace bool
}

type ResultDetailFieldWithoutId struct {
	FieldID string `json:"fieldId" bson:"fieldId"`
	Value   string `json:"value" bson:"value"`
}

func FindScrapeResultExists(client *mongo.Client, endpointId string, groupId primitive.ObjectID, resultFields []models.ScrapeResultDetail, schema []models.Field) (FindScrapeResultExistsResult, error) {
	result := FindScrapeResultExistsResult{
		Exists:       true,
		NeedsReplace: false,
	}

	linkFieldId := ""
	imageFieldId := ""
	for _, field := range schema {
		if field.Key == "link" {
			linkFieldId = field.ID
		}
		if field.Key == "image" {
			imageFieldId = field.ID
		}
	}

	linkValue := ""
	imageValue := ""

	for _, field := range resultFields {
		if field.FieldID == linkFieldId {
			linkValue = field.Value
		}
		if field.FieldID == imageFieldId {
			imageValue = field.Value
		}
	}

	potentialResultHashByLink := GenerateScrapeResultHash(linkValue)

	fmt.Printf("Image id: %s, Image value: %s\n", imageFieldId, imageValue)
	fmt.Printf("Link id: %s, Link value: %s\n", linkFieldId, linkValue)

	query := bson.M{
		"endpointId": endpointId,
		"groupId":    groupId,
		"$or": []bson.M{
			{"uniqueHash": potentialResultHashByLink},
			{
				"fields": bson.M{
					"$elemMatch": bson.M{
						"fieldId": imageFieldId,
						"value":   imageValue,
					},
				},
			},
		},
	}

	collection := client.Database("scrapeit").Collection("scrape_results")

	singleResult := collection.FindOne(context.Background(), query)
	if singleResult.Err() == mongo.ErrNoDocuments {
		fmt.Printf("%s not found", potentialResultHashByLink)
		result.Exists = false
		return result, nil
	} else if singleResult.Err() != nil {
		return result, singleResult.Err()
	}

	fmt.Println("Found existing scrape result")

	var scrapeResult models.ScrapeResult
	err := singleResult.Decode(&scrapeResult)
	if err != nil {
		return result, err
	}

	areSame := true
	for _, field := range scrapeResult.Fields {
		for _, potentialField := range resultFields {
			if field.FieldID == potentialField.FieldID && field.Value != potentialField.Value && field.FieldID != imageFieldId && field.FieldID != linkFieldId {
				fmt.Println("Field", field.FieldID, "is different")
				areSame = false
				break
			}

		}
	}

	result.NeedsReplace = !areSame

	return result, nil
}
