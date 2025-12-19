package claude

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
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

	t.Run("registers with timeout option", func(t *testing.T) {
		hook := func(ctx context.Context, input *PostToolUseInput, hookCtx *HookContext) (*HookOutput, error) {
			return &HookOutput{}, nil
		}

		cfg := &config{}
		WithPostToolUseHook("Bash", hook, HookTimeout(45*time.Second))(cfg)

		matchers := cfg.hooks[PostToolUse]
		if len(matchers) != 1 {
			t.Fatalf("matchers length = %d, want 1", len(matchers))
		}
		if matchers[0].timeout != 45*time.Second {
			t.Errorf("timeout = %v, want 45s", matchers[0].timeout)
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

	t.Run("registers with timeout option", func(t *testing.T) {
		hook := func(ctx context.Context, input *UserPromptSubmitInput, hookCtx *HookContext) (*HookOutput, error) {
			return &HookOutput{}, nil
		}

		cfg := &config{}
		WithUserPromptSubmitHook(hook, HookTimeout(20*time.Second))(cfg)

		matchers := cfg.hooks[UserPromptSubmit]
		if len(matchers) != 1 {
			t.Fatalf("matchers length = %d, want 1", len(matchers))
		}
		if matchers[0].timeout != 20*time.Second {
			t.Errorf("timeout = %v, want 20s", matchers[0].timeout)
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

	t.Run("registers with timeout option", func(t *testing.T) {
		hook := func(ctx context.Context, input *StopInput, hookCtx *HookContext) (*HookOutput, error) {
			return &HookOutput{}, nil
		}

		cfg := &config{}
		WithStopHook(hook, HookTimeout(15*time.Second))(cfg)

		matchers := cfg.hooks[Stop]
		if len(matchers) != 1 {
			t.Fatalf("matchers length = %d, want 1", len(matchers))
		}
		if matchers[0].timeout != 15*time.Second {
			t.Errorf("timeout = %v, want 15s", matchers[0].timeout)
		}
	})
}

func TestWithSubagentStopHook(t *testing.T) {
	t.Run("registers hook function", func(t *testing.T) {
		hook := func(ctx context.Context, input *SubagentStopInput, hookCtx *HookContext) (*HookOutput, error) {
			return &HookOutput{}, nil
		}

		cfg := &config{}
		WithSubagentStopHook(hook)(cfg)

		if len(cfg.hooks) == 0 {
			t.Fatal("hooks map should not be empty")
		}
		if _, ok := cfg.hooks[SubagentStop]; !ok {
			t.Error("SubagentStop hook not registered")
		}
	})

	t.Run("registers with timeout option", func(t *testing.T) {
		hook := func(ctx context.Context, input *SubagentStopInput, hookCtx *HookContext) (*HookOutput, error) {
			return &HookOutput{}, nil
		}

		cfg := &config{}
		WithSubagentStopHook(hook, HookTimeout(30*time.Second))(cfg)

		matchers := cfg.hooks[SubagentStop]
		if len(matchers) != 1 {
			t.Fatalf("matchers length = %d, want 1", len(matchers))
		}
		if matchers[0].timeout != 30*time.Second {
			t.Errorf("timeout = %v, want 30s", matchers[0].timeout)
		}
	})
}

func TestWithPreCompactHook(t *testing.T) {
	t.Run("registers hook function", func(t *testing.T) {
		hook := func(ctx context.Context, input *PreCompactInput, hookCtx *HookContext) (*HookOutput, error) {
			return &HookOutput{}, nil
		}

		cfg := &config{}
		WithPreCompactHook(hook)(cfg)

		if len(cfg.hooks) == 0 {
			t.Fatal("hooks map should not be empty")
		}
		if _, ok := cfg.hooks[PreCompact]; !ok {
			t.Error("PreCompact hook not registered")
		}
	})

	t.Run("registers with timeout option", func(t *testing.T) {
		hook := func(ctx context.Context, input *PreCompactInput, hookCtx *HookContext) (*HookOutput, error) {
			return &HookOutput{}, nil
		}

		cfg := &config{}
		WithPreCompactHook(hook, HookTimeout(60*time.Second))(cfg)

		matchers := cfg.hooks[PreCompact]
		if len(matchers) != 1 {
			t.Fatalf("matchers length = %d, want 1", len(matchers))
		}
		if matchers[0].timeout != 60*time.Second {
			t.Errorf("timeout = %v, want 60s", matchers[0].timeout)
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

func TestHookInitialization(t *testing.T) {
	t.Run("sends initialize request with hooks on connect", func(t *testing.T) {
		hook := func(ctx context.Context, input *PreToolUseInput, hookCtx *HookContext) (*HookOutput, error) {
			return &HookOutput{Decision: HookDecisionAllow}, nil
		}

		mt := newMockTransport()
		client := NewClient(
			WithTransport(mt),
			WithPreToolUseHook("Bash", hook),
		)

		ctx := context.Background()
		if err := client.Connect(ctx); err != nil {
			t.Fatalf("Connect() error = %v", err)
		}
		defer client.Close()

		// Check that an initialize request was sent
		if len(mt.sentMessages) == 0 {
			t.Fatal("no messages sent on connect")
		}

		initMsg := string(mt.sentMessages[0])
		if !strings.Contains(initMsg, "control_request") {
			t.Errorf("first message should be control_request, got: %s", initMsg)
		}
		if !strings.Contains(initMsg, "initialize") {
			t.Errorf("first message should contain initialize subtype, got: %s", initMsg)
		}
		if !strings.Contains(initMsg, "PreToolUse") {
			t.Errorf("initialize should contain PreToolUse hook, got: %s", initMsg)
		}
		if !strings.Contains(initMsg, "hook_0") {
			t.Errorf("initialize should contain callback ID hook_0, got: %s", initMsg)
		}
	})
}

func TestHookCallbackExecution(t *testing.T) {
	t.Run("PreToolUse hook is invoked on control_request", func(t *testing.T) {
		hookCalled := false
		var receivedInput *PreToolUseInput

		hook := func(ctx context.Context, input *PreToolUseInput, hookCtx *HookContext) (*HookOutput, error) {
			hookCalled = true
			receivedInput = input
			return &HookOutput{Decision: HookDecisionAllow}, nil
		}

		mt := newMockTransport()
		client := NewClient(
			WithTransport(mt),
			WithPreToolUseHook("Bash", hook),
		)

		ctx := context.Background()
		if err := client.Connect(ctx); err != nil {
			t.Fatalf("Connect() error = %v", err)
		}
		defer client.Close()

		// Simulate CLI sending a hook_callback control request
		controlRequest := `{"type":"control_request","request_id":"req-123","request":{"subtype":"hook_callback","callback_id":"hook_0","input":{"hook_event_name":"PreToolUse","tool_name":"Bash","tool_input":{"command":"ls -la"},"tool_use_id":"tool-456"}}}`
		mt.QueueMessage([]byte(controlRequest))

		// Also queue a result so the message loop terminates
		mt.QueueMessage([]byte(`{"type":"result","subtype":"success"}`))
		mt.CloseMessages()

		// Consume messages to trigger hook processing
		for range client.Messages() {
		}

		if !hookCalled {
			t.Error("hook was not called")
		}
		if receivedInput == nil {
			t.Fatal("receivedInput is nil")
		}
		if receivedInput.ToolName != "Bash" {
			t.Errorf("ToolName = %q, want 'Bash'", receivedInput.ToolName)
		}
		if receivedInput.ToolInput["command"] != "ls -la" {
			t.Errorf("ToolInput[command] = %v, want 'ls -la'", receivedInput.ToolInput["command"])
		}
	})

	t.Run("control_response is sent after hook execution", func(t *testing.T) {
		hook := func(ctx context.Context, input *PreToolUseInput, hookCtx *HookContext) (*HookOutput, error) {
			return &HookOutput{
				Decision: HookDecisionDeny,
				Reason:   "blocked by test",
			}, nil
		}

		mt := newMockTransport()
		client := NewClient(
			WithTransport(mt),
			WithPreToolUseHook("", hook),
		)

		ctx := context.Background()
		if err := client.Connect(ctx); err != nil {
			t.Fatalf("Connect() error = %v", err)
		}
		defer client.Close()

		// Simulate CLI sending a hook_callback control request
		controlRequest := `{"type":"control_request","request_id":"req-789","request":{"subtype":"hook_callback","callback_id":"hook_0","input":{"hook_event_name":"PreToolUse","tool_name":"Read","tool_input":{"file_path":"/etc/passwd"},"tool_use_id":"tool-999"}}}`
		mt.QueueMessage([]byte(controlRequest))
		mt.QueueMessage([]byte(`{"type":"result","subtype":"success"}`))
		mt.CloseMessages()

		for range client.Messages() {
		}

		// Check that a control_response was sent
		if len(mt.sentMessages) == 0 {
			t.Fatal("no messages sent back to CLI")
		}

		// Find the control_response in sent messages
		var foundResponse bool
		for _, msg := range mt.sentMessages {
			msgStr := string(msg)
			if strings.Contains(msgStr, "control_response") && strings.Contains(msgStr, "req-789") {
				foundResponse = true
				if !strings.Contains(msgStr, "deny") {
					t.Errorf("response should contain deny decision, got: %s", msgStr)
				}
				break
			}
		}
		if !foundResponse {
			t.Errorf("control_response not found in sent messages: %v", mt.sentMessages)
		}
	})

	t.Run("PostToolUse hook is invoked on control_request", func(t *testing.T) {
		hookCalled := false
		var receivedInput *PostToolUseInput

		hook := func(ctx context.Context, input *PostToolUseInput, hookCtx *HookContext) (*HookOutput, error) {
			hookCalled = true
			receivedInput = input
			return &HookOutput{}, nil
		}

		mt := newMockTransport()
		client := NewClient(
			WithTransport(mt),
			WithPostToolUseHook("", hook),
		)

		ctx := context.Background()
		if err := client.Connect(ctx); err != nil {
			t.Fatalf("Connect() error = %v", err)
		}
		defer client.Close()

		controlRequest := `{"type":"control_request","request_id":"req-post","request":{"subtype":"hook_callback","callback_id":"hook_0","input":{"hook_event_name":"PostToolUse","tool_name":"Bash","tool_input":{"command":"echo hello"},"tool_response":"hello\n","is_error":false,"tool_use_id":"tool-post"}}}`
		mt.QueueMessage([]byte(controlRequest))
		mt.QueueMessage([]byte(`{"type":"result","subtype":"success"}`))
		mt.CloseMessages()

		for range client.Messages() {
		}

		if !hookCalled {
			t.Error("PostToolUse hook was not called")
		}
		if receivedInput == nil {
			t.Fatal("receivedInput is nil")
		}
		if receivedInput.ToolName != "Bash" {
			t.Errorf("ToolName = %q, want 'Bash'", receivedInput.ToolName)
		}
		if receivedInput.ToolResponse != "hello\n" {
			t.Errorf("ToolResponse = %v, want 'hello\\n'", receivedInput.ToolResponse)
		}
	})

	t.Run("hook returning error results in continue response", func(t *testing.T) {
		hook := func(ctx context.Context, input *PreToolUseInput, hookCtx *HookContext) (*HookOutput, error) {
			return nil, errors.New("hook failed")
		}

		mt := newMockTransport()
		client := NewClient(
			WithTransport(mt),
			WithPreToolUseHook("", hook),
		)

		ctx := context.Background()
		if err := client.Connect(ctx); err != nil {
			t.Fatalf("Connect() error = %v", err)
		}
		defer client.Close()

		controlRequest := `{"type":"control_request","request_id":"req-err","request":{"subtype":"hook_callback","callback_id":"hook_0","input":{"hook_event_name":"PreToolUse","tool_name":"Bash","tool_input":{},"tool_use_id":"tool-err"}}}`
		mt.QueueMessage([]byte(controlRequest))
		mt.QueueMessage([]byte(`{"type":"result","subtype":"success"}`))
		mt.CloseMessages()

		for range client.Messages() {
		}

		// Response should have continue: true due to error
		var foundResponse bool
		for _, msg := range mt.sentMessages {
			msgStr := string(msg)
			if strings.Contains(msgStr, "control_response") && strings.Contains(msgStr, "req-err") {
				foundResponse = true
				if !strings.Contains(msgStr, `"continue":true`) {
					t.Errorf("response should contain continue:true, got: %s", msgStr)
				}
			}
		}
		if !foundResponse {
			t.Error("control_response not found")
		}
	})

	t.Run("hook with Continue field set", func(t *testing.T) {
		continueVal := false
		hook := func(ctx context.Context, input *PreToolUseInput, hookCtx *HookContext) (*HookOutput, error) {
			return &HookOutput{
				Decision: HookDecisionAllow,
				Continue: &continueVal,
			}, nil
		}

		mt := newMockTransport()
		client := NewClient(
			WithTransport(mt),
			WithPreToolUseHook("", hook),
		)

		ctx := context.Background()
		if err := client.Connect(ctx); err != nil {
			t.Fatalf("Connect() error = %v", err)
		}
		defer client.Close()

		controlRequest := `{"type":"control_request","request_id":"req-cont","request":{"subtype":"hook_callback","callback_id":"hook_0","input":{"hook_event_name":"PreToolUse","tool_name":"Bash","tool_input":{},"tool_use_id":"tool-cont"}}}`
		mt.QueueMessage([]byte(controlRequest))
		mt.QueueMessage([]byte(`{"type":"result","subtype":"success"}`))
		mt.CloseMessages()

		for range client.Messages() {
		}

		// Response should have continue: false
		var foundResponse bool
		for _, msg := range mt.sentMessages {
			msgStr := string(msg)
			if strings.Contains(msgStr, "control_response") && strings.Contains(msgStr, "req-cont") {
				foundResponse = true
				if !strings.Contains(msgStr, `"continue":false`) {
					t.Errorf("response should contain continue:false, got: %s", msgStr)
				}
			}
		}
		if !foundResponse {
			t.Error("control_response not found")
		}
	})
}
