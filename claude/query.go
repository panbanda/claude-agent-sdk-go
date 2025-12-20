package claude

import (
	"context"
	"errors"
)

// ErrNoResult is returned when a query completes without a result message.
var ErrNoResult = errors.New("claude: query completed without result message")

// Query sends a prompt to Claude and returns a channel of messages.
// This is a convenience function for simple, one-shot queries.
// For interactive conversations, use Client directly.
//
// The returned channel receives all messages until the query completes.
// The channel is closed when the query completes or an error occurs.
//
// Example:
//
//	msgs, err := claude.Query(ctx, "What is 2+2?",
//	    claude.WithModel("claude-sonnet-4-5"),
//	)
//	if err != nil {
//	    return err
//	}
//
//	for msg := range msgs {
//	    switch m := msg.(type) {
//	    case *AssistantMessage:
//	        // handle response
//	    case *ResultMessage:
//	        fmt.Printf("Cost: $%.4f\n", m.TotalCostUSD)
//	    }
//	}
func Query(ctx context.Context, prompt string, opts ...Option) (<-chan Message, error) {
	client := NewClient(opts...)

	if err := client.Connect(ctx); err != nil {
		return nil, err
	}

	if err := client.Query(ctx, prompt); err != nil {
		_ = client.Close()
		return nil, err
	}

	// Create output channel
	out := make(chan Message, 100)

	// Start goroutine to read messages and clean up
	go func() {
		defer close(out)
		defer func() { _ = client.Close() }()

		msgs := client.Messages()
		if msgs == nil {
			return
		}

		for msg := range msgs {
			select {
			case <-ctx.Done():
				return
			case out <- msg:
				// Close after sending ResultMessage - query is complete
				if _, isResult := msg.(*ResultMessage); isResult {
					return
				}
			}
		}
	}()

	return out, nil
}

// QueryResult sends a prompt to Claude and returns the final ResultMessage.
// This is a convenience function for simple queries where you only need
// the final result, not intermediate messages.
//
// Example:
//
//	result, err := claude.QueryResult(ctx, "What is 2+2?",
//	    claude.WithModel("claude-sonnet-4-5"),
//	)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Cost: $%.4f\n", result.TotalCostUSD)
func QueryResult(ctx context.Context, prompt string, opts ...Option) (*ResultMessage, error) {
	msgs, err := Query(ctx, prompt, opts...)
	if err != nil {
		return nil, err
	}

	var result *ResultMessage
	for msg := range msgs {
		if r, ok := msg.(*ResultMessage); ok {
			result = r
		}
	}

	if result == nil {
		return nil, ErrNoResult
	}

	return result, nil
}
