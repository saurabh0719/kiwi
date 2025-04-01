package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/saurabh0719/kiwi/internal/config"
	"github.com/saurabh0719/kiwi/internal/llm"
	"github.com/saurabh0719/kiwi/internal/tools"
	"github.com/saurabh0719/kiwi/internal/util"
	"github.com/spf13/cobra"
)

var commandColor = color.New(color.FgGreen)

// statsColor is declared in execute.go

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
	spinner := util.NewSpinner("Generating shell command...")
	spinner.Start()

	startTime := time.Now()
	response, metrics, err := adapter.ChatWithMetrics(context.Background(), messages)
	elapsedTime := time.Since(startTime)

	// Stop the spinner
	spinner.Stop()

	if err != nil {
		return fmt.Errorf("failed to get response: %w", err)
	}

	fmt.Println()
	commandColor.Println(response)

	// Print statistics in blue only when debug mode is enabled
	if cfg.UI.Debug {
		statsColor.Printf("\n[%s] Tokens: %d prompt + %d completion = %d total | Time: %.2fs\n",
			adapter.GetModel(),
			metrics.PromptTokens,
			metrics.CompletionTokens,
			metrics.TotalTokens,
			elapsedTime.Seconds())
	}

	if cfg.LLM.SafeMode {
		fmt.Print("\nDo you want to execute this command? (y/n): ")
		reader := bufio.NewReader(os.Stdin)
		answer, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" {
			return nil
		}
	}

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
