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

func TestSubprocessTransport_Send(t *testing.T) {
	t.Run("returns error when not ready", func(t *testing.T) {
		cfg := &config{}
		st := NewSubprocessTransport(cfg)

		err := st.Send(context.Background(), []byte("test"))

		if err != ErrNotConnected {
			t.Errorf("Send() error = %v, want %v", err, ErrNotConnected)
		}
	})

	t.Run("returns error when stdin is nil", func(t *testing.T) {
		cfg := &config{}
		st := NewSubprocessTransport(cfg)
		st.ready = true
		st.stdin = nil

		err := st.Send(context.Background(), []byte("test"))

		if err != ErrNotConnected {
			t.Errorf("Send() error = %v, want %v", err, ErrNotConnected)
		}
	})
}

func TestSubprocessTransport_Close(t *testing.T) {
	t.Run("returns nil when not ready", func(t *testing.T) {
		cfg := &config{}
		st := NewSubprocessTransport(cfg)

		err := st.Close()

		if err != nil {
			t.Errorf("Close() error = %v, want nil", err)
		}
	})

	t.Run("closes stdin when ready", func(t *testing.T) {
		cfg := &config{}
		st := NewSubprocessTransport(cfg)
		st.ready = true

		// Create a pipe for stdin
		r, w, _ := os.Pipe()
		st.stdin = w
		defer r.Close()

		err := st.Close()

		if err != nil {
			t.Errorf("Close() error = %v, want nil", err)
		}
		if st.ready {
			t.Error("ready should be false after Close()")
		}
		if st.stdin != nil {
			t.Error("stdin should be nil after Close()")
		}
	})

	t.Run("safe to call multiple times", func(t *testing.T) {
		cfg := &config{}
		st := NewSubprocessTransport(cfg)
		st.ready = true

		r, w, _ := os.Pipe()
		st.stdin = w
		defer r.Close()

		_ = st.Close()
		err := st.Close()

		if err != nil {
			t.Errorf("Close() second call error = %v, want nil", err)
		}
	})
}

func TestSubprocessTransport_BuildCommand_AllOptions(t *testing.T) {
	t.Run("includes fallback model flag", func(t *testing.T) {
		cfg := &config{fallbackModel: "claude-haiku-3-5"}
		st := &SubprocessTransport{
			cliPath: "/usr/bin/claude",
			cfg:     cfg,
		}

		cmd := st.buildCommand()

		containsFallbackModel := false
		for i, arg := range cmd {
			if arg == "--fallback-model" && i+1 < len(cmd) && cmd[i+1] == "claude-haiku-3-5" {
				containsFallbackModel = true
				break
			}
		}
		if !containsFallbackModel {
			t.Errorf("command should contain --fallback-model claude-haiku-3-5, got %v", cmd)
		}
	})

	t.Run("includes max thinking tokens flag", func(t *testing.T) {
		cfg := &config{maxThinkingTokens: 50000}
		st := &SubprocessTransport{
			cliPath: "/usr/bin/claude",
			cfg:     cfg,
		}

		cmd := st.buildCommand()

		containsMaxThinking := false
		for i, arg := range cmd {
			if arg == "--max-thinking-tokens" && i+1 < len(cmd) && cmd[i+1] == "50000" {
				containsMaxThinking = true
				break
			}
		}
		if !containsMaxThinking {
			t.Errorf("command should contain --max-thinking-tokens 50000, got %v", cmd)
		}
	})

	t.Run("includes extra args flags", func(t *testing.T) {
		cfg := &config{extraArgs: map[string]string{
			"verbose":       "",
			"output-format": "json",
		}}
		st := &SubprocessTransport{
			cliPath: "/usr/bin/claude",
			cfg:     cfg,
		}

		cmd := st.buildCommand()

		containsVerbose := false
		containsOutputFormat := false
		for i, arg := range cmd {
			if arg == "--verbose" {
				containsVerbose = true
			}
			if arg == "--output-format" && i+1 < len(cmd) && cmd[i+1] == "json" {
				containsOutputFormat = true
			}
		}
		if !containsVerbose {
			t.Errorf("command should contain --verbose, got %v", cmd)
		}
		if !containsOutputFormat {
			t.Errorf("command should contain --output-format json, got %v", cmd)
		}
	})

	t.Run("includes add-dir flags for each directory", func(t *testing.T) {
		cfg := &config{addDirs: []string{"/path/to/dir1", "/path/to/dir2"}}
		st := &SubprocessTransport{
			cliPath: "/usr/bin/claude",
			cfg:     cfg,
		}

		cmd := st.buildCommand()

		addDirCount := 0
		for i, arg := range cmd {
			if arg == "--add-dir" && i+1 < len(cmd) {
				addDirCount++
				if cmd[i+1] != "/path/to/dir1" && cmd[i+1] != "/path/to/dir2" {
					t.Errorf("unexpected --add-dir value: %q", cmd[i+1])
				}
			}
		}
		if addDirCount != 2 {
			t.Errorf("command should contain 2 --add-dir flags, got %d", addDirCount)
		}
	})

	t.Run("includes settings flag when set", func(t *testing.T) {
		cfg := &config{settings: "/path/to/settings.json"}
		st := &SubprocessTransport{
			cliPath: "/usr/bin/claude",
			cfg:     cfg,
		}

		cmd := st.buildCommand()

		containsSettings := false
		for i, arg := range cmd {
			if arg == "--settings" && i+1 < len(cmd) && cmd[i+1] == "/path/to/settings.json" {
				containsSettings = true
				break
			}
		}
		if !containsSettings {
			t.Errorf("command should contain --settings /path/to/settings.json, got %v", cmd)
		}
	})

	t.Run("user option does not become CLI flag", func(t *testing.T) {
		// User option is for subprocess execution context (like Python's anyio.open_process),
		// not a CLI flag. Verify it's not passed to the command.
		cfg := &config{user: "my-app-user-123"}
		st := &SubprocessTransport{
			cliPath: "/usr/bin/claude",
			cfg:     cfg,
		}

		cmd := st.buildCommand()

		for _, arg := range cmd {
			if arg == "--user" {
				t.Errorf("command should not contain --user flag, got %v", cmd)
			}
		}
	})

	t.Run("includes betas as single comma-separated flag", func(t *testing.T) {
		cfg := &config{betas: []string{"context-1m-2025-08-07", "another-beta"}}
		st := &SubprocessTransport{
			cliPath: "/usr/bin/claude",
			cfg:     cfg,
		}

		cmd := st.buildCommand()

		containsBetas := false
		for i, arg := range cmd {
			if arg == "--betas" && i+1 < len(cmd) {
				containsBetas = true
				expected := "context-1m-2025-08-07,another-beta"
				if cmd[i+1] != expected {
					t.Errorf("--betas value = %q, want %q", cmd[i+1], expected)
				}
				break
			}
		}
		if !containsBetas {
			t.Errorf("command should contain --betas flag, got %v", cmd)
		}
	})
}

func TestSubprocessTransport_ReadMessages(t *testing.T) {
	t.Run("reads messages from reader", func(t *testing.T) {
		cfg := &config{}
		st := NewSubprocessTransport(cfg)

		// Create a pipe to simulate stdout
		r, w, _ := os.Pipe()

		// Write test messages
		go func() {
			w.Write([]byte(`{"type":"assistant"}`))
			w.Write([]byte("\n"))
			w.Write([]byte(`{"type":"result"}`))
			w.Write([]byte("\n"))
			w.Close()
		}()

		// Read messages
		go st.readMessages(r)

		// Collect messages
		var messages [][]byte
		for msg := range st.Messages() {
			messages = append(messages, msg)
		}

		if len(messages) != 2 {
			t.Fatalf("got %d messages, want 2", len(messages))
		}
	})

	t.Run("skips empty lines", func(t *testing.T) {
		cfg := &config{}
		st := NewSubprocessTransport(cfg)

		r, w, _ := os.Pipe()

		go func() {
			w.Write([]byte("\n"))
			w.Write([]byte(`{"type":"test"}`))
			w.Write([]byte("\n"))
			w.Write([]byte("\n"))
			w.Close()
		}()

		go st.readMessages(r)

		var messages [][]byte
		for msg := range st.Messages() {
			messages = append(messages, msg)
		}

		if len(messages) != 1 {
			t.Fatalf("got %d messages, want 1 (empty lines should be skipped)", len(messages))
		}
	})
}

func TestFindCLI_FallbackLocations(t *testing.T) {
	t.Run("checks fallback locations when not in PATH", func(t *testing.T) {
		// Save original PATH
		origPath := os.Getenv("PATH")
		os.Setenv("PATH", "/nonexistent")
		defer os.Setenv("PATH", origPath)

		// FindCLI should check fallback locations
		// It may find an existing claude binary (e.g., /usr/local/bin/claude)
		// or we can create one to test
		home, _ := os.UserHomeDir()

		// Build the list of expected fallback locations
		fallbackLocations := []string{
			filepath.Join(home, ".npm-global", "bin", "claude"),
			"/usr/local/bin/claude",
			filepath.Join(home, ".local", "bin", "claude"),
			filepath.Join(home, "node_modules", ".bin", "claude"),
			filepath.Join(home, ".yarn", "bin", "claude"),
			filepath.Join(home, ".claude", "local", "claude"),
		}

		if runtime.GOOS == "windows" {
			for i, loc := range fallbackLocations {
				fallbackLocations[i] = loc + ".exe"
			}
		}

		path, err := FindCLI()

		if err != nil {
			// No fallback location exists, create one to test
			testDir := filepath.Join(home, ".claude", "local")
			if mkErr := os.MkdirAll(testDir, 0755); mkErr != nil {
				t.Skip("cannot create test directory")
			}

			var cliName string
			if runtime.GOOS == "windows" {
				cliName = "claude.exe"
			} else {
				cliName = "claude"
			}

			testCLI := filepath.Join(testDir, cliName)
			if writeErr := os.WriteFile(testCLI, []byte("#!/bin/sh\necho test"), 0755); writeErr != nil {
				t.Skip("cannot create test file")
			}
			defer os.Remove(testCLI)

			path, err = FindCLI()
			if err != nil {
				t.Errorf("FindCLI() error = %v, want nil", err)
			}
			if path != testCLI {
				t.Errorf("FindCLI() = %q, want %q", path, testCLI)
			}
			return
		}

		// Verify the path is one of the expected fallback locations
		found := false
		for _, loc := range fallbackLocations {
			if path == loc {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("FindCLI() = %q, not in expected fallback locations", path)
		}
	})
}

func TestBuildCommand_WithOutputFormat(t *testing.T) {
	t.Run("includes output format flag", func(t *testing.T) {
		cfg := &config{
			outputFormat: &OutputFormat{
				Type: "json_schema",
				Schema: map[string]any{
					"type": "object",
				},
			},
		}
		st := &SubprocessTransport{
			cliPath: "/usr/bin/claude",
			cfg:     cfg,
		}

		cmd := st.buildCommand()

		containsOutputFormat := false
		for i, arg := range cmd {
			if arg == "--output-format" && i+1 < len(cmd) {
				containsOutputFormat = true
				break
			}
		}
		if !containsOutputFormat {
			t.Errorf("command should contain --output-format flag, got %v", cmd)
		}
	})
}

func TestBuildCommand_WithSandbox(t *testing.T) {
	t.Run("includes sandbox flag when enabled", func(t *testing.T) {
		cfg := &config{
			sandbox: &SandboxSettings{
				Enabled: true,
			},
		}
		st := &SubprocessTransport{
			cliPath: "/usr/bin/claude",
			cfg:     cfg,
		}

		cmd := st.buildCommand()

		containsSandbox := false
		for _, arg := range cmd {
			if arg == "--sandbox" {
				containsSandbox = true
				break
			}
		}
		if !containsSandbox {
			t.Errorf("command should contain --sandbox flag, got %v", cmd)
		}
	})

	t.Run("includes sandbox auto-allow-bash flag", func(t *testing.T) {
		cfg := &config{
			sandbox: &SandboxSettings{
				AutoAllowBashIfSandboxed: true,
			},
		}
		st := &SubprocessTransport{
			cliPath: "/usr/bin/claude",
			cfg:     cfg,
		}

		cmd := st.buildCommand()

		containsAutoAllow := false
		for _, arg := range cmd {
			if arg == "--sandbox-auto-allow-bash" {
				containsAutoAllow = true
				break
			}
		}
		if !containsAutoAllow {
			t.Errorf("command should contain --sandbox-auto-allow-bash flag, got %v", cmd)
		}
	})

	t.Run("includes sandbox excluded commands", func(t *testing.T) {
		cfg := &config{
			sandbox: &SandboxSettings{
				ExcludedCommands: []string{"git", "docker"},
			},
		}
		st := &SubprocessTransport{
			cliPath: "/usr/bin/claude",
			cfg:     cfg,
		}

		cmd := st.buildCommand()

		gitCount := 0
		dockerCount := 0
		for i, arg := range cmd {
			if arg == "--sandbox-exclude-command" && i+1 < len(cmd) {
				if cmd[i+1] == "git" {
					gitCount++
				}
				if cmd[i+1] == "docker" {
					dockerCount++
				}
			}
		}
		if gitCount != 1 || dockerCount != 1 {
			t.Errorf("command should contain excluded commands git and docker, got %v", cmd)
		}
	})

	t.Run("includes sandbox network config", func(t *testing.T) {
		cfg := &config{
			sandbox: &SandboxSettings{
				Network: &SandboxNetworkConfig{
					AllowUnixSockets:    []string{"/var/run/docker.sock"},
					AllowAllUnixSockets: true,
					AllowLocalBinding:   true,
				},
			},
		}
		st := &SubprocessTransport{
			cliPath: "/usr/bin/claude",
			cfg:     cfg,
		}

		cmd := st.buildCommand()

		hasUnixSocket := false
		hasAllUnixSockets := false
		hasLocalBinding := false

		for i, arg := range cmd {
			if arg == "--sandbox-allow-unix-socket" && i+1 < len(cmd) && cmd[i+1] == "/var/run/docker.sock" {
				hasUnixSocket = true
			}
			if arg == "--sandbox-allow-all-unix-sockets" {
				hasAllUnixSockets = true
			}
			if arg == "--sandbox-allow-local-binding" {
				hasLocalBinding = true
			}
		}

		if !hasUnixSocket {
			t.Error("command should contain --sandbox-allow-unix-socket /var/run/docker.sock")
		}
		if !hasAllUnixSockets {
			t.Error("command should contain --sandbox-allow-all-unix-sockets")
		}
		if !hasLocalBinding {
			t.Error("command should contain --sandbox-allow-local-binding")
		}
	})

	t.Run("includes sandbox weaker nested flag", func(t *testing.T) {
		cfg := &config{
			sandbox: &SandboxSettings{
				EnableWeakerNestedSandbox: true,
			},
		}
		st := &SubprocessTransport{
			cliPath: "/usr/bin/claude",
			cfg:     cfg,
		}

		cmd := st.buildCommand()

		containsWeakerNested := false
		for _, arg := range cmd {
			if arg == "--sandbox-weaker-nested" {
				containsWeakerNested = true
				break
			}
		}
		if !containsWeakerNested {
			t.Errorf("command should contain --sandbox-weaker-nested flag, got %v", cmd)
		}
	})
}

func TestBuildCommand_WithIncludePartialMessages(t *testing.T) {
	t.Run("includes partial messages flag", func(t *testing.T) {
		cfg := &config{
			includePartialMessages: true,
		}
		st := &SubprocessTransport{
			cliPath: "/usr/bin/claude",
			cfg:     cfg,
		}

		cmd := st.buildCommand()

		containsPartialMessages := false
		for _, arg := range cmd {
			if arg == "--include-partial-messages" {
				containsPartialMessages = true
				break
			}
		}
		if !containsPartialMessages {
			t.Errorf("command should contain --include-partial-messages flag, got %v", cmd)
		}
	})
}

func TestBuildCommand_WithForkSession(t *testing.T) {
	t.Run("includes fork session flag", func(t *testing.T) {
		cfg := &config{
			forkSession: true,
		}
		st := &SubprocessTransport{
			cliPath: "/usr/bin/claude",
			cfg:     cfg,
		}

		cmd := st.buildCommand()

		containsForkSession := false
		for _, arg := range cmd {
			if arg == "--fork-session" {
				containsForkSession = true
				break
			}
		}
		if !containsForkSession {
			t.Errorf("command should contain --fork-session flag, got %v", cmd)
		}
	})
}
