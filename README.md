:heavy_exclamation_mark: This repo is a WIP.


# Kiwi

A command-line interface for interacting with Large Language Models (LLMs) in terminal environments.

## Key Features

- **Multiple LLM Providers**: Support for OpenAI and Claude APIs
- **Interactive Chat**: Maintain context in ongoing conversations
- **Shell Command Assistant**: Get command suggestions for tasks
- **Execute Mode**: Run one-off prompts for quick answers
- **Built-in Tools**: Access filesystem operations and system information

## Building and Installation

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

# Cross-compile for different platforms
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
# Option 1: Move to a directory already in PATH
sudo mv kiwi /usr/local/bin/

# Option 2: Add to PATH in your current session
export PATH=$PATH:$(pwd)

# Option 3: Add to PATH permanently (add to your ~/.bashrc or ~/.zshrc)
echo 'export PATH=$PATH:/path/to/kiwi/directory' >> ~/.bashrc
source ~/.bashrc
```

#### macOS

```bash
# Option 1: Move to a directory already in PATH
sudo mv kiwi /usr/local/bin/

# Option 2: Using Homebrew's directory (if you use Homebrew)
mv kiwi $(brew --prefix)/bin/

# Option 3: Add to PATH permanently (add to your ~/.bash_profile or ~/.zshrc)
echo 'export PATH=$PATH:/path/to/kiwi/directory' >> ~/.zshrc
source ~/.zshrc
```

#### Windows

```powershell
# Option 1: Move to a directory in PATH (PowerShell with admin rights)
Move-Item -Path .\kiwi.exe -Destination "C:\Windows\System32\"

# Option 2: Add current directory to PATH for current session
$env:Path += ";$(Get-Location)"

# Option 3: Add to PATH permanently (requires admin privileges)
[Environment]::SetEnvironmentVariable("Path", $env:Path + ";C:\path\to\kiwi", "User")
```

### Verifying Installation

After installation, verify that Kiwi is working correctly:

```bash
# Check version and basic info
kiwi --help
```

## Usage

### Interactive Chat

```bash
kiwi chat
```

### Shell Command Assistance

```bash
kiwi shell "find all pdf files modified in the last week"
```

### Execute Prompts

```bash
kiwi execute "Explain Docker in simple terms"
```

### Manage Configuration

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

## Configuration

### API Keys

Set your API keys using environment variables:

```bash
# For OpenAI
export OPENAI_API_KEY=your_api_key

# For Claude
export ANTHROPIC_API_KEY=your_api_key
```

### Command Line Options

```bash
--provider string   LLM provider (openai, claude) (default "openai")
--model string      Model to use (default "gpt-3.5-turbo")
--api-key string    API key (if not set via environment variable)
--safe-mode         Enable command confirmation (default true)
```

### Config File

Create `~/.kiwi/config.yaml`:

```yaml
llm:
  provider: openai
  model: gpt-4
  api_key: your_api_key
  safe_mode: true
```

## License

This project is licensed under the Apache-2.0 License - see the LICENSE file for details. 
