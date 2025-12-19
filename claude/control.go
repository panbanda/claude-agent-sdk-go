package claude

import (
	"crypto/rand"
	"encoding/hex"
)

// ControlRequestSubtype identifies the type of control request.
type ControlRequestSubtype string

// MessageTypeControlRequest is the type field value for control requests.
const MessageTypeControlRequest = "control_request"

const (
	// ControlSubtypeInterrupt sends an interrupt signal.
	ControlSubtypeInterrupt ControlRequestSubtype = "interrupt"

	// ControlSubtypeCanUseTool requests permission to use a tool.
	ControlSubtypeCanUseTool ControlRequestSubtype = "can_use_tool"

	// ControlSubtypeInitialize initializes the SDK with hooks.
	ControlSubtypeInitialize ControlRequestSubtype = "initialize"

	// ControlSubtypeSetPermissionMode changes the permission mode.
	ControlSubtypeSetPermissionMode ControlRequestSubtype = "set_permission_mode"

	// ControlSubtypeHookCallback handles a hook callback.
	ControlSubtypeHookCallback ControlRequestSubtype = "hook_callback"

	// ControlSubtypeMcpMessage sends an MCP message.
	ControlSubtypeMcpMessage ControlRequestSubtype = "mcp_message"

	// ControlSubtypeRewindFiles rewinds files to a previous state.
	ControlSubtypeRewindFiles ControlRequestSubtype = "rewind_files"
)

// ControlRequest is a message sent from SDK to CLI.
type ControlRequest struct {
	Type      string              `json:"type"`
	RequestID string              `json:"request_id"`
	Request   *ControlRequestBody `json:"request"`
}

// ControlRequestBody holds the request details.
type ControlRequestBody struct {
	Subtype ControlRequestSubtype `json:"subtype"`

	// For can_use_tool
	ToolName              string         `json:"tool_name,omitempty"`
	Input                 map[string]any `json:"input,omitempty"`
	PermissionSuggestions []any          `json:"permission_suggestions,omitempty"`
	BlockedPath           string         `json:"blocked_path,omitempty"`

	// For initialize (use InitHookDefs for the actual hook config sent to CLI)
	Hooks        map[HookEvent][]HookDefinition    `json:"-"`
	InitHookDefs map[HookEvent][]InitializeHookDef `json:"hooks,omitempty"`

	// For set_permission_mode
	Mode string `json:"mode,omitempty"`

	// For hook_callback
	CallbackID string `json:"callback_id,omitempty"`
	HookInput  any    `json:"hook_input,omitempty"`
	ToolUseID  string `json:"tool_use_id,omitempty"`

	// For mcp_message
	ServerName string `json:"server_name,omitempty"`
	Message    any    `json:"message,omitempty"`

	// For rewind_files
	UserMessageID string `json:"user_message_id,omitempty"`
}

// HookDefinition describes a hook registration for the CLI.
type HookDefinition struct {
	Matcher string `json:"matcher,omitempty"`
	Timeout int    `json:"timeout,omitempty"`
}

// InitializeHookDef describes a hook for the initialize request.
type InitializeHookDef struct {
	Matcher         string   `json:"matcher,omitempty"`
	HookCallbackIDs []string `json:"hookCallbackIds"`
	Timeout         *int     `json:"timeout,omitempty"`
}

// ControlResponse is a message received from CLI in response to a request.
type ControlResponse struct {
	Type     string                  `json:"type"`
	Response *ControlResponsePayload `json:"response"`
}

// ControlResponsePayload holds the response details.
type ControlResponsePayload struct {
	Subtype   string `json:"subtype"`
	RequestID string `json:"request_id"`
	Response  any    `json:"response,omitempty"`
	Error     string `json:"error,omitempty"`
}

// PermissionResultResponse is the response to a can_use_tool request.
type PermissionResultResponse struct {
	Behavior     string         `json:"behavior"`
	Message      string         `json:"message,omitempty"`
	Interrupt    bool           `json:"interrupt,omitempty"`
	UpdatedInput map[string]any `json:"updated_input,omitempty"`
}

// HookCallbackResponse is the response to a hook_callback request.
type HookCallbackResponse struct {
	Continue           bool                `json:"continue"`
	SuppressOutput     bool                `json:"suppressOutput,omitempty"`
	StopReason         string              `json:"stopReason,omitempty"`
	Decision           string              `json:"decision,omitempty"`
	SystemMessage      string              `json:"systemMessage,omitempty"`
	Reason             string              `json:"reason,omitempty"`
	HookSpecificOutput *HookSpecificOutput `json:"hookSpecificOutput,omitempty"`
}

// HookSpecificOutput contains event-specific output fields.
type HookSpecificOutput struct {
	HookEventName            HookEvent      `json:"hookEventName"`
	PermissionDecision       string         `json:"permissionDecision,omitempty"`
	PermissionDecisionReason string         `json:"permissionDecisionReason,omitempty"`
	UpdatedInput             map[string]any `json:"updatedInput,omitempty"`
	AdditionalContext        string         `json:"additionalContext,omitempty"`
}

// generateRequestID creates a unique request ID.
func generateRequestID() string {
	bytes := make([]byte, 8)
	_, _ = rand.Read(bytes)
	return "req-" + hex.EncodeToString(bytes)
}

// NewInterruptRequest creates a new interrupt control request.
func NewInterruptRequest() *ControlRequest {
	return &ControlRequest{
		Type:      MessageTypeControlRequest,
		RequestID: generateRequestID(),
		Request: &ControlRequestBody{
			Subtype: ControlSubtypeInterrupt,
		},
	}
}

// NewInitializeRequest creates an initialize request with hook definitions.
func NewInitializeRequest(hooks map[HookEvent][]HookDefinition) *ControlRequest {
	// Convert HookDefinition to InitializeHookDef
	initHooks := make(map[HookEvent][]InitializeHookDef)
	for event, defs := range hooks {
		for _, def := range defs {
			initDef := InitializeHookDef{
				Matcher:         def.Matcher,
				HookCallbackIDs: []string{}, // Empty for direct use
			}
			if def.Timeout > 0 {
				timeout := def.Timeout // Copy to avoid memory aliasing
				initDef.Timeout = &timeout
			}
			initHooks[event] = append(initHooks[event], initDef)
		}
	}

	return &ControlRequest{
		Type:      MessageTypeControlRequest,
		RequestID: generateRequestID(),
		Request: &ControlRequestBody{
			Subtype:      ControlSubtypeInitialize,
			InitHookDefs: initHooks,
		},
	}
}

// NewSetPermissionModeRequest creates a set permission mode request.
func NewSetPermissionModeRequest(mode PermissionMode) *ControlRequest {
	return &ControlRequest{
		Type:      MessageTypeControlRequest,
		RequestID: generateRequestID(),
		Request: &ControlRequestBody{
			Subtype: ControlSubtypeSetPermissionMode,
			Mode:    string(mode),
		},
	}
}

// NewControlResponseSuccess creates a success response.
func NewControlResponseSuccess(requestID string, response any) *ControlResponse {
	return &ControlResponse{
		Type: "control_response",
		Response: &ControlResponsePayload{
			Subtype:   "success",
			RequestID: requestID,
			Response:  response,
		},
	}
}

// NewControlResponseError creates an error response.
func NewControlResponseError(requestID string, errMsg string) *ControlResponse {
	return &ControlResponse{
		Type: "control_response",
		Response: &ControlResponsePayload{
			Subtype:   "error",
			RequestID: requestID,
			Error:     errMsg,
		},
	}
}
