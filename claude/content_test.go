package claude

import (
	"encoding/json"
	"testing"
)

func TestContentBlockKind(t *testing.T) {
	t.Run("kind constants are distinct", func(t *testing.T) {
		kinds := []ContentBlockKind{BlockText, BlockThinking, BlockToolUse, BlockToolResult}
		seen := make(map[ContentBlockKind]bool)
		for _, k := range kinds {
			if seen[k] {
				t.Errorf("duplicate kind value: %d", k)
			}
			seen[k] = true
		}
	})
}

func TestContentBlock_TextBlock(t *testing.T) {
	t.Run("create text block", func(t *testing.T) {
		block := &ContentBlock{
			Kind: BlockText,
			Text: "Hello, human!",
		}

		if block.Kind != BlockText {
			t.Errorf("Kind = %v, want BlockText", block.Kind)
		}
		if block.Text != "Hello, human!" {
			t.Errorf("Text = %q, want %q", block.Text, "Hello, human!")
		}
	})

	t.Run("IsText returns true for text block", func(t *testing.T) {
		block := &ContentBlock{Kind: BlockText, Text: "test"}
		if !block.IsText() {
			t.Error("IsText() should return true for text block")
		}
		if block.IsThinking() {
			t.Error("IsThinking() should return false for text block")
		}
		if block.IsToolUse() {
			t.Error("IsToolUse() should return false for text block")
		}
		if block.IsToolResult() {
			t.Error("IsToolResult() should return false for text block")
		}
	})
}

func TestContentBlock_ThinkingBlock(t *testing.T) {
	t.Run("create thinking block", func(t *testing.T) {
		block := &ContentBlock{
			Kind:      BlockThinking,
			Thinking:  "I'm thinking...",
			Signature: "sig-123",
		}

		if block.Kind != BlockThinking {
			t.Errorf("Kind = %v, want BlockThinking", block.Kind)
		}
		if block.Thinking != "I'm thinking..." {
			t.Errorf("Thinking = %q, want %q", block.Thinking, "I'm thinking...")
		}
		if block.Signature != "sig-123" {
			t.Errorf("Signature = %q, want %q", block.Signature, "sig-123")
		}
	})

	t.Run("IsThinking returns true for thinking block", func(t *testing.T) {
		block := &ContentBlock{Kind: BlockThinking}
		if !block.IsThinking() {
			t.Error("IsThinking() should return true for thinking block")
		}
		if block.IsText() {
			t.Error("IsText() should return false for thinking block")
		}
	})
}

func TestContentBlock_ToolUseBlock(t *testing.T) {
	t.Run("create tool use block", func(t *testing.T) {
		block := &ContentBlock{
			Kind:      BlockToolUse,
			ToolUseID: "tool-123",
			ToolName:  "Read",
			ToolInput: map[string]any{"file_path": "/test.txt"},
		}

		if block.Kind != BlockToolUse {
			t.Errorf("Kind = %v, want BlockToolUse", block.Kind)
		}
		if block.ToolUseID != "tool-123" {
			t.Errorf("ToolUseID = %q, want %q", block.ToolUseID, "tool-123")
		}
		if block.ToolName != "Read" {
			t.Errorf("ToolName = %q, want %q", block.ToolName, "Read")
		}
		if block.ToolInput["file_path"] != "/test.txt" {
			t.Errorf("ToolInput[file_path] = %v, want /test.txt", block.ToolInput["file_path"])
		}
	})

	t.Run("IsToolUse returns true for tool use block", func(t *testing.T) {
		block := &ContentBlock{Kind: BlockToolUse}
		if !block.IsToolUse() {
			t.Error("IsToolUse() should return true for tool use block")
		}
		if block.IsText() {
			t.Error("IsText() should return false for tool use block")
		}
	})
}

func TestContentBlock_ToolResultBlock(t *testing.T) {
	t.Run("create tool result block", func(t *testing.T) {
		block := &ContentBlock{
			Kind:       BlockToolResult,
			ToolUseID:  "tool-123",
			ToolResult: "File contents here",
			IsError:    false,
		}

		if block.Kind != BlockToolResult {
			t.Errorf("Kind = %v, want BlockToolResult", block.Kind)
		}
		if block.ToolUseID != "tool-123" {
			t.Errorf("ToolUseID = %q, want %q", block.ToolUseID, "tool-123")
		}
		if block.ToolResult != "File contents here" {
			t.Errorf("ToolResult = %v, want %q", block.ToolResult, "File contents here")
		}
		if block.IsError != false {
			t.Error("IsError should be false")
		}
	})

	t.Run("create tool result block with error", func(t *testing.T) {
		block := &ContentBlock{
			Kind:       BlockToolResult,
			ToolUseID:  "tool-456",
			ToolResult: "Error: file not found",
			IsError:    true,
		}

		if !block.IsError {
			t.Error("IsError should be true")
		}
	})

	t.Run("IsToolResult returns true for tool result block", func(t *testing.T) {
		block := &ContentBlock{Kind: BlockToolResult}
		if !block.IsToolResult() {
			t.Error("IsToolResult() should return true for tool result block")
		}
	})
}

func TestContentBlock_JSON(t *testing.T) {
	t.Run("marshal text block to JSON", func(t *testing.T) {
		block := &ContentBlock{
			Kind: BlockText,
			Text: "Hello",
		}

		data, err := json.Marshal(block)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		var parsed map[string]any
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		// Should have kind field
		if _, ok := parsed["kind"]; !ok {
			t.Error("JSON should have 'kind' field")
		}
		if parsed["text"] != "Hello" {
			t.Errorf("JSON text = %v, want 'Hello'", parsed["text"])
		}
	})

	t.Run("unmarshal text block from JSON", func(t *testing.T) {
		jsonData := `{"kind":0,"text":"Hello from JSON"}`

		var block ContentBlock
		if err := json.Unmarshal([]byte(jsonData), &block); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if block.Kind != BlockText {
			t.Errorf("Kind = %v, want BlockText", block.Kind)
		}
		if block.Text != "Hello from JSON" {
			t.Errorf("Text = %q, want 'Hello from JSON'", block.Text)
		}
	})

	t.Run("unmarshal tool use block from JSON", func(t *testing.T) {
		jsonData := `{"kind":2,"id":"tool-123","name":"Read","input":{"file_path":"/test.txt"}}`

		var block ContentBlock
		if err := json.Unmarshal([]byte(jsonData), &block); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if block.Kind != BlockToolUse {
			t.Errorf("Kind = %v, want BlockToolUse", block.Kind)
		}
		if block.ToolUseID != "tool-123" {
			t.Errorf("ToolUseID = %q, want 'tool-123'", block.ToolUseID)
		}
	})
}

// Helper constructors should exist for convenience
func TestContentBlockConstructors(t *testing.T) {
	t.Run("NewTextBlock creates text block", func(t *testing.T) {
		block := NewTextBlock("Hello")
		if block.Kind != BlockText {
			t.Errorf("Kind = %v, want BlockText", block.Kind)
		}
		if block.Text != "Hello" {
			t.Errorf("Text = %q, want 'Hello'", block.Text)
		}
	})

	t.Run("NewThinkingBlock creates thinking block", func(t *testing.T) {
		block := NewThinkingBlock("thinking...", "sig")
		if block.Kind != BlockThinking {
			t.Errorf("Kind = %v, want BlockThinking", block.Kind)
		}
		if block.Thinking != "thinking..." {
			t.Errorf("Thinking = %q, want 'thinking...'", block.Thinking)
		}
		if block.Signature != "sig" {
			t.Errorf("Signature = %q, want 'sig'", block.Signature)
		}
	})

	t.Run("NewToolUseBlock creates tool use block", func(t *testing.T) {
		block := NewToolUseBlock("id-1", "Read", map[string]any{"path": "/file"})
		if block.Kind != BlockToolUse {
			t.Errorf("Kind = %v, want BlockToolUse", block.Kind)
		}
		if block.ToolUseID != "id-1" {
			t.Errorf("ToolUseID = %q, want 'id-1'", block.ToolUseID)
		}
		if block.ToolName != "Read" {
			t.Errorf("ToolName = %q, want 'Read'", block.ToolName)
		}
	})

	t.Run("NewToolResultBlock creates tool result block", func(t *testing.T) {
		block := NewToolResultBlock("id-1", "result content", false)
		if block.Kind != BlockToolResult {
			t.Errorf("Kind = %v, want BlockToolResult", block.Kind)
		}
		if block.ToolUseID != "id-1" {
			t.Errorf("ToolUseID = %q, want 'id-1'", block.ToolUseID)
		}
		if block.ToolResult != "result content" {
			t.Errorf("ToolResult = %v, want 'result content'", block.ToolResult)
		}
		if block.IsError != false {
			t.Error("IsError should be false")
		}
	})
}
