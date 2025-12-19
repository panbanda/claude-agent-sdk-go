package claude

import (
	"context"
	"encoding/json"
	"sync"
)

// Client provides bidirectional communication with the Claude CLI.
//
// For simple one-shot queries, use the Query() convenience function instead.
// The Client is designed for interactive conversations where you need to:
//   - Send multiple messages in a session
//   - React to Claude's responses dynamically
//   - Use streaming for real-time output
//
// Example:
//
//	client := claude.NewClient(
//	    claude.WithModel("claude-sonnet-4-5"),
//	    claude.WithMaxTurns(10),
//	)
//
//	if err := client.Connect(ctx); err != nil {
//	    return err
//	}
//	defer client.Close()
//
//	if err := client.Query(ctx, "What is 2+2?"); err != nil {
//	    return err
//	}
//
//	for msg := range client.Messages() {
//	    switch m := msg.(type) {
//	    case *AssistantMessage:
//	        // handle response
//	    case *ResultMessage:
//	        // query complete
//	    }
//	}
type Client struct {
	cfg       *config
	transport Transport
	messages  chan Message
	connected bool
	mu        sync.RWMutex
}

// NewClient creates a new Claude client with the given options.
func NewClient(opts ...Option) *Client {
	cfg := &config{}
	for _, opt := range opts {
		opt(cfg)
	}

	return &Client{
		cfg: cfg,
	}
}

// Connect establishes a connection to the Claude CLI.
// It must be called before Query or Messages.
func (c *Client) Connect(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Use configured transport or create default subprocess transport
	if c.cfg.transport != nil {
		c.transport = c.cfg.transport
	} else {
		c.transport = NewSubprocessTransport(c.cfg)
	}

	if err := c.transport.Connect(ctx); err != nil {
		return err
	}

	// Create message parsing goroutine
	c.messages = make(chan Message, 100)
	go c.readMessages()

	c.connected = true
	return nil
}

// readMessages reads from transport and parses into Message types.
func (c *Client) readMessages() {
	defer close(c.messages)

	for data := range c.transport.Messages() {
		msg := c.parseMessage(data)
		if msg != nil {
			c.messages <- msg
		}
	}
}

// parseMessage converts raw JSON into a Message type.
// Returns nil if the message cannot be parsed.
func (c *Client) parseMessage(data []byte) Message {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil
	}

	msgType, _ := raw["type"].(string)

	switch msgType {
	case "user":
		return c.parseUserMessage(raw)
	case "assistant":
		return c.parseAssistantMessage(raw)
	case "system":
		return c.parseSystemMessage(raw)
	case "result":
		return c.parseResultMessage(raw)
	default:
		return nil
	}
}

func (c *Client) parseUserMessage(raw map[string]any) *UserMessage {
	msg := &UserMessage{}

	if m, ok := raw["message"].(map[string]any); ok {
		if content, ok := m["content"].(string); ok {
			msg.Content = content
		}
	}

	if uuid, ok := raw["uuid"].(string); ok {
		msg.UUID = uuid
	}

	if parentID, ok := raw["parent_tool_use_id"].(string); ok {
		msg.ParentToolUseID = parentID
	}

	return msg
}

func (c *Client) parseAssistantMessage(raw map[string]any) *AssistantMessage {
	msg := &AssistantMessage{}

	if m, ok := raw["message"].(map[string]any); ok {
		if model, ok := m["model"].(string); ok {
			msg.Model = model
		}

		if content, ok := m["content"].([]any); ok {
			msg.Content = c.parseContentBlocks(content)
		}
	}

	if parentID, ok := raw["parent_tool_use_id"].(string); ok {
		msg.ParentToolUseID = parentID
	}

	if errStr, ok := raw["error"].(string); ok {
		msg.Error = errStr
	}

	return msg
}

func (c *Client) parseContentBlocks(content []any) []*ContentBlock {
	blocks := make([]*ContentBlock, 0, len(content))

	for _, item := range content {
		block, ok := item.(map[string]any)
		if !ok {
			continue
		}

		blockType, _ := block["type"].(string)
		switch blockType {
		case "text":
			text, _ := block["text"].(string)
			blocks = append(blocks, NewTextBlock(text))

		case "thinking":
			thinking, _ := block["thinking"].(string)
			signature, _ := block["signature"].(string)
			blocks = append(blocks, NewThinkingBlock(thinking, signature))

		case "tool_use":
			id, _ := block["id"].(string)
			name, _ := block["name"].(string)
			input, _ := block["input"].(map[string]any)
			blocks = append(blocks, NewToolUseBlock(id, name, input))

		case "tool_result":
			toolUseID, _ := block["tool_use_id"].(string)
			result := block["content"]
			isError, _ := block["is_error"].(bool)
			blocks = append(blocks, NewToolResultBlock(toolUseID, result, isError))
		}
	}

	return blocks
}

func (c *Client) parseSystemMessage(raw map[string]any) *SystemMessage {
	msg := &SystemMessage{
		Data: make(map[string]any),
	}

	if subtype, ok := raw["subtype"].(string); ok {
		msg.Subtype = subtype
	}

	if data, ok := raw["data"].(map[string]any); ok {
		msg.Data = data
	}

	return msg
}

func (c *Client) parseResultMessage(raw map[string]any) *ResultMessage {
	msg := &ResultMessage{}

	if subtype, ok := raw["subtype"].(string); ok {
		msg.Subtype = subtype
	}

	if duration, ok := raw["duration_ms"].(float64); ok {
		msg.DurationMS = int(duration)
	}

	if duration, ok := raw["duration_api_ms"].(float64); ok {
		msg.DurationAPIMS = int(duration)
	}

	if isError, ok := raw["is_error"].(bool); ok {
		msg.IsError = isError
	}

	if numTurns, ok := raw["num_turns"].(float64); ok {
		msg.NumTurns = int(numTurns)
	}

	if sessionID, ok := raw["session_id"].(string); ok {
		msg.SessionID = sessionID
	}

	if cost, ok := raw["total_cost_usd"].(float64); ok {
		msg.TotalCostUSD = cost
	}

	if usage, ok := raw["usage"].(map[string]any); ok {
		msg.Usage = usage
	}

	if result, ok := raw["result"].(string); ok {
		msg.Result = result
	}

	if output, ok := raw["structured_output"]; ok {
		msg.StructuredOutput = output
	}

	return msg
}

// Close disconnects from the Claude CLI and releases resources.
// It is safe to call Close multiple times.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil
	}

	c.connected = false

	if c.transport != nil {
		return c.transport.Close()
	}

	return nil
}

// Query sends a prompt to Claude.
// Connect must be called before Query.
func (c *Client) Query(ctx context.Context, prompt string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	c.mu.RLock()
	if !c.connected {
		c.mu.RUnlock()
		return ErrNotConnected
	}
	transport := c.transport
	c.mu.RUnlock()

	msg := map[string]any{
		"type": "user",
		"message": map[string]any{
			"role":    "user",
			"content": prompt,
		},
		"parent_tool_use_id": nil,
		"session_id":         "default",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// Append newline for JSONL format
	data = append(data, '\n')

	return transport.Send(ctx, data)
}

// Messages returns a channel that receives parsed messages from Claude.
// Returns nil if the client is not connected.
func (c *Client) Messages() <-chan Message {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil
	}

	return c.messages
}

// IsConnected returns true if the client is connected to Claude.
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}
