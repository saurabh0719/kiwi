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

### Execute Prompts

```bash
# Shorthand command
kiwi e "Explain Docker in simple terms"

# Full command
kiwi execute "Explain Docker in simple terms"
```

**Example Output:**
```
Docker is a platform that packages software into standardized units called containers. Think of containers like shipping containers in the real world:

1. Standardized: Just as shipping containers have standard sizes and fittings, Docker containers package everything an application needs (code, libraries, settings) in a consistent way.

2. Portable: Containers run the same way regardless of environment - your laptop, a server, or the cloud.

3. Isolated: Each container operates independently without interfering with other containers or the host system.

4. Efficient: Unlike virtual machines, containers share the host OS kernel, making them lightweight and quick to start.

Docker makes development and deployment simpler because it eliminates "it works on my machine" problems, as the container includes everything needed to run the application consistently across different environments.
```

### Shell Command Assistance

```bash
# Shorthand command
kiwi s "find all pdf files modified in the last week"

# Full command
kiwi shell "find all pdf files modified in the last week"
```

**Example Output:**
```
find /path/to/search -name "*.pdf" -type f -mtime -7

Do you want to execute this command? (y/n): 
```

### Interactive Chat

```bash
# Shorthand command
kiwi c

# Full command 
kiwi chat
```

**Example Output:**
```
Created new session: session_1712082042
Chat session started. Type 'exit' to end the session.
Using openai model: gpt-4o
----------------------------------------

You: What can you help me with?
```

### Debug Mode

You can enable debug mode to see detailed information about token usage and response time:

```bash
# Using command-line flag
kiwi e "What is a smartphone" --debug

# Or set it permanently in your config
kiwi config set ui.debug true
```

**Example Output with Debug Mode:**
```
A smartphone is a mobile device that combines cellular and mobile computing functions into one unit. Key features of a smartphone include:

- **Touchscreen Interface**: Typically large, high-resolution displays for interaction.
- **Operating System**: Runs a mobile OS like Android or iOS.
- **Connectivity**: Offers cellular connectivity for calls and text, along with Wi-Fi, Bluetooth, and often GPS.
- **App Ecosystem**: Supports a wide range of applications for productivity, communication, and entertainment.
- **Camera**: Integrated high-quality cameras for photos and videos.
- **Sensors**: Includes accelerometers, gyroscopes, and often biometric sensors (e.g., fingerprint, face recognition).

Smartphones are versatile devices used for communication, internet access, media consumption, and various other applications.

[gpt-4o] Tokens: 501 prompt + 162 completion = 663 total | Time: 3.92s
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
--debug             Enable debug mode with verbose output and statistics (default false)
```

### Config File

Create `~/.kiwi/config.yaml`:

```yaml
llm:
  provider: openai
  model: gpt-4
  api_key: your_api_key
  safe_mode: true
ui:
  debug: false
```

## License

This project is licensed under the Apache-2.0 License - see the LICENSE file for details. 
