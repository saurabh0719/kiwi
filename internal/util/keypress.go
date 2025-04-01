package util

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

// ReadSingleKey reads a single keypress without requiring Enter.
// It returns the key as a byte.
func ReadSingleKey() (byte, error) {
	// Save the current state of the terminal
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return 0, err
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	// Read a single byte
	var b [1]byte
	_, err = os.Stdin.Read(b[:])
	if err != nil {
		return 0, err
	}

	return b[0], nil
}

// PromptForConfirmation asks the user for confirmation with a single keypress.
// Returns true if the user confirms (by pressing 'y' or 'Y'), false otherwise.
func PromptForConfirmation(prompt string) (bool, error) {
	fmt.Print(prompt)

	// Read a single keystroke
	key, err := ReadSingleKey()
	if err != nil {
		return false, err
	}

	// Print the pressed key and a newline for better UX
	fmt.Printf("%c\n", key)

	// Check if the key is 'y' or 'Y'
	return key == 'y' || key == 'Y', nil
}
