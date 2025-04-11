package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// LLM configuration flags
	provider       string
	model          string
	apiKey         string
	safeMode       bool
	debug          bool   // Debug mode flag
	streaming      bool   // Streaming mode flag
	configPath     string // Path to config file
	renderMarkdown bool   // Markdown rendering flag

	// Command declarations
	assistantCmd *cobra.Command
	configCmd    *cobra.Command
	// sessionsCmd is already declared in sessions.go

	// Root command declaration
	rootCmd *cobra.Command

	// Shorthand command flags for root
	assistantFlag bool // Start assistant session flag
	configFlag    bool // Config management flag
	sessionsFlag  bool // Sessions flag - duplicated here for root command use
	clearFlag     bool // Clear sessions flag - duplicated here for root command use
	// Other session flags are defined in sessions.go

	// Config flags
	configGet string   // Config key to get
	configSet []string // Config key-value pair to set
)

func init() {
	// Initialize commands first so we can use them
	initSessionsCmd()
	initAssistantCmd()
	initConfigCmd()

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

  # Execute shell commands
  kiwi create a shell script to backup my Documents folder

  # Start a new assistant session (chat)
  kiwi -a

  # Manage assistant sessions
  kiwi -s            # List all sessions
  kiwi -s -l         # Also lists all sessions
  kiwi -s -n my_project    # Create a new named session
  kiwi -s -o 1234567       # Continue a session by ID
  kiwi -s -d 1234567       # Delete a session by ID
  kiwi -s -C                    # Clear all sessions

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
			// Directly handle common high-level flags
			switch {
			case assistantFlag:
				return startAssistant(cmd, args)
			case configFlag:
				if len(configGet) > 0 {
					return handleConfigGet(cmd, []string{configGet})
				} else if len(configSet) == 2 {
					return handleConfigSet(cmd, configSet)
				} else {
					return handleConfigList(cmd, args)
				}
			case sessionsFlag && clearFlag:
				// Handle clear sessions directly
				return handleSessionsClear(cmd, args)
			case sessionsFlag && continueFlag:
				// Handle continue session directly with any provided ID
				if len(args) > 0 {
					return handleSessionsContinue(cmd, args)
				}
				return fmt.Errorf("session ID required with -c flag")
			case sessionsFlag && deleteFlag:
				// Handle delete session directly with any provided ID
				if len(args) > 0 {
					return handleSessionsDelete(cmd, args)
				}
				return fmt.Errorf("session ID required with -d flag")
			case sessionsFlag && newFlag:
				// Handle new session directly with any provided name
				if len(args) > 0 {
					return handleSessionsNew(cmd, args)
				}
				return fmt.Errorf("session name required with -n flag")
			case sessionsFlag:
				// If no specific flags, default to list sessions
				return handleSessionsList(cmd, args)
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
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "D", false, "Enable debug mode with verbose output and statistics")
	rootCmd.PersistentFlags().BoolVarP(&streaming, "streaming", "S", true, "Enable streaming mode for incremental response display")
	rootCmd.PersistentFlags().StringVarP(&configPath, "config-path", "p", "", "Path to config file")
	rootCmd.PersistentFlags().BoolVarP(&renderMarkdown, "render-markdown", "r", false, "Enable Markdown rendering for output")

	// Shorthand command flags
	rootCmd.Flags().BoolVarP(&assistantFlag, "assistant", "a", false, "Start a new assistant session (interactive chat)")
	rootCmd.Flags().BoolVarP(&sessionsFlag, "sessions", "s", false, "Manage assistant sessions")
	rootCmd.Flags().BoolVarP(&clearFlag, "clear", "C", false, "Clear all sessions (use with -s)")
	rootCmd.Flags().BoolVarP(&configFlag, "configure", "c", false, "Manage configuration (defaults to list)")

	// Register session flags with the root command
	rootCmd.Flags().BoolVarP(&continueFlag, "continue", "o", false, "Continue a session (requires ID as argument)")
	rootCmd.Flags().BoolVarP(&deleteFlag, "delete", "d", false, "Delete a session (requires ID as argument)")
	rootCmd.Flags().BoolVarP(&newFlag, "new", "n", false, "Create a new session (requires name as argument)")
	rootCmd.Flags().StringVar(&sessionID, "id", "", "Session ID (alternative to providing as argument)")

	// Config flags
	rootCmd.Flags().StringVar(&configGet, "get", "", "Get a specific configuration value")
	rootCmd.Flags().StringSliceVar(&configSet, "set", []string{}, "Set a configuration value (key and value)")

	// Add commands to root
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(sessionsCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
