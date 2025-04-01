package util

import "github.com/fatih/color"

// Colors for consistent UI across the application
var (
	// UserColor for user messages in chat mode
	UserColor = color.New(color.FgYellow, color.Bold)

	// AssistantColor for assistant responses in chat mode
	AssistantColor = color.RGB(52, 235, 155).Add(color.Bold)

	// StatsColor for statistics and metrics
	StatsColor = color.New(color.FgCyan)

	// DividerColor for visual separators
	DividerColor = color.New(color.FgHiBlack)

	// OutputColor for highlighting output in execute mode
	OutputColor = color.New(color.FgGreen)

	// ErrorColor for error messages
	ErrorColor = color.New(color.FgRed)

	// SuccessColor for success messages
	SuccessColor = color.New(color.FgGreen, color.Bold)

	// InfoColor for informational messages
	InfoColor = color.New(color.FgHiBlue)

	// PromptColor for input prompts
	PromptColor = color.New(color.FgMagenta)
)
