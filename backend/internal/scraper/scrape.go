package scraper

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/devices"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/stealth"
)

var (
	mu sync.Mutex
)

func GetBrowser() *rod.Browser {
	rodBrowserWsURL := os.Getenv("ROD_BROWSER_WS_URL")
	if rodBrowserWsURL == "" {
		panic("ROD_BROWSER_WS_URL is not set")
	}

	mu.Lock()
	defer mu.Unlock()

	browser := rod.New().ControlURL(rodBrowserWsURL).MustConnect().DefaultDevice(devices.LaptopWithHiDPIScreen)

	return browser
}

func GetStealthPage(browser *rod.Browser, url string, elementToWaitFor string) (*rod.Page, error) {
	page := stealth.MustPage(browser)

	// Navigate to the URL
	if err := page.Navigate(url); err != nil {
		return page, fmt.Errorf("error navigating: %w", err)
	}

	// Wait for load and element to be visible (max 5 seconds)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	err := page.Context(ctx).MustWaitLoad().MustWaitIdle().MustWaitStable().WaitElementsMoreThan(elementToWaitFor, 0)

	fmt.Println("Waited for load and element")

	if err != nil {
		if err == context.DeadlineExceeded {
			fmt.Println("Load and element wait timeout exceeded")
		} else {
			fmt.Println("Error waiting for load and element:", err)
		}

		// If timeout, try to wait for navigation (max 5 seconds)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		page.Context(ctx).WaitNavigation(proto.PageLifecycleEventNameNetworkAlmostIdle)
		fmt.Println("Waited for navigation")
	}

	// Set viewport and take screenshot
	page.MustSetViewport(1920, 1080, 2.0, false)

	return page, nil
}

func ScrapeTest() (map[string]string, error) {
	browser := GetBrowser()

	defer browser.Close()

	page, err := GetStealthPage(browser, "https://app.zenrows.com/register", ".min-h-screen.flex.bg-secondary")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
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
	scrollDelay = 200 // milliseconds
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

		time.Sleep(scrollDelay * time.Millisecond)
		page.MustWaitRequestIdle()

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

		time.Sleep(scrollDelay * time.Millisecond)
		page.MustWaitRequestIdle()

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
