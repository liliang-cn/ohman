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
	// Method 1: Try reading the hook-recorded file
	if cmd, err := readFailedFromHook(); err == nil {
		return cmd, nil
	}

	// Method 2: Try getting from shell built-in variables (requires user configuration)
	// Here we provide a fallback solution

	return nil, fmt.Errorf("unable to get failed command information")
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

	default:
		return ""
	}
}
