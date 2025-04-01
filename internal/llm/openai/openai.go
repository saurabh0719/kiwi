package openai

import (
	"context"
	"fmt"
	"os"

	openaiapi "github.com/sashabaranov/go-openai"
	"github.com/saurabh0719/kiwi/internal/llm/core"
	"github.com/saurabh0719/kiwi/internal/tools"
)

// Adapter implements the Adapter interface for OpenAI
type Adapter struct {
	client *openaiapi.Client
	model  string
	tools  *tools.Registry
}

// New creates a new OpenAI adapter
func New(model, apiKey string, tools *tools.Registry) (*Adapter, error) {
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("OpenAI API key not found")
		}
	}

	client := openaiapi.NewClient(apiKey)
	return &Adapter{
		client: client,
		model:  model,
		tools:  tools,
	}, nil
}

// Chat sends a message to OpenAI and returns the response
func (a *Adapter) Chat(ctx context.Context, messages []core.Message) (string, error) {
	// Build system prompt with tools
	systemPrompt := core.DefaultSystemPrompt
	if a.tools != nil {
		systemPrompt += "\n\n" + a.tools.GetToolsDescription()
	}

	openaiMessages := []openaiapi.ChatCompletionMessage{
		{
			Role:    "system",
			Content: systemPrompt,
		},
	}

	for _, msg := range messages {
		openaiMessages = append(openaiMessages, openaiapi.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	resp, err := a.client.CreateChatCompletion(
		ctx,
		openaiapi.ChatCompletionRequest{
			Model:       a.model,
			Messages:    openaiMessages,
			Temperature: 0.7,
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to create chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no completion choices returned")
	}

	return resp.Choices[0].Message.Content, nil
}

// GetModel returns the model name being used
func (a *Adapter) GetModel() string {
	return a.model
}

// GetProvider returns the provider name
func (a *Adapter) GetProvider() string {
	return "openai"
}

// Complete sends a completion request to OpenAI
func (a *Adapter) Complete(ctx context.Context, prompt string) (string, error) {
	resp, err := a.client.CreateCompletion(
		ctx,
		openaiapi.CompletionRequest{
			Model:       a.model,
			Prompt:      prompt,
			MaxTokens:   2000,
			Temperature: 0.7,
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to create completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no completion choices returned")
	}

	return resp.Choices[0].Text, nil
}

// Name returns the name of the adapter
func (a *Adapter) Name() string {
	return "OpenAI"
}
