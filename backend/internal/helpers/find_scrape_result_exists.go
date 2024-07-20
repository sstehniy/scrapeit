package helpers

import (
	"context"
	"fmt"
	"scrapeit/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	LinkFieldUniqueID = "unique_identifier"
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

	uniqueIdFieldId := ""
	for _, field := range schema {
		if field.Key == LinkFieldUniqueID {
			uniqueIdFieldId = field.ID
		}
	}

	uniqueIdValue := ""
	for _, field := range resultFields {
		switch field.FieldID {
		case uniqueIdFieldId:
			uniqueIdValue = field.Value.(string)
		}
	}

	potentialResultHash := GenerateScrapeResultHash(endpointId + uniqueIdValue)

	query := bson.M{
		"endpointId": endpointId,
		"groupId":    groupId,
		"uniqueHash": potentialResultHash,
	}

	fieldIdsWithLinkType := []string{}
	for _, field := range schema {
		if field.Type == models.FieldTypeLink {
			fieldIdsWithLinkType = append(fieldIdsWithLinkType, field.ID)
		}
	}

	fieldIdsWithImageType := []string{}
	for _, field := range schema {
		if field.Type == models.FieldTypeImage {
			fieldIdsWithImageType = append(fieldIdsWithImageType, field.ID)
		}
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
			isImageTypeValue := false
			for _, fieldId := range fieldIdsWithImageType {
				if fieldId == potentialField.FieldID {
					isImageTypeValue = true
					break
				}
			}
			if isImageTypeValue && potentialField.Value.(string) == "" && field.Value != "" {
				continue
			}

			isLinkTypeValue := false

			for _, fieldId := range fieldIdsWithLinkType {
				if fieldId == potentialField.FieldID {
					isLinkTypeValue = true

				}
			}

			if potentialField.FieldID != uniqueIdFieldId &&
				field.FieldID == potentialField.FieldID &&
				field.Value != potentialField.Value && !isLinkTypeValue {
				fieldName := ""
				for _, schemaField := range schema {
					if schemaField.ID == field.FieldID {
						fieldName = schemaField.Name
					}
				}

				fmt.Printf("Field %s is different: %s != %s\n", fieldName, field.Value, potentialField.Value)

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
