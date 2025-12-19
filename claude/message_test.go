package claude

import (
	"encoding/json"
	"testing"
)

func TestMessageInterface(t *testing.T) {
	t.Run("all message types implement Message interface", func(t *testing.T) {
		// Compile-time check that all types implement Message
		var _ Message = &UserMessage{}
		var _ Message = &AssistantMessage{}
		var _ Message = &SystemMessage{}
		var _ Message = &ResultMessage{}
		var _ Message = &StreamEvent{}
	})
}

func TestUserMessage(t *testing.T) {
	t.Run("create with string content", func(t *testing.T) {
		msg := &UserMessage{
			Content: "Hello, Claude!",
		}

		if msg.Content != "Hello, Claude!" {
			t.Errorf("Content = %q, want %q", msg.Content, "Hello, Claude!")
		}
	})

	t.Run("with optional fields", func(t *testing.T) {
		msg := &UserMessage{
			Content:         "Hello",
			UUID:            "uuid-123",
			ParentToolUseID: "tool-456",
		}

		if msg.UUID != "uuid-123" {
			t.Errorf("UUID = %q, want %q", msg.UUID, "uuid-123")
		}
		if msg.ParentToolUseID != "tool-456" {
			t.Errorf("ParentToolUseID = %q, want %q", msg.ParentToolUseID, "tool-456")
		}
	})
}

func TestAssistantMessage(t *testing.T) {
	t.Run("create with text content", func(t *testing.T) {
		textBlock := NewTextBlock("Hello, human!")
		msg := &AssistantMessage{
			Content: []*ContentBlock{textBlock},
			Model:   "claude-opus-4-1-20250805",
		}

		if len(msg.Content) != 1 {
			t.Fatalf("Content length = %d, want 1", len(msg.Content))
		}
		if msg.Content[0].Text != "Hello, human!" {
			t.Errorf("Content[0].Text = %q, want %q", msg.Content[0].Text, "Hello, human!")
		}
		if msg.Model != "claude-opus-4-1-20250805" {
			t.Errorf("Model = %q, want %q", msg.Model, "claude-opus-4-1-20250805")
		}
	})

	t.Run("create with thinking content", func(t *testing.T) {
		thinkingBlock := NewThinkingBlock("I'm thinking...", "sig-123")
		msg := &AssistantMessage{
			Content: []*ContentBlock{thinkingBlock},
			Model:   "claude-opus-4-1-20250805",
		}

		if len(msg.Content) != 1 {
			t.Fatalf("Content length = %d, want 1", len(msg.Content))
		}
		if msg.Content[0].Thinking != "I'm thinking..." {
			t.Errorf("Thinking = %q, want %q", msg.Content[0].Thinking, "I'm thinking...")
		}
		if msg.Content[0].Signature != "sig-123" {
			t.Errorf("Signature = %q, want %q", msg.Content[0].Signature, "sig-123")
		}
	})

	t.Run("with error field", func(t *testing.T) {
		msg := &AssistantMessage{
			Content: []*ContentBlock{},
			Model:   "claude-opus-4-1-20250805",
			Error:   "rate_limit",
		}

		if msg.Error != "rate_limit" {
			t.Errorf("Error = %q, want %q", msg.Error, "rate_limit")
		}
	})

	t.Run("with parent tool use ID", func(t *testing.T) {
		msg := &AssistantMessage{
			Content:         []*ContentBlock{NewTextBlock("response")},
			Model:           "claude-opus-4-1-20250805",
			ParentToolUseID: "tool-789",
		}

		if msg.ParentToolUseID != "tool-789" {
			t.Errorf("ParentToolUseID = %q, want %q", msg.ParentToolUseID, "tool-789")
		}
	})
}

func TestSystemMessage(t *testing.T) {
	t.Run("create system message", func(t *testing.T) {
		msg := &SystemMessage{
			Subtype: "init",
			Data:    map[string]any{"version": "1.0"},
		}

		if msg.Subtype != "init" {
			t.Errorf("Subtype = %q, want %q", msg.Subtype, "init")
		}
		if msg.Data["version"] != "1.0" {
			t.Errorf("Data[version] = %v, want '1.0'", msg.Data["version"])
		}
	})
}

func TestResultMessage(t *testing.T) {
	t.Run("create result message", func(t *testing.T) {
		msg := &ResultMessage{
			Subtype:       "success",
			DurationMS:    1500,
			DurationAPIMS: 1200,
			IsError:       false,
			NumTurns:      1,
			SessionID:     "session-123",
			TotalCostUSD:  0.01,
		}

		if msg.Subtype != "success" {
			t.Errorf("Subtype = %q, want %q", msg.Subtype, "success")
		}
		if msg.TotalCostUSD != 0.01 {
			t.Errorf("TotalCostUSD = %f, want 0.01", msg.TotalCostUSD)
		}
		if msg.SessionID != "session-123" {
			t.Errorf("SessionID = %q, want %q", msg.SessionID, "session-123")
		}
		if msg.DurationMS != 1500 {
			t.Errorf("DurationMS = %d, want 1500", msg.DurationMS)
		}
		if msg.DurationAPIMS != 1200 {
			t.Errorf("DurationAPIMS = %d, want 1200", msg.DurationAPIMS)
		}
		if msg.NumTurns != 1 {
			t.Errorf("NumTurns = %d, want 1", msg.NumTurns)
		}
		if msg.IsError {
			t.Error("IsError should be false")
		}
	})

	t.Run("with optional fields", func(t *testing.T) {
		msg := &ResultMessage{
			Subtype:          "success",
			DurationMS:       1000,
			DurationAPIMS:    800,
			IsError:          false,
			NumTurns:         2,
			SessionID:        "sess-1",
			TotalCostUSD:     0.05,
			Usage:            map[string]any{"input_tokens": 100, "output_tokens": 50},
			Result:           "Final result text",
			StructuredOutput: map[string]any{"key": "value"},
		}

		if msg.Usage["input_tokens"] != 100 {
			t.Errorf("Usage[input_tokens] = %v, want 100", msg.Usage["input_tokens"])
		}
		if msg.Result != "Final result text" {
			t.Errorf("Result = %q, want %q", msg.Result, "Final result text")
		}
		if msg.StructuredOutput == nil {
			t.Error("StructuredOutput should not be nil")
		}
	})
}

func TestStreamEvent(t *testing.T) {
	t.Run("create stream event", func(t *testing.T) {
		msg := &StreamEvent{
			UUID:      "event-123",
			SessionID: "session-456",
			Event:     map[string]any{"type": "content_block_delta"},
		}

		if msg.UUID != "event-123" {
			t.Errorf("UUID = %q, want %q", msg.UUID, "event-123")
		}
		if msg.SessionID != "session-456" {
			t.Errorf("SessionID = %q, want %q", msg.SessionID, "session-456")
		}
		if msg.Event["type"] != "content_block_delta" {
			t.Errorf("Event[type] = %v, want 'content_block_delta'", msg.Event["type"])
		}
	})

	t.Run("with parent tool use ID", func(t *testing.T) {
		msg := &StreamEvent{
			UUID:            "event-1",
			SessionID:       "sess-1",
			Event:           map[string]any{},
			ParentToolUseID: "tool-999",
		}

		if msg.ParentToolUseID != "tool-999" {
			t.Errorf("ParentToolUseID = %q, want %q", msg.ParentToolUseID, "tool-999")
		}
	})
}

func TestMessageJSON(t *testing.T) {
	t.Run("marshal UserMessage", func(t *testing.T) {
		msg := &UserMessage{Content: "Hello"}
		data, err := json.Marshal(msg)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		var parsed map[string]any
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if parsed["content"] != "Hello" {
			t.Errorf("JSON content = %v, want 'Hello'", parsed["content"])
		}
	})

	t.Run("marshal AssistantMessage", func(t *testing.T) {
		msg := &AssistantMessage{
			Content: []*ContentBlock{NewTextBlock("Hello")},
			Model:   "claude-opus-4-1-20250805",
		}
		data, err := json.Marshal(msg)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		var parsed map[string]any
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if parsed["model"] != "claude-opus-4-1-20250805" {
			t.Errorf("JSON model = %v, want 'claude-opus-4-1-20250805'", parsed["model"])
		}
	})

	t.Run("marshal ResultMessage", func(t *testing.T) {
		msg := &ResultMessage{
			Subtype:       "success",
			SessionID:     "sess-1",
			TotalCostUSD:  0.01,
			DurationMS:    1000,
			DurationAPIMS: 800,
			NumTurns:      1,
		}
		data, err := json.Marshal(msg)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		var parsed map[string]any
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if parsed["session_id"] != "sess-1" {
			t.Errorf("JSON session_id = %v, want 'sess-1'", parsed["session_id"])
		}
	})
}
