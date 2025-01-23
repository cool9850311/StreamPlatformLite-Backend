package util

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
)

func TrimPathToBase(path, base string) string {
	index := strings.Index(path, base)
	if index == -1 {
		return ""
	}

	trimmedPath := path[:index+len(base)]
	return trimmedPath
}
func GenerateRandomBase64String(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
