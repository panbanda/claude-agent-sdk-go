// Example: hooks-security
// Demonstrates how pre-tool hooks can filter and block commands.
//
// This example blocks commands containing "/etc" to demonstrate
// the filtering mechanism. In production, you would configure
// your own security policies.
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/panbanda/claude-agent-sdk-go/claude"
)

// blockedPatterns demonstrates command filtering (for demo purposes).
var blockedPatterns = []string{
	"/etc",     // Block access to system config (for demo)
	"/private", // Block access to private dirs (for demo)
}

func main() {
	ctx := context.Background()

	// This prompt will trigger the hook - Claude will try to read /etc/hosts
	// which our hook will block, demonstrating the security filtering
	prompt := `Read the file /etc/hosts and show its contents.
If that's blocked, just say "Access was blocked by security hook".`

	fmt.Println("Prompt:", prompt)
	fmt.Println()

	msgs, err := claude.Query(ctx,
		prompt,
		claude.WithPreToolUseHook("", func(
			ctx context.Context,
			input *claude.PreToolUseInput,
			hookCtx *claude.HookContext,
		) (*claude.HookOutput, error) {
			// Check Bash commands
			if input.ToolName == "Bash" {
				cmd, _ := input.ToolInput["command"].(string)
				for _, pattern := range blockedPatterns {
					if strings.Contains(cmd, pattern) {
						fmt.Printf("[BLOCKED] Command accessing %s: %s\n", pattern, cmd)
						return &claude.HookOutput{
							Decision: claude.HookDecisionDeny,
							Reason:   fmt.Sprintf("Access to %s is not allowed", pattern),
						}, nil
					}
				}
			}

			// Check Read tool file paths
			if input.ToolName == "Read" {
				path, _ := input.ToolInput["file_path"].(string)
				for _, pattern := range blockedPatterns {
					if strings.Contains(path, pattern) {
						fmt.Printf("[BLOCKED] Read access to %s: %s\n", pattern, path)
						return &claude.HookOutput{
							Decision: claude.HookDecisionDeny,
							Reason:   fmt.Sprintf("Reading files in %s is not allowed", pattern),
						}, nil
					}
				}
			}

			fmt.Printf("[ALLOWED] %s\n", input.ToolName)
			return &claude.HookOutput{Decision: claude.HookDecisionAllow}, nil
		}),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n--- Response ---")
	for msg := range msgs {
		if m, ok := msg.(*claude.AssistantMessage); ok {
			for _, block := range m.Content {
				if block.IsText() {
					fmt.Print(block.Text)
				}
			}
		}
	}
	fmt.Println()
}
