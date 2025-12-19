# Claude Agent SDK for Go

<p align="center">
  <img src="assets/header-image.jpg" alt="Claude Agent SDK for Go" width="100%">
</p>

<p align="center">
  <a href="https://pkg.go.dev/github.com/panbanda/claude-agent-sdk-go"><img src="https://pkg.go.dev/badge/github.com/panbanda/claude-agent-sdk-go.svg" alt="Go Reference"></a>
  <a href="https://goreportcard.com/report/github.com/panbanda/claude-agent-sdk-go"><img src="https://goreportcard.com/badge/github.com/panbanda/claude-agent-sdk-go" alt="Go Report Card"></a>
  <a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License: MIT"></a>
  <a href="https://go.dev/"><img src="https://img.shields.io/github/go-mod/go-version/panbanda/claude-agent-sdk-go" alt="Go Version"></a>
  <a href="https://github.com/panbanda/claude-agent-sdk-go/actions/workflows/test.yml"><img src="https://github.com/panbanda/claude-agent-sdk-go/actions/workflows/test.yml/badge.svg" alt="Tests"></a>
  <a href="https://codecov.io/gh/panbanda/claude-agent-sdk-go"><img src="https://codecov.io/gh/panbanda/claude-agent-sdk-go/branch/main/graph/badge.svg" alt="Coverage"></a>
</p>

<p align="center">
  The official Go SDK for building AI agents with <a href="https://docs.anthropic.com/en/docs/claude-code">Claude Code</a>.<br>
  A clean, idiomatic Go interface for interacting with Claude's agentic capabilities.
</p>

## Features

- **Idiomatic Go API** - Functional options, channels, and context support
- **Streaming Messages** - Real-time message streaming via Go channels
- **Hook System** - Intercept and modify tool usage with typed hooks
- **Full Type Safety** - Strongly typed messages, content blocks, and options
- **Subprocess Transport** - Manages Claude CLI lifecycle automatically
- **Control Protocol** - Full support for SDK-CLI communication

## Requirements

- Go 1.21 or later
- [Claude Code CLI](https://docs.anthropic.com/en/docs/claude-code) installed

## Installation

```bash
go get github.com/panbanda/claude-agent-sdk-go
```

## Quick Start

### Simple Query

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/panbanda/claude-agent-sdk-go/claude"
)

func main() {
    ctx := context.Background()

    // One-shot query with streaming
    msgs, err := claude.Query(ctx, "What is 2+2?",
        claude.WithModel("claude-sonnet-4-5"),
    )
    if err != nil {
        log.Fatal(err)
    }

    for msg := range msgs {
        switch m := msg.(type) {
        case *claude.AssistantMessage:
            for _, block := range m.Content {
                if block.IsText() {
                    fmt.Println(block.Text)
                }
            }
        case *claude.ResultMessage:
            fmt.Printf("Cost: $%.4f\n", m.TotalCostUSD)
        }
    }
}
```

### Get Final Result Only

```go
result, err := claude.QueryResult(ctx, "Explain quantum computing",
    claude.WithModel("claude-sonnet-4-5"),
    claude.WithMaxTurns(5),
)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Response: %s\n", result.Result)
fmt.Printf("Cost: $%.4f\n", result.TotalCostUSD)
```

### Interactive Client

```go
client := claude.NewClient(
    claude.WithModel("claude-sonnet-4-5"),
    claude.WithMaxTurns(10),
    claude.WithSystemPrompt("You are a helpful coding assistant."),
)

if err := client.Connect(ctx); err != nil {
    log.Fatal(err)
}
defer client.Close()

// Send a query
if err := client.Query(ctx, "Write a hello world in Go"); err != nil {
    log.Fatal(err)
}

// Stream responses
for msg := range client.Messages() {
    switch m := msg.(type) {
    case *claude.AssistantMessage:
        // Handle assistant response
    case *claude.SystemMessage:
        // Handle system events
    case *claude.ResultMessage:
        // Query complete
    }
}
```

## Configuration Options

### Model Settings

```go
claude.WithModel("claude-sonnet-4-5")      // Primary model
claude.WithFallbackModel("claude-haiku")   // Fallback if primary unavailable
```

### Limits

```go
claude.WithMaxTurns(10)           // Maximum conversation turns
claude.WithMaxBudgetUSD(1.0)      // Spending limit in USD
claude.WithMaxThinkingTokens(500) // Token budget for extended thinking
```

### Permissions

```go
claude.WithPermissionMode(claude.PermissionDefault)     // Default prompting
claude.WithPermissionMode(claude.PermissionAcceptEdits) // Auto-accept edits
claude.WithPermissionMode(claude.PermissionPlan)        // Plan mode
claude.WithPermissionMode(claude.PermissionBypass)      // Bypass all checks
```

### Tools

```go
claude.WithAllowedTools("Read", "Write", "Bash")  // Whitelist tools
claude.WithDisallowedTools("Bash")                // Blacklist tools
```

### Environment

```go
claude.WithWorkingDir("/path/to/project")         // Working directory
claude.WithCLIPath("/custom/path/to/claude")      // Custom CLI path
claude.WithEnv(map[string]string{"KEY": "value"}) // Environment variables
```

### Sessions

```go
claude.WithContinueConversation(true)  // Continue prior conversation
claude.WithResume("session-id")        // Resume specific session
```

## Hooks

Hooks allow you to intercept and modify Claude's behavior at key points.

### Pre-Tool Use Hook

Intercept tool calls before execution:

```go
client := claude.NewClient(
    claude.WithPreToolUseHook("Bash", func(ctx context.Context, input *claude.PreToolUseInput, hookCtx *claude.HookContext) (*claude.HookOutput, error) {
        // Block dangerous commands
        if cmd, ok := input.ToolInput["command"].(string); ok {
            if strings.Contains(cmd, "rm -rf") {
                return &claude.HookOutput{
                    Decision: claude.HookDecisionDeny,
                    Reason:   "Dangerous command blocked",
                }, nil
            }
        }
        return &claude.HookOutput{Decision: claude.HookDecisionAllow}, nil
    }),
)
```

### Post-Tool Use Hook

React to tool results:

```go
claude.WithPostToolUseHook("", func(ctx context.Context, input *claude.PostToolUseInput, hookCtx *claude.HookContext) (*claude.HookOutput, error) {
    if input.IsError {
        log.Printf("Tool %s failed: %v", input.ToolName, input.ToolResponse)
    }
    return &claude.HookOutput{}, nil
})
```

### Available Hook Events

| Event | Description |
|-------|-------------|
| `PreToolUse` | Before tool execution |
| `PostToolUse` | After tool execution |
| `UserPromptSubmit` | When user sends a prompt |
| `Stop` | When agent stops |
| `SubagentStop` | When a subagent stops |
| `PreCompact` | Before conversation compaction |

## Message Types

### AssistantMessage

Contains Claude's response with content blocks:

```go
case *claude.AssistantMessage:
    for _, block := range m.Content {
        switch {
        case block.IsText():
            fmt.Println(block.Text)
        case block.IsThinking():
            fmt.Printf("Thinking: %s\n", block.Thinking)
        case block.IsToolUse():
            fmt.Printf("Using tool: %s\n", block.Name)
        case block.IsToolResult():
            fmt.Printf("Tool result: %v\n", block.Content)
        }
    }
```

### ResultMessage

Final message with usage and cost information:

```go
case *claude.ResultMessage:
    fmt.Printf("Session: %s\n", m.SessionID)
    fmt.Printf("Turns: %d\n", m.NumTurns)
    fmt.Printf("Duration: %dms\n", m.DurationMS)
    fmt.Printf("Cost: $%.4f\n", m.TotalCostUSD)
```

### SystemMessage

System events and metadata:

```go
case *claude.SystemMessage:
    fmt.Printf("Event: %s, Data: %v\n", m.Subtype, m.Data)
```

## Error Handling

```go
import "errors"

msgs, err := claude.Query(ctx, "Hello")
if err != nil {
    if errors.Is(err, claude.ErrCLINotFound) {
        log.Fatal("Claude CLI not installed")
    }
    if errors.Is(err, claude.ErrNotConnected) {
        log.Fatal("Not connected to Claude")
    }
    if procErr, ok := err.(*claude.ProcessError); ok {
        log.Printf("Process failed (exit %d): %s", procErr.ExitCode, procErr.Stderr)
    }
    log.Fatal(err)
}
```

## Best Practices

### Use Context for Cancellation

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

msgs, err := claude.Query(ctx, "Long running task...")
```

### Always Close the Client

```go
client := claude.NewClient()
if err := client.Connect(ctx); err != nil {
    log.Fatal(err)
}
defer client.Close() // Always close to clean up subprocess
```

### Handle All Message Types

```go
for msg := range client.Messages() {
    switch m := msg.(type) {
    case *claude.UserMessage:
        // Echo of user input
    case *claude.AssistantMessage:
        // Claude's response
    case *claude.SystemMessage:
        // System events
    case *claude.ResultMessage:
        // Final result
    default:
        log.Printf("Unknown message type: %T", msg)
    }
}
```

### Set Appropriate Limits

```go
client := claude.NewClient(
    claude.WithMaxTurns(20),        // Prevent runaway conversations
    claude.WithMaxBudgetUSD(5.0),   // Cost control
)
```

## Testing

The SDK includes a mock transport for testing:

```go
func TestMyAgent(t *testing.T) {
    // Create mock transport
    mt := claude.NewMockTransport()

    // Queue expected responses
    mt.QueueMessage([]byte(`{"type": "assistant", "message": {...}}`))
    mt.QueueMessage([]byte(`{"type": "result", "subtype": "success"}`))
    mt.CloseMessages()

    // Use mock in client
    client := claude.NewClient(claude.WithTransport(mt))
    // ... test your code
}
```

## Contributing

Contributions are welcome! Please read our [Contributing Guidelines](CONTRIBUTING.md) before submitting a PR.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Write tests for your changes
4. Ensure all tests pass (`go test ./...`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Related Projects

- [Claude Code](https://docs.anthropic.com/en/docs/claude-code) - The CLI this SDK wraps
- [claude-agent-sdk-python](https://github.com/anthropics/claude-agent-sdk-python) - Official Python SDK
