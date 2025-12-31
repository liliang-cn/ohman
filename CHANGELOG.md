# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planned

- Session history persistence
- Command completion (bash/zsh completion)
- More LLM provider support

## [0.1.0] - 2024-XX-XX

### Added

- 🎉 Initial release
- ✨ Command Q&A feature: `ohman <command> "<question>"`
- ✨ Interactive mode: `ohman <command>`
- ✨ Failed command diagnosis: `ohman` (no arguments)
- ✨ Configuration wizard: `ohman config`
- 🔌 OpenAI API support
- 🔌 Anthropic API support
- 🔌 Ollama local model support
- 🔌 Custom OpenAI-compatible API support
- 📝 Shell hook support (Zsh/Bash)
- 🌐 Linux and macOS support
- 📖 Complete documentation

### Technical Details

- Developed with Go 1.21
- CLI built with Cobra
- YAML configuration file support
- Environment variable override support

---

## Version Notes

- **Major**: Incompatible API changes
- **Minor**: Backward-compatible feature additions
- **Patch**: Backward-compatible bug fixes

[Unreleased]: https://github.com/liliang-cn/ohman/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/liliang-cn/ohman/releases/tag/v0.1.0
