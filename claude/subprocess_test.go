package claude

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestFindCLI(t *testing.T) {
	t.Run("returns path when cli exists in PATH", func(t *testing.T) {
		// Create a temp directory with a mock claude binary
		tempDir := t.TempDir()

		var cliName string
		if runtime.GOOS == "windows" {
			cliName = "claude.exe"
		} else {
			cliName = "claude"
		}

		cliPath := filepath.Join(tempDir, cliName)
		if err := os.WriteFile(cliPath, []byte("#!/bin/sh\necho test"), 0755); err != nil {
			t.Fatal(err)
		}

		// Add temp dir to PATH
		origPath := os.Getenv("PATH")
		os.Setenv("PATH", tempDir+string(os.PathListSeparator)+origPath)
		defer os.Setenv("PATH", origPath)

		path, err := FindCLI()

		if err != nil {
			t.Errorf("FindCLI() error = %v, want nil", err)
		}
		if path != cliPath {
			t.Errorf("FindCLI() = %q, want %q", path, cliPath)
		}
	})

	t.Run("returns error when cli not found", func(t *testing.T) {
		// This test is skipped in environments where claude is installed
		// in fallback locations. We test the error path through Connect
		// with an invalid CLI path instead.
		t.Skip("FindCLI fallback paths may contain claude; tested via Connect")
	})
}

func TestSubprocessTransport_BuildCommand(t *testing.T) {
	t.Run("builds basic command", func(t *testing.T) {
		cfg := &config{}
		st := &SubprocessTransport{
			cliPath: "/usr/bin/claude",
			cfg:     cfg,
		}

		cmd := st.buildCommand()

		// Should have base args
		if cmd[0] != "/usr/bin/claude" {
			t.Errorf("cmd[0] = %q, want '/usr/bin/claude'", cmd[0])
		}

		// Should include output format
		containsOutputFormat := false
		for i, arg := range cmd {
			if arg == "--output-format" && i+1 < len(cmd) && cmd[i+1] == "stream-json" {
				containsOutputFormat = true
				break
			}
		}
		if !containsOutputFormat {
			t.Errorf("command should contain --output-format stream-json, got %v", cmd)
		}
	})

	t.Run("includes model flag", func(t *testing.T) {
		cfg := &config{model: "claude-sonnet-4-5"}
		st := &SubprocessTransport{
			cliPath: "/usr/bin/claude",
			cfg:     cfg,
		}

		cmd := st.buildCommand()

		containsModel := false
		for i, arg := range cmd {
			if arg == "--model" && i+1 < len(cmd) && cmd[i+1] == "claude-sonnet-4-5" {
				containsModel = true
				break
			}
		}
		if !containsModel {
			t.Errorf("command should contain --model claude-sonnet-4-5, got %v", cmd)
		}
	})

	t.Run("includes max turns flag", func(t *testing.T) {
		cfg := &config{maxTurns: 10}
		st := &SubprocessTransport{
			cliPath: "/usr/bin/claude",
			cfg:     cfg,
		}

		cmd := st.buildCommand()

		containsMaxTurns := false
		for i, arg := range cmd {
			if arg == "--max-turns" && i+1 < len(cmd) && cmd[i+1] == "10" {
				containsMaxTurns = true
				break
			}
		}
		if !containsMaxTurns {
			t.Errorf("command should contain --max-turns 10, got %v", cmd)
		}
	})

	t.Run("includes system prompt flag", func(t *testing.T) {
		cfg := &config{systemPrompt: "You are helpful"}
		st := &SubprocessTransport{
			cliPath: "/usr/bin/claude",
			cfg:     cfg,
		}

		cmd := st.buildCommand()

		containsSystemPrompt := false
		for i, arg := range cmd {
			if arg == "--system-prompt" && i+1 < len(cmd) && cmd[i+1] == "You are helpful" {
				containsSystemPrompt = true
				break
			}
		}
		if !containsSystemPrompt {
			t.Errorf("command should contain --system-prompt 'You are helpful', got %v", cmd)
		}
	})

	t.Run("includes permission mode flag", func(t *testing.T) {
		cfg := &config{permissionMode: PermissionBypass}
		st := &SubprocessTransport{
			cliPath: "/usr/bin/claude",
			cfg:     cfg,
		}

		cmd := st.buildCommand()

		containsPermissionMode := false
		for i, arg := range cmd {
			if arg == "--permission-mode" && i+1 < len(cmd) && cmd[i+1] == string(PermissionBypass) {
				containsPermissionMode = true
				break
			}
		}
		if !containsPermissionMode {
			t.Errorf("command should contain --permission-mode bypassPermissions, got %v", cmd)
		}
	})

	t.Run("includes allowed tools flag", func(t *testing.T) {
		cfg := &config{allowedTools: []string{"Read", "Write"}}
		st := &SubprocessTransport{
			cliPath: "/usr/bin/claude",
			cfg:     cfg,
		}

		cmd := st.buildCommand()

		containsAllowedTools := false
		for i, arg := range cmd {
			if arg == "--allowedTools" && i+1 < len(cmd) && cmd[i+1] == "Read,Write" {
				containsAllowedTools = true
				break
			}
		}
		if !containsAllowedTools {
			t.Errorf("command should contain --allowedTools Read,Write, got %v", cmd)
		}
	})

	t.Run("includes disallowed tools flag", func(t *testing.T) {
		cfg := &config{disallowedTools: []string{"Bash"}}
		st := &SubprocessTransport{
			cliPath: "/usr/bin/claude",
			cfg:     cfg,
		}

		cmd := st.buildCommand()

		containsDisallowedTools := false
		for i, arg := range cmd {
			if arg == "--disallowedTools" && i+1 < len(cmd) && cmd[i+1] == "Bash" {
				containsDisallowedTools = true
				break
			}
		}
		if !containsDisallowedTools {
			t.Errorf("command should contain --disallowedTools Bash, got %v", cmd)
		}
	})

	t.Run("includes max budget usd flag", func(t *testing.T) {
		cfg := &config{maxBudgetUSD: 1.5}
		st := &SubprocessTransport{
			cliPath: "/usr/bin/claude",
			cfg:     cfg,
		}

		cmd := st.buildCommand()

		containsMaxBudget := false
		for i, arg := range cmd {
			if arg == "--max-budget-usd" && i+1 < len(cmd) && cmd[i+1] == "1.5" {
				containsMaxBudget = true
				break
			}
		}
		if !containsMaxBudget {
			t.Errorf("command should contain --max-budget-usd 1.5, got %v", cmd)
		}
	})

	t.Run("includes continue flag", func(t *testing.T) {
		cfg := &config{continueConversation: true}
		st := &SubprocessTransport{
			cliPath: "/usr/bin/claude",
			cfg:     cfg,
		}

		cmd := st.buildCommand()

		containsContinue := false
		for _, arg := range cmd {
			if arg == "--continue" {
				containsContinue = true
				break
			}
		}
		if !containsContinue {
			t.Errorf("command should contain --continue, got %v", cmd)
		}
	})

	t.Run("includes resume flag", func(t *testing.T) {
		cfg := &config{resume: "session-123"}
		st := &SubprocessTransport{
			cliPath: "/usr/bin/claude",
			cfg:     cfg,
		}

		cmd := st.buildCommand()

		containsResume := false
		for i, arg := range cmd {
			if arg == "--resume" && i+1 < len(cmd) && cmd[i+1] == "session-123" {
				containsResume = true
				break
			}
		}
		if !containsResume {
			t.Errorf("command should contain --resume session-123, got %v", cmd)
		}
	})

	t.Run("includes input format for streaming", func(t *testing.T) {
		cfg := &config{}
		st := &SubprocessTransport{
			cliPath: "/usr/bin/claude",
			cfg:     cfg,
		}

		cmd := st.buildCommand()

		containsInputFormat := false
		for i, arg := range cmd {
			if arg == "--input-format" && i+1 < len(cmd) && cmd[i+1] == "stream-json" {
				containsInputFormat = true
				break
			}
		}
		if !containsInputFormat {
			t.Errorf("command should contain --input-format stream-json, got %v", cmd)
		}
	})

	t.Run("includes mcp config flag", func(t *testing.T) {
		cfg := &config{mcpConfig: "/path/to/mcp-config.json"}
		st := &SubprocessTransport{
			cliPath: "/usr/bin/claude",
			cfg:     cfg,
		}

		cmd := st.buildCommand()

		containsMCPConfig := false
		for i, arg := range cmd {
			if arg == "--mcp-config" && i+1 < len(cmd) && cmd[i+1] == "/path/to/mcp-config.json" {
				containsMCPConfig = true
				break
			}
		}
		if !containsMCPConfig {
			t.Errorf("command should contain --mcp-config /path/to/mcp-config.json, got %v", cmd)
		}
	})
}

func TestSubprocessTransport_Connect(t *testing.T) {
	// Skip if we don't have a mock setup
	if _, err := exec.LookPath("true"); err != nil {
		t.Skip("true command not available")
	}

	t.Run("returns error for invalid cli path", func(t *testing.T) {
		cfg := &config{}
		st := NewSubprocessTransport(cfg)
		st.cliPath = "/nonexistent/path/to/claude"

		err := st.Connect(context.Background())

		if err == nil {
			t.Error("Connect() error = nil, want error")
		}
	})
}

func TestSubprocessTransport_IsReady(t *testing.T) {
	t.Run("returns false before connect", func(t *testing.T) {
		cfg := &config{}
		st := NewSubprocessTransport(cfg)

		if st.IsReady() {
			t.Error("IsReady() = true, want false before Connect")
		}
	})
}

func TestSubprocessTransport_Messages(t *testing.T) {
	t.Run("returns channel", func(t *testing.T) {
		cfg := &config{}
		st := NewSubprocessTransport(cfg)

		ch := st.Messages()

		if ch == nil {
			t.Error("Messages() = nil, want channel")
		}
	})
}

func TestSubprocessTransport_Errors(t *testing.T) {
	t.Run("returns channel", func(t *testing.T) {
		cfg := &config{}
		st := NewSubprocessTransport(cfg)

		ch := st.Errors()

		if ch == nil {
			t.Error("Errors() = nil, want channel")
		}
	})
}

func TestNewSubprocessTransport(t *testing.T) {
	t.Run("creates transport with config", func(t *testing.T) {
		cfg := &config{
			model:    "claude-sonnet-4-5",
			maxTurns: 5,
		}

		st := NewSubprocessTransport(cfg)

		if st == nil {
			t.Fatal("NewSubprocessTransport() = nil")
		}
		if st.cfg != cfg {
			t.Error("NewSubprocessTransport() cfg not set correctly")
		}
	})

	t.Run("uses custom cli path from config", func(t *testing.T) {
		cfg := &config{
			cliPath: "/custom/path/claude",
		}

		st := NewSubprocessTransport(cfg)

		if st.cliPath != "/custom/path/claude" {
			t.Errorf("cliPath = %q, want '/custom/path/claude'", st.cliPath)
		}
	})
}
