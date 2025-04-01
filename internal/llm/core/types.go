package core

import (
	"context"
	"time"

	"github.com/saurabh0719/kiwi/internal/tools"
)

// Message represents a message in a conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ResponseMetrics contains metrics about the response
type ResponseMetrics struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	ResponseTime     time.Duration
}

// Adapter is the interface for LLM adapters
type Adapter interface {
	// Chat sends a message to the LLM and returns the response
	Chat(ctx context.Context, messages []Message) (string, error)
	// ChatWithMetrics sends a message to the LLM and returns the response with metrics
	ChatWithMetrics(ctx context.Context, messages []Message) (string, *ResponseMetrics, error)
	// GetModel returns the model name
	GetModel() string
	// GetProvider returns the provider name
	GetProvider() string
}

// Factory is a function type that creates new adapters
type Factory func(model, apiKey string, tools *tools.Registry) (Adapter, error)
