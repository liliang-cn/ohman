package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDetectShell(t *testing.T) {
	tests := []struct {
		name     string
		shellEnv string
		want     string
	}{
		{"zsh", "/bin/zsh", "zsh"},
		{"bash", "/bin/bash", "bash"},
		{"fish", "/usr/bin/fish", "fish"},
		{"unknown", "/bin/sh", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldShell := os.Getenv("SHELL")
			_ = os.Setenv("SHELL", tt.shellEnv)
			defer func() { _ = os.Setenv("SHELL", oldShell) }()

			if got := DetectShell(); got != tt.want {
				t.Errorf("DetectShell() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetShellHookScript(t *testing.T) {
	tests := []struct {
		shellType string
		wantEmpty bool
	}{
		{"zsh", false},
		{"bash", false},
		{"fish", true}, // Not supported yet
		{"unknown", true},
	}

	for _, tt := range tests {
		t.Run(tt.shellType, func(t *testing.T) {
			script := GetShellHookScript(tt.shellType)
			if tt.wantEmpty && script != "" {
				t.Errorf("expected empty script for %s", tt.shellType)
			}
			if !tt.wantEmpty && script == "" {
				t.Errorf("expected non-empty script for %s", tt.shellType)
			}
		})
	}
}

func TestReadLastLines(t *testing.T) {
	// Create temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_history")

	content := "line1\nline2\nline3\nline4\nline5\n"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name  string
		limit int
		want  int
	}{
		{"all lines", 10, 5},
		{"limited", 3, 3},
		{"exact", 5, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines, err := readLastLines(tmpFile, tt.limit)
			if err != nil {
				t.Fatalf("readLastLines() error = %v", err)
			}
			if len(lines) != tt.want {
				t.Errorf("readLastLines() got %d lines, want %d", len(lines), tt.want)
			}
		})
	}
}

func TestGetHistory(t *testing.T) {
	// This test depends on the system environment
	lines, err := GetHistory(5)
	if err != nil {
		t.Logf("GetHistory() returned error (may be expected): %v", err)
		return
	}

	t.Logf("Got %d history lines", len(lines))
}

func TestReadFailedFromHookFile(t *testing.T) {
	// Create test hook file with current timestamp
	currentTime := time.Now().Unix()
	hookContent := fmt.Sprintf("1|ls -la nonexistent|%d", currentTime)
	hookFile := "/tmp/.ohman_last_failed"

	if err := os.WriteFile(hookFile, []byte(hookContent), 0644); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(hookFile) }()

	cmd, err := readFailedFromHook()
	if err != nil {
		t.Fatalf("readFailedFromHook() error = %v", err)
	}

	if cmd.ExitCode != 1 {
		t.Errorf("exit code = %d, want 1", cmd.ExitCode)
	}

	if cmd.Command != "ls -la nonexistent" {
		t.Errorf("command = %s, want 'ls -la nonexistent'", cmd.Command)
	}
}
