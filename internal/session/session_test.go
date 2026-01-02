package session

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	oldDir := os.Getenv("OHMAN_CONFIG_DIR")
	os.Setenv("OHMAN_CONFIG_DIR", tmpDir)
	defer os.Setenv("OHMAN_CONFIG_DIR", oldDir)

	m, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if m == nil {
		t.Fatal("New() returned nil manager")
	}

	if m.Count() != 0 {
		t.Errorf("New() initial count = %d, want 0", m.Count())
	}
}

func TestAddAndGet(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("OHMAN_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("OHMAN_CONFIG_DIR")

	m, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	entry := Entry{
		Command:   "ls",
		Question:  "How to show hidden files?",
		Answer:    "Use the -a flag",
		Type:      "question",
		Timestamp: time.Now(),
	}

	err = m.Add(entry)
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	if m.Count() != 1 {
		t.Errorf("Count() = %d, want 1", m.Count())
	}

	entries := m.GetAll()
	if len(entries) != 1 {
		t.Fatalf("GetAll() returned %d entries, want 1", len(entries))
	}

	if entries[0].Command != "ls" {
		t.Errorf("Entry command = %s, want ls", entries[0].Command)
	}

	if entries[0].Question != "How to show hidden files?" {
		t.Errorf("Entry question = %s, want 'How to show hidden files?'", entries[0].Question)
	}
}

func TestGetLastN(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("OHMAN_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("OHMAN_CONFIG_DIR")

	m, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Add 5 entries
	for i := 1; i <= 5; i++ {
		entry := Entry{
			Command:  fmt.Sprintf("cmd%d", i),
			Question: fmt.Sprintf("question%d", i),
			Type:     "question",
		}
		if err := m.Add(entry); err != nil {
			t.Fatalf("Add() error = %v", err)
		}
	}

	// Get last 3
	last3 := m.GetLastN(3)
	if len(last3) != 3 {
		t.Fatalf("GetLastN(3) returned %d entries, want 3", len(last3))
	}

	// Should be cmd3, cmd4, cmd5
	if last3[0].Command != "cmd3" {
		t.Errorf("GetLastN(3)[0].Command = %s, want cmd3", last3[0].Command)
	}
	if last3[2].Command != "cmd5" {
		t.Errorf("GetLastN(3)[2].Command = %s, want cmd5", last3[2].Command)
	}
}

func TestClear(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("OHMAN_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("OHMAN_CONFIG_DIR")

	m, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Add an entry
	entry := Entry{
		Command:  "ls",
		Question: "test",
		Type:     "question",
	}
	if err := m.Add(entry); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// Clear
	if err := m.Clear(); err != nil {
		t.Fatalf("Clear() error = %v", err)
	}

	if m.Count() != 0 {
		t.Errorf("Count() after Clear = %d, want 0", m.Count())
	}
}

func TestPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("OHMAN_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("OHMAN_CONFIG_DIR")

	// Create first manager and add entry
	m1, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	entry := Entry{
		Command:  "tar",
		Question: "How to extract?",
		Answer:   "Use -x flag",
		Type:     "question",
	}
	if err := m1.Add(entry); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// Create second manager (should load from disk)
	m2, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if m2.Count() != 1 {
		t.Errorf("Second manager Count() = %d, want 1", m2.Count())
	}

	entries := m2.GetAll()
	if len(entries) != 1 {
		t.Fatalf("Second manager GetAll() returned %d entries, want 1", len(entries))
	}

	if entries[0].Command != "tar" {
		t.Errorf("Persisted entry command = %s, want tar", entries[0].Command)
	}
}

func TestTrimTo100Entries(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("OHMAN_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("OHMAN_CONFIG_DIR")

	m, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Add 105 entries
	for i := 1; i <= 105; i++ {
		entry := Entry{
			Command:  fmt.Sprintf("cmd%d", i),
			Question: "test",
			Type:     "question",
		}
		if err := m.Add(entry); err != nil {
			t.Fatalf("Add() error = %v", err)
		}
	}

	// Should only have 100
	if m.Count() != 100 {
		t.Errorf("Count() after adding 105 entries = %d, want 100", m.Count())
	}

	// First entry should be cmd6 (cmd1-cmd5 were trimmed)
	entries := m.GetAll()
	if entries[0].Command != "cmd6" {
		t.Errorf("First entry after trim = %s, want cmd6", entries[0].Command)
	}
}
