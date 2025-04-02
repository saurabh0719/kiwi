package shell

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/saurabh0719/kiwi/internal/tools/core"
)

// Tool provides shell command execution
type Tool struct {
	name        string
	description string
	parameters  map[string]core.Parameter
}

// New creates a new ShellTool
func New() *Tool {
	parameters := map[string]core.Parameter{
		"command": {
			Type:        "string",
			Description: "Command to execute",
			Required:    true,
		},
	}

	return &Tool{
		name:        "shell",
		description: "Executes shell commands",
		parameters:  parameters,
	}
}

// Name returns the name of the tool
func (t *Tool) Name() string {
	return t.name
}

// Description returns the description of the tool
func (t *Tool) Description() string {
	return t.description
}

// Parameters returns the parameters for the tool
func (t *Tool) Parameters() map[string]core.Parameter {
	return t.parameters
}

// RequiresConfirmation returns true because shell commands should always require confirmation
func (t *Tool) RequiresConfirmation() bool {
	return true
}

// Execute executes the tool with the given arguments
func (t *Tool) Execute(ctx context.Context, args map[string]interface{}) (core.ToolExecutionResult, error) {
	result := core.ToolExecutionResult{
		ToolMethod: "",
		Output:     "",
	}

	commandLine, ok := args["command"].(string)
	if !ok {
		return result, fmt.Errorf("command must be a string")
	}

	// Extract the base command for the method name
	methodName := "execute"
	parts := strings.Fields(commandLine)
	if len(parts) > 0 {
		methodName = parts[0]
	}
	result.ToolMethod = methodName

	result.AddStep(fmt.Sprintf("Command requested: %s", commandLine))

	// Process the command
	var output string
	var err error

	// Always use shell for command execution to properly handle
	// command sequences (&&, ||, ;) and special characters
	result.AddStep("Executing command with shell...")
	output, err = t.executeWithShell(ctx, commandLine)
	if err != nil {
		result.AddStep(fmt.Sprintf("Command execution failed: %v", err))
		return result, err
	}

	outputLines := strings.Count(output, "\n") + 1
	outputBytes := len(output)
	result.AddStep(fmt.Sprintf("Command completed successfully with %d lines (%d bytes) of output", outputLines, outputBytes))

	result.Output = output
	return result, nil
}

// executeWithShell runs a command using the shell to handle pipes, flags, and command sequences
func (t *Tool) executeWithShell(ctx context.Context, commandLine string) (string, error) {
	// Use bash to execute the command with proper handling of flags and operators
	cmd := exec.CommandContext(ctx, "bash", "-c", commandLine)

	// Get current working directory to execute in
	currentDir, err := os.Getwd()
	if err == nil {
		cmd.Dir = currentDir
	}

	// Set up environment
	cmd.Env = os.Environ()

	// Run command and capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}
