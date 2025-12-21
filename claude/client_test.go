package claude

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	t.Run("creates client with default options", func(t *testing.T) {
		client := NewClient()

		if client == nil {
			t.Fatal("NewClient() returned nil")
		}
	})

	t.Run("creates client with options", func(t *testing.T) {
		client := NewClient(
			WithModel("claude-sonnet-4-5"),
			WithMaxTurns(10),
		)

		if client == nil {
			t.Fatal("NewClient() returned nil")
		}
	})

	t.Run("creates client with custom transport", func(t *testing.T) {
		mt := newMockTransport()
		client := NewClient(WithTransport(mt))

		if client == nil {
			t.Fatal("NewClient() returned nil")
		}
	})
}

func TestClientConnect(t *testing.T) {
	t.Run("connects with transport", func(t *testing.T) {
		mt := newMockTransport()
		client := NewClient(WithTransport(mt))

		err := client.Connect(context.Background())

		if err != nil {
			t.Errorf("Connect() error = %v, want nil", err)
		}
	})

	t.Run("returns transport connect error", func(t *testing.T) {
		mt := newMockTransport()
		mt.connectErr = ErrCLINotFound
		client := NewClient(WithTransport(mt))

		err := client.Connect(context.Background())

		if !errors.Is(err, ErrCLINotFound) {
			t.Errorf("Connect() error = %v, want %v", err, ErrCLINotFound)
		}
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		mt := newMockTransport()
		client := NewClient(WithTransport(mt))

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := client.Connect(ctx)

		if !errors.Is(err, context.Canceled) {
			t.Errorf("Connect() error = %v, want %v", err, context.Canceled)
		}
	})
}

func TestClientClose(t *testing.T) {
	t.Run("closes connected client", func(t *testing.T) {
		mt := newMockTransport()
		client := NewClient(WithTransport(mt))
		_ = client.Connect(context.Background())

		err := client.Close()

		if err != nil {
			t.Errorf("Close() error = %v, want nil", err)
		}
	})

	t.Run("safe to call close on unconnected client", func(t *testing.T) {
		client := NewClient()

		err := client.Close()

		if err != nil {
			t.Errorf("Close() error = %v, want nil", err)
		}
	})

	t.Run("safe to call close multiple times", func(t *testing.T) {
		mt := newMockTransport()
		client := NewClient(WithTransport(mt))
		_ = client.Connect(context.Background())

		_ = client.Close()
		err := client.Close()

		if err != nil {
			t.Errorf("Close() second call error = %v, want nil", err)
		}
	})
}

func TestClientQuery(t *testing.T) {
	t.Run("sends prompt to transport", func(t *testing.T) {
		mt := newMockTransport()
		client := NewClient(WithTransport(mt))
		_ = client.Connect(context.Background())

		err := client.Query(context.Background(), "What is 2+2?")

		if err != nil {
			t.Errorf("Query() error = %v, want nil", err)
		}
		if len(mt.sentMessages) != 1 {
			t.Fatalf("sentMessages length = %d, want 1", len(mt.sentMessages))
		}

		// Verify message format
		var msg map[string]any
		if err := json.Unmarshal(mt.sentMessages[0], &msg); err != nil {
			t.Fatalf("failed to unmarshal sent message: %v", err)
		}
		if msg["type"] != "user" {
			t.Errorf("message type = %v, want 'user'", msg["type"])
		}
	})

	t.Run("fails when not connected", func(t *testing.T) {
		client := NewClient()

		err := client.Query(context.Background(), "test")

		if !errors.Is(err, ErrNotConnected) {
			t.Errorf("Query() error = %v, want %v", err, ErrNotConnected)
		}
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		mt := newMockTransport()
		client := NewClient(WithTransport(mt))
		_ = client.Connect(context.Background())

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := client.Query(ctx, "test")

		if !errors.Is(err, context.Canceled) {
			t.Errorf("Query() error = %v, want %v", err, context.Canceled)
		}
	})
}

func TestClientMessages(t *testing.T) {
	t.Run("returns message channel", func(t *testing.T) {
		mt := newMockTransport()
		client := NewClient(WithTransport(mt))
		_ = client.Connect(context.Background())

		// Queue a message
		assistantMsg := map[string]any{
			"type": "assistant",
			"message": map[string]any{
				"role":    "assistant",
				"content": []map[string]any{{"type": "text", "text": "4"}},
				"model":   "claude-sonnet-4-5",
			},
		}
		msgBytes, _ := json.Marshal(assistantMsg)
		mt.QueueMessage(msgBytes)
		mt.CloseMessages()

		msgs := client.Messages()
		if msgs == nil {
			t.Fatal("Messages() returned nil channel")
		}

		msg, ok := <-msgs
		if !ok {
			t.Fatal("channel closed without message")
		}
		if msg == nil {
			t.Fatal("received nil message")
		}
	})

	t.Run("returns nil when not connected", func(t *testing.T) {
		client := NewClient()

		msgs := client.Messages()

		if msgs != nil {
			t.Errorf("Messages() = %v, want nil when not connected", msgs)
		}
	})
}

func TestClientIsConnected(t *testing.T) {
	t.Run("returns false when not connected", func(t *testing.T) {
		client := NewClient()

		if client.IsConnected() {
			t.Error("IsConnected() = true, want false")
		}
	})

	t.Run("returns true when connected", func(t *testing.T) {
		mt := newMockTransport()
		client := NewClient(WithTransport(mt))
		_ = client.Connect(context.Background())

		if !client.IsConnected() {
			t.Error("IsConnected() = false after Connect(), want true")
		}
	})

	t.Run("returns false after close", func(t *testing.T) {
		mt := newMockTransport()
		client := NewClient(WithTransport(mt))
		_ = client.Connect(context.Background())
		_ = client.Close()

		if client.IsConnected() {
			t.Error("IsConnected() = true after Close(), want false")
		}
	})
}

func TestClientContextIntegration(t *testing.T) {
	t.Run("timeout cancels connect", func(t *testing.T) {
		mt := newMockTransport()
		// Make connect block
		mt.connectErr = nil
		client := NewClient(WithTransport(mt))

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		// Give the timeout a chance to fire
		time.Sleep(5 * time.Millisecond)

		err := client.Connect(ctx)

		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("Connect() error = %v, want %v", err, context.DeadlineExceeded)
		}
	})
}

func TestClientParseUserMessage(t *testing.T) {
	t.Run("parses user message from raw JSON", func(t *testing.T) {
		mt := newMockTransport()
		client := NewClient(WithTransport(mt))
		_ = client.Connect(context.Background())
		defer client.Close()

		userMsg := map[string]any{
			"type": "user",
			"message": map[string]any{
				"content": "Hello Claude",
			},
			"uuid":               "user-uuid-123",
			"parent_tool_use_id": "tool-789",
		}
		msgBytes, _ := json.Marshal(userMsg)
		mt.QueueMessage(msgBytes)
		mt.CloseMessages()

		msg := <-client.Messages()
		um, ok := msg.(*UserMessage)
		if !ok {
			t.Fatalf("expected *UserMessage, got %T", msg)
		}
		if um.Content != "Hello Claude" {
			t.Errorf("Content = %q, want 'Hello Claude'", um.Content)
		}
		if um.UUID != "user-uuid-123" {
			t.Errorf("UUID = %q, want 'user-uuid-123'", um.UUID)
		}
		if um.ParentToolUseID != "tool-789" {
			t.Errorf("ParentToolUseID = %q, want 'tool-789'", um.ParentToolUseID)
		}
	})

	t.Run("handles empty user message", func(t *testing.T) {
		mt := newMockTransport()
		client := NewClient(WithTransport(mt))
		_ = client.Connect(context.Background())
		defer client.Close()

		userMsg := map[string]any{
			"type":    "user",
			"message": map[string]any{},
		}
		msgBytes, _ := json.Marshal(userMsg)
		mt.QueueMessage(msgBytes)
		mt.CloseMessages()

		msg := <-client.Messages()
		um, ok := msg.(*UserMessage)
		if !ok {
			t.Fatalf("expected *UserMessage, got %T", msg)
		}
		if um.Content != "" {
			t.Errorf("Content = %q, want empty", um.Content)
		}
	})
}

func TestClientParseSystemMessage(t *testing.T) {
	t.Run("parses system message from raw JSON", func(t *testing.T) {
		mt := newMockTransport()
		client := NewClient(WithTransport(mt))
		_ = client.Connect(context.Background())
		defer client.Close()

		sysMsg := map[string]any{
			"type":    "system",
			"subtype": "init",
			"data": map[string]any{
				"version": "1.0",
				"session": "sess-456",
			},
		}
		msgBytes, _ := json.Marshal(sysMsg)
		mt.QueueMessage(msgBytes)
		mt.CloseMessages()

		msg := <-client.Messages()
		sm, ok := msg.(*SystemMessage)
		if !ok {
			t.Fatalf("expected *SystemMessage, got %T", msg)
		}
		if sm.Subtype != "init" {
			t.Errorf("Subtype = %q, want 'init'", sm.Subtype)
		}
		if sm.Data["version"] != "1.0" {
			t.Errorf("Data[version] = %v, want '1.0'", sm.Data["version"])
		}
	})

	t.Run("handles system message without data", func(t *testing.T) {
		mt := newMockTransport()
		client := NewClient(WithTransport(mt))
		_ = client.Connect(context.Background())
		defer client.Close()

		sysMsg := map[string]any{
			"type":    "system",
			"subtype": "status",
		}
		msgBytes, _ := json.Marshal(sysMsg)
		mt.QueueMessage(msgBytes)
		mt.CloseMessages()

		msg := <-client.Messages()
		sm, ok := msg.(*SystemMessage)
		if !ok {
			t.Fatalf("expected *SystemMessage, got %T", msg)
		}
		if sm.Data == nil {
			t.Error("Data should be initialized to empty map, not nil")
		}
	})
}

func TestClientParseAssistantMessage(t *testing.T) {
	t.Run("parses assistant message with parent_tool_use_id and error", func(t *testing.T) {
		mt := newMockTransport()
		client := NewClient(WithTransport(mt))
		_ = client.Connect(context.Background())
		defer client.Close()

		assistantMsg := map[string]any{
			"type":               "assistant",
			"parent_tool_use_id": "tool-parent-123",
			"error":              "something went wrong",
			"message": map[string]any{
				"model":   "claude-sonnet-4-5",
				"content": []any{},
			},
		}
		msgBytes, _ := json.Marshal(assistantMsg)
		mt.QueueMessage(msgBytes)
		mt.CloseMessages()

		msg := <-client.Messages()
		am, ok := msg.(*AssistantMessage)
		if !ok {
			t.Fatalf("expected *AssistantMessage, got %T", msg)
		}
		if am.ParentToolUseID != "tool-parent-123" {
			t.Errorf("ParentToolUseID = %q, want 'tool-parent-123'", am.ParentToolUseID)
		}
		if am.Error != "something went wrong" {
			t.Errorf("Error = %q, want 'something went wrong'", am.Error)
		}
	})
}

func TestClientParseContentBlocks(t *testing.T) {
	t.Run("parses tool_result content block", func(t *testing.T) {
		mt := newMockTransport()
		client := NewClient(WithTransport(mt))
		_ = client.Connect(context.Background())
		defer client.Close()

		assistantMsg := map[string]any{
			"type": "assistant",
			"message": map[string]any{
				"model": "claude-sonnet-4-5",
				"content": []any{
					map[string]any{
						"type":        "tool_result",
						"tool_use_id": "tool-result-123",
						"content":     "command output",
						"is_error":    true,
					},
				},
			},
		}
		msgBytes, _ := json.Marshal(assistantMsg)
		mt.QueueMessage(msgBytes)
		mt.CloseMessages()

		msg := <-client.Messages()
		am, ok := msg.(*AssistantMessage)
		if !ok {
			t.Fatalf("expected *AssistantMessage, got %T", msg)
		}
		if len(am.Content) != 1 {
			t.Fatalf("Content length = %d, want 1", len(am.Content))
		}
		block := am.Content[0]
		if block.Kind != BlockToolResult {
			t.Errorf("Kind = %q, want 'tool_result'", block.Kind)
		}
		if block.ToolUseID != "tool-result-123" {
			t.Errorf("ToolUseID = %q, want 'tool-result-123'", block.ToolUseID)
		}
		if !block.IsError {
			t.Error("IsError should be true")
		}
	})

	t.Run("parses tool_use content block", func(t *testing.T) {
		mt := newMockTransport()
		client := NewClient(WithTransport(mt))
		_ = client.Connect(context.Background())
		defer client.Close()

		assistantMsg := map[string]any{
			"type": "assistant",
			"message": map[string]any{
				"model": "claude-sonnet-4-5",
				"content": []any{
					map[string]any{
						"type":  "tool_use",
						"id":    "tool-use-456",
						"name":  "Bash",
						"input": map[string]any{"command": "ls"},
					},
				},
			},
		}
		msgBytes, _ := json.Marshal(assistantMsg)
		mt.QueueMessage(msgBytes)
		mt.CloseMessages()

		msg := <-client.Messages()
		am, ok := msg.(*AssistantMessage)
		if !ok {
			t.Fatalf("expected *AssistantMessage, got %T", msg)
		}
		if len(am.Content) != 1 {
			t.Fatalf("Content length = %d, want 1", len(am.Content))
		}
		block := am.Content[0]
		if block.Kind != BlockToolUse {
			t.Errorf("Kind = %q, want 'tool_use'", block.Kind)
		}
		if block.ToolUseID != "tool-use-456" {
			t.Errorf("ToolUseID = %q, want 'tool-use-456'", block.ToolUseID)
		}
		if block.ToolName != "Bash" {
			t.Errorf("ToolName = %q, want 'Bash'", block.ToolName)
		}
	})

	t.Run("parses thinking content block", func(t *testing.T) {
		mt := newMockTransport()
		client := NewClient(WithTransport(mt))
		_ = client.Connect(context.Background())
		defer client.Close()

		assistantMsg := map[string]any{
			"type": "assistant",
			"message": map[string]any{
				"model": "claude-sonnet-4-5",
				"content": []any{
					map[string]any{
						"type":      "thinking",
						"thinking":  "Let me think about this...",
						"signature": "sig-abc",
					},
				},
			},
		}
		msgBytes, _ := json.Marshal(assistantMsg)
		mt.QueueMessage(msgBytes)
		mt.CloseMessages()

		msg := <-client.Messages()
		am, ok := msg.(*AssistantMessage)
		if !ok {
			t.Fatalf("expected *AssistantMessage, got %T", msg)
		}
		if len(am.Content) != 1 {
			t.Fatalf("Content length = %d, want 1", len(am.Content))
		}
		block := am.Content[0]
		if block.Kind != BlockThinking {
			t.Errorf("Kind = %q, want 'thinking'", block.Kind)
		}
		if block.Thinking != "Let me think about this..." {
			t.Errorf("Thinking = %q, want 'Let me think about this...'", block.Thinking)
		}
		if block.Signature != "sig-abc" {
			t.Errorf("Signature = %q, want 'sig-abc'", block.Signature)
		}
	})

	t.Run("skips invalid content block items", func(t *testing.T) {
		mt := newMockTransport()
		client := NewClient(WithTransport(mt))
		_ = client.Connect(context.Background())
		defer client.Close()

		assistantMsg := map[string]any{
			"type": "assistant",
			"message": map[string]any{
				"model": "claude-sonnet-4-5",
				"content": []any{
					"invalid string item",
					map[string]any{"type": "text", "text": "valid"},
					nil,
					123,
				},
			},
		}
		msgBytes, _ := json.Marshal(assistantMsg)
		mt.QueueMessage(msgBytes)
		mt.CloseMessages()

		msg := <-client.Messages()
		am, ok := msg.(*AssistantMessage)
		if !ok {
			t.Fatalf("expected *AssistantMessage, got %T", msg)
		}
		// Only the valid text block should be parsed
		if len(am.Content) != 1 {
			t.Fatalf("Content length = %d, want 1 (only valid text block)", len(am.Content))
		}
		if am.Content[0].Text != "valid" {
			t.Errorf("Text = %q, want 'valid'", am.Content[0].Text)
		}
	})

	t.Run("handles unknown content block type", func(t *testing.T) {
		mt := newMockTransport()
		client := NewClient(WithTransport(mt))
		_ = client.Connect(context.Background())
		defer client.Close()

		assistantMsg := map[string]any{
			"type": "assistant",
			"message": map[string]any{
				"model": "claude-sonnet-4-5",
				"content": []any{
					map[string]any{"type": "unknown_type", "data": "some data"},
					map[string]any{"type": "text", "text": "valid text"},
				},
			},
		}
		msgBytes, _ := json.Marshal(assistantMsg)
		mt.QueueMessage(msgBytes)
		mt.CloseMessages()

		msg := <-client.Messages()
		am, ok := msg.(*AssistantMessage)
		if !ok {
			t.Fatalf("expected *AssistantMessage, got %T", msg)
		}
		// Unknown types are skipped, only valid text block should be parsed
		if len(am.Content) != 1 {
			t.Fatalf("Content length = %d, want 1", len(am.Content))
		}
	})
}

func TestClientParseMessage(t *testing.T) {
	t.Run("returns nil for invalid JSON", func(t *testing.T) {
		mt := newMockTransport()
		client := NewClient(WithTransport(mt))
		_ = client.Connect(context.Background())
		defer client.Close()

		mt.QueueMessage([]byte("not valid json"))
		mt.QueueMessage([]byte(`{"type":"result","subtype":"success"}`))
		mt.CloseMessages()

		// First message (invalid) should be skipped, second should come through
		msg := <-client.Messages()
		if msg == nil {
			t.Fatal("expected non-nil message")
		}
		_, ok := msg.(*ResultMessage)
		if !ok {
			t.Fatalf("expected *ResultMessage, got %T", msg)
		}
	})

	t.Run("returns nil for unknown message type", func(t *testing.T) {
		mt := newMockTransport()
		client := NewClient(WithTransport(mt))
		_ = client.Connect(context.Background())
		defer client.Close()

		mt.QueueMessage([]byte(`{"type":"unknown_type"}`))
		mt.QueueMessage([]byte(`{"type":"result","subtype":"success"}`))
		mt.CloseMessages()

		// First message (unknown type) should be skipped
		msg := <-client.Messages()
		_, ok := msg.(*ResultMessage)
		if !ok {
			t.Fatalf("expected *ResultMessage, got %T", msg)
		}
	})
}

func TestClientCloseWithTransportError(t *testing.T) {
	t.Run("close returns transport error", func(t *testing.T) {
		mt := newMockTransport()
		mt.closeErr = errors.New("close failed")
		client := NewClient(WithTransport(mt))
		_ = client.Connect(context.Background())

		err := client.Close()

		if err == nil || err.Error() != "close failed" {
			t.Errorf("Close() error = %v, want 'close failed'", err)
		}
	})
}

func TestClientConnectSendInitializeError(t *testing.T) {
	t.Run("connect fails when sendInitialize fails", func(t *testing.T) {
		hook := func(ctx context.Context, input *PreToolUseInput, hookCtx *HookContext) (*HookOutput, error) {
			return &HookOutput{Decision: HookDecisionAllow}, nil
		}

		mt := newMockTransport()
		mt.sendErr = errors.New("send failed")
		client := NewClient(
			WithTransport(mt),
			WithPreToolUseHook("Bash", hook),
		)

		err := client.Connect(context.Background())

		if err == nil || err.Error() != "send failed" {
			t.Errorf("Connect() error = %v, want 'send failed'", err)
		}
	})
}

func TestClientSendInitializeWithTimeout(t *testing.T) {
	t.Run("sendInitialize includes timeout in hook definition", func(t *testing.T) {
		hook := func(ctx context.Context, input *PreToolUseInput, hookCtx *HookContext) (*HookOutput, error) {
			return &HookOutput{Decision: HookDecisionAllow}, nil
		}

		mt := newMockTransport()
		client := NewClient(
			WithTransport(mt),
			WithPreToolUseHook("Bash", hook, HookTimeout(30*time.Second)),
		)

		err := client.Connect(context.Background())
		if err != nil {
			t.Fatalf("Connect() error = %v, want nil", err)
		}
		defer client.Close()

		// Check that the initialize request was sent with timeout
		if len(mt.sentMessages) == 0 {
			t.Fatal("no messages sent on connect")
		}

		var initReq map[string]any
		if err := json.Unmarshal(mt.sentMessages[0], &initReq); err != nil {
			t.Fatalf("failed to unmarshal init request: %v", err)
		}

		// Navigate to the hook definition to check timeout
		request, _ := initReq["request"].(map[string]any)
		hooks, _ := request["hooks"].(map[string]any)
		preToolUse, _ := hooks["PreToolUse"].([]any)
		if len(preToolUse) == 0 {
			t.Fatal("PreToolUse hooks not found in initialize request")
		}

		hookDef := preToolUse[0].(map[string]any)
		timeout, ok := hookDef["timeout"]
		if !ok {
			t.Error("timeout not found in hook definition")
		}
		if timeout.(float64) != 30 {
			t.Errorf("timeout = %v, want 30", timeout)
		}
	})
}

func TestClientParseResultMessage(t *testing.T) {
	t.Run("parses result message with all fields", func(t *testing.T) {
		mt := newMockTransport()
		client := NewClient(WithTransport(mt))
		_ = client.Connect(context.Background())
		defer client.Close()

		resultMsg := map[string]any{
			"type":            "result",
			"subtype":         "success",
			"duration_ms":     1500.0,
			"duration_api_ms": 1200.0,
			"is_error":        false,
			"num_turns":       3.0,
			"session_id":      "sess-abc123",
			"total_cost_usd":  0.05,
			"usage": map[string]any{
				"input_tokens":  100.0,
				"output_tokens": 200.0,
			},
			"result":            "Task completed successfully",
			"structured_output": map[string]any{"key": "value"},
		}
		msgBytes, _ := json.Marshal(resultMsg)
		mt.QueueMessage(msgBytes)
		mt.CloseMessages()

		msg := <-client.Messages()
		rm, ok := msg.(*ResultMessage)
		if !ok {
			t.Fatalf("expected *ResultMessage, got %T", msg)
		}
		if rm.Subtype != "success" {
			t.Errorf("Subtype = %q, want 'success'", rm.Subtype)
		}
		if rm.DurationMS != 1500 {
			t.Errorf("DurationMS = %d, want 1500", rm.DurationMS)
		}
		if rm.DurationAPIMS != 1200 {
			t.Errorf("DurationAPIMS = %d, want 1200", rm.DurationAPIMS)
		}
		if rm.NumTurns != 3 {
			t.Errorf("NumTurns = %d, want 3", rm.NumTurns)
		}
		if rm.SessionID != "sess-abc123" {
			t.Errorf("SessionID = %q, want 'sess-abc123'", rm.SessionID)
		}
		if rm.TotalCostUSD != 0.05 {
			t.Errorf("TotalCostUSD = %f, want 0.05", rm.TotalCostUSD)
		}
		if rm.Usage == nil {
			t.Error("Usage should not be nil")
		}
		if rm.Result != "Task completed successfully" {
			t.Errorf("Result = %q, want 'Task completed successfully'", rm.Result)
		}
		if rm.StructuredOutput == nil {
			t.Error("StructuredOutput should not be nil")
		}
	})
}

func TestClientHandleControlRequestMalformed(t *testing.T) {
	t.Run("handles control_request with missing request field", func(t *testing.T) {
		mt := newMockTransport()
		client := NewClient(WithTransport(mt))
		_ = client.Connect(context.Background())
		defer client.Close()

		// Control request without proper request field
		controlReq := `{"type":"control_request","request_id":"req-123"}`
		mt.QueueMessage([]byte(controlReq))
		mt.QueueMessage([]byte(`{"type":"result","subtype":"success"}`))
		mt.CloseMessages()

		// Should not panic, just ignore the malformed request
		for range client.Messages() {
		}
	})

	t.Run("handles control_request with non-hook_callback subtype", func(t *testing.T) {
		mt := newMockTransport()
		client := NewClient(WithTransport(mt))
		_ = client.Connect(context.Background())
		defer client.Close()

		// Control request with unknown subtype
		controlReq := `{"type":"control_request","request_id":"req-123","request":{"subtype":"unknown"}}`
		mt.QueueMessage([]byte(controlReq))
		mt.QueueMessage([]byte(`{"type":"result","subtype":"success"}`))
		mt.CloseMessages()

		for range client.Messages() {
		}
	})

	t.Run("handles control_request with unknown callback_id", func(t *testing.T) {
		hook := func(ctx context.Context, input *PreToolUseInput, hookCtx *HookContext) (*HookOutput, error) {
			return &HookOutput{Decision: HookDecisionAllow}, nil
		}

		mt := newMockTransport()
		client := NewClient(
			WithTransport(mt),
			WithPreToolUseHook("Bash", hook),
		)
		_ = client.Connect(context.Background())
		defer client.Close()

		// Control request with callback_id that doesn't exist
		controlReq := `{"type":"control_request","request_id":"req-unknown","request":{"subtype":"hook_callback","callback_id":"unknown_callback","input":{"hook_event_name":"PreToolUse","tool_name":"Bash","tool_input":{}}}}`
		mt.QueueMessage([]byte(controlReq))
		mt.QueueMessage([]byte(`{"type":"result","subtype":"success"}`))
		mt.CloseMessages()

		// Should not panic
		for range client.Messages() {
		}
	})
}

func TestClient_Interrupt(t *testing.T) {
	t.Run("sends interrupt control request", func(t *testing.T) {
		mt := newMockTransport()
		client := NewClient(WithTransport(mt))
		_ = client.Connect(context.Background())

		err := client.Interrupt(context.Background())

		if err != nil {
			t.Errorf("Interrupt() error = %v, want nil", err)
		}
		if len(mt.sentMessages) != 1 {
			t.Fatalf("sentMessages length = %d, want 1", len(mt.sentMessages))
		}

		var msg map[string]any
		if err := json.Unmarshal(mt.sentMessages[0], &msg); err != nil {
			t.Fatalf("failed to unmarshal sent message: %v", err)
		}
		if msg["type"] != MessageTypeControlRequest {
			t.Errorf("message type = %v, want %v", msg["type"], MessageTypeControlRequest)
		}
		request, _ := msg["request"].(map[string]any)
		if request["subtype"] != string(ControlSubtypeInterrupt) {
			t.Errorf("request subtype = %v, want %v", request["subtype"], ControlSubtypeInterrupt)
		}
	})
}

func TestClient_Interrupt_NotConnected(t *testing.T) {
	t.Run("fails when not connected", func(t *testing.T) {
		client := NewClient()

		err := client.Interrupt(context.Background())

		if !errors.Is(err, ErrNotConnected) {
			t.Errorf("Interrupt() error = %v, want %v", err, ErrNotConnected)
		}
	})
}

func TestClient_SetPermissionMode(t *testing.T) {
	t.Run("sends set permission mode control request", func(t *testing.T) {
		mt := newMockTransport()
		client := NewClient(WithTransport(mt))
		_ = client.Connect(context.Background())

		err := client.SetPermissionMode(context.Background(), PermissionAcceptEdits)

		if err != nil {
			t.Errorf("SetPermissionMode() error = %v, want nil", err)
		}
		if len(mt.sentMessages) != 1 {
			t.Fatalf("sentMessages length = %d, want 1", len(mt.sentMessages))
		}

		var msg map[string]any
		if err := json.Unmarshal(mt.sentMessages[0], &msg); err != nil {
			t.Fatalf("failed to unmarshal sent message: %v", err)
		}
		if msg["type"] != MessageTypeControlRequest {
			t.Errorf("message type = %v, want %v", msg["type"], MessageTypeControlRequest)
		}
		request, _ := msg["request"].(map[string]any)
		if request["subtype"] != string(ControlSubtypeSetPermissionMode) {
			t.Errorf("request subtype = %v, want %v", request["subtype"], ControlSubtypeSetPermissionMode)
		}
		if request["mode"] != string(PermissionAcceptEdits) {
			t.Errorf("request mode = %v, want %v", request["mode"], PermissionAcceptEdits)
		}
	})
}

func TestClient_SetPermissionMode_NotConnected(t *testing.T) {
	t.Run("fails when not connected", func(t *testing.T) {
		client := NewClient()

		err := client.SetPermissionMode(context.Background(), PermissionDefault)

		if !errors.Is(err, ErrNotConnected) {
			t.Errorf("SetPermissionMode() error = %v, want %v", err, ErrNotConnected)
		}
	})
}

func TestClient_SetModel(t *testing.T) {
	t.Run("sends set model control request", func(t *testing.T) {
		mt := newMockTransport()
		client := NewClient(WithTransport(mt))
		_ = client.Connect(context.Background())

		err := client.SetModel(context.Background(), "claude-sonnet-4-5")

		if err != nil {
			t.Errorf("SetModel() error = %v, want nil", err)
		}
		if len(mt.sentMessages) != 1 {
			t.Fatalf("sentMessages length = %d, want 1", len(mt.sentMessages))
		}

		var msg map[string]any
		if err := json.Unmarshal(mt.sentMessages[0], &msg); err != nil {
			t.Fatalf("failed to unmarshal sent message: %v", err)
		}
		if msg["type"] != MessageTypeControlRequest {
			t.Errorf("message type = %v, want %v", msg["type"], MessageTypeControlRequest)
		}
		request, _ := msg["request"].(map[string]any)
		if request["subtype"] != string(ControlSubtypeSetModel) {
			t.Errorf("request subtype = %v, want %v", request["subtype"], ControlSubtypeSetModel)
		}
		if request["model"] != "claude-sonnet-4-5" {
			t.Errorf("request model = %v, want 'claude-sonnet-4-5'", request["model"])
		}
	})
}

func TestClient_SetModel_EmptyString(t *testing.T) {
	t.Run("sends nil model for empty string", func(t *testing.T) {
		mt := newMockTransport()
		client := NewClient(WithTransport(mt))
		_ = client.Connect(context.Background())

		err := client.SetModel(context.Background(), "")

		if err != nil {
			t.Errorf("SetModel() error = %v, want nil", err)
		}
		if len(mt.sentMessages) != 1 {
			t.Fatalf("sentMessages length = %d, want 1", len(mt.sentMessages))
		}

		var msg map[string]any
		if err := json.Unmarshal(mt.sentMessages[0], &msg); err != nil {
			t.Fatalf("failed to unmarshal sent message: %v", err)
		}
		request, _ := msg["request"].(map[string]any)
		if _, hasModel := request["model"]; hasModel {
			t.Errorf("request should not have model field for empty string, got %v", request["model"])
		}
	})
}

func TestClient_SetModel_NotConnected(t *testing.T) {
	t.Run("fails when not connected", func(t *testing.T) {
		client := NewClient()

		err := client.SetModel(context.Background(), "claude-sonnet-4-5")

		if !errors.Is(err, ErrNotConnected) {
			t.Errorf("SetModel() error = %v, want %v", err, ErrNotConnected)
		}
	})
}

func TestClient_GetServerInfo(t *testing.T) {
	t.Run("returns nil when no server info captured", func(t *testing.T) {
		client := NewClient()

		info := client.GetServerInfo()

		if info != nil {
			t.Errorf("GetServerInfo() = %v, want nil", info)
		}
	})
}

func TestClient_GetServerInfo_CapturedFromInit(t *testing.T) {
	t.Run("returns server info from init message", func(t *testing.T) {
		mt := newMockTransport()
		client := NewClient(WithTransport(mt))
		_ = client.Connect(context.Background())
		defer client.Close()

		initMsg := map[string]any{
			"type":    "system",
			"subtype": "init",
			"data": map[string]any{
				"slash_commands": []string{"/help", "/commit"},
				"output_styles":  []string{"default", "verbose"},
			},
		}
		msgBytes, _ := json.Marshal(initMsg)
		mt.QueueMessage(msgBytes)
		mt.CloseMessages()

		// Read the message to trigger parsing
		<-client.Messages()

		info := client.GetServerInfo()
		if info == nil {
			t.Fatal("GetServerInfo() = nil, want non-nil")
		}
		commands, _ := info["slash_commands"].([]any)
		if len(commands) != 2 {
			t.Errorf("slash_commands length = %d, want 2", len(commands))
		}
		styles, _ := info["output_styles"].([]any)
		if len(styles) != 2 {
			t.Errorf("output_styles length = %d, want 2", len(styles))
		}
	})
}
