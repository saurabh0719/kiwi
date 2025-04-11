package shell

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"

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

	// Process the command
	var output string
	var err error

	// Always use shell for command execution to properly handle
	// command sequences (&&, ||, ;) and special characters

	// Just execute the command - clean output stream
	output, err = t.executeWithShell(ctx, commandLine)
	if err != nil {
		if strings.Contains(err.Error(), "command interrupted") {
			// Just return without retry if interrupted by signal
			result.Output = output
			return result, nil
		}
		// Only add error step after command is fully completed
		result.AddStep(fmt.Sprintf("Command execution failed: %v", err))
		return result, err
	}

	// Only add completion step after command is fully completed
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

	// Set up pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Create a signal channel to handle interrupts
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	// Create a channel to signal when the command is done
	doneChan := make(chan struct{})

	// Create a context that can be canceled to stop goroutines
	cancelCtx, cancelFunc := context.WithCancel(ctx)
	defer cancelFunc()

	// Start the command
	if err := cmd.Start(); err != nil {
		signal.Stop(signalChan)
		close(signalChan)
		return "", fmt.Errorf("failed to start command: %w", err)
	}

	// Signal handling flag to track if we've been interrupted
	interrupted := false

	// Handle signals in a goroutine
	go func() {
		select {
		case <-cancelCtx.Done():
			// Context was canceled, exit goroutine
			return
		case sig := <-signalChan:
			interrupted = true
			// Subtle message without disrupting clean output too much
			fmt.Printf("\n")

			// Send the signal to the process group to ensure all child processes are terminated
			if cmd.Process != nil {
				// First try to kill just the process
				err := cmd.Process.Signal(sig)
				if err != nil {
					// If that fails, try a harder kill
					cmd.Process.Kill()
				}
			}

			cancelFunc() // Cancel all goroutines reading from pipes
		}
	}()

	// Create a buffer to store the complete output
	var combinedOutput strings.Builder
	var outputMutex sync.Mutex
	var wg sync.WaitGroup

	// Stream stdout in real-time
	wg.Add(1)
	go func() {
		defer wg.Done()
		buffer := make([]byte, 1024)
		for {
			select {
			case <-cancelCtx.Done():
				return
			default:
				n, err := stdout.Read(buffer)
				if n > 0 {
					output := string(buffer[:n])
					// Thread-safe write to combinedOutput
					outputMutex.Lock()
					combinedOutput.WriteString(output)
					outputMutex.Unlock()
					// Print to stdout for real-time streaming
					fmt.Print(output)
				}
				if err != nil {
					return
				}
			}
		}
	}()

	// Stream stderr in real-time
	wg.Add(1)
	go func() {
		defer wg.Done()
		buffer := make([]byte, 1024)
		for {
			select {
			case <-cancelCtx.Done():
				return
			default:
				n, err := stderr.Read(buffer)
				if n > 0 {
					output := string(buffer[:n])
					// Thread-safe write to combinedOutput
					outputMutex.Lock()
					combinedOutput.WriteString(output)
					outputMutex.Unlock()
					// Print to stderr for real-time streaming
					fmt.Fprint(os.Stderr, output)
				}
				if err != nil {
					return
				}
			}
		}
	}()

	// Wait for the command to complete in a goroutine
	go func() {
		cmd.Wait()
		close(doneChan)
		cancelFunc() // Cancel context to stop the reading goroutines
	}()

	// Wait for either completion or cancellation
	select {
	case <-doneChan:
		// Command completed normally
	case <-cancelCtx.Done():
		// Context was canceled, or we received signal
		if interrupted {
			// If we were interrupted by signal, return a specific error
			// so Execute() knows not to retry
			return combinedOutput.String(), fmt.Errorf("command interrupted by signal")
		} else if ctx.Err() != context.Canceled {
			// If it wasn't our own cancellation, but parent context
			return combinedOutput.String(), fmt.Errorf("command interrupted: %w", ctx.Err())
		}
	}

	// Wait for all goroutines to finish reading pipes
	wg.Wait()

	// Stop the signal handler
	signal.Stop(signalChan)
	close(signalChan)

	// Check if there was an error running the command
	if err := cmd.Wait(); err != nil && ctx.Err() == nil && !interrupted {
		return combinedOutput.String(), fmt.Errorf("command failed: %w", err)
	}

	// Add a newline at the end of output if it doesn't end with one
	// This helps ensure proper formatting when returning to the assistant interface
	outputStr := combinedOutput.String()
	if len(outputStr) > 0 && !strings.HasSuffix(outputStr, "\n") {
		outputStr += "\n"
	}

	return outputStr, nil
}
