package helpers

import (
	"fmt"
	"regexp"
)

func ExtractStringWithRegex(input, pattern string) (string, error) {
	// Compile the regex pattern
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", fmt.Errorf("invalid regex pattern: %w", err)
	}

	// Find the first match
	match := re.FindStringSubmatch(input)

	// If a match is found, return the first captured group (or the entire match if no group)
	if len(match) > 1 {
		return match[1], nil
	} else if len(match) == 1 {
		return match[0], nil
	}

	// Return an empty string and an error if no match is found
	return "", fmt.Errorf("no match found")
}
