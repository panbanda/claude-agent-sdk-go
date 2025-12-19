// Example: result-only
// When you just want the final answer without handling streaming messages.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/panbanda/claude-agent-sdk-go/claude"
)

func main() {
	ctx := context.Background()

	result, err := claude.QueryResult(ctx, "What is the capital of the Philippines? One word only.")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Answer: %s\n", result.Result)
	fmt.Printf("Cost: $%.4f\n", result.TotalCostUSD)
	fmt.Printf("Duration: %dms\n", result.DurationMS)
}
