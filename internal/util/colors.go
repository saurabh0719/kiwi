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

	PromptColor = color.New(color.FgMagenta)
)
