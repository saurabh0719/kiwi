# Kiwi

A command-line interface for interacting with Large Language Models (LLMs) in terminal environments.

## Overview

Kiwi provides a seamless interface to interact with LLMs like OpenAI and Claude directly from your terminal. It offers multiple interaction modes, built-in tools, and a configurable environment to enhance your productivity.

## Features

- **Multiple LLM Providers**: Support for OpenAI and Claude APIs
- **Interactive Chat**: Maintain context in ongoing conversations
- **Shell Command Assistant**: Get command suggestions for terminal tasks
- **Execute Mode**: Run one-off prompts for quick answers
- **Debug Mode**: View token usage and response time statistics
- **Built-in Tools**:
  - Filesystem operations (read, write, list files)
  - Shell command execution with safety controls
  - System information retrieval

## Installation

### Prerequisites

- Go 1.18 or later
- Git

### Build from Source

```bash
# Clone the repository
git clone https://github.com/saurabh0719/kiwi.git
cd kiwi

# Basic build
go build -o kiwi ./cmd/kiwi

# Optimized build (smaller binary size)
go build -ldflags="-s -w" -o kiwi ./cmd/kiwi

# Using make
make build
```

### Cross-compilation

```bash
# For Windows
GOOS=windows GOARCH=amd64 go build -o kiwi.exe ./cmd/kiwi

# For macOS
GOOS=darwin GOARCH=amd64 go build -o kiwi ./cmd/kiwi

# For Linux
GOOS=linux GOARCH=amd64 go build -o kiwi ./cmd/kiwi
```

### Adding to System Path

#### Linux/Ubuntu

```bash
# Move to a directory already in PATH
sudo mv kiwi /usr/local/bin/

# Add to PATH permanently (add to your shell config)
echo 'export PATH=$PATH:/path/to/kiwi/directory' >> ~/.bashrc
source ~/.bashrc
```

#### macOS

```bash
# Move to a directory already in PATH
sudo mv kiwi /usr/local/bin/

# Using Homebrew's directory
mv kiwi $(brew --prefix)/bin/

# Add to PATH permanently
echo 'export PATH=$PATH:/path/to/kiwi/directory' >> ~/.zshrc
source ~/.zshrc
```

#### Windows

```powershell
# Move to a directory in PATH (PowerShell with admin rights)
Move-Item -Path .\kiwi.exe -Destination "C:\Windows\System32\"

# Add to PATH permanently (requires admin privileges)
[Environment]::SetEnvironmentVariable("Path", $env:Path + ";C:\path\to\kiwi", "User")
```

## Usage

### API Keys

Set your API keys using environment variables:

```bash
kiwi config
kiwi config set llm.provider <provider>
kiwi config set llm.model <model>
kiwi config set llm.api_key <your-api-key>
```

### Execute Prompts

Get quick answers without starting a full chat session:

```bash
# Shorthand command
kiwi e "Explain Docker in simple terms"

# Full command
kiwi execute "Explain Docker in simple terms"
```

### Shell Command Assistance

Get help with shell commands:

```bash
# Shorthand command
kiwi s "find all pdf files modified in the last week"

# Full command
kiwi shell "find all pdf files modified in the last week"
```

The tool will suggest a command and ask for confirmation before executing it:

```
find /path/to/search -name "*.pdf" -type f -mtime -7

Do you want to execute this command? (y/n): 
```

> **Note about Shell Command Limitations**: The shell command tool currently has limitations with certain operations:
> - Pipeline commands (using `|`) are not fully supported in shell mode
> - Only a limited set of base commands are allowed for security reasons
> - For complex commands with pipelines, use the execute mode (`kiwi e`) which can leverage filesystem and shell tools for better handling

### Interactive Chat

Start an interactive chat session:

```bash
# Shorthand command
kiwi c

# Full command 
kiwi chat
```

Example interaction:

```
Created new session: session_1712082042
Chat session started. Type 'exit' to end the session.
Using openai model: gpt-4o
----------------------------------------

You: How can I use the filesystem tools in Kiwi?

Kiwi: You can use filesystem tools in Kiwi to perform operations like reading files, listing directories, and more. These tools are available within chat sessions.

For example, to list files in a directory:
- Use `list_files: <directory_path>` to see files in a specific directory
- Use `read_file: <file_path>` to view the contents of a file

You: read_file: ~/.bashrc

Kiwi: [Shows content of your bashrc file]
```

### Debug Mode

Enable debug mode to see detailed information about token usage and response time:

```bash
# Using command-line flag
kiwi e "What is a smartphone" --debug

# Or set it permanently in your config
kiwi config set ui.debug true
```

Output with debug mode enabled:

```
A smartphone is a mobile device that combines cellular and mobile computing functions into one unit. Key features include:

- Touchscreen interface
- Mobile operating system (Android or iOS)
- Connectivity options (cellular, Wi-Fi, Bluetooth)
- App ecosystem
- Camera functionality
- Various sensors

[gpt-4o] Tokens: 501 prompt + 162 completion = 663 total | Time: 3.92s
```

#### Tool Execution Debugging

In debug mode, Kiwi shows which tools are being executed with collapsible output sections:

```
🔧 [Tool: filesystem:readFile] executed in 0.123s [ID: section-1 - use 'expand section-1' to toggle]
```

Each tool execution reports:
- The tool category being used (filesystem, shell, or sysinfo)
- The specific method that was called
- The execution time
- A section ID that can be used to expand/collapse the output

You can interact with the collapsible sections using the following commands in chat mode:
- `expand section-1` - Expands the output of section-1
- `collapse section-1` - Collapses the output of section-1
- `sections` - Lists all available sections

When expanded, you'll see the full output of the tool:

```
🔧 [Tool: filesystem:readFile] executed in 0.123s [ID: section-1 - use 'expand section-1' to toggle]
  # Contents of README.md
  A command-line interface for interacting with Large Language Models (LLMs)...
```

This feature helps keep the terminal clean while still allowing you to inspect tool outputs when needed.

## Configuration

### Command Line Options

```bash
--provider string   LLM provider (openai, claude) (default "openai")
--model string      Model to use (default "gpt-3.5-turbo")
--api-key string    API key (if not set via environment variable)
--safe-mode         Enable command confirmation (default true)
--debug             Enable debug mode (default false)
```

### Config Commands

Kiwi provides a configuration system that persists settings between runs:

```bash
# List all settings
kiwi config list

# Get a specific setting
kiwi config get llm.provider

# Update settings
kiwi config set llm.provider claude
kiwi config set llm.model claude-3-opus-20240229
kiwi config set llm.api_key your_api_key
kiwi config set llm.safe_mode false
```

### Config File

Kiwi stores configuration in `~/.kiwi/config.yaml`:

```yaml
llm:
  provider: openai
  model: gpt-4
  api_key: your_api_key
  safe_mode: true
ui:
  debug: false
```

## Extending Kiwi: Custom Tools

Kiwi can be extended with custom tools by implementing the core Tool interface.

### Creating a Custom Tool

1. **Create a new package**
```
mkdir -p internal/tools/mytool
```

2. **Implement the Tool interface**
```go
// mytool.go
package mytool

import (
	"context"
	"fmt"

	"github.com/saurabh0719/kiwi/internal/tools/core"
)

// Tool implements the core.Tool interface
type Tool struct {
	name        string
	description string
	parameters  map[string]core.Parameter
}

// New creates a new tool instance
func New() *Tool {
	return &Tool{
		name:        "mytool",
		description: "Description of what this tool does",
		parameters: map[string]core.Parameter{
			"param1": {
				Type:        "string",
				Description: "Parameter description",
				Required:    true,
			},
		},
	}
}

// Name returns the tool name
func (t *Tool) Name() string { return t.name }

// Description returns the tool description
func (t *Tool) Description() string { return t.description }

// Parameters returns the tool parameters
func (t *Tool) Parameters() map[string]core.Parameter { return t.parameters }

// Execute performs the tool's operation
func (t *Tool) Execute(ctx context.Context, args map[string]interface{}) (core.ToolExecutionResult, error) {
	// Extract and validate parameters
	param1, ok := args["param1"].(string)
	if !ok {
		return core.ToolExecutionResult{}, fmt.Errorf("param1 must be a string")
	}
	
	// Implement tool logic
	output := fmt.Sprintf("Processed: %s", param1)
	
	return core.ToolExecutionResult{
		ToolMethod: "process", // Name of specific operation performed
		Output:     output,
	}, nil
}
```

3. **Register your tool in tools.go**
```go
// Add import
import "github.com/saurabh0719/kiwi/internal/tools/mytool"

// Add factory function
func NewMyTool() core.Tool {
	return mytool.New()
}

// Update registration function
func RegisterStandardTools(registry *Registry) {
	registry.Register(NewFileSystemTool())
	registry.Register(NewShellTool())
	registry.Register(NewSystemInfoTool())
	registry.Register(NewMyTool()) // Register your tool
}
```

### Best Practices

- **Validate parameters** thoroughly
- Use **descriptive method names** in ToolMethod field
- Return **clear error messages** for better debugging
- Consider **security implications** for filesystem/shell operations
- Write **unit tests** for your tools

### API-Based Tool Example

For tools that call external APIs (simplified):

```go
// Execute for an API-based tool
func (t *Tool) Execute(ctx context.Context, args map[string]interface{}) (core.ToolExecutionResult, error) {
	// Extract parameters
	query, ok := args["query"].(string)
	if !ok {
		return core.ToolExecutionResult{}, fmt.Errorf("query must be a string")
	}
	
	// Call external API
	resp, err := http.Get(fmt.Sprintf("https://api.example.com/data?q=%s&key=%s", 
		url.QueryEscape(query), t.apiKey))
	if err != nil {
		return core.ToolExecutionResult{}, err
	}
	defer resp.Body.Close()
	
	// Process response
	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return core.ToolExecutionResult{}, err
	}
	
	return core.ToolExecutionResult{
		ToolMethod: "search",
		Output:     fmt.Sprintf("Results: %v", data["results"]),
	}, nil
}
```

## Built-in Tools

Kiwi provides several built-in tools that can be accessed during chat sessions:

### Filesystem Tools

Allow you to interact with the file system:

```
- list_files: List files in a directory
- read_file: Read the contents of a file
- write_file: Write content to a file
- delete_file: Delete a file
```

### Shell Tools

Execute and get help with shell commands:

```
- run_command: Execute a shell command (with safe mode confirmation)
- command_help: Get help with a specific command
```

Allowed commands:
```
ls, cat, grep, find, pwd, head, tail, wc, echo, date, ps, df, du, free, top
```

### System Info Tools

Retrieve system information:

```
- system_info: Get information about the operating system
- disk_usage: Check disk space usage
- memory_info: Get memory usage statistics
```

## Project Structure

```
kiwi/
├── cmd/
│   └── kiwi/         # Main entry point
├── internal/
│   ├── cli/          # CLI commands implementation
│   ├── config/       # Configuration management
│   ├── input/        # User input handling
│   ├── llm/          # LLM provider interfaces
│   │   ├── core/     # Core LLM interfaces
│   │   ├── claude/   # Claude API implementation
│   │   └── openai/   # OpenAI API implementation
│   ├── session/      # Chat session management
│   ├── tools/        # Built-in tools
│   │   ├── core/     # Tool interface definitions
│   │   ├── filesystem/  # Filesystem operations
│   │   ├── shell/    # Shell command operations
│   │   └── sysinfo/  # System information tools
│   └── util/         # Utility functions
├── go.mod            # Go module definition
├── go.sum            # Go dependencies checksum
├── LICENSE           # License information
├── Makefile          # Build automation
└── README.md         # This documentation
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the Apache-2.0 License - see the LICENSE file for details.