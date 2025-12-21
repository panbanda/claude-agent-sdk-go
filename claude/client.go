// Package claude provides a Go SDK for interacting with the Claude CLI.
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
	cfg        *config
	transport  Transport
	messages   chan Message
	connected  bool
	serverInfo map[string]any
	mu         sync.RWMutex
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

	// Send initialize request if hooks are configured
	if len(c.cfg.hooks) > 0 {
		if err := c.sendInitialize(ctx); err != nil {
			c.connected = false
			_ = c.transport.Close()
			return err
		}
	}

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
// Returns nil if the message cannot be parsed or if it was handled internally.
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
	case "stream_event":
		return c.parseStreamEvent(raw)
	case MessageTypeControlRequest:
		c.handleControlRequest(raw)
		return nil
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

		if msg.Subtype == "init" {
			c.mu.Lock()
			c.serverInfo = data
			c.mu.Unlock()
		}
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

func (c *Client) parseStreamEvent(raw map[string]any) *StreamEvent {
	event := &StreamEvent{}

	if uuid, ok := raw["uuid"].(string); ok {
		event.UUID = uuid
	}
	if sessionID, ok := raw["session_id"].(string); ok {
		event.SessionID = sessionID
	}
	if e, ok := raw["event"].(map[string]any); ok {
		event.Event = e
	}
	if parentID, ok := raw["parent_tool_use_id"].(string); ok {
		event.ParentToolUseID = parentID
	}

	return event
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

// Interrupt sends an interrupt signal to stop the current operation.
// This is only effective during an active query.
func (c *Client) Interrupt(ctx context.Context) error {
	c.mu.RLock()
	if !c.connected {
		c.mu.RUnlock()
		return ErrNotConnected
	}
	transport := c.transport
	c.mu.RUnlock()

	req := &ControlRequest{
		Type:      MessageTypeControlRequest,
		RequestID: generateRequestID(),
		Request: &ControlRequestBody{
			Subtype: ControlSubtypeInterrupt,
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		return err
	}
	data = append(data, '\n')

	return transport.Send(ctx, data)
}

// SetPermissionMode changes the permission mode during a conversation.
// Valid modes: "default", "acceptEdits", "plan", "bypassPermissions".
func (c *Client) SetPermissionMode(ctx context.Context, mode PermissionMode) error {
	c.mu.RLock()
	if !c.connected {
		c.mu.RUnlock()
		return ErrNotConnected
	}
	transport := c.transport
	c.mu.RUnlock()

	req := &ControlRequest{
		Type:      MessageTypeControlRequest,
		RequestID: generateRequestID(),
		Request: &ControlRequestBody{
			Subtype: ControlSubtypeSetPermissionMode,
			Mode:    string(mode),
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		return err
	}
	data = append(data, '\n')

	return transport.Send(ctx, data)
}

// SetModel changes the AI model during a conversation.
// Pass empty string to use the default model.
func (c *Client) SetModel(ctx context.Context, model string) error {
	c.mu.RLock()
	if !c.connected {
		c.mu.RUnlock()
		return ErrNotConnected
	}
	transport := c.transport
	c.mu.RUnlock()

	var modelPtr *string
	if model != "" {
		modelPtr = &model
	}

	req := &ControlRequest{
		Type:      MessageTypeControlRequest,
		RequestID: generateRequestID(),
		Request: &ControlRequestBody{
			Subtype: ControlSubtypeSetModel,
			Model:   modelPtr,
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		return err
	}
	data = append(data, '\n')

	return transport.Send(ctx, data)
}

// GetServerInfo returns server initialization info including available
// commands and output styles. Returns nil if not available.
func (c *Client) GetServerInfo() map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.serverInfo
}

// sendInitialize sends an initialize request with hook configurations to the CLI.
func (c *Client) sendInitialize(ctx context.Context) error {
	// Build hook definitions for the CLI
	hookDefs := make(map[HookEvent][]InitializeHookDef)

	for event, matchers := range c.cfg.hooks {
		for _, m := range matchers {
			def := InitializeHookDef{
				Matcher:         m.matcher,
				HookCallbackIDs: m.callbackIDs,
			}
			if m.timeout > 0 {
				timeoutSecs := int(m.timeout.Seconds())
				def.Timeout = &timeoutSecs
			}
			hookDefs[event] = append(hookDefs[event], def)
		}
	}

	req := &ControlRequest{
		Type:      MessageTypeControlRequest,
		RequestID: generateRequestID(),
		Request: &ControlRequestBody{
			Subtype:      ControlSubtypeInitialize,
			InitHookDefs: hookDefs,
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		return err
	}
	data = append(data, '\n')

	return c.transport.Send(ctx, data)
}

// handleControlRequest processes a control request from the CLI.
func (c *Client) handleControlRequest(raw map[string]any) {
	requestID, _ := raw["request_id"].(string)
	request, ok := raw["request"].(map[string]any)
	if !ok {
		return
	}

	subtype, _ := request["subtype"].(string)
	if subtype != "hook_callback" {
		return
	}

	callbackID, _ := request["callback_id"].(string)
	input, _ := request["input"].(map[string]any)

	// Look up the callback
	if c.cfg.hookCallbacks == nil {
		return
	}
	callback, ok := c.cfg.hookCallbacks[callbackID]
	if !ok {
		return
	}

	// Extract hook event name to determine how to invoke
	hookEventName, _ := input["hook_event_name"].(string)

	var response *HookCallbackResponse
	ctx := context.Background()
	hookCtx := &HookContext{}

	switch hookEventName {
	case "PreToolUse":
		if hook, ok := callback.(PreToolUseHook); ok {
			hookInput := &PreToolUseInput{
				ToolName:  getString(input, "tool_name"),
				ToolInput: getMap(input, "tool_input"),
				ToolUseID: getString(input, "tool_use_id"),
			}
			output, err := hook(ctx, hookInput, hookCtx)
			response = c.buildHookResponse(output, err, PreToolUse)
		}
	case "PostToolUse":
		if hook, ok := callback.(PostToolUseHook); ok {
			hookInput := &PostToolUseInput{
				ToolName:     getString(input, "tool_name"),
				ToolInput:    getMap(input, "tool_input"),
				ToolUseID:    getString(input, "tool_use_id"),
				ToolResponse: input["tool_response"],
				IsError:      getBool(input, "is_error"),
			}
			output, err := hook(ctx, hookInput, hookCtx)
			response = c.buildHookResponse(output, err, PostToolUse)
		}
	}

	if response == nil {
		response = &HookCallbackResponse{Continue: true}
	}

	// Send the response
	c.sendControlResponse(requestID, response)
}

func (c *Client) buildHookResponse(output *HookOutput, err error, event HookEvent) *HookCallbackResponse {
	if err != nil || output == nil {
		return &HookCallbackResponse{Continue: true}
	}

	resp := &HookCallbackResponse{
		Continue: true,
	}

	if output.Continue != nil {
		resp.Continue = *output.Continue
	}

	resp.StopReason = output.StopReason
	resp.SystemMessage = output.SystemMessage
	resp.Reason = output.Reason

	if output.Decision != HookDecisionNone {
		resp.HookSpecificOutput = &HookSpecificOutput{
			HookEventName:     event,
			UpdatedInput:      output.UpdatedInput,
			AdditionalContext: output.AdditionalContext,
		}

		switch output.Decision {
		case HookDecisionAllow:
			resp.HookSpecificOutput.PermissionDecision = string(HookDecisionAllow)
		case HookDecisionDeny:
			resp.HookSpecificOutput.PermissionDecision = string(HookDecisionDeny)
			resp.HookSpecificOutput.PermissionDecisionReason = output.Reason
		case HookDecisionNone:
			// Already handled by outer if check
		}
	}

	return resp
}

func (c *Client) sendControlResponse(requestID string, response *HookCallbackResponse) {
	resp := NewControlResponseSuccess(requestID, response)
	data, err := json.Marshal(resp)
	if err != nil {
		return
	}
	data = append(data, '\n')

	c.mu.RLock()
	transport := c.transport
	c.mu.RUnlock()

	if transport != nil {
		_ = transport.Send(context.Background(), data)
	}
}

func getString(m map[string]any, key string) string {
	v, _ := m[key].(string)
	return v
}

func getMap(m map[string]any, key string) map[string]any {
	v, _ := m[key].(map[string]any)
	return v
}

func getBool(m map[string]any, key string) bool {
	v, _ := m[key].(bool)
	return v
}
