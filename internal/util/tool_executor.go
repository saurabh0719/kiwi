package util

import (
	"context"
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/saurabh0719/kiwi/internal/tools/core"
)

var (
	toolColor = color.New(color.FgYellow)
)

// ExecuteToolWithFeedback executes a tool with visual feedback (spinner and colored output)
func ExecuteToolWithFeedback(ctx context.Context, tool core.Tool, args map[string]interface{}) (string, error) {
	toolName := tool.Name()

	// Get the global spinner manager and start a tool spinner
	spinnerManager := GetGlobalSpinnerManager()

	// Only start a new spinner if we're not already in a tool state
	// This should prevent flickering between tool executions
	if spinnerManager.GetCurrentState() != SpinnerStateTool {
		// Add a slight delay before starting a new spinner to prevent visual glitches
		time.Sleep(100 * time.Millisecond)
		spinnerManager.StartToolSpinner(fmt.Sprintf("[Tool: %s] executing...", toolName))
	}

	startTime := time.Now()
	toolExecutionResult, err := tool.Execute(ctx, args)
	elapsedTime := time.Since(startTime)

	// Only transition to response state on successful tool execution
	// Don't transition here if there was an error - let the caller handle it
	if err == nil {
		// Add a slight delay before transitioning to ensure spinner is visible
		time.Sleep(150 * time.Millisecond)
		spinnerManager.TransitionToResponse()
	}

	// Show that the tool ran, regardless of success/failure
	toolColor.Printf("🔧 [Tool: %s:%s] executed in %.3fs\n", toolName, toolExecutionResult.ToolMethod, elapsedTime.Seconds())

	return toolExecutionResult.Output, err
}
