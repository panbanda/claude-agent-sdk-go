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
