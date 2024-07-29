package handlers

import (
	"context"
	"fmt"
	"net/http"
	"scrapeit/internal/cron"
	"scrapeit/internal/helpers"
	"scrapeit/internal/models"
	"scrapeit/internal/scraper"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ScraperEndpointsHandlerRequest struct {
	EndpointIds []string `json:"endpointIds"`
	GroupId     string   `json:"groupId"`
}

type ScraperEndpointsHandlerResponse struct {
	NewResults      []models.ScrapeResult `json:"newResults"`
	ReplacedResults []models.ScrapeResult `json:"replacedResults"`
}

func ScrapeEndpointsHandler(c echo.Context) error {
	var body ScraperEndpointsHandlerRequest
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	dbClient, _ := models.GetDbClient()

	cronManager := cron.GetCronManager()

	stopRunningJobs(cronManager, body.GroupId, body.EndpointIds)
	defer startJobs(cronManager, body.GroupId, body.EndpointIds)

	group, err := getScrapeGroup(c, dbClient, body.GroupId)
	if err != nil {
		return err
	}

	endpointsToScrape := filterIdleEndpoints(group, body.EndpointIds)
	if len(endpointsToScrape) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "No idle endpoints to scrape"})
	}

	if err := updateEndpointStatuses(dbClient, group.ID, endpointsToScrape, models.ScrapeStatusRunning); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update group"})
	}
	defer updateEndpointStatuses(dbClient, group.ID, endpointsToScrape, models.ScrapeStatusIdle)

	results, toReplaceResults := scrapeEndpoints(c, dbClient, group, endpointsToScrape)

	if err := insertNewResults(dbClient, results); err != nil {
		fmt.Printf("Failed to insert new results %v\n", err)
	}

	if err := updateExistingResults(c, dbClient, toReplaceResults); err != nil {
		fmt.Printf("Failed to update existing results: %v\n", err)
	}

	go func() {
		if err := notifyResults(c, dbClient, group, results, toReplaceResults); err != nil {
			fmt.Printf("Failed to notify results: %v\n", err)
		}

	}()
	return c.JSON(http.StatusOK, ScraperEndpointsHandlerResponse{
		NewResults:      results,
		ReplacedResults: toReplaceResults,
	})
}

func stopRunningJobs(cronManager *cron.CronManager, groupId string, endpointIds []string) {
	for _, id := range endpointIds {
		existingJob := cronManager.GetJob(groupId, id)
		if existingJob != nil && existingJob.Status == cron.CronJobStatusRunning {
			continue
		}
		cronManager.StopJob(groupId, id)
	}
}

func startJobs(cronManager *cron.CronManager, groupId string, endpointIds []string) {
	for _, id := range endpointIds {
		cronManager.StartJob(groupId, id)
	}
}

func getScrapeGroup(c echo.Context, dbClient *mongo.Client, groupId string) (*models.ScrapeGroup, error) {
	groupCollection := dbClient.Database("scrapeit").Collection("scrape_groups")
	groupObjectId, err := primitive.ObjectIDFromHex(groupId)
	if err != nil {
		return nil, c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	groupQuery := bson.M{"_id": groupObjectId}
	groupResult := groupCollection.FindOne(c.Request().Context(), groupQuery)
	if err := groupResult.Err(); err != nil {
		return nil, c.JSON(http.StatusBadRequest, map[string]string{"error": groupResult.Err().Error()})
	}

	var group models.ScrapeGroup
	if err := groupResult.Decode(&group); err != nil {
		return nil, c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	return &group, nil
}

func filterIdleEndpoints(group *models.ScrapeGroup, endpointIds []string) []*models.Endpoint {
	var endpoints []*models.Endpoint
	for _, endpointId := range endpointIds {
		endpoint := group.GetEndpointById(endpointId)
		if endpoint != nil && endpoint.Status == models.ScrapeStatusIdle {
			endpoints = append(endpoints, endpoint)
		}
	}
	return endpoints
}

func updateEndpointStatuses(dbClient *mongo.Client, groupID primitive.ObjectID, endpoints []*models.Endpoint, status models.ScrapeStatus) error {
	groupCollection := dbClient.Database("scrapeit").Collection("scrape_groups")

	var updates []mongo.WriteModel
	for _, endpoint := range endpoints {
		// updates = append(updates, bson.M{
		// 	"q": bson.M{"_id": groupID, "endpoints.id": endpoint.ID},
		// 	"u": bson.M{"$set": bson.M{"endpoints.$.status": status}},
		// })
		update := mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": groupID, "endpoints.id": endpoint.ID}).
			SetUpdate(bson.M{"$set": bson.M{"endpoints.$.status": status}})
		updates = append(updates, update)
	}

	_, err := groupCollection.BulkWrite(context.TODO(), updates, options.BulkWrite().SetOrdered(false))
	if err != nil {
		return fmt.Errorf("failed to update endpoint statuses: %w", err)
	}

	// Verify the update
	groupResult := groupCollection.FindOne(context.TODO(), bson.M{"_id": groupID})
	var groupUpdated models.ScrapeGroup
	if err := groupResult.Decode(&groupUpdated); err != nil {
		return fmt.Errorf("failed to retrieve updated group: %w", err)
	}

	for _, endpoint := range groupUpdated.Endpoints {
		fmt.Printf("Endpoint %s status is %s\n", endpoint.ID, endpoint.Status)
	}

	return nil
}
func updateGroupEndpoints(dbClient *mongo.Client, group *models.ScrapeGroup) error {

	groupCollection := dbClient.Database("scrapeit").Collection("scrape_groups")
	_, err := groupCollection.UpdateOne(context.TODO(), bson.M{"_id": group.ID}, bson.M{"$set": bson.M{"endpoints": group.Endpoints}})
	// test get and see ednpoints status
	groupResult := groupCollection.FindOne(context.TODO(), bson.M{"_id": group.ID})
	var groupUpdated models.ScrapeGroup
	if err := groupResult.Decode(&groupUpdated); err != nil {
		return err
	}

	for _, endpoint := range groupUpdated.Endpoints {
		fmt.Println("Status is ", endpoint.Status)
	}

	return err
}

func scrapeEndpoints(c echo.Context, dbClient *mongo.Client, group *models.ScrapeGroup, endpoints []*models.Endpoint) ([]models.ScrapeResult, []models.ScrapeResult) {
	resultsChan := make(chan []models.ScrapeResult)
	toReplaceChan := make(chan []models.ScrapeResult)
	browser := scraper.GetBrowser()

	for _, endpoint := range endpoints {
		go func(endpoint models.Endpoint) {
			results, toReplace, err := scraper.ScrapeEndpoint(endpoint, *group, dbClient, browser)
			if err != nil {
				resultsChan <- nil
				toReplaceChan <- nil
			} else {
				resultsChan <- results
				toReplaceChan <- toReplace
			}
		}(*endpoint)
	}

	var results []models.ScrapeResult
	var toReplaceResults []models.ScrapeResult
	for range endpoints {
		results = append(results, <-resultsChan...)
		toReplaceResults = append(toReplaceResults, <-toReplaceChan...)
	}
	return results, toReplaceResults
}

func insertNewResults(dbClient *mongo.Client, results []models.ScrapeResult) error {
	seenUniqueHashes := make(map[string]bool)
	var toInsert []interface{}

	for _, r := range results {
		if seenUniqueHashes[r.UniqueHash] {
			continue
		}
		seenUniqueHashes[r.UniqueHash] = true
		r.ID = primitive.NewObjectID()
		toInsert = append(toInsert, r)
	}

	allResultsCollection := dbClient.Database("scrapeit").Collection("scrape_results")
	_, err := allResultsCollection.InsertMany(context.TODO(), toInsert, &options.InsertManyOptions{})
	return err
}

func updateExistingResults(c echo.Context, dbClient *mongo.Client, toReplaceResults []models.ScrapeResult) error {
	var bulkWrites []mongo.WriteModel
	for _, r := range toReplaceResults {
		update := mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": r.ID}).
			SetUpdate(bson.M{"$set": bson.M{
				"fields":              r.Fields,
				"timestampLastUpdate": r.TimestampLastUpdate,
			}})
		bulkWrites = append(bulkWrites, update)
	}

	allResultsCollection := dbClient.Database("scrapeit").Collection("scrape_results")
	_, err := allResultsCollection.BulkWrite(c.Request().Context(), bulkWrites)
	return err
}

func notifyResults(c echo.Context, dbClient *mongo.Client, group *models.ScrapeGroup, results, toReplaceResults []models.ScrapeResult) error {
	notificationConfigs, err := getNotificationConfigs(c, dbClient, group.ID)
	if err != nil {
		return err
	}

	if len(notificationConfigs) > 0 {
		go helpers.HandleNotifyResults(notificationConfigs, *group, results, toReplaceResults)
	}
	return nil
}

func getNotificationConfigs(c echo.Context, dbClient *mongo.Client, groupId primitive.ObjectID) ([]models.NotificationConfig, error) {
	var notificationConfigs []models.NotificationConfig
	notificationConfigResult, err := dbClient.Database("scrapeit").Collection("notification_configs").Find(c.Request().Context(), bson.M{"groupId": groupId})
	if err != nil {
		return nil, fmt.Errorf("failed to get notification configs: %w", err)
	}
	defer notificationConfigResult.Close(context.Background())

	for notificationConfigResult.Next(context.Background()) {
		var config models.NotificationConfig
		if err := notificationConfigResult.Decode(&config); err != nil {
			return nil, fmt.Errorf("failed to decode config: %w", err)
		}
		notificationConfigs = append(notificationConfigs, config)
	}
	return notificationConfigs, nil
}
