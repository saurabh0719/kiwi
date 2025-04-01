package claude

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/saurabh0719/kiwi/internal/llm/core"
	"github.com/saurabh0719/kiwi/internal/tools"
)

// Adapter implements the Adapter interface for Claude
type Adapter struct {
	client *openai.Client
	model  string
	tools  *tools.Registry
}

// New creates a new Claude adapter
func New(model, apiKey string, tools *tools.Registry) (*Adapter, error) {
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("Anthropic API key not found")
		}
	}

	client := openai.NewClient(apiKey)
	return &Adapter{
		client: client,
		model:  model,
		tools:  tools,
	}, nil
}

// Chat sends a message to Claude and returns the response
func (a *Adapter) Chat(ctx context.Context, messages []core.Message) (string, error) {
	response, _, err := a.ChatWithMetrics(ctx, messages)
	return response, err
}

// ChatWithMetrics sends a message to Claude and returns the response with metrics
func (a *Adapter) ChatWithMetrics(ctx context.Context, messages []core.Message) (string, *core.ResponseMetrics, error) {
	startTime := time.Now()

	// Build the system prompt
	var userPrompt string
	var systemPrompt string
	if a.tools != nil {
		systemPrompt = core.DefaultSystemPrompt + "\n\n" + a.tools.GetToolsDescription()
	} else {
		systemPrompt = core.DefaultSystemPrompt
	}

	// Extract the user's message
	// This is a simplified approach - in reality, Claude API has different requirements
	// for message formats than the OpenAI API
	for _, msg := range messages {
		if msg.Role == "user" {
			userPrompt = msg.Content
			break
		}
	}

	// Claude API typically expects a single prompt so we combine the system and user prompts
	combinedPrompt := fmt.Sprintf("System: %s\n\nHuman: %s\n\nAssistant:", systemPrompt, userPrompt)

	// Note: This is a simplified example - Claude API integration would need to be implemented properly
	// Here we mock using OpenAI client for example, but would need to be replaced with Claude's API client
	resp, err := a.client.CreateCompletion(
		ctx,
		openai.CompletionRequest{
			Model:       a.model,
			Prompt:      combinedPrompt,
			MaxTokens:   2000,
			Temperature: 0.7,
		},
	)
	responseTime := time.Since(startTime)

	if err != nil {
		return "", nil, fmt.Errorf("failed to create completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", nil, fmt.Errorf("no completion choices returned")
	}

	// Claude API doesn't provide token usage in the same way as OpenAI
	// This is a simplified approach estimating token usage
	estimatedPromptTokens := len(combinedPrompt) / 4
	estimatedCompletionTokens := len(resp.Choices[0].Text) / 4

	metrics := &core.ResponseMetrics{
		PromptTokens:     estimatedPromptTokens,
		CompletionTokens: estimatedCompletionTokens,
		TotalTokens:      estimatedPromptTokens + estimatedCompletionTokens,
		ResponseTime:     responseTime,
	}

	return resp.Choices[0].Text, metrics, nil
}

// GetModel returns the model name being used
func (a *Adapter) GetModel() string {
	return a.model
}

// GetProvider returns the provider name
func (a *Adapter) GetProvider() string {
	return "claude"
}

// Name returns the name of the adapter
func (a *Adapter) Name() string {
	return "Claude"
}
