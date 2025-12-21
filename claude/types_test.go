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

func TestAgentDefinition_Fields(t *testing.T) {
	t.Run("can create agent with all fields", func(t *testing.T) {
		agent := AgentDefinition{
			Description: "Test agent",
			Prompt:      "You are a test agent",
			Tools:       []string{"Read", "Write"},
			Model:       "sonnet",
		}

		if agent.Description != "Test agent" {
			t.Errorf("Description = %q, want %q", agent.Description, "Test agent")
		}
		if agent.Prompt != "You are a test agent" {
			t.Errorf("Prompt = %q, want %q", agent.Prompt, "You are a test agent")
		}
		if len(agent.Tools) != 2 {
			t.Fatalf("Tools length = %d, want 2", len(agent.Tools))
		}
		if agent.Tools[0] != "Read" || agent.Tools[1] != "Write" {
			t.Errorf("Tools = %v, want [Read Write]", agent.Tools)
		}
		if agent.Model != "sonnet" {
			t.Errorf("Model = %q, want %q", agent.Model, "sonnet")
		}
	})

	t.Run("can create agent with nil tools", func(t *testing.T) {
		agent := AgentDefinition{
			Description: "Test agent",
			Prompt:      "You are a test agent",
			Tools:       nil,
			Model:       "opus",
		}

		if agent.Tools != nil {
			t.Errorf("Tools = %v, want nil", agent.Tools)
		}
	})

	t.Run("can create agent with empty model", func(t *testing.T) {
		agent := AgentDefinition{
			Description: "Test agent",
			Prompt:      "You are a test agent",
			Model:       "",
		}

		if agent.Model != "" {
			t.Errorf("Model = %q, want empty", agent.Model)
		}
	})
}

func TestSettingSource_Constants(t *testing.T) {
	t.Run("setting source constants exist", func(t *testing.T) {
		sources := []SettingSource{
			SettingSourceUser,
			SettingSourceProject,
			SettingSourceLocal,
		}

		// All sources should be non-empty strings
		for _, s := range sources {
			if s == "" {
				t.Error("SettingSource should not be empty")
			}
		}

		// All sources should be distinct
		seen := make(map[SettingSource]bool)
		for _, s := range sources {
			if seen[s] {
				t.Errorf("duplicate source: %s", s)
			}
			seen[s] = true
		}
	})

	t.Run("setting source has expected values", func(t *testing.T) {
		if SettingSourceUser != "user" {
			t.Errorf("SettingSourceUser = %q, want %q", SettingSourceUser, "user")
		}
		if SettingSourceProject != "project" {
			t.Errorf("SettingSourceProject = %q, want %q", SettingSourceProject, "project")
		}
		if SettingSourceLocal != "local" {
			t.Errorf("SettingSourceLocal = %q, want %q", SettingSourceLocal, "local")
		}
	})
}

func TestPluginConfig_Fields(t *testing.T) {
	t.Run("can create plugin config with all fields", func(t *testing.T) {
		plugin := PluginConfig{
			Type: PluginTypeLocal,
			Path: "/path/to/plugin",
		}

		if plugin.Type != PluginTypeLocal {
			t.Errorf("Type = %q, want %q", plugin.Type, PluginTypeLocal)
		}
		if plugin.Path != "/path/to/plugin" {
			t.Errorf("Path = %q, want %q", plugin.Path, "/path/to/plugin")
		}
	})

	t.Run("can create plugin config with empty fields", func(t *testing.T) {
		plugin := PluginConfig{}

		if plugin.Type != "" {
			t.Errorf("Type = %q, want empty", plugin.Type)
		}
		if plugin.Path != "" {
			t.Errorf("Path = %q, want empty", plugin.Path)
		}
	})
}
