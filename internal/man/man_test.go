package man

import (
	"os/exec"
	"testing"
)

func TestGet(t *testing.T) {
	// Skip if man command is not available
	if _, err := exec.LookPath("man"); err != nil {
		t.Skip("man command not available")
	}

	tests := []struct {
		name    string
		command string
		section int
		wantErr bool
	}{
		{
			name:    "existing command ls",
			command: "ls",
			section: 0,
			wantErr: false,
		},
		{
			name:    "existing command grep",
			command: "grep",
			section: 0,
			wantErr: false,
		},
		{
			name:    "non-existing command",
			command: "nonexistentcmd123456",
			section: 0,
			wantErr: true,
		},
		{
			name:    "specific section",
			command: "ls",
			section: 1,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page, err := Get(tt.command, tt.section)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if page == nil {
					t.Error("Get() returned nil page")
					return
				}
				if page.Command != tt.command {
					t.Errorf("Get() command = %v, want %v", page.Command, tt.command)
				}
				if page.Content == "" {
					t.Error("Get() returned empty content")
				}
			}
		})
	}
}

func TestExists(t *testing.T) {
	if _, err := exec.LookPath("man"); err != nil {
		t.Skip("man command not available")
	}

	tests := []struct {
		name    string
		command string
		want    bool
	}{
		{"ls exists", "ls", true},
		{"grep exists", "grep", true},
		{"nonexistent", "nonexistentcmd123456", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Exists(tt.command); got != tt.want {
				t.Errorf("Exists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetWhatis(t *testing.T) {
	if _, err := exec.LookPath("whatis"); err != nil {
		t.Skip("whatis command not available")
	}

	desc, err := GetWhatis("ls")
	if err != nil {
		t.Logf("GetWhatis() warning: %v", err)
		return
	}

	if desc == "" {
		t.Error("GetWhatis() returned empty description")
	}
}
