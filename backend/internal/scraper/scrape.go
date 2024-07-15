package scraper

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/devices"
	"github.com/go-rod/stealth"
)

var (
	mu sync.Mutex
)

func GetBrowser() (*rod.Browser, error) {
	rodBrowserWsURL := os.Getenv("ROD_BROWSER_WS_URL")
	if rodBrowserWsURL == "" {
		return nil, fmt.Errorf("ROD_BROWSER_WS_URL environment variable not set")
	}

	mu.Lock()
	defer mu.Unlock()

	browser := rod.New().ControlURL(rodBrowserWsURL).MustConnect().DefaultDevice(devices.LaptopWithHiDPIScreen)

	return browser, nil
}

func GetStealthPage(browser *rod.Browser, url string) (*rod.Page, error) {

	err := stealth.MustPage(browser).Navigate(url)
	if err != nil {
		fmt.Println("Error navigating: ", err)
		return nil, err

	}
	page := stealth.MustPage(browser).MustNavigate(url)
	page.MustWaitLoad()

	page.MustSetViewport(1920, 1080,
		2.0,
		false,
	)

	return page, nil
}

func ScrapeTest() (map[string]string, error) {
	browser, err := GetBrowser()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer browser.Close()

	page := browser.MustPage("https://brightdata.com/solutions/rotating-proxies").MustWaitLoad().MustWindowFullscreen()
	// get current page dimensions
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)
		}
	}()

	title := page.MustElement("title").MustText()
	return map[string]string{"titleddsaffadssdaf": title}, nil
}

const (
	scrollDelay = 150 // milliseconds
)

func SlowScrollToHalf(page *rod.Page) error {
	fmt.Println("Scrolling to half")
	var totalHeight int
	var viewportHeight int

	// Get the initial document height and viewport height
	result := page.MustEval(`() => [document.documentElement.scrollHeight, window.innerHeight]`).String()

	// parse the result [totalHeight viewportHeight]
	_, err := fmt.Sscanf(result, "[%d %d]", &totalHeight, &viewportHeight)

	if err != nil {
		return fmt.Errorf("failed to parse result: %w", err)
	}

	fmt.Println("Total height: ", totalHeight)
	// Scroll loop
	for currentScroll := 0; currentScroll < totalHeight/2; currentScroll += viewportHeight {
		// screenshot into separate folder
		// create folder if not exists

		fmt.Println("Current scroll: ", currentScroll)
		// Scroll to the new position
		_, err = page.Eval(fmt.Sprintf("() => window.scrollTo(0, %d)", currentScroll))
		if err != nil {
			fmt.Println("Error scrolling: ", err)
			return fmt.Errorf("failed to scroll: %w", err)
		}

		// Wait for a short time to allow content to load
		time.Sleep(scrollDelay * time.Millisecond)

		// Print progress
		fmt.Printf("\rScrolling: %d/%d pixels", currentScroll, totalHeight)

		// Check if the document height has changed (in case of dynamically loaded content)
		var newHeight int
		err = page.MustEval(`() => document.documentElement.scrollHeight`).Unmarshal(&newHeight)
		if err != nil {
			fmt.Println("Error getting new height: ", err)
			return fmt.Errorf("failed to get new height: %w", err)
		}

		if newHeight > totalHeight {
			totalHeight = newHeight
		}

		// Check if we've reached the bottom
		if currentScroll+viewportHeight >= totalHeight {
			break
		}
	}

	fmt.Println() // Print a newline after the progress indicator
	return nil

}

func SlowScrollToBottom(page *rod.Page) error {
	fmt.Println("Scrolling to bottom")
	var totalHeight int
	var viewportHeight int

	// Get the initial document height and viewport height
	result := page.MustEval(`() => [document.documentElement.scrollHeight, window.innerHeight]`).String()

	// parse the result [totalHeight viewportHeight]
	_, err := fmt.Sscanf(result, "[%d %d]", &totalHeight, &viewportHeight)

	if err != nil {
		return fmt.Errorf("failed to parse result: %w", err)
	}

	fmt.Println("Total height: ", totalHeight)
	// Scroll loop
	for currentScroll := 0; currentScroll < totalHeight; currentScroll += viewportHeight {
		// screenshot into separate folder
		// create folder if not exists

		fmt.Println("Current scroll: ", currentScroll)
		// Scroll to the new position
		_, err = page.Eval(fmt.Sprintf("() => window.scrollTo(0, %d)", currentScroll))
		if err != nil {
			fmt.Println("Error scrolling: ", err)
			return fmt.Errorf("failed to scroll: %w", err)
		}

		// Wait for a short time to allow content to load
		time.Sleep(scrollDelay * time.Millisecond)

		// Print progress
		fmt.Printf("\rScrolling: %d/%d pixels", currentScroll, totalHeight)

		// Check if the document height has changed (in case of dynamically loaded content)
		var newHeight int
		err = page.MustEval(`() => document.documentElement.scrollHeight`).Unmarshal(&newHeight)
		if err != nil {
			fmt.Println("Error getting new height: ", err)
			return fmt.Errorf("failed to get new height: %w", err)
		}

		if newHeight > totalHeight {
			totalHeight = newHeight
		}

		// Check if we've reached the bottom
		if currentScroll+viewportHeight >= totalHeight {
			break
		}
	}

	fmt.Println() // Print a newline after the progress indicator
	return nil
}
