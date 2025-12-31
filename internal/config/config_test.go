package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	if cfg.LLM.Provider != "openai" {
		t.Errorf("expected provider 'openai', got '%s'", cfg.LLM.Provider)
	}

	if cfg.LLM.Model != "gpt-4o-mini" {
		t.Errorf("expected model 'gpt-4o-mini', got '%s'", cfg.LLM.Model)
	}

	if cfg.LLM.MaxTokens != 4096 {
		t.Errorf("expected max_tokens 4096, got %d", cfg.LLM.MaxTokens)
	}

	if !cfg.Output.Color {
		t.Error("expected color to be true")
	}

	if !cfg.Output.Markdown {
		t.Error("expected markdown to be true")
	}
}

func TestLoadNonExistent(t *testing.T) {
	// Set a non-existent config path
	_ = os.Setenv("OHMAN_CONFIG", "/tmp/nonexistent_ohman_config_12345.yaml")
	defer func() { _ = os.Unsetenv("OHMAN_CONFIG") }()

	cfg, err := Load()
	if err != nil {
		t.Errorf("Load() should return default config for non-existent file, got error: %v", err)
	}

	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}

	// Should return default config
	if cfg.LLM.Provider != "openai" {
		t.Errorf("expected default provider 'openai', got '%s'", cfg.LLM.Provider)
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	_ = os.Setenv("OHMAN_CONFIG", configPath)
	defer func() { _ = os.Unsetenv("OHMAN_CONFIG") }()

	// Create config
	cfg := &Config{
		LLM: LLMConfig{
			Provider:    "anthropic",
			APIKey:      "test-key",
			Model:       "claude-3",
			MaxTokens:   2048,
			Temperature: 0.5,
			Timeout:     30,
		},
		Output: OutputConfig{
			Color:    true,
			Markdown: true,
			Language: "en-US",
		},
	}

	// Save
	if err := Save(cfg); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("config file was not created")
	}

	// Load and verify
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if loaded.LLM.Provider != cfg.LLM.Provider {
		t.Errorf("provider mismatch: got %s, want %s", loaded.LLM.Provider, cfg.LLM.Provider)
	}

	if loaded.LLM.Model != cfg.LLM.Model {
		t.Errorf("model mismatch: got %s, want %s", loaded.LLM.Model, cfg.LLM.Model)
	}
}

func TestGetConfigPath(t *testing.T) {
	// Test environment variable override
	testPath := "/custom/path/config.yaml"
	_ = os.Setenv("OHMAN_CONFIG", testPath)
	defer func() { _ = os.Unsetenv("OHMAN_CONFIG") }()

	if got := GetConfigPath(); got != testPath {
		t.Errorf("GetConfigPath() = %s, want %s", got, testPath)
	}
}

func TestEnvironmentVariableOverride(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	_ = os.Setenv("OHMAN_CONFIG", configPath)
	defer func() { _ = os.Unsetenv("OHMAN_CONFIG") }()

	// Save config without API Key
	cfg := DefaultConfig()
	cfg.LLM.APIKey = ""
	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}

	// Set environment variable API Key
	_ = os.Setenv("OHMAN_API_KEY", "env-api-key")
	defer func() { _ = os.Unsetenv("OHMAN_API_KEY") }()

	// Load should use environment variable API Key
	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	if loaded.LLM.APIKey != "env-api-key" {
		t.Errorf("API key should be overridden by env var, got %s", loaded.LLM.APIKey)
	}
}
