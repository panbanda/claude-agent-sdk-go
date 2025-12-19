// Example: code-reviewer
// A practical code review agent that analyzes files for issues.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/panbanda/claude-agent-sdk-go/claude"
)

func main() {
	path := flag.String("path", ".", "Path to review")
	flag.Parse()

	ctx := context.Background()

	prompt := fmt.Sprintf(`Review the code in %s for:
1. Bugs or potential issues
2. Security vulnerabilities
3. Performance problems
4. Code style issues

Be concise. List only actual problems found, not suggestions for improvements.
If no issues found, say "No issues found."`, *path)

	msgs, err := claude.Query(ctx, prompt,
		claude.WithWorkingDir(*path),
		claude.WithAllowedTools("Read", "Glob", "Grep"),
		claude.WithSystemPrompt("You are a code reviewer. Be direct and specific. Only report actual issues."),
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
					fmt.Printf("[Reading: %s]\n", getFileName(block.ToolInput))
				}
			}
		case *claude.ResultMessage:
			fmt.Printf("\n\nReview completed in %dms (cost: $%.4f)\n",
				m.DurationMS, m.TotalCostUSD)
		}
	}
}

func getFileName(input map[string]any) string {
	if p, ok := input["file_path"].(string); ok {
		return p
	}
	if p, ok := input["pattern"].(string); ok {
		return p
	}
	return "..."
}
