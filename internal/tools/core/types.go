package core

import "context"

// Parameter represents a parameter for a tool
type Parameter struct {
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Default     interface{} `json:"default,omitempty"`
}

// Tool is the interface for tools that can be used by LLMs
type Tool interface {
	// Name returns the name of the tool
	Name() string

	// Description returns a description of the tool
	Description() string

	// Parameters returns the parameters for the tool
	Parameters() map[string]Parameter

	// Execute runs the tool with the provided parameters
	Execute(ctx context.Context, params map[string]interface{}) (ToolExecutionResult, error)

	// RequiresConfirmation returns true if this tool should ask for user confirmation before executing
	RequiresConfirmation() bool
}

// ToolExecutionResult is the result of a tool execution
type ToolExecutionResult struct {
	ToolMethod         string   `json:"tool_method,omitempty"`
	ToolExecutionSteps []string `json:"tool_execution_steps,omitempty"`
	Output             string   `json:"output"`
}

// AddStep adds an execution step to the result
func (r *ToolExecutionResult) AddStep(step string) {
	r.ToolExecutionSteps = append(r.ToolExecutionSteps, step)
}

// Factory is a function type that creates new tools
type Factory func() Tool
