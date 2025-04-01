package cli

import (
	"github.com/spf13/cobra"
)

var (
	// LLM configuration flags
	provider string
	model    string
	apiKey   string
	safeMode bool

	// Command declarations
	chatCmd    *cobra.Command
	shellCmd   *cobra.Command
	executeCmd *cobra.Command
	configCmd  *cobra.Command
)

var rootCmd = &cobra.Command{
	Use:   "kiwi",
	Short: "Kiwi - A CLI tool for interacting with LLMs",
	Long: `Kiwi is a CLI tool that helps you interact with Large Language Models (LLMs).
It supports multiple LLM providers and provides various tools for enhanced functionality.

Examples:
  # Start a new chat session
  kiwi chat

  # Get help with a shell command
  kiwi shell "list all files in this directory"

  # Execute a prompt
  kiwi execute "What is the capital of France?"

  # Manage configuration
  kiwi config list
  kiwi config get llm.provider
  kiwi config set llm.model gpt-4

Configuration:
  The tool can be configured using:
  - Environment variables (KIWI_PROVIDER, KIWI_MODEL, KIWI_API_KEY)
  - Command line flags (--provider, --model, --api-key)
  - Config file (~/.kiwi/config.yaml)
  - Config commands (kiwi config set)`,
}

func init() {
	// LLM configuration flags
	rootCmd.PersistentFlags().StringVar(&provider, "provider", "openai", "LLM provider (openai, claude)")
	rootCmd.PersistentFlags().StringVar(&model, "model", "gpt-3.5-turbo", "LLM model to use")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "API key for the LLM provider")
	rootCmd.PersistentFlags().BoolVar(&safeMode, "safe-mode", true, "Enable safe mode with command confirmation")

	// Initialize commands
	initChatCmd()
	initShellCmd()
	initExecuteCmd()
	initConfigCmd()

	// Add commands to root
	rootCmd.AddCommand(chatCmd)
	rootCmd.AddCommand(shellCmd)
	rootCmd.AddCommand(executeCmd)
	rootCmd.AddCommand(configCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
