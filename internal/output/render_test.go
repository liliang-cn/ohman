package output

import (
	"testing"

	"github.com/liliang-cn/ohman/internal/config"
)

func TestNewRenderer(t *testing.T) {
	cfg := config.OutputConfig{
		Color:    true,
		Markdown: true,
	}

	renderer := NewRenderer(cfg)
	if renderer == nil {
		t.Fatal("NewRenderer() returned nil")
	}

	if renderer.cfg.Color != cfg.Color {
		t.Error("Color config not set correctly")
	}

	if renderer.cfg.Markdown != cfg.Markdown {
		t.Error("Markdown config not set correctly")
	}
}

func TestColorize(t *testing.T) {
	text := "test"

	result := Colorize(text, Red)
	if result != Red+text+Reset {
		t.Errorf("Colorize() = %q, want %q", result, Red+text+Reset)
	}

	result = Colorize(text, Green)
	if result != Green+text+Reset {
		t.Errorf("Colorize() = %q, want %q", result, Green+text+Reset)
	}
}
