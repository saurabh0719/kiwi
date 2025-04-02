package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/saurabh0719/kiwi/internal/tools/core"
)

var (
	// ToolColor is used for tool execution messages
	ToolColor = color.New(color.FgYellow)
)

// ExecuteToolWithFeedback executes a tool with visual feedback
func ExecuteToolWithFeedback(ctx context.Context, tool core.Tool, args map[string]interface{}) (string, error) {
	toolName := tool.Name()

	// Execute the tool
	startTime := time.Now()
	toolExecutionResult, err := tool.Execute(ctx, args)
	elapsedTime := time.Since(startTime)

	// Make sure we're at the beginning of a line and any previous output is cleared
	fmt.Print("\r\033[K")

	// Show that the tool ran, regardless of success/failure
	ToolColor.Printf("🔧 [Tool: %s:%s] executed in %.3fs\n", toolName, toolExecutionResult.ToolMethod, elapsedTime.Seconds())

	return toolExecutionResult.Output, err
}
