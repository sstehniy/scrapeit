package scraper

import (
	"context"
	"fmt"
	"scrapeit/internal/helpers"
	"scrapeit/internal/models"

	"github.com/go-rod/rod"
)

// PageData stores the page and its main element
type PageData struct {
	Page       *rod.Page
	Element    *rod.Element
	ActualLink string
}

func getMainElements(page *rod.Page, endpoint models.Endpoint, scrapeType ScrapeType, limit int) ([]PageData, error) {
	switch scrapeType {
	case Previews:
		elements := page.MustElements(endpoint.MainElementSelector)
		pageData := make([]PageData, len(elements))
		for i, elem := range elements {
			pageData[i] = PageData{Page: nil, Element: elem}
		}
		return pageData, nil

	case PreviewsWithDetails:
		elems, err := page.Elements(endpoint.MainElementSelector)
		if err != nil {
			return nil, fmt.Errorf("error getting main elements: %w", err)
		}
		fmt.Printf("Found %v elements\n", len(elems))

		var detailPages []PageData

		for _, elem := range elems {
			linkElem, err := elem.Element(endpoint.DetailedViewTriggerSelector)
			if err != nil {
				fmt.Printf("error getting link element: %v", err)
				continue
			}

			attr, err := linkElem.Attribute("href")
			if err != nil {
				fmt.Printf("error getting href attribute: %v", err)
				continue
			}

			fullUrl := helpers.GetFullUrl(endpoint.URL, *attr)
			fmt.Println("Full URL: ", fullUrl)

			newPage, err := GetStealthPage(context.Background(), page.Browser(), fullUrl, endpoint.DetailedViewMainElementSelector)
			if err != nil {
				fmt.Printf("error getting detailed view page: %v", err)
				newPage.MustClose()
				continue
			}

			// SlowScrollToBottom(newPage)
			newPage.MustWaitStable()

			fmt.Println("Navigated to detailed view")
			// newPage.MustScreenshot("screenshot.png")

			detailElem := newPage.MustElement(endpoint.DetailedViewMainElementSelector)
			detailPages = append(detailPages, PageData{Page: newPage, Element: detailElem, ActualLink: fullUrl})
			// check if not null
			if detailElem != nil {
				fmt.Println("Found detailed view element")
			} else {
				fmt.Println("Detailed view element is null")
			}

			if limit != -1 && len(detailPages) >= limit {
				break
			}
		}
		return detailPages, nil

	case PureDetails:
		element := page.MustElement(endpoint.DetailedViewMainElementSelector)
		return []PageData{{Page: nil, Element: element, ActualLink: page.MustInfo().URL}}, nil

	default:
		return nil, fmt.Errorf("unknown scrape type: %v", scrapeType)
	}
}
