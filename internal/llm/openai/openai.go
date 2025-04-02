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
	startTime := time.Now()
	var llmTime time.Duration
	var toolTime time.Duration

	// Ensure any existing spinner is stopped at the beginning
	spinnerManager := util.GetGlobalSpinnerManager()
	spinnerManager.TransitionToResponse()

	// Prepare the initial messages and request
	openaiMessages := a.prepareInitialMessages(messages)
	req := a.createChatCompletionRequest(openaiMessages, false) // false = not streaming

	var finalResponse string
	var totalPromptTokens, totalCompletionTokens int

	// Maximum number of function call iterations to prevent infinite loops
	maxCalls := 10
	callCount := 0

	// Process the conversation with potential function calls
	for callCount < maxCalls {
		callCount++

		// Start spinner for waiting for response
		if callCount == 1 { // Only for the first call
			spinnerManager.StartThinkingSpinner("Waiting for response...")
		}

		llmStartTime := time.Now()
		resp, err := a.client.CreateChatCompletion(ctx, req)
		llmTime += time.Since(llmStartTime)
		if err != nil {
			// Stop spinner on error
			spinnerManager.TransitionToResponse()
			return "", nil, fmt.Errorf("failed to create chat completion: %w", err)
		}

		if len(resp.Choices) == 0 {
			// Stop spinner when no choices are returned
			spinnerManager.TransitionToResponse()
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

			// Process each tool call and add results to messages
			// (Tools manage their own spinners)
			toolStartTime := time.Now()
			toolCallResults := a.processToolCalls(ctx, choice.Message.ToolCalls)
			toolTime += time.Since(toolStartTime)
			openaiMessages = append(openaiMessages, toolCallResults...)

			// Update the request with the new messages
			req.Messages = openaiMessages

			// Start spinner for next iteration
			spinnerManager.StartThinkingSpinner("Continuing conversation...")
			continue
		}

		// No tool call, so we have our final response
		spinnerManager.TransitionToResponse() // Stop spinner before returning response
		finalResponse = choice.Message.Content
		break
	}

	// Ensure all spinners are stopped
	spinnerManager.StopAllSpinners()

	responseTime := time.Since(startTime)
	metrics := &core.ResponseMetrics{
		PromptTokens:     totalPromptTokens,
		CompletionTokens: totalCompletionTokens,
		TotalTokens:      totalPromptTokens + totalCompletionTokens,
		ResponseTime:     responseTime,
		LLMTime:          llmTime,
		ToolTime:         toolTime,
	}

	return finalResponse, metrics, nil
}

// prepareInitialMessages converts core.Message array to OpenAI messages with system prompt
func (a *Adapter) prepareInitialMessages(messages []core.Message) []openaiapi.ChatCompletionMessage {
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

	return openaiMessages
}

// createChatCompletionRequest creates a request object for the OpenAI API
func (a *Adapter) createChatCompletionRequest(messages []openaiapi.ChatCompletionMessage, streaming bool) openaiapi.ChatCompletionRequest {
	// Set up the request
	req := openaiapi.ChatCompletionRequest{
		Model:       a.model,
		Messages:    messages,
		Temperature: 0.7,
		Stream:      streaming,
	}

	// Add tool calling capability if we have tools
	if a.tools != nil && len(a.tools.List()) > 0 {
		tools := a.prepareTools()
		if len(tools) > 0 {
			req.Tools = tools
			req.ToolChoice = "auto"
		}
	}

	return req
}

// processToolCalls processes a slice of tool calls and returns their results as messages
func (a *Adapter) processToolCalls(ctx context.Context, toolCalls []openaiapi.ToolCall) []openaiapi.ChatCompletionMessage {
	var resultMessages []openaiapi.ChatCompletionMessage

	// Get the global spinner manager and clear any spinners before tool execution
	spinnerManager := util.GetGlobalSpinnerManager()
	spinnerManager.TransitionToResponse()

	for _, toolCall := range toolCalls {
		if toolCall.Type != openaiapi.ToolTypeFunction {
			continue
		}

		// Get the function name and arguments
		functionName := toolCall.Function.Name
		functionArgs := toolCall.Function.Arguments

		var functionResult string

		// Verify function arguments are present and valid
		if functionArgs == "" || functionArgs == "{}" || !json.Valid([]byte(functionArgs)) {
			functionResult = fmt.Sprintf("Error: Missing or invalid arguments for function %s. Please provide valid arguments.", functionName)
		} else {
			// Try to execute the function
			functionResult = a.executeToolCall(ctx, functionName, functionArgs, toolCall.ID)
		}

		// Add the function result to our conversation
		resultMessages = append(resultMessages, openaiapi.ChatCompletionMessage{
			Role:       "tool",
			ToolCallID: toolCall.ID,
			Content:    functionResult,
		})
	}

	return resultMessages
}

// executeToolCall executes a single tool call and returns the result
func (a *Adapter) executeToolCall(ctx context.Context, functionName, functionArgs, toolCallID string) string {
	// Get the global spinner manager
	spinnerManager := util.GetGlobalSpinnerManager()

	// Get the tool to execute
	tool, exists := a.tools.Get(functionName)
	if !exists {
		// Stop any spinner if no tool is found
		spinnerManager.TransitionToResponse()
		return fmt.Sprintf("Error: Function %s not found", functionName)
	}

	// Parse the arguments
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(functionArgs), &args); err != nil {
		// Stop any spinner on error
		spinnerManager.TransitionToResponse()
		return fmt.Sprintf("Error parsing arguments: %v", err)
	}

	// The spinners will be managed by the ExecuteToolWithFeedback function
	// So we just need to clear any existing spinners here
	spinnerManager.TransitionToResponse()

	// Execute the tool with feedback (will handle its own spinners)
	result, err := tools.ExecuteToolWithFeedback(ctx, tool, args)

	// Always clear spinners after tool execution
	spinnerManager.TransitionToResponse()

	// Start the next thinking spinner
	spinnerManager.StartThinkingSpinner("Continuing conversation...")

	if err != nil {
		return fmt.Sprintf("Error executing function: %v", err)
	}

	return result
}

// ChatStream sends a message to OpenAI and streams the response tokens to the handler function
func (a *Adapter) ChatStream(ctx context.Context, messages []core.Message, handler core.StreamHandler) (*core.ResponseMetrics, error) {
	startTime := time.Now()
	var llmTime time.Duration
	var toolTime time.Duration

	// Ensure any existing spinner is stopped at the beginning of a new chat stream
	spinnerManager := util.GetGlobalSpinnerManager()
	spinnerManager.TransitionToResponse()

	// Prepare initial messages and variables for tracking
	openaiMessages := a.prepareInitialMessages(messages)
	var totalTokensGenerated int
	var totalPromptTokens, totalCompletionTokens int

	// Maximum number of function call iterations
	maxCalls := 10
	callCount := 0

	// Process conversation with potential tool calls in a loop
	for callCount < maxCalls {
		callCount++

		// Process stream and check for tool calls
		llmStartTime := time.Now()
		toolCallDetected, streamTokens, err := a.processStream(ctx, openaiMessages, handler)
		llmTime += time.Since(llmStartTime)
		if err != nil {
			// Ensure any spinner is stopped on error
			spinnerManager.TransitionToResponse()
			return nil, err
		}

		totalTokensGenerated += streamTokens

		// If no tool calls were detected, we're done
		if !toolCallDetected {
			break
		}

		// Process tool calls through non-streaming API
		toolStartTime := time.Now()
		updatedMessages, promptTokens, completionTokens, err := a.processToolCallsNonStreaming(ctx, openaiMessages)
		toolTime += time.Since(toolStartTime)
		if err != nil {
			// Ensure any spinner is stopped on error
			spinnerManager.TransitionToResponse()
			return nil, err
		}

		openaiMessages = updatedMessages
		totalPromptTokens += promptTokens
		totalCompletionTokens += completionTokens
	}

	// Ensure all spinners are stopped at the end of the conversation
	spinnerManager.StopAllSpinners()

	responseTime := time.Since(startTime)

	// With streaming, we don't get accurate token counts, so this is an estimation
	metrics := &core.ResponseMetrics{
		PromptTokens:     totalPromptTokens,
		CompletionTokens: totalCompletionTokens + totalTokensGenerated,
		TotalTokens:      totalPromptTokens + totalCompletionTokens + totalTokensGenerated,
		ResponseTime:     responseTime,
		LLMTime:          llmTime,
		ToolTime:         toolTime,
	}

	return metrics, nil
}

// processStream handles the streaming part of the response and detects tool calls
func (a *Adapter) processStream(ctx context.Context, messages []openaiapi.ChatCompletionMessage, handler core.StreamHandler) (bool, int, error) {
	// Get the global spinner manager
	spinnerManager := util.GetGlobalSpinnerManager()

	// Create streaming request
	req := a.createChatCompletionRequest(messages, true) // true = streaming

	stream, err := a.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return false, 0, fmt.Errorf("failed to create chat completion stream: %w", err)
	}
	defer stream.Close()

	// Flag to track if we need to handle tool calls
	toolCallDetected := false
	var messageContent string
	tokensGenerated := 0
	firstToken := true

	// Process the stream
	for {
		response, err := stream.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return false, tokensGenerated, fmt.Errorf("stream error: %w", err)
		}

		// Skip empty choices
		if len(response.Choices) == 0 {
			continue
		}

		// Process each delta
		for _, choice := range response.Choices {
			// Collect content from the stream
			if choice.Delta.Content != "" {
				messageContent += choice.Delta.Content

				// On first token, ensure any tool execution spinner is stopped
				if firstToken {
					firstToken = false
					// Explicitly stop any running spinner to ensure clean transition
					spinnerManager.TransitionToResponse()
				}

				// Send the chunk to the handler
				if err := handler(choice.Delta.Content); err != nil {
					return false, tokensGenerated, fmt.Errorf("handler error: %w", err)
				}
				tokensGenerated++
			}

			// Check if there's a tool call in the delta
			if choice.Delta.ToolCalls != nil && len(choice.Delta.ToolCalls) > 0 {
				toolCallDetected = true
			}

			// If we get a finish reason, we need to consider if we're done
			if choice.FinishReason != "" {
				// If finish reason is "tool_calls", we need to process them
				if choice.FinishReason == "tool_calls" {
					toolCallDetected = true
				}
			}
		}
	}

	return toolCallDetected, tokensGenerated, nil
}

// processToolCallsNonStreaming handles tool calls by making a non-streaming request
func (a *Adapter) processToolCallsNonStreaming(ctx context.Context, messages []openaiapi.ChatCompletionMessage) ([]openaiapi.ChatCompletionMessage, int, int, error) {
	// Get the global spinner manager
	spinnerManager := util.GetGlobalSpinnerManager()

	// Start a spinner for this process
	spinnerManager.StartThinkingSpinner("Processing request...")

	// Create non-streaming request
	nonStreamReq := a.createChatCompletionRequest(messages, false) // false = not streaming

	// Make the request
	resp, err := a.client.CreateChatCompletion(ctx, nonStreamReq)
	if err != nil {
		// Stop spinner on error
		spinnerManager.TransitionToResponse()
		return messages, 0, 0, fmt.Errorf("failed to create non-streaming chat completion: %w", err)
	}

	// Track token usage
	promptTokens := resp.Usage.PromptTokens
	completionTokens := resp.Usage.CompletionTokens

	if len(resp.Choices) == 0 {
		// Stop spinner when no choices are available
		spinnerManager.TransitionToResponse()
		return messages, promptTokens, completionTokens, fmt.Errorf("no completion choices returned")
	}

	choice := resp.Choices[0]

	// Check if there's a tool call in the response
	if choice.Message.ToolCalls != nil && len(choice.Message.ToolCalls) > 0 {
		// Add the assistant's message with the tool calls to our conversation
		messages = append(messages, choice.Message)

		// Process each tool call - spinner management is done in the executeToolCall method
		toolCallResults := a.processToolCalls(ctx, choice.Message.ToolCalls)
		messages = append(messages, toolCallResults...)
	} else {
		// If no tool calls, stop the spinner
		spinnerManager.TransitionToResponse()
	}

	return messages, promptTokens, completionTokens, nil
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
