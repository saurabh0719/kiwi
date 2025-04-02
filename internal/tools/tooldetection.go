package tools

import (
	"errors"
	"strings"
)

// Common error types
var ErrNullContent = errors.New("null content returned")

// IsNullContentError checks if an error is a null content error
func IsNullContentError(err error) bool {
	return err != nil && errors.Is(err, ErrNullContent)
}

// ExecutionDetector is a function that checks if a response chunk contains
// evidence of tool execution
type ExecutionDetector func(chunk string) bool

// ExecutionState tracks details about tool execution
type ExecutionState struct {
	Executed bool   // Whether a tool was executed at all
	Failed   bool   // Whether the execution failed
	Output   string // Output of the execution if any
}

// DefaultExecutionDetector provides a standard way to detect tool execution markers
func DefaultExecutionDetector(chunk string) bool {
	return strings.Contains(chunk, "tool") ||
		strings.Contains(chunk, "executing:") ||
		strings.Contains(chunk, "executed in")
}

// IsExecutionFailed checks if a tool execution failed
func IsExecutionFailed(chunk string) bool {
	return strings.Contains(chunk, "execution failed") ||
		strings.Contains(chunk, "failed:") ||
		strings.Contains(chunk, "All 3 attempts failed")
}

// GetExecutionState analyzes a response to determine tool execution state
func GetExecutionState(response string) ExecutionState {
	state := ExecutionState{
		Executed: false,
		Failed:   false,
		Output:   "",
	}

	// Check if any tool execution was attempted
	if DefaultExecutionDetector(response) {
		state.Executed = true

		// Check if the execution failed
		if IsExecutionFailed(response) {
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
	return IsNullContentError(err) && toolExecuted
}
