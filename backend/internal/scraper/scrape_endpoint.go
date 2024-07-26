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
	var results []models.ScrapeResult
	scrapeType := GetScrapeType(endpointToScrape)

	switch scrapeType {
	case PureDetails:
		page, err := GetStealthPage(browser, endpointToScrape.URL, endpointToScrape.DetailedViewMainElementSelector)
		if err != nil {
			return nil, nil, fmt.Errorf("error getting page: %w", err)
		}
		defer page.Close()

		SlowScrollToBottom(page)
		page.MustWaitStable()

		elements, err := getMainElements(page, endpointToScrape, scrapeType, 1)
		if err != nil {
			return nil, nil, fmt.Errorf("error finding elements: %w", err)
		}

		scraped, err := processElements(elements, endpointToScrape, relevantGroup)
		if err != nil {
			return nil, nil, fmt.Errorf("error processing elements: %w", err)
		}
		results = scraped
	case Previews:
		scraped, err := scrapePreviewsPages(endpointToScrape, relevantGroup, browser)
		if err != nil {
			return nil, nil, fmt.Errorf("error scraping previews pages: %w", err)
		}
		results = scraped

	case PreviewsWithDetails:
		scraped, err := scrapePreviewsWithDetails(endpointToScrape, relevantGroup, browser)
		if err != nil {
			return nil, nil, fmt.Errorf("error scraping previews with details: %w", err)
		}
		results = scraped

	default:
		return nil, nil, fmt.Errorf("unknown scrape type: %v", scrapeType)
	}

	return filterElements(relevantGroup.Fields, results, endpointToScrape.ID, relevantGroup.ID, client)
}

func scrapePreviewsPages(endpointToScrape models.Endpoint, relevantGroup models.ScrapeGroup, browser *rod.Browser) ([]models.ScrapeResult, error) {
	var allElements []PageData

	for i := endpointToScrape.PaginationConfig.Start; i <= endpointToScrape.PaginationConfig.End; i += endpointToScrape.PaginationConfig.Step {
		urlWithPagination := buildPaginationURL(endpointToScrape.URL, endpointToScrape.PaginationConfig, i)
		fmt.Println("Scraping URL: ", urlWithPagination)

		page, err := GetStealthPage(browser, urlWithPagination, endpointToScrape.MainElementSelector)
		if err != nil {
			return nil, fmt.Errorf("error getting page: %w", err)
		}

		defer page.Close()

		SlowScrollToBottom(page)
		page.MustWaitStable()

		elements, err := getMainElements(page, endpointToScrape, Previews, -1)
		if err != nil {
			return nil, fmt.Errorf("error finding elements: %w", err)
		}

		allElements = append(allElements, elements...)
	}

	return processElements(allElements, endpointToScrape, relevantGroup)
}

func scrapePreviewsWithDetails(endpointToScrape models.Endpoint, relevantGroup models.ScrapeGroup, browser *rod.Browser) ([]models.ScrapeResult, error) {
	var results []models.ScrapeResult

	for i := endpointToScrape.PaginationConfig.Start; i <= endpointToScrape.PaginationConfig.End; i += endpointToScrape.PaginationConfig.Step {
		urlWithPagination := buildPaginationURL(endpointToScrape.URL, endpointToScrape.PaginationConfig, i)
		fmt.Println("Scraping URL: ", urlWithPagination)

		page, err := GetStealthPage(browser, urlWithPagination, endpointToScrape.MainElementSelector)
		if err != nil {
			return nil, fmt.Errorf("error getting page: %w", err)
		}

		defer page.Close()

		SlowScrollToBottom(page)
		page.MustWaitStable()

		elems, err := page.Elements(endpointToScrape.MainElementSelector)
		if err != nil {
			return nil, fmt.Errorf("error getting main elements: %w", err)
		}

		for _, elem := range elems {
			linkElem, err := elem.Element(endpointToScrape.DetailedViewTriggerSelector)
			if err != nil {
				fmt.Printf("error getting link element: %v", err)
				continue
			}

			attr, err := linkElem.Attribute("href")
			if err != nil {
				fmt.Printf("error getting href attribute: %v", err)
				continue
			}

			fullUrl := helpers.GetFullUrl(endpointToScrape.URL, *attr)
			fmt.Println("Full URL: ", fullUrl)

			detailPage, err := GetStealthPage(browser, fullUrl, endpointToScrape.DetailedViewMainElementSelector)
			if err != nil {
				fmt.Printf("error getting detailed view page: %v", err)
				continue
			}

			detailPage.MustWaitStable()

			detailElem := detailPage.MustElement(endpointToScrape.DetailedViewMainElementSelector)
			if detailElem == nil {
				fmt.Println("Detailed view element is null")
				detailPage.Close()
				continue
			}

			pageData := []PageData{{Page: nil, Element: detailElem, ActualLink: fullUrl}}
			pageResults, err := processElements(pageData, endpointToScrape, relevantGroup)
			if err != nil {
				fmt.Printf("error processing detail page: %v", err)
				detailPage.Close()
				continue
			}

			results = append(results, pageResults...)
			detailPage.Close()
		}
	}

	return results, nil
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
	if strings.Contains(baseURL, "?") {
		if strings.Contains(baseURL, parameter) {
			re := regexp.MustCompile(parameter + `=\d+`)
			return re.ReplaceAllString(baseURL, fmt.Sprintf("%s=%d", parameter, page))
		}
		return fmt.Sprintf("%s&%s=%d", baseURL, parameter, page)
	}
	return fmt.Sprintf("%s?%s=%d", baseURL, parameter, page)
}

func buildURLPathPagination(baseURL string, parameter string, page int, urlRegexToInsert string) string {
	if urlRegexToInsert == "" {
		urlRegexToInsert = `\d+`
	}
	re, err := regexp.Compile(urlRegexToInsert)
	if err != nil {
		fmt.Println("Error extracting regex for pagination: ", err)
		return baseURL
	}
	replacement := fmt.Sprintf("%s%d", parameter, page)

	if re.MatchString(baseURL) {
		return re.ReplaceAllString(baseURL, replacement)
	}

	return baseURL
}

func findLinkFieldId(fields []models.Field) string {
	for _, field := range fields {
		if field.Type == models.FieldTypeLink && field.Name == "Link" {
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
			} else {
				text = 0.0
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

func processElements(elements []PageData, endpointToScrape models.Endpoint, relevantGroup models.ScrapeGroup) ([]models.ScrapeResult, error) {
	results := make([]models.ScrapeResult, 0, len(elements))
	linkFieldId := findLinkFieldId(relevantGroup.Fields)

	for _, element := range elements {
		details, err := getElementDetails(element.Element, endpointToScrape.DetailFieldSelectors, relevantGroup.Fields)
		if err != nil {
			return nil, fmt.Errorf("error getting element details: %w", err)
		}

		if element.ActualLink != "" {
			for i, detail := range details {
				if detail.FieldID == linkFieldId {
					details[i].Value = element.ActualLink
				}
			}
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
	return results, nil
}

func ScrapeEndpointTest(endpointToScrape models.Endpoint, relevantGroup models.ScrapeGroup, client *mongo.Client, browser *rod.Browser) ([]models.ScrapeResultTest, []models.ScrapeResultTest, error) {
	fmt.Println("Scraping endpoint test")

	var results []models.ScrapeResultTest
	scrapeType := GetScrapeType(endpointToScrape)

	switch scrapeType {
	case PureDetails:
		scraped, err := scrapeTestPureDetails(endpointToScrape, relevantGroup, browser)
		if err != nil {
			return nil, nil, fmt.Errorf("error scraping pure details: %w", err)
		}
		results = scraped

	case Previews:
		scraped, err := scrapeTestPreviewsPages(endpointToScrape, relevantGroup, browser)
		if err != nil {
			return nil, nil, fmt.Errorf("error scraping previews pages: %w", err)
		}
		results = scraped

	case PreviewsWithDetails:
		scraped, err := scrapeTestPreviewsWithDetails(endpointToScrape, relevantGroup, browser)
		if err != nil {
			return nil, nil, fmt.Errorf("error scraping previews with details: %w", err)
		}
		results = scraped

	default:
		return nil, nil, fmt.Errorf("unknown scrape type: %v", scrapeType)
	}

	return results, nil, nil
}

func scrapeTestPureDetails(endpointToScrape models.Endpoint, relevantGroup models.ScrapeGroup, browser *rod.Browser) ([]models.ScrapeResultTest, error) {
	page, err := GetStealthPage(browser, endpointToScrape.URL, endpointToScrape.DetailedViewMainElementSelector)
	if err != nil {
		return nil, fmt.Errorf("error getting page: %w", err)
	}
	defer page.Close()

	SlowScrollToBottom(page)
	page.MustWaitStable()

	elements, err := getMainElements(page, endpointToScrape, PureDetails, 1)
	if err != nil {
		return nil, fmt.Errorf("error finding elements: %w", err)
	}

	return processTestElements(elements, endpointToScrape, relevantGroup)
}

func scrapeTestPreviewsPages(endpointToScrape models.Endpoint, relevantGroup models.ScrapeGroup, browser *rod.Browser) ([]models.ScrapeResultTest, error) {
	var allElements []PageData

	page, err := GetStealthPage(browser, endpointToScrape.URL, endpointToScrape.MainElementSelector)
	if err != nil {
		return nil, fmt.Errorf("error getting page: %w", err)
	}
	defer page.Close()

	SlowScrollToBottom(page)
	page.MustWaitStable()

	elements, err := getMainElements(page, endpointToScrape, Previews, 5)
	if err != nil {
		return nil, fmt.Errorf("error finding elements: %w", err)
	}

	allElements = append(allElements, elements...)

	return processTestElements(allElements, endpointToScrape, relevantGroup)
}

func scrapeTestPreviewsWithDetails(endpointToScrape models.Endpoint, relevantGroup models.ScrapeGroup, browser *rod.Browser) ([]models.ScrapeResultTest, error) {
	var results []models.ScrapeResultTest

	page, err := GetStealthPage(browser, endpointToScrape.URL, endpointToScrape.MainElementSelector)
	if err != nil {
		return nil, fmt.Errorf("error getting page: %w", err)
	}
	defer page.Close()

	SlowScrollToBottom(page)
	page.MustWaitStable()

	elems, err := page.Elements(endpointToScrape.MainElementSelector)
	if err != nil {
		return nil, fmt.Errorf("error getting main elements: %w", err)
	}

	for i, elem := range elems {
		if i >= 5 {
			break // Limit to 5 elements for testing
		}

		linkElem, err := elem.Element(endpointToScrape.DetailedViewTriggerSelector)
		if err != nil {
			fmt.Printf("error getting link element: %v", err)
			continue
		}

		attr, err := linkElem.Attribute("href")
		if err != nil {
			fmt.Printf("error getting href attribute: %v", err)
			continue
		}

		fullUrl := helpers.GetFullUrl(endpointToScrape.URL, *attr)
		fmt.Println("Full URL: ", fullUrl)

		detailPage, err := GetStealthPage(browser, fullUrl, endpointToScrape.DetailedViewMainElementSelector)
		if err != nil {
			fmt.Printf("error getting detailed view page: %v", err)
			continue
		}

		detailPage.MustWaitStable()

		detailElem := detailPage.MustElement(endpointToScrape.DetailedViewMainElementSelector)
		if detailElem == nil {
			fmt.Println("Detailed view element is null")
			detailPage.Close()
			continue
		}

		pageData := []PageData{{Page: nil, Element: detailElem, ActualLink: fullUrl}}
		pageResults, err := processTestElements(pageData, endpointToScrape, relevantGroup)
		if err != nil {
			fmt.Printf("error processing detail page: %v", err)
			detailPage.Close()
			continue
		}

		results = append(results, pageResults...)
		detailPage.Close()
	}

	return results, nil
}

func processTestElements(elements []PageData, endpointToScrape models.Endpoint, relevantGroup models.ScrapeGroup) ([]models.ScrapeResultTest, error) {
	results := make([]models.ScrapeResultTest, 0, len(elements))
	linkFieldId := findLinkFieldId(relevantGroup.Fields)

	for _, element := range elements {
		details, err := getElementDetailsTest(element.Element, endpointToScrape.DetailFieldSelectors, relevantGroup.Fields)
		if err != nil {
			return nil, fmt.Errorf("error getting element details: %w", err)
		}

		if element.ActualLink != "" {
			for i, detail := range details {
				if detail.FieldID == linkFieldId {
					details[i].Value = element.ActualLink
				}
			}
		}

		uniqueId := getFieldValueByFieldKeyTest(relevantGroup.Fields, "unique_identifier", details).(string)
		if uniqueId == "" {
			fmt.Println("Unique ID is empty, skipping")
			continue
		}

		result := models.ScrapeResultTest{
			ID:         primitive.NewObjectID(),
			UniqueHash: helpers.GenerateScrapeResultHash(endpointToScrape.ID + uniqueId),
			EndpointID: endpointToScrape.ID,
			GroupId:    relevantGroup.ID,
			Fields:     details,
			Timestamp:  time.Now().Format(time.RFC3339),
		}
		results = append(results, result)
	}

	return results, nil
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
			} else {
				text = 0.0
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
