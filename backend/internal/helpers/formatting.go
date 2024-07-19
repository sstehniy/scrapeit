package helpers

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

func ExtractStringWithRegex(input, pattern string, regexMatchIndexToUse int) (string, error) {
	// Compile the regex pattern
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", fmt.Errorf("invalid regex pattern: %w", err)
	}

	// Find the first match
	match := re.FindStringSubmatch(input)

	// fmt.Printf("Matches: ")
	// for _, m := range match {
	// 	fmt.Printf("%s ", m)
	// }
	// fmt.Println("regexMatchIndexToUse: ", regexMatchIndexToUse)

	// If a match is found, return the first captured group (or the entire match if no group)
	if len(match) > 0 {
		if len(match) < regexMatchIndexToUse {
			return match[0], nil
		}
		return match[regexMatchIndexToUse], nil
	}

	// Return an empty string and an error if no match is found
	return "", fmt.Errorf("no match found")
}

func CastTextNumberStringToFloat(priceStrAny interface{}) float64 {
	if priceStrAny == nil {
		return 0
	}
	// Convert the input to a string
	priceStr := fmt.Sprintf("%v", priceStrAny)

	// Remove currency symbols and spaces
	currencySymbols := regexp.MustCompile(`[^\d.,-]`)
	cleanedStr := currencySymbols.ReplaceAllString(priceStr, "")
	cleanedStr = strings.ReplaceAll(cleanedStr, " ", "")

	// Identify and handle different formatting styles
	if strings.Count(cleanedStr, ".") > 1 {
		// Case: Dots as thousand separators, comma as decimal
		cleanedStr = strings.ReplaceAll(cleanedStr, ".", "")
		cleanedStr = strings.Replace(cleanedStr, ",", ".", 1)
	} else if strings.Count(cleanedStr, ",") > 1 {
		// Case: Commas as thousand separators, dot as decimal
		cleanedStr = strings.ReplaceAll(cleanedStr, ",", "")
	} else if strings.Count(cleanedStr, ",") == 1 && strings.Count(cleanedStr, ".") == 1 {
		// Mixed case: Assume last occurrence is the decimal separator
		if strings.LastIndex(cleanedStr, ".") > strings.LastIndex(cleanedStr, ",") {
			cleanedStr = strings.Replace(cleanedStr, ",", "", -1)
		} else {
			cleanedStr = strings.Replace(cleanedStr, ".", "", -1)
			cleanedStr = strings.Replace(cleanedStr, ",", ".", 1)
		}
	} else if strings.Count(cleanedStr, ",") == 1 {
		// Case: Single comma, assume decimal separator
		cleanedStr = strings.Replace(cleanedStr, ",", ".", 1)
	}

	// Convert the cleaned string to float64
	floatVal, err := strconv.ParseFloat(cleanedStr, 64)
	if err != nil {
		return 0
	}

	fmt.Println("Float value: ", math.Round(floatVal*100)/100)

	return math.Round(floatVal*100) / 100
}
