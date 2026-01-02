package cli

import (
	"testing"
)

func TestLooksLikeErrorMessage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "multiline error",
			input:    "error: something failed\nat line 42",
			expected: true,
		},
		{
			name:     "error with colon",
			input:    "error: command not found",
			expected: true,
		},
		{
			name:     "permission denied",
			input:    "bash: ./script.sh: Permission denied",
			expected: true,
		},
		{
			name:     "failed keyword",
			input:    "Build failed with exit code 1",
			expected: true,
		},
		{
			name:     "cannot keyword",
			input:    "cannot open file",
			expected: true,
		},
		{
			name:     "no such file",
			input:    "No such file or directory",
			expected: true,
		},
		{
			name:     "segmentation fault",
			input:    "Segmentation fault (core dumped)",
			expected: true,
		},
		{
			name:     "fatal error",
			input:    "fatal: repository not found",
			expected: true,
		},
		{
			name:     "long input",
			input:    "this is a very long input that exceeds 150 characters and should be detected as an error message because it's likely pasted output from some command that failed for some reason",
			expected: true,
		},
		{
			name:     "simple command",
			input:    "ls",
			expected: false,
		},
		{
			name:     "command with question",
			input:    "grep how to search recursively",
			expected: false,
		},
		{
			name:     "short text",
			input:    "hello world",
			expected: false,
		},
		{
			name:     "command with options",
			input:    "tar -xvf file.tar.gz",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := looksLikeErrorMessage(tt.input)
			if result != tt.expected {
				t.Errorf("looksLikeErrorMessage(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
