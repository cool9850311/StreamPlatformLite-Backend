package util

import (
	"net/url"
	"strings"
)

func EncodeRFC5987(s string) string {
	// URL encode the string
	encoded := url.QueryEscape(s)
	// Replace spaces with %20 instead of +
	return strings.ReplaceAll(encoded, "+", "%20")
}
