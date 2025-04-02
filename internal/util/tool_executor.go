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

	// Show tool being executed with spinner
	spinner := NewSpinner(fmt.Sprintf("[Tool: %s] executing...", toolName))
	spinner.Start()

	startTime := time.Now()
	toolExecutionResult, err := tool.Execute(ctx, args)
	elapsedTime := time.Since(startTime)

	// Stop spinner
	spinner.Stop()

	// Show that the tool ran, regardless of success/failure
	toolColor.Printf("🔧 [Tool: %s:%s] executed in %.3fs\n", toolName, toolExecutionResult.ToolMethod, elapsedTime.Seconds())

	return toolExecutionResult.Output, err
}
