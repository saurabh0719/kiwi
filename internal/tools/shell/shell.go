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
	}
}

// Execute executes the tool with the given arguments
func (t *Tool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	commandLine, ok := args["command"].(string)
	if !ok {
		return "", fmt.Errorf("command must be a string")
	}

	// Split the command line into command and arguments
	parts := strings.Fields(commandLine)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty command")
	}

	baseCommand := parts[0]
	cmdArgs := parts[1:]

	// Check if base command is allowed
	if !t.isCommandAllowed(baseCommand) {
		return "", fmt.Errorf("command not allowed: %s", baseCommand)
	}

	// Create command
	cmd := exec.CommandContext(ctx, baseCommand, cmdArgs...)

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
