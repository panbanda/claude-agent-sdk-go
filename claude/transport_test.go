package claude

import (
	"context"
	"errors"
	"testing"
)

func TestTransportInterface(t *testing.T) {
	t.Run("interface has required methods", func(t *testing.T) {
		// Compile-time check that the interface has all required methods.
		// This test passes if the code compiles.
		var _ Transport = (*mockTransport)(nil)
	})
}

func TestMockTransport(t *testing.T) {
	t.Run("Connect succeeds by default", func(t *testing.T) {
		mt := newMockTransport()

		err := mt.Connect(context.Background())

		if err != nil {
			t.Errorf("Connect() error = %v, want nil", err)
		}
		if !mt.IsReady() {
			t.Error("IsReady() = false after Connect(), want true")
		}
	})

	t.Run("Connect returns configured error", func(t *testing.T) {
		mt := newMockTransport()
		mt.connectErr = ErrCLINotFound

		err := mt.Connect(context.Background())

		if !errors.Is(err, ErrCLINotFound) {
			t.Errorf("Connect() error = %v, want %v", err, ErrCLINotFound)
		}
	})

	t.Run("Send writes to buffer", func(t *testing.T) {
		mt := newMockTransport()
		_ = mt.Connect(context.Background())

		data := []byte(`{"type": "user", "content": "hello"}`)
		err := mt.Send(context.Background(), data)

		if err != nil {
			t.Errorf("Send() error = %v, want nil", err)
		}
		if len(mt.sentMessages) != 1 {
			t.Fatalf("sentMessages length = %d, want 1", len(mt.sentMessages))
		}
		if string(mt.sentMessages[0]) != string(data) {
			t.Errorf("sentMessages[0] = %q, want %q", mt.sentMessages[0], data)
		}
	})

	t.Run("Send fails when not connected", func(t *testing.T) {
		mt := newMockTransport()

		err := mt.Send(context.Background(), []byte("test"))

		if !errors.Is(err, ErrNotConnected) {
			t.Errorf("Send() error = %v, want %v", err, ErrNotConnected)
		}
	})

	t.Run("Messages returns channel", func(t *testing.T) {
		mt := newMockTransport()
		_ = mt.Connect(context.Background())

		// Queue a message
		mt.QueueMessage([]byte(`{"type": "assistant"}`))
		mt.CloseMessages()

		msgs := mt.Messages()
		if msgs == nil {
			t.Fatal("Messages() returned nil channel")
		}

		msg, ok := <-msgs
		if !ok {
			t.Fatal("channel closed without message")
		}
		if string(msg) != `{"type": "assistant"}` {
			t.Errorf("message = %q, want %q", msg, `{"type": "assistant"}`)
		}
	})

	t.Run("Errors returns channel", func(t *testing.T) {
		mt := newMockTransport()
		_ = mt.Connect(context.Background())

		testErr := errors.New("test error")
		mt.QueueError(testErr)
		mt.CloseErrors()

		errs := mt.Errors()
		if errs == nil {
			t.Fatal("Errors() returned nil channel")
		}

		err, ok := <-errs
		if !ok {
			t.Fatal("channel closed without error")
		}
		if err != testErr {
			t.Errorf("error = %v, want %v", err, testErr)
		}
	})

	t.Run("Close cleans up", func(t *testing.T) {
		mt := newMockTransport()
		_ = mt.Connect(context.Background())

		err := mt.Close()

		if err != nil {
			t.Errorf("Close() error = %v, want nil", err)
		}
		if mt.IsReady() {
			t.Error("IsReady() = true after Close(), want false")
		}
	})

	t.Run("context cancellation stops operations", func(t *testing.T) {
		mt := newMockTransport()
		_ = mt.Connect(context.Background())

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := mt.Send(ctx, []byte("test"))

		if !errors.Is(err, context.Canceled) {
			t.Errorf("Send() error = %v, want %v", err, context.Canceled)
		}
	})
}

// mockTransport is a test implementation of Transport.
type mockTransport struct {
	ready        bool
	connectErr   error
	sendErr      error
	closeErr     error
	sentMessages [][]byte
	messagesCh   chan []byte
	errorsCh     chan error
}

func newMockTransport() *mockTransport {
	return &mockTransport{
		messagesCh: make(chan []byte, 10),
		errorsCh:   make(chan error, 10),
	}
}

func (m *mockTransport) Connect(ctx context.Context) error {
	if m.connectErr != nil {
		return m.connectErr
	}
	m.ready = true
	return nil
}

func (m *mockTransport) Send(ctx context.Context, data []byte) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if !m.ready {
		return ErrNotConnected
	}
	if m.sendErr != nil {
		return m.sendErr
	}
	m.sentMessages = append(m.sentMessages, data)
	return nil
}

func (m *mockTransport) Messages() <-chan []byte {
	return m.messagesCh
}

func (m *mockTransport) Errors() <-chan error {
	return m.errorsCh
}

func (m *mockTransport) Close() error {
	m.ready = false
	return m.closeErr
}

func (m *mockTransport) IsReady() bool {
	return m.ready
}

// Test helpers

func (m *mockTransport) QueueMessage(data []byte) {
	m.messagesCh <- data
}

func (m *mockTransport) CloseMessages() {
	close(m.messagesCh)
}

func (m *mockTransport) QueueError(err error) {
	m.errorsCh <- err
}

func (m *mockTransport) CloseErrors() {
	close(m.errorsCh)
}
