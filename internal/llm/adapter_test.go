package llm

import (
	"context"
	"os"
	"testing"

	"github.com/saurabh0719/kiwi/internal/tools"
)

func TestNewAdapter(t *testing.T) {
	toolRegistry := tools.NewRegistry()

	tests := []struct {
		name      string
		provider  string
		model     string
		apiKey    string
		wantError bool
	}{
		{
			name:      "OpenAI valid",
			provider:  "openai",
			model:     "gpt-3.5-turbo",
			apiKey:    "test-key",
			wantError: false,
		},
		// Claude support will be added in the future
		// {
		//     name:      "Claude valid",
		//     provider:  "claude",
		//     model:     "claude-3-opus-20240229",
		//     apiKey:    "test-key",
		//     wantError: false,
		// },
		{
			name:      "Unknown provider",
			provider:  "unknown",
			model:     "model",
			apiKey:    "key",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter, err := NewAdapter(tt.provider, tt.model, tt.apiKey, toolRegistry)
			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if adapter.GetProvider() != tt.provider {
				t.Errorf("wrong provider: got %s, want %s", adapter.GetProvider(), tt.provider)
			}
			if adapter.GetModel() != tt.model {
				t.Errorf("wrong model: got %s, want %s", adapter.GetModel(), tt.model)
			}
		})
	}
}

func TestOpenAIAdapter(t *testing.T) {
	// Skip if no API key is available
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	// Create adapter using NewAdapter instead of direct factory method
	adapter, err := NewAdapter("openai", "gpt-3.5-turbo", apiKey, nil)
	if err != nil {
		t.Fatalf("failed to create adapter: %v", err)
	}

	ctx := context.Background()
	messages := []Message{
		{Role: "user", Content: "Say hello"},
	}

	response, err := adapter.Chat(ctx, messages)
	if err != nil {
		t.Errorf("Chat failed: %v", err)
	}
	if response == "" {
		t.Error("empty response")
	}
}

// Claude support will be added in the future
// func TestClaudeAdapter(t *testing.T) {
//     // Skip if no API key is available
//     apiKey := os.Getenv("ANTHROPIC_API_KEY")
//     if apiKey == "" {
//         t.Skip("ANTHROPIC_API_KEY not set")
//     }
//
//     // Create adapter using NewAdapter instead of direct factory method
//     adapter, err := NewAdapter("claude", "claude-3-opus-20240229", apiKey, nil)
//     if err != nil {
//         t.Fatalf("failed to create adapter: %v", err)
//     }
//
//     ctx := context.Background()
//     messages := []Message{
//         {Role: "user", Content: "Say hello"},
//     }
//
//     response, err := adapter.Chat(ctx, messages)
//     if err != nil {
//         t.Errorf("Chat failed: %v", err)
//     }
//     if response == "" {
//         t.Error("empty response")
//     }
// }
