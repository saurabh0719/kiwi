package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// LLMConfig represents configuration for a specific LLM provider
type LLMConfig struct {
	Provider string            `mapstructure:"provider"`
	Model    string            `mapstructure:"model"`
	APIKey   string            `mapstructure:"api_key"`
	Options  map[string]string `mapstructure:"options"`
	SafeMode bool              `mapstructure:"safe_mode"`
}

// UIConfig represents UI and display settings
type UIConfig struct {
	Debug     bool `mapstructure:"debug"`
	Streaming bool `mapstructure:"streaming"`
}

// ToolsConfig represents configuration for various tools
type ToolsConfig struct {
	SerperAPIKey string `mapstructure:"serper_api_key"`
}

// Config represents the overall application configuration
type Config struct {
	LLM   LLMConfig   `mapstructure:"llm"`
	UI    UIConfig    `mapstructure:"ui"`
	Tools ToolsConfig `mapstructure:"tools"`
}

func getConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return filepath.Join(homeDir, ".kiwi"), nil
}

// Load loads configuration from various sources
func Load(rootCmd *cobra.Command) (*Config, error) {
	// Set up viper
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// Get OS-specific config directory
	configDir, err := getConfigDir()
	if err != nil {
		return nil, err
	}
	v.AddConfigPath(configDir)

	// Set defaults
	v.SetDefault("llm.provider", "openai")
	v.SetDefault("llm.model", "gpt-3.5-turbo")
	v.SetDefault("llm.options", map[string]string{})
	v.SetDefault("llm.safe_mode", true)

	// Set UI defaults
	v.SetDefault("ui.debug", false)
	v.SetDefault("ui.streaming", true)

	// Set Tools defaults
	v.SetDefault("tools.serper_api_key", "")

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Read from environment
	v.SetEnvPrefix("KIWI")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Map environment variables for tools
	if serperKey := os.Getenv("SERPER_API_KEY"); serperKey != "" {
		v.Set("tools.serper_api_key", serperKey)
	}

	// Read from flags
	if err := v.BindPFlag("llm.provider", rootCmd.Flags().Lookup("provider")); err != nil {
		return nil, fmt.Errorf("failed to bind provider flag: %w", err)
	}
	if err := v.BindPFlag("llm.model", rootCmd.Flags().Lookup("model")); err != nil {
		return nil, fmt.Errorf("failed to bind model flag: %w", err)
	}
	if err := v.BindPFlag("llm.api_key", rootCmd.Flags().Lookup("api-key")); err != nil {
		return nil, fmt.Errorf("failed to bind api-key flag: %w", err)
	}
	if err := v.BindPFlag("llm.safe_mode", rootCmd.Flags().Lookup("safe-mode")); err != nil {
		return nil, fmt.Errorf("failed to bind safe-mode flag: %w", err)
	}
	if err := v.BindPFlag("ui.debug", rootCmd.Flags().Lookup("debug")); err != nil {
		return nil, fmt.Errorf("failed to bind debug flag: %w", err)
	}
	if err := v.BindPFlag("ui.streaming", rootCmd.Flags().Lookup("streaming")); err != nil {
		return nil, fmt.Errorf("failed to bind streaming flag: %w", err)
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// Save saves the current configuration to disk
func (c *Config) Save() error {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// Get OS-specific config directory
	configDir, err := getConfigDir()
	if err != nil {
		return err
	}
	v.AddConfigPath(configDir)

	// Set values
	v.Set("llm.provider", c.LLM.Provider)
	v.Set("llm.model", c.LLM.Model)
	v.Set("llm.api_key", c.LLM.APIKey)
	v.Set("llm.options", c.LLM.Options)
	v.Set("llm.safe_mode", c.LLM.SafeMode)

	// Set UI values
	v.Set("ui.debug", c.UI.Debug)
	v.Set("ui.streaming", c.UI.Streaming)

	// Set Tools values
	v.Set("tools.serper_api_key", c.Tools.SerperAPIKey)

	// Write to file
	configPath := filepath.Join(configDir, "config.yaml")
	return v.WriteConfigAs(configPath)
}

// GetToolsConfig returns a map of tool configuration values
func (c *Config) GetToolsConfig() map[string]string {
	return map[string]string{
		"SERPER_API_KEY": c.Tools.SerperAPIKey,
	}
}
