package shell

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/saurabh0719/kiwi/internal/tools/core"
)

// Tool provides sandboxed shell command execution
type Tool struct {
	// allowedCommands is a whitelist of commands that can be executed
	allowedCommands []string
}

// New creates a new ShellTool
func New() *Tool {
	return &Tool{
		allowedCommands: []string{
			"ls", "cat", "grep", "find", "pwd",
			"head", "tail", "wc", "echo", "date",
			"ps", "df", "du", "free", "top",
		},
	}
}

// Name returns the name of the tool
func (t *Tool) Name() string {
	return "shell"
}

// Description returns the description of the tool
func (t *Tool) Description() string {
	return "Executes shell commands in a sandboxed environment"
}

// Parameters returns the parameters for the tool
func (t *Tool) Parameters() map[string]core.Parameter {
	return map[string]core.Parameter{
		"command": {
			Type:        "string",
			Description: "Command to execute",
			Required:    true,
		},
		"args": {
			Type:        "array",
			Description: "Command arguments",
			Required:    false,
			Default:     []string{},
		},
	}
}

// Execute executes the tool with the given arguments
func (t *Tool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	command, ok := args["command"].(string)
	if !ok {
		return "", fmt.Errorf("command must be a string")
	}

	// Check if command is allowed
	if !t.isCommandAllowed(command) {
		return "", fmt.Errorf("command not allowed: %s", command)
	}

	// Get command arguments
	var cmdArgs []string
	if argsRaw, ok := args["args"].([]interface{}); ok {
		for _, arg := range argsRaw {
			if strArg, ok := arg.(string); ok {
				cmdArgs = append(cmdArgs, strArg)
			}
		}
	}

	// Create command
	cmd := exec.CommandContext(ctx, command, cmdArgs...)

	// Run command and capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}

// isCommandAllowed checks if a command is in the whitelist
func (t *Tool) isCommandAllowed(command string) bool {
	command = strings.TrimSpace(command)
	for _, allowed := range t.allowedCommands {
		if command == allowed {
			return true
		}
	}
	return false
}
