package util

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func GenerateCsrfToken(secret, userID string) (string, error) {
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	nonceHex := hex.EncodeToString(nonce)
	data := nonceHex + ":" + userID
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(data))
	sigHex := hex.EncodeToString(mac.Sum(nil))
	return nonceHex + "." + sigHex, nil
}

func ValidateCsrfToken(token, secret, userID string) bool {
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return false
	}
	data := parts[0] + ":" + userID
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(data))
	expectedHex := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(parts[1]), []byte(expectedHex))
}
