# Contributing to Claude Agent SDK for Go

Thank you for your interest in contributing! This document provides guidelines and information for contributors.

## Code of Conduct

Please be respectful and constructive in all interactions. We are committed to providing a welcoming and inclusive environment.

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/claude-agent-sdk-go.git
   cd claude-agent-sdk-go
   ```
3. **Add the upstream remote**:
   ```bash
   git remote add upstream https://github.com/panbanda/claude-agent-sdk-go.git
   ```

## Development Setup

### Prerequisites

- Go 1.24 or later
- [golangci-lint](https://golangci-lint.run/usage/install/) for linting

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with race detection
go test -race ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Running Linter

```bash
golangci-lint run
```

## Making Changes

### Branch Naming

Use descriptive branch names:
- `feature/add-streaming-support`
- `fix/connection-timeout`
- `docs/update-readme`

### Commit Messages

Follow conventional commit format:
```
type(scope): description

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Adding or updating tests
- `refactor`: Code refactoring
- `chore`: Maintenance tasks

Examples:
```
feat(client): add support for streaming responses
fix(transport): handle connection timeout correctly
docs(readme): add usage examples
test(hooks): add tests for PreToolUse hook
```

### Code Style

This project follows standard Go conventions:

1. **Format code** with `gofmt` or `goimports`
2. **Follow** [Effective Go](https://go.dev/doc/effective_go) guidelines
3. **Use meaningful names** - complete words, not abbreviations
4. **Write tests** for new functionality
5. **Add comments** only when the code's purpose isn't obvious

### Testing Requirements

- All new features must include tests
- All bug fixes should include a test that would have caught the bug
- Maintain or improve code coverage
- Tests should be deterministic and not depend on external services

## Pull Request Process

1. **Update your fork** with the latest upstream changes:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature
   ```

3. **Make your changes** and commit them

4. **Push to your fork**:
   ```bash
   git push origin feature/your-feature
   ```

5. **Open a Pull Request** on GitHub

### PR Checklist

- [ ] Tests pass locally (`go test ./...`)
- [ ] Linter passes (`golangci-lint run`)
- [ ] New code has tests
- [ ] Documentation updated if needed
- [ ] Commit messages follow conventions
- [ ] PR description explains the changes

### PR Review

- PRs require at least one approval before merging
- Address review feedback promptly
- Keep PRs focused - one feature or fix per PR
- Large changes should be discussed in an issue first

## Reporting Issues

### Bug Reports

Include:
- Go version (`go version`)
- Operating system
- Claude CLI version (`claude --version`)
- Steps to reproduce
- Expected vs actual behavior
- Relevant error messages or logs

### Feature Requests

Include:
- Use case description
- Proposed API (if applicable)
- Alternatives considered

## Project Structure

```
claude-agent-sdk-go/
├── claude/              # Main package
│   ├── client.go        # Client implementation
│   ├── client_test.go   # Client tests
│   ├── content.go       # Content block types
│   ├── control.go       # Control protocol
│   ├── errors.go        # Error types
│   ├── hooks.go         # Hook system
│   ├── message.go       # Message types
│   ├── options.go       # Functional options
│   ├── query.go         # Query functions
│   ├── subprocess.go    # Subprocess transport
│   └── transport.go     # Transport interface
├── .github/workflows/   # CI configuration
├── go.mod               # Module definition
├── LICENSE              # MIT license
├── README.md            # Documentation
└── CONTRIBUTING.md      # This file
```

## Questions?

Feel free to open an issue for questions or discussions about the project.

Thank you for contributing!
