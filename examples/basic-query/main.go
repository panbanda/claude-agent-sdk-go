// Example: basic-query
// Simplest possible usage of the SDK - send a prompt and print responses.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/panbanda/claude-agent-sdk-go/claude"
)

func main() {
	ctx := context.Background()

	msgs, err := claude.Query(ctx, "What is 2 + 2? Answer briefly.")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	for msg := range msgs {
		switch m := msg.(type) {
		case *claude.AssistantMessage:
			for _, block := range m.Content {
				if block.IsText() {
					fmt.Print(block.Text)
				}
			}
		case *claude.ResultMessage:
			fmt.Printf("\n\nCost: $%.4f\n", m.TotalCostUSD)
		}
	}
}
