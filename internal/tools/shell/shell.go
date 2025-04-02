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
	name            string
	description     string
	parameters      map[string]core.Parameter
	allowedCommands []string
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
		description: "Executes shell commands in a sandboxed environment",
		parameters:  parameters,
		allowedCommands: []string{
			"ls", "cat", "grep", "find", "pwd",
			"head", "tail", "wc", "echo", "date",
			"ps", "df", "du", "free", "top",
		},
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

	// Improved command handling for pipes
	if strings.Contains(commandLine, "|") {
		result.AddStep("Detected pipe in command, using shell executor")

		// Verify pipeline commands
		commands := strings.Split(commandLine, "|")
		for i, cmd := range commands {
			trimmed := strings.TrimSpace(cmd)
			baseParts := strings.Fields(trimmed)
			if len(baseParts) == 0 {
				continue
			}

			baseCmd := baseParts[0]
			if !t.isCommandAllowed(baseCmd) {
				result.AddStep(fmt.Sprintf("Command not allowed in pipeline: %s", baseCmd))
				return result, fmt.Errorf("command not allowed in pipeline: %s", baseCmd)
			}
			result.AddStep(fmt.Sprintf("Verified command %d in pipeline: %s", i+1, baseCmd))
		}

		result.AddStep("Executing command with shell...")
		output, err = t.executeWithShell(ctx, commandLine)
		if err != nil {
			result.AddStep(fmt.Sprintf("Command execution failed: %v", err))
			return result, err
		}
	} else {
		// For simple commands, continue with whitelist check
		// Split the command line into command and arguments
		if len(parts) == 0 {
			result.AddStep("Error: Empty command")
			return result, fmt.Errorf("empty command")
		}

		baseCommand := parts[0]
		cmdArgs := parts[1:]

		result.AddStep(fmt.Sprintf("Checking if command is allowed: %s", baseCommand))

		// Check if base command is allowed
		if !t.isCommandAllowed(baseCommand) {
			result.AddStep(fmt.Sprintf("Command not allowed: %s", baseCommand))
			return result, fmt.Errorf("command not allowed: %s", baseCommand)
		}

		result.AddStep(fmt.Sprintf("Command allowed: %s", baseCommand))
		result.AddStep(fmt.Sprintf("Executing: %s %s", baseCommand, strings.Join(cmdArgs, " ")))

		// Create command
		cmd := exec.CommandContext(ctx, baseCommand, cmdArgs...)

		// Run command and capture output
		outputBytes, err := cmd.CombinedOutput()
		if err != nil {
			result.AddStep(fmt.Sprintf("Command execution failed: %v", err))
			return result, fmt.Errorf("command failed: %w\nOutput: %s", err, string(outputBytes))
		}
		output = string(outputBytes)
	}

	outputLines := strings.Count(output, "\n") + 1
	outputBytes := len(output)
	result.AddStep(fmt.Sprintf("Command completed successfully with %d lines (%d bytes) of output", outputLines, outputBytes))

	result.Output = output
	return result, nil
}

// executeWithShell runs a command using the shell to handle pipes and redirects
func (t *Tool) executeWithShell(ctx context.Context, commandLine string) (string, error) {
	// Verify that all base commands in the pipeline are allowed
	commands := strings.Split(commandLine, "|")
	for _, cmd := range commands {
		trimmed := strings.TrimSpace(cmd)
		baseParts := strings.Fields(trimmed)
		if len(baseParts) == 0 {
			continue
		}

		baseCmd := baseParts[0]
		if !t.isCommandAllowed(baseCmd) {
			return "", fmt.Errorf("command not allowed in pipeline: %s", baseCmd)
		}
	}

	// Use bash to execute the command, which handles pipes correctly
	cmd := exec.CommandContext(ctx, "bash", "-c", commandLine)

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
