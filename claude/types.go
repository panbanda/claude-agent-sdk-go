package claude

// OutputFormat configures structured output with JSON schema validation.
type OutputFormat struct {
	// Type is the output format type. Currently only "json_schema" is supported.
	Type string `json:"type"`

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
