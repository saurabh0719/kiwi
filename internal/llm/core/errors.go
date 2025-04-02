package core

import (
	"errors"
	"fmt"
)

// Common error types that can be returned by any LLM adapter
var (
	// ErrNullContent is returned when an LLM returns a null content response after a tool call
	ErrNullContent = errors.New("null content received from LLM")

	// ErrRateLimited is returned when the provider rate limits the request
	ErrRateLimited = errors.New("rate limited by provider")

	// ErrInvalidResponse is returned when the provider returns an invalid response
	ErrInvalidResponse = errors.New("invalid response from provider")

	// ErrContextTooLarge is returned when the input exceeds the model's context window
	ErrContextTooLarge = errors.New("input exceeds model context window")
)

// IsNullContentError checks if an error is or wraps a null content error
func IsNullContentError(err error) bool {
	return errors.Is(err, ErrNullContent)
}

// WrapError wraps an error with additional context
func WrapError(err error, message string) error {
	return fmt.Errorf("%s: %w", message, err)
}
