package claude

import (
	"context"
	"testing"
)

func TestHookEvent(t *testing.T) {
	t.Run("hook event constants exist", func(t *testing.T) {
		events := []HookEvent{
			PreToolUse,
			PostToolUse,
			UserPromptSubmit,
			Stop,
			SubagentStop,
			PreCompact,
		}

		// All events should be non-empty strings
		for _, e := range events {
			if e == "" {
				t.Error("HookEvent should not be empty")
			}
		}

		// All events should be distinct
		seen := make(map[HookEvent]bool)
		for _, e := range events {
			if seen[e] {
				t.Errorf("duplicate event: %s", e)
			}
			seen[e] = true
		}
	})
}

func TestPreToolUseHook(t *testing.T) {
	t.Run("hook input contains tool info", func(t *testing.T) {
		input := &PreToolUseInput{
			ToolName:  "Bash",
			ToolInput: map[string]any{"command": "ls"},
		}

		if input.ToolName != "Bash" {
			t.Errorf("ToolName = %q, want 'Bash'", input.ToolName)
		}
		if input.ToolInput["command"] != "ls" {
			t.Errorf("ToolInput[command] = %v, want 'ls'", input.ToolInput["command"])
		}
	})

	t.Run("hook output can deny permission", func(t *testing.T) {
		output := &HookOutput{
			Decision: HookDecisionDeny,
			Reason:   "Command blocked by policy",
		}

		if output.Decision != HookDecisionDeny {
			t.Errorf("Decision = %v, want %v", output.Decision, HookDecisionDeny)
		}
	})

	t.Run("hook output can allow permission", func(t *testing.T) {
		output := &HookOutput{
			Decision: HookDecisionAllow,
		}

		if output.Decision != HookDecisionAllow {
			t.Errorf("Decision = %v, want %v", output.Decision, HookDecisionAllow)
		}
	})

	t.Run("hook output can modify input", func(t *testing.T) {
		output := &HookOutput{
			Decision: HookDecisionAllow,
			UpdatedInput: map[string]any{
				"command": "ls -la",
			},
		}

		if output.UpdatedInput["command"] != "ls -la" {
			t.Errorf("UpdatedInput[command] = %v, want 'ls -la'", output.UpdatedInput["command"])
		}
	})
}

func TestPostToolUseHook(t *testing.T) {
	t.Run("hook input contains tool response", func(t *testing.T) {
		input := &PostToolUseInput{
			ToolName:     "Bash",
			ToolInput:    map[string]any{"command": "ls"},
			ToolResponse: "file1.txt\nfile2.txt",
			IsError:      false,
		}

		if input.ToolResponse != "file1.txt\nfile2.txt" {
			t.Errorf("ToolResponse = %q, want 'file1.txt\\nfile2.txt'", input.ToolResponse)
		}
		if input.IsError {
			t.Error("IsError should be false")
		}
	})
}

func TestHookContext(t *testing.T) {
	t.Run("context contains session info", func(t *testing.T) {
		hookCtx := &HookContext{
			SessionID:      "session-123",
			TranscriptPath: "/tmp/transcript.json",
			WorkingDir:     "/home/user/project",
			PermissionMode: string(PermissionAcceptEdits),
		}

		if hookCtx.SessionID != "session-123" {
			t.Errorf("SessionID = %q, want 'session-123'", hookCtx.SessionID)
		}
	})
}

func TestWithPreToolUseHook(t *testing.T) {
	t.Run("registers hook function", func(t *testing.T) {
		hook := func(ctx context.Context, input *PreToolUseInput, hookCtx *HookContext) (*HookOutput, error) {
			return &HookOutput{Decision: HookDecisionAllow}, nil
		}

		cfg := &config{}
		WithPreToolUseHook("Bash", hook)(cfg)

		if len(cfg.hooks) == 0 {
			t.Fatal("hooks map should not be empty")
		}
		if _, ok := cfg.hooks[PreToolUse]; !ok {
			t.Error("PreToolUse hook not registered")
		}
	})

	t.Run("registers multiple hooks for same event", func(t *testing.T) {
		hook1 := func(ctx context.Context, input *PreToolUseInput, hookCtx *HookContext) (*HookOutput, error) {
			return &HookOutput{Decision: HookDecisionAllow}, nil
		}
		hook2 := func(ctx context.Context, input *PreToolUseInput, hookCtx *HookContext) (*HookOutput, error) {
			return &HookOutput{Decision: HookDecisionAllow}, nil
		}

		cfg := &config{}
		WithPreToolUseHook("Bash", hook1)(cfg)
		WithPreToolUseHook("Read", hook2)(cfg)

		matchers := cfg.hooks[PreToolUse]
		if len(matchers) != 2 {
			t.Errorf("matchers length = %d, want 2", len(matchers))
		}
	})
}

func TestWithPostToolUseHook(t *testing.T) {
	t.Run("registers hook function", func(t *testing.T) {
		hook := func(ctx context.Context, input *PostToolUseInput, hookCtx *HookContext) (*HookOutput, error) {
			return &HookOutput{}, nil
		}

		cfg := &config{}
		WithPostToolUseHook("Bash", hook)(cfg)

		if len(cfg.hooks) == 0 {
			t.Fatal("hooks map should not be empty")
		}
		if _, ok := cfg.hooks[PostToolUse]; !ok {
			t.Error("PostToolUse hook not registered")
		}
	})
}

func TestWithUserPromptSubmitHook(t *testing.T) {
	t.Run("registers hook function", func(t *testing.T) {
		hook := func(ctx context.Context, input *UserPromptSubmitInput, hookCtx *HookContext) (*HookOutput, error) {
			return &HookOutput{
				AdditionalContext: "Extra context",
			}, nil
		}

		cfg := &config{}
		WithUserPromptSubmitHook(hook)(cfg)

		if len(cfg.hooks) == 0 {
			t.Fatal("hooks map should not be empty")
		}
		if _, ok := cfg.hooks[UserPromptSubmit]; !ok {
			t.Error("UserPromptSubmit hook not registered")
		}
	})
}

func TestWithStopHook(t *testing.T) {
	t.Run("registers hook function", func(t *testing.T) {
		hook := func(ctx context.Context, input *StopInput, hookCtx *HookContext) (*HookOutput, error) {
			return &HookOutput{}, nil
		}

		cfg := &config{}
		WithStopHook(hook)(cfg)

		if len(cfg.hooks) == 0 {
			t.Fatal("hooks map should not be empty")
		}
		if _, ok := cfg.hooks[Stop]; !ok {
			t.Error("Stop hook not registered")
		}
	})
}

func TestHookDecision(t *testing.T) {
	t.Run("decision constants exist", func(t *testing.T) {
		decisions := []HookDecision{
			HookDecisionAllow,
			HookDecisionDeny,
			HookDecisionNone,
		}

		// All decisions should be distinct
		seen := make(map[HookDecision]bool)
		for _, d := range decisions {
			if seen[d] {
				t.Errorf("duplicate decision: %s", d)
			}
			seen[d] = true
		}
	})
}
