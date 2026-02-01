package app

import (
	"fmt"
	"strings"

	"github.com/liliang-cn/ohman/internal/config"
	"github.com/liliang-cn/ohman/internal/input"
	"github.com/liliang-cn/ohman/internal/llm"
	"github.com/liliang-cn/ohman/internal/log"
	"github.com/liliang-cn/ohman/internal/man"
	"github.com/liliang-cn/ohman/internal/output"
	"github.com/liliang-cn/ohman/internal/session"
	"github.com/liliang-cn/ohman/internal/shell"
)

// App is the main application structure
type App struct {
	cfg        *config.Config
	llmClient  llm.Client
	renderer   *output.Renderer
	sessionMgr *session.Manager
}

// New creates a new application instance
func New(cfg *config.Config) *App {
	sessionMgr, err := session.New()
	if err != nil {
		// If session manager fails to initialize, continue without it
		sessionMgr = nil
	}

	return &App{
		cfg:        cfg,
		renderer:   output.NewRenderer(cfg.Output),
		sessionMgr: sessionMgr,
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
	response, err := client.Chat(messages)
	if err != nil {
		return fmt.Errorf("failed to call LLM: %w", err)
	}

	// 4. Save to session history
	if a.sessionMgr != nil {
		_ = a.sessionMgr.Add(session.Entry{
			Command:  command,
			Question: question,
			Answer:   response.Content,
			Type:     "question",
		})
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
	content := "(no documentation available)"
	if err != nil {
		// Try --help flag as fallback
		fmt.Printf("⚠️  No man page for %s, trying --help...\n", cmdName)
		helpOutput, helpErr := man.GetHelpOutput(cmdName)
		if helpErr == nil {
			content = helpOutput
			fmt.Printf("✅ Got help output for %s\n", cmdName)
		} else {
			fmt.Printf("⚠️  Unable to get documentation for %s, but will still try to diagnose\n", cmdName)
		}
	} else {
		content = manPage.Content
	}

	// 4. Initialize LLM client
	client, err := a.getLLMClient()
	if err != nil {
		return err
	}

	// 5. Build diagnose prompt and call LLM
	messages := llm.BuildDiagnosePrompt(failedCmd.Command, failedCmd.ExitCode, failedCmd.Error, content)

	fmt.Println("🔧 Analyzing...")
	fmt.Println()
	response, err := client.Chat(messages)
	if err != nil {
		return fmt.Errorf("failed to call LLM: %w", err)
	}

	// Save to session history
	if a.sessionMgr != nil {
		_ = a.sessionMgr.Add(session.Entry{
			Command: cmdName,
			Answer:  response.Content,
			Type:    "diagnose",
		})
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

	reader := input.New("❓ ")
	history := llm.BuildQuestionPrompt(command, manPage.Content, "")

	for {
		question, err := reader.ReadLine()
		if err != nil {
			break
		}

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

		// Save to session history
		if a.sessionMgr != nil {
			_ = a.sessionMgr.Add(session.Entry{
				Command:  command,
				Question: question,
				Answer:   response.Content,
				Type:     "interactive",
			})
		}

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

// AnalyzeError analyzes an error message and provides suggestions
func (a *App) AnalyzeError(errorMsg string) error {
	// Initialize LLM client
	client, err := a.getLLMClient()
	if err != nil {
		return err
	}

	fmt.Println("🔍 Analyzing error message...")
	fmt.Println()

	// Build error analysis prompt
	messages := llm.BuildErrorPrompt(errorMsg)

	response, err := client.Chat(messages)
	if err != nil {
		return fmt.Errorf("failed to call LLM: %w", err)
	}

	// Save to session history
	if a.sessionMgr != nil {
		_ = a.sessionMgr.Add(session.Entry{
			Question: errorMsg,
			Answer:   response.Content,
			Type:     "error",
		})
	}

	// Streaming output is already printed, just add a newline
	fmt.Println()

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

// Chat starts an interactive chat session with optional log context
func (a *App) Chat(logContext string) error {
	// Initialize LLM client
	client, err := a.getLLMClient()
	if err != nil {
		return err
	}

	// Build initial messages
	var messages []llm.Message

	// If log context is provided, add it as system prompt
	if logContext != "" {
		messages = append(messages, llm.Message{
			Role: "system",
			Content: "You are a log analysis expert. The following log content has been analyzed. " +
				"Use this context to answer the user's questions about the logs.\n\n" + logContext,
		})
		fmt.Println("💬 Chat mode started with log context")
		fmt.Println("   You can ask questions about the analyzed logs.")
		fmt.Println()
	} else {
		messages = append(messages, llm.Message{
			Role: "system",
			Content: "You are a Linux/Unix command-line expert assistant. " +
				"Help users with commands, errors, and technical problems.",
		})
		fmt.Println("💬 Chat mode started")
		fmt.Println("   Type your question or 'exit' to quit.")
		fmt.Println()
	}

	reader := input.New("❓ ")

	for {
		question, err := reader.ReadLine()
		if err != nil {
			break
		}

		if question == "" {
			continue
		}

		// Check for exit commands
		if question == "exit" || question == "quit" || question == "q" || question == "bye" {
			fmt.Println("👋 Goodbye!")
			break
		}

		// Add user question to history
		messages = append(messages, llm.Message{
			Role:    "user",
			Content: question,
		})

		// Call LLM
		fmt.Println()
		response, err := client.Chat(messages)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			continue
		}

		// Add assistant response to history
		messages = append(messages, llm.Message{
			Role:    "assistant",
			Content: response.Content,
		})

		// Save to session history
		if a.sessionMgr != nil {
			entryType := "chat"
			if logContext != "" {
				entryType = "chat-log"
			}
			_ = a.sessionMgr.Add(session.Entry{
				Command:  "chat",
				Question: question,
				Answer:   response.Content,
				Type:     entryType,
			})
		}

		// Add newline for readability
		fmt.Println()
		fmt.Println()
	}

	return nil
}

// AnalyzeLogFile analyzes a log file and provides AI-powered insights
func (a *App) AnalyzeLogFile(filePath string, limit int) error {
	// Initialize LLM client
	client, err := a.getLLMClient()
	if err != nil {
		return err
	}

	fmt.Printf("📋 Analyzing log file: %s\n", filePath)
	fmt.Println()

	// Analyze the log file
	analyzer := log.New(filePath)
	result, err := analyzer.Analyze(limit)
	if err != nil {
		return fmt.Errorf("failed to analyze log file: %w", err)
	}

	// Format the analysis result for AI
	logContent := result.ToFormat()

	fmt.Printf("📊 Found %d log entries\n", result.Total)
	if len(result.Errors) > 0 {
		fmt.Printf("   Errors: %d\n", len(result.Errors))
	}
	if len(result.Warnings) > 0 {
		fmt.Printf("   Warnings: %d\n", len(result.Warnings))
	}
	fmt.Println()

	// Build log prompt and call LLM
	messages := llm.BuildLogPrompt(logContent)

	fmt.Println("🔍 Analyzing...")
	fmt.Println()
	response, err := client.Chat(messages)
	if err != nil {
		return fmt.Errorf("failed to call LLM: %w", err)
	}

	// Save to session history
	if a.sessionMgr != nil {
		_ = a.sessionMgr.Add(session.Entry{
			Command:  "log",
			Question: fmt.Sprintf("file:%s (limit:%d)", filePath, limit),
			Answer:   response.Content,
			Type:     "log",
		})
	}

	// Streaming output is already printed, just add a newline
	fmt.Println()

	return nil
}

// AnalyzeLogContent analyzes log content from a string
func (a *App) AnalyzeLogContent(content string) error {
	// Initialize LLM client
	client, err := a.getLLMClient()
	if err != nil {
		return err
	}

	fmt.Println("📋 Analyzing log content...")
	fmt.Println()

	// Analyze the log content
	result := log.AnalyzeString(content)

	// Format as analysis result for AI
	logContent := result.ToFormat()

	fmt.Printf("📊 Found %d log entries\n", result.Total)
	if len(result.Errors) > 0 {
		fmt.Printf("   Errors: %d\n", len(result.Errors))
	}
	if len(result.Warnings) > 0 {
		fmt.Printf("   Warnings: %d\n", len(result.Warnings))
	}
	fmt.Println()

	// Build log prompt and call LLM
	messages := llm.BuildLogPrompt(logContent)

	fmt.Println("🔍 Analyzing...")
	fmt.Println()
	response, err := client.Chat(messages)
	if err != nil {
		return fmt.Errorf("failed to call LLM: %w", err)
	}

	// Save to session history
	if a.sessionMgr != nil {
		_ = a.sessionMgr.Add(session.Entry{
			Command:  "log",
			Question: content[:min(len(content), 100)],
			Answer:   response.Content,
			Type:     "log",
		})
	}

	// Streaming output is already printed, just add a newline
	fmt.Println()

	return nil
}

// AnalyzeJournalctlUnit analyzes logs from journalctl for a specific unit
func (a *App) AnalyzeJournalctlUnit(unit string, limit int) error {
	// Initialize LLM client
	client, err := a.getLLMClient()
	if err != nil {
		return err
	}

	fmt.Printf("📋 Analyzing journalctl logs for unit: %s\n", unit)
	fmt.Println()

	// Get logs from journalctl
	journalContent, err := log.GetJournalctlLogs(unit, limit)
	if err != nil {
		return fmt.Errorf("failed to get journalctl logs: %w", err)
	}

	// Analyze the journal content
	result := log.AnalyzeString(journalContent)

	// Format as analysis result for AI
	logContent := result.ToFormat()

	fmt.Printf("📊 Found %d log entries\n", result.Total)
	if len(result.Errors) > 0 {
		fmt.Printf("   Errors: %d\n", len(result.Errors))
	}
	if len(result.Warnings) > 0 {
		fmt.Printf("   Warnings: %d\n", len(result.Warnings))
	}
	fmt.Println()

	// Build log prompt and call LLM
	messages := llm.BuildLogPrompt(logContent)

	fmt.Println("🔍 Analyzing...")
	fmt.Println()
	response, err := client.Chat(messages)
	if err != nil {
		return fmt.Errorf("failed to call LLM: %w", err)
	}

	// Save to session history
	if a.sessionMgr != nil {
		_ = a.sessionMgr.Add(session.Entry{
			Command:  "journalctl",
			Question: fmt.Sprintf("unit:%s (limit:%d)", unit, limit),
			Answer:   response.Content,
			Type:     "log",
		})
	}

	// Streaming output is already printed, just add a newline
	fmt.Println()

	return nil
}
