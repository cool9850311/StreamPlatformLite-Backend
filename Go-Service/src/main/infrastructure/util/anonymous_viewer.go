package util

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
)

// GenerateViewerIDFromIP derives a deterministic viewer ID from a client IP.
// Same IP + same secret always produces the same viewer ID.
func GenerateViewerIDFromIP(clientIP, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(clientIP))
	h := mac.Sum(nil)
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		h[0:4], h[4:6], h[6:8], h[8:10], h[10:16])
}
