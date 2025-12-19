// Example: tool-control
// Restrict Claude to read-only operations by disallowing write tools.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/panbanda/claude-agent-sdk-go/claude"
)

func main() {
	ctx := context.Background()

	// Create a read-only agent that cannot modify files or run commands
	msgs, err := claude.Query(ctx,
		"List the files in the current directory and show me the contents of go.mod",
		claude.WithDisallowedTools("Write", "Edit", "Bash"),
		claude.WithAllowedTools("Read", "Glob", "Grep"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	for msg := range msgs {
		switch m := msg.(type) {
		case *claude.AssistantMessage:
			for _, block := range m.Content {
				switch {
				case block.IsText():
					fmt.Print(block.Text)
				case block.IsToolUse():
					fmt.Printf("\n[Using tool: %s]\n", block.ToolName)
				}
			}
		case *claude.ResultMessage:
			fmt.Printf("\n\nCost: $%.4f\n", m.TotalCostUSD)
		}
	}
}
