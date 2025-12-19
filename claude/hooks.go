package claude

import "context"

// HookDecision represents the permission decision from a hook.
type HookDecision string

const (
	// HookDecisionNone indicates no explicit decision (pass through).
	HookDecisionNone HookDecision = ""

	// HookDecisionAllow explicitly allows the tool use.
	HookDecisionAllow HookDecision = "allow"

	// HookDecisionDeny explicitly denies the tool use.
	HookDecisionDeny HookDecision = "deny"
)

// HookContext provides context information to hook callbacks.
type HookContext struct {
	// SessionID is the current session identifier.
	SessionID string

	// TranscriptPath is the path to the conversation transcript.
	TranscriptPath string

	// WorkingDir is the current working directory.
	WorkingDir string

	// PermissionMode is the current permission mode.
	PermissionMode string
}

// HookOutput is the response from a hook callback.
type HookOutput struct {
	// Decision is the permission decision (allow/deny/none).
	Decision HookDecision

	// Reason explains the decision (shown to the model).
	Reason string

	// SystemMessage is a message to inject into the conversation.
	SystemMessage string

	// AdditionalContext adds context to the conversation.
	AdditionalContext string

	// UpdatedInput contains modified tool input (when allowing).
	UpdatedInput map[string]any

	// Continue controls whether to continue execution (default true).
	Continue *bool

	// StopReason explains why execution was stopped.
	StopReason string
}

// PreToolUseInput contains information about a tool use before execution.
type PreToolUseInput struct {
	// ToolName is the name of the tool being invoked.
	ToolName string

	// ToolInput is the input parameters to the tool.
	ToolInput map[string]any

	// ToolUseID is the unique identifier for this tool use.
	ToolUseID string
}

// PostToolUseInput contains information about a tool use after execution.
type PostToolUseInput struct {
	// ToolName is the name of the tool that was invoked.
	ToolName string

	// ToolInput is the input parameters that were passed to the tool.
	ToolInput map[string]any

	// ToolUseID is the unique identifier for this tool use.
	ToolUseID string

	// ToolResponse is the output from the tool.
	ToolResponse any

	// IsError indicates if the tool returned an error.
	IsError bool
}

// UserPromptSubmitInput contains information about a user prompt.
type UserPromptSubmitInput struct {
	// Prompt is the user's message content.
	Prompt string

	// SessionID is the session identifier.
	SessionID string
}

// StopInput contains information when the agent stops.
type StopInput struct {
	// Reason is why the agent stopped.
	Reason string

	// SessionID is the session identifier.
	SessionID string
}

// SubagentStopInput contains information when a subagent stops.
type SubagentStopInput struct {
	// SubagentID is the identifier of the subagent.
	SubagentID string

	// Reason is why the subagent stopped.
	Reason string

	// SessionID is the session identifier.
	SessionID string
}

// PreCompactInput contains information before conversation compaction.
type PreCompactInput struct {
	// SessionID is the session identifier.
	SessionID string

	// MessageCount is the number of messages in the conversation.
	MessageCount int
}

// Hook function types for each event.

// PreToolUseHook is called before a tool is executed.
// Return HookDecisionDeny to block the tool use.
// Return HookDecisionAllow to explicitly allow.
// Return HookDecisionNone to let the system decide.
type PreToolUseHook func(ctx context.Context, input *PreToolUseInput, hookCtx *HookContext) (*HookOutput, error)

// PostToolUseHook is called after a tool is executed.
type PostToolUseHook func(ctx context.Context, input *PostToolUseInput, hookCtx *HookContext) (*HookOutput, error)

// UserPromptSubmitHook is called when a user submits a prompt.
type UserPromptSubmitHook func(ctx context.Context, input *UserPromptSubmitInput, hookCtx *HookContext) (*HookOutput, error)

// StopHook is called when the agent stops.
type StopHook func(ctx context.Context, input *StopInput, hookCtx *HookContext) (*HookOutput, error)

// SubagentStopHook is called when a subagent stops.
type SubagentStopHook func(ctx context.Context, input *SubagentStopInput, hookCtx *HookContext) (*HookOutput, error)

// PreCompactHook is called before conversation compaction.
type PreCompactHook func(ctx context.Context, input *PreCompactInput, hookCtx *HookContext) (*HookOutput, error)

// WithPreToolUseHook registers a hook to be called before tool execution.
// The matcher specifies which tools to match (e.g., "Bash", "Read|Write").
// Use empty string to match all tools.
func WithPreToolUseHook(matcher string, hook PreToolUseHook, opts ...HookOption) Option {
	return func(c *config) {
		c.initHookMaps()

		hc := &hookConfig{}
		for _, opt := range opts {
			opt(hc)
		}

		callbackID := c.generateCallbackID()
		c.hookCallbacks[callbackID] = hook

		c.hooks[PreToolUse] = append(c.hooks[PreToolUse], hookMatcher{
			matcher:     matcher,
			callbackIDs: []string{callbackID},
			timeout:     hc.timeout,
		})
	}
}

// WithPostToolUseHook registers a hook to be called after tool execution.
// The matcher specifies which tools to match (e.g., "Bash", "Read|Write").
// Use empty string to match all tools.
func WithPostToolUseHook(matcher string, hook PostToolUseHook, opts ...HookOption) Option {
	return func(c *config) {
		c.initHookMaps()

		hc := &hookConfig{}
		for _, opt := range opts {
			opt(hc)
		}

		callbackID := c.generateCallbackID()
		c.hookCallbacks[callbackID] = hook

		c.hooks[PostToolUse] = append(c.hooks[PostToolUse], hookMatcher{
			matcher:     matcher,
			callbackIDs: []string{callbackID},
			timeout:     hc.timeout,
		})
	}
}

// WithUserPromptSubmitHook registers a hook to be called when a user submits a prompt.
func WithUserPromptSubmitHook(hook UserPromptSubmitHook, opts ...HookOption) Option {
	return func(c *config) {
		c.initHookMaps()

		hc := &hookConfig{}
		for _, opt := range opts {
			opt(hc)
		}

		callbackID := c.generateCallbackID()
		c.hookCallbacks[callbackID] = hook

		c.hooks[UserPromptSubmit] = append(c.hooks[UserPromptSubmit], hookMatcher{
			matcher:     "",
			callbackIDs: []string{callbackID},
			timeout:     hc.timeout,
		})
	}
}

// WithStopHook registers a hook to be called when the agent stops.
func WithStopHook(hook StopHook, opts ...HookOption) Option {
	return func(c *config) {
		c.initHookMaps()

		hc := &hookConfig{}
		for _, opt := range opts {
			opt(hc)
		}

		callbackID := c.generateCallbackID()
		c.hookCallbacks[callbackID] = hook

		c.hooks[Stop] = append(c.hooks[Stop], hookMatcher{
			matcher:     "",
			callbackIDs: []string{callbackID},
			timeout:     hc.timeout,
		})
	}
}

// WithSubagentStopHook registers a hook to be called when a subagent stops.
func WithSubagentStopHook(hook SubagentStopHook, opts ...HookOption) Option {
	return func(c *config) {
		c.initHookMaps()

		hc := &hookConfig{}
		for _, opt := range opts {
			opt(hc)
		}

		callbackID := c.generateCallbackID()
		c.hookCallbacks[callbackID] = hook

		c.hooks[SubagentStop] = append(c.hooks[SubagentStop], hookMatcher{
			matcher:     "",
			callbackIDs: []string{callbackID},
			timeout:     hc.timeout,
		})
	}
}

// WithPreCompactHook registers a hook to be called before conversation compaction.
func WithPreCompactHook(hook PreCompactHook, opts ...HookOption) Option {
	return func(c *config) {
		c.initHookMaps()

		hc := &hookConfig{}
		for _, opt := range opts {
			opt(hc)
		}

		callbackID := c.generateCallbackID()
		c.hookCallbacks[callbackID] = hook

		c.hooks[PreCompact] = append(c.hooks[PreCompact], hookMatcher{
			matcher:     "",
			callbackIDs: []string{callbackID},
			timeout:     hc.timeout,
		})
	}
}
