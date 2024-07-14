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

	// Regular expression to match data attributes
	dataAttrRegex := regexp.MustCompile(`\s*data-[a-zA-Z0-9\-_]+\s*=\s*"[^"]*"`)

	// remove all svgs
	svgRegex := regexp.MustCompile(`<svg[^>]*>.*?</svg>`)

	// Remove all inline styles and data attributes
	withoutStyles := styleRegex.ReplaceAllString(noNewlines, "")

	withoutDataAttrs := dataAttrRegex.ReplaceAllString(withoutStyles, "")

	withoutSVGs := svgRegex.ReplaceAllString(withoutDataAttrs, "")

	return withoutSVGs
}

func GetMainElementHTMLContent(url, elementSelector string, maxElements int) (string, error) {
	browser, err := GetBrowser()
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer browser.Close()

	page, err := GetStealthPage(browser, url)
	if err != nil {

		return "", err
	}

	page.MustWaitElementsMoreThan(
		elementSelector, 0,
	).MustSetViewport(1920, 1080,
		2.0,
		false,
	)
	// scroll to the bottom of the page
	SlowScrollToHalf(page)
	page.MustWaitLoad()

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
