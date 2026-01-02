package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Entry represents a session entry
type Entry struct {
	ID        string    `json:"id"`
	Command   string    `json:"command"`
	Question  string    `json:"question"`
	Answer    string    `json:"answer"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"` // "question", "diagnose", "interactive"
}

// Manager manages session history
type Manager struct {
	filePath string
	mu       sync.RWMutex
	entries  []Entry
}

// New creates a new session manager
func New() (*Manager, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return nil, err
	}

	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	filePath := filepath.Join(configDir, "history.json")

	m := &Manager{
		filePath: filePath,
		entries:  make([]Entry, 0),
	}

	// Load existing history
	if err := m.load(); err != nil {
		// If file doesn't exist, that's okay
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	return m, nil
}

// Add adds a new session entry
func (m *Manager) Add(entry Entry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if entry.ID == "" {
		entry.ID = generateID()
	}
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	m.entries = append(m.entries, entry)

	// Trim to last 100 entries if too large
	if len(m.entries) > 100 {
		m.entries = m.entries[len(m.entries)-100:]
	}

	return m.save()
}

// GetAll returns all entries
func (m *Manager) GetAll() []Entry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to avoid race conditions
	result := make([]Entry, len(m.entries))
	copy(result, m.entries)
	return result
}

// GetLastN returns the last n entries
func (m *Manager) GetLastN(n int) []Entry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if n <= 0 || n > len(m.entries) {
		n = len(m.entries)
	}

	start := len(m.entries) - n
	result := make([]Entry, n)
	copy(result, m.entries[start:])
	return result
}

// Clear removes all entries
func (m *Manager) Clear() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.entries = make([]Entry, 0)
	return m.save()
}

// Count returns the number of entries
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.entries)
}

// load loads entries from file
func (m *Manager) load() error {
	data, err := os.ReadFile(m.filePath)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return nil
	}

	return json.Unmarshal(data, &m.entries)
}

// save saves entries to file
func (m *Manager) save() error {
	data, err := json.MarshalIndent(m.entries, "", "  ")
	if err != nil {
		return err
	}

	// Write to temporary file first, then rename for atomicity
	tmpPath := m.filePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return err
	}

	return os.Rename(tmpPath, m.filePath)
}

// getConfigDir returns the config directory path
func getConfigDir() (string, error) {
	// Check environment variable first
	if dir := os.Getenv("OHMAN_CONFIG_DIR"); dir != "" {
		return dir, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".config", "ohman"), nil
}

// generateID generates a unique ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
