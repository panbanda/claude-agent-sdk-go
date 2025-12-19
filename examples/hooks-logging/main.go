// Example: hooks-logging
// Log all tool usage for auditing and debugging.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/panbanda/claude-agent-sdk-go/claude"
)

type ToolLog struct {
	Timestamp string         `json:"timestamp"`
	Tool      string         `json:"tool"`
	Input     map[string]any `json:"input"`
	Output    any            `json:"output,omitempty"`
	IsError   bool           `json:"is_error,omitempty"`
	Duration  string         `json:"duration,omitempty"`
}

var toolStartTimes = make(map[string]time.Time)

func main() {
	ctx := context.Background()

	msgs, err := claude.Query(ctx,
		"Read the go.mod file and tell me the module name",
		claude.WithPreToolUseHook("", func(
			ctx context.Context,
			input *claude.PreToolUseInput,
			hookCtx *claude.HookContext,
		) (*claude.HookOutput, error) {
			toolStartTimes[input.ToolUseID] = time.Now()
			logEntry := ToolLog{
				Timestamp: time.Now().Format(time.RFC3339),
				Tool:      input.ToolName,
				Input:     input.ToolInput,
			}
			printLog("PRE", logEntry)
			return &claude.HookOutput{Decision: claude.HookDecisionAllow}, nil
		}),
		claude.WithPostToolUseHook("", func(
			ctx context.Context,
			input *claude.PostToolUseInput,
			hookCtx *claude.HookContext,
		) (*claude.HookOutput, error) {
			duration := ""
			if start, ok := toolStartTimes[input.ToolUseID]; ok {
				duration = time.Since(start).String()
				delete(toolStartTimes, input.ToolUseID)
			}
			logEntry := ToolLog{
				Timestamp: time.Now().Format(time.RFC3339),
				Tool:      input.ToolName,
				Input:     input.ToolInput,
				Output:    truncateOutput(input.ToolResponse),
				IsError:   input.IsError,
				Duration:  duration,
			}
			printLog("POST", logEntry)
			return nil, nil
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

func printLog(phase string, entry ToolLog) {
	data, _ := json.MarshalIndent(entry, "", "  ")
	fmt.Printf("[%s] %s\n", phase, string(data))
}

func truncateOutput(v any) any {
	s, ok := v.(string)
	if !ok {
		return v
	}
	if len(s) > 100 {
		return s[:100] + "..."
	}
	return s
}
