package version

import (
	"strings"
	"testing"
)

func TestString(t *testing.T) {
	result := String()

	if result == "" {
		t.Error("String() returned empty string")
	}

	if !strings.Contains(result, Version) {
		t.Errorf("String() should contain version %s", Version)
	}
}

func TestShort(t *testing.T) {
	result := Short()

	if result != Version {
		t.Errorf("Short() = %s, want %s", result, Version)
	}
}
