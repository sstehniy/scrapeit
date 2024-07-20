package helpers

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

func ExtractStringWithRegex(input, pattern string, regexMatchIndexToUse int) (string, []string, error) {
	input = strings.ReplaceAll(input, "\u00A0", " ")

	// Compile the regex pattern
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", []string{}, fmt.Errorf("invalid regex pattern: %w", err)
	}

	// Find the first match
	match := re.FindStringSubmatch(input)

	fmt.Printf("Matches: ")
	for _, m := range match {
		fmt.Printf("%s ", m)
	}
	fmt.Println("regexMatchIndexToUse: ", regexMatchIndexToUse)

	// If a match is found, return the first captured group (or the entire match if no group)
	if len(match) > 0 {
		if len(match) < regexMatchIndexToUse {
			return match[0], match, nil
		}
		return match[regexMatchIndexToUse], match, nil
	}

	// Return an empty string and an error if no match is found
	return "", match, fmt.Errorf("no match found")
}

func CastPriceStringToFloat(priceStr string) float64 {
	// Remove currency symbols and spaces
	currencySymbols := regexp.MustCompile(`[^\d.,-]`)
	cleanedStr := currencySymbols.ReplaceAllString(priceStr, "")
	cleanedStr = strings.ReplaceAll(cleanedStr, " ", "")

	// Handle different formatting styles
	// Case: Multiple dots or multiple commas
	if strings.Count(cleanedStr, ".") > 1 {
		// Assume dots as thousand separators and comma as decimal
		cleanedStr = strings.ReplaceAll(cleanedStr, ".", "")
		cleanedStr = strings.Replace(cleanedStr, ",", ".", 1)
	} else if strings.Count(cleanedStr, ",") > 1 {
		// Assume commas as thousand separators and dot as decimal
		cleanedStr = strings.ReplaceAll(cleanedStr, ",", "")
	} else if strings.Count(cleanedStr, ",") == 1 && strings.Count(cleanedStr, ".") == 1 {
		// Mixed case: Assume the last occurrence is the decimal separator
		if strings.LastIndex(cleanedStr, ".") > strings.LastIndex(cleanedStr, ",") {
			cleanedStr = strings.Replace(cleanedStr, ",", "", -1)
		} else {
			cleanedStr = strings.Replace(cleanedStr, ".", "", -1)
			cleanedStr = strings.Replace(cleanedStr, ",", ".", 1)
		}
	} else if strings.Count(cleanedStr, ",") == 1 {
		// Case: Single comma, assume decimal separator
		cleanedStr = strings.Replace(cleanedStr, ",", ".", 1)
	} else if strings.Count(cleanedStr, ".") == 1 {
		// Case: Single dot
		parts := strings.Split(cleanedStr, ".")
		if len(parts[1]) > 2 {
			// If there are more than two decimal places, assume it's a thousand separator
			cleanedStr = strings.ReplaceAll(cleanedStr, ".", "")
		}
	}

	// Convert the cleaned string to float64
	floatVal, err := strconv.ParseFloat(cleanedStr, 64)
	if err != nil {
		return 0
	}
	fmt.Printf("Original value: %v, Float value: %f \n", priceStr, math.Round(floatVal*100)/100)

	return math.Round(floatVal*100) / 100
}
