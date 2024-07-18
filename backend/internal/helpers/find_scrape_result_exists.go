package helpers

import (
	"context"
	"scrapeit/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	LinkFieldKey  = "link"
	ImageFieldKey = "image"
)

type FindScrapeResultExistsResult struct {
	Exists       bool
	NeedsReplace bool
	ReplaceID    *primitive.ObjectID
}

type ResultDetailFieldWithoutId struct {
	FieldID string `json:"fieldId" bson:"fieldId"`
	Value   string `json:"value" bson:"value"`
}

// FindScrapeResultExists checks if a scrape result already exists in the database.
func FindScrapeResultExists(ctx context.Context, client *mongo.Client, endpointId string, groupId primitive.ObjectID, resultFields []models.ScrapeResultDetail, schema []models.Field) (FindScrapeResultExistsResult, error) {
	result := FindScrapeResultExistsResult{
		Exists:       true,
		NeedsReplace: false,
	}

	linkFieldId, imageFieldId := "", ""
	for _, field := range schema {
		switch field.Key {
		case LinkFieldKey:
			linkFieldId = field.ID
		case ImageFieldKey:
			imageFieldId = field.ID
		}
	}

	linkValue, imageValue := "", ""
	for _, field := range resultFields {
		switch field.FieldID {
		case linkFieldId:
			linkValue = field.Value
		case imageFieldId:
			imageValue = field.Value
		}
	}

	potentialResultHashByLink := GenerateScrapeResultHash(linkValue)

	query := bson.M{
		"endpointId": endpointId,
		"groupId":    groupId,
		"$or": []bson.M{
			{"uniqueHash": potentialResultHashByLink},
			{
				"fields": bson.M{
					"$all": []bson.M{
						{
							"$elemMatch": bson.M{
								"fieldId": imageFieldId,
								"value":   imageValue,
							},
						},
						{
							"$elemMatch": bson.M{
								"fieldId": linkFieldId,
								"value":   linkValue,
							},
						},
					},
				},
			},
		},
	}

	collection := client.Database("scrapeit").Collection("scrape_results")

	singleResult := collection.FindOne(ctx, query)
	if singleResult.Err() == mongo.ErrNoDocuments {
		result.Exists = false
		return result, nil
	} else if singleResult.Err() != nil {
		return result, singleResult.Err()
	}

	var scrapeResult models.ScrapeResult
	if err := singleResult.Decode(&scrapeResult); err != nil {
		return result, err
	}

	areSame := true
	for _, field := range scrapeResult.Fields {
		for _, potentialField := range resultFields {
			if potentialField.FieldID != linkFieldId &&
				field.FieldID == potentialField.FieldID &&
				field.Value != potentialField.Value {
				areSame = false
				break
			}
		}
	}

	result.NeedsReplace = !areSame
	if !areSame {
		result.ReplaceID = &scrapeResult.ID
	}

	return result, nil
}
