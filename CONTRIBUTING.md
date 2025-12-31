# Contributing Guide

Thank you for your interest in the Oh Man! project! We welcome contributions of all kinds.

## How to Contribute

### Reporting Bugs

1. Verify the bug hasn't been reported (check [Issues](https://github.com/liliang-cn/ohman/issues))
2. Create a new issue using the Bug template
3. Provide detailed reproduction steps, system information, and error messages

### Suggesting Features

1. Check if a similar suggestion exists
2. Create a new issue using the Feature Request template
3. Clearly describe the purpose and expected behavior

### Submitting Code

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Write code and tests
4. Ensure tests pass: `make test`
5. Ensure code style is correct: `make lint`
6. Commit changes: `git commit -m 'Add amazing feature'`
7. Push branch: `git push origin feature/amazing-feature`
8. Create a Pull Request

## Development Guide

### Requirements

- Go 1.21+
- make

### Local Development

```bash
# Clone repository
git clone https://github.com/liliang-cn/ohman.git
cd ohman

# Install dependencies
go mod download

# Build
make build

# Run tests
make test

# Code linting
make lint
```

### Project Structure

```
ohman/
├── cmd/ohman/      # Program entry point
├── internal/       # Internal packages
│   ├── app/        # Application logic
│   ├── cli/        # Command line handling
│   ├── config/     # Configuration management
│   ├── llm/        # LLM clients
│   ├── man/        # Man page processing
│   ├── output/     # Output rendering
│   └── shell/      # Shell history
├── pkg/            # Public packages
└── scripts/        # Scripts
```

### Code Style

- Follow [Effective Go](https://golang.org/doc/effective_go)
- Use `gofmt` to format code
- Use meaningful variable and function names
- Add necessary comments

### Commit Messages

Use conventional commit format:

```
type(scope): description

[optional body]

[optional footer]
```

Types:

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting)
- `refactor`: Code refactoring
- `test`: Adding or modifying tests
- `chore`: Build process or auxiliary tool changes

### Testing

- Write unit tests for new features
- Ensure existing tests still pass
- Aim for at least 80% coverage for new code

## Code Review

All submissions require review. We use GitHub pull requests for this purpose.

### Review Criteria

- Code correctness
- Test coverage
- Documentation completeness
- Code style consistency

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
