package tools

import (
	"context"
	"encoding/json"

	"github.com/saurabh0719/kiwi/internal/tools/core"
	"github.com/saurabh0719/kiwi/internal/tools/filesystem"
	"github.com/saurabh0719/kiwi/internal/tools/shell"
	"github.com/saurabh0719/kiwi/internal/tools/sysinfo"
)

// Parameter represents a parameter for a tool
type Parameter struct {
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Default     interface{} `json:"default,omitempty"`
}

// Tool is the interface for all tools
type Tool interface {
	// Name returns the name of the tool
	Name() string

	// Description returns the description of the tool
	Description() string

	// Parameters returns the parameters for the tool
	Parameters() map[string]Parameter

	// Execute executes the tool with the given arguments
	Execute(ctx context.Context, args map[string]interface{}) (interface{}, error)
}

// Factory is a function type that creates new tools
type Factory func() Tool

// Registry manages the available tools
type Registry struct {
	tools map[string]core.Tool
}

// NewRegistry creates a new tool registry
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]core.Tool),
	}
}

// Register adds a tool to the registry
func (r *Registry) Register(tool core.Tool) {
	r.tools[tool.Name()] = tool
}

// Get returns a tool by name
func (r *Registry) Get(name string) (core.Tool, bool) {
	tool, ok := r.tools[name]
	return tool, ok
}

// List returns all registered tools
func (r *Registry) List() []core.Tool {
	tools := make([]core.Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

// GetToolsDescription returns a description of all tools in a format suitable for the LLM
func (r *Registry) GetToolsDescription() string {
	desc := "Available tools:\n"
	for _, tool := range r.tools {
		desc += "- " + tool.Name() + ": " + tool.Description() + "\n"
		desc += "  Parameters:\n"
		for name, param := range tool.Parameters() {
			required := ""
			if param.Required {
				required = " (required)"
			}
			desc += "  - " + name + required + ": " + param.Description + "\n"
		}
		desc += "\n"
	}
	return desc
}

// ToJSON converts a tool's parameters to a JSON schema
func (r *Registry) ToJSON() (string, error) {
	schema := make(map[string]interface{})
	for name, tool := range r.tools {
		schema[name] = map[string]interface{}{
			"description": tool.Description(),
			"parameters":  tool.Parameters(),
		}
	}

	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// RegisterStandardTools initializes and registers the default set of tools
func RegisterStandardTools(registry *Registry) {
	// Register default tools
	registry.Register(NewFileSystemTool())
	registry.Register(NewShellTool())
	registry.Register(NewSystemInfoTool())
}

// NewFileSystemTool creates a new FileSystemTool
func NewFileSystemTool() core.Tool {
	// Direct implementation that returns ToolExecutionResult
	return filesystem.New()
}

// NewShellTool creates a new ShellTool
func NewShellTool() core.Tool {
	// Direct implementation that returns ToolExecutionResult
	return shell.New()
}

// NewSystemInfoTool creates a new SystemInfoTool
func NewSystemInfoTool() core.Tool {
	// Direct implementation that returns ToolExecutionResult
	return sysinfo.New()
}
