package claude

import (
	"encoding/json"
	"testing"
)

func TestControlRequestSubtype(t *testing.T) {
	t.Run("subtype constants are defined", func(t *testing.T) {
		subtypes := []ControlRequestSubtype{
			ControlSubtypeInterrupt,
			ControlSubtypeCanUseTool,
			ControlSubtypeInitialize,
			ControlSubtypeSetPermissionMode,
			ControlSubtypeHookCallback,
			ControlSubtypeMcpMessage,
			ControlSubtypeRewindFiles,
		}

		// All should be non-empty strings
		for _, s := range subtypes {
			if s == "" {
				t.Error("ControlRequestSubtype should not be empty")
			}
		}

		// All should be distinct
		seen := make(map[ControlRequestSubtype]bool)
		for _, s := range subtypes {
			if seen[s] {
				t.Errorf("duplicate subtype: %s", s)
			}
			seen[s] = true
		}
	})
}

func TestControlRequest(t *testing.T) {
	t.Run("marshal interrupt request", func(t *testing.T) {
		req := NewInterruptRequest()

		data, err := json.Marshal(req)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}

		var result map[string]any
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}

		if result["type"] != "control_request" {
			t.Errorf("type = %v, want 'control_request'", result["type"])
		}

		request, ok := result["request"].(map[string]any)
		if !ok {
			t.Fatal("request should be a map")
		}

		if request["subtype"] != "interrupt" {
			t.Errorf("subtype = %v, want 'interrupt'", request["subtype"])
		}
	})

	t.Run("marshal initialize request", func(t *testing.T) {
		hooks := map[HookEvent][]HookDefinition{
			PreToolUse: {
				{Matcher: "Bash", Timeout: 30000},
			},
		}
		req := NewInitializeRequest(hooks)

		data, err := json.Marshal(req)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}

		var result map[string]any
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}

		request, ok := result["request"].(map[string]any)
		if !ok {
			t.Fatal("request should be a map")
		}

		if request["subtype"] != "initialize" {
			t.Errorf("subtype = %v, want 'initialize'", request["subtype"])
		}

		hooksData, ok := request["hooks"].(map[string]any)
		if !ok {
			t.Fatal("hooks should be a map")
		}

		if hooksData["PreToolUse"] == nil {
			t.Error("hooks should contain PreToolUse")
		}
	})

	t.Run("marshal permission request response", func(t *testing.T) {
		resp := &ControlResponse{
			Type: "control_response",
			Response: &ControlResponsePayload{
				Subtype:   "success",
				RequestID: "req-123",
				Response: &PermissionResultResponse{
					Behavior: "allow",
				},
			},
		}

		data, err := json.Marshal(resp)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}

		var result map[string]any
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}

		if result["type"] != "control_response" {
			t.Errorf("type = %v, want 'control_response'", result["type"])
		}
	})
}

func TestControlResponse(t *testing.T) {
	t.Run("unmarshal success response", func(t *testing.T) {
		data := `{
			"type": "control_response",
			"response": {
				"subtype": "success",
				"request_id": "req-123",
				"response": null
			}
		}`

		var resp ControlResponse
		if err := json.Unmarshal([]byte(data), &resp); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}

		if resp.Type != "control_response" {
			t.Errorf("Type = %q, want 'control_response'", resp.Type)
		}
		if resp.Response.Subtype != "success" {
			t.Errorf("Subtype = %q, want 'success'", resp.Response.Subtype)
		}
		if resp.Response.RequestID != "req-123" {
			t.Errorf("RequestID = %q, want 'req-123'", resp.Response.RequestID)
		}
	})

	t.Run("unmarshal error response", func(t *testing.T) {
		data := `{
			"type": "control_response",
			"response": {
				"subtype": "error",
				"request_id": "req-456",
				"error": "Something went wrong"
			}
		}`

		var resp ControlResponse
		if err := json.Unmarshal([]byte(data), &resp); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}

		if resp.Response.Subtype != "error" {
			t.Errorf("Subtype = %q, want 'error'", resp.Response.Subtype)
		}
		if resp.Response.Error != "Something went wrong" {
			t.Errorf("Error = %q, want 'Something went wrong'", resp.Response.Error)
		}
	})
}

func TestCanUseToolRequest(t *testing.T) {
	t.Run("unmarshal permission request", func(t *testing.T) {
		data := `{
			"type": "control_request",
			"request_id": "req-789",
			"request": {
				"subtype": "can_use_tool",
				"tool_name": "Bash",
				"input": {"command": "ls"},
				"permission_suggestions": null,
				"blocked_path": null
			}
		}`

		var req ControlRequest
		if err := json.Unmarshal([]byte(data), &req); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}

		if req.Type != "control_request" {
			t.Errorf("Type = %q, want 'control_request'", req.Type)
		}
		if req.RequestID != "req-789" {
			t.Errorf("RequestID = %q, want 'req-789'", req.RequestID)
		}
		if req.Request.Subtype != ControlSubtypeCanUseTool {
			t.Errorf("Subtype = %q, want 'can_use_tool'", req.Request.Subtype)
		}
		if req.Request.ToolName != "Bash" {
			t.Errorf("ToolName = %q, want 'Bash'", req.Request.ToolName)
		}
	})
}

func TestHookCallbackRequest(t *testing.T) {
	t.Run("unmarshal hook callback request", func(t *testing.T) {
		data := `{
			"type": "control_request",
			"request_id": "req-hook-1",
			"request": {
				"subtype": "hook_callback",
				"callback_id": "cb-123",
				"input": {
					"session_id": "sess-1",
					"transcript_path": "/tmp/transcript.json",
					"cwd": "/home/user",
					"hook_event_name": "PreToolUse",
					"tool_name": "Bash",
					"tool_input": {"command": "ls"}
				},
				"tool_use_id": "tu-456"
			}
		}`

		var req ControlRequest
		if err := json.Unmarshal([]byte(data), &req); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}

		if req.Request.Subtype != ControlSubtypeHookCallback {
			t.Errorf("Subtype = %q, want 'hook_callback'", req.Request.Subtype)
		}
		if req.Request.CallbackID != "cb-123" {
			t.Errorf("CallbackID = %q, want 'cb-123'", req.Request.CallbackID)
		}
		if req.Request.ToolUseID != "tu-456" {
			t.Errorf("ToolUseID = %q, want 'tu-456'", req.Request.ToolUseID)
		}
	})
}

func TestHookDefinition(t *testing.T) {
	t.Run("marshal hook definition", func(t *testing.T) {
		def := HookDefinition{
			Matcher: "Bash|Read",
			Timeout: 60000,
		}

		data, err := json.Marshal(def)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}

		var result map[string]any
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}

		if result["matcher"] != "Bash|Read" {
			t.Errorf("matcher = %v, want 'Bash|Read'", result["matcher"])
		}
		if result["timeout"] != float64(60000) {
			t.Errorf("timeout = %v, want 60000", result["timeout"])
		}
	})
}

func TestPermissionResultResponse(t *testing.T) {
	t.Run("allow response", func(t *testing.T) {
		resp := &PermissionResultResponse{
			Behavior: "allow",
		}

		data, err := json.Marshal(resp)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}

		var result map[string]any
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}

		if result["behavior"] != "allow" {
			t.Errorf("behavior = %v, want 'allow'", result["behavior"])
		}
	})

	t.Run("deny response with message", func(t *testing.T) {
		resp := &PermissionResultResponse{
			Behavior:  "deny",
			Message:   "Not allowed",
			Interrupt: true,
		}

		data, err := json.Marshal(resp)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}

		var result map[string]any
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}

		if result["behavior"] != "deny" {
			t.Errorf("behavior = %v, want 'deny'", result["behavior"])
		}
		if result["message"] != "Not allowed" {
			t.Errorf("message = %v, want 'Not allowed'", result["message"])
		}
		if result["interrupt"] != true {
			t.Errorf("interrupt = %v, want true", result["interrupt"])
		}
	})

	t.Run("allow response with updated input", func(t *testing.T) {
		resp := &PermissionResultResponse{
			Behavior: "allow",
			UpdatedInput: map[string]any{
				"command": "ls -la",
			},
		}

		data, err := json.Marshal(resp)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}

		var result map[string]any
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}

		updatedInput, ok := result["updated_input"].(map[string]any)
		if !ok {
			t.Fatal("updated_input should be a map")
		}
		if updatedInput["command"] != "ls -la" {
			t.Errorf("updated_input.command = %v, want 'ls -la'", updatedInput["command"])
		}
	})
}

func TestHookCallbackResponse(t *testing.T) {
	t.Run("basic response", func(t *testing.T) {
		resp := &HookCallbackResponse{
			Continue: true,
		}

		data, err := json.Marshal(resp)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}

		var result map[string]any
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}

		if result["continue"] != true {
			t.Errorf("continue = %v, want true", result["continue"])
		}
	})

	t.Run("response with permission decision", func(t *testing.T) {
		resp := &HookCallbackResponse{
			Continue: true,
			HookSpecificOutput: &HookSpecificOutput{
				HookEventName:      PreToolUse,
				PermissionDecision: "allow",
			},
		}

		data, err := json.Marshal(resp)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}

		var result map[string]any
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}

		output, ok := result["hookSpecificOutput"].(map[string]any)
		if !ok {
			t.Fatal("hookSpecificOutput should be a map")
		}
		if output["hookEventName"] != string(PreToolUse) {
			t.Errorf("hookEventName = %v, want 'PreToolUse'", output["hookEventName"])
		}
		if output["permissionDecision"] != "allow" {
			t.Errorf("permissionDecision = %v, want 'allow'", output["permissionDecision"])
		}
	})

	t.Run("response with stop reason", func(t *testing.T) {
		resp := &HookCallbackResponse{
			Continue:   false,
			StopReason: "User requested stop",
		}

		data, err := json.Marshal(resp)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}

		var result map[string]any
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}

		if result["continue"] != false {
			t.Errorf("continue = %v, want false", result["continue"])
		}
		if result["stopReason"] != "User requested stop" {
			t.Errorf("stopReason = %v, want 'User requested stop'", result["stopReason"])
		}
	})
}

func TestNewInterruptRequest(t *testing.T) {
	t.Run("creates valid interrupt request", func(t *testing.T) {
		req := NewInterruptRequest()

		if req.Type != "control_request" {
			t.Errorf("Type = %q, want 'control_request'", req.Type)
		}
		if req.RequestID == "" {
			t.Error("RequestID should not be empty")
		}
		if req.Request.Subtype != ControlSubtypeInterrupt {
			t.Errorf("Subtype = %q, want 'interrupt'", req.Request.Subtype)
		}
	})
}

func TestNewInitializeRequest(t *testing.T) {
	t.Run("creates valid initialize request", func(t *testing.T) {
		hooks := map[HookEvent][]HookDefinition{
			PreToolUse: {
				{Matcher: "Bash"},
			},
		}
		req := NewInitializeRequest(hooks)

		if req.Type != "control_request" {
			t.Errorf("Type = %q, want 'control_request'", req.Type)
		}
		if req.RequestID == "" {
			t.Error("RequestID should not be empty")
		}
		if req.Request.Subtype != ControlSubtypeInitialize {
			t.Errorf("Subtype = %q, want 'initialize'", req.Request.Subtype)
		}
	})
}

func TestNewControlResponseSuccess(t *testing.T) {
	t.Run("creates success response", func(t *testing.T) {
		resp := NewControlResponseSuccess("req-123", nil)

		if resp.Type != "control_response" {
			t.Errorf("Type = %q, want 'control_response'", resp.Type)
		}
		if resp.Response.Subtype != "success" {
			t.Errorf("Subtype = %q, want 'success'", resp.Response.Subtype)
		}
		if resp.Response.RequestID != "req-123" {
			t.Errorf("RequestID = %q, want 'req-123'", resp.Response.RequestID)
		}
	})
}

func TestNewControlResponseError(t *testing.T) {
	t.Run("creates error response", func(t *testing.T) {
		resp := NewControlResponseError("req-456", "Something went wrong")

		if resp.Type != "control_response" {
			t.Errorf("Type = %q, want 'control_response'", resp.Type)
		}
		if resp.Response.Subtype != "error" {
			t.Errorf("Subtype = %q, want 'error'", resp.Response.Subtype)
		}
		if resp.Response.Error != "Something went wrong" {
			t.Errorf("Error = %q, want 'Something went wrong'", resp.Response.Error)
		}
	})
}

func TestNewSetPermissionModeRequest(t *testing.T) {
	t.Run("creates set permission mode request", func(t *testing.T) {
		req := NewSetPermissionModeRequest(PermissionBypass)

		if req.Type != "control_request" {
			t.Errorf("Type = %q, want 'control_request'", req.Type)
		}
		if req.RequestID == "" {
			t.Error("RequestID should not be empty")
		}
		if req.Request.Subtype != ControlSubtypeSetPermissionMode {
			t.Errorf("Subtype = %q, want 'set_permission_mode'", req.Request.Subtype)
		}
		if req.Request.Mode != string(PermissionBypass) {
			t.Errorf("Mode = %q, want '%s'", req.Request.Mode, PermissionBypass)
		}
	})

	t.Run("creates request with accept edits mode", func(t *testing.T) {
		req := NewSetPermissionModeRequest(PermissionAcceptEdits)

		if req.Request.Mode != string(PermissionAcceptEdits) {
			t.Errorf("Mode = %q, want '%s'", req.Request.Mode, PermissionAcceptEdits)
		}
	})
}
