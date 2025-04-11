package cli

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
	"github.com/spf13/cobra"
)

// handleExecute processes a direct query/prompt without starting an interactive session
func handleExecute(cmd *cobra.Command, prompt string) error {
	cfg, err := config.Load(cmd)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	toolRegistry := tools.NewRegistry()
	tools.RegisterStandardTools(toolRegistry)

	adapter, err := llm.NewAdapter(cfg.LLM.Provider, cfg.LLM.Model, cfg.LLM.APIKey, toolRegistry)
	if err != nil {
		return fmt.Errorf("failed to create LLM adapter: %w", err)
	}

	messages := []llm.Message{
		{
			Role: "system",
			Content: `You are Kiwi in execute mode. This mode is designed for one-off, direct queries that require crisp, focused responses.

Since this is a single interaction:
- Provide complete but concise responses
- Don't ask clarifying questions - do your best with the information provided
- Optimize for efficiency and immediate utility
- Format responses for terminal readability
- Always use Markdown for your responses to enhance readability

When handling shell commands:
- ALWAYS use the shell tool to execute commands when users ask for file operations, git commands, or system tasks
- If the user's request implies running a terminal command, use the shell tool rather than just showing commands
- Examples: "list files," "find large files," "add files to git," etc. should all use the shell tool

For technical content:
- Use code blocks for commands and code snippets
- Include brief explanations when needed
- Format output to be easily read in terminal environments
- Focus on practical solutions over background information

As a developer-focused CLI tool:
- Interpret commands in a development context first and foremost
- Recognize common developer intentions and take direct action when appropriate
- For requests like "build", "run", "test", or similar operations, explore the project structure first, then execute relevant commands
- Understand project conventions and standard development workflows without requiring explicit details
- When executing developer operations, take initiative to determine the appropriate command based on project type

Remember that users in execute mode typically want quick, actionable information without ongoing conversation, and expect you to understand the development context of their requests.`,
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	// Get the global spinner manager
	spinnerManager := util.GetGlobalSpinnerManager()

	// Start the thinking spinner
	spinnerManager.StartThinkingSpinner("Generating response...")

	// Print the divider before the response begins
	util.PrintExecuteStartDivider()

	// Track time for metrics
	startTime := time.Now()
	var metrics *core.ResponseMetrics
	var completeResponse string
	var toolExecuted bool

	// For buffering partial chunks in streaming mode
	var responseBuffer strings.Builder
	const flushThreshold = 100

	// Check if we should render markdown
	shouldRenderMarkdown := util.ShouldRenderMarkdown(cfg.UI.RenderMarkdown)

	if cfg.UI.Streaming {
		// Use our shared streaming handler with tool detection
		metrics, toolExecuted, completeResponse, err = llm.ProcessStreamWithToolDetection(
			context.Background(),
			adapter,
			messages,
			func(chunk string) error {
				// On first chunk, make sure no spinner is active
				if completeResponse == "" {
					// Clear spinner before printing any output
					util.PrepareForResponse(spinnerManager)
				}

				// Collect the chunk in our buffer
				responseBuffer.WriteString(chunk)

				// If we've collected enough text or the chunk ends with a newline,
				// render and flush the buffer
				if responseBuffer.Len() >= flushThreshold ||
					strings.HasSuffix(chunk, "\n") ||
					strings.HasSuffix(chunk, "\r") {

					// Render and print the buffered content
					fmt.Print(util.RenderMarkdown(responseBuffer.String(), shouldRenderMarkdown))

					// Clear the buffer for the next chunks
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

		// Handle specific error cases gracefully
		if err != nil {
			// Use shared error handler for null content after tool execution
			if tools.HandleNullContentError(err, toolExecuted) {
				// We'll print the response divider and continue as normal
				if completeResponse == "" {
					fmt.Println("\nCommand executed successfully.")
				}
				util.PrintExecuteEndDivider()

				// Create minimal metrics
				metrics = &core.ResponseMetrics{
					ResponseTime: time.Since(startTime),
				}

				// Print debug info if enabled
				if cfg.UI.Debug {
					core.PrintResponseMetrics(metrics, adapter.GetModel())
				}

				// Return without error
				return nil
			}

			// For other errors, show the divider and return the error
			util.PrintExecuteEndDivider()
			return fmt.Errorf("failed to get response: %w", err)
		}
	} else {
		// Use shared non-streaming handler
		var response string
		response, toolExecuted, metrics, err = llm.HandleNonStreamingResponse(
			context.Background(),
			adapter,
			messages,
			tools.DefaultExecutionDetector,
		)

		// Handle null content errors in non-streaming mode
		if err != nil {
			// Use shared error handler
			if tools.HandleNullContentError(err, toolExecuted) {
				// Print a generic success message
				util.PrepareForResponse(spinnerManager)
				fmt.Println("Command executed successfully.")

				// Create minimal metrics for debug mode
				metrics = &core.ResponseMetrics{
					ResponseTime: time.Since(startTime),
				}

				util.PrintExecuteEndDivider()

				// Print debug info if enabled
				if cfg.UI.Debug {
					core.PrintResponseMetrics(metrics, adapter.GetModel())
				}

				// Return without error
				return nil
			}

			util.PrintExecuteEndDivider()
			return fmt.Errorf("failed to get response: %w", err)
		}

		// Clear spinner before printing any output
		util.PrepareForResponse(spinnerManager)

		// Print the complete response with Markdown rendering
		fmt.Println(util.RenderMarkdown(response, shouldRenderMarkdown))
		completeResponse = response
	}

	// No need to stop spinners again, as it's already done before printing the output
	// Just print the divider after the response
	util.PrintExecuteEndDivider()

	// If metrics is nil (can happen if the stream fails), create empty metrics
	if metrics == nil {
		metrics = &core.ResponseMetrics{
			ResponseTime: time.Since(startTime),
		}
	}

	if cfg.UI.Debug {
		core.PrintResponseMetrics(metrics, adapter.GetModel())
	}

	return nil
}
