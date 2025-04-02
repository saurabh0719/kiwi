package util

import (
	"fmt"
)

// ClearSpinner stops all spinners and clears the line
// This ensures that any spinner messages are removed before displaying any output
func ClearSpinner(spinnerManager *SpinnerManager) {
	// Stop all spinners
	spinnerManager.StopAllSpinners()
	// Clear the line to remove any spinner messages
	fmt.Printf("\r%s\r", ClearLine())
}

// PrepareForResponse handles the common pattern of stopping spinners and clearing the line
// before displaying a response
func PrepareForResponse(spinnerManager *SpinnerManager) {
	ClearSpinner(spinnerManager)
}
