package scraper

import (
	"fmt"
	"regexp"
	"scrapeit/internal/helpers"
	"scrapeit/internal/models"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ScrapeEndpointError struct {
	Message string
}

func (e ScrapeEndpointError) Error() string {
	return fmt.Sprintf("scrape endpoint error: %s", e.Message)
}

func ScrapeEndpoint(endpointToScrape models.Endpoint, relevantGroup models.ScrapeGroup, client *mongo.Client, browser *rod.Browser) ([]models.ScrapeResult, []models.ScrapeResult, error) {

	var allElements rod.Elements

	for i := endpointToScrape.PaginationConfig.Start; i <= endpointToScrape.PaginationConfig.End; i++ {
		urlWithPagination := buildPaginationURL(endpointToScrape.URL, endpointToScrape.PaginationConfig, i)
		fmt.Println("Scraping URL: ", urlWithPagination)

		page, err := GetStealthPage(browser, urlWithPagination, endpointToScrape.MainElementSelector)
		if err != nil {
			return nil, nil, fmt.Errorf("error getting page: %w", err)
		}

		defer page.Close()

		SlowScrollToBottom(page)

		elements, err := page.Elements(endpointToScrape.MainElementSelector)

		if err != nil {
			return nil, nil, fmt.Errorf("error finding elements: %w", err)
		}

		allElements = append(allElements, elements...)
	}

	linkFieldId := findLinkFieldId(relevantGroup.Fields)
	endpointLinkSelector := findLinkSelector(endpointToScrape.DetailFieldSelectors, linkFieldId)
	fmt.Println("Link selector: ", endpointLinkSelector)

	results := make([]models.ScrapeResult, 0, len(allElements))
	for _, element := range allElements {
		details, err := getElementDetails(element, endpointToScrape.DetailFieldSelectors)

		if err != nil {
			return nil, nil, fmt.Errorf("error getting element details: %w", err)
		}
		result := models.ScrapeResult{
			ID:         primitive.NewObjectID(),
			UniqueHash: helpers.GenerateScrapeResultHash(getFieldValueByFieldName(relevantGroup.Fields, "link", details)),
			EndpointID: endpointToScrape.ID,
			GroupId:    relevantGroup.ID,
			Fields:     details,
			Timestamp:  time.Now().Format(time.RFC3339),
		}
		results = append(results, result)
	}

	return filterElements(relevantGroup.Fields, results, endpointToScrape.ID, relevantGroup.ID, client)

}

func buildPaginationURL(baseURL string, config models.PaginationConfig, page int) string {
	switch config.Type {
	case "url_parameter":
		return buildURLParameterPagination(baseURL, config.Parameter, page)
	case "url_path":
		return buildURLPathPagination(baseURL, config.Parameter, page, *config.UrlRegexToInsert)
	default:
		return baseURL // Return original URL if type is not recognized
	}
}

func buildURLParameterPagination(baseURL, parameter string, page int) string {
	// Existing implementation for URL parameter pagination
	if strings.Contains(baseURL, "?") {
		if strings.Contains(baseURL, parameter) {
			re := regexp.MustCompile(parameter + `=\d+`)
			return re.ReplaceAllString(baseURL, fmt.Sprintf("%s=%d", parameter, page))
		}
		return fmt.Sprintf("%s&%s=%d", baseURL, parameter, page)
	}
	return fmt.Sprintf("%s?%s=%d", baseURL, parameter, page)
}

func buildURLPathPagination(baseURL, parameter string, page int, urlRegexToInsert string) string {
	if urlRegexToInsert == "" {
		// Default regex if not provided
		urlRegexToInsert = fmt.Sprintf(`(%s:\d+)`, parameter)
	}
	re := regexp.MustCompile(urlRegexToInsert)
	replacement := fmt.Sprintf("%s:%d", parameter, page)

	if re.MatchString(baseURL) {
		return re.ReplaceAllString(baseURL, replacement)
	}

	// If the pattern is not found, insert it before the last path segment
	parts := strings.Split(baseURL, "/")
	if len(parts) > 1 {
		parts[len(parts)-1] = replacement + "/" + parts[len(parts)-1]
	}
	return strings.Join(parts, "/")
}

func findLinkFieldId(fields []models.Field) string {
	for _, field := range fields {
		if field.Type == models.FieldTypeLink {
			return field.ID
		}
	}
	return ""
}

func getFieldValueByFieldName(fields []models.Field, fieldName string, details []models.ScrapeResultDetail) string {
	for _, field := range fields {
		if field.Key == fieldName {
			for _, detail := range details {
				if detail.FieldID == field.ID {
					return detail.Value
				}
			}
		}
	}
	return ""
}

func findLinkSelector(selectors []models.FieldSelector, linkFieldId string) models.FieldSelector {
	for _, selector := range selectors {
		if selector.FieldID == linkFieldId {
			return selector
		}
	}
	return models.FieldSelector{}
}

func filterElements(fields []models.Field, results []models.ScrapeResult, endpointId string, groupId primitive.ObjectID, client *mongo.Client) ([]models.ScrapeResult, []models.ScrapeResult, error) {

	var filtered []models.ScrapeResult
	var toReplace []models.ScrapeResult
	for _, element := range results {

		filterResult, err := helpers.FindScrapeResultExists(client, endpointId, groupId, element.Fields, fields)

		if err != nil {
			fmt.Println("Failed to find existing", err)
			continue
		}

		if filterResult.NeedsReplace {
			toReplace = append(toReplace, element)
		} else if !filterResult.Exists {
			filtered = append(filtered, element)
		}
	}

	toReplaceIds := make([]string, 0, len(toReplace))
	for _, r := range toReplace {
		toReplaceIds = append(toReplaceIds, r.ID.Hex())
	}

	fmt.Println("Filtered results: ", len(filtered))
	fmt.Println("To replace results: ", toReplaceIds)

	return filtered, toReplace, nil
}

func getElementDetails(element *rod.Element, selectors []models.FieldSelector) ([]models.ScrapeResultDetail, error) {
	details := make([]models.ScrapeResultDetail, 0, len(selectors))
	for _, selector := range selectors {
		fieldElement, err := element.Element(selector.Selector)
		if err != nil {

			fmt.Printf("Error finding element for selector %s: %v\n", selector.Selector, err)
			details = append(details, models.ScrapeResultDetail{
				ID:      uuid.New().String(),
				FieldID: selector.FieldID,
				Value:   "",
			})
			continue
		}
		text := fieldElement.MustEval("() => this.textContent").String()

		if selector.AttributeToGet != "" {
			attr, err := fieldElement.Attribute(selector.AttributeToGet)
			if err == nil {
				text = *attr
			}
		}
		if selector.Regex != "" {
			text, err = helpers.ExtractStringWithRegex(text, selector.Regex)
			if err != nil {
				// fmt.Printf("Error extracting regex for selector %s: %v\n", selector.Selector, err)
				text = ""
			}
		}

		detail := models.ScrapeResultDetail{
			ID:      uuid.New().String(),
			FieldID: selector.FieldID,
			Value:   text,
		}
		details = append(details, detail)
	}
	return details, nil
}

type ScrapeEndpointTestError struct {
	Message string
}

func (e ScrapeEndpointTestError) Error() string {
	return fmt.Sprintf("scrape endpoint test error: %s", e.Message)
}

func ScrapeEndpointTest(endpointToScrape models.Endpoint, relevantGroup models.ScrapeGroup, client *mongo.Client, browser *rod.Browser) ([]models.ScrapeResultTest, []models.ScrapeResultTest, error) {
	fmt.Println(("Scraping endpoint test"))

	var allElements rod.Elements

	for i := endpointToScrape.PaginationConfig.Start; i <= endpointToScrape.PaginationConfig.End; i++ {
		urlWithPagination := buildPaginationURL(endpointToScrape.URL, endpointToScrape.PaginationConfig, i)
		fmt.Println("Scraping URL: ", urlWithPagination)

		page, err := GetStealthPage(browser, urlWithPagination, endpointToScrape.MainElementSelector)
		if err != nil {
			return nil, nil, fmt.Errorf("error getting page: %w", err)
		}

		defer page.Close()

		SlowScrollToBottom(page)

		elements, err := page.Elements(endpointToScrape.MainElementSelector)

		if err != nil {
			return nil, nil, fmt.Errorf("error finding elements: %w", err)
		}

		allElements = append(allElements, elements...)
	}

	linkFieldId := findLinkFieldId(relevantGroup.Fields)
	endpointLinkSelector := findLinkSelector(endpointToScrape.DetailFieldSelectors, linkFieldId)
	fmt.Println("Link selector: ", endpointLinkSelector)

	results := make([]models.ScrapeResultTest, 0, len(allElements))
	for _, element := range allElements {
		details, err := getElementDetailsTest(element, endpointToScrape.DetailFieldSelectors)

		if err != nil {
			return nil, nil, fmt.Errorf("error getting element details: %w", err)
		}
		result := models.ScrapeResultTest{
			ID:         primitive.NewObjectID(),
			UniqueHash: "",
			EndpointID: endpointToScrape.ID,
			GroupId:    relevantGroup.ID,
			Fields:     details,
			Timestamp:  time.Now().Format(time.RFC3339),
		}
		results = append(results, result)
	}

	return results, nil, nil
}

func getElementDetailsTest(element *rod.Element, selectors []models.FieldSelector) ([]models.ScrapeResultDetailTest, error) {
	details := make([]models.ScrapeResultDetailTest, 0, len(selectors))
	for _, selector := range selectors {
		fieldElement, err := element.Element(selector.Selector)
		if err != nil {

			fmt.Printf("Error finding element for selector %s: %v\n", selector.Selector, err)
			details = append(details, models.ScrapeResultDetailTest{
				ID:      uuid.New().String(),
				FieldID: selector.FieldID,
				Value:   "",
				RawData: "",
			})
			continue
		}
		text := fieldElement.MustEval("() => this.textContent").String()

		if selector.AttributeToGet != "" {
			attr, err := fieldElement.Attribute(selector.AttributeToGet)
			if err == nil {
				text = *attr
			}
		}
		if selector.Regex != "" {
			text, err = helpers.ExtractStringWithRegex(text, selector.Regex)
			if err != nil {
				// fmt.Printf("Error extracting regex for selector %s: %v\n", selector.Selector, err)
				text = ""
			}
		}

		detail := models.ScrapeResultDetailTest{
			ID:      uuid.New().String(),
			FieldID: selector.FieldID,
			Value:   text,
			RawData: fieldElement.MustHTML(),
		}
		details = append(details, detail)
	}

	return details, nil
}
