package scraper

import (
	"fmt"
	"regexp"
	"scrapeit/internal/models"
	"scrapeit/internal/utils"
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

func ScrapeEndpoint(endpointToScrape models.Endpoint, relevantGroup models.ScrapeGroup, filterExisiting bool, client *mongo.Client) ([]models.ScrapeResult, error) {

	browser, err := GetBrowser()
	if err != nil {
		return nil, fmt.Errorf("error getting browser: %w", err)
	}
	defer browser.Close()

	var allElements rod.Elements

	for i := endpointToScrape.PaginationConfig.Start; i <= endpointToScrape.PaginationConfig.End; i++ {
		urlWithPagination := buildPaginationURL(endpointToScrape.URL, endpointToScrape.PaginationConfig.Parameter, i)
		fmt.Println("Scraping URL: ", urlWithPagination)

		page, err := GetStealthPage(browser, urlWithPagination)
		if err != nil {
			return nil, fmt.Errorf("error getting page: %w", err)
		}

		defer page.Close()

		SlowScrollToBottom(page)

		elements, err := page.Elements(endpointToScrape.MainElementSelector)
		if err != nil {
			return nil, fmt.Errorf("error finding elements: %w", err)
		}

		allElements = append(allElements, elements...)
	}

	linkFieldId := findLinkFieldId(relevantGroup.Fields)
	endpointLinkSelector := findLinkSelector(endpointToScrape.DetailFieldSelectors, linkFieldId)
	fmt.Println("Link selector: ", endpointLinkSelector)
	var filteredElements rod.Elements
	if filterExisiting {
		filteredElements = filterElements(allElements, endpointLinkSelector, endpointToScrape.ID, relevantGroup.ID, client)
	} else {
		filteredElements = allElements
	}

	results := make([]models.ScrapeResult, 0, len(filteredElements))
	for _, element := range filteredElements {
		details, err := getElementDetails(element, endpointToScrape.DetailFieldSelectors)
		if err != nil {
			return nil, fmt.Errorf("error getting element details: %w", err)
		}
		fmt.Println("--------------------")
		fmt.Println(fmt.Sprintf("URL: %s, hash: %s", getFieldValueByFieldName(relevantGroup.Fields, "link", details), utils.GenerateScrapeResultHash(getFieldValueByFieldName(relevantGroup.Fields, "link", details))))
		fmt.Println("--------------------")
		result := models.ScrapeResult{
			ID:         primitive.NewObjectID(),
			UniqueHash: utils.GenerateScrapeResultHash(getFieldValueByFieldName(relevantGroup.Fields, "link", details)),
			EndpointID: endpointToScrape.ID,
			GroupId:    relevantGroup.ID,
			Fields:     details,
			Timestamp:  time.Now().Format(time.RFC3339),
		}
		results = append(results, result)
	}

	return results, nil
}

func buildPaginationURL(baseURL, parameter string, page int) string {
	if strings.Contains(baseURL, "?") {
		if strings.Contains(baseURL, parameter) {
			// replace the existing page number with regex
			re := regexp.MustCompile(parameter + `=\d+`)
			return re.ReplaceAllString(baseURL, fmt.Sprintf("%s=%d", parameter, page))
		}
		return fmt.Sprintf("%s&%s=%d", baseURL, parameter, page)
	}
	return fmt.Sprintf("%s?%s=%d", baseURL, parameter, page)
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

func filterElements(elements rod.Elements, linkSelector models.FieldSelector, endpointId string, groupId primitive.ObjectID, client *mongo.Client) rod.Elements {
	var filtered rod.Elements
	for _, element := range elements {
		linkElement, err := element.Element(linkSelector.Selector)
		if err != nil {
			continue
		}
		hrefAttr, err := linkElement.Attribute("href")
		if err != nil || hrefAttr == nil {
			continue
		}
		if !utils.FindScrapeResultExists(*hrefAttr, endpointId, groupId, client) {
			filtered = append(filtered, element)
		}
	}
	return filtered
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
			text, err = utils.ExtractStringWithRegex(text, selector.Regex)
			if err != nil {
				fmt.Printf("Error extracting regex for selector %s: %v\n", selector.Selector, err)
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
