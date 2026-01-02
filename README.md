# Oh Man! 🤯

> Making man pages less cryptic — AI-powered command line help

`ohman` is an intelligent command-line assistant powered by LLM. It combines traditional `man` pages with AI, allowing you to ask questions about command usage and parameters in natural language, and even get automatic fix suggestions when commands fail.

## ✨ Features

- 🔍 **Smart Q&A**: Ask questions about any command in natural language
- 🔧 **Failure Diagnosis**: Automatically diagnose the last failed command and suggest fixes
- 💬 **Error Analysis**: Paste error messages directly for instant analysis and solutions
- 📝 **Session History**: View and manage your query history with `ohman history` and `ohman clear`
- 📡 **Streaming Output**: Real-time streaming responses for better user experience
- 🎯 **OpenAI Compatible**: Works with any OpenAI-compatible API (OpenAI, DeepSeek, Ollama, etc.)
- 🖥️ **Cross-Platform**: Supports Linux and macOS
- 🆘 **Smart Fallback**: Tries `--help` flag when man page is not available

## 📦 Installation

### Using Go

```bash
go install github.com/liliang-cn/ohman@latest
```

### Build from Source

```bash
git clone https://github.com/liliang-cn/ohman.git
cd ohman
make build
sudo make install
```

### Download Pre-built Binary

Download the binary for your platform from the [Releases](https://github.com/liliang-cn/ohman/releases) page.

## 🚀 Quick Start

### 1. Configure LLM

First-time setup requires LLM configuration:

```bash
ohman config
```

Or manually edit the config file `~/.config/ohman/config.yaml`:

```yaml
llm:
  api_key: sk-xxx # Your API Key
  base_url: "" # Optional, for OpenAI-compatible APIs (e.g., https://api.deepseek.com/v1)
  model: gpt-4o-mini # Model name
  max_tokens: 4096 # Maximum output tokens
  temperature: 0.7 # Temperature parameter
  timeout: 60 # Request timeout in seconds

shell:
  history_file: "" # Leave empty for auto-detection

output:
  color: true # Colored output
  markdown: true # Markdown rendering
```

### Using with Different Providers

**OpenAI (default):**

```yaml
llm:
  api_key: sk-xxx
  model: gpt-4o-mini
```

**DeepSeek:**

```yaml
llm:
  api_key: sk-xxx
  base_url: https://api.deepseek.com/v1
  model: deepseek-chat
```

**Azure OpenAI:**

```yaml
llm:
  api_key: your-azure-key
  base_url: https://your-resource.openai.azure.com/openai/deployments/your-deployment
  model: gpt-4
```

**Local LLM (via OpenAI-compatible server like LM Studio, Ollama with OpenAI API):**

```yaml
llm:
  api_key: not-needed
  base_url: http://localhost:1234/v1
  model: local-model
```

### 2. Start Using

```bash
# Ask about grep usage
ohman grep "How to search recursively?"

# Ask about tar parameters
ohman tar "What do the xvf parameters mean?"

# After a command fails, run ohman to get suggestions
$ tar -cvf backup  # Oops, forgot to specify files
tar: Cowardly refusing to create an empty archive

$ ohman
# AI will analyze the failed tar command and tell you the correct usage
```

## 📖 Usage Guide

### Basic Command Format

```
ohman [command] [question]
```

| Parameter  | Description                                                  |
| ---------- | ------------------------------------------------------------ |
| `command`  | The command name to query (e.g., grep, tar, find)            |
| `question` | Your question (optional, enters interactive mode if omitted) |

### Use Cases

#### Case 1: Ask about Command Parameters

```bash
ohman find "How to find files modified 7 days ago and delete them?"
```

#### Case 2: Understand Complex Commands

```bash
ohman awk "Explain the NR and NF variables"
```

#### Case 3: Interactive Exploration

```bash
ohman git
# Enter interactive mode for continuous questions
> How to undo the last commit?
> How to view the modification history of a file?
> exit
```

#### Case 4: Failed Command Diagnosis

```bash
$ chmod 777 /etc/passwd
chmod: changing permissions of '/etc/passwd': Operation not permitted

$ ohman
# AI will explain why it failed and provide the correct approach
```

#### Case 5: Direct Error Message Analysis (New!)

```bash
# Simply paste any error message
ohman "error: failed to push some refs to 'https://github.com/user/repo.git'"

ohman "bash: ./script.sh: Permission denied"

ohman "segmentation fault (core dumped)"
# AI detects error keywords and provides solutions
```

#### Case 6: Session Management

```bash
# View your query history
ohman history

# Clear all history
ohman clear
```

### Advanced Usage

#### Specify Man Section

```bash
ohman -s 3 printf "C language printf format string"
```

#### View Raw Man Content

```bash
ohman --raw grep
```

#### Specify LLM Model

```bash
ohman --model gpt-4 find "Complex query"
```

## 🔧 Command Reference

```
ohman - AI-powered man page assistant

Usage:
  ohman [command] [question]
  ohman [command]

Available Commands:
  clear       Clear session cache
  completion  Generate the autocompletion script for the specified shell
  config      Configure ohman
  help        Help about any command
  history     View session history

Flags:
  -c, --config string   config file path
  -h, --help            help for ohman
  -i, --interactive     force interactive mode
  -m, --model string    LLM model name
  -r, --raw             show raw man content only
  -s, --section int     man page section (1-8)
  -v, --verbose         verbose output
      --version         version for ohman

Examples:
  ohman grep "How to only show matching filenames?"
  ohman -s 5 passwd "What's the config file format?"
  ohman -c /path/to/config.yaml tar "How to extract?"
  ohman tar
  ohman
```

## 🏗️ How It Works

```
┌─────────────────────────────────────────────────────────────────┐
│                         ohman Workflow                           │
├─────────────────────────────────────────────────────────────────┤
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐  │
│  │  User    │───>│  Parse   │───>│  Get     │───>│  Build   │  │
│  │  Input   │    │  Args    │    │  Man     │    │  Prompt  │  │
│  └──────────┘    └──────────┘    └──────────┘    └──────────┘  │
│                                         │              │        │
│                                         │              ▼        │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐  │
│  │  Render  │<───│  Parse   │<───│  LLM     │<───│  Send    │  │
│  │  Output  │    │  Response│    │  API     │    │  Request │  │
│  └──────────┘    └──────────┘    └──────────┘    └──────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

## 🔐 Privacy & Security

- 📤 **Data Sent**: Only man page content and your questions are sent to the LLM API
- 🔒 **API Key Security**: API Key is stored in local config file with 600 permissions
- 🚫 **No Telemetry**: No usage data is collected
- 💻 **Local Option**: Supports any OpenAI-compatible local LLM server (LM Studio, Ollama with OpenAI API, etc.)

## 🤝 Contributing

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## 📜 License

MIT License - See [LICENSE](LICENSE) file for details.

---

**Oh Man!** - Because reading man pages shouldn't be this painful 🎉
