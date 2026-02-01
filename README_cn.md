# Oh Man! 🤯

> 让 man 页面不再难懂 — AI 驱动的命令行助手

`ohman` 是一个由大语言模型驱动的智能命令行助手。它结合了传统的 `man` 页面和 AI，让你可以用自然语言询问命令用法和参数，甚至在命令失败时获得自动修复建议。

## ✨ 特性

- 🔍 **智能问答**: 用自然语言询问任何命令的使用方法
- 🔧 **失败诊断**: 自动诊断上一个失败的命令并给出修复建议
- 💬 **错误分析**: 直接粘贴错误消息，立即获得分析和解决方案
- 📊 **日志分析**: 分析日志文件、systemd 服务日志或直接粘贴日志内容，获得 AI 驱动的洞察
- 🔌 **管道支持**: 支持通过管道（pipe）传递日志数据，实现实时分析和过滤
- 💭 **交互式聊天**: 启动聊天会话，可选择附上日志进行深入讨论
- 📝 **会话历史**: 使用 `ohman history` 和 `ohman clear` 查看和管理查询历史
- 📡 **流式输出**: 实时流式响应，提供更好的用户体验
- 🎯 **OpenAI 兼容**: 支持任何 OpenAI 兼容的 API（OpenAI、DeepSeek、Ollama 等）
- 🖥️ **跨平台**: 支持 Linux 和 macOS
- 🆘 **智能降级**: 当 man 页面不可用时，自动尝试 `--help` 标志

## 📦 安装

### 使用 Go

```bash
go install github.com/liliang-cn/ohman@latest
```

### 从源码构建

```bash
git clone https://github.com/liliang-cn/ohman.git
cd ohman
make build
sudo make install
```

### 下载预编译二进制文件

从 [Releases](https://github.com/liliang-cn/ohman/releases) 页面下载适合你平台的二进制文件。

## 🚀 快速开始

### 1. 配置 LLM

首次使用需要配置 LLM：

```bash
ohman config
```

或手动编辑配置文件 `~/.config/ohman/config.yaml`：

```yaml
llm:
  api_key: sk-xxx # 你的 API Key
  base_url: "" # 可选，用于 OpenAI 兼容的 API（如 https://api.deepseek.com/v1）
  model: gpt-4o-mini # 模型名称
  max_tokens: 4096 # 最大输出 token 数
  temperature: 0.7 # 温度参数
  timeout: 60 # 请求超时时间（秒）

shell:
  history_file: "" # 留空以自动检测

output:
  color: true # 彩色输出
  markdown: true # Markdown 渲染
```

### 使用不同的服务提供商

**OpenAI（默认）：**

```yaml
llm:
  api_key: sk-xxx
  model: gpt-4o-mini
```

**DeepSeek：**

```yaml
llm:
  api_key: sk-xxx
  base_url: https://api.deepseek.com/v1
  model: deepseek-chat
```

**Azure OpenAI：**

```yaml
llm:
  api_key: your-azure-key
  base_url: https://your-resource.openai.azure.com/openai/deployments/your-deployment
  model: gpt-4
```

**本地 LLM（通过 OpenAI 兼容的服务器，如 LM Studio、带 OpenAI API 的 Ollama）：**

```yaml
llm:
  api_key: not-needed
  base_url: http://localhost:1234/v1
  model: local-model
```

### 2. 开始使用

```bash
# 询问 grep 用法
ohman grep "如何递归搜索？"

# 询问 tar 参数
ohman tar "xvf 参数是什么意思？"

# 命令失败后运行 ohman 获取建议
$ tar -cvf backup  # 哎呀，忘记指定文件了
tar: Cowardly refusing to create an empty archive

$ ohman
# AI 会分析失败的 tar 命令并告诉你正确的用法
```

## 📖 使用指南

### 基本命令格式

```
ohman [命令] [问题]
```

| 参数     | 描述                                                         |
| -------- | ------------------------------------------------------------ |
| `命令`   | 要查询的命令名称（如 grep、tar、find）                        |
| `问题`   | 你的问题（可选，如果省略则进入交互模式）                     |

### 使用场景

#### 场景 1：询问命令参数

```bash
ohman find "如何查找 7 天前修改的文件并删除？"
```

#### 场景 2：理解复杂命令

```bash
ohman awk "解释 NR 和 NF 变量"
```

#### 场景 3：交互式探索

```bash
ohman git
# 进入交互模式持续提问
> 如何撤销最后一次提交？
> 如何查看文件的修改历史？
> exit
```

#### 场景 4：失败命令诊断

```bash
$ chmod 777 /etc/passwd
chmod: changing permissions of '/etc/passwd': Operation not permitted

$ ohman
# AI 会解释为什么失败并提供正确的方法
```

#### 场景 5：直接错误消息分析

```bash
# 只需粘贴任何错误消息
ohman "error: failed to push some refs to 'https://github.com/user/repo.git'"

ohman "bash: ./script.sh: Permission denied"

ohman "segmentation fault (core dumped)"
# AI 会检测错误关键词并提供解决方案
```

#### 场景 6：日志分析

```bash
# 分析日志文件
ohman log /var/log/app/error.log

# 分析并限制条目数
ohman log -n 100 /var/log/nginx/access.log

# 直接粘贴日志内容
ohman log "2025-02-01 10:23:45 [ERROR] Database connection timeout
2025-02-01 10:23:46 [WARN] Retrying connection..."

# AI 会分析日志，识别错误，并提供解决方案
```

#### 场景 7：Systemd 服务日志分析

```bash
# 分析 systemd 服务日志
ohman log -u nginx.service

# 限制分析条目数
ohman log -u docker -n 50

# 使用完整格式
ohman log --unit mysql.service

# AI 会分析 journalctl 日志并提供深入洞察
```

#### 场景 8：管道支持（新功能！）

```bash
# 通过管道分析日志
tail -f /var/log/app.log | ohman log

# 过滤后分析
grep ERROR /var/log/app.log | ohman log

# 分析最近的错误
tail -n 100 /var/log/syslog | ohman log

# 与 journalctl 配合使用
journalctl -u nginx | ohman log

# 链接多个命令
journalctl -f -u docker | grep ERROR | ohman log

# 实时监控
tail -f /var/log/app.log | grep ERROR | ohman log
```

#### 场景 9：会话管理

```bash
# 查看查询历史
ohman history

# 清空所有历史
ohman clear
```

#### 场景 10：交互式聊天（新功能！）

```bash
# 启动通用聊天会话
ohman chat

# 带日志上下文的聊天
ohman chat /var/log/app/error.log

# 追问关于日志的问题
> 发生了什么错误？
> 超时是什么原因导致的？
> 如何修复这个问题？
> exit
```

### 高级用法

#### 指定 Man 章节

```bash
ohman -s 3 printf "C 语言 printf 格式化字符串"
```

#### 查看原始 Man 内容

```bash
ohman --raw grep
```

#### 指定 LLM 模型

```bash
ohman --model gpt-4 find "复杂查询"
```

## 🔧 命令参考

```
ohman - AI 驱动的 man 页面助手

用法:
  ohman [命令] [问题]
  ohman [命令]

可用命令:
  chat        启动交互式聊天会话
  clear       清空会话缓存
  completion  为指定的 shell 生成自动补全脚本
  config      配置 ohman
  help        关于任何命令的帮助
  history     查看会话历史
  log         分析日志文件、systemd 服务日志或通过管道接收日志内容

标志:
  -c, --config string   配置文件路径
  -h, --help            ohman 的帮助
  -i, --interactive     强制交互模式
  -m, --model string    LLM 模型名称
  -r, --raw             仅显示原始 man 内容
  -s, --section int     man 页面章节 (1-8)
  -v, --verbose         详细输出
      --version         ohman 的版本

示例:
  ohman grep "如何只显示匹配的文件名？"
  ohman -s 5 passwd "配置文件格式是什么？"
  ohman -c /path/to/config.yaml tar "如何解压？"
  ohman log /var/log/app/error.log
  ohman log -n 50 "2025-02-01 ERROR: Connection timeout"
  ohman log -u nginx.service
  ohman log --unit docker -n 100
  tail -f /var/log/app.log | ohman log
  ohman tar
  ohman
```

## 🏗️ 工作原理

```
┌─────────────────────────────────────────────────────────────────┐
│                         ohman 工作流程                           │
├─────────────────────────────────────────────────────────────────┤
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐  │
│  │  用户    │───>│  解析    │───>│  获取    │───>│  构建    │  │
│  │  输入    │    │  参数    │    │  日志    │    │  提示    │  │
│  └──────────┘    └──────────┘    └──────────┘    └──────────┘  │
│                                         │              │        │
│                                         │              ▼        │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐  │
│  │  渲染    │<───│  解析    │<───│  LLM     │<───│  发送    │  │
│  │  输出    │    │  响应    │    │  API     │    │  请求    │  │
│  └──────────┘    └──────────┘    └──────────┘    └──────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

### 日志分析流程

ohman 的日志分析支持三种输入方式：

1. **文件分析**
   ```
   ohman log /var/log/app/error.log
   ```
   - 读取指定的日志文件
   - 解析日志级别（DEBUG、INFO、WARN、ERROR、FATAL）
   - 自动识别日志类型（应用、系统、访问、错误日志）
   - 统计各级别数量
   - 提取样本供 AI 分析

2. **systemd Journalctl 分析**
   ```
   ohman log -u nginx.service
   ```
   - 调用 `journalctl -u <unit>` 获取服务日志
   - 支持常见的 systemd 单元类型
   - 可使用 `-n` 参数限制条目数
   - 智能识别服务名称（如 nginx、docker、mysql）

3. **管道输入**
   ```
   tail -f /var/log/app.log | ohman log
   ```
   - 自动检测管道输入
   - 从 stdin 读取数据
   - 支持与 grep、awk、sed 等命令组合
   - 适合实时监控和过滤场景

### 日志分析能力

AI 会自动识别：
- **错误模式** - 重复的错误和根本原因
- **警告趋势** - 潜在问题的早期信号
- **性能问题** - 超时、资源耗尽等
- **配置错误** - 配置文件或参数问题
- **依赖问题** - 数据库连接、API 调用失败等

并提供：
- **问题诊断** - 清晰的问题描述
- **根本原因** - 分析问题的真正原因
- **具体解决方案** - 可执行的修复命令和配置
- **预防措施** - 避免类似问题的建议

### 错误检测规则

`ohman` 自动检测输入是否为错误消息：

1. **多行输入** - 如果粘贴了多行（包含 `\n`）
2. **错误关键词** - 包含常见错误指示符：
   - `error:`, `failed`, `cannot`, `permission denied`
   - `no such file`, `command not found`
   - `segmentation fault`, `core dumped`
   - `fatal`, `exception`, `undefined`, `not found`
   - `connection refused`, `timeout`
3. **长输入** - 输入超过 150 字符（可能是粘贴的输出）

### 会话历史

所有会话自动保存到 `~/.config/ohman/history.json`：

- **question**: 直接问答
- **diagnose**: 失败命令诊断
- **error**: 错误消息分析
- **interactive**: 交互模式会话

最多保存 100 条最近记录。

## 🔐 隐私与安全

- 📤 **发送的数据**: 仅将 man 页面内容、日志内容和你的问题发送到 LLM API
- 🔒 **API 密钥安全**: API 密钥存储在本地配置文件中，权限为 600
- 🚫 **无遥测**: 不收集任何使用数据
- 💻 **本地选项**: 支持任何 OpenAI 兼容的本地 LLM 服务器（LM Studio、Ollama with OpenAI API 等）

### 日志数据隐私

- 日志内容仅发送到你配置的 LLM API，不会存储或分享
- 不从你的系统读取敏感信息（除非显式指定日志文件）
- 支持使用本地 LLM 进行完全离线的日志分析

## 🤝 贡献

欢迎贡献！详见 [CONTRIBUTING.md](CONTRIBUTING.md)。

## 📜 许可证

MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。

---

## 💡 高级日志分析技巧

### 实时监控

```bash
# 实时监控应用错误
tail -f /var/log/app.log | grep ERROR | ohman log

# 监控多个服务的错误
journalctl -f -u nginx -u docker | grep -i error | ohman log

# 结合系统日志
tail -f /var/log/syslog /var/log/app.log | grep -E "ERROR|FATAL" | ohman log
```

### 批量分析

```bash
# 分析最近 1000 行的错误
tail -n 1000 /var/log/app.log | ohman log

# 分析所有日志文件中的错误
cat /var/log/app/*.log | grep CRITICAL | ohman log

# 按日期筛选后分析
grep "2025-02-01" /var/log/app.log | ohman log
```

### 与其他工具组合

```bash
# 使用 awk 过滤特定字段
awk '/ERROR/ {print $0}' /var/log/app.log | ohman log

# 使用 sed 清理格式后分析
sed 's/\[.*\]//' /var/log/app.log | ohman log

# 使用 jq 解析 JSON 日志
cat app.log | jq -r '. | select(.level=="ERROR") | .message' | ohman log
```

### Systemd 日志高级用法

```bash
# 查看今天的错误
journalctl --since today -u nginx | ohman log

# 查看特定时间范围
journalctl --since "1 hour ago" -u docker | ohman log

# 查看引导过程的日志
journalctl -b | ohman log

# 组合多个单元
journalctl -u nginx -u mysql | ohman log
```

## ❓ 常见问题

### Q: ohman log 和 journalctl 有什么区别？

A: `ohman log` 可以直接分析 journalctl 的输出，但提供了：
- AI 驱动的智能分析
- 自动错误识别和分类
- 可执行的解决方案和命令
- 与其他日志文件统一的分析接口

### Q: 支持哪些日志格式？

A: ohman 支持多种日志格式：
- 标准时间戳格式：`2025-02-01 10:23:45`
- 括号级别格式：`[ERROR] Failed...`
- Systemd 格式：`Feb 01 10:23:45 host service[123]: ...`
- 自定义格式：自动识别错误级别和消息

### Q: 管道输入的优先级如何？

A: 管道输入具有最高优先级：
1. 管道输入（如果有）
2. `-u/--unit` 参数（systemd 服务）
3. 文件路径（如果文件存在）
4. 日志内容字符串

### Q: 可以分析多大的日志文件？

A: ohman 可以处理任意大小的日志文件：
- 默认分析全部内容
- 使用 `-n` 参数限制分析的条目数
- 大文件建议使用管道过滤后再分析：
  ```bash
  tail -n 1000 /var/log/app.log | ohman log
  ```

### Q: 日志分析会修改原文件吗？

A: 不会。ohman 只是读取和分析日志内容，不会：
- 修改原始日志文件
- 创建额外的日志文件
- 存储日志内容到其他位置

### Q: 如何确保敏感信息不被发送？

A: 使用本地 LLM 或过滤敏感信息：
```bash
# 过滤敏感信息后再分析
sed 's/password=****/password=[REDACTED]/' app.log | ohman log

# 或使用本地 LLM 完全离线分析
ohman log app.log --model local-model
```

## 📚 更多资源

- [设计文档](docs/DESIGN.md) - 技术架构和设计决策
- [使用指南](docs/USAGE.md) - 详细的使用说明和示例
- [更新日志](CHANGELOG.md) - 版本历史和更新内容

---

**Oh Man!** - 让读 man 页和查日志不再痛苦 🎉
