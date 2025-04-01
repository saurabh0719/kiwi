package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileSystemTool(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "kiwi-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := "Hello, World!"
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create FileSystemTool instance
	fsTool := NewFileSystemTool()

	// Test reading a file
	args := map[string]interface{}{
		"operation": "read",
		"path":      testFile,
	}
	result, err := fsTool.Execute(context.Background(), args)
	if err != nil {
		t.Errorf("Execute(read) failed: %v", err)
	}
	if result.Output != testContent {
		t.Errorf("Execute(read) content mismatch: got %q, want %q", result.Output, testContent)
	}
	// Verify the toolMethod is correctly set
	if result.ToolMethod != "read" {
		t.Errorf("Execute(read) toolMethod mismatch: got %q, want %q", result.ToolMethod, "read")
	}

	// Test listing directory
	args = map[string]interface{}{
		"operation": "list",
		"path":      tmpDir,
	}
	result, err = fsTool.Execute(context.Background(), args)
	if err != nil {
		t.Errorf("Execute(list) failed: %v", err)
	}

	// Check if the listed files contain our test file
	if !strings.Contains(result.Output, "test.txt") {
		t.Errorf("Execute(list) should contain test.txt, got: %v", result.Output)
	}
	// Verify the toolMethod is correctly set
	if result.ToolMethod != "list" {
		t.Errorf("Execute(list) toolMethod mismatch: got %q, want %q", result.ToolMethod, "list")
	}
}

func TestShellTool(t *testing.T) {
	shellTool := NewShellTool()

	// Test simple echo command
	args := map[string]interface{}{
		"command": "echo test",
	}
	result, err := shellTool.Execute(context.Background(), args)
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}
	if !strings.Contains(result.Output, "test") {
		t.Errorf("Execute output should contain 'test', got %q", result.Output)
	}
	// Verify the toolMethod is correctly set
	if result.ToolMethod != "echo" {
		t.Errorf("Execute toolMethod mismatch: got %q, want %q", result.ToolMethod, "echo")
	}
}

func TestSystemInfoTool(t *testing.T) {
	sysInfo := NewSystemInfoTool()

	// Test basic info
	args := map[string]interface{}{
		"type": "basic",
	}
	result, err := sysInfo.Execute(context.Background(), args)
	if err != nil {
		t.Errorf("Execute(basic) failed: %v", err)
	}

	// Check if basic information is included
	if !strings.Contains(result.Output, "os:") || !strings.Contains(result.Output, "arch:") {
		t.Error("Execute(basic) missing required fields in output:", result.Output)
	}
	// Verify the toolMethod is correctly set
	if result.ToolMethod != "basic" {
		t.Errorf("Execute toolMethod mismatch: got %q, want %q", result.ToolMethod, "basic")
	}

	// Test memory info
	args = map[string]interface{}{
		"type": "memory",
	}
	result, err = sysInfo.Execute(context.Background(), args)
	if err != nil {
		t.Errorf("Execute(memory) failed: %v", err)
	}

	// Check if memory information is included
	if !strings.Contains(result.Output, "alloc:") || !strings.Contains(result.Output, "sys:") {
		t.Error("Execute(memory) missing required fields in output:", result.Output)
	}
	// Verify the toolMethod is correctly set
	if result.ToolMethod != "memory" {
		t.Errorf("Execute toolMethod mismatch: got %q, want %q", result.ToolMethod, "memory")
	}
}

func TestRegistry(t *testing.T) {
	registry := NewRegistry()

	// Test registering tools
	fsTool := NewFileSystemTool()
	shellTool := NewShellTool()
	sysInfo := NewSystemInfoTool()

	registry.Register(fsTool)
	registry.Register(shellTool)
	registry.Register(sysInfo)

	// Test getting tools
	if tool, ok := registry.Get("filesystem"); !ok || tool == nil {
		t.Error("Failed to get filesystem tool")
	}
	if tool, ok := registry.Get("shell"); !ok || tool == nil {
		t.Error("Failed to get shell tool")
	}
	if tool, ok := registry.Get("sysinfo"); !ok || tool == nil {
		t.Error("Failed to get sysinfo tool")
	}

	// Test getting non-existent tool
	if tool, ok := registry.Get("nonexistent"); ok || tool != nil {
		t.Error("Got non-existent tool")
	}

	// Test tools description
	desc := registry.GetToolsDescription()
	if desc == "" {
		t.Error("GetToolsDescription returned empty string")
	}
}
