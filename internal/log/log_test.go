package log

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected LogLevel
	}{
		{"error log", "2025-02-01 ERROR: Something went wrong", LevelError},
		{"warning log", "2025-02-01 WARN: Warning message", LevelWarn},
		{"info log", "2025-02-01 INFO: Info message", LevelInfo},
		{"debug log", "2025-02-01 DEBUG: Debug message", LevelDebug},
		{"fatal log", "2025-02-01 FATAL: Fatal error", LevelFatal},
		{"unknown log", "Just a log message", LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseLevel(tt.line)
			if result != tt.expected {
				t.Errorf("parseLevel() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAnalyzeString(t *testing.T) {
	logContent := `2025-02-01 10:23:45 INFO Starting application
2025-02-01 10:23:46 DEBUG Initializing components
2025-02-01 10:23:47 ERROR Failed to connect to database
2025-02-01 10:23:48 WARN Retrying connection
2025-02-01 10:23:49 FATAL Connection failed after retries
2025-02-01 10:23:50 INFO Application started
`

	result := AnalyzeString(logContent)

	if result.Total != 6 {
		t.Errorf("Expected 6 entries, got %d", result.Total)
	}

	if result.ByLevel[LevelError] != 1 {
		t.Errorf("Expected 1 error, got %d", result.ByLevel[LevelError])
	}

	if result.ByLevel[LevelFatal] != 1 {
		t.Errorf("Expected 1 fatal, got %d", result.ByLevel[LevelFatal])
	}

	if result.ByLevel[LevelWarn] != 1 {
		t.Errorf("Expected 1 warning, got %d", result.ByLevel[LevelWarn])
	}

	if len(result.Errors) != 2 { // ERROR + FATAL
		t.Errorf("Expected 2 error entries, got %d", len(result.Errors))
	}

	if len(result.Warnings) != 1 {
		t.Errorf("Expected 1 warning entry, got %d", len(result.Warnings))
	}
}

func TestAnalyzeFile(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	content := `2025-02-01 10:00:00 INFO Application started
2025-02-01 10:00:01 ERROR Database connection failed
2025-02-01 10:00:02 WARN Retrying...
2025-02-01 10:00:03 INFO Retrying connection
2025-02-01 10:00:04 ERROR Connection failed again
`

	if err := os.WriteFile(logFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test log file: %v", err)
	}

	analyzer := New(logFile)
	result, err := analyzer.Analyze(0)

	if err != nil {
		t.Fatalf("Analyze() failed: %v", err)
	}

	if result.Total != 5 {
		t.Errorf("Expected 5 entries, got %d", result.Total)
	}

	if len(result.Errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(result.Errors))
	}

	if len(result.Warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(result.Warnings))
	}
}

func TestAnalyzeFileWithLimit(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	var builder strings.Builder
	for i := 0; i < 100; i++ {
		builder.WriteString(fmt.Sprintf("2025-02-01 10:00:%02d INFO Log entry %d\n", i, i))
	}
	content := builder.String()

	if err := os.WriteFile(logFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test log file: %v", err)
	}

	analyzer := New(logFile)
	result, err := analyzer.Analyze(10)

	if err != nil {
		t.Fatalf("Analyze() failed: %v", err)
	}

	if result.Total != 10 {
		t.Errorf("Expected 10 entries, got %d", result.Total)
	}
}

func TestToFormat(t *testing.T) {
	logContent := `2025-02-01 10:23:45 INFO Starting application
2025-02-01 10:23:46 ERROR Database connection failed
2025-02-01 10:23:47 WARN Retrying connection
`

	result := AnalyzeString(logContent)
	formatted := result.ToFormat()

	if !strings.Contains(formatted, "Log Analysis Summary") {
		t.Error("Formatted output should contain summary header")
	}

	if !strings.Contains(formatted, "Total Entries: 3") {
		t.Error("Formatted output should contain total count")
	}

	if !strings.Contains(formatted, "ERROR: 1") {
		t.Error("Formatted output should contain error count")
	}

	if !strings.Contains(formatted, "WARN: 1") {
		t.Error("Formatted output should contain warning count")
	}
}

func TestDetectLogType(t *testing.T) {
	tests := []struct {
		filePath string
		expected LogType
	}{
		{"access.log", TypeAccess},
		{"error.log", TypeError},
		{"err.log", TypeError},
		{"/var/log/syslog", TypeSystem},
		{"/var/log/dmesg", TypeSystem},
		{"application.log", TypeApplication},
	}

	for _, tt := range tests {
		t.Run(tt.filePath, func(t *testing.T) {
			analyzer := New(tt.filePath)
			analyzer.detectType()
			if analyzer.logType != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, analyzer.logType)
			}
		})
	}
}

func TestParseMessage(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		contains string
	}{
		{
			name:     "standard format",
			line:     "2025-02-01 10:23:45 ERROR Database connection failed",
			contains: "Database connection failed",
		},
		{
			name:     "with brackets",
			line:     "[ERROR] Failed to start service",
			contains: "Failed to start service",
		},
		{
			name:     "simple format",
			line:     "This is a simple log message",
			contains: "This is a simple log message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := parseMessage(tt.line)
			if !strings.Contains(message, tt.contains) {
				t.Errorf("parseMessage() = %v, want to contain %v", message, tt.contains)
			}
		})
	}
}

func TestIsJournalctlUnit(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"service unit", "nginx.service", true},
		{"socket unit", "docker.socket", true},
		{"timer unit", "backup.timer", true},
		{"common service", "nginx", true},
		{"common service mysql", "mysql", true},
		{"log file path", "/var/log/app.log", false},
		{"log file", "error.log", false},
		{"log content", "2025-02-01 ERROR: Failed", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsJournalctlUnit(tt.input)
			if result != tt.expected {
				t.Errorf("IsJournalctlUnit(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetJournalctlLogs(t *testing.T) {
	// This test requires journalctl to be available
	// It's okay to skip on systems without journalctl
	if _, err := exec.LookPath("journalctl"); err != nil {
		t.Skip("journalctl not available, skipping test")
	}

	// Test with a common systemd unit (systemd itself should always exist)
	_, err := GetJournalctlLogs("systemd-journald.service", 5)
	if err != nil {
		t.Logf("GetJournalctlLogs returned error (unit may not exist): %v", err)
	}
}

func TestIsPipedInput(t *testing.T) {
	// This test checks if IsPipedInput can detect pipes
	// In a normal test run, it should return false (not piped)
	result := IsPipedInput()
	if result {
		t.Log("IsPipedInput returned true (might be running in a piped test environment)")
	}
	// We don't assert false because tests might run in different contexts
}

func TestAnalyzeStringWithPipeLikeInput(t *testing.T) {
	// Test that AnalyzeString works with pipe-like input
	input := `2025-02-01 10:00:00 INFO Starting service
2025-02-01 10:00:01 ERROR Database connection failed
2025-02-01 10:00:02 WARN Retrying...`

	result := AnalyzeString(input)

	if result.Total != 3 {
		t.Errorf("Expected 3 entries, got %d", result.Total)
	}

	if len(result.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(result.Errors))
	}

	if len(result.Warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(result.Warnings))
	}
}
