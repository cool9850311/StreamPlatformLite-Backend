package errors

import "errors"

var (
	ErrUnauthorized = errors.New("unauthorized access")
	ErrNotFound     = errors.New("resource not found")
	// Add other error types as needed
)
