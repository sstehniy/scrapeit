package scraper

import (
	"fmt"
	"os"

	"github.com/go-rod/rod"
)

var browser *rod.Browser

func Scrape() (map[string]string, error) {
	rodBrowserWsURL := os.Getenv("ROD_BROWSER_WS_URL")

	if rodBrowserWsURL == "" {
		fmt.Println("ROD_BROWSER_WS_URL environment variable not set")
		return nil, fmt.Errorf("ROD_BROWSER_WS_URL environment variable not set")
	}

	if browser == nil {
		// Connect to the browser
		browser = rod.New().ControlURL(rodBrowserWsURL).MustConnect()
	}

	// Navigate to a page
	page := browser.MustPage("https://example.com")

	// Interact with the page
	title := page.MustElement("title").MustText()
	return map[string]string{"title": title}, nil
}
