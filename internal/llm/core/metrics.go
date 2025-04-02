package core

import (
	"github.com/fatih/color"
)

var (
	// StatsColor is used for printing metrics
	StatsColor = color.New(color.FgCyan)
)

// PrintResponseMetrics formats and prints the response metrics in a consistent way
// This includes token usage and detailed timing metrics
func PrintResponseMetrics(metrics *ResponseMetrics, modelName string) {
	// Calculate timing information
	totalTime := metrics.ResponseTime.Seconds()
	llmTime := metrics.LLMTime.Seconds()
	toolTime := metrics.ToolTime.Seconds()
	otherTime := totalTime - llmTime - toolTime

	// Print detailed timing information
	StatsColor.Printf("\n[%s] Tokens: %d prompt + %d completion = %d total | Time: %.2fs (LLM: %.2fs, Tools: %.2fs, Other: %.2fs)\n",
		modelName,
		metrics.PromptTokens,
		metrics.CompletionTokens,
		metrics.TotalTokens,
		totalTime,
		llmTime,
		toolTime,
		otherTime)
}
