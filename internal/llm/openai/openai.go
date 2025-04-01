package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	openaiapi "github.com/sashabaranov/go-openai"
	"github.com/saurabh0719/kiwi/internal/llm/core"
	"github.com/saurabh0719/kiwi/internal/tools"
	"github.com/saurabh0719/kiwi/internal/util"
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
	response, _, err := a.ChatWithMetrics(ctx, messages)
	return response, err
}

// prepareTools converts the available tools to OpenAI tool definitions
func (a *Adapter) prepareTools() []openaiapi.Tool {
	if a.tools == nil {
		return nil
	}

	var tools []openaiapi.Tool
	toolsList := a.tools.List()

	for _, tool := range toolsList {
		// Create a JSON schema for the function parameters
		params := tool.Parameters()

		// Create a JSON schema structure
		schema := map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
			"required":   []string{},
		}

		properties := schema["properties"].(map[string]interface{})
		required := schema["required"].([]string)

		for name, param := range params {
			properties[name] = map[string]interface{}{
				"type":        param.Type,
				"description": param.Description,
			}

			if param.Required {
				required = append(required, name)
			}
		}

		schema["required"] = required

		// Create a tool with a function definition
		tools = append(tools, openaiapi.Tool{
			Type: openaiapi.ToolTypeFunction,
			Function: &openaiapi.FunctionDefinition{
				Name:        tool.Name(),
				Description: tool.Description(),
				Parameters:  schema,
			},
		})
	}

	return tools
}

// ChatWithMetrics sends a message to OpenAI and returns the response with metrics
func (a *Adapter) ChatWithMetrics(ctx context.Context, messages []core.Message) (string, *core.ResponseMetrics, error) {
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

	// Set up the request
	req := openaiapi.ChatCompletionRequest{
		Model:       a.model,
		Messages:    openaiMessages,
		Temperature: 0.7,
	}

	// Add tool calling capability if we have tools
	if a.tools != nil && len(a.tools.List()) > 0 {
		tools := a.prepareTools()
		if len(tools) > 0 {
			req.Tools = tools
			req.ToolChoice = "auto"
		}
	}

	startTime := time.Now()
	var finalResponse string
	var responseTime time.Duration
	var totalPromptTokens, totalCompletionTokens int

	// Maximum number of function call iterations to prevent infinite loops
	maxCalls := 10
	callCount := 0

	// Process the conversation with potential function calls
	for callCount < maxCalls {
		callCount++

		resp, err := a.client.CreateChatCompletion(ctx, req)
		if err != nil {
			return "", nil, fmt.Errorf("failed to create chat completion: %w", err)
		}

		if len(resp.Choices) == 0 {
			return "", nil, fmt.Errorf("no completion choices returned")
		}

		// Track token usage
		totalPromptTokens += resp.Usage.PromptTokens
		totalCompletionTokens += resp.Usage.CompletionTokens

		choice := resp.Choices[0]

		// Check if there's a tool call in the response
		if choice.Message.ToolCalls != nil && len(choice.Message.ToolCalls) > 0 {
			// Add the assistant's message with the tool calls to our conversation
			openaiMessages = append(openaiMessages, choice.Message)

			// Process each tool call
			for _, toolCall := range choice.Message.ToolCalls {
				if toolCall.Type != openaiapi.ToolTypeFunction {
					continue
				}

				// Get the function name and arguments
				functionName := toolCall.Function.Name
				functionArgs := toolCall.Function.Arguments

				// Get the tool to execute
				tool, exists := a.tools.Get(functionName)
				if !exists {
					functionResult := fmt.Sprintf("Error: Function %s not found", functionName)

					// Add the function result to our conversation
					openaiMessages = append(openaiMessages, openaiapi.ChatCompletionMessage{
						Role:       "tool",
						ToolCallID: toolCall.ID,
						Content:    functionResult,
					})

					continue
				}

				// Parse the arguments
				var args map[string]interface{}
				if err := json.Unmarshal([]byte(functionArgs), &args); err != nil {
					functionResult := fmt.Sprintf("Error parsing arguments: %v", err)

					// Add the function result to our conversation
					openaiMessages = append(openaiMessages, openaiapi.ChatCompletionMessage{
						Role:       "tool",
						ToolCallID: toolCall.ID,
						Content:    functionResult,
					})

					continue
				}

				// Execute the tool with visual feedback
				functionResult, err := util.ExecuteToolWithFeedback(ctx, tool, args)
				if err != nil {
					functionResult = fmt.Sprintf("Error executing function: %v", err)
				}

				// Add the function result to our conversation
				openaiMessages = append(openaiMessages, openaiapi.ChatCompletionMessage{
					Role:       "tool",
					ToolCallID: toolCall.ID,
					Content:    functionResult,
				})
			}

			// Update the request with the new messages
			req.Messages = openaiMessages

			continue
		}

		// No tool call, so we have our final response
		finalResponse = choice.Message.Content
		responseTime = time.Since(startTime)
		break
	}

	metrics := &core.ResponseMetrics{
		PromptTokens:     totalPromptTokens,
		CompletionTokens: totalCompletionTokens,
		TotalTokens:      totalPromptTokens + totalCompletionTokens,
		ResponseTime:     responseTime,
	}

	return finalResponse, metrics, nil
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
