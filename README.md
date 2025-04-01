# Kiwi

![Image](https://github.com/user-attachments/assets/9e323659-f603-4b8f-8a24-da5f19169c38)

A command-line interface (CLI) for interacting with Large Language Models (LLMs) directly from your terminal.

[![GitHub license](https://img.shields.io/github/license/saurabh0719/kiwi)](https://github.com/saurabh0719/kiwi/blob/main/LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.18+-00ADD8?style=flat&logo=go)](https://github.com/saurabh0719/kiwi)

## Installation

### Quick Install (Linux and macOS)

```bash
# Install with a single command
curl -fsSL https://raw.githubusercontent.com/saurabh0719/kiwi/main/install.sh | bash
```

### Download Prebuilt Binaries

You can download prebuilt binaries for your platform from the [GitHub Releases page](https://github.com/saurabh0719/kiwi/releases).

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
- **Default Execute Mode**: Run prompts directly without specifying a command

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
# Direct execution (default behavior)
$ kiwi "What is Docker?"
----------------------------------------------------------------

Docker is an open-source platform that automates the deployment, scaling, and management of applications using containerization. Containers allow you to package an application with all its dependencies into a standardized unit for software development. This ensures that the application runs consistently across different environments.

Key features of Docker include:
- **Portability**: Containers can run on any system that supports Docker, making it easy to move applications between environments.
- **Isolation**: Each container runs in a separate environment, ensuring that applications do not interfere with each other.
- **Efficiency**: Containers share the host OS kernel, making them lightweight and fast to start compared to virtual machines.
- **Scalability**: Docker makes it easy to scale applications by deploying multiple containers across different nodes.

Docker is widely used for developing, shipping, and running applications in a consistent and automated manner.

----------------------------------------------------------------

[gpt-4o] Tokens: 716 prompt + 173 completion = 889 total | Time: 2.92s
```

You can also use shorthand commands:

```bash
# Shorthand command
$ kiwi e "What is version control?"

# Full command
$ kiwi execute "What is version control?"
```

### Shell Command Assistance

Get help with shell commands:

```bash
# Shorthand command
$ kiwi s "find all pdf files modified in the last week"
----------------------------------------------------------------

Suggested command:
find ~ -name "*.pdf" -type f -mtime -7

This command will:
- Search in your home directory (~)
- Find all files with .pdf extension
- Only include regular files (not directories)
- Filter for files modified in the last 7 days

----------------------------------------------------------------

Do you want to execute this command? (y/n): y

/home/user/Documents/report.pdf
/home/user/Downloads/manual.pdf
/home/user/Projects/presentation.pdf
```

The tool will suggest a command and ask for confirmation before executing it.

> **Note**: For complex commands with pipelines, use the execute mode (`kiwi e`) which provides better handling.

### Interactive Chat

Start an interactive chat session that maintains context:

```bash
# Shorthand command
$ kiwi c

# Full command 
$ kiwi chat
```

Example interaction:

```
$ kiwi c

Created new session: session_1743543868
Chat session started. Type 'exit' to end the session.
Using openai model: gpt-4o
----------------------------------------

You: What is HTML?

Kiwi: HTML, or Hypertext Markup Language, is the standard language used to create and design documents on the web. It structures web pages by using a system of tags and attributes to define elements such as headings, paragraphs, links, images, and other content types.

### Key Features of HTML:
- **Tags and Elements**: HTML uses tags to structure content. Tags are enclosed in angle brackets, like `<tag>`. Most elements have an opening tag and a closing tag, e.g., `<p>Content</p>` for paragraphs.
- **Attributes**: Tags can have attributes that provide additional information about elements, such as `id`, `class`, or `style`.
- **Hyperlinks**: HTML supports hyperlinks, allowing users to navigate between web pages using the `<a>` tag.
- **Media**: HTML can embed images, videos, and audio using specific tags like `<img>`, `<video>`, and `<audio>`.
- **Semantic Elements**: Newer versions of HTML introduce semantic elements like `<header>`, `<footer>`, `<article>`, and `<section>`, which provide more meaningful structure to a document.

HTML is a cornerstone technology of the World Wide Web, working alongside CSS (Cascading Style Sheets) and JavaScript to create interactive and visually appealing web pages.

[gpt-4o] Tokens: 713 prompt + 268 completion = 981 total | Time: 5.82s
```

### Debug Mode

Enable debug mode to see token usage and response time:

```bash
# Using command-line flag
$ kiwi e "What is a smartphone" --debug

# Or set it permanently in your config
$ kiwi config set ui.debug true
```

Output with debug mode:

```
$ kiwi "Explain Docker in simple terms" --debug
----------------------------------------------------------------

Docker is like a shipping container for software. It packages your application and all its dependencies into a standardized unit (a container) that can run anywhere Docker is installed.

Key benefits:
- Consistency: Works the same in development and production
- Isolation: Applications run in their own environment
- Portability: Run on any system with Docker installed
- Efficiency: Uses fewer resources than virtual machines

----------------------------------------------------------------

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