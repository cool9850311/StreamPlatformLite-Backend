package errors

import "errors"

var (
	ErrUnauthorized     = errors.New("unauthorized access")
	ErrNotFound         = errors.New("resource not found")
	ErrInternal         = errors.New("internal server error")
	ErrInvalidInput     = errors.New("invalid input")
	ErrConnectionClosed = errors.New("connection closed")
	ErrExists           = errors.New("already exists")
	// Add other error types as needed
)
