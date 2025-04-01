package input

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ReadMultiLine reads multiple lines of input until a terminator is encountered
func ReadMultiLine(terminator string) (string, error) {
	var lines []string
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Printf("Enter your message (terminate with %s):\n", terminator)

	for scanner.Scan() {
		line := scanner.Text()
		if line == terminator {
			break
		}
		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading input: %w", err)
	}

	return strings.Join(lines, "\n"), nil
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
