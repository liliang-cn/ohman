package app

import (
	"testing"
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
