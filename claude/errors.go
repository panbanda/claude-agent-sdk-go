package claude

import (
	"errors"
	"fmt"
)

// Sentinel errors for simple error cases.
// Use errors.Is() to check for these.
var (
	// ErrCLINotFound indicates the Claude CLI binary was not found.
	ErrCLINotFound = errors.New("claude: CLI not found")

	// ErrNotConnected indicates an operation was attempted before connecting.
	ErrNotConnected = errors.New("claude: not connected")

	// ErrCLIConnection indicates a failure to connect to the CLI process.
	ErrCLIConnection = errors.New("claude: CLI connection failed")
)

// ProcessError represents a CLI process failure with exit code and stderr output.
// Use errors.As() to extract this from wrapped errors.
type ProcessError struct {
	ExitCode int
	Stderr   string
}

func (e *ProcessError) Error() string {
	if e.Stderr != "" {
		return fmt.Sprintf("claude: process exited with code %d: %s", e.ExitCode, e.Stderr)
	}
	return fmt.Sprintf("claude: process exited with code %d", e.ExitCode)
}

// JSONDecodeError represents a failure to parse JSON from the CLI.
// Wraps the original json error and includes the problematic line.
type JSONDecodeError struct {
	Line          string
	OriginalError error
}

func (e *JSONDecodeError) Error() string {
	return fmt.Sprintf("claude: failed to decode JSON: %v", e.OriginalError)
}

// Unwrap returns the original error for use with errors.Is/As.
func (e *JSONDecodeError) Unwrap() error {
	return e.OriginalError
}
