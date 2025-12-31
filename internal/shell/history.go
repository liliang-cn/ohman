package shell

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// FailedCommand represents a failed command
type FailedCommand struct {
	Command  string
	ExitCode int
	Error    string
	Time     time.Time
}

// GetLastFailed gets the last failed command
func GetLastFailed() (*FailedCommand, error) {
	// Method 1: Try reading the hook-recorded file (most accurate)
	if cmd, err := readFailedFromHook(); err == nil {
		return cmd, nil
	}

	// Method 2: Fallback - read last command from shell history
	// Assume the user just ran a failed command and immediately runs ohman
	if cmd, err := getLastCommandFromHistory(); err == nil {
		return cmd, nil
	}

	return nil, fmt.Errorf("unable to get failed command information")
}

// getLastCommandFromHistory reads the last command from shell history file
func getLastCommandFromHistory() (*FailedCommand, error) {
	history, err := GetHistory(5) // Get a few more in case some are ohman itself
	if err != nil || len(history) == 0 {
		return nil, fmt.Errorf("cannot read history")
	}

	shellType := DetectShell()

	// Find the last non-ohman command
	for i := len(history) - 1; i >= 0; i-- {
		cmd := parseHistoryLine(history[i], shellType)
		if cmd == "" {
			continue
		}
		// Skip if command is ohman itself
		if strings.HasPrefix(cmd, "ohman") || strings.HasPrefix(cmd, "./bin/ohman") {
			continue
		}
		return &FailedCommand{
			Command:  cmd,
			ExitCode: 1, // Unknown exit code
			Time:     time.Now(),
		}, nil
	}

	return nil, fmt.Errorf("no previous command found")
}

// parseHistoryLine parses a history line based on shell type
func parseHistoryLine(line, shellType string) string {
	line = strings.TrimSpace(line)
	if line == "" {
		return ""
	}

	switch shellType {
	case "zsh":
		// Zsh extended history format: ": timestamp:0;command"
		if strings.HasPrefix(line, ": ") {
			if idx := strings.Index(line, ";"); idx != -1 {
				return strings.TrimSpace(line[idx+1:])
			}
		}
		return line
	case "fish":
		// Fish format: "- cmd: command" or just the command
		if strings.HasPrefix(line, "- cmd: ") {
			return strings.TrimSpace(line[7:])
		}
		return line
	default:
		// Bash and others: plain command
		return line
	}
}

// readFailedFromHook reads failed command from hook file
func readFailedFromHook() (*FailedCommand, error) {
	// Hook file format: exitcode|command|timestamp
	pid := os.Getppid() // Use parent process PID (i.e., the shell's PID)
	hookFile := fmt.Sprintf("/tmp/.ohman_last_failed_%d", pid)

	data, err := os.ReadFile(hookFile)
	if err != nil {
		// Try the generic file without PID
		hookFile = "/tmp/.ohman_last_failed"
		data, err = os.ReadFile(hookFile)
		if err != nil {
			return nil, err
		}
	}

	parts := strings.SplitN(strings.TrimSpace(string(data)), "|", 3)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid hook file format")
	}

	exitCode, _ := strconv.Atoi(parts[0])
	command := parts[1]

	var timestamp time.Time
	if len(parts) >= 3 {
		if ts, err := strconv.ParseInt(parts[2], 10, 64); err == nil {
			timestamp = time.Unix(ts, 0)
		}
	}

	// Check if command is too old (more than 5 minutes)
	if !timestamp.IsZero() && time.Since(timestamp) > 5*time.Minute {
		return nil, fmt.Errorf("the latest failed command has expired")
	}

	return &FailedCommand{
		Command:  command,
		ExitCode: exitCode,
		Time:     timestamp,
	}, nil
}

// GetHistory gets shell history records
func GetHistory(limit int) ([]string, error) {
	historyFile := getHistoryFile()
	if historyFile == "" {
		return nil, fmt.Errorf("unable to determine history file location")
	}

	return readLastLines(historyFile, limit)
}

// getHistoryFile gets the history file path
func getHistoryFile() string {
	// Check environment variable
	if histFile := os.Getenv("HISTFILE"); histFile != "" {
		return histFile
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	// Detect shell type
	shell := os.Getenv("SHELL")
	if strings.Contains(shell, "zsh") {
		return filepath.Join(home, ".zsh_history")
	}
	if strings.Contains(shell, "bash") {
		return filepath.Join(home, ".bash_history")
	}

	// Default to bash history
	return filepath.Join(home, ".bash_history")
}

// readLastLines reads the last n lines of a file
func readLastLines(filePath string, n int) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > n {
			lines = lines[1:]
		}
	}

	return lines, scanner.Err()
}

// DetectShell detects the current shell
func DetectShell() string {
	shell := os.Getenv("SHELL")
	if strings.Contains(shell, "zsh") {
		return "zsh"
	}
	if strings.Contains(shell, "bash") {
		return "bash"
	}
	if strings.Contains(shell, "fish") {
		return "fish"
	}
	return "unknown"
}

// GetShellHookScript gets the shell hook script
func GetShellHookScript(shellType string) string {
	switch shellType {
	case "zsh":
		return `# Oh Man! Failed command recording hook
ohman_precmd() {
    local exit_code=$?
    if [[ $exit_code -ne 0 ]]; then
        echo "$exit_code|$(fc -ln -1)|$(date +%s)" > /tmp/.ohman_last_failed_$$
    fi
}
precmd_functions+=(ohman_precmd)`

	case "bash":
		return `# Oh Man! Failed command recording hook
ohman_prompt_command() {
    local exit_code=$?
    if [[ $exit_code -ne 0 ]]; then
        echo "$exit_code|$(history 1 | sed 's/^[ ]*[0-9]*[ ]*//')|$(date +%s)" > /tmp/.ohman_last_failed_$$
    fi
}
PROMPT_COMMAND="ohman_prompt_command${PROMPT_COMMAND:+; $PROMPT_COMMAND}"`

	case "fish":
		return `# Oh Man! Failed command recording hook
function ohman_postexec --on-event fish_postexec
    set -l exit_code $status
    if test $exit_code -ne 0
        echo "$exit_code|$argv|"(date +%s) > /tmp/.ohman_last_failed_%self
    end
end`

	default:
		return ""
	}
}

// getShellConfigFile returns the shell config file path
func getShellConfigFile(shellType string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	switch shellType {
	case "zsh":
		return filepath.Join(home, ".zshrc")
	case "bash":
		// Prefer .bashrc, fallback to .bash_profile
		bashrc := filepath.Join(home, ".bashrc")
		if _, err := os.Stat(bashrc); err == nil {
			return bashrc
		}
		return filepath.Join(home, ".bash_profile")
	case "fish":
		return filepath.Join(home, ".config", "fish", "config.fish")
	default:
		return ""
	}
}

// IsHookInstalled checks if the shell hook is already installed
func IsHookInstalled() bool {
	shellType := DetectShell()
	configFile := getShellConfigFile(shellType)
	if configFile == "" {
		return false
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return false
	}

	content := string(data)
	switch shellType {
	case "zsh":
		return strings.Contains(content, "ohman_precmd")
	case "bash":
		return strings.Contains(content, "ohman_prompt_command")
	case "fish":
		return strings.Contains(content, "ohman_postexec")
	}
	return false
}

// EnsureHookInstalled checks and installs the shell hook if needed
// Returns: (installed, needsReload, error)
func EnsureHookInstalled() (installed bool, needsReload bool, err error) {
	if IsHookInstalled() {
		return true, false, nil
	}

	shellType := DetectShell()
	if shellType == "unknown" {
		return false, false, fmt.Errorf("unsupported shell")
	}

	configFile := getShellConfigFile(shellType)
	if configFile == "" {
		return false, false, fmt.Errorf("cannot determine shell config file")
	}

	hookScript := GetShellHookScript(shellType)
	if hookScript == "" {
		return false, false, fmt.Errorf("no hook script for shell: %s", shellType)
	}

	// Ensure directory exists for fish
	if shellType == "fish" {
		dir := filepath.Dir(configFile)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return false, false, fmt.Errorf("failed to create config directory: %w", err)
		}
	}

	// Append hook to config file
	f, err := os.OpenFile(configFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return false, false, fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString("\n" + hookScript + "\n"); err != nil {
		return false, false, fmt.Errorf("failed to write hook: %w", err)
	}

	return true, true, nil
}
