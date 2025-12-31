# Oh Man! User Guide

This document provides detailed usage instructions and best practices for Oh Man!

## Table of Contents

- [Quick Start](#quick-start)
- [Command Reference](#command-reference)
- [Configuration](#configuration)
- [Shell Hooks](#shell-hooks)
- [FAQ](#faq)
- [Advanced Usage](#advanced-usage)

## Quick Start

### 1. Installation

```bash
# Install with Go
go install github.com/liliang-cn/ohman@latest

# Or use the install script
curl -sSL https://raw.githubusercontent.com/liliang-cn/ohman/main/scripts/install.sh | bash
```

### 2. Configure LLM

```bash
ohman config
```

Follow the prompts to select an LLM provider and enter your API Key.

### 3. Start Using

```bash
# Ask about command usage
ohman grep "How do I ignore case?"

# Diagnose a failed command
ohman
```

## Command Reference

### Basic Syntax

```
ohman [flags] [command] [question]
```

### Usage Modes

#### Mode 1: Command Q&A

```bash
ohman <command> "<question>"
```

Examples:

```bash
ohman find "How do I find files larger than 100MB?"
ohman awk "What do the NR and NF variables mean?"
ohman sed "How do I replace file contents in place?"
```

#### Mode 2: Interactive Mode

```bash
ohman <command>
```

After entering interactive mode, you can ask multiple questions:

```bash
$ ohman git
📖 Loaded man page for git, entering interactive mode
   Ask questions, type 'exit' or 'quit' to exit

❓ How do I view commit history?
[AI response...]

❓ How do I revert to the previous version?
[AI response...]

❓ exit
👋 Goodbye!
```

#### Mode 3: Failed Command Diagnosis

After a command fails, simply run `ohman` (with no arguments):

```bash
$ chmod 777 /etc/passwd
chmod: changing permissions of '/etc/passwd': Operation not permitted

$ ohman
🔍 Detected failed command: chmod 777 /etc/passwd
   Exit code: 1

🔧 Analyzing...

## 🔍 Problem Analysis
This command failed because /etc/passwd is a critical system file...

## ✅ Solution
...
```

### Command Flags

| Flag            | Short | Description                    |
| --------------- | ----- | ------------------------------ |
| `--section`     | `-s`  | Specify man page section (1-8) |
| `--model`       | `-m`  | Temporarily specify LLM model  |
| `--raw`         | `-r`  | Show raw man content only      |
| `--interactive` | `-i`  | Force interactive mode         |
| `--config`      | `-c`  | Specify config file path       |
| `--verbose`     | `-v`  | Verbose output mode            |
| `--help`        | `-h`  | Show help information          |
| `--version`     |       | Show version information       |

## Configuration

Config file location: `~/.config/ohman/config.yaml`

### Complete Configuration Example

```yaml
# LLM Configuration
llm:
  provider: openai # openai, anthropic, ollama, custom
  api_key: sk-xxx # API Key
  base_url: "" # Custom endpoint (optional)
  model: gpt-4o-mini # Model name
  max_tokens: 4096 # Maximum output tokens
  temperature: 0.7 # Temperature (0-2)
  timeout: 60 # Timeout in seconds

# Shell Configuration
shell:
  history_file: "" # History file path (leave empty for auto-detect)
  auto_install_hook: true # Auto install hook

# Output Configuration
output:
  color: true # Colored output
  markdown: true # Markdown rendering
  language: en-US # Language

# Debug Configuration
debug:
  enabled: false
  show_prompt: false
  show_tokens: false
```

### Environment Variables

| Variable        | Description                                 |
| --------------- | ------------------------------------------- |
| `OHMAN_CONFIG`  | Config file path                            |
| `OHMAN_API_KEY` | API Key (takes precedence over config file) |

### Supported LLM Providers

#### OpenAI

```yaml
llm:
  provider: openai
  api_key: sk-xxx
  model: gpt-4o-mini # or gpt-4o, gpt-4-turbo
```

#### Anthropic

```yaml
llm:
  provider: anthropic
  api_key: sk-ant-xxx
  model: claude-3-5-sonnet-20241022
```

#### Ollama (Local)

```yaml
llm:
  provider: ollama
  base_url: http://localhost:11434
  model: llama3 # or other installed models
```

#### Custom (OpenAI API Compatible)

```yaml
llm:
  provider: custom
  api_key: your-key
  base_url: https://your-api-endpoint.com/v1
  model: your-model
```

## Shell Hooks

To enable automatic failed command detection, you need to install a shell hook.

### Automatic Installation

The install script will automatically add the hook, or run:

```bash
ohman config  # Follow the prompts
```

### Manual Installation

#### Zsh

Add to `~/.zshrc`:

```zsh
# Oh Man! Failed command recording hook
ohman_precmd() {
    local exit_code=$?
    if [[ $exit_code -ne 0 ]]; then
        echo "$exit_code|$(fc -ln -1)|$(date +%s)" > /tmp/.ohman_last_failed_$$
    fi
}
precmd_functions+=(ohman_precmd)
```

#### Bash

Add to `~/.bashrc`:

```bash
# Oh Man! Failed command recording hook
ohman_prompt_command() {
    local exit_code=$?
    if [[ $exit_code -ne 0 ]]; then
        echo "$exit_code|$(history 1 | sed 's/^[ ]*[0-9]*[ ]*//')|$(date +%s)" > /tmp/.ohman_last_failed_$$
    fi
}
PROMPT_COMMAND="ohman_prompt_command${PROMPT_COMMAND:+; $PROMPT_COMMAND}"
```

After installation, reload your configuration:

```bash
source ~/.zshrc  # or source ~/.bashrc
```

## FAQ

### Q: Got "API Key not configured" error

Run `ohman config` to configure your API Key, or set an environment variable:

```bash
export OHMAN_API_KEY="your-api-key"
```

### Q: Man page not found

1. Verify the command name is correct
2. Try specifying a section: `ohman -s 1 command`
3. Some commands may not have man pages

### Q: Failed command diagnosis not working

1. Confirm the shell hook is installed
2. Confirm you've reloaded your shell configuration
3. The hook only records failed commands from the last 5 minutes

### Q: Response is too slow

1. Check your network connection
2. Try using a faster model (e.g., gpt-4o-mini)
3. Consider using local Ollama

### Q: How do I use a proxy?

Set environment variables:

```bash
export HTTP_PROXY=http://proxy:port
export HTTPS_PROXY=http://proxy:port
```

## Advanced Usage

### Piping

```bash
echo "How do I search recursively?" | ohman grep
```

### Script Integration

```bash
#!/bin/bash
# Automatically query command help
result=$(ohman tar "How do I extract .tar.gz" 2>&1)
echo "$result"
```

### Combining with fzf

```bash
# Interactively select a command and query
ls /usr/bin | fzf | xargs -I {} ohman {}
```

### Batch Queries

```bash
for cmd in grep sed awk; do
    echo "=== $cmd ==="
    ohman "$cmd" "What are the most commonly used options?"
    echo
done
```
