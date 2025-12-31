package app

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/liliang-cn/ohman/internal/config"
	"github.com/liliang-cn/ohman/internal/llm"
	"github.com/liliang-cn/ohman/internal/man"
	"github.com/liliang-cn/ohman/internal/output"
	"github.com/liliang-cn/ohman/internal/shell"
)

// App is the main application structure
type App struct {
	cfg       *config.Config
	llmClient llm.Client
	renderer  *output.Renderer
}

// New creates a new application instance
func New(cfg *config.Config) *App {
	return &App{
		cfg:      cfg,
		renderer: output.NewRenderer(cfg.Output),
	}
}

// Ask asks a question about a command
func (a *App) Ask(command string, section int, question string) error {
	// 1. Get man page
	manPage, err := man.Get(command, section)
	if err != nil {
		return fmt.Errorf("failed to get man page: %w", err)
	}

	// 2. Initialize LLM client
	client, err := a.getLLMClient()
	if err != nil {
		return err
	}

	// 3. Build prompt and call LLM
	messages := llm.BuildQuestionPrompt(command, manPage.Content, question)

	fmt.Println("🤔 Thinking...")
	fmt.Println()
	_, err = client.Chat(messages)
	if err != nil {
		return fmt.Errorf("failed to call LLM: %w", err)
	}

	// Streaming output is already printed, just add a newline
	fmt.Println()

	return nil
}

// DiagnoseLastFailed diagnoses the last failed command
func (a *App) DiagnoseLastFailed() error {
	// Get last command (from hook file or history)
	failedCmd, err := shell.GetLastFailed()
	if err != nil {
		fmt.Println("✅ No recent command found to diagnose.")
		fmt.Println()
		fmt.Println("💡 Tip: You can use 'ohman <command> [question]' to ask about command usage")
		return nil
	}

	fmt.Printf("🔍 Analyzing command: %s\n", failedCmd.Command)
	fmt.Println()

	// Parse command name
	cmdName := parseCommandName(failedCmd.Command)
	if cmdName == "" {
		return fmt.Errorf("unable to parse command name")
	}

	// Get related man content
	manPage, err := man.Get(cmdName, 0)
	if err != nil {
		fmt.Printf("⚠️  Unable to get man page for %s, but will still try to diagnose\n", cmdName)
		manPage = &man.ManPage{Command: cmdName, Content: "(man page not available)"}
	}

	// 4. Initialize LLM client
	client, err := a.getLLMClient()
	if err != nil {
		return err
	}

	// 5. Build diagnose prompt and call LLM
	messages := llm.BuildDiagnosePrompt(failedCmd.Command, failedCmd.ExitCode, failedCmd.Error, manPage.Content)

	fmt.Println("🔧 Analyzing...")
	fmt.Println()
	_, err = client.Chat(messages)
	if err != nil {
		return fmt.Errorf("failed to call LLM: %w", err)
	}

	// Streaming output is already printed, just add a newline
	fmt.Println()

	return nil
}

// Interactive enters interactive mode
func (a *App) Interactive(command string, section int) error {
	// 1. Get man page
	manPage, err := man.Get(command, section)
	if err != nil {
		return fmt.Errorf("failed to get man page: %w", err)
	}

	// 2. Initialize LLM client
	client, err := a.getLLMClient()
	if err != nil {
		return err
	}

	fmt.Printf("📖 Loaded man page for %s, entering interactive mode\n", command)
	fmt.Println("   Type your question, or 'exit' / 'quit' to exit")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	history := llm.BuildQuestionPrompt(command, manPage.Content, "")

	for {
		fmt.Print("❓ ")
		input, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		question := strings.TrimSpace(input)
		if question == "" {
			continue
		}

		if question == "exit" || question == "quit" || question == "q" {
			fmt.Println("👋 Goodbye!")
			break
		}

		// Add user question to history
		history = append(history, llm.Message{Role: "user", Content: question})

		fmt.Println()
		response, err := client.Chat(history)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			continue
		}

		// Add assistant response to history
		history = append(history, llm.Message{Role: "assistant", Content: response.Content})

		// Streaming output is already printed, just add newlines
		fmt.Println()
		fmt.Println()
	}

	return nil
}

// ShowManPage shows the raw man page
func (a *App) ShowManPage(command string, section int) error {
	manPage, err := man.Get(command, section)
	if err != nil {
		return fmt.Errorf("failed to get man page: %w", err)
	}

	fmt.Println(manPage.Content)
	return nil
}

// getLLMClient gets the LLM client
func (a *App) getLLMClient() (llm.Client, error) {
	if a.llmClient != nil {
		return a.llmClient, nil
	}

	if a.cfg.LLM.APIKey == "" && a.cfg.LLM.Provider != "ollama" {
		return nil, fmt.Errorf("API Key not configured, please run 'ohman config'")
	}

	client, err := llm.NewClient(a.cfg.LLM)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM client: %w", err)
	}

	a.llmClient = client
	return client, nil
}

// parseCommandName parses the command name from a full command
func parseCommandName(fullCmd string) string {
	// Skip environment variable settings like "VAR=value command"
	parts := strings.Fields(fullCmd)
	for _, part := range parts {
		if !strings.Contains(part, "=") {
			// Handle paths like "/usr/bin/ls" -> "ls"
			if idx := strings.LastIndex(part, "/"); idx != -1 {
				return part[idx+1:]
			}
			return part
		}
	}
	return ""
}
