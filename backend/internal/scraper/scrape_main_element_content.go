package scraper

import (
	"fmt"
	"regexp"
	"scrapeit/internal/models"
	"strings"
)

func cleanHTML(input string) string {
	// Remove all newline characters
	noNewlines := strings.ReplaceAll(input, "\n", "")

	// Regular expression to match inline styles
	styleRegex := regexp.MustCompile(`\s*style\s*=\s*"[^"]*"`)

	// remove all svgs
	svgRegex := regexp.MustCompile(`<svg[^>]*>.*?</svg>`)

	// Remove all inline styles and data attributes
	withoutStyles := styleRegex.ReplaceAllString(noNewlines, "")

	withoutSVGs := svgRegex.ReplaceAllString(withoutStyles, "")

	return withoutSVGs
}

func GetMainElementHTMLContent(endpoint models.Endpoint, maxElements int) (string, error) {
	browser := GetBrowser()
	defer browser.Close()
	scrapeType := GetScrapeType(endpoint)
	elementToWaitFor := endpoint.MainElementSelector
	if scrapeType == PureDetails {
		elementToWaitFor = endpoint.DetailedViewMainElementSelector
	}

	fmt.Printf("Element selector: %v\n", elementToWaitFor)

	fmt.Println("Scrape type: ", GetScrapeType(endpoint))
	page, err := GetStealthPage(browser, endpoint.URL, elementToWaitFor)
	if err != nil {
		return "", err
	}

	defer page.Close()

	// scroll to the bottom of the page
	SlowScrollToHalf(page)
	page.MustWaitLoad().MustWaitStable()

	elems, err := getMainElements(page, endpoint, GetScrapeType(endpoint), maxElements)
	fmt.Printf("Found %v elements\n", len(elems))
	if err != nil {
		return "", err
	}
	defer func() {
		for _, elem := range elems {
			if elem.Page != nil {
				elem.Page.Close()
			}
		}
	}()
	html := ""
	for idx, elem := range elems {
		html += elem.Element.MustHTML()
		if idx >= maxElements {
			break
		}
	}
	fmt.Printf("HTML: %v\n", len(strings.TrimSpace(html)))
	cleanedHTML := cleanHTML(html)
	fmt.Printf("Cleaned HTML: %v\n", len(strings.TrimSpace(cleanedHTML)))
	return cleanedHTML, nil
}
