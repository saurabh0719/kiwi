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
			Description: "Operation to perform: 'list' (list directory contents), 'read' (read existing file), 'write' (write to file, creates it if doesn't exist), 'delete' (delete a file)",
			Required:    true,
		},
		"path": {
			Type:        "string",
			Description: "Path to file or directory. For 'write' operations, this is the file to write to (will be created if it doesn't exist).",
			Required:    true,
		},
		"content": {
			Type:        "string",
			Description: "Content to write (for write operation only). For example, 'content': 'Hello, world!' will write that text to the file.",
			Required:    false,
		},
	}

	return &Tool{
		name:        "filesystem",
		description: "Provides file system operations like listing files, reading from existing files, writing to files (creates files if they don't exist), and deleting files. For writing to files, use operation='write', path='filename.txt', and content='text to write'. Example: To create a file called notes.txt with content 'Meeting notes', use these parameters: {\"operation\": \"write\", \"path\": \"notes.txt\", \"content\": \"Meeting notes\"}.",
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
		ToolMethod:         "",
		ToolExecutionSteps: []string{},
		Output:             "",
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
	result.AddStep(fmt.Sprintf("Requested operation: %s", operation))

	// Check if path is a valid string
	pathVal, ok := args["path"]
	if !ok || pathVal == nil {
		return result, fmt.Errorf("path parameter is required")
	}

	path, ok := pathVal.(string)
	if !ok {
		return result, fmt.Errorf("path must be a string")
	}

	result.AddStep(fmt.Sprintf("Validating path: %s", path))

	// Validate the path for safety
	if !isPathSafe(path) {
		result.AddStep(fmt.Sprintf("Path safety check failed for: %s", path))
		return result, fmt.Errorf("path is not safe: %s", path)
	}

	result.AddStep(fmt.Sprintf("Path safety check passed"))

	// Additional validation based on operation type
	switch operation {
	case "write":
		// For write operations, ensure content parameter is present
		contentVal, hasContent := args["content"]
		if !hasContent || contentVal == nil {
			result.AddStep("Error: Missing required 'content' parameter for write operation")
			return result, fmt.Errorf("content parameter is required for write operation")
		}

		// Validate content is a string
		_, ok := contentVal.(string)
		if !ok {
			result.AddStep("Error: Content must be a string")
			return result, fmt.Errorf("content must be a string")
		}
	case "read":
		// For read operations, check if the file exists first
		if _, err := os.Stat(path); os.IsNotExist(err) {
			result.AddStep(fmt.Sprintf("Error: File '%s' does not exist", path))
			return result, fmt.Errorf("file does not exist: %s. Use 'write' operation first to create it", path)
		}
	}

	var err error
	var output string

	// Execute the requested operation
	switch operation {
	case "list":
		result.AddStep(fmt.Sprintf("Listing files in directory: %s", path))
		output, err = t.listFiles(path)
		if err != nil {
			result.AddStep(fmt.Sprintf("Error listing files: %v", err))
		} else {
			fileCount := strings.Count(output, "\n")
			result.AddStep(fmt.Sprintf("Listed %d files/directories in %s", fileCount, path))
		}
	case "read":
		result.AddStep(fmt.Sprintf("Reading file content: %s", path))
		output, err = t.readFile(path)
		if err != nil {
			result.AddStep(fmt.Sprintf("Error reading file: %v", err))
		} else {
			lineCount := strings.Count(output, "\n") + 1
			result.AddStep(fmt.Sprintf("Successfully read %s (%d lines, %d bytes)", path, lineCount, len(output)))
		}
	case "write":
		content := args["content"].(string)
		contentLength := len(content)
		result.AddStep(fmt.Sprintf("Writing %d bytes to file: %s", contentLength, path))
		output, err = t.writeFile(path, content)
		if err != nil {
			result.AddStep(fmt.Sprintf("Error writing to file: %v", err))
		} else {
			result.AddStep(fmt.Sprintf("Successfully wrote to file %s", path))
		}
	case "delete":
		result.AddStep(fmt.Sprintf("Deleting file: %s", path))
		output, err = t.deleteFile(path)
		if err != nil {
			result.AddStep(fmt.Sprintf("Error deleting file: %v", err))
		} else {
			result.AddStep(fmt.Sprintf("Successfully deleted file %s", path))
		}
	default:
		result.AddStep(fmt.Sprintf("Unknown operation requested: %s", operation))
		err = fmt.Errorf("unknown operation: %s, supported operations are: list, read, write, delete", operation)
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
		return "", fmt.Errorf("file does not exist: %s. Use 'write' operation first to create it", path)
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
