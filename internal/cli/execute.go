package cli

import (
	"context"
	"fmt"
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

When handling shell commands:
- ALWAYS use the shell tool to execute commands when users ask for file operations, git commands, or system tasks
- If the user's request implies running a terminal command, use the shell tool rather than just showing commands
- Examples: "list files," "find large files," "add files to git," etc. should all use the shell tool

For technical content:
- Use code blocks for commands and code snippets
- Include brief explanations when needed
- Format output to be easily read in terminal environments
- Focus on practical solutions over background information

Remember that users in execute mode typically want quick, actionable information without ongoing conversation.`,
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
				// Print the chunk without a newline
				fmt.Print(chunk)
				return nil
			},
			llm.DefaultToolExecutionDetector,
		)

		// Handle specific error cases gracefully
		if err != nil {
			// Use shared error handler for null content after tool execution
			if llm.HandleNullContentError(err, toolExecuted) {
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
			llm.DefaultToolExecutionDetector,
		)

		// Handle null content errors in non-streaming mode
		if err != nil {
			// Use shared error handler
			if llm.HandleNullContentError(err, toolExecuted) {
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

		// Print the complete response
		fmt.Println(response)
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
