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

func initExecuteCmd() {
	executeCmd = &cobra.Command{
		Use:   "execute",
		Short: "Execute a prompt",
		Long:  `Execute a prompt and get a response from the LLM.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleExecute(cmd, args[0])
		},
	}
}

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

	// Initialize a complete response string to store the entire response
	completeResponse := ""

	// Start the loading spinner
	spinner := util.NewSpinner("Generating response...")
	spinner.Start()

	// Print the divider before the response begins
	util.OutputColor.Println("----------------------------------------------------------------\n")

	// Track time for metrics
	startTime := time.Now()
	var metrics *core.ResponseMetrics

	if cfg.UI.Streaming {
		// Stream the response
		metrics, err = adapter.ChatStream(context.Background(), messages, func(chunk string) error {
			// Stop the spinner on first token
			if spinner != nil {
				spinner.Stop()
				spinner = nil
			}
			// Print the chunk without a newline
			fmt.Print(chunk)
			// Append the chunk to the complete response
			completeResponse += chunk
			return nil
		})
	} else {
		// Get the complete response at once
		var response string
		response, metrics, err = adapter.ChatWithMetrics(context.Background(), messages)

		// Stop the spinner when response is received
		if spinner != nil {
			spinner.Stop()
			spinner = nil
		}

		// Print the complete response
		fmt.Println(response)
		completeResponse = response
	}

	// If the spinner is still running (no tokens received), stop it
	if spinner != nil {
		spinner.Stop()
	}

	// Print the divider after the response
	util.OutputColor.Println("\n\n----------------------------------------------------------------")

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
		util.StatsColor.Printf("\n[%s] Tokens: %d prompt + %d completion = %d total | Time: %.2fs\n",
			adapter.GetModel(),
			metrics.PromptTokens,
			metrics.CompletionTokens,
			metrics.TotalTokens,
			metrics.ResponseTime.Seconds())
	}

	return nil
}
