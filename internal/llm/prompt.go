package llm

import (
	"fmt"
	"strings"
)

// System Prompt Templates

const systemPromptQuestion = `You are a Linux/Unix command-line expert assistant. The user will provide man page content for a command and ask related questions.

Please answer the user's questions accurately based on the man page content. When answering:
1. Prioritize information from the man page
2. Use clear and concise language
3. Provide specific command examples (using shell code block format)
4. If the information is not in the man page, clearly state so
5. For complex option combinations, explain each option's function

Current command: %s

=== MAN PAGE CONTENT ===
%s
=== END OF MAN PAGE ===`

const systemPromptDiagnose = `You are a command-line expert. Analyze the failed command and provide a fix.

Be concise. Use this format:

## Problem
Brief explanation of why it failed.

## Fix
` + "```bash" + `
correct command here
` + "```" + `

One-line explanation if needed.

Failed command: %s
Exit code: %d
Error: %s

=== MAN PAGE ===
%s
===`

const systemPromptInteractive = `You are a Linux/Unix command-line expert assistant, having a conversation with the user about the %s command.

The user has loaded the man page for this command, and you can answer questions based on the content. When answering:
1. Be concise and direct
2. Provide practical command examples
3. Feel free to recommend related useful options or tips

=== MAN PAGE CONTENT ===
%s
=== END OF MAN PAGE ===`

const systemPromptError = `You are a Linux/Unix command-line expert. Analyze the following error message and provide a solution.

Be concise and use this format:

## Problem
Brief explanation of what went wrong.

## Solution
` + "```bash" + `
fixed command here
` + "```" + `

One-line explanation if needed.

=== ERROR MESSAGE ===
%s
=== END OF ERROR ===`

const systemPromptFix = `You are a command fixing assistant. Analyze the failed command and return ONLY the fixed command.

CRITICAL OUTPUT FORMAT:
- Return ONLY the fixed shell command
- Wrap exactly in: __CMD__your command here__CMD__
- NO explanations, NO markdown, NO "here is" prefixes

Example: __CMD__git pull --rebase__CMD__

If multiple fixes are possible, choose the most likely one. Context includes previous attempts - don't repeat them.`

const systemPromptLog = `You are a log analysis expert. Analyze the following log content and provide insights.

When analyzing:
1. Identify the root causes of errors and warnings
2. Provide specific solutions for each issue found
3. Suggest preventive measures
4. Highlight patterns or recurring issues
5. If applicable, recommend configuration or code changes

Use this format:

## Summary
Brief overview of the log analysis.

## Issues Found
For each issue:
- **Issue**: Description of the problem
- **Root Cause**: What caused it
- **Solution**: ` + "```bash" + `
  specific fix or command
  ` + "```" + `

## Recommendations
Actionable recommendations to prevent future issues.

=== LOG ANALYSIS ===
%s
=== END OF LOG ANALYSIS ===`

// BuildQuestionPrompt builds a question prompt
func BuildQuestionPrompt(command, manContent, question string) []Message {
	// Limit man content length to avoid exceeding token limits
	manContent = truncateContent(manContent, 50000)

	messages := []Message{
		{
			Role:    "system",
			Content: fmt.Sprintf(systemPromptQuestion, command, manContent),
		},
	}

	if question != "" {
		messages = append(messages, Message{
			Role:    "user",
			Content: question,
		})
	}

	return messages
}

// BuildDiagnosePrompt builds a diagnose prompt
func BuildDiagnosePrompt(command string, exitCode int, errorMsg, manContent string) []Message {
	// Limit man content length
	manContent = truncateContent(manContent, 50000)

	return []Message{
		{
			Role:    "system",
			Content: fmt.Sprintf(systemPromptDiagnose, command, exitCode, errorMsg, manContent),
		},
		{
			Role:    "user",
			Content: "Please analyze why this command failed and provide fix suggestions.",
		},
	}
}

// BuildInteractivePrompt builds an interactive mode prompt
func BuildInteractivePrompt(command, manContent string) []Message {
	manContent = truncateContent(manContent, 50000)

	return []Message{
		{
			Role:    "system",
			Content: fmt.Sprintf(systemPromptInteractive, command, manContent),
		},
	}
}

// BuildErrorPrompt builds an error analysis prompt
func BuildErrorPrompt(errorMsg string) []Message {
	return []Message{
		{
			Role:    "system",
			Content: fmt.Sprintf(systemPromptError, errorMsg),
		},
		{
			Role:    "user",
			Content: "Please analyze this error and provide a fix.",
		},
	}
}

// BuildLogPrompt builds a log analysis prompt
func BuildLogPrompt(logContent string) []Message {
	return []Message{
		{
			Role:    "system",
			Content: fmt.Sprintf(systemPromptLog, logContent),
		},
		{
			Role:    "user",
			Content: "Please analyze these logs and provide solutions for any issues found.",
		},
	}
}

// FixAttempt represents a single fix attempt for context
type FixAttempt struct {
	Command  string
	ExitCode int
	Stdout   string
	Stderr   string
}

// BuildFixPrompt builds messages for command fixing
func BuildFixPrompt(originalCommand string, attempts []FixAttempt) []Message {
	var ctx strings.Builder
	ctx.WriteString(fmt.Sprintf("Original command: %s\n\n", originalCommand))

	if len(attempts) > 0 {
		ctx.WriteString("Previous attempts:\n")
		for i, a := range attempts {
			ctx.WriteString(fmt.Sprintf("\n[Attempt %d]\n", i+1))
			ctx.WriteString(fmt.Sprintf("Command: %s\n", a.Command))
			ctx.WriteString(fmt.Sprintf("Exit code: %d\n", a.ExitCode))
			if a.Stderr != "" {
				ctx.WriteString(fmt.Sprintf("Error: %s\n", truncateContent(a.Stderr, 500)))
			}
			if a.Stdout != "" {
				ctx.WriteString(fmt.Sprintf("Output: %s\n", truncateContent(a.Stdout, 200)))
			}
		}
	}

	return []Message{
		{
			Role:    "system",
			Content: systemPromptFix,
		},
		{
			Role:    "user",
			Content: ctx.String(),
		},
	}
}

// ExtractCommand extracts the fixed command from LLM response
func ExtractCommand(response string) (string, error) {
	response = strings.TrimSpace(response)

	// Try to extract using __CMD__ tags first
	if start := strings.Index(response, "__CMD__"); start != -1 {
		start += len("__CMD__")
		if end := strings.Index(response[start:], "__CMD__"); end != -1 {
			cmd := strings.TrimSpace(response[start : start+end])
			if cmd != "" {
				return cmd, nil
			}
		}
	}

	// Fallback: extract from markdown code block
	lines := strings.Split(response, "\n")
	inCodeBlock := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "```bash" || trimmed == "```sh" || trimmed == "```" {
			inCodeBlock = !inCodeBlock
			continue
		}
		if inCodeBlock && trimmed != "" {
			return trimmed, nil
		}
	}

	// Last resort: if response is short and looks like a command
	if len(response) < 200 && !strings.ContainsAny(response, "\n") {
		return response, nil
	}

	return "", fmt.Errorf("no command found in response")
}

// truncateContent truncates content to specified character count
func truncateContent(content string, maxChars int) string {
	// Use rune count to correctly handle multi-byte characters
	runes := []rune(content)
	if len(runes) <= maxChars {
		return content
	}

	// Try to truncate at paragraph boundary
	truncated := string(runes[:maxChars])
	if idx := strings.LastIndex(truncated, "\n\n"); idx > maxChars*3/4 {
		truncated = truncated[:idx]
	}

	return truncated + "\n\n... (content truncated)"
}
