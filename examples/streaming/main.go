// Example: streaming
// Display messages in real-time as they arrive, showing all message types.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/panbanda/claude-agent-sdk-go/claude"
)

func main() {
	ctx := context.Background()

	msgs, err := claude.Query(ctx, "Write a haiku about programming.")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	for msg := range msgs {
		switch m := msg.(type) {
		case *claude.UserMessage:
			fmt.Printf("[User] %s\n", m.Content)

		case *claude.AssistantMessage:
			for _, block := range m.Content {
				switch {
				case block.IsText():
					fmt.Printf("[Text] %s\n", block.Text)
				case block.IsToolUse():
					fmt.Printf("[Tool] %s(%v)\n", block.ToolName, block.ToolInput)
				case block.IsToolResult():
					fmt.Printf("[Result] %v\n", block.ToolResult)
				case block.IsThinking():
					fmt.Printf("[Thinking] %s...\n", truncate(block.Thinking, 50))
				}
			}

		case *claude.SystemMessage:
			fmt.Printf("[System] %s: %v\n", m.Subtype, m.Data)

		case *claude.ResultMessage:
			fmt.Println()
			fmt.Printf("Completed in %dms, cost $%.4f\n", m.DurationMS, m.TotalCostUSD)
		}
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
