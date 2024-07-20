package scraper

import (
	"context"
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

		time.Sleep(1 * time.Second)

		SlowScrollToBottom(page)
		page.MustWaitStable()

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
		details, err := getElementDetails(element, endpointToScrape.DetailFieldSelectors, relevantGroup.Fields)

		if err != nil {
			return nil, nil, fmt.Errorf("error getting element details: %w", err)
		}
		result := models.ScrapeResult{
			ID:                  primitive.NewObjectID(),
			UniqueHash:          helpers.GenerateScrapeResultHash(endpointToScrape.ID + getFieldValueByFieldKey(relevantGroup.Fields, "unique_identifier", details).(string)),
			EndpointID:          endpointToScrape.ID,
			GroupId:             relevantGroup.ID,
			Fields:              details,
			TimestampInitial:    time.Now().Format(time.RFC3339),
			TimestampLastUpdate: time.Now().Format(time.RFC3339),
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

func getFieldValueByFieldKey(fields []models.Field, fieldKey string, details []models.ScrapeResultDetail) interface{} {
	for _, field := range fields {
		if field.Key == fieldKey {
			for _, detail := range details {
				if detail.FieldID == field.ID {
					return detail.Value
				}
			}
		}
	}
	return ""
}

func getFieldValueByFieldKeyTest(fields []models.Field, fieldKey string, details []models.ScrapeResultDetailTest) interface{} {
	for _, field := range fields {
		if field.Key == fieldKey {
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
		uniqueId := getFieldValueByFieldKey(fields, "unique_identifier", element.Fields)
		fmt.Println("Unique ID: ", uniqueId)
		if uniqueId == "" {
			fmt.Println("Unique ID is empty, skipping")
			continue
		}

		filterResult, err := helpers.FindScrapeResultExists(context.Background(), client, endpointId, groupId, element.Fields, fields)

		if err != nil {
			fmt.Println("Failed to find existing", err)
			continue
		}

		if filterResult.NeedsReplace {
			element.ID = *filterResult.ReplaceID
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

func getElementDetails(element *rod.Element, selectors []models.FieldSelector, fields []models.Field) ([]models.ScrapeResultDetail, error) {
	details := make([]models.ScrapeResultDetail, 0, len(selectors))

	for _, selector := range selectors {
		var text interface{} = ""
		fieldElement, err := element.Element(selector.Selector)

		if err == nil {
			text = fieldElement.MustEval("() => this.textContent").String()
			if strings.TrimSpace(selector.AttributeToGet) != "" {
				if attr, err := fieldElement.Attribute(selector.AttributeToGet); err == nil && attr != nil {
					text = *attr
				}
			}
			if strings.TrimSpace(selector.Regex) != "" {
				if extractedText, _, err := helpers.ExtractStringWithRegex(text.(string), selector.Regex, selector.RegexMatchIndexToUse); err == nil {
					text = extractedText
				}
			}
		}

		if strings.TrimSpace(text.(string)) == "" {
			// Try to get data from the element itself if fieldElement is not found or text is empty
			if strings.TrimSpace(selector.AttributeToGet) != "" {
				if attr, err := element.Attribute(selector.AttributeToGet); err == nil && attr != nil {
					text = *attr
				}
			}
			if strings.TrimSpace(selector.Regex) != "" {
				if extractedText, _, err := helpers.ExtractStringWithRegex(text.(string), selector.Regex, selector.RegexMatchIndexToUse); err == nil {
					text = extractedText
				}
			}
		}

		var relevantFieldType models.FieldType
		for _, field := range fields {
			if field.ID == selector.FieldID {
				relevantFieldType = field.Type
			}
		}
		if relevantFieldType == "number" {
			if text != "" {
				text = helpers.CastPriceStringToFloat(text.(string))
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

OUTER:
	for i := endpointToScrape.PaginationConfig.Start; i <= endpointToScrape.PaginationConfig.End; i++ {
		urlWithPagination := buildPaginationURL(endpointToScrape.URL, endpointToScrape.PaginationConfig, i)
		fmt.Println("Scraping URL: ", urlWithPagination)

		page, err := GetStealthPage(browser, urlWithPagination, endpointToScrape.MainElementSelector)
		if err != nil {
			return nil, nil, fmt.Errorf("error getting page: %w", err)
		}

		defer page.Close()
		time.Sleep(1 * time.Second)

		SlowScrollToBottom(page)
		page.MustWaitStable()

		elements, err := page.Elements(endpointToScrape.MainElementSelector)

		if err != nil {
			return nil, nil, fmt.Errorf("error finding elements: %w", err)
		}

		for _, element := range elements {
			allElements = append(allElements, element)
			if len(allElements) == 5 {
				break OUTER
			}
		}

		// allElements = append(allElements, elements...)
		// if len(allElements) == 5 {
		// 	break
		// }
	}

	linkFieldId := findLinkFieldId(relevantGroup.Fields)
	endpointLinkSelector := findLinkSelector(endpointToScrape.DetailFieldSelectors, linkFieldId)
	fmt.Println("Link selector: ", endpointLinkSelector)

	results := make([]models.ScrapeResultTest, 0, len(allElements))
	for _, element := range allElements {
		details, err := getElementDetailsTest(element, endpointToScrape.DetailFieldSelectors, relevantGroup.Fields)
		uniqueId := getFieldValueByFieldKeyTest(relevantGroup.Fields, "unique_identifier", details).(string)
		fmt.Println("Unique ID: ", uniqueId)
		if uniqueId == "" {
			fmt.Println("Unique ID is empty, skipping")
			continue
		}

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

func getElementDetailsTest(element *rod.Element, selectors []models.FieldSelector, fields []models.Field) ([]models.ScrapeResultDetailTest, error) {
	details := make([]models.ScrapeResultDetailTest, 0, len(selectors))

	for _, selector := range selectors {
		var text interface{} = ""
		var extractMatches []string
		fieldElement, err := element.Element(selector.Selector)

		if err == nil {
			text = fieldElement.MustEval("() => this.textContent").String()
			if strings.TrimSpace(selector.AttributeToGet) != "" {
				if attr, err := fieldElement.Attribute(selector.AttributeToGet); err == nil && attr != nil {
					text = *attr
				}
			}
			if strings.TrimSpace(selector.Regex) != "" {
				fmt.Println("Original text: ", text, selector.Regex)
				if extractedText, matches, err := helpers.ExtractStringWithRegex(text.(string), selector.Regex, selector.RegexMatchIndexToUse); err == nil {
					fmt.Println("Extracted text: ", extractedText)
					text = extractedText
					extractMatches = matches
				}
			}
		}

		if text == "" {
			// Try to get data from the element itself if fieldElement is not found or text is empty
			if strings.TrimSpace(selector.AttributeToGet) != "" {
				if attr, err := element.Attribute(selector.AttributeToGet); err == nil && attr != nil {
					text = *attr
				}
			}
			if strings.TrimSpace(selector.Regex) != "" {
				if extractedText, matches, err := helpers.ExtractStringWithRegex(text.(string), selector.Regex, selector.RegexMatchIndexToUse); err == nil {
					text = extractedText
					extractMatches = matches
				}
			}
		}

		rawData := ""
		if fieldElement != nil {
			rawData = fieldElement.MustHTML()
		}

		var relevantFieldType models.FieldType
		for _, field := range fields {
			if field.ID == selector.FieldID {
				relevantFieldType = field.Type
			}
		}
		if relevantFieldType == "number" {
			if text != "" {
				text = helpers.CastPriceStringToFloat(text.(string))
			}
		}
		detail := models.ScrapeResultDetailTest{
			ID:           uuid.New().String(),
			FieldID:      selector.FieldID,
			Value:        text,
			RawData:      rawData,
			RegexMatches: extractMatches,
		}

		details = append(details, detail)
	}

	return details, nil
}
