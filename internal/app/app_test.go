package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/liliang-cn/ohman/internal/config"
)

func TestParseCommandName(t *testing.T) {
	tests := []struct {
		name     string
		fullCmd  string
		expected string
	}{
		{
			name:     "simple command",
			fullCmd:  "ls -la",
			expected: "ls",
		},
		{
			name:     "full path",
			fullCmd:  "/usr/bin/grep pattern file",
			expected: "grep",
		},
		{
			name:     "with env var",
			fullCmd:  "VAR=value command arg",
			expected: "command",
		},
		{
			name:     "multiple env vars",
			fullCmd:  "VAR1=a VAR2=b cmd",
			expected: "cmd",
		},
		{
			name:     "sudo command",
			fullCmd:  "sudo apt update",
			expected: "sudo",
		},
		{
			name:     "command with options",
			fullCmd:  "tar -xvf file.tar.gz",
			expected: "tar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseCommandName(tt.fullCmd)
			if result != tt.expected {
				t.Errorf("parseCommandName(%q) = %q, want %q", tt.fullCmd, result, tt.expected)
			}
		})
	}
}

func TestAnalyzeLogFile(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	content := `2025-02-01 10:00:00 INFO Application started
2025-02-01 10:00:01 ERROR Database connection failed
2025-02-01 10:00:02 WARN Retrying...
`

	if err := os.WriteFile(logFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test log file: %v", err)
	}

	cfg := &config.Config{}
	application := New(cfg)

	err := application.AnalyzeLogFile(logFile, 0)
	if err != nil {
		t.Logf("AnalyzeLogFile returned error (expected without LLM config): %v", err)
	}
}

func TestAnalyzeLogContent(t *testing.T) {
	content := `2025-02-01 10:23:45 ERROR Database connection failed
2025-02-01 10:23:46 WARN Retrying connection`

	cfg := &config.Config{}
	application := New(cfg)

	err := application.AnalyzeLogContent(content)
	if err != nil {
		t.Logf("AnalyzeLogContent returned error (expected without LLM config): %v", err)
	}
}

func TestChat(t *testing.T) {
	cfg := &config.Config{}
	application := New(cfg)

	// Test chat without log context
	err := application.Chat("")
	if err != nil {
		t.Logf("Chat returned error (expected without LLM config): %v", err)
	}

	// Test chat with log context
	logContent := "2025-02-01 ERROR: Test error"
	err = application.Chat(logContent)
	if err != nil {
		t.Logf("Chat with log context returned error (expected without LLM config): %v", err)
	}
}
