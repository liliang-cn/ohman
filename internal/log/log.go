package log

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// LogLevel represents log level
type LogLevel string

const (
	LevelDebug LogLevel = "DEBUG"
	LevelInfo  LogLevel = "INFO"
	LevelWarn  LogLevel = "WARN"
	LevelError LogLevel = "ERROR"
	LevelFatal LogLevel = "FATAL"
)

// LogEntry represents a parsed log entry
type LogEntry struct {
	Timestamp string
	Level     LogLevel
	Message   string
	Raw       string
}

// LogType represents the type of log file
type LogType string

const (
	TypeApplication LogType = "application"
	TypeSystem      LogType = "system"
	TypeAccess      LogType = "access"
	TypeError       LogType = "error"
	TypeUnknown     LogType = "unknown"
)

// Analyzer analyzes log files
type Analyzer struct {
	filePath string
	logType  LogType
}

// New creates a new log analyzer
func New(filePath string) *Analyzer {
	return &Analyzer{
		filePath: filePath,
		logType:  TypeUnknown,
	}
}

// GetJournalctlLogs retrieves logs from journalctl for a specific service unit
func GetJournalctlLogs(unit string, limit int) (string, error) {
	// Check if journalctl exists
	if _, err := exec.LookPath("journalctl"); err != nil {
		return "", fmt.Errorf("journalctl not found: %w", err)
	}

	// Build journalctl command
	args := []string{"-u", unit, "--no-pager"}

	// Add limit if specified
	if limit > 0 {
		args = append(args, "-n", fmt.Sprintf("%d", limit))
	}

	// Execute journalctl
	cmd := exec.Command("journalctl", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Check if unit doesn't exist
		errStr := stderr.String()
		if strings.Contains(errStr, "No journal files were found") ||
			strings.Contains(errStr, "Unit") ||
			strings.Contains(errStr, "not found") {
			return "", fmt.Errorf("journalctl failed: %s", errStr)
		}
	}

	output := stdout.String()
	if output == "" {
		return "", fmt.Errorf("no logs found for unit %s", unit)
	}

	return output, nil
}

// IsJournalctlUnit checks if input looks like a systemd unit
func IsJournalctlUnit(input string) bool {
	// Common systemd unit patterns
	unitPatterns := []string{
		".service",
		".socket",
		".timer",
		".path",
		".slice",
		".scope",
		".device",
		".mount",
		".automount",
		".swap",
		".target",
	}

	for _, pattern := range unitPatterns {
		if strings.HasSuffix(input, pattern) {
			return true
		}
	}

	// Check for common service names (like nginx, docker, sshd)
	commonServices := []string{
		"nginx", "apache2", "apache", "httpd",
		"docker", "containerd",
		"sshd", "ssh",
		"mysql", "mariadb", "postgresql", "postgres",
		"redis", "mongodb",
		"systemd", "network",
	}

	lowerInput := strings.ToLower(input)
	for _, service := range commonServices {
		if lowerInput == service {
			return true
		}
	}

	return false
}

// AnalyzeFile analyzes a log file and returns the analysis result
func AnalyzeFile(filePath string, limit int) (*AnalysisResult, error) {
	analyzer := New(filePath)
	return analyzer.Analyze(limit)
}

// Analyze analyzes the log file and returns key findings
func (a *Analyzer) Analyze(limit int) (*AnalysisResult, error) {
	// Detect log type
	a.detectType()

	// Read and parse log entries
	entries, err := a.readEntries(limit)
	if err != nil {
		return nil, err
	}

	// Analyze entries
	result := &AnalysisResult{
		LogType:  a.logType,
		Total:    len(entries),
		ByLevel:  make(map[LogLevel]int),
		Samples:  getSampleEntries(entries, 20),
		Errors:   filterByLevel(entries, []LogLevel{LevelError, LevelFatal}),
		Warnings: filterByLevel(entries, []LogLevel{LevelWarn}),
	}

	for _, entry := range entries {
		result.ByLevel[entry.Level]++
	}

	return result, nil
}

// AnalyzeString analyzes log content from string
func AnalyzeString(content string) *AnalysisResult {
	lines := strings.Split(content, "\n")
	entries := make([]LogEntry, 0)

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		entry := parseLine(line)
		entries = append(entries, entry)
	}

	result := &AnalysisResult{
		LogType:  TypeApplication,
		Total:    len(entries),
		ByLevel:  make(map[LogLevel]int),
		Samples:  getSampleEntries(entries, 20),
		Errors:   filterByLevel(entries, []LogLevel{LevelError, LevelFatal}),
		Warnings: filterByLevel(entries, []LogLevel{LevelWarn}),
	}

	for _, entry := range entries {
		result.ByLevel[entry.Level]++
	}

	return result
}

// AnalysisResult represents the analysis result
type AnalysisResult struct {
	LogType  LogType
	Total    int
	ByLevel  map[LogLevel]int
	Samples  []LogEntry
	Errors   []LogEntry
	Warnings []LogEntry
}

// ToFormat converts the result to a formatted string for AI analysis
func (r *AnalysisResult) ToFormat() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("## Log Analysis Summary\n\n"))
	sb.WriteString(fmt.Sprintf("Log Type: %s\n", r.LogType))
	sb.WriteString(fmt.Sprintf("Total Entries: %d\n\n", r.Total))

	sb.WriteString("## Statistics by Level\n")
	for level, count := range r.ByLevel {
		if count > 0 {
			sb.WriteString(fmt.Sprintf("- %s: %d\n", level, count))
		}
	}
	sb.WriteString("\n")

	if len(r.Errors) > 0 {
		sb.WriteString(fmt.Sprintf("## Error Entries (%d)\n", len(r.Errors)))
		errorSamples := getSampleEntries(r.Errors, 10)
		for _, entry := range errorSamples {
			sb.WriteString(fmt.Sprintf("%s\n\n", entry.Raw))
		}
		sb.WriteString("\n")
	}

	if len(r.Warnings) > 0 {
		sb.WriteString(fmt.Sprintf("## Warning Entries (%d)\n", len(r.Warnings)))
		warnSamples := getSampleEntries(r.Warnings, 5)
		for _, entry := range warnSamples {
			sb.WriteString(fmt.Sprintf("%s\n\n", entry.Raw))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Sample Log Entries\n")
	for _, entry := range r.Samples {
		sb.WriteString(fmt.Sprintf("%s\n\n", entry.Raw))
	}

	return sb.String()
}

// readEntries reads log entries from file
func (a *Analyzer) readEntries(limit int) ([]LogEntry, error) {
	file, err := os.Open(a.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	var entries []LogEntry
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		entry := parseLine(line)
		entries = append(entries, entry)

		if limit > 0 && len(entries) >= limit {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read log file: %w", err)
	}

	return entries, nil
}

// detectType detects the type of log file
func (a *Analyzer) detectType() {
	if strings.Contains(a.filePath, "access") {
		a.logType = TypeAccess
	} else if strings.Contains(a.filePath, "error") || strings.Contains(a.filePath, "err") {
		a.logType = TypeError
	} else if strings.Contains(a.filePath, "syslog") || strings.Contains(a.filePath, "dmesg") {
		a.logType = TypeSystem
	} else {
		a.logType = TypeApplication
	}
}

// parseLine parses a log line
func parseLine(line string) LogEntry {
	level := parseLevel(line)

	return LogEntry{
		Timestamp: parseTimestamp(line),
		Level:     level,
		Message:   parseMessage(line),
		Raw:       line,
	}
}

// parseLevel parses log level from line
func parseLevel(line string) LogLevel {
	lower := strings.ToLower(line)

	for _, level := range []LogLevel{LevelFatal, LevelError, LevelWarn, LevelDebug, LevelInfo} {
		if strings.Contains(lower, strings.ToLower(string(level))) {
			return level
		}
	}

	return LevelInfo
}

// parseTimestamp extracts timestamp from log line
func parseTimestamp(line string) string {
	// Common timestamp patterns
	patterns := []string{
		`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}`,
		`\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}`,
		`\d{2}/\d{2}/\d{4} \d{2}:\d{2}:\d{2}`,
		`\w{3} \d{1,2} \d{2}:\d{2}:\d{2}`,
	}

	for _, pattern := range patterns {
		if strings.Contains(line, pattern[:6]) {
			return extractTimestampByPattern(line, pattern[:6])
		}
	}

	return ""
}

func extractTimestampByPattern(line, pattern string) string {
	idx := strings.Index(line, pattern)
	if idx == -1 {
		return ""
	}

	// Extract up to 30 chars after pattern match
	end := idx + 30
	if end > len(line) {
		end = len(line)
	}

	return strings.TrimSpace(line[idx:end])
}

// parseMessage extracts message from log line
func parseMessage(line string) string {
	// Remove common log prefixes
	removed := false

	patterns := []string{
		`\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}`,
		`\[?\w+\]?\s*:\s*`,
	}

	for _, pattern := range patterns {
		if idx := strings.Index(line, pattern); idx != -1 {
			line = line[idx+len(pattern):]
			removed = true
			break
		}
	}

	// Find message after level
	levelMarkers := []string{"DEBUG", "INFO", "WARN", "WARNING", "ERROR", "FATAL"}
	for _, marker := range levelMarkers {
		if idx := strings.Index(line, marker); idx != -1 {
			after := line[idx+len(marker):]
			return strings.TrimSpace(after)
		}
	}

	if removed {
		return strings.TrimSpace(line)
	}

	return line
}

// filterByLevel filters entries by log level
func filterByLevel(entries []LogEntry, levels []LogLevel) []LogEntry {
	var filtered []LogEntry

	for _, entry := range entries {
		for _, level := range levels {
			if entry.Level == level {
				filtered = append(filtered, entry)
				break
			}
		}
	}

	return filtered
}

// ReadFromStdin reads log content from stdin (for pipe support)
func ReadFromStdin() (string, error) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return "", err
	}

	// Check if stdin is a pipe or redirected file
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// It's a pipe or redirected file
		var buffer bytes.Buffer
		if _, err := buffer.ReadFrom(os.Stdin); err != nil {
			return "", fmt.Errorf("failed to read from stdin: %w", err)
		}
		return buffer.String(), nil
	}

	return "", nil
}

// IsPipedInput checks if there is piped input
func IsPipedInput() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}

	// If not a character device, it's likely a pipe or redirect
	return (stat.Mode() & os.ModeCharDevice) == 0
}

// getSampleEntries returns sample entries
func getSampleEntries(entries []LogEntry, max int) []LogEntry {
	if len(entries) <= max {
		return entries
	}

	// Return samples from beginning, middle, and end
	samples := make([]LogEntry, 0, max)

	begin := max / 3
	middle := len(entries) / 2

	samples = append(samples, entries[:begin]...)
	if middle < len(entries)-begin {
		samples = append(samples, entries[middle:middle+begin]...)
	}
	samples = append(samples, entries[len(entries)-begin:]...)

	return samples
}
