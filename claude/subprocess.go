package claude

import (
	"bufio"
	"context"
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

	// System prompt
	if cfg.systemPrompt != "" {
		cmd = append(cmd, "--system-prompt", cfg.systemPrompt)
	}

	// Model
	if cfg.model != "" {
		cmd = append(cmd, "--model", cfg.model)
	}

	// Fallback model
	if cfg.fallbackModel != "" {
		cmd = append(cmd, "--fallback-model", cfg.fallbackModel)
	}

	// Max turns
	if cfg.maxTurns > 0 {
		cmd = append(cmd, "--max-turns", strconv.Itoa(cfg.maxTurns))
	}

	// Max budget
	if cfg.maxBudgetUSD > 0 {
		cmd = append(cmd, "--max-budget-usd", strconv.FormatFloat(cfg.maxBudgetUSD, 'f', -1, 64))
	}

	// Permission mode
	if cfg.permissionMode != "" {
		cmd = append(cmd, "--permission-mode", string(cfg.permissionMode))
	}

	// Allowed tools
	if len(cfg.allowedTools) > 0 {
		cmd = append(cmd, "--allowedTools", strings.Join(cfg.allowedTools, ","))
	}

	// Disallowed tools
	if len(cfg.disallowedTools) > 0 {
		cmd = append(cmd, "--disallowedTools", strings.Join(cfg.disallowedTools, ","))
	}

	// Continue conversation
	if cfg.continueConversation {
		cmd = append(cmd, "--continue")
	}

	// Resume session
	if cfg.resume != "" {
		cmd = append(cmd, "--resume", cfg.resume)
	}

	// Max thinking tokens
	if cfg.maxThinkingTokens > 0 {
		cmd = append(cmd, "--max-thinking-tokens", strconv.Itoa(cfg.maxThinkingTokens))
	}

	// MCP config
	if cfg.mcpConfig != "" {
		cmd = append(cmd, "--mcp-config", cfg.mcpConfig)
	}

	// Streaming mode: use --input-format stream-json
	cmd = append(cmd, "--input-format", "stream-json")

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
	const maxScanTokenSize = 1024 * 1024 // 1MB
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
