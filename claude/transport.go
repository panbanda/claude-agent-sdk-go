package claude

import (
	"context"
)

// Transport is the interface for CLI communication.
// Implementations handle the actual process management and I/O.
//
// The transport lifecycle is:
//  1. Create transport (implementation-specific)
//  2. Connect(ctx) - start the subprocess and establish communication
//  3. Send(ctx, data) - write messages to the CLI
//  4. Read from Messages() channel - receive CLI responses
//  5. Close() - shut down the transport
//
// Example:
//
//	transport := NewSubprocessTransport(opts...)
//	if err := transport.Connect(ctx); err != nil {
//	    return err
//	}
//	defer transport.Close()
//
//	for msg := range transport.Messages() {
//	    // handle message
//	}
type Transport interface {
	// Connect starts the subprocess and establishes communication.
	// It should be called before Send or Messages.
	Connect(ctx context.Context) error

	// Send writes a message to the CLI.
	// Returns ErrNotConnected if Connect has not been called.
	Send(ctx context.Context, data []byte) error

	// Messages returns a channel of incoming messages from the CLI.
	// The channel is closed when the transport is closed or an error occurs.
	Messages() <-chan []byte

	// Errors returns a channel of transport errors.
	// Non-fatal errors are sent here; fatal errors close the Messages channel.
	Errors() <-chan error

	// Close shuts down the transport and releases resources.
	// It is safe to call Close multiple times.
	Close() error

	// IsReady returns true if the transport is connected and ready for communication.
	IsReady() bool
}
