package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/saurabh0719/kiwi/internal/tools/core"
	"github.com/saurabh0719/kiwi/internal/util"
)

// ExecuteToolWithFeedback executes a tool with visual feedback
func ExecuteToolWithFeedback(ctx context.Context, tool core.Tool, args map[string]interface{}) (string, error) {
	toolName := tool.Name()
	const maxRetries = 3
	var lastErr error
	var toolExecutionResult core.ToolExecutionResult

	// Make sure we're at the beginning of a line and any previous output is cleared
	// Do this before execution to avoid clearing retry messages
	fmt.Print("\r\033[K")

	// Start the execution timer
	startTime := time.Now()

	// Try executing the tool up to maxRetries times
	for attempt := 1; attempt <= maxRetries; attempt++ {
		var err error
		toolExecutionResult, err = tool.Execute(ctx, args)

		if err == nil {
			break
		}

		lastErr = err
		if attempt < maxRetries {
			util.StepColor.Printf("  → Attempt %d failed: %s. Retrying...\n", attempt, err.Error())
			// Short delay before retry (could be exponential backoff if needed)
			time.Sleep(500 * time.Millisecond)
		}
	}

	// Calculate elapsed time after all attempts
	elapsedTime := time.Since(startTime)

	// Clear spinner and any previous output
	// Move cursor to beginning of current line and clear
	fmt.Print("\r\033[K")
	// Move cursor up one line (in case of spinner) and clear that line too
	fmt.Print("\033[1A\033[K")
	// In case there are multiple lines of spinner output, clear one more line
	fmt.Print("\033[1A\033[K")
	// Return to original position
	fmt.Print("\033[1B")

	// Show that the tool ran, regardless of success/failure
	util.ToolColor.Printf("🔧 [Tool: %s:%s] executed in %.3fs\n", toolName, toolExecutionResult.ToolMethod, elapsedTime.Seconds())

	// Print execution steps
	if len(toolExecutionResult.ToolExecutionSteps) > 0 {
		for _, step := range toolExecutionResult.ToolExecutionSteps {
			util.StepColor.Printf("  → %s\n", strings.TrimSpace(step))
		}
	}

	// Return the error from the last attempt if all retries failed
	if lastErr != nil {
		util.ErrorColor.Printf("  → All %d attempts failed. Last error: %s\n", maxRetries, lastErr.Error())
		return "", lastErr
	}
	fmt.Println()
	return toolExecutionResult.Output, nil
}

// ExecuteTool executes a tool with no visual feedback
func ExecuteTool(ctx context.Context, tool core.Tool, args map[string]interface{}) (string, error) {
	toolExecutionResult, err := tool.Execute(ctx, args)
	if err != nil {
		return "", err
	}

	return toolExecutionResult.Output, nil
}
