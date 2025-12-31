package output

import (
	"fmt"
	"os"

	"github.com/liliang-cn/ohman/internal/config"
)

// Renderer is an output renderer
type Renderer struct {
	cfg config.OutputConfig
}

// NewRenderer creates a new renderer
func NewRenderer(cfg config.OutputConfig) *Renderer {
	return &Renderer{cfg: cfg}
}

// Render renders the output
func (r *Renderer) Render(content string) {
	if !r.cfg.Markdown {
		fmt.Println(content)
		return
	}

	// Try to render Markdown using glamour
	// If it fails, output the original text
	rendered, err := r.renderMarkdown(content)
	if err != nil {
		fmt.Println(content)
		return
	}

	fmt.Print(rendered)
}

// renderMarkdown renders Markdown content
func (r *Renderer) renderMarkdown(content string) (string, error) {
	// Simple implementation: just return the content
	// Full implementation requires importing the glamour library
	// Here we provide basic code block highlighting

	// If terminal doesn't support colors, return as-is
	if !r.cfg.Color || !isTerminal() {
		return content, nil
	}

	return content, nil
}

// isTerminal checks if output is a terminal
func isTerminal() bool {
	fileInfo, _ := os.Stdout.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// Color constants
const (
	Reset   = "\033[0m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"
	Bold    = "\033[1m"
)

// Colorize adds color to text
func Colorize(text, color string) string {
	return color + text + Reset
}

// Success prints a success message
func Success(format string, args ...interface{}) {
	fmt.Printf(Green+"✅ "+format+Reset+"\n", args...)
}

// Error prints an error message
func Error(format string, args ...interface{}) {
	fmt.Printf(Red+"❌ "+format+Reset+"\n", args...)
}

// Warning prints a warning message
func Warning(format string, args ...interface{}) {
	fmt.Printf(Yellow+"⚠️  "+format+Reset+"\n", args...)
}

// Info prints an info message
func Info(format string, args ...interface{}) {
	fmt.Printf(Blue+"ℹ️  "+format+Reset+"\n", args...)
}
