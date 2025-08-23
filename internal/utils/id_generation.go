package utils

import (
	"crypto/sha256"
	"encoding/base64"
)

const shortIDLength = 8

func GenerateID(url string) string {
	hash := sha256.Sum256([]byte(url))
	encoded := base64.URLEncoding.EncodeToString(hash[:])
	if shortIDLength > len(encoded) {
		return encoded
	}
	return encoded[:shortIDLength]
}
