package util

import "github.com/fatih/color"

var (
	UserColor = color.New(color.FgYellow, color.Bold)

	AssistantColor = color.RGB(52, 235, 155).Add(color.Bold)

	StatsColor = color.New(color.FgCyan)

	DividerColor = color.New(color.FgHiBlack)

	OutputColor = color.New(color.FgGreen)

	ErrorColor = color.New(color.FgRed)

	SuccessColor = color.New(color.FgGreen, color.Bold)

	InfoColor = color.New(color.FgHiBlue)

	// WarningColor is used for warnings and non-critical errors
	WarningColor = color.New(color.FgYellow)

	PromptColor = color.New(color.FgMagenta)

	// ToolColor is used for tool execution messages
	ToolColor = color.New(color.FgYellow)

	// StepColor is used for tool execution steps (faded/subtle)
	StepColor = color.New(color.FgHiBlack)

	// HeaderColor is used for section headers
	HeaderColor = color.New(color.FgHiMagenta, color.Bold)

	// SessionIDColor is used for session IDs
	SessionIDColor = color.New(color.FgCyan, color.Bold)

	// SummaryColor is used for session summary labels
	SummaryColor = color.New(color.FgGreen)

	// CommandColor is used for command bullets
	CommandColor = color.New(color.FgHiYellow)

	// HighlightColor is used for highlighting command examples
	HighlightColor = color.New(color.FgHiWhite)
)
