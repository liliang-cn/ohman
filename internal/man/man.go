package man

import (
	"fmt"
	"os/exec"
	"strings"
)

// ManPage represents a man page
type ManPage struct {
	Command string
	Section int
	Content string
}

// Get retrieves the man page content
func Get(command string, section int) (*ManPage, error) {
	// Build man command arguments
	args := []string{}
	if section > 0 {
		args = append(args, fmt.Sprintf("%d", section))
	}
	args = append(args, command)

	// Use col -b to remove formatting control characters for plain text
	cmdStr := fmt.Sprintf("man %s 2>/dev/null | col -b", strings.Join(args, " "))
	cmd := exec.Command("sh", "-c", cmdStr)

	output, err := cmd.Output()
	if err != nil {
		// Try without col (some systems may not have it)
		cmd = exec.Command("man", args...)
		output, err = cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("man page for %s not found", command)
		}
	}

	content := string(output)
	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("man page for %s not found", command)
	}

	return &ManPage{
		Command: command,
		Section: section,
		Content: content,
	}, nil
}

// Exists checks if a man page exists
func Exists(command string) bool {
	cmd := exec.Command("man", "-w", command)
	return cmd.Run() == nil
}

// GetSections gets which sections have man pages for a command
func GetSections(command string) []int {
	var sections []int
	for i := 1; i <= 8; i++ {
		cmd := exec.Command("man", "-w", fmt.Sprintf("%d", i), command)
		if cmd.Run() == nil {
			sections = append(sections, i)
		}
	}
	return sections
}

// GetWhatis gets a short description of a command
func GetWhatis(command string) (string, error) {
	cmd := exec.Command("whatis", command)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// GetHelpOutput tries to get help output using --help, -h, or -help flags
func GetHelpOutput(command string) (string, error) {
	// Common help flags to try, in order of preference
	helpFlags := []string{"--help", "-h", "-help"}

	for _, flag := range helpFlags {
		cmd := exec.Command(command, flag)
		output, err := cmd.Output()
		if err == nil {
			content := strings.TrimSpace(string(output))
			if content != "" {
				return content, nil
			}
		}
		// Try with combined stderr (some programs output help to stderr)
		cmd = exec.Command(command, flag)
		output, err = cmd.CombinedOutput()
		if err == nil {
			content := strings.TrimSpace(string(output))
			if content != "" {
				return content, nil
			}
		}
	}

	return "", fmt.Errorf("no help output found for %s", command)
}
