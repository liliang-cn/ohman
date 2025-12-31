package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/liliang-cn/ohman/internal/app"
	"github.com/liliang-cn/ohman/internal/config"
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

	command := args[0]
	question := ""
	if len(args) > 1 {
		question = strings.Join(args[1:], " ")
	}

	// Case 2: Show raw man content
	if rawMode {
		return application.ShowManPage(command, section)
	}

	// Case 3: Interactive mode
	if interactive || question == "" {
		return application.Interactive(command, section)
	}

	// Case 4: Direct Q&A
	return application.Ask(command, section, question)
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
