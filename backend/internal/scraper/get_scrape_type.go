package scraper

import (
	"scrapeit/internal/models"
	"strings"
)

type ScrapeType string

const (
	Previews            ScrapeType = "previews"
	PreviewsWithDetails ScrapeType = "previews_with_details"
	PureDetails         ScrapeType = "pure_details"
)

func GetScrapeType(endpoint models.Endpoint) ScrapeType {
	listElementsSelector := strings.TrimSpace(endpoint.MainElementSelector)
	withDetailedView := endpoint.WithDetailedView
	detailedViewTriggerSelector := strings.TrimSpace(endpoint.DetailedViewTriggerSelector)
	detailedViewMainElementSelector := strings.TrimSpace(endpoint.DetailedViewMainElementSelector)

	// Config 1: Main Element Selector only (Previews)
	if !withDetailedView && listElementsSelector != "" {
		return Previews
	}

	// Config 2: With Detailed View and Trigger
	if withDetailedView && detailedViewTriggerSelector != "" && listElementsSelector != "" && detailedViewMainElementSelector != "" {
		return PreviewsWithDetails
	}

	// Config 3: With Detailed View, no Trigger
	if withDetailedView && detailedViewTriggerSelector == "" && detailedViewMainElementSelector != "" {
		return PureDetails
	}

	// Default case
	return Previews
}
