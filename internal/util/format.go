package util

import (
	"fmt"
)

// Constants for divider lines
const (
	// ChatDivider is the divider pattern used in chat mode
	ChatDivider = "----------------------------------------"

	// ExecuteDivider is the divider pattern used in execute mode
	ExecuteDivider = "----------------------------------------------------------------"
)

// PrintChatDivider prints a divider line for chat mode
func PrintChatDivider() {
	DividerColor.Println(ChatDivider)
}

// PrintExecuteDivider prints a divider line for execute mode
func PrintExecuteDivider() {
	OutputColor.Println(ExecuteDivider)
}

// PrintExecuteStartDivider prints a divider and newline at the start of execute mode output
func PrintExecuteStartDivider() {
	PrintExecuteDivider()
	fmt.Println()
}

// PrintExecuteEndDivider prints a newline and divider at the end of execute mode output
func PrintExecuteEndDivider() {
	fmt.Println()
	PrintExecuteDivider()
}
