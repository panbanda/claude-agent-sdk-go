package claude

// Message is the interface for all message types in a conversation.
//
// Design rationale: Using an interface with separate concrete types rather than
// a single struct with a discriminator because:
// - Message variants have very different shapes (ResultMessage has ~10 fields
//   that UserMessage doesn't need)
// - Interface provides compile-time type safety via type switches
// - More memory efficient (each type only has its relevant fields)
// - IDE autocomplete only shows relevant fields after type assertion
//
// Use type switch to handle different message types:
//
//	switch m := msg.(type) {
//	case *UserMessage:
//	    fmt.Println(m.Content)
//	case *AssistantMessage:
//	    for _, block := range m.Content { ... }
//	case *ResultMessage:
//	    fmt.Printf("Cost: $%.4f\n", m.TotalCostUSD)
//	}
type Message interface {
	// messageMarker is an unexported method that seals the interface.
	// Only types in this package can implement Message.
	messageMarker()
}

// UserMessage represents a message from the user.
type UserMessage struct {
	// Content is the user's message text.
	Content string `json:"content"`

	// UUID is the unique identifier for this message (optional).
	UUID string `json:"uuid,omitempty"`

	// ParentToolUseID links this message to a tool use (optional).
	ParentToolUseID string `json:"parent_tool_use_id,omitempty"`
}

func (*UserMessage) messageMarker() {}

// AssistantMessage represents a response from the assistant.
type AssistantMessage struct {
	// Content contains the response blocks (text, thinking, tool use, etc.)
	Content []*ContentBlock `json:"content"`

	// Model is the model that generated this response.
	Model string `json:"model"`

	// ParentToolUseID links this message to a tool use (optional).
	ParentToolUseID string `json:"parent_tool_use_id,omitempty"`

	// Error indicates an error type if the response failed (optional).
	// Possible values: "authentication_failed", "billing_error", "rate_limit",
	// "invalid_request", "server_error", "unknown"
	Error string `json:"error,omitempty"`
}

func (*AssistantMessage) messageMarker() {}

// SystemMessage represents a system-level message with metadata.
type SystemMessage struct {
	// Subtype indicates the type of system message.
	Subtype string `json:"subtype"`

	// Data contains the message payload.
	Data map[string]any `json:"data"`
}

func (*SystemMessage) messageMarker() {}

// ResultMessage contains the final result of a query including cost and usage.
type ResultMessage struct {
	// Subtype indicates the result type (e.g., "success").
	Subtype string `json:"subtype"`

	// DurationMS is the total duration in milliseconds.
	DurationMS int `json:"duration_ms"`

	// DurationAPIMS is the API call duration in milliseconds.
	DurationAPIMS int `json:"duration_api_ms"`

	// IsError indicates if the query ended with an error.
	IsError bool `json:"is_error"`

	// NumTurns is the number of conversation turns.
	NumTurns int `json:"num_turns"`

	// SessionID is the session identifier.
	SessionID string `json:"session_id"`

	// TotalCostUSD is the total cost in USD (optional).
	TotalCostUSD float64 `json:"total_cost_usd,omitempty"`

	// Usage contains token usage information (optional).
	Usage map[string]any `json:"usage,omitempty"`

	// Result contains the final text result (optional).
	Result string `json:"result,omitempty"`

	// StructuredOutput contains structured output if requested (optional).
	StructuredOutput any `json:"structured_output,omitempty"`
}

func (*ResultMessage) messageMarker() {}

// StreamEvent represents a streaming event for partial message updates.
type StreamEvent struct {
	// UUID is the unique identifier for this event.
	UUID string `json:"uuid"`

	// SessionID is the session identifier.
	SessionID string `json:"session_id"`

	// Event contains the raw Anthropic API stream event.
	Event map[string]any `json:"event"`

	// ParentToolUseID links this event to a tool use (optional).
	ParentToolUseID string `json:"parent_tool_use_id,omitempty"`
}

func (*StreamEvent) messageMarker() {}
