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
# For OpenAI
export OPENAI_API_KEY=your_api_key

# For Claude
export ANTHROPIC_API_KEY=your_api_key
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

### System Info Tools

Retrieve system information:

```
- system_info: Get information about the operating system
- disk_usage: Check disk space usage
- memory_info: Get memory usage statistics
```

## Development

### Project Structure

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

### Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the Apache-2.0 License - see the LICENSE file for details. 
