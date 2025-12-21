package claude

import (
	"encoding/json"
	"testing"
)

func TestOutputFormat_JSON(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"answer": map[string]any{"type": "string"},
		},
	}

	format := &OutputFormat{
		Type:   OutputFormatTypeJSONSchema,
		Schema: schema,
	}

	if format.Type != OutputFormatTypeJSONSchema {
		t.Errorf("expected type %q, got %q", OutputFormatTypeJSONSchema, format.Type)
	}

	if format.Schema == nil {
		t.Fatal("expected schema to be set")
	}

	// Verify it can be marshaled to JSON
	data, err := json.Marshal(format)
	if err != nil {
		t.Fatalf("failed to marshal OutputFormat: %v", err)
	}

	// Verify it can be unmarshaled
	var decoded OutputFormat
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal OutputFormat: %v", err)
	}

	if decoded.Type != format.Type {
		t.Errorf("expected type '%s', got '%s'", format.Type, decoded.Type)
	}
}

func TestSandboxSettings_Defaults(t *testing.T) {
	settings := &SandboxSettings{}

	if settings.Enabled {
		t.Error("expected Enabled to default to false")
	}

	if settings.AutoAllowBashIfSandboxed {
		t.Error("expected AutoAllowBashIfSandboxed to default to false")
	}

	if settings.AllowUnsandboxedCommands {
		t.Error("expected AllowUnsandboxedCommands to default to false")
	}

	if settings.EnableWeakerNestedSandbox {
		t.Error("expected EnableWeakerNestedSandbox to default to false")
	}

	if settings.ExcludedCommands != nil {
		t.Error("expected ExcludedCommands to be nil")
	}

	if settings.Network != nil {
		t.Error("expected Network to be nil")
	}

	if settings.IgnoreViolations != nil {
		t.Error("expected IgnoreViolations to be nil")
	}
}

func TestSandboxNetworkConfig_Fields(t *testing.T) {
	network := &SandboxNetworkConfig{
		AllowUnixSockets:    []string{"/var/run/docker.sock"},
		AllowAllUnixSockets: false,
		AllowLocalBinding:   true,
		HTTPProxyPort:       8080,
		SOCKSProxyPort:      1080,
	}

	if len(network.AllowUnixSockets) != 1 {
		t.Errorf("expected 1 unix socket, got %d", len(network.AllowUnixSockets))
	}

	if network.AllowUnixSockets[0] != "/var/run/docker.sock" {
		t.Errorf("expected '/var/run/docker.sock', got '%s'", network.AllowUnixSockets[0])
	}

	if network.AllowAllUnixSockets {
		t.Error("expected AllowAllUnixSockets to be false")
	}

	if !network.AllowLocalBinding {
		t.Error("expected AllowLocalBinding to be true")
	}

	if network.HTTPProxyPort != 8080 {
		t.Errorf("expected HTTPProxyPort 8080, got %d", network.HTTPProxyPort)
	}

	if network.SOCKSProxyPort != 1080 {
		t.Errorf("expected SOCKSProxyPort 1080, got %d", network.SOCKSProxyPort)
	}

	// Test JSON marshaling
	data, err := json.Marshal(network)
	if err != nil {
		t.Fatalf("failed to marshal SandboxNetworkConfig: %v", err)
	}

	var decoded SandboxNetworkConfig
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal SandboxNetworkConfig: %v", err)
	}

	if decoded.HTTPProxyPort != network.HTTPProxyPort {
		t.Errorf("expected HTTPProxyPort %d, got %d", network.HTTPProxyPort, decoded.HTTPProxyPort)
	}
}

func TestSandboxIgnoreViolations(t *testing.T) {
	violations := &SandboxIgnoreViolations{
		File:    []string{"/tmp/test.txt"},
		Network: []string{"example.com"},
	}

	if len(violations.File) != 1 {
		t.Errorf("expected 1 file violation, got %d", len(violations.File))
	}

	if len(violations.Network) != 1 {
		t.Errorf("expected 1 network violation, got %d", len(violations.Network))
	}

	// Test JSON marshaling
	data, err := json.Marshal(violations)
	if err != nil {
		t.Fatalf("failed to marshal SandboxIgnoreViolations: %v", err)
	}

	var decoded SandboxIgnoreViolations
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal SandboxIgnoreViolations: %v", err)
	}

	if len(decoded.File) != len(violations.File) {
		t.Errorf("expected %d file violations, got %d", len(violations.File), len(decoded.File))
	}
}
