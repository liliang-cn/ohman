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
