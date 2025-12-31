# Oh Man! Technical Design Document

## 1. Overview

### 1.1 Project Goals

`ohman` is a command-line tool that enhances the traditional `man` page experience using LLMs (Large Language Models). The core concept is to send the complete man page content as context to an LLM, allowing users to ask questions in natural language and receive accurate answers.

### 1.2 Core Features

1. **Command Q&A**: `ohman <command> [question]` - Ask about a specific command's usage
2. **Failure Diagnosis**: `ohman` (no arguments) - Analyze the last failed command and provide suggestions
3. **Interactive Mode**: `ohman <command>` - Enter interactive Q&A mode
4. **Configuration Management**: `ohman config` - Configure LLM and other settings

## 2. System Architecture

### 2.1 Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              ohman CLI                                   │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐    │
│  │  CLI Layer  │  │Business Layer│ │Service Layer│  │  Base Layer │    │
│  ├─────────────┤  ├─────────────┤  ├─────────────┤  ├─────────────┤    │
│  │ • Arg Parse │  │ • Q&A Logic │  │ • LLM Client│  │ • Config    │    │
│  │ • Routing   │  │ • Diagnosis │  │ • Man Parser│  │ • Logging   │    │
│  │ • Interactive│ │ • Sessions  │  │ • Shell Hist│  │ • FileSystem│    │
│  │ • Rendering │  │ • Prompts   │  │ • HTTP      │  │ • Errors    │    │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘    │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### 2.2 Module Responsibilities

#### CLI Layer (`internal/cli/`)

- Parse command-line arguments
- Route to corresponding handlers
- Manage interactive sessions
- Render output results

#### Business Layer (`internal/app/`)

- Coordinate services to complete business logic
- Build LLM prompts
- Manage session context
- Handle user questions

#### Service Layer (`internal/service/`)

- LLM API calls
- Man page retrieval and parsing
- Shell history reading
- Failed command detection

#### Base Layer (`internal/pkg/`)

- Configuration file read/write
- Logging
- Common utility functions

## 3. Core Flows

### 3.1 Command Q&A Flow

```go
// Pseudocode flow
func HandleQuestion(command, question string) {
    // 1. Get man page content
    manContent := man.Get(command, section)

    // 2. Build prompt
    prompt := prompt.Build(PromptTypeQuestion, PromptData{
        Command:    command,
        ManContent: manContent,
        Question:   question,
    })

    // 3. Call LLM
    response := llm.Chat(prompt)

    // 4. Render output
    output.Render(response)
}
```

### 3.2 Failed Command Diagnosis Flow

```go
// Pseudocode flow
func HandleDiagnose() {
    // 1. Get shell history
    history := shell.GetHistory()

    // 2. Get last failed command
    failedCmd := shell.GetLastFailed(history)
    if failedCmd == nil {
        fmt.Println("No failed command detected")
        return
    }

    // 3. Parse command name
    cmdName := parseCommandName(failedCmd.Command)

    // 4. Get corresponding man content
    manContent := man.Get(cmdName, 0)

    // 5. Build diagnosis prompt
    prompt := prompt.Build(PromptTypeDiagnose, PromptData{
        Command:     failedCmd.Command,
        Error:       failedCmd.Error,
        ExitCode:    failedCmd.ExitCode,
        ManContent:  manContent,
    })

    // 6. Call LLM for diagnosis suggestions
    response := llm.Chat(prompt)

    // 7. Render output
    output.Render(response)
}
```

### 3.3 Interactive Mode Flow

```go
func HandleInteractive(command string) {
    manContent := man.Get(command, section)
    session := session.New(command, manContent)

    for {
        question := readInput()
        if question == "exit" || question == "quit" {
            break
        }

        response := session.Ask(question)
        output.Render(response)
    }
}
```

## 4. Core Module Design

### 4.1 LLM Client Design

```go
// internal/llm/client.go

// Client is the LLM client interface
type Client interface {
    Chat(ctx context.Context, messages []Message) (*Response, error)
    ChatStream(ctx context.Context, messages []Message) (<-chan StreamChunk, error)
}

// Message represents a chat message
type Message struct {
    Role    string `json:"role"`    // system, user, assistant
    Content string `json:"content"`
}

// Response represents an LLM response
type Response struct {
    Content     string
    TokensUsed  int
    FinishReason string
}

// Config LLM configuration
type Config struct {
    Provider    string  // openai, anthropic, ollama, custom
    APIKey      string
    BaseURL     string
    Model       string
    MaxTokens   int
    Temperature float64
}

// NewClient creates an LLM client
func NewClient(cfg Config) (Client, error) {
    switch cfg.Provider {
    case "openai":
        return NewOpenAIClient(cfg)
    case "anthropic":
        return NewAnthropicClient(cfg)
    case "ollama":
        return NewOllamaClient(cfg)
    case "custom":
        return NewCustomClient(cfg)
    default:
        return nil, fmt.Errorf("unsupported provider: %s", cfg.Provider)
    }
}
```

### 4.2 Man Page Module Design

```go
// internal/man/man.go

// ManPage represents a man page
type ManPage struct {
    Command     string
    Section     int
    Content     string
    Synopsis    string
    Description string
    Options     []Option
}

// Option represents a command option
type Option struct {
    Short       string
    Long        string
    Description string
}

// Get retrieves man page content
func Get(command string, section int) (*ManPage, error) {
    args := []string{}
    if section > 0 {
        args = append(args, fmt.Sprintf("%d", section))
    }
    args = append(args, command)

    // Use col -b to remove formatting control characters
    cmd := exec.Command("sh", "-c", fmt.Sprintf("man %s | col -b", strings.Join(args, " ")))
    output, err := cmd.Output()
    if err != nil {
        return nil, fmt.Errorf("failed to get man page: %w", err)
    }

    return &ManPage{
        Command: command,
        Section: section,
        Content: string(output),
    }, nil
}

// Exists checks if a man page exists
func Exists(command string) bool {
    cmd := exec.Command("man", "-w", command)
    return cmd.Run() == nil
}
```

### 4.3 Shell History Module Design

```go
// internal/shell/history.go

// HistoryEntry represents a history entry
type HistoryEntry struct {
    Command   string
    Timestamp time.Time
    ExitCode  int
    Output    string // if available
}

// FailedCommand represents a failed command
type FailedCommand struct {
    Command  string
    ExitCode int
    Error    string
    Time     time.Time
}

// GetLastFailed gets the last failed command
func GetLastFailed() (*FailedCommand, error) {
    shellType := detectShell()

    switch shellType {
    case "zsh":
        return getZshLastFailed()
    case "bash":
        return getBashLastFailed()
    default:
        return nil, fmt.Errorf("unsupported shell: %s", shellType)
    }
}

// detectShell detects the current shell
func detectShell() string {
    shell := os.Getenv("SHELL")
    if strings.Contains(shell, "zsh") {
        return "zsh"
    }
    if strings.Contains(shell, "bash") {
        return "bash"
    }
    return "unknown"
}
```

### 4.4 Prompt Template Design

```go
// internal/llm/prompt.go

// PromptType represents prompt type
type PromptType int

const (
    PromptTypeQuestion PromptType = iota
    PromptTypeDiagnose
    PromptTypeInteractive
)

// System prompt templates
const systemPromptQuestion = `You are a Linux/Unix command-line expert assistant. The user will provide man page content for a command and ask related questions.

Please answer the user's question accurately based on the man page content:
1. Prioritize information from the man page
2. Use clear and concise language
3. Provide specific command examples
4. If the man page doesn't contain relevant information, state that clearly

Current command: {{.Command}}

=== MAN PAGE CONTENT ===
{{.ManContent}}
=== END OF MAN PAGE ===
`

const systemPromptDiagnose = `You are a Linux/Unix command-line diagnostic expert. The user executed a command that failed. Please analyze the cause and provide solutions.

Please answer in the following format:
1. **Problem Analysis**: Explain why the command failed
2. **Solution**: Provide the correct command syntax
3. **Additional Notes**: Related tips or best practices

Failed command: {{.Command}}
Exit code: {{.ExitCode}}
Error message: {{.Error}}

=== RELATED MAN PAGE ===
{{.ManContent}}
=== END OF MAN PAGE ===
`

// Build builds the prompt
func Build(ptype PromptType, data PromptData) []Message {
    var systemPrompt string

    switch ptype {
    case PromptTypeQuestion:
        systemPrompt = renderTemplate(systemPromptQuestion, data)
    case PromptTypeDiagnose:
        systemPrompt = renderTemplate(systemPromptDiagnose, data)
    }

    messages := []Message{
        {Role: "system", Content: systemPrompt},
    }

    if data.Question != "" {
        messages = append(messages, Message{
            Role:    "user",
            Content: data.Question,
        })
    }

    return messages
}
```

### 4.5 Configuration Management Design

```go
// internal/config/config.go

// Config represents application configuration
type Config struct {
    LLM    LLMConfig    `yaml:"llm"`
    Shell  ShellConfig  `yaml:"shell"`
    Output OutputConfig `yaml:"output"`
}

// LLMConfig represents LLM configuration
type LLMConfig struct {
    Provider    string  `yaml:"provider"`
    APIKey      string  `yaml:"api_key"`
    BaseURL     string  `yaml:"base_url"`
    Model       string  `yaml:"model"`
    MaxTokens   int     `yaml:"max_tokens"`
    Temperature float64 `yaml:"temperature"`
}

// ShellConfig represents shell configuration
type ShellConfig struct {
    HistoryFile string `yaml:"history_file"`
}

// OutputConfig represents output configuration
type OutputConfig struct {
    Color    bool `yaml:"color"`
    Markdown bool `yaml:"markdown"`
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
    return &Config{
        LLM: LLMConfig{
            Provider:    "openai",
            Model:       "gpt-4o-mini",
            MaxTokens:   4096,
            Temperature: 0.7,
        },
        Output: OutputConfig{
            Color:    true,
            Markdown: true,
        },
    }
}

// Load loads configuration
func Load() (*Config, error) {
    configPath := getConfigPath()

    if _, err := os.Stat(configPath); os.IsNotExist(err) {
        return DefaultConfig(), nil
    }

    data, err := os.ReadFile(configPath)
    if err != nil {
        return nil, err
    }

    var cfg Config
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, err
    }

    return &cfg, nil
}

// getConfigPath returns the config file path
func getConfigPath() string {
    // Priority: OHMAN_CONFIG > ~/.config/ohman/config.yaml
    if path := os.Getenv("OHMAN_CONFIG"); path != "" {
        return path
    }

    home, _ := os.UserHomeDir()
    return filepath.Join(home, ".config", "ohman", "config.yaml")
}
```

## 5. Failed Command Detection Strategy

### 5.1 Detection Methods

Since different shells handle history records and exit codes differently, we employ multiple strategies:

#### Method 1: Using Shell Built-in Variables (Recommended)

Users can add hooks in their shell configuration to record failed commands:

**Zsh (`~/.zshrc`)**:

```zsh
# Record failed commands to a temporary file
ohman_precmd() {
    local exit_code=$?
    if [[ $exit_code -ne 0 ]]; then
        echo "$exit_code|$(fc -ln -1)|$(date +%s)" > /tmp/.ohman_last_failed_$$
    fi
}
precmd_functions+=(ohman_precmd)
```

**Bash (`~/.bashrc`)**:

```bash
# Record failed commands to a temporary file
ohman_prompt_command() {
    local exit_code=$?
    if [[ $exit_code -ne 0 ]]; then
        echo "$exit_code|$(history 1 | sed 's/^[ ]*[0-9]*[ ]*//')|$(date +%s)" > /tmp/.ohman_last_failed_$$
    fi
}
PROMPT_COMMAND="ohman_prompt_command${PROMPT_COMMAND:+; $PROMPT_COMMAND}"
```

#### Method 2: Analyzing Shell History (Fallback)

```go
// Read shell history file to get recent commands
// Note: This method cannot get the exit code
func getLastCommandFromHistory() (string, error) {
    historyFile := getHistoryFile()

    // Read last few lines
    lines, err := readLastLines(historyFile, 10)
    if err != nil {
        return "", err
    }

    // Return the last non-empty command
    for i := len(lines) - 1; i >= 0; i-- {
        cmd := strings.TrimSpace(lines[i])
        if cmd != "" && !strings.HasPrefix(cmd, "ohman") {
            return cmd, nil
        }
    }

    return "", fmt.Errorf("no command found in history")
}
```

### 5.2 Implementation

```go
// internal/shell/failed.go

func GetLastFailed() (*FailedCommand, error) {
    // Method 1: Try reading the hook-recorded file
    if cmd, err := readFailedFromHook(); err == nil {
        return cmd, nil
    }

    // Method 2: Try getting from shell history (fallback)
    // In this case, inform the user we cannot confirm if the command failed
    cmd, err := getLastCommandFromHistory()
    if err != nil {
        return nil, err
    }

    return &FailedCommand{
        Command:  cmd,
        ExitCode: -1, // Unknown
        Error:    "(Cannot get error info, please confirm this is the failed command)",
    }, nil
}
```

## 6. Dependency Management

### 6.1 Main Dependencies

```go
// go.mod
module github.com/liliang-cn/ohman

go 1.21

require (
    github.com/spf13/cobra v1.8.0      // CLI framework
    github.com/spf13/viper v1.18.0     // Configuration management
    github.com/charmbracelet/glamour v0.6.0  // Markdown rendering
    github.com/charmbracelet/lipgloss v0.9.0 // Terminal styling
    github.com/charmbracelet/bubbletea v0.25.0 // TUI framework (interactive mode)
    gopkg.in/yaml.v3 v3.0.1            // YAML parsing
)
```

### 6.2 Optional Dependencies

- `github.com/sashabaranov/go-openai` - OpenAI SDK
- `github.com/liushuangls/go-anthropic` - Anthropic SDK

## 7. Error Handling

### 7.1 Error Types

```go
// internal/errors/errors.go

var (
    ErrManNotFound      = errors.New("man page not found")
    ErrLLMUnavailable   = errors.New("LLM service unavailable")
    ErrConfigNotFound   = errors.New("configuration not found")
    ErrAPIKeyMissing    = errors.New("API key not configured")
    ErrShellNotSupported = errors.New("shell not supported")
    ErrNoFailedCommand  = errors.New("no failed command detected")
)
```

### 7.2 User-Friendly Messages

```go
func handleError(err error) {
    switch {
    case errors.Is(err, ErrManNotFound):
        fmt.Println("❌ Man page not found for this command. Please verify the command name.")
    case errors.Is(err, ErrAPIKeyMissing):
        fmt.Println("❌ API Key not configured. Please run 'ohman config' to set it up.")
    case errors.Is(err, ErrNoFailedCommand):
        fmt.Println("✅ No failed command detected. Everything looks good!")
    default:
        fmt.Printf("❌ An error occurred: %v\n", err)
    }
}
```

## 8. Testing Strategy

### 8.1 Unit Tests

```go
// internal/man/man_test.go

func TestGet(t *testing.T) {
    tests := []struct {
        name    string
        command string
        wantErr bool
    }{
        {"existing command", "ls", false},
        {"non-existing command", "nonexistentcmd123", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := Get(tt.command, 0)
            if (err != nil) != tt.wantErr {
                t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### 8.2 Integration Tests

```go
// internal/app/app_test.go

func TestHandleQuestion(t *testing.T) {
    // Use mock LLM client
    mockClient := &MockLLMClient{
        Response: "This is a test response",
    }

    app := NewApp(WithLLMClient(mockClient))

    result, err := app.HandleQuestion("ls", "How to show hidden files")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    if result == "" {
        t.Error("expected non-empty result")
    }
}
```

## 9. Release & Distribution

### 9.1 Build Script

````makefile
# Makefile

VERSION := $(shell git describe --tags --always --dirty)
LDFLAGS := -ldflags "-X github.com/liliang-cn/ohman/pkg/version.Version=$(VERSION)"

.PHONY: build
build:
	go build $(LDFLAGS) -o bin/ohman ./cmd/ohman

.PHONY: build-all
build-all:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/ohman-linux-amd64 ./cmd/ohman
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/ohman-linux-arm64 ./cmd/ohman
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/ohman-darwin-amd64 ./cmd/ohman
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/ohman-darwin-arm64 ./cmd/ohman

.PHONY: install
install: build
	cp bin/ohman /usr/local/bin/

.PHONY: test
test:
	go test -v ./...

### 9.2 GitHub Actions

```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags:
      - "v*"

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.21"
      - name: Build
        run: make build-all
      - name: Release
        uses: softprops/action-gh-release@v1
        with:
          files: bin/*
````

## 10. Future Roadmap

### 10.1 v1.0 Features

- [x] Basic command Q&A
- [x] Failed command diagnosis
- [x] Interactive mode
- [x] Multi-LLM support
- [x] Configuration management

### 10.2 v1.1 Plans

- [ ] Session history persistence
- [ ] Command completion (bash/zsh completion)
- [ ] Offline mode (local LLM)
- [ ] Multi-language support

### 10.3 v2.0 Vision

- [ ] Command recommendation system
- [ ] Learn user preferences
- [ ] Community-shared Q&A library
- [ ] IDE integration plugins
