package claude

import (
	"testing"
	"time"
)

func TestPermissionMode(t *testing.T) {
	t.Run("permission mode constants exist", func(t *testing.T) {
		modes := []PermissionMode{
			PermissionDefault,
			PermissionAcceptEdits,
			PermissionPlan,
			PermissionBypass,
		}

		// All modes should be non-empty strings
		for _, m := range modes {
			if m == "" {
				t.Error("PermissionMode should not be empty")
			}
		}

		// All modes should be distinct
		seen := make(map[PermissionMode]bool)
		for _, m := range modes {
			if seen[m] {
				t.Errorf("duplicate mode: %s", m)
			}
			seen[m] = true
		}
	})
}

func TestOptions(t *testing.T) {
	t.Run("default config has zero values", func(t *testing.T) {
		cfg := &config{}
		applyOptions(cfg)

		if cfg.model != "" {
			t.Errorf("model = %q, want empty", cfg.model)
		}
		if cfg.maxTurns != 0 {
			t.Errorf("maxTurns = %d, want 0", cfg.maxTurns)
		}
		if cfg.permissionMode != "" {
			t.Errorf("permissionMode = %q, want empty", cfg.permissionMode)
		}
	})

	t.Run("WithModel sets model", func(t *testing.T) {
		cfg := &config{}
		applyOptions(cfg, WithModel("claude-sonnet-4-5"))

		if cfg.model != "claude-sonnet-4-5" {
			t.Errorf("model = %q, want %q", cfg.model, "claude-sonnet-4-5")
		}
	})

	t.Run("WithFallbackModel sets fallback model", func(t *testing.T) {
		cfg := &config{}
		applyOptions(cfg, WithFallbackModel("claude-haiku-3-5"))

		if cfg.fallbackModel != "claude-haiku-3-5" {
			t.Errorf("fallbackModel = %q, want %q", cfg.fallbackModel, "claude-haiku-3-5")
		}
	})

	t.Run("WithMaxTurns sets max turns", func(t *testing.T) {
		cfg := &config{}
		applyOptions(cfg, WithMaxTurns(5))

		if cfg.maxTurns != 5 {
			t.Errorf("maxTurns = %d, want 5", cfg.maxTurns)
		}
	})

	t.Run("WithMaxBudgetUSD sets budget", func(t *testing.T) {
		cfg := &config{}
		applyOptions(cfg, WithMaxBudgetUSD(1.50))

		if cfg.maxBudgetUSD != 1.50 {
			t.Errorf("maxBudgetUSD = %f, want 1.50", cfg.maxBudgetUSD)
		}
	})

	t.Run("WithPermissionMode sets mode", func(t *testing.T) {
		cfg := &config{}
		applyOptions(cfg, WithPermissionMode(PermissionBypass))

		if cfg.permissionMode != PermissionBypass {
			t.Errorf("permissionMode = %q, want %q", cfg.permissionMode, PermissionBypass)
		}
	})

	t.Run("WithSystemPrompt sets prompt", func(t *testing.T) {
		cfg := &config{}
		applyOptions(cfg, WithSystemPrompt("You are a helpful assistant."))

		if cfg.systemPrompt != "You are a helpful assistant." {
			t.Errorf("systemPrompt = %q, want %q", cfg.systemPrompt, "You are a helpful assistant.")
		}
	})

	t.Run("WithAllowedTools sets allowed tools", func(t *testing.T) {
		cfg := &config{}
		applyOptions(cfg, WithAllowedTools("Read", "Write", "Edit"))

		if len(cfg.allowedTools) != 3 {
			t.Fatalf("allowedTools length = %d, want 3", len(cfg.allowedTools))
		}
		if cfg.allowedTools[0] != "Read" || cfg.allowedTools[1] != "Write" || cfg.allowedTools[2] != "Edit" {
			t.Errorf("allowedTools = %v, want [Read Write Edit]", cfg.allowedTools)
		}
	})

	t.Run("WithDisallowedTools sets disallowed tools", func(t *testing.T) {
		cfg := &config{}
		applyOptions(cfg, WithDisallowedTools("Bash"))

		if len(cfg.disallowedTools) != 1 {
			t.Fatalf("disallowedTools length = %d, want 1", len(cfg.disallowedTools))
		}
		if cfg.disallowedTools[0] != "Bash" {
			t.Errorf("disallowedTools[0] = %q, want %q", cfg.disallowedTools[0], "Bash")
		}
	})

	t.Run("WithWorkingDir sets working directory", func(t *testing.T) {
		cfg := &config{}
		applyOptions(cfg, WithWorkingDir("/tmp/project"))

		if cfg.workingDir != "/tmp/project" {
			t.Errorf("workingDir = %q, want %q", cfg.workingDir, "/tmp/project")
		}
	})

	t.Run("WithCLIPath sets CLI path", func(t *testing.T) {
		cfg := &config{}
		applyOptions(cfg, WithCLIPath("/usr/local/bin/claude"))

		if cfg.cliPath != "/usr/local/bin/claude" {
			t.Errorf("cliPath = %q, want %q", cfg.cliPath, "/usr/local/bin/claude")
		}
	})

	t.Run("WithEnv sets environment variables", func(t *testing.T) {
		cfg := &config{}
		applyOptions(cfg, WithEnv(map[string]string{"FOO": "bar", "BAZ": "qux"}))

		if len(cfg.env) != 2 {
			t.Fatalf("env length = %d, want 2", len(cfg.env))
		}
		if cfg.env["FOO"] != "bar" {
			t.Errorf("env[FOO] = %q, want %q", cfg.env["FOO"], "bar")
		}
	})

	t.Run("WithContinueConversation enables continuation", func(t *testing.T) {
		cfg := &config{}
		applyOptions(cfg, WithContinueConversation(true))

		if !cfg.continueConversation {
			t.Error("continueConversation should be true")
		}
	})

	t.Run("WithResume sets session ID", func(t *testing.T) {
		cfg := &config{}
		applyOptions(cfg, WithResume("session-123"))

		if cfg.resume != "session-123" {
			t.Errorf("resume = %q, want %q", cfg.resume, "session-123")
		}
	})

	t.Run("WithMaxThinkingTokens sets thinking tokens", func(t *testing.T) {
		cfg := &config{}
		applyOptions(cfg, WithMaxThinkingTokens(10000))

		if cfg.maxThinkingTokens != 10000 {
			t.Errorf("maxThinkingTokens = %d, want 10000", cfg.maxThinkingTokens)
		}
	})

	t.Run("WithMCPConfig sets MCP config path", func(t *testing.T) {
		cfg := &config{}
		applyOptions(cfg, WithMCPConfig("/path/to/mcp-config.json"))

		if cfg.mcpConfig != "/path/to/mcp-config.json" {
			t.Errorf("mcpConfig = %q, want %q", cfg.mcpConfig, "/path/to/mcp-config.json")
		}
	})

	t.Run("multiple options compose", func(t *testing.T) {
		cfg := &config{}
		applyOptions(cfg,
			WithModel("claude-sonnet-4-5"),
			WithMaxTurns(10),
			WithPermissionMode(PermissionAcceptEdits),
			WithAllowedTools("Read", "Write"),
		)

		if cfg.model != "claude-sonnet-4-5" {
			t.Errorf("model = %q, want %q", cfg.model, "claude-sonnet-4-5")
		}
		if cfg.maxTurns != 10 {
			t.Errorf("maxTurns = %d, want 10", cfg.maxTurns)
		}
		if cfg.permissionMode != PermissionAcceptEdits {
			t.Errorf("permissionMode = %q, want %q", cfg.permissionMode, PermissionAcceptEdits)
		}
		if len(cfg.allowedTools) != 2 {
			t.Errorf("allowedTools length = %d, want 2", len(cfg.allowedTools))
		}
	})

	t.Run("later options override earlier", func(t *testing.T) {
		cfg := &config{}
		applyOptions(cfg,
			WithModel("first"),
			WithModel("second"),
		)

		if cfg.model != "second" {
			t.Errorf("model = %q, want %q", cfg.model, "second")
		}
	})

	t.Run("WithExtraArgs sets extra args map", func(t *testing.T) {
		cfg := &config{}
		applyOptions(cfg, WithExtraArgs(map[string]string{
			"verbose":       "",
			"output-format": "json",
		}))

		if len(cfg.extraArgs) != 2 {
			t.Fatalf("extraArgs length = %d, want 2", len(cfg.extraArgs))
		}
		if cfg.extraArgs["verbose"] != "" {
			t.Errorf("extraArgs[verbose] = %q, want empty string", cfg.extraArgs["verbose"])
		}
		if cfg.extraArgs["output-format"] != "json" {
			t.Errorf("extraArgs[output-format] = %q, want %q", cfg.extraArgs["output-format"], "json")
		}
	})

	t.Run("WithAddDirs sets directories", func(t *testing.T) {
		cfg := &config{}
		applyOptions(cfg, WithAddDirs("/path/to/dir1", "/path/to/dir2"))

		if len(cfg.addDirs) != 2 {
			t.Fatalf("addDirs length = %d, want 2", len(cfg.addDirs))
		}
		if cfg.addDirs[0] != "/path/to/dir1" {
			t.Errorf("addDirs[0] = %q, want %q", cfg.addDirs[0], "/path/to/dir1")
		}
		if cfg.addDirs[1] != "/path/to/dir2" {
			t.Errorf("addDirs[1] = %q, want %q", cfg.addDirs[1], "/path/to/dir2")
		}
	})

	t.Run("WithSettings sets settings path", func(t *testing.T) {
		cfg := &config{}
		applyOptions(cfg, WithSettings("/path/to/settings.json"))

		if cfg.settings != "/path/to/settings.json" {
			t.Errorf("settings = %q, want %q", cfg.settings, "/path/to/settings.json")
		}
	})

	t.Run("WithUser sets user identifier", func(t *testing.T) {
		cfg := &config{}
		applyOptions(cfg, WithUser("my-app-user-123"))

		if cfg.user != "my-app-user-123" {
			t.Errorf("user = %q, want %q", cfg.user, "my-app-user-123")
		}
	})

	t.Run("WithBetas sets beta features", func(t *testing.T) {
		cfg := &config{}
		applyOptions(cfg, WithBetas("context-1m-2025-08-07", "another-beta"))

		if len(cfg.betas) != 2 {
			t.Fatalf("betas length = %d, want 2", len(cfg.betas))
		}
		if cfg.betas[0] != "context-1m-2025-08-07" {
			t.Errorf("betas[0] = %q, want %q", cfg.betas[0], "context-1m-2025-08-07")
		}
		if cfg.betas[1] != "another-beta" {
			t.Errorf("betas[1] = %q, want %q", cfg.betas[1], "another-beta")
		}
	})

	t.Run("WithMaxBufferSize sets buffer size", func(t *testing.T) {
		cfg := &config{}
		applyOptions(cfg, WithMaxBufferSize(2048000))

		if cfg.maxBufferSize != 2048000 {
			t.Errorf("maxBufferSize = %d, want 2048000", cfg.maxBufferSize)
		}
	})
}

func TestHookOptions(t *testing.T) {
	t.Run("HookTimeout sets timeout", func(t *testing.T) {
		hcfg := &hookConfig{}
		HookTimeout(30 * time.Second)(hcfg)

		if hcfg.timeout != 30*time.Second {
			t.Errorf("timeout = %v, want 30s", hcfg.timeout)
		}
	})
}

func TestWithCanUseTool(t *testing.T) {
	t.Run("sets canUseTool callback", func(t *testing.T) {
		fn := func(toolName string, input map[string]any) (PermissionResult, error) {
			return PermissionResult{Allow: true}, nil
		}

		cfg := &config{}
		applyOptions(cfg, WithCanUseTool(fn))

		if cfg.canUseTool == nil {
			t.Error("canUseTool should not be nil")
		}
	})

	t.Run("callback is invocable", func(t *testing.T) {
		called := false
		fn := func(toolName string, input map[string]any) (PermissionResult, error) {
			called = true
			if toolName != "Bash" {
				t.Errorf("toolName = %q, want 'Bash'", toolName)
			}
			if input["command"] != "ls" {
				t.Errorf("input[command] = %v, want 'ls'", input["command"])
			}
			return PermissionResult{Allow: true, Message: "allowed"}, nil
		}

		cfg := &config{}
		applyOptions(cfg, WithCanUseTool(fn))

		result, err := cfg.canUseTool("Bash", map[string]any{"command": "ls"})
		if err != nil {
			t.Errorf("canUseTool error = %v, want nil", err)
		}
		if !called {
			t.Error("callback was not called")
		}
		if !result.Allow {
			t.Error("result.Allow should be true")
		}
		if result.Message != "allowed" {
			t.Errorf("result.Message = %q, want 'allowed'", result.Message)
		}
	})

	t.Run("callback can deny with updated input", func(t *testing.T) {
		fn := func(toolName string, input map[string]any) (PermissionResult, error) {
			return PermissionResult{
				Allow:        false,
				Message:      "denied",
				UpdatedInput: map[string]any{"command": "echo denied"},
			}, nil
		}

		cfg := &config{}
		applyOptions(cfg, WithCanUseTool(fn))

		result, _ := cfg.canUseTool("Bash", map[string]any{"command": "rm -rf /"})
		if result.Allow {
			t.Error("result.Allow should be false")
		}
		if result.UpdatedInput["command"] != "echo denied" {
			t.Errorf("UpdatedInput[command] = %v, want 'echo denied'", result.UpdatedInput["command"])
		}
	})
}

func TestWithAgents(t *testing.T) {
	t.Run("sets agents", func(t *testing.T) {
		agents := map[string]AgentDefinition{
			"code-reviewer": {
				Description: "Reviews code",
				Prompt:      "You are a code reviewer",
				Tools:       []string{"Read", "Grep"},
				Model:       "sonnet",
			},
			"doc-writer": {
				Description: "Writes docs",
				Prompt:      "You are a technical writer",
				Tools:       []string{"Read", "Write", "Edit"},
			},
		}

		cfg := &config{}
		applyOptions(cfg, WithAgents(agents))

		if len(cfg.agents) != 2 {
			t.Fatalf("agents length = %d, want 2", len(cfg.agents))
		}

		reviewer := cfg.agents["code-reviewer"]
		if reviewer.Description != "Reviews code" {
			t.Errorf("code-reviewer.Description = %q, want %q", reviewer.Description, "Reviews code")
		}
		if reviewer.Prompt != "You are a code reviewer" {
			t.Errorf("code-reviewer.Prompt = %q, want %q", reviewer.Prompt, "You are a code reviewer")
		}
		if len(reviewer.Tools) != 2 {
			t.Fatalf("code-reviewer.Tools length = %d, want 2", len(reviewer.Tools))
		}
		if reviewer.Model != "sonnet" {
			t.Errorf("code-reviewer.Model = %q, want %q", reviewer.Model, "sonnet")
		}

		writer := cfg.agents["doc-writer"]
		if writer.Description != "Writes docs" {
			t.Errorf("doc-writer.Description = %q, want %q", writer.Description, "Writes docs")
		}
	})

	t.Run("empty agents map", func(t *testing.T) {
		cfg := &config{}
		applyOptions(cfg, WithAgents(map[string]AgentDefinition{}))

		if len(cfg.agents) != 0 {
			t.Errorf("agents length = %d, want 0", len(cfg.agents))
		}
	})
}

func TestWithSettingSources(t *testing.T) {
	t.Run("sets single source", func(t *testing.T) {
		cfg := &config{}
		applyOptions(cfg, WithSettingSources(SettingSourceUser))

		if len(cfg.settingSources) != 1 {
			t.Fatalf("settingSources length = %d, want 1", len(cfg.settingSources))
		}
		if cfg.settingSources[0] != SettingSourceUser {
			t.Errorf("settingSources[0] = %q, want %q", cfg.settingSources[0], SettingSourceUser)
		}
	})

	t.Run("sets multiple sources", func(t *testing.T) {
		cfg := &config{}
		applyOptions(cfg, WithSettingSources(SettingSourceUser, SettingSourceProject, SettingSourceLocal))

		if len(cfg.settingSources) != 3 {
			t.Fatalf("settingSources length = %d, want 3", len(cfg.settingSources))
		}
		if cfg.settingSources[0] != SettingSourceUser {
			t.Errorf("settingSources[0] = %q, want %q", cfg.settingSources[0], SettingSourceUser)
		}
		if cfg.settingSources[1] != SettingSourceProject {
			t.Errorf("settingSources[1] = %q, want %q", cfg.settingSources[1], SettingSourceProject)
		}
		if cfg.settingSources[2] != SettingSourceLocal {
			t.Errorf("settingSources[2] = %q, want %q", cfg.settingSources[2], SettingSourceLocal)
		}
	})

	t.Run("empty sources", func(t *testing.T) {
		cfg := &config{}
		applyOptions(cfg, WithSettingSources())

		if len(cfg.settingSources) != 0 {
			t.Errorf("settingSources length = %d, want 0", len(cfg.settingSources))
		}
	})
}

func TestWithPlugins(t *testing.T) {
	t.Run("sets single plugin", func(t *testing.T) {
		cfg := &config{}
		applyOptions(cfg, WithPlugins(PluginConfig{Type: "local", Path: "/path/to/plugin"}))

		if len(cfg.plugins) != 1 {
			t.Fatalf("plugins length = %d, want 1", len(cfg.plugins))
		}
		if cfg.plugins[0].Type != "local" {
			t.Errorf("plugins[0].Type = %q, want %q", cfg.plugins[0].Type, "local")
		}
		if cfg.plugins[0].Path != "/path/to/plugin" {
			t.Errorf("plugins[0].Path = %q, want %q", cfg.plugins[0].Path, "/path/to/plugin")
		}
	})

	t.Run("sets multiple plugins", func(t *testing.T) {
		cfg := &config{}
		applyOptions(cfg, WithPlugins(
			PluginConfig{Type: "local", Path: "/path/to/plugin1"},
			PluginConfig{Type: "local", Path: "/path/to/plugin2"},
		))

		if len(cfg.plugins) != 2 {
			t.Fatalf("plugins length = %d, want 2", len(cfg.plugins))
		}
		if cfg.plugins[0].Path != "/path/to/plugin1" {
			t.Errorf("plugins[0].Path = %q, want %q", cfg.plugins[0].Path, "/path/to/plugin1")
		}
		if cfg.plugins[1].Path != "/path/to/plugin2" {
			t.Errorf("plugins[1].Path = %q, want %q", cfg.plugins[1].Path, "/path/to/plugin2")
		}
	})

	t.Run("empty plugins", func(t *testing.T) {
		cfg := &config{}
		applyOptions(cfg, WithPlugins())

		if len(cfg.plugins) != 0 {
			t.Errorf("plugins length = %d, want 0", len(cfg.plugins))
		}
	})
}

// Helper to apply options
func applyOptions(cfg *config, opts ...Option) {
	for _, opt := range opts {
		opt(cfg)
	}
}

func TestWithOutputFormat(t *testing.T) {
	t.Run("sets output format", func(t *testing.T) {
		format := &OutputFormat{
			Type: OutputFormatTypeJSONSchema,
			Schema: map[string]any{
				"type": "object",
			},
		}

		cfg := &config{}
		applyOptions(cfg, WithOutputFormat(format))

		if cfg.outputFormat == nil {
			t.Fatal("outputFormat should not be nil")
		}

		if cfg.outputFormat.Type != OutputFormatTypeJSONSchema {
			t.Errorf("outputFormat.Type = %q, want %q", cfg.outputFormat.Type, OutputFormatTypeJSONSchema)
		}

		if cfg.outputFormat.Schema == nil {
			t.Error("outputFormat.Schema should not be nil")
		}
	})
}

func TestWithJSONSchema(t *testing.T) {
	t.Run("creates output format with schema", func(t *testing.T) {
		schema := map[string]any{
			"type": "object",
			"properties": map[string]any{
				"answer": map[string]any{"type": "string"},
			},
		}

		cfg := &config{}
		applyOptions(cfg, WithJSONSchema(schema))

		if cfg.outputFormat == nil {
			t.Fatal("outputFormat should not be nil")
		}

		if cfg.outputFormat.Type != OutputFormatTypeJSONSchema {
			t.Errorf("outputFormat.Type = %q, want %q", cfg.outputFormat.Type, OutputFormatTypeJSONSchema)
		}

		if cfg.outputFormat.Schema == nil {
			t.Fatal("outputFormat.Schema should not be nil")
		}

		props, ok := cfg.outputFormat.Schema["properties"].(map[string]any)
		if !ok {
			t.Fatal("schema properties should be map[string]any")
		}

		if _, exists := props["answer"]; !exists {
			t.Error("schema should have 'answer' property")
		}
	})
}

func TestWithSandbox(t *testing.T) {
	t.Run("sets sandbox settings", func(t *testing.T) {
		settings := &SandboxSettings{
			Enabled:                  true,
			AutoAllowBashIfSandboxed: true,
			ExcludedCommands:         []string{"git", "docker"},
		}

		cfg := &config{}
		applyOptions(cfg, WithSandbox(settings))

		if cfg.sandbox == nil {
			t.Fatal("sandbox should not be nil")
		}

		if !cfg.sandbox.Enabled {
			t.Error("sandbox.Enabled should be true")
		}

		if !cfg.sandbox.AutoAllowBashIfSandboxed {
			t.Error("sandbox.AutoAllowBashIfSandboxed should be true")
		}

		if len(cfg.sandbox.ExcludedCommands) != 2 {
			t.Errorf("sandbox.ExcludedCommands length = %d, want 2", len(cfg.sandbox.ExcludedCommands))
		}
	})

	t.Run("sets sandbox with network config", func(t *testing.T) {
		settings := &SandboxSettings{
			Enabled: true,
			Network: &SandboxNetworkConfig{
				AllowUnixSockets:  []string{"/var/run/docker.sock"},
				AllowLocalBinding: true,
			},
		}

		cfg := &config{}
		applyOptions(cfg, WithSandbox(settings))

		if cfg.sandbox == nil {
			t.Fatal("sandbox should not be nil")
		}

		if cfg.sandbox.Network == nil {
			t.Fatal("sandbox.Network should not be nil")
		}

		if len(cfg.sandbox.Network.AllowUnixSockets) != 1 {
			t.Errorf("AllowUnixSockets length = %d, want 1", len(cfg.sandbox.Network.AllowUnixSockets))
		}

		if !cfg.sandbox.Network.AllowLocalBinding {
			t.Error("AllowLocalBinding should be true")
		}
	})
}

func TestWithIncludePartialMessages(t *testing.T) {
	t.Run("enables partial messages", func(t *testing.T) {
		cfg := &config{}
		applyOptions(cfg, WithIncludePartialMessages(true))

		if !cfg.includePartialMessages {
			t.Error("includePartialMessages should be true")
		}
	})

	t.Run("disables partial messages", func(t *testing.T) {
		cfg := &config{}
		applyOptions(cfg, WithIncludePartialMessages(false))

		if cfg.includePartialMessages {
			t.Error("includePartialMessages should be false")
		}
	})
}

func TestWithForkSession(t *testing.T) {
	t.Run("enables fork session", func(t *testing.T) {
		cfg := &config{}
		applyOptions(cfg, WithForkSession(true))

		if !cfg.forkSession {
			t.Error("forkSession should be true")
		}
	})

	t.Run("disables fork session", func(t *testing.T) {
		cfg := &config{}
		applyOptions(cfg, WithForkSession(false))

		if cfg.forkSession {
			t.Error("forkSession should be false")
		}
	})
}
