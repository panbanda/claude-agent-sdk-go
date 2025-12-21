package claude

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

// osWindows is the GOOS value for Windows.
const osWindows = "windows"

// SubprocessTransport implements Transport using the Claude CLI subprocess.
type SubprocessTransport struct {
	cliPath  string
	cfg      *config
	cmd      *exec.Cmd
	stdin    io.WriteCloser
	stdout   io.ReadCloser
	messages chan []byte
	errors   chan error
	ready    bool
	mu       sync.RWMutex
}

// NewSubprocessTransport creates a new subprocess transport.
func NewSubprocessTransport(cfg *config) *SubprocessTransport {
	st := &SubprocessTransport{
		cfg:      cfg,
		messages: make(chan []byte, 100),
		errors:   make(chan error, 10),
	}

	// Use custom CLI path if provided
	if cfg.cliPath != "" {
		st.cliPath = cfg.cliPath
	}

	return st
}

// FindCLI locates the Claude CLI binary.
func FindCLI() (string, error) {
	// First check PATH
	if path, err := exec.LookPath("claude"); err == nil {
		return path, nil
	}

	// Check common installation locations
	home, _ := os.UserHomeDir()
	locations := []string{
		filepath.Join(home, ".npm-global", "bin", "claude"),
		"/usr/local/bin/claude",
		filepath.Join(home, ".local", "bin", "claude"),
		filepath.Join(home, "node_modules", ".bin", "claude"),
		filepath.Join(home, ".yarn", "bin", "claude"),
		filepath.Join(home, ".claude", "local", "claude"),
	}

	// On Windows, also check for .exe
	if runtime.GOOS == osWindows {
		for i, loc := range locations {
			locations[i] = loc + ".exe"
		}
	}

	for _, path := range locations {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path, nil
		}
	}

	return "", ErrCLINotFound
}

// buildCommand constructs the CLI command with arguments.
func (st *SubprocessTransport) buildCommand() []string {
	cmd := []string{st.cliPath, "--output-format", "stream-json", "--verbose"}
	cfg := st.cfg

	cmd = st.addBasicOptions(cmd, cfg)
	cmd = st.addToolOptions(cmd, cfg)
	cmd = st.addSessionOptions(cmd, cfg)
	cmd = st.addAdvancedOptions(cmd, cfg)
	cmd = st.addOutputOptions(cmd, cfg)
	cmd = st.addSandboxOptions(cmd, cfg)

	// Streaming mode: use --input-format stream-json
	cmd = append(cmd, "--input-format", "stream-json")

	return cmd
}

// addBasicOptions adds basic configuration options.
func (st *SubprocessTransport) addBasicOptions(cmd []string, cfg *config) []string {
	if cfg.systemPrompt != "" {
		cmd = append(cmd, "--system-prompt", cfg.systemPrompt)
	}
	if cfg.model != "" {
		cmd = append(cmd, "--model", cfg.model)
	}
	if cfg.fallbackModel != "" {
		cmd = append(cmd, "--fallback-model", cfg.fallbackModel)
	}
	if cfg.maxTurns > 0 {
		cmd = append(cmd, "--max-turns", strconv.Itoa(cfg.maxTurns))
	}
	if cfg.maxBudgetUSD > 0 {
		cmd = append(cmd, "--max-budget-usd", strconv.FormatFloat(cfg.maxBudgetUSD, 'f', -1, 64))
	}
	return cmd
}

// addToolOptions adds tool and permission configuration options.
func (st *SubprocessTransport) addToolOptions(cmd []string, cfg *config) []string {
	if cfg.permissionMode != "" {
		cmd = append(cmd, "--permission-mode", string(cfg.permissionMode))
	}
	if len(cfg.allowedTools) > 0 {
		cmd = append(cmd, "--allowedTools", strings.Join(cfg.allowedTools, ","))
	}
	if len(cfg.disallowedTools) > 0 {
		cmd = append(cmd, "--disallowedTools", strings.Join(cfg.disallowedTools, ","))
	}
	return cmd
}

// addSessionOptions adds session and thinking configuration options.
func (st *SubprocessTransport) addSessionOptions(cmd []string, cfg *config) []string {
	if cfg.continueConversation {
		cmd = append(cmd, "--continue")
	}
	if cfg.resume != "" {
		cmd = append(cmd, "--resume", cfg.resume)
	}
	if cfg.maxThinkingTokens > 0 {
		cmd = append(cmd, "--max-thinking-tokens", strconv.Itoa(cfg.maxThinkingTokens))
	}
	if cfg.mcpConfig != "" {
		cmd = append(cmd, "--mcp-config", cfg.mcpConfig)
	}
	if cfg.forkSession {
		cmd = append(cmd, "--fork-session")
	}
	return cmd
}

// addAdvancedOptions adds extra args, directories, settings, betas, agents, and plugins.
func (st *SubprocessTransport) addAdvancedOptions(cmd []string, cfg *config) []string {
	for key, value := range cfg.extraArgs {
		if value == "" {
			cmd = append(cmd, "--"+key)
		} else {
			cmd = append(cmd, "--"+key, value)
		}
	}
	for _, dir := range cfg.addDirs {
		cmd = append(cmd, "--add-dir", dir)
	}
	if cfg.settings != "" {
		cmd = append(cmd, "--settings", cfg.settings)
	}
	// Note: cfg.user is for subprocess execution context (reserved for future use),
	// not a CLI flag. Python SDK passes it to anyio.open_process(user=...).
	if len(cfg.betas) > 0 {
		cmd = append(cmd, "--betas", strings.Join(cfg.betas, ","))
	}

	// Add agents as a single JSON dict (matching Python SDK)
	if len(cfg.agents) > 0 {
		agentsDict := make(map[string]any)
		for name, agent := range cfg.agents {
			agentMap := map[string]any{
				"description": agent.Description,
				"prompt":      agent.Prompt,
			}
			// Only include non-nil/non-empty fields (matching Python SDK behavior)
			if agent.Tools != nil {
				agentMap["tools"] = agent.Tools
			}
			if agent.Model != "" {
				agentMap["model"] = agent.Model
			}
			agentsDict[name] = agentMap
		}
		agentsJSON, err := json.Marshal(agentsDict)
		if err == nil {
			cmd = append(cmd, "--agents", string(agentsJSON))
		}
	}

	// Add setting sources as comma-separated value (always included, matching Python SDK)
	sourcesValue := ""
	if len(cfg.settingSources) > 0 {
		sources := make([]string, len(cfg.settingSources))
		for i, s := range cfg.settingSources {
			sources[i] = string(s)
		}
		sourcesValue = strings.Join(sources, ",")
	}
	cmd = append(cmd, "--setting-sources", sourcesValue)

	// Add plugin directories (matching Python SDK)
	for _, plugin := range cfg.plugins {
		if plugin.Type == "local" {
			cmd = append(cmd, "--plugin-dir", plugin.Path)
		}
	}

	return cmd
}

// addOutputOptions adds output format and streaming options.
func (st *SubprocessTransport) addOutputOptions(cmd []string, cfg *config) []string {
	// Extract schema from output_format and pass as --json-schema (matching Python SDK)
	if cfg.outputFormat != nil && cfg.outputFormat.Type == OutputFormatTypeJSONSchema && cfg.outputFormat.Schema != nil {
		schemaJSON, err := json.Marshal(cfg.outputFormat.Schema)
		if err == nil {
			cmd = append(cmd, "--json-schema", string(schemaJSON))
		}
	}
	if cfg.includePartialMessages {
		cmd = append(cmd, "--include-partial-messages")
	}
	return cmd
}

// addSandboxOptions adds sandbox configuration options.
func (st *SubprocessTransport) addSandboxOptions(cmd []string, cfg *config) []string {
	if cfg.sandbox == nil {
		return cmd
	}
	if cfg.sandbox.Enabled {
		cmd = append(cmd, "--sandbox")
	}
	if cfg.sandbox.AutoAllowBashIfSandboxed {
		cmd = append(cmd, "--sandbox-auto-allow-bash")
	}
	for _, excludedCmd := range cfg.sandbox.ExcludedCommands {
		cmd = append(cmd, "--sandbox-exclude-command", excludedCmd)
	}
	if cfg.sandbox.AllowUnsandboxedCommands {
		cmd = append(cmd, "--sandbox-allow-unsandboxed")
	}
	cmd = st.addSandboxNetworkOptions(cmd, cfg.sandbox)
	if cfg.sandbox.EnableWeakerNestedSandbox {
		cmd = append(cmd, "--sandbox-weaker-nested")
	}
	return cmd
}

// addSandboxNetworkOptions adds sandbox network configuration options.
func (st *SubprocessTransport) addSandboxNetworkOptions(cmd []string, sandbox *SandboxSettings) []string {
	if sandbox.Network == nil {
		return cmd
	}
	for _, socket := range sandbox.Network.AllowUnixSockets {
		cmd = append(cmd, "--sandbox-allow-unix-socket", socket)
	}
	if sandbox.Network.AllowAllUnixSockets {
		cmd = append(cmd, "--sandbox-allow-all-unix-sockets")
	}
	if sandbox.Network.AllowLocalBinding {
		cmd = append(cmd, "--sandbox-allow-local-binding")
	}
	return cmd
}

// Connect starts the subprocess.
func (st *SubprocessTransport) Connect(ctx context.Context) error {
	st.mu.Lock()
	defer st.mu.Unlock()

	if st.ready {
		return nil
	}

	// Find CLI if not already set
	if st.cliPath == "" {
		path, err := FindCLI()
		if err != nil {
			return err
		}
		st.cliPath = path
	}

	// Build command
	args := st.buildCommand()
	st.cmd = exec.CommandContext(ctx, args[0], args[1:]...) //nolint:gosec // args are from trusted config

	// Set working directory if specified
	if st.cfg.workingDir != "" {
		st.cmd.Dir = st.cfg.workingDir
	}

	// Set environment
	st.cmd.Env = os.Environ()
	st.cmd.Env = append(st.cmd.Env, "CLAUDE_CODE_ENTRYPOINT=sdk-go")

	if st.cfg.env != nil {
		for k, v := range st.cfg.env {
			st.cmd.Env = append(st.cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// Setup pipes
	stdinPipe, err := st.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdoutPipe, err := st.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// Start the process
	if err := st.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start claude process: %w", err)
	}

	// Store pipes for writing/reading
	st.stdin = stdinPipe
	st.stdout = stdoutPipe

	// Start reading messages
	go st.readMessages(stdoutPipe)

	st.ready = true
	return nil
}

// readMessages reads from stdout and sends to messages channel.
func (st *SubprocessTransport) readMessages(stdout interface{ Read([]byte) (int, error) }) {
	defer close(st.messages)

	scanner := bufio.NewScanner(stdout)
	// Set a larger buffer for potentially large JSON messages
	maxScanTokenSize := 1024 * 1024 // 1MB default
	if st.cfg.maxBufferSize > 0 {
		maxScanTokenSize = st.cfg.maxBufferSize
	}
	buf := make([]byte, maxScanTokenSize)
	scanner.Buffer(buf, maxScanTokenSize)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		// Copy the line data since scanner reuses the buffer
		data := make([]byte, len(line))
		copy(data, line)

		select {
		case st.messages <- data:
		default:
			// Channel full, drop message
		}
	}

	if err := scanner.Err(); err != nil {
		select {
		case st.errors <- err:
		default:
		}
	}

	// Wait for process to exit
	if st.cmd != nil {
		if err := st.cmd.Wait(); err != nil {
			select {
			case st.errors <- err:
			default:
			}
		}
	}

	close(st.errors)
}

// Send writes data to the subprocess stdin.
func (st *SubprocessTransport) Send(_ context.Context, data []byte) error {
	st.mu.RLock()
	defer st.mu.RUnlock()

	if !st.ready || st.stdin == nil {
		return ErrNotConnected
	}

	_, err := st.stdin.Write(data)
	return err
}

// Messages returns the channel receiving parsed messages.
func (st *SubprocessTransport) Messages() <-chan []byte {
	return st.messages
}

// Errors returns the channel receiving errors.
func (st *SubprocessTransport) Errors() <-chan error {
	return st.errors
}

// Close terminates the subprocess.
func (st *SubprocessTransport) Close() error {
	st.mu.Lock()
	defer st.mu.Unlock()

	if !st.ready {
		return nil
	}

	st.ready = false

	// Close stdin to signal we're done
	if st.stdin != nil {
		_ = st.stdin.Close()
		st.stdin = nil
	}

	// Kill the process if still running
	if st.cmd != nil && st.cmd.Process != nil {
		_ = st.cmd.Process.Kill()
	}

	return nil
}

// IsReady returns true if the transport is ready for communication.
func (st *SubprocessTransport) IsReady() bool {
	st.mu.RLock()
	defer st.mu.RUnlock()
	return st.ready
}
