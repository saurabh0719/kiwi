package llm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/saurabh0719/kiwi/internal/llm/core"
)

// ChatMessage represents a chat message for the shared components
type ChatMessage struct {
	Role    string
	Content string
}

// ToolExecutionDetector is a function that checks if a response chunk contains
// evidence of tool execution
type ToolExecutionDetector func(chunk string) bool

// ToolExecutionState tracks details about tool execution
type ToolExecutionState struct {
	Executed bool   // Whether a tool was executed at all
	Failed   bool   // Whether the execution failed
	Output   string // Output of the execution if any
}

// DefaultToolExecutionDetector provides a standard way to detect tool execution markers
func DefaultToolExecutionDetector(chunk string) bool {
	return strings.Contains(chunk, "tool") ||
		strings.Contains(chunk, "executing:") ||
		strings.Contains(chunk, "executed in")
}

// IsToolExecutionFailed checks if a tool execution failed
func IsToolExecutionFailed(chunk string) bool {
	return strings.Contains(chunk, "execution failed") ||
		strings.Contains(chunk, "failed:") ||
		strings.Contains(chunk, "All 3 attempts failed")
}

// GetToolExecutionState analyzes a response to determine tool execution state
func GetToolExecutionState(response string) ToolExecutionState {
	state := ToolExecutionState{
		Executed: false,
		Failed:   false,
		Output:   "",
	}

	// Check if any tool execution was attempted
	if DefaultToolExecutionDetector(response) {
		state.Executed = true

		// Check if the execution failed
		if IsToolExecutionFailed(response) {
			state.Failed = true
		}

		// Try to extract the tool output if available
		// This is a simple heuristic - could be improved
		if strings.Contains(response, "Output:") {
			parts := strings.Split(response, "Output:")
			if len(parts) > 1 {
				state.Output = strings.TrimSpace(parts[1])
			}
		}
	}

	return state
}

// HandleNullContentError checks if an error is a null content error and determines
// if it should be treated as a successful response based on the toolExecuted flag
func HandleNullContentError(err error, toolExecuted bool) bool {
	return core.IsNullContentError(err) && toolExecuted
}

// ProcessStreamWithToolDetection handles a streaming response while detecting tool usage
func ProcessStreamWithToolDetection(
	ctx context.Context,
	adapter core.Adapter,
	messages []Message,
	streamHandler core.StreamHandler,
	toolDetector ToolExecutionDetector) (*core.ResponseMetrics, bool, string, error) {

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
		IsToolExecutionFailed(chunk)

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

// HandleStreamError processes errors from stream responses, with special handling for
// null content errors after tool execution
func HandleStreamError(
	err error,
	toolExecuted bool,
	startTime time.Time) (*core.ResponseMetrics, error) {

	// If this is a null content error after successful tool execution,
	// we can ignore it since we've already gotten output from tool execution
	if HandleNullContentError(err, toolExecuted) {
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
	toolDetector ToolExecutionDetector) (string, bool, *core.ResponseMetrics, error) {

	// Get the response at once
	response, metrics, err := adapter.ChatWithMetrics(ctx, messages)

	// Check if the response contains evidence of tool execution
	toolExecuted := false
	if err == nil && response != "" {
		toolExecuted = toolDetector(response)
	}

	return response, toolExecuted, metrics, err
}
