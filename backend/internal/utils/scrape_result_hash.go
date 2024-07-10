package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

// should always be the same for the same input
func GenerateScrapeResultHash(uniqueString string) string {
	hasher := sha256.New()
	hasher.Write([]byte(uniqueString))
	return hex.EncodeToString(hasher.Sum(nil))
}
