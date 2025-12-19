package claude

// OutputFormatType specifies the type of output format.
type OutputFormatType string

const (
	// OutputFormatTypeJSONSchema enables JSON schema validation for structured output.
	OutputFormatTypeJSONSchema OutputFormatType = "json_schema"
)

// OutputFormat configures structured output with JSON schema validation.
type OutputFormat struct {
	// Type is the output format type. Currently only "json_schema" is supported.
	Type OutputFormatType `json:"type"`

	// Schema is the JSON schema for validating output.
	Schema map[string]any `json:"schema"`
}

// SandboxSettings configures bash command sandboxing.
type SandboxSettings struct {
	// Enabled enables bash sandboxing (macOS/Linux only).
	Enabled bool `json:"enabled,omitempty"`

	// AutoAllowBashIfSandboxed auto-approves bash commands when sandboxed.
	AutoAllowBashIfSandboxed bool `json:"autoAllowBashIfSandboxed,omitempty"`

	// ExcludedCommands are commands that run outside the sandbox.
	ExcludedCommands []string `json:"excludedCommands,omitempty"`

	// AllowUnsandboxedCommands allows commands to bypass sandbox.
	AllowUnsandboxedCommands bool `json:"allowUnsandboxedCommands,omitempty"`

	// Network configures network access in sandbox.
	Network *SandboxNetworkConfig `json:"network,omitempty"`

	// IgnoreViolations specifies violations to ignore.
	IgnoreViolations *SandboxIgnoreViolations `json:"ignoreViolations,omitempty"`

	// EnableWeakerNestedSandbox enables weaker sandbox for Docker (Linux only).
	EnableWeakerNestedSandbox bool `json:"enableWeakerNestedSandbox,omitempty"`
}

// SandboxNetworkConfig configures network access in sandbox.
type SandboxNetworkConfig struct {
	// AllowUnixSockets are Unix socket paths accessible in sandbox.
	AllowUnixSockets []string `json:"allowUnixSockets,omitempty"`

	// AllowAllUnixSockets allows all Unix sockets (less secure).
	AllowAllUnixSockets bool `json:"allowAllUnixSockets,omitempty"`

	// AllowLocalBinding allows binding to localhost ports (macOS only).
	AllowLocalBinding bool `json:"allowLocalBinding,omitempty"`

	// HTTPProxyPort is the HTTP proxy port if using own proxy.
	HTTPProxyPort int `json:"httpProxyPort,omitempty"`

	// SOCKSProxyPort is the SOCKS5 proxy port if using own proxy.
	SOCKSProxyPort int `json:"socksProxyPort,omitempty"`
}

// SandboxIgnoreViolations specifies violations to ignore.
type SandboxIgnoreViolations struct {
	// File paths for which violations should be ignored.
	File []string `json:"file,omitempty"`

	// Network hosts for which violations should be ignored.
	Network []string `json:"network,omitempty"`
}

// AgentDefinition defines a custom agent with specific capabilities.
type AgentDefinition struct {
	// Description is a short description of what the agent does.
	Description string

	// Prompt is the system prompt for the agent.
	Prompt string

	// Tools is the list of tools the agent can use.
	// If nil, the agent uses default tools.
	Tools []string

	// Model specifies the model to use: "sonnet", "opus", "haiku", or "inherit".
	// If empty, inherits from parent.
	Model string
}

// SettingSource specifies where to load settings from.
type SettingSource string

const (
	// SettingSourceUser loads global user settings (~/.claude/).
	SettingSourceUser SettingSource = "user"

	// SettingSourceProject loads project-level settings (.claude/ in project).
	SettingSourceProject SettingSource = "project"

	// SettingSourceLocal loads local gitignored settings (.claude-local/).
	SettingSourceLocal SettingSource = "local"
)

// PluginConfig configures a plugin to load.
type PluginConfig struct {
	// Type is the plugin type. Currently only "local" is supported.
	Type string

	// Path is the path to the plugin directory.
	Path string
}
