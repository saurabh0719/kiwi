package core

import (
	"context"

	"github.com/saurabh0719/kiwi/internal/tools"
)

// Message represents a message in a conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Adapter is the interface for LLM adapters
type Adapter interface {
	// Chat sends a message to the LLM and returns the response
	Chat(ctx context.Context, messages []Message) (string, error)
	// GetModel returns the model name
	GetModel() string
	// GetProvider returns the provider name
	GetProvider() string
}

// Factory is a function type that creates new adapters
type Factory func(model, apiKey string, tools *tools.Registry) (Adapter, error)
