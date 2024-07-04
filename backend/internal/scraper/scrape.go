package scraper

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/devices"
	"github.com/go-rod/stealth"
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

	browser := rod.New().ControlURL(rodBrowserWsURL).MustConnect().DefaultDevice(devices.Device{
		Screen: devices.Screen{Horizontal: devices.ScreenSize{
			Width:  1920,
			Height: 1080,
		}},
	})
	return browser, nil
}

func Scrape() (map[string]string, error) {
	browser, err := getBrowser()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer browser.Close()

	page := browser.MustPage("https://brightdata.com/solutions/rotating-proxies").MustWaitLoad().MustWindowFullscreen()
	// get current page dimensions
	res := page.MustEval(`() => {
		return "Page dimensions:," + window.innerWidth + " " + window.innerHeight;
	}`)
	fmt.Println(res.String())
	page.MustScreenshot("screenshot.png")
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)
		}
	}()

	title := page.MustElement("title").MustText()
	return map[string]string{"titleddsaffadssdaf": title}, nil
}

func GetMainElementHTMLContent(url, elementSelector string) (string, error) {
	browser, err := getBrowser()
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer browser.Close()

	page := stealth.MustPage(browser).MustNavigate(url).MustWaitLoad().MustWindowFullscreen()

	// scroll to the bottom of the page
	page.MustEval(`() => {
		window.scrollTo(0, document.body.scrollHeight);
	}`)

	fmt.Printf("Element selector: %v\n", elementSelector)

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)
		}
	}()

	html := page.MustElement(elementSelector).MustHTML()
	fmt.Printf("HTML: %v\n", len(strings.TrimSpace(html)))
	return html, nil
}
