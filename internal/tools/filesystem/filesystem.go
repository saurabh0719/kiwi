package filesystem

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/saurabh0719/kiwi/internal/tools/core"
)

// Tool provides file system operations
type Tool struct {
	name        string
	description string
	parameters  map[string]core.Parameter
}

// New creates a new FileSystemTool
func New() *Tool {
	parameters := map[string]core.Parameter{
		"operation": {
			Type:        "string",
			Description: "Operation to perform (list or read)",
			Required:    true,
		},
		"path": {
			Type:        "string",
			Description: "Path to file or directory",
			Required:    true,
		},
	}

	return &Tool{
		name:        "filesystem",
		description: "Provides file system operations like listing files and reading file contents",
		parameters:  parameters,
	}
}

// Name returns the name of the tool
func (t *Tool) Name() string {
	return t.name
}

// Description returns the description of the tool
func (t *Tool) Description() string {
	return t.description
}

// Parameters returns the parameters for the tool
func (t *Tool) Parameters() map[string]core.Parameter {
	return t.parameters
}

// Execute runs the filesystem operation
func (t *Tool) Execute(ctx context.Context, args map[string]interface{}) (core.ToolExecutionResult, error) {
	result := core.ToolExecutionResult{
		ToolMethod: "",
		Output:     "",
	}

	// Check if operation is a valid string
	operationVal, ok := args["operation"]
	if !ok || operationVal == nil {
		return result, fmt.Errorf("operation parameter is required")
	}

	operation, ok := operationVal.(string)
	if !ok {
		return result, fmt.Errorf("operation must be a string")
	}

	result.ToolMethod = operation

	// Check if path is a valid string
	pathVal, ok := args["path"]
	if !ok || pathVal == nil {
		return result, fmt.Errorf("path parameter is required")
	}

	path, ok := pathVal.(string)
	if !ok {
		return result, fmt.Errorf("path must be a string")
	}

	// Validate the path for safety
	if !isPathSafe(path) {
		return result, fmt.Errorf("path is not safe: %s", path)
	}

	var err error
	var output string

	// Execute the requested operation
	switch operation {
	case "list":
		output, err = t.listFiles(path)
	case "read":
		output, err = t.readFile(path)
	case "write":
		content, _ := args["content"].(string)
		output, err = t.writeFile(path, content)
	case "delete":
		output, err = t.deleteFile(path)
	default:
		err = fmt.Errorf("unknown operation: %s", operation)
	}

	if err != nil {
		return result, err
	}

	result.Output = output
	return result, nil
}

// listFiles lists the files in a directory
func (t *Tool) listFiles(path string) (string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return "", fmt.Errorf("failed to read directory: %w", err)
	}

	var result strings.Builder
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			name += "/"
		}
		result.WriteString(name + "\n")
	}

	return result.String(), nil
}

// readFile reads the content of a file
func (t *Tool) readFile(path string) (string, error) {
	// Validate the path for safety
	if !isPathSafe(path) {
		return "", fmt.Errorf("path is not safe: %s", path)
	}

	// Check if the file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", fmt.Errorf("file does not exist: %s", path)
	}

	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return string(data), nil
}

// writeFile writes content to a file
func (t *Tool) writeFile(path string, content string) (string, error) {
	// Validate the path for safety
	if !isPathSafe(path) {
		return "", fmt.Errorf("path is not safe: %s", path)
	}

	// Write the file
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return fmt.Sprintf("Successfully wrote %d bytes to %s", len(content), path), nil
}

// deleteFile deletes a file
func (t *Tool) deleteFile(path string) (string, error) {
	// Validate the path for safety
	if !isPathSafe(path) {
		return "", fmt.Errorf("path is not safe: %s", path)
	}

	// Check if the file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", fmt.Errorf("file does not exist: %s", path)
	}

	// Delete the file
	err := os.Remove(path)
	if err != nil {
		return "", fmt.Errorf("failed to delete file: %w", err)
	}

	return fmt.Sprintf("Successfully deleted %s", path), nil
}

// isPathSafe checks if a path is safe to access
func isPathSafe(path string) bool {
	// Get absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	// Allow temporary paths for testing
	if strings.HasPrefix(absPath, os.TempDir()) {
		return true
	}

	// Get working directory
	wd, err := os.Getwd()
	if err != nil {
		return false
	}

	// Ensure path is under working directory
	rel, err := filepath.Rel(wd, absPath)
	if err != nil {
		return false
	}

	// Check if path tries to escape working directory
	return !filepath.IsAbs(rel) && !strings.HasPrefix(rel, "..")
}
