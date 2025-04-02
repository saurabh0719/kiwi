package llm

import (
	"context"
	"fmt"
	"time"

	"github.com/saurabh0719/kiwi/internal/llm/core"
	"github.com/saurabh0719/kiwi/internal/tools"
	"github.com/saurabh0719/kiwi/internal/util"
)

// Chat represents a complete conversation with an LLM
func Chat(ctx context.Context, adapter core.Adapter, messages []Message) (string, error) {
	// Simple chat without streaming or metrics
	response, _, err := adapter.ChatWithMetrics(ctx, messages)
	return response, err
}

// ChatWithToolDetection adds tool detection to streaming or non-streaming chat
func ChatWithToolDetection(
	ctx context.Context,
	adapter core.Adapter,
	messages []Message,
	useStreaming bool,
	spinner *util.SpinnerManager) (string, *core.ResponseMetrics, error) {

	if useStreaming {
		// Process with streaming to handle tools
		metrics, toolExecuted, response, err := ProcessStreamWithToolDetection(
			ctx,
			adapter,
			messages,
			DefaultStreamHandler(spinner),
			tools.DefaultExecutionDetector)

		// Special handling for null content errors after tool execution
		if err != nil {
			startTime := time.Now()
			handledMetrics, handledErr := HandleStreamError(err, toolExecuted, startTime)
			if handledErr == nil {
				// Error was handled, use metrics and empty response
				return "Command executed successfully.", handledMetrics, nil
			}
			// Error wasn't handled, pass through
			return "", nil, err
		}

		// Successful response
		return response, metrics, nil

	} else {
		// For non-streaming, use the non-streaming handler
		response, toolExecuted, metrics, err := HandleNonStreamingResponse(
			ctx,
			adapter,
			messages,
			tools.DefaultExecutionDetector)

		// Handle special error cases
		if err != nil && tools.HandleNullContentError(err, toolExecuted) {
			return "Command executed successfully.", metrics, nil
		}

		return response, metrics, err
	}
}

// HandleStreamError processes errors from stream responses, with special handling for
// null content errors after tool execution
func HandleStreamError(
	err error,
	toolExecuted bool,
	startTime time.Time) (*core.ResponseMetrics, error) {

	// If this is a null content error after successful tool execution,
	// we can ignore it since we've already gotten output from tool execution
	if tools.HandleNullContentError(err, toolExecuted) {
		// Create minimal metrics
		metrics := &core.ResponseMetrics{
			ResponseTime: time.Since(startTime),
		}

		// Return without error
		return metrics, nil
	}

	// For other errors, return the error
	return nil, fmt.Errorf("failed to get response: %w", err)
}

// HandleNonStreamingResponse processes a complete non-streaming response
func HandleNonStreamingResponse(
	ctx context.Context,
	adapter core.Adapter,
	messages []Message,
	toolDetector func(string) bool) (string, bool, *core.ResponseMetrics, error) {

	// Get the response at once
	response, metrics, err := adapter.ChatWithMetrics(ctx, messages)

	// Check if the response contains evidence of tool execution
	toolExecuted := false
	if err == nil && response != "" {
		toolExecuted = toolDetector(response)
	}

	return response, toolExecuted, metrics, err
}

// DefaultStreamHandler returns a standard stream handler function
func DefaultStreamHandler(spinner *util.SpinnerManager) func(string) error {
	// Define standard behavior for handling each chunk
	return func(chunk string) error {
		// If spinner is provided, ensure it's not running while we print
		if spinner != nil {
			spinner.TransitionToResponse()
		}

		// Print each chunk as it comes in
		fmt.Print(chunk)
		return nil
	}
}

// ProcessStreamWithToolDetection handles a streaming response while detecting tool usage
func ProcessStreamWithToolDetection(
	ctx context.Context,
	adapter core.Adapter,
	messages []Message,
	streamHandler core.StreamHandler,
	toolDetector func(string) bool) (*core.ResponseMetrics, bool, string, error) {

	// Initialize tracking variables
	var completeResponse string
	toolExecuted := false

	// Process the streaming response
	metrics, err := adapter.ChatStream(ctx, messages, func(chunk string) error {
		// Check for tool execution markers
		if toolDetector(chunk) {
			toolExecuted = true
		}

		// Check for tool failure markers (just log but don't store for now)
		// Can be extended in the future if needed
		tools.IsExecutionFailed(chunk)

		// Append the chunk to the complete response
		completeResponse += chunk

		// Pass to the original handler
		return streamHandler(chunk)
	})

	// If metrics is nil (can happen if the stream fails), create minimal metrics
	if metrics == nil {
		metrics = &core.ResponseMetrics{
			ResponseTime: 0,
		}
	}

	return metrics, toolExecuted, completeResponse, err
}
