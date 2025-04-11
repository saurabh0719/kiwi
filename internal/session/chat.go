package session

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/saurabh0719/kiwi/internal/config"
	"github.com/saurabh0719/kiwi/internal/llm"
	"github.com/saurabh0719/kiwi/internal/llm/core"
	"github.com/saurabh0719/kiwi/internal/tools"
	"github.com/saurabh0719/kiwi/internal/util"
)

// System prompt for assistant sessions
const AssistantSystemPrompt = `You are Kiwi, a CLI-based developer assistant. You understand developer workflows and context, and provide helpful responses to technical questions and requests.

Core principles:
- Understand that you are operating in a developer's terminal environment
- Interpret requests in the context of software development, CLI usage, and project management
- Leverage your shell tool capabilities to take direct action when appropriate
- Always use Markdown formatting for your responses to enhance readability

For this assistant session:
- Retain context from previous messages
- Provide comprehensive responses when appropriate
- Ask clarifying questions when user requests are ambiguous
- Balance brevity with completeness based on user's engagement style

When handling shell commands:
- ALWAYS use the shell tool to execute commands when users ask for file operations, git commands, or system tasks
- If the user's request implies running a terminal command, use the shell tool rather than just showing commands
- Examples: "list files," "find large files," "add files to git," etc. should all use the shell tool

For build-related commands:
- When asked to build a project, FIRST examine the project structure with 'ls -la' and look for build files
- For Go projects, look for go.mod and use 'go build'
- For Node.js projects, look for package.json and use 'npm install' followed by 'npm run build'
- For Python projects, look for setup.py, requirements.txt, or pyproject.toml
- For C/C++ projects, look for Makefile or CMakeLists.txt
- ALWAYS execute the actual build commands rather than just suggesting them

When handling user requests:
- Recognize common developer patterns and intentions without needing explicit instructions
- Understand when a request implies running commands or performing file operations
- Execute appropriate actions based on project context and development conventions 
- Assume requests are made in a development context - interpret ambiguous requests as developer tasks
- For requests like "build", "run", "test", or "deploy", take initiative to explore the project and execute appropriate commands

When interacting with users:
- Be conversational yet efficient
- Show your reasoning when solving complex problems
- Present multiple approaches for complex questions
- Adapt your technical level to match the user's demonstrated expertise
- If a user indicates something isn't working, actively diagnose the issue
- Provide context-aware assistance that understands developer workflows and project structures

Remember this is an ongoing conversation where context builds over time, and you're expected to be proactive in helping with development tasks.`

// ProcessChatMessage handles a single message in the chat, either in interactive or non-interactive mode
// It adds the user message to the session, gets a response from the LLM, and adds the response to the session
func ProcessChatMessage(mgr *Manager, sess Session, cfg config.Config, adapter core.Adapter, userInput string) error {
	if err := mgr.AddMessage(sess.ID, "user", userInput); err != nil {
		return fmt.Errorf("failed to add user message: %w", err)
	}

	updatedSess, err := mgr.GetSession(sess.ID)
	if err != nil {
		return fmt.Errorf("failed to get updated session: %w", err)
	}

	var messages []llm.Message
	if len(updatedSess.Messages) == 1 {
		messages = append(messages, llm.Message{
			Role:    "system",
			Content: AssistantSystemPrompt,
		})
	}

	for _, msg := range updatedSess.Messages {
		messages = append(messages, llm.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Initialize a complete response string to store the entire response
	completeResponse := ""

	// For buffering partial chunks in streaming mode
	var responseBuffer strings.Builder
	const flushThreshold = 100

	// Check if we should render markdown
	shouldRenderMarkdown := util.ShouldRenderMarkdown(cfg.UI.RenderMarkdown)

	// Define a stream handler for handling the streaming response
	streamHandler := func(chunk string) error {
		// Add chunk to the buffer
		responseBuffer.WriteString(chunk)

		// Append the chunk to the complete response (unformatted for storage)
		completeResponse += chunk

		// If we've accumulated enough text or hit a natural break, format and print
		if responseBuffer.Len() >= flushThreshold ||
			strings.HasSuffix(chunk, "\n") ||
			strings.HasSuffix(chunk, "\r") {

			// Render and print the buffered content
			fmt.Print(util.RenderMarkdown(responseBuffer.String(), shouldRenderMarkdown))

			// Clear the buffer for next chunks
			responseBuffer.Reset()
		}

		return nil
	}

	// Start a spinner that will stop once we get the first token
	// Using the global spinner manager
	spinnerManager := util.GetGlobalSpinnerManager()
	spinnerManager.StartThinkingSpinner("Thinking...")

	// Track time for metrics
	startTime := time.Now()
	var metrics *core.ResponseMetrics

	// Print a newline before any response
	fmt.Println()

	// Flag to track if we've printed the prefix
	prefixPrinted := false

	if cfg.UI.Streaming {
		// Use streaming API with the handler
		metrics, err = adapter.ChatStream(context.Background(), messages, func(chunk string) error {
			// On first chunk, make sure no spinner is active and print prefix
			if !prefixPrinted {
				// Clear spinner before printing any output
				util.PrepareForResponse(spinnerManager)
				// Print the Kiwi prefix
				util.AssistantColor.Print("Kiwi: ")
				prefixPrinted = true
			}

			return streamHandler(chunk)
		})

		// Flush any remaining content in the buffer
		if responseBuffer.Len() > 0 {
			fmt.Print(util.RenderMarkdown(responseBuffer.String(), shouldRenderMarkdown))
		}
	} else {
		// Use non-streaming API for complete response at once
		var response string
		response, metrics, err = adapter.ChatWithMetrics(context.Background(), messages)

		// Clear spinner before printing any output
		util.PrepareForResponse(spinnerManager)
		// Print the Kiwi prefix
		util.AssistantColor.Print("Kiwi: ")
		prefixPrinted = true

		// Print the complete response
		fmt.Print(util.RenderMarkdown(response, shouldRenderMarkdown))
		completeResponse = response
	}

	// At the end of the function, after processing the response
	// Just print a newline after the response
	fmt.Println()

	if err != nil {
		return fmt.Errorf("failed to get response: %w", err)
	}

	// If metrics is nil (can happen if the stream fails), create empty metrics
	if metrics == nil {
		metrics = &core.ResponseMetrics{
			ResponseTime: time.Since(startTime),
		}
	}

	if cfg.UI.Debug {
		core.PrintResponseMetrics(metrics, adapter.GetModel())
		util.PrintChatDivider()
	}

	if err := mgr.AddMessage(sess.ID, "assistant", completeResponse); err != nil {
		return fmt.Errorf("failed to add assistant message: %w", err)
	}

	return nil
}

// ProcessStream processes a streaming response from the LLM
// This is used for one-off requests that don't need to be stored in a session
func ProcessStream(adapter llm.Adapter, messages []llm.Message, userPrompt string, exitCode *int, isFirstResponse *bool, isExecuteMode bool, renderMarkdown bool) error {
	spinnerManager := util.GetGlobalSpinnerManager()
	spinnerManager.StartThinkingSpinner("Thinking...")

	// For buffering partial chunks in streaming mode
	var responseBuffer strings.Builder
	const flushThreshold = 100

	// Use the passed rendering preference
	shouldRenderMarkdown := util.ShouldRenderMarkdown(renderMarkdown)

	// Process the streaming response with shared tool detection
	_, toolCallDetected, _, err := llm.ProcessStreamWithToolDetection(
		context.Background(),
		adapter,
		messages,
		func(chunk string) error {
			// Print the first bit of output
			if *isFirstResponse {
				// Clear thinking spinner and prepare for output
				util.PrepareForResponse(spinnerManager)

				if !isExecuteMode {
					fmt.Print("\n")
					util.AssistantColor.Print("Kiwi: ")
				}
				*isFirstResponse = false
			}

			// Add chunk to the buffer
			responseBuffer.WriteString(chunk)

			// If we've accumulated enough text or hit a natural break, format and print
			if responseBuffer.Len() >= flushThreshold ||
				strings.HasSuffix(chunk, "\n") ||
				strings.HasSuffix(chunk, "\r") {

				// Render and print the buffered content
				fmt.Print(util.RenderMarkdown(responseBuffer.String(), shouldRenderMarkdown))

				// Clear the buffer for next chunks
				responseBuffer.Reset()
			}

			return nil
		},
		tools.DefaultExecutionDetector,
	)

	// Flush any remaining content in the buffer
	if responseBuffer.Len() > 0 {
		fmt.Print(util.RenderMarkdown(responseBuffer.String(), shouldRenderMarkdown))
	}

	// Handle null content errors gracefully using shared handler
	if err != nil {
		// Check if this is a null content error that can be gracefully handled
		if tools.HandleNullContentError(err, toolCallDetected) {
			// If we detected a tool call but got null content, this is likely after a successful
			// command execution. We'll print a fallback message.
			if *isFirstResponse {
				util.PrepareForResponse(spinnerManager)
				if !isExecuteMode {
					fmt.Print("\n")
					util.AssistantColor.Print("Kiwi: ")
				}
				*isFirstResponse = false
				fmt.Println("Command executed successfully.")
			}

			// Return without error since we handled it
			return nil
		}

		return fmt.Errorf("error getting streaming response: %w", err)
	}

	// Check if exitCode was updated by a tool
	if exitCode != nil && *exitCode != 0 {
		return fmt.Errorf("operation failed with exit code %d", *exitCode)
	}

	return nil
}

// UpdateSessionSummaryLLM generates a summary by finding the most significant user message
// This simplified version doesn't use the LLM but instead finds and trims the longest user message
func UpdateSessionSummaryLLM(mgr *Manager, sessionID string, cfg *config.Config) error {
	sess, err := mgr.GetSession(sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// If there are no messages, nothing to do
	if len(sess.Messages) == 0 {
		return nil
	}

	// Find the most significant (longest) user message
	var significantMessage string
	var maxLength int

	for _, msg := range sess.Messages {
		if msg.Role == "user" {
			// Remove any common markdown formatting and whitespace
			cleanContent := strings.TrimSpace(msg.Content)

			// Consider this message if it's longer than what we have
			if len(cleanContent) > maxLength {
				maxLength = len(cleanContent)
				significantMessage = cleanContent
			}
		}
	}

	// If we didn't find any user messages, use first message or default
	if significantMessage == "" {
		if len(sess.Messages) > 0 {
			significantMessage = strings.TrimSpace(sess.Messages[0].Content)
		} else {
			significantMessage = "New chat session"
		}
	}

	// Truncate to the first line or sentence if very long
	if idx := strings.IndexAny(significantMessage, "\n\r"); idx > 0 {
		significantMessage = significantMessage[:idx]
	}

	// Truncate if still too long (max 60 chars)
	if len(significantMessage) > 60 {
		significantMessage = significantMessage[:57] + "..."
	}

	// Log the summary being set
	util.InfoColor.Printf("Setting session summary: %s\n", significantMessage)

	// Update the session with the summary
	return mgr.UpdateSessionSummary(sessionID, significantMessage)
}
