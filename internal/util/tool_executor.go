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
	spinner := NewSpinner(fmt.Sprintf("Executing %s tool...", toolName))
	spinner.Start()

	startTime := time.Now()
	result, err := tool.Execute(ctx, args)
	elapsedTime := time.Since(startTime)

	// Stop spinner
	spinner.Stop()

	// Show that the tool ran, regardless of success/failure
	toolColor.Printf("🔧 [Tool: %s] executed in %.2fs\n\n", toolName, elapsedTime.Seconds())

	return result, err
}
