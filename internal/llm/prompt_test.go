package llm

import (
	"strings"
	"testing"
)

func TestBuildQuestionPrompt(t *testing.T) {
	command := "grep"
	manContent := "GREP(1) - print lines matching a pattern"
	question := "How to search recursively?"

	messages := BuildQuestionPrompt(command, manContent, question)

	if len(messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(messages))
	}

	if messages[0].Role != "system" {
		t.Errorf("first message should be system, got %s", messages[0].Role)
	}

	if !strings.Contains(messages[0].Content, command) {
		t.Error("system prompt should contain command name")
	}

	if !strings.Contains(messages[0].Content, manContent) {
		t.Error("system prompt should contain man content")
	}

	if messages[1].Role != "user" {
		t.Errorf("second message should be user, got %s", messages[1].Role)
	}

	if messages[1].Content != question {
		t.Errorf("user message should be question, got %s", messages[1].Content)
	}
}

func TestBuildQuestionPromptNoQuestion(t *testing.T) {
	messages := BuildQuestionPrompt("ls", "LS(1)", "")

	if len(messages) != 1 {
		t.Errorf("expected 1 message without question, got %d", len(messages))
	}
}

func TestBuildDiagnosePrompt(t *testing.T) {
	command := "chmod 777 /etc/passwd"
	exitCode := 1
	errorMsg := "Operation not permitted"
	manContent := "CHMOD(1) - change file mode bits"

	messages := BuildDiagnosePrompt(command, exitCode, errorMsg, manContent)

	if len(messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(messages))
	}

	systemContent := messages[0].Content
	if !strings.Contains(systemContent, command) {
		t.Error("should contain failed command")
	}
	if !strings.Contains(systemContent, errorMsg) {
		t.Error("should contain error message")
	}
}

func TestBuildLogPrompt(t *testing.T) {
	logContent := `2025-02-01 10:23:45 ERROR Database connection failed
2025-02-01 10:23:46 WARN Retrying connection
2025-02-01 10:23:47 INFO Application started`

	messages := BuildLogPrompt(logContent)

	if len(messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(messages))
	}

	if messages[0].Role != "system" {
		t.Errorf("first message should be system, got %s", messages[0].Role)
	}

	if !strings.Contains(messages[0].Content, logContent) {
		t.Error("system prompt should contain log content")
	}

	if messages[1].Role != "user" {
		t.Errorf("second message should be user, got %s", messages[1].Role)
	}
}

func TestTruncateContent(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		maxChars int
		wantLen  int
	}{
		{
			name:     "short content",
			content:  "hello",
			maxChars: 100,
			wantLen:  5,
		},
		{
			name:     "long content",
			content:  strings.Repeat("a", 1000),
			maxChars: 100,
			wantLen:  100 + len("\n\n... (content truncated)"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateContent(tt.content, tt.maxChars)
			if len(tt.content) <= tt.maxChars {
				if result != tt.content {
					t.Error("short content should not be modified")
				}
			} else {
				if !strings.HasSuffix(result, "... (content truncated)") {
					t.Error("truncated content should have suffix")
				}
			}
		})
	}
}
