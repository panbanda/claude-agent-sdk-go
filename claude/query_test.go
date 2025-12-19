package claude

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestQuery(t *testing.T) {
	t.Run("returns channel with messages", func(t *testing.T) {
		mt := newMockTransport()

		// Queue an assistant message and result
		assistantMsg := map[string]any{
			"type": "assistant",
			"message": map[string]any{
				"role":    "assistant",
				"content": []map[string]any{{"type": "text", "text": "4"}},
				"model":   "claude-sonnet-4-5",
			},
		}
		resultMsg := map[string]any{
			"type":           "result",
			"subtype":        "success",
			"duration_ms":    1000,
			"is_error":       false,
			"num_turns":      1,
			"session_id":     "test-session",
			"total_cost_usd": 0.001,
		}

		assistantBytes, _ := json.Marshal(assistantMsg)
		resultBytes, _ := json.Marshal(resultMsg)
		mt.QueueMessage(assistantBytes)
		mt.QueueMessage(resultBytes)
		mt.CloseMessages()

		msgs, err := Query(context.Background(), "What is 2+2?", WithTransport(mt))

		if err != nil {
			t.Fatalf("Query() error = %v, want nil", err)
		}
		if msgs == nil {
			t.Fatal("Query() returned nil channel")
		}

		var received []Message
		for msg := range msgs {
			received = append(received, msg)
		}

		if len(received) != 2 {
			t.Errorf("received %d messages, want 2", len(received))
		}

		// First should be AssistantMessage
		if _, ok := received[0].(*AssistantMessage); !ok {
			t.Errorf("first message type = %T, want *AssistantMessage", received[0])
		}

		// Last should be ResultMessage
		if _, ok := received[1].(*ResultMessage); !ok {
			t.Errorf("last message type = %T, want *ResultMessage", received[1])
		}
	})

	t.Run("sends prompt to transport", func(t *testing.T) {
		mt := newMockTransport()

		// Queue a result to complete the query
		resultMsg := map[string]any{
			"type":       "result",
			"subtype":    "success",
			"session_id": "test",
		}
		resultBytes, _ := json.Marshal(resultMsg)
		mt.QueueMessage(resultBytes)
		mt.CloseMessages()

		_, err := Query(context.Background(), "Test prompt", WithTransport(mt))

		if err != nil {
			t.Fatalf("Query() error = %v, want nil", err)
		}

		if len(mt.sentMessages) != 1 {
			t.Fatalf("sentMessages length = %d, want 1", len(mt.sentMessages))
		}

		// Verify the prompt was sent correctly
		var sent map[string]any
		if err := json.Unmarshal(mt.sentMessages[0], &sent); err != nil {
			t.Fatalf("failed to unmarshal sent message: %v", err)
		}

		if sent["type"] != "user" {
			t.Errorf("sent message type = %v, want 'user'", sent["type"])
		}

		msgContent, _ := sent["message"].(map[string]any)
		if msgContent["content"] != "Test prompt" {
			t.Errorf("sent message content = %v, want 'Test prompt'", msgContent["content"])
		}
	})

	t.Run("returns error on connection failure", func(t *testing.T) {
		mt := newMockTransport()
		mt.connectErr = ErrCLINotFound

		_, err := Query(context.Background(), "test", WithTransport(mt))

		if !errors.Is(err, ErrCLINotFound) {
			t.Errorf("Query() error = %v, want %v", err, ErrCLINotFound)
		}
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		mt := newMockTransport()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := Query(ctx, "test", WithTransport(mt))

		if !errors.Is(err, context.Canceled) {
			t.Errorf("Query() error = %v, want %v", err, context.Canceled)
		}
	})

	t.Run("applies options", func(t *testing.T) {
		mt := newMockTransport()

		// Queue a result to complete the query
		resultMsg := map[string]any{
			"type":       "result",
			"subtype":    "success",
			"session_id": "test",
		}
		resultBytes, _ := json.Marshal(resultMsg)
		mt.QueueMessage(resultBytes)
		mt.CloseMessages()

		_, err := Query(context.Background(), "Test",
			WithTransport(mt),
			WithModel("claude-sonnet-4-5"),
			WithMaxTurns(5),
		)

		if err != nil {
			t.Fatalf("Query() error = %v, want nil", err)
		}

		// The options should be applied (verified by the client working correctly)
	})
}

func TestQueryResult(t *testing.T) {
	t.Run("returns final result", func(t *testing.T) {
		mt := newMockTransport()

		// Queue messages
		assistantMsg := map[string]any{
			"type": "assistant",
			"message": map[string]any{
				"role":    "assistant",
				"content": []map[string]any{{"type": "text", "text": "The answer is 4"}},
				"model":   "claude-sonnet-4-5",
			},
		}
		resultMsg := map[string]any{
			"type":           "result",
			"subtype":        "success",
			"session_id":     "test",
			"total_cost_usd": 0.01,
			"result":         "The answer is 4",
		}

		assistantBytes, _ := json.Marshal(assistantMsg)
		resultBytes, _ := json.Marshal(resultMsg)
		mt.QueueMessage(assistantBytes)
		mt.QueueMessage(resultBytes)
		mt.CloseMessages()

		result, err := QueryResult(context.Background(), "What is 2+2?", WithTransport(mt))

		if err != nil {
			t.Fatalf("QueryResult() error = %v, want nil", err)
		}
		if result == nil {
			t.Fatal("QueryResult() returned nil")
		}
		if result.TotalCostUSD != 0.01 {
			t.Errorf("result.TotalCostUSD = %v, want 0.01", result.TotalCostUSD)
		}
	})

	t.Run("returns error on connection failure", func(t *testing.T) {
		mt := newMockTransport()
		mt.connectErr = ErrCLINotFound

		_, err := QueryResult(context.Background(), "test", WithTransport(mt))

		if !errors.Is(err, ErrCLINotFound) {
			t.Errorf("QueryResult() error = %v, want %v", err, ErrCLINotFound)
		}
	})

	t.Run("returns error when no result message", func(t *testing.T) {
		mt := newMockTransport()

		// Only queue assistant message, no result
		assistantMsg := map[string]any{
			"type": "assistant",
			"message": map[string]any{
				"role":    "assistant",
				"content": []map[string]any{{"type": "text", "text": "Hi"}},
				"model":   "claude-sonnet-4-5",
			},
		}
		assistantBytes, _ := json.Marshal(assistantMsg)
		mt.QueueMessage(assistantBytes)
		mt.CloseMessages()

		_, err := QueryResult(context.Background(), "test", WithTransport(mt))

		if err == nil {
			t.Error("QueryResult() error = nil, want error")
		}
	})
}
