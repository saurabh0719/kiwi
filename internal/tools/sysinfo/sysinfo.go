package sysinfo

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/saurabh0719/kiwi/internal/tools/core"
)

// Tool provides system information
type Tool struct {
	name        string
	description string
	parameters  map[string]core.Parameter
}

// New creates a new SystemInfoTool
func New() *Tool {
	parameters := map[string]core.Parameter{
		"type": {
			Type:        "string",
			Description: "Type of information to retrieve (basic, memory, env)",
			Required:    true,
		},
	}

	return &Tool{
		name:        "sysinfo",
		description: "Provides system information and status",
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

// Execute executes the tool with the given arguments
func (t *Tool) Execute(ctx context.Context, args map[string]interface{}) (core.ToolExecutionResult, error) {
	result := core.ToolExecutionResult{
		ToolMethod: "",
		Output:     "",
	}

	infoType, ok := args["type"].(string)
	if !ok {
		return result, fmt.Errorf("type must be a string")
	}

	result.ToolMethod = infoType
	result.AddStep(fmt.Sprintf("Requested information type: %s", infoType))

	var output string
	var err error

	switch infoType {
	case "basic":
		result.AddStep("Gathering basic system information...")
		output, err = t.getBasicInfo()
		if err != nil {
			result.AddStep(fmt.Sprintf("Error getting basic info: %v", err))
		} else {
			result.AddStep("Successfully retrieved basic system information")
		}
	case "memory":
		result.AddStep("Gathering memory usage statistics...")
		output, err = t.getMemoryInfo()
		if err != nil {
			result.AddStep(fmt.Sprintf("Error getting memory info: %v", err))
		} else {
			result.AddStep("Successfully retrieved memory usage information")
		}
	case "env":
		result.AddStep("Gathering environment variables...")
		output, err = t.getEnvironmentInfo()
		if err != nil {
			result.AddStep(fmt.Sprintf("Error getting environment info: %v", err))
		} else {
			varCount := strings.Count(output, "\n")
			result.AddStep(fmt.Sprintf("Successfully retrieved %d environment variables (excluding sensitive ones)", varCount))
		}
	default:
		result.AddStep(fmt.Sprintf("Unknown information type requested: %s", infoType))
		return result, fmt.Errorf("unknown info type: %s", infoType)
	}

	if err != nil {
		return result, err
	}

	result.Output = output
	return result, nil
}

// getBasicInfo returns basic system information
func (t *Tool) getBasicInfo() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	info := map[string]interface{}{
		"hostname":   hostname,
		"os":         runtime.GOOS,
		"arch":       runtime.GOARCH,
		"cpus":       runtime.NumCPU(),
		"time":       time.Now().Format(time.RFC3339),
		"uptime":     time.Since(startTime).String(),
		"go_version": runtime.Version(),
	}

	var result strings.Builder
	for k, v := range info {
		result.WriteString(fmt.Sprintf("%s: %v\n", k, v))
	}

	return result.String(), nil
}

// getMemoryInfo returns memory usage information
func (t *Tool) getMemoryInfo() (string, error) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	info := map[string]string{
		"alloc":       fmt.Sprintf("%v MB", m.Alloc/1024/1024),
		"total_alloc": fmt.Sprintf("%v MB", m.TotalAlloc/1024/1024),
		"sys":         fmt.Sprintf("%v MB", m.Sys/1024/1024),
		"num_gc":      fmt.Sprintf("%d", m.NumGC),
	}

	var result strings.Builder
	for k, v := range info {
		result.WriteString(fmt.Sprintf("%s: %s\n", k, v))
	}

	return result.String(), nil
}

// getEnvironmentInfo returns filtered environment variables
func (t *Tool) getEnvironmentInfo() (string, error) {
	// Filter sensitive environment variables
	var result strings.Builder
	for _, env := range os.Environ() {
		if key, value, ok := strings.Cut(env, "="); ok {
			// Skip sensitive variables
			if !isSensitiveEnvVar(key) {
				result.WriteString(fmt.Sprintf("%s=%s\n", key, value))
			}
		}
	}
	return result.String(), nil
}

// isSensitiveEnvVar checks if an environment variable name is sensitive
func isSensitiveEnvVar(name string) bool {
	sensitiveKeys := []string{
		"KEY", "SECRET", "PASSWORD", "TOKEN", "CREDENTIAL",
		"PRIVATE", "AUTH", "ACCESS", "API_KEY",
	}

	name = strings.ToUpper(name)
	for _, key := range sensitiveKeys {
		if strings.Contains(name, key) {
			return true
		}
	}
	return false
}

// startTime is used to calculate uptime
var startTime = time.Now()
