package input

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// ReadMultiLine reads multiple lines of input until a terminator is encountered
func ReadMultiLine(terminator string) (string, error) {
	var lines []string
	scanner := bufio.NewScanner(os.Stdin)

	// If terminator is empty, we're using Enter/Return as the delimiter
	if terminator == "" {
		// Use the advanced reader that supports Shift+Enter for multiline
		return ReadMultiLineWithShiftEnter()
	} else {
		fmt.Printf("Enter your message (terminate with %s):\n", terminator)

		for scanner.Scan() {
			line := scanner.Text()
			if line == terminator {
				break
			}
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading input: %w", err)
	}

	// Make sure we always return at least an empty string, never nil
	if len(lines) == 0 {
		return "", nil
	}

	return strings.Join(lines, "\n"), nil
}

// ReadMultiLineWithShiftEnter reads user input supporting both single-line and multi-line modes:
// - Pressing Enter submits the input as a single line
// - Using Shift+Enter allows for line breaks within the input for multi-line content
// It returns the combined text with proper newlines as a single string.
func ReadMultiLineWithShiftEnter() (string, error) {
	var builder strings.Builder
	reader := bufio.NewReader(os.Stdin)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// At EOF, use what we have
				return builder.String(), nil
			}
			return "", fmt.Errorf("error reading input: %w", err)
		}

		// Trim the trailing newline
		line = strings.TrimSuffix(line, "\n")
		line = strings.TrimSuffix(line, "\r") // For Windows compatibility

		// Check if this is an empty line (just Enter was pressed)
		if line == "" {
			// If builder is empty, this is the first line and user just pressed Enter
			// In this case, return an empty string
			if builder.Len() == 0 {
				return "", nil
			}

			// Otherwise, this is the end of multi-line input
			break
		}

		// Append the line to our builder
		builder.WriteString(line)

		// When in multi-line mode (user pressed Shift+Enter),
		// we'll receive lines normally and should add a proper newline
		builder.WriteString("\n")
	}

	// Trim any trailing newline for clean output
	result := strings.TrimSuffix(builder.String(), "\n")
	return result, nil
}

// ReadPrompt reads a single line of input with a prompt
func ReadPrompt(prompt string) (string, error) {
	fmt.Print(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return "", fmt.Errorf("error reading input")
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading input: %w", err)
	}
	return scanner.Text(), nil
}

// IsInputPiped checks if input is being piped from stdin rather than coming from an interactive terminal
// Returns true if input is piped, and the piped content as a string
func IsInputPiped() (bool, string) {
	// Check if stdin is a pipe
	stat, _ := os.Stdin.Stat()
	isPipe := (stat.Mode() & os.ModeCharDevice) == 0

	// If it is a pipe, read all input
	if isPipe {
		scanner := bufio.NewScanner(os.Stdin)
		var input strings.Builder

		for scanner.Scan() {
			input.WriteString(scanner.Text())
			input.WriteString("\n")
		}

		// Return the input without the final newline
		inputStr := input.String()
		return true, strings.TrimSpace(inputStr)
	}

	// Not a pipe, return empty string
	return false, ""
}
