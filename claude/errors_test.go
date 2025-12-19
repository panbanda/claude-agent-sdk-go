package claude

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestSentinelErrors(t *testing.T) {
	t.Run("ErrCLINotFound is an error", func(t *testing.T) {
		var err error = ErrCLINotFound
		if err == nil {
			t.Error("ErrCLINotFound should not be nil")
		}
		if !errors.Is(err, ErrCLINotFound) {
			t.Error("errors.Is should match ErrCLINotFound")
		}
	})

	t.Run("ErrNotConnected is an error", func(t *testing.T) {
		var err error = ErrNotConnected
		if err == nil {
			t.Error("ErrNotConnected should not be nil")
		}
	})

	t.Run("ErrCLIConnection is an error", func(t *testing.T) {
		var err error = ErrCLIConnection
		if err == nil {
			t.Error("ErrCLIConnection should not be nil")
		}
	})
}

func TestProcessError(t *testing.T) {
	t.Run("has exit code and stderr", func(t *testing.T) {
		err := &ProcessError{
			ExitCode: 1,
			Stderr:   "Command not found",
		}

		if err.ExitCode != 1 {
			t.Errorf("ExitCode = %d, want 1", err.ExitCode)
		}
		if err.Stderr != "Command not found" {
			t.Errorf("Stderr = %q, want %q", err.Stderr, "Command not found")
		}
	})

	t.Run("error message includes exit code", func(t *testing.T) {
		err := &ProcessError{
			ExitCode: 42,
			Stderr:   "something failed",
		}

		msg := err.Error()
		if msg == "" {
			t.Error("Error() should not return empty string")
		}
		// Should mention exit code
		if !contains(msg, "42") {
			t.Errorf("Error() = %q, should contain exit code 42", msg)
		}
	})

	t.Run("error message without stderr", func(t *testing.T) {
		err := &ProcessError{
			ExitCode: 127,
			Stderr:   "",
		}

		msg := err.Error()
		if msg == "" {
			t.Error("Error() should not return empty string")
		}
		// Should mention exit code
		if !contains(msg, "127") {
			t.Errorf("Error() = %q, should contain exit code 127", msg)
		}
		// Should not mention stderr since it's empty
		if contains(msg, "something failed") {
			t.Errorf("Error() = %q, should not contain stderr when empty", msg)
		}
	})

	t.Run("implements error interface", func(t *testing.T) {
		var err error = &ProcessError{ExitCode: 1}
		if err == nil {
			t.Error("ProcessError should implement error interface")
		}
	})

	t.Run("works with errors.As", func(t *testing.T) {
		err := &ProcessError{ExitCode: 1, Stderr: "test"}
		var wrapped error = err

		var pe *ProcessError
		if !errors.As(wrapped, &pe) {
			t.Error("errors.As should extract ProcessError")
		}
		if pe.ExitCode != 1 {
			t.Errorf("ExitCode = %d, want 1", pe.ExitCode)
		}
	})
}

func TestJSONDecodeError(t *testing.T) {
	t.Run("stores line and original error", func(t *testing.T) {
		// Create a real JSON decode error
		var jsonErr *json.SyntaxError
		err := json.Unmarshal([]byte("{invalid}"), &struct{}{})
		if !errors.As(err, &jsonErr) {
			t.Fatal("expected json.SyntaxError")
		}

		decodeErr := &JSONDecodeError{
			Line:          "{invalid}",
			OriginalError: err,
		}

		if decodeErr.Line != "{invalid}" {
			t.Errorf("Line = %q, want %q", decodeErr.Line, "{invalid}")
		}
		if decodeErr.OriginalError == nil {
			t.Error("OriginalError should not be nil")
		}
	})

	t.Run("error message mentions JSON decode", func(t *testing.T) {
		decodeErr := &JSONDecodeError{
			Line:          "bad json",
			OriginalError: errors.New("syntax error"),
		}

		msg := decodeErr.Error()
		if !contains(msg, "JSON") && !contains(msg, "json") {
			t.Errorf("Error() = %q, should mention JSON", msg)
		}
	})

	t.Run("implements error interface", func(t *testing.T) {
		var err error = &JSONDecodeError{Line: "test"}
		if err == nil {
			t.Error("JSONDecodeError should implement error interface")
		}
	})

	t.Run("Unwrap returns original error", func(t *testing.T) {
		original := errors.New("original")
		decodeErr := &JSONDecodeError{
			Line:          "test",
			OriginalError: original,
		}

		if decodeErr.Unwrap() != original {
			t.Error("Unwrap() should return OriginalError")
		}
	})
}

// contains checks if s contains substr (simple helper)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && searchString(s, substr)))
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
