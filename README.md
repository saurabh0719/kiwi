# Kiwi

![Image](https://github.com/user-attachments/assets/9e323659-f603-4b8f-8a24-da5f19169c38)

A command-line interface (CLI) for interacting with Large Language Models (LLMs) directly from your terminal.

[![GitHub license](https://img.shields.io/github/license/saurabh0719/kiwi)](https://github.com/saurabh0719/kiwi/blob/main/LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.18+-00ADD8?style=flat&logo=go)](https://github.com/saurabh0719/kiwi)


![Image](https://github.com/user-attachments/assets/682a1f30-62b1-471f-9dd7-348c661dc4ed)

<span id="installation"></span>
## 📦 Installation

### Quick Install (Linux and macOS)

```bash
# Install with a single command
curl -fsSL https://raw.githubusercontent.com/saurabh0719/kiwi/main/install.sh | bash
```

## ✨ Demo 
[![asciicast](https://asciinema.org/a/pf9axJfiITOzW4dZYlwIOu3qd.svg)](https://asciinema.org/a/pf9axJfiITOzW4dZYlwIOu3qd)

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
- **Interactive Assistant**: Maintain context in ongoing conversations
- **Built-in Tools**: Filesystem operations, shell commands, system information
- **Terminal Command Assistance**: Get terminal command suggestions with confirmation

## 📑 Table of Contents

* [Installation](#installation)
* [Usage](#usage)
  * [API Keys](#api-keys)
  * [Execute Mode](#execute-mode)
  * [Interactive Assistant](#interactive-assistant)
  * [Tool Calls](#tool-calls)
  * [Built-in Tools](#built-in-tools)
  * [Terminal Command Assistance](#terminal-command-assistance)
  * [Debug Mode](#debug-mode)
* [Configuration](#configuration)
* [Contributing](#contributing)
  * [Adding New LLM Providers](#adding-new-llm-providers)
  * [Creating Custom Tools](#creating-custom-tools)
* [License](#license)

<span id="usage"></span>
## 🚀 Usage

<span id="api-keys"></span>
### 🔑 API Keys

You'll need to provide your own API keys for the LLM providers you want to use. Set your API keys using the config command:

```bash
kiwi -c set llm.provider openai
kiwi -c set llm.model gpt-4o
kiwi -c set llm.api_key your-api-key-here
```

> **Note**: Currently only OpenAI is supported. This repository is open for contributions to add more LLM providers.

<span id="execute-mode"></span>
### ⚡ Execute Mode

Get quick answers without starting a full assistant session:

```bash
# Direct execution (default behavior) - no quotes needed
$ kiwi What is Docker?
----------------------------------------------------------------

Docker is an open-source platform that automates the deployment, scaling, and management of applications using containerization. Containers allow you to package an application with all its dependencies into a standardized unit for software development. This ensures that the application runs consistently across different environments.

Key features of Docker include:
- **Portability**: Containers can run on any system that supports Docker, making it easy to move applications between environments.
- **Isolation**: Each container runs in a separate environment, ensuring that applications do not interfere with each other.
- **Efficiency**: Containers share the host OS kernel, making them lightweight and fast to start compared to virtual machines.
- **Scalability**: Docker makes it easy to scale applications by deploying multiple containers across different nodes.

Docker is widely used for developing, shipping, and running applications in a consistent and automated manner.

----------------------------------------------------------------

[gpt-4o] Tokens: 716 prompt + 173 completion = 889 total | Time: 2.92s (LLM: 2.82s, Tools: 0.00s, Other: 0.10s)
```

This example shows the timing breakdown in execute mode, demonstrating that for simple queries without tool calls, almost all the time is spent in LLM processing.

![Image](https://github.com/user-attachments/assets/85527dba-6e84-41f5-9e3f-e98b72814a00)


<span id="interactive-assistant"></span>
### 💬 Interactive Assistant

Start an interactive assistant session that maintains context:

```bash
# Assistant mode
$ kiwi -a
```

Example interaction:

```
$ kiwi -a

Created new session: session_1743543868
Assistant session started. Type 'exit' to end the session.
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

[gpt-4o] Tokens: 713 prompt + 268 completion = 981 total | Time: 5.82s (LLM: 5.74s, Tools: 0.00s, Other: 0.08s)
```

In the interactive assistant mode, the timing breakdown shows that most of the time (5.74s) is spent in LLM processing, with minimal overhead (0.08s) and no tool usage for this simple query.

![Image](https://github.com/user-attachments/assets/c7e32bd8-22b6-4141-814f-554c0d3db87b)

<span id="tool-calls"></span>
### 🛠️ Tool Calls

Kiwi can use built-in tools to perform tasks like interacting with the filesystem:

![Image](https://github.com/user-attachments/assets/6d5ea17c-b49a-4dc9-8c6f-344950323853)

```bash
...

⠋ [Tool: filesystem] executing...🔧 [Tool: filesystem:list] executed in 0.000s
  → Requested operation: list
  → Validating path: .
  → Path safety check passed
  → Listing files in directory: .
  → Listed 15 files/directories in .

⠋ [Tool: filesystem] executing...🔧 [Tool: filesystem:read] executed in 0.000s
  → Requested operation: read
  → Validating path: README.md
  → Path safety check passed
  → Reading file content: README.md
  → Successfully read README.md (303 lines, 11300 bytes)

...
```

In this example, notice how the time is broken down between LLM processing (4.25s), tool execution (0.32s for filesystem:list), and other overhead (0.89s), giving you insights into where the processing time is being spent.

<span id="built-in-tools"></span>
### 🧰 Built-in Tools

Kiwi comes with several built-in tools that extend its capabilities beyond simple text responses.

#### 📂 Filesystem Tool

**Operations:**
- **list**: Lists all files and directories in a specified path
- **read**: Reads the contents of a specified file
- **write**: Creates or updates a file with specified content
- **delete**: Deletes a specified file

#### 🖥️ Shell Tool

**Features:**
- Limited to a whitelist of safe commands (`ls`, `cat`, `grep`, `find`, `pwd`, `head`, `tail`, `wc`, `echo`, `date`, `ps`, `df`, `du`, `free`, `top`)
- Pipeline support for combining simple commands
- Command validation to prevent dangerous operations

#### 🔍 System Info Tool

**Information Types:**
- **basic**: General system information (OS, architecture, CPU count, hostname, etc.)
- **memory**: Memory usage statistics
- **env**: Non-sensitive environment variables

#### 🌐 Web Search Tool

**Features:**
- HTML content extraction to provide readable text
- Content truncation for very large pages

<span id="terminal-command-assistance"></span>
### 🔧 Terminal Command Assistance

Get help with terminal commands using the `-t` flag:

```bash
# Terminal command assistance
$ kiwi -t find all pdf files modified in the last week
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

> **Note**: For complex commands with pipelines, use execute mode which provides better handling.

<span id="debug-mode"></span>
### 🐞 Debug Mode

Enable debug mode to see token usage and detailed timing metrics:

```bash
# Using command-line flag
$ kiwi What is a smartphone --debug

# Or set it permanently in your config
$ kiwi -c set ui.debug true
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

[gpt-4o] Tokens: 501 prompt + 162 completion = 663 total | Time: 3.92s (LLM: 3.42s, Tools: 0.00s, Other: 0.50s)
```

#### Understanding Timing Metrics

When debug mode is enabled, Kiwi provides a detailed breakdown of execution time:

- **Total Time**: The overall time taken from sending the request to receiving the complete response
- **LLM Time**: Time spent in the language model processing your request and generating a response
- **Tools Time**: Time spent executing tools (file operations, web searches, etc.)
- **Other Time**: Overhead time spent on networking, parsing, and other internal operations

These metrics can help you understand where time is being spent during request processing, which can be valuable for diagnosing performance issues or optimizing your workflows.

### 📺 Streaming Mode

Control whether responses appear incrementally (streaming) or all at once:

```bash
# Using command-line flag
$ kiwi What is quantum computing? --streaming=false  # Display complete answer at once

# Or set it permanently in your config
$ kiwi -c set ui.streaming false  # Disable streaming for all commands
```

When streaming is enabled (default), you'll see the response being generated word by word.
When disabled, the response will appear all at once after it's complete, which can be useful for scripting or when you prefer to see the finished response.

<span id="configuration"></span>
## ⚙️ Configuration

Kiwi provides a simple configuration system:

```bash
# List all settings
kiwi -c

# Get/set specific settings
kiwi -c get llm.provider
kiwi -c set llm.provider openai
kiwi -c set llm.model gpt-4o
kiwi -c set llm.api_key your_api_key
kiwi -c set ui.debug true
kiwi -c set ui.streaming true
```

### UI Options

- **Debug Mode** (`ui.debug`): When enabled, shows token usage, response time, and API cost statistics after each response
- **Streaming Mode** (`ui.streaming`): Controls whether responses are displayed incrementally (true) or all at once when completed (false)


<span id="contributing"></span>
## 🤝 Contributing

Contributions to Kiwi are welcome! If you're interested in contributing, please follow these steps:

1. Fork the repository
2. Create a new branch
3. Make your changes
4. Run the tests
5. Submit a pull request

Kiwi needs more contribution to add support for more LLM providers and create custom tools to enhance functionality for CLI specific workflows.

<span id="adding-new-llm-providers"></span>
### 🧠 Adding New LLM Providers

You can extend Kiwi to work with additional LLM providers beyond the built-in ones:

```go
// Create a new provider that implements the Adapter interface
type ClaudeAdapter struct {
    apiKey string
    model  string
    tools  *tools.Registry
}

// Implement required methods
func (c *ClaudeAdapter) Chat(ctx context.Context, messages []core.Message) (string, error) {
    // Implementation for Claude API
    // ...
}

func (c *ClaudeAdapter) ChatWithMetrics(ctx context.Context, messages []core.Message) (string, *core.ResponseMetrics, error) {
    // Implementation with metrics
    // ...
}

func (c *ClaudeAdapter) ChatStream(ctx context.Context, messages []core.Message, handler core.StreamHandler) (*core.ResponseMetrics, error) {
    // Implementation for streaming responses
    // ...
}

func (c *ClaudeAdapter) GetModel() string {
    return c.model
}

func (c *ClaudeAdapter) GetProvider() string {
    return "claude"
}

// Register your provider factory
func init() {
    llm.RegisterAdapter("claude", func(model, apiKey string, tools *tools.Registry) (core.Adapter, error) {
        return &ClaudeAdapter{
            apiKey: apiKey,
            model:  model,
            tools:  tools,
        }, nil
    })
}
```

<span id="creating-custom-tools"></span>
### 🛠️ Creating Custom Tools

Add your own tools to extend Kiwi's capabilities:

```go
// Define a new tool
type WeatherTool struct {
    apiKey string
}

// Implement the Tool interface
func (w *WeatherTool) Name() string {
    return "weather"
}

func (w *WeatherTool) Description() string {
    return "Get current weather information for a location"
}

func (w *WeatherTool) Parameters() map[string]core.Parameter {
    return map[string]core.Parameter{
        "location": {
            Type:        "string",
            Description: "City name or coordinates",
            Required:    true,
        },
        "units": {
            Type:        "string",
            Description: "Temperature units (metric/imperial)",
            Required:    false,
            Default:     "metric",
        },
    }
}

func (w *WeatherTool) Execute(ctx context.Context, params map[string]interface{}) (core.ToolExecutionResult, error) {
    result := core.ToolExecutionResult{
        ToolMethod: "weather",
    }
    
    location, ok := params["location"].(string)
    if !ok {
        return result, errors.New("location parameter is required")
    }
    
    // Call weather API
    result.AddStep("Retrieving weather data for " + location)
    
    // Process and return results
    result.Output = "Weather information for " + location + "..."
    
    return result, nil
}

// Register your tool
func init() {
    tools.Register("weather", func() core.Tool {
        return &WeatherTool{apiKey: os.Getenv("WEATHER_API_KEY")}
    })
}
```

<span id="license"></span>
## 📄 License

Kiwi is licensed under the MIT License. See the [LICENSE](https://github.com/saurabh0719/kiwi/blob/main/LICENSE) file for more details.
