package cli

import (
	"github.com/spf13/cobra"
)

var (
	// LLM configuration flags
	provider  string
	model     string
	apiKey    string
	safeMode  bool
	debug     bool // Debug mode flag
	streaming bool // Streaming mode flag

	// Command declarations
	assistantCmd *cobra.Command
	shellCmd     *cobra.Command
	configCmd    *cobra.Command

	// Shorthand flags
	assistantFlag bool
	terminalFlag  string
	configFlag    bool
	configGet     string
	configSet     []string

	// Root command declaration
	rootCmd *cobra.Command
)

func init() {
	// Initialize root command
	rootCmd = &cobra.Command{
		Use:   "kiwi [prompt]",
		Short: "Kiwi - A CLI tool for interacting with LLMs directly from your terminal",
		Long: `Kiwi is a CLI tool that helps you interact with Large Language Models (LLMs).
It supports multiple LLM providers and provides various tools for enhanced functionality.

When run without a command but with arguments, Kiwi treats the arguments as a prompt for the execute mode.

Examples:
  # Execute a prompt directly (no quotes needed)
  kiwi what is the capital of France

  # Terminal command assistance
  kiwi -t list all files in this directory

  # Start a new assistant session (chat)
  kiwi -a

  # Configuration
  kiwi -c
  kiwi -c get llm.provider
  kiwi -c set llm.provider openai

Configuration:
  The tool can be configured using:
  - Environment variables (KIWI_PROVIDER, KIWI_MODEL, KIWI_API_KEY)
  - Command line flags (--provider, --model, --api-key)
  - Config file (~/.kiwi/config.yaml)
  - Config commands (kiwi -c set)`,
		// Allow arbitrary args to support default execute mode
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no command is provided but shorthand flags are, handle them
			switch {
			case assistantFlag:
				return startAssistant(cmd, args)
			case terminalFlag != "":
				return handleTerminalHelp(cmd, terminalFlag)
			case configFlag:
				if len(configGet) > 0 {
					return handleConfigGet(cmd, []string{configGet})
				} else if len(configSet) == 2 {
					return handleConfigSet(cmd, configSet)
				} else {
					return handleConfigList(cmd, args)
				}
			default:
				// If arguments are provided, treat them as an execute command
				if len(args) > 0 {
					// Join all arguments into a single prompt string
					prompt := args[0]
					for i := 1; i < len(args); i++ {
						prompt += " " + args[i]
					}
					return handleExecute(cmd, prompt)
				}
				// Display help if no command, flag, or args are provided
				return cmd.Help()
			}
		},
	}

	// LLM configuration flags
	rootCmd.PersistentFlags().StringVar(&provider, "provider", "openai", "LLM provider (currently only openai)")
	rootCmd.PersistentFlags().StringVar(&model, "model", "gpt-3.5-turbo", "LLM model to use")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "API key for the LLM provider")
	rootCmd.PersistentFlags().BoolVar(&safeMode, "safe-mode", true, "Enable safe mode with command confirmation")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug mode with verbose output and statistics")
	rootCmd.PersistentFlags().BoolVar(&streaming, "streaming", true, "Enable streaming mode for incremental response display")

	// Shorthand command flags
	rootCmd.Flags().BoolVarP(&assistantFlag, "assistant", "a", false, "Start a new assistant session (interactive chat)")
	rootCmd.Flags().StringVarP(&terminalFlag, "terminal", "t", "", "Get help with a terminal command")

	// Config flags
	rootCmd.Flags().BoolVarP(&configFlag, "config", "c", false, "Manage configuration (defaults to list)")
	rootCmd.Flags().StringVar(&configGet, "get", "", "Get a specific configuration value")
	rootCmd.Flags().StringSliceVar(&configSet, "set", []string{}, "Set a configuration value (key and value)")

	// Initialize commands
	setupCommands()
}

// Setup all commands
func setupCommands() {
	// Initialize commands
	initAssistantCmd()
	initShellCmd()
	initConfigCmd()

	// Add commands to root
	rootCmd.AddCommand(assistantCmd)
	rootCmd.AddCommand(shellCmd)
	rootCmd.AddCommand(configCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
