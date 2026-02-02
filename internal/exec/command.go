package exec

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	pipeit "github.com/liliang-cn/pipeit"
)

// Result represents the result of a command execution
type Result struct {
	Command  string
	ExitCode int
	Stdout   string
	Stderr   string
	Duration time.Duration
}

// Execute runs a command and returns its result
func Execute(command string) (*Result, error) {
	start := time.Now()

	var stdoutBuf, stderrBuf bytes.Buffer

	config := pipeit.Config{
		Command: "sh",
		Args:    []string{"-c", command},
		OnOutput: func(data []byte) {
			stdoutBuf.Write(data)
		},
		OnError: func(data []byte) {
			stderrBuf.Write(data)
		},
	}

	pm := pipeit.NewWithConfig(config)
	if err := pm.StartWithPipes(); err != nil {
		return nil, fmt.Errorf("failed to start: %w", err)
	}
	defer pm.Stop()

	err := pm.Wait()

	result := &Result{
		Command:  command,
		Stdout:   stdoutBuf.String(),
		Stderr:   stderrBuf.String(),
		Duration: time.Since(start),
	}

	// Get exit code
	if err != nil {
		result.ExitCode = 1
	}

	return result, nil
}

// ExecuteWithPTY runs a command with PTY support for interactive programs
func ExecuteWithPTY(command string) (*Result, error) {
	start := time.Now()

	var stdoutBuf, stderrBuf bytes.Buffer

	config := pipeit.Config{
		Command: "sh",
		Args:    []string{"-c", command},
		OnOutput: func(data []byte) {
			stdoutBuf.Write(data)
		},
		OnError: func(data []byte) {
			stderrBuf.Write(data)
		},
	}

	pm := pipeit.NewWithConfig(config)
	if err := pm.StartWithPTY(); err != nil {
		return nil, fmt.Errorf("failed to start: %w", err)
	}
	defer pm.Stop()

	err := pm.Wait()

	result := &Result{
		Command:  command,
		Stdout:   stdoutBuf.String(),
		Stderr:   stderrBuf.String(),
		Duration: time.Since(start),
	}

	// Get exit code
	if err != nil {
		result.ExitCode = 1
	}

	return result, nil
}

// Success returns true if the command succeeded
func (r *Result) Success() bool {
	return r.ExitCode == 0
}

// String returns a formatted string representation
func (r *Result) String() string {
	if r.Success() {
		return fmt.Sprintf("✓ %s (exit: 0)", r.Command)
	}
	return fmt.Sprintf("✗ %s (exit: %d)", r.Command, r.ExitCode)
}

// HasError returns true if there was any stderr output
func (r *Result) HasError() bool {
	return strings.TrimSpace(r.Stderr) != ""
}
