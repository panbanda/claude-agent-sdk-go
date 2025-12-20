package claude

import (
	"fmt"
	"time"
)

// PermissionMode controls how tool permissions are handled.
type PermissionMode string

const (
	// PermissionDefault uses the CLI's default permission prompting.
	PermissionDefault PermissionMode = "default"

	// PermissionAcceptEdits auto-accepts file edit operations.
	PermissionAcceptEdits PermissionMode = "acceptEdits"

	// PermissionPlan enables plan mode for reviewing changes.
	PermissionPlan PermissionMode = "plan"

	// PermissionBypass bypasses all permission checks (use with caution).
	PermissionBypass PermissionMode = "bypassPermissions"
)

// config holds the internal configuration built from options.
// This is not exported; users interact via Option functions.
type config struct {
	// Model settings
	model         string
	fallbackModel string

	// Limits
	maxTurns     int
	maxBudgetUSD float64

	// Permissions
	permissionMode PermissionMode

	// Prompts
	systemPrompt string

	// Tools
	allowedTools    []string
	disallowedTools []string

	// Paths
	workingDir string
	cliPath    string

	// Environment
	env map[string]string

	// Session
	continueConversation bool
	resume               string

	// Advanced
	maxThinkingTokens int

	// MCP
	mcpConfig string

	// Transport (for testing)
	transport Transport

	// Hooks
	hooks map[HookEvent][]hookMatcher

	// Hook callbacks indexed by ID for control request handling
	hookCallbacks map[string]any

	// Counter for generating unique callback IDs
	nextCallbackID int

	// Internal callback for tool permissions
	canUseTool CanUseToolFunc

	// Additional CLI options
	extraArgs     map[string]string
	addDirs       []string
	settings      string
	user          string
	betas         []string
	maxBufferSize int
}

// Option is a function that configures the client.
// Use With* functions to create options.
type Option func(*config)

// initHookMaps ensures hook maps are initialized.
func (c *config) initHookMaps() {
	if c.hooks == nil {
		c.hooks = make(map[HookEvent][]hookMatcher)
	}
	if c.hookCallbacks == nil {
		c.hookCallbacks = make(map[string]any)
	}
}

// generateCallbackID creates a unique callback ID.
func (c *config) generateCallbackID() string {
	id := fmt.Sprintf("hook_%d", c.nextCallbackID)
	c.nextCallbackID++
	return id
}

// WithModel sets the model to use (e.g., "claude-sonnet-4-5").
func WithModel(model string) Option {
	return func(c *config) {
		c.model = model
	}
}

// WithFallbackModel sets a fallback model if the primary is unavailable.
func WithFallbackModel(model string) Option {
	return func(c *config) {
		c.fallbackModel = model
	}
}

// WithMaxTurns limits the number of conversation turns.
func WithMaxTurns(n int) Option {
	return func(c *config) {
		c.maxTurns = n
	}
}

// WithMaxBudgetUSD sets a spending limit in USD.
func WithMaxBudgetUSD(amount float64) Option {
	return func(c *config) {
		c.maxBudgetUSD = amount
	}
}

// WithPermissionMode sets how tool permissions are handled.
func WithPermissionMode(mode PermissionMode) Option {
	return func(c *config) {
		c.permissionMode = mode
	}
}

// WithSystemPrompt sets a custom system prompt.
func WithSystemPrompt(prompt string) Option {
	return func(c *config) {
		c.systemPrompt = prompt
	}
}

// WithAllowedTools specifies which tools Claude can use.
func WithAllowedTools(tools ...string) Option {
	return func(c *config) {
		c.allowedTools = tools
	}
}

// WithDisallowedTools specifies which tools Claude cannot use.
func WithDisallowedTools(tools ...string) Option {
	return func(c *config) {
		c.disallowedTools = tools
	}
}

// WithWorkingDir sets the working directory for the CLI.
func WithWorkingDir(dir string) Option {
	return func(c *config) {
		c.workingDir = dir
	}
}

// WithCLIPath sets a custom path to the Claude CLI binary.
func WithCLIPath(path string) Option {
	return func(c *config) {
		c.cliPath = path
	}
}

// WithEnv sets environment variables for the CLI process.
func WithEnv(env map[string]string) Option {
	return func(c *config) {
		c.env = env
	}
}

// WithContinueConversation enables resuming prior conversations.
func WithContinueConversation(enabled bool) Option {
	return func(c *config) {
		c.continueConversation = enabled
	}
}

// WithResume specifies a session ID to resume.
func WithResume(sessionID string) Option {
	return func(c *config) {
		c.resume = sessionID
	}
}

// WithMaxThinkingTokens sets the token budget for extended thinking.
func WithMaxThinkingTokens(tokens int) Option {
	return func(c *config) {
		c.maxThinkingTokens = tokens
	}
}

// WithMCPConfig sets the path to an MCP server configuration file.
// The config file specifies MCP servers that Claude can use as tools.
func WithMCPConfig(path string) Option {
	return func(c *config) {
		c.mcpConfig = path
	}
}

// WithTransport sets a custom transport (primarily for testing).
func WithTransport(t Transport) Option {
	return func(c *config) {
		c.transport = t
	}
}

// hookConfig holds configuration for a single hook registration.
type hookConfig struct {
	timeout time.Duration
}

// HookOption configures a hook registration.
type HookOption func(*hookConfig)

// HookTimeout sets a timeout for hook execution.
func HookTimeout(d time.Duration) HookOption {
	return func(hc *hookConfig) {
		hc.timeout = d
	}
}

// hookMatcher pairs a pattern with callback IDs and timeout.
type hookMatcher struct {
	matcher     string
	callbackIDs []string // IDs referencing hookCallbacks map
	timeout     time.Duration
}

// HookEvent represents the type of hook event.
type HookEvent string

const (
	// PreToolUse fires before a tool is executed.
	PreToolUse HookEvent = "PreToolUse"

	// PostToolUse fires after a tool is executed.
	PostToolUse HookEvent = "PostToolUse"

	// UserPromptSubmit fires when a user prompt is submitted.
	UserPromptSubmit HookEvent = "UserPromptSubmit"

	// Stop fires when the agent stops.
	Stop HookEvent = "Stop"

	// SubagentStop fires when a subagent stops.
	SubagentStop HookEvent = "SubagentStop"

	// PreCompact fires before conversation compaction.
	PreCompact HookEvent = "PreCompact"
)

// CanUseToolFunc is a callback for custom tool permission logic.
type CanUseToolFunc func(toolName string, input map[string]any) (PermissionResult, error)

// PermissionResult represents the result of a permission check.
type PermissionResult struct {
	// Allow indicates whether the tool use is allowed.
	Allow bool

	// Message is an optional message (used when denying).
	Message string

	// UpdatedInput allows modifying the tool input (when allowing).
	UpdatedInput map[string]any
}

// WithCanUseTool sets a callback for custom tool permission logic.
func WithCanUseTool(fn CanUseToolFunc) Option {
	return func(c *config) {
		c.canUseTool = fn
	}
}

// WithExtraArgs passes arbitrary CLI flags.
// Keys are flag names (without --), values are flag values.
// Use empty string for boolean flags.
func WithExtraArgs(args map[string]string) Option {
	return func(c *config) {
		c.extraArgs = args
	}
}

// WithAddDirs adds additional directories for Claude to access.
func WithAddDirs(dirs ...string) Option {
	return func(c *config) {
		c.addDirs = dirs
	}
}

// WithSettings sets the path to a settings file.
func WithSettings(path string) Option {
	return func(c *config) {
		c.settings = path
	}
}

// WithUser sets a user identifier.
func WithUser(user string) Option {
	return func(c *config) {
		c.user = user
	}
}

// WithBetas enables SDK beta features.
func WithBetas(betas ...string) Option {
	return func(c *config) {
		c.betas = betas
	}
}

// WithMaxBufferSize sets the maximum buffer size for stdout.
func WithMaxBufferSize(size int) Option {
	return func(c *config) {
		c.maxBufferSize = size
	}
}
