package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	LLM    LLMConfig    `yaml:"llm"`
	Shell  ShellConfig  `yaml:"shell"`
	Output OutputConfig `yaml:"output"`
	Debug  DebugConfig  `yaml:"debug"`
}

// LLMConfig represents LLM configuration
type LLMConfig struct {
	Provider    string  `yaml:"provider"`
	APIKey      string  `yaml:"api_key"`
	BaseURL     string  `yaml:"base_url"`
	Model       string  `yaml:"model"`
	MaxTokens   int     `yaml:"max_tokens"`
	Temperature float64 `yaml:"temperature"`
	Timeout     int     `yaml:"timeout"`
}

// ShellConfig represents shell configuration
type ShellConfig struct {
	HistoryFile     string `yaml:"history_file"`
	AutoInstallHook bool   `yaml:"auto_install_hook"`
}

// OutputConfig represents output configuration
type OutputConfig struct {
	Color    bool   `yaml:"color"`
	Markdown bool   `yaml:"markdown"`
	Language string `yaml:"language"`
}

// DebugConfig represents debug configuration
type DebugConfig struct {
	Enabled    bool `yaml:"enabled"`
	ShowPrompt bool `yaml:"show_prompt"`
	ShowTokens bool `yaml:"show_tokens"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		LLM: LLMConfig{
			Provider:    "openai",
			Model:       "gpt-4o-mini",
			MaxTokens:   4096,
			Temperature: 0.7,
			Timeout:     60,
		},
		Shell: ShellConfig{
			AutoInstallHook: true,
		},
		Output: OutputConfig{
			Color:    true,
			Markdown: true,
			Language: "en-US",
		},
		Debug: DebugConfig{
			Enabled: false,
		},
	}
}

// Load loads the configuration from file
func Load() (*Config, error) {
	configPath := GetConfigPath()

	// If config file doesn't exist, return default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Environment variable overrides
	if apiKey := os.Getenv("OHMAN_API_KEY"); apiKey != "" {
		cfg.LLM.APIKey = apiKey
	}

	return cfg, nil
}

// Save saves the configuration to file
func Save(cfg *Config) error {
	configPath := GetConfigPath()
	configDir := filepath.Dir(configPath)

	// Create config directory
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	// Write config file with 600 permissions (user read/write only)
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetConfigPath returns the config file path
func GetConfigPath() string {
	if path := os.Getenv("OHMAN_CONFIG"); path != "" {
		return path
	}

	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "ohman", "config.yaml")
}

// InteractiveSetup runs the interactive configuration wizard
func InteractiveSetup() error {
	reader := bufio.NewReader(os.Stdin)
	cfg := DefaultConfig()

	fmt.Println("🔧 Oh Man! Configuration")
	fmt.Println("========================")
	fmt.Println()

	// API Base URL
	fmt.Print("API Base URL (e.g., https://api.openai.com/v1): ")
	baseURL, _ := reader.ReadString('\n')
	cfg.LLM.BaseURL = strings.TrimSpace(baseURL)

	// API Key
	fmt.Print("API Key: ")
	apiKey, _ := reader.ReadString('\n')
	cfg.LLM.APIKey = strings.TrimSpace(apiKey)

	// Model
	fmt.Print("Model name [gpt-4o-mini]: ")
	model, _ := reader.ReadString('\n')
	model = strings.TrimSpace(model)
	if model != "" {
		cfg.LLM.Model = model
	}

	// Save configuration
	if err := Save(cfg); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("✅ Saved to:", GetConfigPath())
	fmt.Println()
	fmt.Println("Try: ohman grep \"how to search recursively?\"")

	return nil
}
