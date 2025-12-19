// Example: extended-thinking
// Enable Claude's extended thinking for complex reasoning tasks.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/panbanda/claude-agent-sdk-go/claude"
)

func main() {
	ctx := context.Background()

	prompt := `A farmer has a fox, a chicken, and a bag of grain. He needs to cross a river
in a boat that can only carry him and one item at a time. If left alone:
- The fox will eat the chicken
- The chicken will eat the grain

How can the farmer get everything across safely? Think through this step by step.`

	msgs, err := claude.Query(ctx, prompt,
		claude.WithMaxThinkingTokens(10000),
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
				case block.IsThinking():
					fmt.Println("--- Thinking ---")
					fmt.Println(block.Thinking)
					fmt.Println("--- End Thinking ---")
					fmt.Println()
				case block.IsText():
					fmt.Println("--- Response ---")
					fmt.Print(block.Text)
				}
			}
		case *claude.ResultMessage:
			fmt.Printf("\n\nTokens used for thinking: check usage\n")
			fmt.Printf("Cost: $%.4f\n", m.TotalCostUSD)
		}
	}
}
