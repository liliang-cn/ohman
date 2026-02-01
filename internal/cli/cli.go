package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/liliang-cn/ohman/internal/app"
	"github.com/liliang-cn/ohman/internal/config"
	"github.com/liliang-cn/ohman/internal/log"
	"github.com/liliang-cn/ohman/internal/session"
	"github.com/liliang-cn/ohman/pkg/version"
	"github.com/spf13/cobra"
)

var (
	cfgFile     string
	section     int
	model       string
	rawMode     bool
	interactive bool
	verbose     bool
)

// rootCmd is the root command
var rootCmd = &cobra.Command{
	Use:   "ohman [command] [question]",
	Short: "Oh Man! - AI-powered man page assistant",
	Long: `Oh Man! is an intelligent command-line assistant powered by LLM.
It combines traditional man pages with AI, allowing you to ask questions
about command usage and parameters in natural language, and even get
automatic fix suggestions when commands fail.

Examples:
  ohman grep "How to search recursively?"    Ask about grep usage
  ohman tar "What does xvf mean?"            Ask about tar parameters
  ohman git                                  Enter interactive mode
  ohman                                      Diagnose last failed command`,
	Version:               version.String(),
	Args:                  cobra.ArbitraryArgs,
	DisableFlagsInUseLine: true,
	RunE:                  runRoot,
}

// Execute executes the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file path")
	rootCmd.PersistentFlags().IntVarP(&section, "section", "s", 0, "man page section (1-8)")
	rootCmd.PersistentFlags().StringVarP(&model, "model", "m", "", "LLM model name")
	rootCmd.PersistentFlags().BoolVarP(&rawMode, "raw", "r", false, "show raw man content only")
	rootCmd.PersistentFlags().BoolVarP(&interactive, "interactive", "i", false, "force interactive mode")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Subcommands
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(historyCmd)
	rootCmd.AddCommand(clearCmd)
	rootCmd.AddCommand(logCmd)
	rootCmd.AddCommand(chatCmd)
}

func initConfig() {
	if cfgFile != "" {
		_ = os.Setenv("OHMAN_CONFIG", cfgFile)
	}
}

func runRoot(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Override model config
	if model != "" {
		cfg.LLM.Model = model
	}

	application := app.New(cfg)

	// Case 1: No args - diagnose failed command
	if len(args) == 0 {
		return application.DiagnoseLastFailed()
	}

	input := strings.Join(args, " ")

	// Case 2: Error message analysis
	// If input looks like an error message (multiline or contains error keywords)
	if looksLikeErrorMessage(input) {
		return application.AnalyzeError(input)
	}

	command := args[0]
	question := ""
	if len(args) > 1 {
		question = strings.Join(args[1:], " ")
	}

	// Case 3: Show raw man content
	if rawMode {
		return application.ShowManPage(command, section)
	}

	// Case 4: Interactive mode
	if interactive || question == "" {
		return application.Interactive(command, section)
	}

	// Case 5: Direct Q&A
	return application.Ask(command, section, question)
}

// looksLikeErrorMessage checks if input looks like an error message
func looksLikeErrorMessage(input string) bool {
	// Check for multiline input (user pasted error output)
	if strings.Contains(input, "\n") {
		return true
	}

	// Check for common error keywords (case-insensitive)
	lowerInput := strings.ToLower(input)
	errorKeywords := []string{
		"error:",
		"error: ",
		"failed",
		"cannot",
		"permission denied",
		"no such file",
		"command not found",
		"segmentation fault",
		"core dumped",
		"fatal",
		"exception",
		"undefined",
		"not found",
		"connection refused",
		"timeout",
	}

	for _, keyword := range errorKeywords {
		if strings.Contains(lowerInput, keyword) {
			return true
		}
	}

	// Check if input is very long (likely pasted output)
	if len(input) > 150 {
		return true
	}

	return false
}

// configCmd is the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure ohman",
	Long:  "Interactive configuration for LLM settings",
	RunE: func(cmd *cobra.Command, args []string) error {
		return config.InteractiveSetup()
	},
}

// historyCmd is the history command
var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "View session history",
	Long:  "View your session history of previous queries and diagnoses",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runHistory(cmd, args)
	},
}

// clearCmd is the clear command
var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear session cache",
	Long:  "Clear all session history",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runClear(cmd, args)
	},
}

var (
	logLimit int
	logUnit  string
)

// logCmd is the log analysis command
var logCmd = &cobra.Command{
	Use:   "log [file|content]",
	Short: "Analyze log files or log content",
	Long:  "Analyze log files, paste log content, or analyze systemd journalctl logs to get AI-powered insights and solutions",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runLog(cmd, args)
	},
}

func init() {
	logCmd.Flags().IntVarP(&logLimit, "limit", "n", 0, "limit number of log entries to analyze (0 = all)")
	logCmd.Flags().StringVarP(&logUnit, "unit", "u", "", "analyze systemd journalctl logs for specific unit (e.g., nginx.service)")
}

// chatCmd is the chat command for interactive conversations
var chatCmd = &cobra.Command{
	Use:   "chat [log-file]",
	Short: "Start an interactive chat session",
	Long:  "Start an interactive chat session with AI. Optionally provide a log file path to analyze and discuss.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runChat(cmd, args)
	},
}

func runLog(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Override model config
	if model != "" {
		cfg.LLM.Model = model
	}

	application := app.New(cfg)

	// Case 1: Check for piped input
	if log.IsPipedInput() {
		content, err := log.ReadFromStdin()
		if err != nil {
			return fmt.Errorf("failed to read from pipe: %w", err)
		}
		if content == "" {
			return fmt.Errorf("no input received from pipe")
		}
		return application.AnalyzeLogContent(content)
	}

	// Case 2: Analyze systemd journalctl unit
	if logUnit != "" {
		return application.AnalyzeJournalctlUnit(logUnit, logLimit)
	}

	// Case 3: No args, provide usage info
	if len(args) == 0 {
		fmt.Println("📋 Log Analysis Usage")
		fmt.Println()
		fmt.Println("Analyze a log file:")
		fmt.Println("  ohman log /var/log/app/error.log")
		fmt.Println()
		fmt.Println("Analyze with limit:")
		fmt.Println("  ohman log -n 100 /var/log/app.log")
		fmt.Println()
		fmt.Println("Analyze systemd service:")
		fmt.Println("  ohman log -u nginx.service")
		fmt.Println("  ohman log --unit docker -n 50")
		fmt.Println()
		fmt.Println("Paste log content directly:")
		fmt.Println("  ohman log \"2025-02-01 ERROR: Database connection failed\"")
		fmt.Println()
		fmt.Println("Use pipe:")
		fmt.Println("  tail -f /var/log/app.log | ohman log")
		fmt.Println("  grep ERROR /var/log/app.log | ohman log")
		fmt.Println("  journalctl -u nginx | ohman log")
		fmt.Println()
		return nil
	}

	input := args[0]

	// Case 4: Auto-detect systemd unit
	if log.IsJournalctlUnit(input) && logLimit == 0 {
		// Ask user if they want to use journalctl
		fmt.Printf("⚠️  '%s' looks like a systemd unit.\n", input)
		fmt.Println("   Use 'ohman log -u <unit>' to analyze journalctl logs.")
		fmt.Println("   Continuing as log content...")
	}

	// Check if input looks like a file path (contains / or starts with . or ~)
	if strings.Contains(input, "\n") || strings.Contains(input, "[ERROR]") || strings.Contains(input, "[FATAL]") {
		// Looks like pasted log content
		return application.AnalyzeLogContent(input)
	}

	// Check if file exists
	if _, err := os.Stat(input); err == nil {
		return application.AnalyzeLogFile(input, logLimit)
	}

	// Treat as log content string
	return application.AnalyzeLogContent(input)
}

func runHistory(cmd *cobra.Command, args []string) error {
	manager, err := session.New()
	if err != nil {
		return fmt.Errorf("failed to initialize session manager: %w", err)
	}

	entries := manager.GetAll()
	count := len(entries)

	if count == 0 {
		fmt.Println("📝 No session history found")
		fmt.Println()
		fmt.Println("💡 Tip: Use 'ohman <command> [question]' to start asking questions")
		return nil
	}

	fmt.Printf("📝 Session History (%d entries)\n", count)
	fmt.Println()

	for i, entry := range entries {
		fmt.Printf("  [%d] %s\n", i+1, formatTimestamp(entry.Timestamp))

		switch entry.Type {
		case "error":
			fmt.Printf("      Type: Error Message Analysis\n")
			if entry.Question != "" {
				fmt.Printf("      Error: %s\n", truncateString(entry.Question, 60))
			}
		default:
			fmt.Printf("      Command: %s\n", entry.Command)
			switch entry.Type {
			case "question":
				if entry.Question != "" {
					fmt.Printf("      Question: %s\n", truncateString(entry.Question, 60))
				}
			case "diagnose":
				fmt.Printf("      Type: Failed Command Diagnosis\n")
			case "interactive":
				if entry.Question != "" {
					fmt.Printf("      Question: %s\n", truncateString(entry.Question, 60))
				}
			case "chat":
				fmt.Printf("      Type: Chat Session\n")
				if entry.Question != "" {
					fmt.Printf("      Question: %s\n", truncateString(entry.Question, 60))
				}
			case "chat-log":
				fmt.Printf("      Type: Chat with Log Analysis\n")
				if entry.Question != "" {
					fmt.Printf("      Question: %s\n", truncateString(entry.Question, 60))
				}
			}
		}

		fmt.Println()
	}

	return nil
}

func runClear(cmd *cobra.Command, args []string) error {
	manager, err := session.New()
	if err != nil {
		return fmt.Errorf("failed to initialize session manager: %w", err)
	}

	count := manager.Count()
	if count == 0 {
		fmt.Println("✅ Session cache is already empty")
		return nil
	}

	fmt.Printf("⚠️  This will clear all %d session entries. Continue? [y/N] ", count)
	var response string
	fmt.Scanln(&response)

	if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
		fmt.Println("❌ Cancelled")
		return nil
	}

	if err := manager.Clear(); err != nil {
		return fmt.Errorf("failed to clear session: %w", err)
	}

	fmt.Println("✅ Session cache cleared")
	return nil
}

func formatTimestamp(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "Just now"
	} else if diff < time.Hour {
		return fmt.Sprintf("%d minutes ago", int(diff.Minutes()))
	} else if diff < 24*time.Hour {
		return fmt.Sprintf("%d hours ago", int(diff.Hours()))
	} else if diff < 7*24*time.Hour {
		return fmt.Sprintf("%d days ago", int(diff.Hours()/24))
	} else {
		return t.Format("2006-01-02 15:04")
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func runChat(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Override model config
	if model != "" {
		cfg.LLM.Model = model
	}

	application := app.New(cfg)

	// If log file is provided, analyze it first
	var logContext string
	if len(args) > 0 {
		logFile := args[0]

		// Check if file exists
		if _, err := os.Stat(logFile); err == nil {
			fmt.Printf("📋 Analyzing log file: %s\n", logFile)
			fmt.Println()

			// Analyze the log file
			result, err := log.AnalyzeFile(logFile, 0)
			if err != nil {
				return fmt.Errorf("failed to analyze log file: %w", err)
			}

			logContext = result.ToFormat()
			fmt.Printf("📊 Found %d log entries\n", result.Total)
			if len(result.Errors) > 0 {
				fmt.Printf("   Errors: %d\n", len(result.Errors))
			}
			if len(result.Warnings) > 0 {
				fmt.Printf("   Warnings: %d\n", len(result.Warnings))
			}
			fmt.Println()
			fmt.Println("Log analysis complete. You can now ask questions about these logs.")
			fmt.Println()
		} else {
			fmt.Printf("⚠️  Log file not found: %s\n", logFile)
			fmt.Println("   Continuing with chat mode...")
			fmt.Println()
		}
	}

	// Start interactive chat
	return application.Chat(logContext)
}
