# Kiwi

![Image](https://github.com/user-attachments/assets/9e323659-f603-4b8f-8a24-da5f19169c38)

A command-line interface (CLI) for interacting with Large Language Models (LLMs) directly from your terminal.

[![GitHub license](https://img.shields.io/github/license/saurabh0719/kiwi)](https://github.com/saurabh0719/kiwi/blob/main/LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.18+-00ADD8?style=flat&logo=go)](https://github.com/saurabh0719/kiwi)

```bash
# Quick installation
git clone https://github.com/saurabh0719/kiwi.git
cd kiwi
go build -o kiwi ./cmd/kiwi
# Move to a directory in PATH
sudo mv kiwi /usr/local/bin/
```

## Features

- **Multiple LLM Providers**: Support for OpenAI and Claude APIs
- **Interactive Chat**: Maintain context in ongoing conversations
- **Shell Command Assistance**: Get terminal command suggestions with confirmation
- **Execute Mode**: Run one-off prompts for quick answers
- **Debug Mode**: View token usage and response time statistics
- **Built-in Tools**: Filesystem operations, shell commands, system information

## Table of Contents

* [Installation](#installation)
* [Usage](#usage)
  * [API Keys](#api-keys)
  * [Execute Mode](#execute-mode)
  * [Shell Command Assistance](#shell-command-assistance)
  * [Interactive Chat](#interactive-chat)
  * [Debug Mode](#debug-mode)
* [Configuration](#configuration)
* [Built-in Tools](#built-in-tools)
* [Custom Tools](#custom-tools)
* [Project Structure](#project-structure)
* [Contributing](#contributing)
* [License](#license)

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

### Adding to System Path

#### Linux/macOS

```bash
# Move to a directory already in PATH
sudo mv kiwi /usr/local/bin/

# Add to PATH (add to your shell config)
echo 'export PATH=$PATH:/path/to/kiwi/directory' >> ~/.bashrc  # or ~/.zshrc
source ~/.bashrc  # or ~/.zshrc
```

#### Windows

```powershell
# Move to a directory in PATH (PowerShell with admin rights)
Move-Item -Path .\kiwi.exe -Destination "C:\Windows\System32\"

# Add to PATH (requires admin privileges)
[Environment]::SetEnvironmentVariable("Path", $env:Path + ";C:\path\to\kiwi", "User")
```

## Usage

### API Keys

Set your API keys using the config command:

```bash
kiwi config set llm.provider openai  # or claude
kiwi config set llm.model gpt-4o     # or claude-3-opus-20240229
kiwi config set llm.api_key your-api-key-here
```

### Execute Mode

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

> **Note**: For complex commands with pipelines, use the execute mode (`kiwi e`) which provides better handling.

### Interactive Chat

Start an interactive chat session that maintains context:

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

Kiwi: You can use filesystem tools in Kiwi to perform operations like reading files, 
listing directories, and more. These tools are available within chat sessions.

For example, to list files in a directory:
- Use `list_files: <directory_path>` to see files in a specific directory
- Use `read_file: <file_path>` to view the contents of a file

You: read_file: ~/.bashrc

Kiwi: [Shows content of your bashrc file]
```

### Debug Mode

Enable debug mode to see token usage and response time:

```bash
# Using command-line flag
kiwi e "What is a smartphone" --debug

# Or set it permanently in your config
kiwi config set ui.debug true
```

Output with debug mode:

```
A smartphone is a mobile device that combines cellular and mobile computing functions into one unit.
Key features include:

- Touchscreen interface
- Mobile operating system (Android or iOS)
- Connectivity options (cellular, Wi-Fi, Bluetooth)
- App ecosystem
- Camera functionality
- Various sensors

[gpt-4o] Tokens: 501 prompt + 162 completion = 663 total | Time: 3.92s
```

## Configuration

Kiwi provides a simple configuration system:

```bash
# List all settings
kiwi config list

# Get/set specific settings
kiwi config get llm.provider
kiwi config set llm.provider claude
kiwi config set llm.model claude-3-opus-20240229
kiwi config set llm.api_key your_api_key
kiwi config set ui.debug true
```

Config is stored in `~/.kiwi/config.yaml`:

```yaml
llm:
  provider: openai
  model: gpt-4
  api_key: your_api_key
  safe_mode: true
ui:
  debug: false
```

## Built-in Tools

Kiwi comes with several built-in tools accessible in chat sessions:

### Filesystem Tools
```
- list_files: <directory>       # List files in a directory
- read_file: <file_path>        # Read the contents of a file
- write_file: <file_path>       # Write content to a file
- delete_file: <file_path>      # Delete a file
```

### Shell Tools
```
- run_command: <command>        # Execute a shell command
- command_help: <command>       # Get help for a command
```

### System Info Tools
```
- system_info                   # Get OS information
- disk_usage                    # Check disk space
- memory_info                   # Get memory statistics
```

## Custom Tools

Kiwi can be extended with custom tools by implementing the core Tool interface:

1. **Create a new tool package**:

```go
// mytool.go
package mytool

import (
	"context"
	"fmt"
	"github.com/saurabh0719/kiwi/internal/tools/core"
)

type Tool struct {
	name        string
	description string
	parameters  map[string]core.Parameter
}

func New() *Tool {
	return &Tool{
		name:        "mytool",
		description: "A helpful description of what this tool does",
		parameters: map[string]core.Parameter{
			"param1": {
				Type:        "string",
				Description: "Description of parameter",
				Required:    true,
			},
		},
	}
}

func (t *Tool) Name() string { return t.name }
func (t *Tool) Description() string { return t.description }
func (t *Tool) Parameters() map[string]core.Parameter { return t.parameters }

func (t *Tool) Execute(ctx context.Context, args map[string]interface{}) (core.ToolExecutionResult, error) {
	param1, ok := args["param1"].(string)
	if !ok {
		return core.ToolExecutionResult{}, fmt.Errorf("param1 must be a string")
	}
	
	// Tool logic here
	output := fmt.Sprintf("Processed: %s", param1)
	
	return core.ToolExecutionResult{
		ToolMethod: "process",
		Output:     output,
	}, nil
}
```

2. **Register your tool**:

```go
// In tools.go
import "github.com/saurabh0719/kiwi/internal/tools/mytool"

func NewMyTool() core.Tool {
	return mytool.New()
}

func RegisterStandardTools(registry *Registry) {
	registry.Register(NewFileSystemTool())
	registry.Register(NewShellTool())
	registry.Register(NewSystemInfoTool())
	registry.Register(NewMyTool()) // Your tool
}
```

We welcome contributions to expand Kiwi's capabilities with new tools! Some ideas for tools:
- Weather information
- Web search
- Calendar integration
- Note-taking
- Translation

## Project Structure

```
kiwi/
├── cmd/kiwi/         # Main entry point
├── internal/
│   ├── cli/          # CLI commands implementation
│   ├── config/       # Configuration management
│   ├── input/        # User input handling
│   ├── llm/          # LLM provider interfaces
│   ├── session/      # Chat session management
│   ├── tools/        # Built-in tools
│   └── util/         # Utility functions
```

## Contributing

We welcome contributions from developers of all skill levels!

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

Ways to contribute:
- Add new tools to expand functionality
- Improve documentation
- Fix bugs or improve existing features
- Add support for new LLM providers
- Enhance UI/UX in the terminal

## License

This project is licensed under the Apache-2.0 License - see the [LICENSE](LICENSE) file for details.