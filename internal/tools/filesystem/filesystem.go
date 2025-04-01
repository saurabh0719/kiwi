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
type Tool struct{}

// New creates a new FileSystemTool
func New() *Tool {
	return &Tool{}
}

// Name returns the name of the tool
func (t *Tool) Name() string {
	return "filesystem"
}

// Description returns the description of the tool
func (t *Tool) Description() string {
	return "Provides file system operations like listing files and reading file contents"
}

// Parameters returns the parameters for the tool
func (t *Tool) Parameters() map[string]core.Parameter {
	return map[string]core.Parameter{
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
}

// Execute executes the tool with the given arguments
func (t *Tool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	operation, ok := args["operation"].(string)
	if !ok {
		return "", fmt.Errorf("operation must be a string")
	}

	path, ok := args["path"].(string)
	if !ok {
		return "", fmt.Errorf("path must be a string")
	}

	// Ensure the path is safe
	cleanPath := filepath.Clean(path)
	if !isPathSafe(cleanPath) {
		return "", fmt.Errorf("path is not safe: %s", path)
	}

	switch operation {
	case "list":
		return t.listFiles(cleanPath)
	case "read":
		return t.readFile(cleanPath)
	default:
		return "", fmt.Errorf("unknown operation: %s", operation)
	}
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

// readFile reads a file and returns its contents
func (t *Tool) readFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Limit the size of the response
	const maxSize = 1024 * 1024 // 1MB
	if len(data) > maxSize {
		return string(data[:maxSize]) + "\n... (file truncated)", nil
	}

	return string(data), nil
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
