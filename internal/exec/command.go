package exec

import (
	"bytes"
	"fmt"
	"os/exec"
	"syscall"
	"time"
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

	// Use sh -c for shell command support
	cmd := exec.Command("sh", "-c", command)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := &Result{
		Command:  command,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: time.Since(start),
	}

	// Get exit code
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				result.ExitCode = status.ExitStatus()
			} else {
				result.ExitCode = 1
			}
		} else {
			result.ExitCode = 1
		}
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
