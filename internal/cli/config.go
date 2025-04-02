package cli

import (
	"fmt"
	"strings"

	"github.com/saurabh0719/kiwi/internal/config"
	"github.com/spf13/cobra"
)

var (
	configCmd          *cobra.Command
	listCmd            *cobra.Command
	getCmd             *cobra.Command
	setCmd             *cobra.Command
	configSetOpenAICmd *cobra.Command
	configSetSerperCmd *cobra.Command
	configListCmd      *cobra.Command
)

func initConfigCmd() {
	configCmd = &cobra.Command{
		Use:   "config",
		Short: "Manage Kiwi configuration",
		Long: `Manage Kiwi configuration settings.

Examples:
  # List all config settings
  kiwi config
  kiwi config list

  # Get a specific config value
  kiwi config get llm.provider
  kiwi config get ui.debug

  # Set a config value
  kiwi config set llm.provider openai
  kiwi config set llm.model gpt-4
  kiwi config set llm.api_key your_api_key
  kiwi config set llm.safe_mode true
  kiwi config set ui.debug true
  kiwi config set ui.streaming true`,
		// Run list command by default when no subcommand is specified
		RunE: handleConfigList,
	}

	// List command
	listCmd = &cobra.Command{
		Use:   "list",
		Short: "List all configuration settings",
		Long:  "Display all current configuration settings",
		Args:  cobra.NoArgs,
		RunE:  handleConfigList,
	}

	// Get command
	getCmd = &cobra.Command{
		Use:   "get [key]",
		Short: "Get a configuration value",
		Long:  "Get the value of a specific configuration key",
		Args:  cobra.ExactArgs(1),
		RunE:  handleConfigGet,
	}

	// Set command
	setCmd = &cobra.Command{
		Use:   "set [key] [value]",
		Short: "Set a configuration value",
		Long:  "Set a specific configuration key to the given value",
		Args:  cobra.ExactArgs(2),
		RunE:  handleConfigSet,
	}

	// Set OpenAI API key command
	configSetOpenAICmd = &cobra.Command{
		Use:   "set-openai-key",
		Short: "Set OpenAI API key",
		Long:  "Set OpenAI API key for Kiwi",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return configSetOpenAIKey(cmd, args)
		},
	}

	// Set Serper API key command
	configSetSerperCmd = &cobra.Command{
		Use:   "set-serper-key",
		Short: "Set Serper API key for web search",
		Long:  "Set Serper API key to enable web search functionality in Kiwi",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return configSetSerperKey(cmd, args)
		},
	}

	// Add subcommands to config
	configCmd.AddCommand(listCmd)
	configCmd.AddCommand(getCmd)
	configCmd.AddCommand(setCmd)
	configCmd.AddCommand(configSetOpenAICmd)
	configCmd.AddCommand(configSetSerperCmd)
}

func handleConfigList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(rootCmd)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("Current configuration:")
	fmt.Printf("  llm.provider: %s\n", cfg.LLM.Provider)
	fmt.Printf("  llm.model: %s\n", cfg.LLM.Model)

	// Only show API key if present, but mask it
	if cfg.LLM.APIKey != "" {
		maskedKey := maskString(cfg.LLM.APIKey)
		fmt.Printf("  llm.api_key: %s\n", maskedKey)
	} else {
		fmt.Printf("  llm.api_key: <not set>\n")
	}

	fmt.Printf("  llm.safe_mode: %t\n", cfg.LLM.SafeMode)

	// Show options if there are any
	if len(cfg.LLM.Options) > 0 {
		fmt.Println("  llm.options:")
		for k, v := range cfg.LLM.Options {
			fmt.Printf("    %s: %s\n", k, v)
		}
	}

	fmt.Printf("  ui.debug: %t\n", cfg.UI.Debug)
	fmt.Printf("  ui.streaming: %t\n", cfg.UI.Streaming)

	return nil
}

func handleConfigGet(cmd *cobra.Command, args []string) error {
	key := args[0]
	cfg, err := config.Load(rootCmd)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	switch key {
	case "llm.provider":
		fmt.Println(cfg.LLM.Provider)
	case "llm.model":
		fmt.Println(cfg.LLM.Model)
	case "llm.api_key":
		if cfg.LLM.APIKey == "" {
			fmt.Println("<not set>")
		} else {
			maskedKey := maskString(cfg.LLM.APIKey)
			fmt.Println(maskedKey)
		}
	case "llm.safe_mode":
		fmt.Println(cfg.LLM.SafeMode)
	case "ui.debug":
		fmt.Println(cfg.UI.Debug)
	case "ui.streaming":
		fmt.Println(cfg.UI.Streaming)
	default:
		// Check if it's an option
		if strings.HasPrefix(key, "llm.options.") {
			optKey := strings.TrimPrefix(key, "llm.options.")
			if val, ok := cfg.LLM.Options[optKey]; ok {
				fmt.Println(val)
			} else {
				fmt.Println("<not set>")
			}
		} else {
			return fmt.Errorf("unknown config key: %s", key)
		}
	}

	return nil
}

func handleConfigSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	cfg, err := config.Load(rootCmd)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	switch key {
	case "llm.provider":
		if value != "openai" {
			return fmt.Errorf("provider must be 'openai' (Claude support will be added in the future)")
		}
		cfg.LLM.Provider = value
	case "llm.model":
		cfg.LLM.Model = value
	case "llm.api_key":
		cfg.LLM.APIKey = value
	case "llm.safe_mode":
		if value == "true" {
			cfg.LLM.SafeMode = true
		} else if value == "false" {
			cfg.LLM.SafeMode = false
		} else {
			return fmt.Errorf("safe_mode must be 'true' or 'false'")
		}
	case "ui.debug":
		if value == "true" {
			cfg.UI.Debug = true
		} else if value == "false" {
			cfg.UI.Debug = false
		} else {
			return fmt.Errorf("debug must be 'true' or 'false'")
		}
	case "ui.streaming":
		if value == "true" {
			cfg.UI.Streaming = true
		} else if value == "false" {
			cfg.UI.Streaming = false
		} else {
			return fmt.Errorf("streaming must be 'true' or 'false'")
		}
	default:
		// Check if it's an option
		if strings.HasPrefix(key, "llm.options.") {
			optKey := strings.TrimPrefix(key, "llm.options.")
			if cfg.LLM.Options == nil {
				cfg.LLM.Options = make(map[string]string)
			}
			cfg.LLM.Options[optKey] = value
		} else {
			return fmt.Errorf("unknown config key: %s", key)
		}
	}

	// Save the updated config
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Config updated: %s = %s\n", key, value)
	return nil
}

// Mask a string (like an API key) for display
func maskString(input string) string {
	if len(input) <= 8 {
		return "****"
	}

	// Show the first 4 and last 4 characters, mask the rest
	return input[:4] + "..." + input[len(input)-4:]
}

// configSetOpenAIKey sets the OpenAI API key
func configSetOpenAIKey(cmd *cobra.Command, args []string) error {
	// Load existing config
	cfg, err := config.Load(cmd)
	if err != nil {
		return fmt.Errorf("Failed to load config: %v", err)
	}

	// Set the API key
	cfg.LLM.APIKey = args[0]

	// Save the config
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("Failed to save config: %v", err)
	}

	fmt.Println("OpenAI API key set successfully.")
	return nil
}

// configSetSerperKey sets the Serper API key
func configSetSerperKey(cmd *cobra.Command, args []string) error {
	// Load existing config
	cfg, err := config.Load(cmd)
	if err != nil {
		return fmt.Errorf("Failed to load config: %v", err)
	}

	// Set the API key
	cfg.Tools.SerperAPIKey = args[0]

	// Save the config
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("Failed to save config: %v", err)
	}

	fmt.Println("Serper API key set successfully. Web search is now available.")
	return nil
}

// configList lists all current configuration settings
func configList(cmd *cobra.Command, args []string) error {
	// Load existing config
	cfg, err := config.Load(cmd)
	if err != nil {
		return fmt.Errorf("Failed to load config: %v", err)
	}

	// Print configuration values
	fmt.Println("Current Configuration:")
	fmt.Println("----------------------")
	fmt.Printf("LLM Provider: %s\n", cfg.LLM.Provider)
	fmt.Printf("LLM Model: %s\n", cfg.LLM.Model)

	// Show API keys as masked values for security
	if cfg.LLM.APIKey != "" {
		fmt.Printf("OpenAI API Key: %s\n", maskAPIKey(cfg.LLM.APIKey))
	} else {
		fmt.Printf("OpenAI API Key: [not set]\n")
	}

	if cfg.Tools.SerperAPIKey != "" {
		fmt.Printf("Serper API Key: %s\n", maskAPIKey(cfg.Tools.SerperAPIKey))
		fmt.Println("Web Search: Enabled")
	} else {
		fmt.Printf("Serper API Key: [not set]\n")
		fmt.Println("Web Search: Disabled")
	}

	fmt.Printf("Debug Mode: %v\n", cfg.UI.Debug)
	fmt.Printf("Streaming Mode: %v\n", cfg.UI.Streaming)

	return nil
}

// maskAPIKey masks an API key for display, showing only first and last 4 characters
func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return strings.Repeat("*", len(key))
	}
	return key[:4] + strings.Repeat("*", len(key)-8) + key[len(key)-4:]
}
