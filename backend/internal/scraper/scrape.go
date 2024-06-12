package scraper

import (
	"fmt"
	"os"
	"sync"

	"github.com/go-rod/rod"
)

var (
	mu sync.Mutex
)

func getBrowser() (*rod.Browser, error) {
	rodBrowserWsURL := os.Getenv("ROD_BROWSER_WS_URL")
	if rodBrowserWsURL == "" {
		return nil, fmt.Errorf("ROD_BROWSER_WS_URL environment variable not set")
	}

	mu.Lock()
	defer mu.Unlock()

	browser := rod.New().ControlURL(rodBrowserWsURL).MustConnect()

	return browser, nil
}

func Scrape() (map[string]string, error) {
	browser, err := getBrowser()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer browser.Close()

	page := browser.MustPage("https://example.com")

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)
		}
	}()

	title := page.MustElement("title").MustText()
	return map[string]string{"title": title}, nil
}
