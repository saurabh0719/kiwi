package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/saurabh0719/kiwi/internal/config"
	"github.com/saurabh0719/kiwi/internal/llm"
	"github.com/saurabh0719/kiwi/internal/tools"
	"github.com/saurabh0719/kiwi/internal/util"
	"github.com/spf13/cobra"
)

func initShellCmd() {
	shellCmd = &cobra.Command{
		Use:   "shell",
		Short: "Get help with a shell command",
		Long:  `Get help with a shell command. The LLM will provide the appropriate command to accomplish the task.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleShellHelp(cmd, args[0])
		},
	}
}

func handleShellHelp(cmd *cobra.Command, prompt string) error {
	cfg, err := config.Load(cmd)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	toolRegistry := tools.NewRegistry()
	toolRegistry.Register(tools.NewShellTool())

	adapter, err := llm.NewAdapter(cfg.LLM.Provider, cfg.LLM.Model, cfg.LLM.APIKey, toolRegistry)
	if err != nil {
		return fmt.Errorf("failed to create LLM adapter: %w", err)
	}

	messages := []llm.Message{
		{
			Role: "system",
			Content: `You are Kiwi in shell command mode. Your sole purpose is to generate the most appropriate shell command for the user's request.

Response format:
Provide ONLY the command itself with no explanations, preamble, or follow-up text.

Command requirements:
- Must be valid shell syntax for Unix/Linux environments
- Include all necessary arguments and options
- Use appropriate flags for usability
- Chain commands with pipes (|) or operators (&&, ||) when needed
- Use variable substitution or subshells when appropriate

Examples:
User: "List all text files in the current directory"
Command: find . -type f -name "*.txt"

User: "Check system memory usage"
Command: free -h

User: "Create a compressed backup of my documents"
Command: tar -czvf backup_$(date +%Y%m%d).tar.gz ~/Documents

This is a pure command generation mode - the interface will handle execution and safety confirmations.`,
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	// Start the loading spinner
	spinnerManager := util.GetGlobalSpinnerManager()
	spinnerManager.StartThinkingSpinner("Generating shell command...")

	startTime := time.Now()
	response, metrics, err := adapter.ChatWithMetrics(context.Background(), messages)
	elapsedTime := time.Since(startTime)

	// Stop the spinner
	spinnerManager.TransitionToResponse()

	if err != nil {
		return fmt.Errorf("failed to get response: %w", err)
	}

	// Display the command in green, without a box
	util.OutputColor.Println(response)
	fmt.Println() // Add space after command

	// Print statistics only when debug mode is enabled
	if cfg.UI.Debug {
		util.StatsColor.Printf("[%s] Tokens: %d prompt + %d completion = %d total | Time: %.2fs\n",
			adapter.GetModel(),
			metrics.PromptTokens,
			metrics.CompletionTokens,
			metrics.TotalTokens,
			elapsedTime.Seconds())
		fmt.Println() // Add space after debug info
	}

	if cfg.LLM.SafeMode {
		// Use single-key confirmation
		confirmed, err := util.PromptForConfirmation("Do you want to execute this command? (y/N): ")
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		if !confirmed {
			return nil
		}
	}

	// Add a line break before output header
	fmt.Println()
	util.SuccessColor.Println("Output:")

	shellTool := tools.NewShellTool()
	result, err := shellTool.Execute(context.Background(), map[string]interface{}{
		"command": response,
	})
	if err != nil {
		return fmt.Errorf("failed to execute command: %w", err)
	}

	fmt.Println(result)
	return nil
}
