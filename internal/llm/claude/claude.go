package claude

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/saurabh0719/kiwi/internal/llm/core"
	"github.com/saurabh0719/kiwi/internal/tools"
)

const (
	apiEndpoint = "https://api.anthropic.com/v1/messages"
)

// Adapter implements the Adapter interface for Anthropic's Claude
type Adapter struct {
	apiKey string
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

	return &Adapter{
		apiKey: apiKey,
		model:  model,
		tools:  tools,
	}, nil
}

type claudeRequest struct {
	Model       string         `json:"model"`
	Messages    []core.Message `json:"messages"`
	MaxTokens   int            `json:"max_tokens"`
	Temperature float64        `json:"temperature"`
}

type claudeResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
}

// Chat sends a message to Claude and returns the response
func (a *Adapter) Chat(ctx context.Context, messages []core.Message) (string, error) {
	systemPrompt := core.DefaultSystemPrompt
	if a.tools != nil {
		systemPrompt += "\n\n" + a.tools.GetToolsDescription()
	}

	claudeMessages := []core.Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
	}

	claudeMessages = append(claudeMessages, messages...)

	reqBody := claudeRequest{
		Model:       a.model,
		Messages:    claudeMessages,
		MaxTokens:   4096,
		Temperature: 0.7,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiEndpoint, strings.NewReader(string(jsonData)))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var claudeResp claudeResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(claudeResp.Content) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	return claudeResp.Content[0].Text, nil
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
