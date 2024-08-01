package scraper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"scrapeit/internal/helpers"
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

var (
	browser     *rod.Browser
	browserOnce sync.Once
)

func GetBrowser() *rod.Browser {
	browserOnce.Do(func() {
		rodBrowserWsURL := os.Getenv("ROD_BROWSER_WS_URL")
		if rodBrowserWsURL == "" {
			panic("ROD_BROWSER_WS_URL is not set")
		}
		log.Printf("Connecting to browser at %s", rodBrowserWsURL)
		err := rod.New().ControlURL(rodBrowserWsURL).Connect()
		if err != nil {
			fmt.Println("Error connecting to browser: ", err)
			panic(err)
		}
		browser = rod.New().ControlURL(rodBrowserWsURL).MustConnect().DefaultDevice(devices.LaptopWithHiDPIScreen)
		browser.MustIgnoreCertErrors(true)
	})

	return browser
}

func GetStealthPage(ctx context.Context, browser *rod.Browser, url string, elementToWaitFor string) (*rod.Page, error) {
	// Load the cookie store
	store, err := LoadCookieStore()
	if err != nil {
		log.Fatalf("Failed to load cookie store: %v", err)
		return nil, err
	}

	// Check if we have valid cookies
	baseURL := helpers.GetBaseURL(url)
	cookies, valid := GetValidCookies(store, baseURL)
	// Parse the JSON response
	var response struct {
		Status   string `json:"status"`
		Message  string `json:"message"`
		Solution struct {
			URL       string            `json:"url"`
			Status    int               `json:"status"`
			Headers   map[string]string `json:"headers"`
			Response  string            `json:"response"`
			Cookies   []Cookie          `json:"cookies"`
			UserAgent string            `json:"userAgent"`
		} `json:"solution"`
	}

	if !valid {
		// Make the request to get cookies if not valid
		resp, err := http.Post("http://flaresolverr:8191/v1", "application/json", bytes.NewBuffer([]byte(fmt.Sprintf(`{
			"cmd": "request.get",
			"url": "%s",
			"maxTimeout": 30000,
			"returnOnlyCookies": false
		}`, url))))
		if err != nil {
			log.Fatalf("Failed to make request: %v", err)
			return nil, err
		}
		defer resp.Body.Close()

		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("Failed to read response body: %v", err)
			return nil, err
		}

		if err := json.Unmarshal(body, &response); err != nil {
			log.Fatalf("Failed to unmarshal JSON response: %v", err)
			return nil, err
		}

		cookies = UserAgentWithCookies{
			Cookie:      response.Solution.Cookies,
			UserAgent:   response.Solution.UserAgent,
			LastUpdated: time.Now(),
		}
		// Save the new cookies
		SetCookies(store, baseURL, cookies)
	}

	page := stealth.MustPage(browser)
	page.MustSetViewport(1920, 1080, 2.0, false)

	cookiesToSet := make([]*proto.NetworkCookieParam, len(cookies.Cookie))
	for idx, c := range cookies.Cookie {
		cookiesToSet[idx] = &proto.NetworkCookieParam{
			Name:     c.Name,
			Value:    c.Value,
			Domain:   c.Domain,
			Path:     c.Path,
			Expires:  proto.TimeSinceEpoch(float64(c.Expiry)),
			Secure:   c.Secure,
			HTTPOnly: c.HttpOnly,
			SameSite: proto.NetworkCookieSameSite(c.SameSite),
		}
	}
	page.MustSetCookies(cookiesToSet...)

	// Set the User-Agent if provided
	if cookies.UserAgent != "" {
		fmt.Printf("Setting User-Agent: %s\n", cookies.UserAgent)
		page.MustSetUserAgent(&proto.NetworkSetUserAgentOverride{
			UserAgent: cookies.UserAgent,
		})
	}

	navCtx, navCancel := context.WithTimeout(ctx, 10*time.Second)
	defer navCancel()

	if err := page.Context(navCtx).Navigate(url); err != nil {
		if err == context.DeadlineExceeded {
			log.Printf("Timeout reached while navigating to %s: %v", url, err)
		} else {
			log.Printf("Error navigating to %s: %v", url, err)
		}
		return nil, err
	}

	loadCtx, loadCancel := context.WithTimeout(ctx, 10*time.Second)
	defer loadCancel()

	if err := page.Context(loadCtx).WaitLoad(); err != nil {
		if err == context.DeadlineExceeded {
			log.Printf("Timeout reached while waiting for page to load: %v", err)
		} else {
			log.Printf("Error waiting for page to load: %v", err)
		}
		return nil, err
	}

	elementCtx, elementCancel := context.WithTimeout(ctx, 10*time.Second)
	defer elementCancel()

	_, err = page.Context(elementCtx).Element(elementToWaitFor)
	if err != nil {
		if err == context.DeadlineExceeded {
			log.Printf("Timeout reached while waiting for element %s: %v", elementToWaitFor, err)
			page.MustScreenshot("timeout_screenshot.png")
		} else {
			log.Printf("Error finding element %s: %v", elementToWaitFor, err)
			page.MustScreenshot("error_screenshot.png")
		}
		return nil, err
	}

	return page, nil
}
func ScrapeTest() (map[string]string, error) {
	browser := GetBrowser()

	page, err := GetStealthPage(context.Background(), browser, "https://app.zenrows.com/register", ".min-h-screen.flex.bg-secondary")
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

		fmt.Println("Current scroll: ", currentScroll)
		// Scroll to the new position
		_, err = page.Eval(fmt.Sprintf("() => window.scrollTo(0, %d)", currentScroll))
		if err != nil {
			fmt.Println("Error scrolling: ", err)
			return fmt.Errorf("failed to scroll: %w", err)
		}

		time.Sleep(scrollDelay * time.Millisecond)
		page.MustWaitLoad().MustWaitIdle()

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

	page.MustEval("() => window.scrollTo(0, 0)")

	time.Sleep(1 * time.Second)

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

	// fmt.Println("Total height: ", totalHeight)
	// Scroll loop
	for currentScroll := 0; currentScroll < totalHeight; currentScroll += viewportHeight {

		// fmt.Println("Current scroll: ", currentScroll)
		// Scroll to the new position
		_, err = page.Eval(fmt.Sprintf("() => window.scrollTo(0, %d)", currentScroll))
		if err != nil {
			fmt.Println("Error scrolling: ", err)
			return fmt.Errorf("failed to scroll: %w", err)
		}

		time.Sleep(scrollDelay * time.Millisecond)
		page.MustWaitLoad().MustWaitIdle()

		// Print progress
		// fmt.Printf("\rScrolling: %d/%d pixels", currentScroll, totalHeight)

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

	page.MustEval("() => window.scrollTo(0, 0)")

	time.Sleep(1 * time.Second)

	fmt.Println() // Print a newline after the progress indicator
	return nil
}
