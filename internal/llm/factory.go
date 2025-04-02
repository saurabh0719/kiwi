package llm

import (
	"fmt"

	"github.com/saurabh0719/kiwi/internal/llm/core"
	"github.com/saurabh0719/kiwi/internal/llm/openai"
	"github.com/saurabh0719/kiwi/internal/tools"
)

// Message is an alias for core.Message for backward compatibility
type Message = core.Message

// Adapter is an alias for core.Adapter for backward compatibility
type Adapter = core.Adapter

// DefaultSystemPrompt is an alias for core.DefaultSystemPrompt for backward compatibility
var DefaultSystemPrompt = core.DefaultSystemPrompt

// NewAdapter creates a new adapter for the specified provider
func NewAdapter(provider, model, apiKey string, tools *tools.Registry) (Adapter, error) {
	switch provider {
	case "openai":
		return openai.New(model, apiKey, tools)
	// Claude support will be added in the future
	// case "claude":
	//   return claude.New(model, apiKey, tools)
	default:
		return nil, fmt.Errorf("unsupported provider: %s (only 'openai' is currently supported)", provider)
	}
}
