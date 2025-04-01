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
	Execute(ctx context.Context, params map[string]interface{}) (string, error)
}

// Factory is a function type that creates new tools
type Factory func() Tool
