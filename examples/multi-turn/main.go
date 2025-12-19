// Example: multi-turn
// Interactive conversation using the Client for multiple exchanges.
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/panbanda/claude-agent-sdk-go/claude"
)

func main() {
	ctx := context.Background()

	client := claude.NewClient(
		claude.WithMaxTurns(10),
		claude.WithSystemPrompt("You are a helpful assistant. Keep responses brief."),
	)

	if err := client.Connect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Chat with Claude (type 'quit' to exit)")
	fmt.Println()

	for {
		fmt.Print("You: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "quit" {
			break
		}

		if err := client.Query(ctx, input); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			continue
		}

		fmt.Print("Claude: ")
	readLoop:
		for msg := range client.Messages() {
			switch m := msg.(type) {
			case *claude.AssistantMessage:
				for _, block := range m.Content {
					if block.IsText() {
						fmt.Print(block.Text)
					}
				}
			case *claude.ResultMessage:
				// Query complete, break to prompt for next input
				break readLoop
			}
		}
		fmt.Println()
		fmt.Println()
	}
}
