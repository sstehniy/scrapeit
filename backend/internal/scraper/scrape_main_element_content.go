package scraper

import (
	"fmt"
	"regexp"
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

func GetMainElementHTMLContent(url, elementSelector string, maxElements int) (string, error) {
	browser := GetBrowser()
	defer browser.Close()

	page, err := GetStealthPage(browser, url, elementSelector)
	if err != nil {
		return "", err
	}

	defer page.Close()

	// scroll to the bottom of the page
	SlowScrollToHalf(page)
	page.MustWaitLoad().MustWaitStable()

	fmt.Printf("Element selector: %v\n", elementSelector)

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)
		}
	}()

	elems := page.MustElements(elementSelector)
	html := ""
	for idx, elem := range elems {
		html += elem.MustHTML()
		if idx >= maxElements {
			break
		}
	}
	fmt.Printf("HTML: %v\n", len(strings.TrimSpace(html)))
	cleanedHTML := cleanHTML(html)
	fmt.Printf("Cleaned HTML: %v\n", len(strings.TrimSpace(cleanedHTML)))
	return cleanedHTML, nil
}
