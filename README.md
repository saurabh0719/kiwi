# Kiwi

![Image](https://github.com/user-attachments/assets/9e323659-f603-4b8f-8a24-da5f19169c38)

A command-line interface (CLI) for interacting with Large Language Models (LLMs) directly from your terminal.

[![GitHub license](https://img.shields.io/github/license/saurabh0719/kiwi)](https://github.com/saurabh0719/kiwi/blob/main/LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.18+-00ADD8?style=flat&logo=go)](https://github.com/saurabh0719/kiwi)

## 📦 Installation

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

## ✨ Features

- **Execute Mode**: Run one-off prompts for quick answers
- **Interactive Chat**: Maintain context in ongoing conversations
- **Built-in Tools**: Filesystem operations, shell commands, system information
- **Shell Command Assistance**: Get terminal command suggestions with confirmation

## 📑 Table of Contents

* [Installation](#-installation)
* [Usage](#-usage)
  * [API Keys](#-api-keys)
  * [Execute Mode](#-execute-mode)
  * [Interactive Chat](#-interactive-chat)
  * [Tool Calls](#-tool-calls)
  * [Shell Command Assistance](#-shell-command-assistance)
  * [Debug Mode](#-debug-mode)
* [Configuration](#-configuration)
* [Built-in Tools](#-built-in-tools)
* [Custom Tools](#-custom-tools)
* [Project Structure](#-project-structure)
* [Contributing](#-contributing)
* [License](#-license)

## 🚀 Usage

### 🔑 API Keys

Set your API keys using the config command:

```bash
kiwi config set llm.provider openai
kiwi config set llm.model gpt-4o
kiwi config set llm.api_key your-api-key-here
```

> **Note**: Currently only OpenAI is supported. Claude support will be added in a future update.

### ⚡ Execute Mode

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

### 💬 Interactive Chat

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

### 🛠️ Tool Calls

Kiwi can use built-in tools to perform tasks like interacting with the filesystem:

```bash
$ kiwi "summarise the contents of this directory"
🔧 [Tool: filesystem:list] executed in 0.000s
----------------------------------------------------------------

Here's a summary of the contents of the current directory:

- **Directories:**
  - `.git/`: Version control directory for Git
  - `.github/`: Contains GitHub-specific configurations or workflows
  - `bin/`: Likely contains compiled binaries or scripts
  - `cmd/`: Typically used for command-line tools or applications
  - `docs/`: Documentation files
  - `internal/`: Contains internal packages or modules

- **Files:**
  - `.gitignore`: Specifies files and directories to be ignored by Git
  - `LICENSE`: License information for the project
  - `Makefile`: Contains build instructions and commands
  - `README.md`: Project description and usage instructions
  - `go.mod`: Module definition file for Go projects
  - `go.sum`: Checksums for Go module dependencies
  - `install.sh`: Shell script for installation
  - `kiwi`: Possibly an executable or script related to the project

This structure suggests a Go-based project with version control and documentation setup.

----------------------------------------------------------------

[gpt-4o] Tokens: 1500 prompt + 236 completion = 1736 total | Time: 5.46s
```

### 🔧 Shell Command Assistance

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

### 🐞 Debug Mode

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

### 📺 Streaming Mode

Control whether responses appear incrementally (streaming) or all at once:

```bash
# Using command-line flag
$ kiwi "What is quantum computing?" --streaming=false  # Display complete answer at once

# Or set it permanently in your config
$ kiwi config set ui.streaming false  # Disable streaming for all commands
```

When streaming is enabled (default), you'll see the response being generated word by word.
When disabled, the response will appear all at once after it's complete, which can be useful for scripting or when you prefer to see the finished response.

## ⚙️ Configuration

Kiwi provides a simple configuration system:

```bash
# List all settings
kiwi config list

# Get/set specific settings
kiwi config get llm.provider
kiwi config set llm.provider openai
kiwi config set llm.model gpt-4o
kiwi config set llm.api_key your_api_key
kiwi config set ui.debug true
kiwi config set ui.streaming true
```

### UI Options

- **Debug Mode** (`ui.debug`): When enabled, shows token usage, response time, and API cost statistics after each response
- **Streaming Mode** (`ui.streaming`): Controls whether responses are displayed incrementally (true) or all at once when completed (false)