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
	debug    bool // Debug mode flag

	// Command declarations
	chatCmd    *cobra.Command
	shellCmd   *cobra.Command
	executeCmd *cobra.Command
	configCmd  *cobra.Command

	// Shorthand flags
	executeFlag string
	chatFlag    bool
	shellFlag   string
)

var rootCmd = &cobra.Command{
	Use:   "kiwi",
	Short: "Kiwi - A CLI tool for interacting with LLMs",
	Long: `Kiwi is a CLI tool that helps you interact with Large Language Models (LLMs).
It supports multiple LLM providers and provides various tools for enhanced functionality.

Examples:
  # Start a new chat session
  kiwi chat
  kiwi -c

  # Get help with a shell command
  kiwi shell "list all files in this directory"
  kiwi -s "list all files in this directory"

  # Execute a prompt
  kiwi execute "What is the capital of France?"
  kiwi -e "What is the capital of France?"

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
	RunE: func(cmd *cobra.Command, args []string) error {
		// If no command is provided but shorthand flags are, handle them
		switch {
		case executeFlag != "":
			return handleExecute(cmd, executeFlag)
		case chatFlag:
			return startNewChat(cmd, args)
		case shellFlag != "":
			return handleShellHelp(cmd, shellFlag)
		default:
			// Display help if no command or flag is provided
			return cmd.Help()
		}
	},
}

func init() {
	// LLM configuration flags
	rootCmd.PersistentFlags().StringVar(&provider, "provider", "openai", "LLM provider (openai, claude)")
	rootCmd.PersistentFlags().StringVar(&model, "model", "gpt-3.5-turbo", "LLM model to use")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "API key for the LLM provider")
	rootCmd.PersistentFlags().BoolVar(&safeMode, "safe-mode", true, "Enable safe mode with command confirmation")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug mode with verbose output and statistics")

	// Shorthand command flags
	rootCmd.Flags().StringVarP(&executeFlag, "execute", "e", "", "Execute a prompt (shorthand)")
	rootCmd.Flags().BoolVarP(&chatFlag, "chat", "c", false, "Start a new chat session (shorthand)")
	rootCmd.Flags().StringVarP(&shellFlag, "shell", "s", "", "Get help with a shell command (shorthand)")

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

	// Add shorthand commands at the root level
	rootCmd.AddCommand(&cobra.Command{
		Use:   "e",
		Short: "Shorthand for execute",
		Long:  `Shorthand for the execute command. Execute a prompt and get a response from the LLM.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleExecute(cmd, args[0])
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "c",
		Short: "Shorthand for chat",
		Long:  `Shorthand for the chat command. Start a new chat session with the LLM.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return startNewChat(cmd, args)
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "s",
		Short: "Shorthand for shell",
		Long:  `Shorthand for the shell command. Get help with a shell command.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleShellHelp(cmd, args[0])
		},
	})
}

func Execute() error {
	return rootCmd.Execute()
}
