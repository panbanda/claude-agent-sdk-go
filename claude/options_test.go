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

// Helper to apply options
func applyOptions(cfg *config, opts ...Option) {
	for _, opt := range opts {
		opt(cfg)
	}
}
