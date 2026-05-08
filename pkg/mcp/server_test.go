package mcp

import (
	"encoding/json"
	"testing"
)

// echoTool is a simple test tool that echoes back its input.
// This pattern — creating a minimal implementation for testing —
// is very common in Go.
type echoTool struct{}

func (e *echoTool) Definition() ToolDefinition {
	return ToolDefinition{
		Name:        "echo",
		Description: "Echoes back the input message",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"message": {Type: "string", Description: "Message to echo"},
			},
			Required: []string{"message"},
		},
	}
}

func (e *echoTool) Execute(args json.RawMessage) ToolResult {
	var input struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return ErrorResult("invalid input: " + err.Error())
	}
	return TextResult("echo: " + input.Message)
}

func TestHandleMessage(t *testing.T) {
	// Table-driven tests — the idiomatic Go testing pattern.
	// Each test case is a row in the table with a name, input, and
	// expected output. This makes it trivial to add new cases.
	tests := []struct {
		name       string
		input      string
		wantMethod string // empty means we check error instead
		wantError  bool
	}{
		{
			name:      "initialize",
			input:     `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
			wantError: false,
		},
		{
			name:      "tools/list returns registered tools",
			input:     `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`,
			wantError: false,
		},
		{
			name:      "tools/call with valid tool",
			input:     `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"echo","arguments":{"message":"hello"}}}`,
			wantError: false,
		},
		{
			name:      "tools/call with unknown tool",
			input:     `{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"nonexistent","arguments":{}}}`,
			wantError: false, // returns a result with isError, not a JSON-RPC error
		},
		{
			name:      "unknown method",
			input:     `{"jsonrpc":"2.0","id":5,"method":"unknown/method"}`,
			wantError: false, // returns JSON-RPC error in response
		},
	}

	server := NewServer("test", "0.0.1")
	server.RegisterTool(&echoTool{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := server.HandleMessage([]byte(tt.input))

			if tt.wantError && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp == nil && !tt.wantError {
				// nil response is okay for notifications
				return
			}

			// Verify response is valid JSON
			if resp != nil {
				out, err := json.Marshal(resp)
				if err != nil {
					t.Fatalf("failed to marshal response: %v", err)
				}
				t.Logf("response: %s", string(out))
			}
		})
	}
}

func TestInitializeResponse(t *testing.T) {
	server := NewServer("vaultguard", "0.1.0")
	resp, err := server.HandleMessage([]byte(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the response contains expected server info
	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatal("result is not a map")
	}

	serverInfo, ok := result["serverInfo"].(map[string]any)
	if !ok {
		t.Fatal("serverInfo is not a map")
	}

	if serverInfo["name"] != "vaultguard" {
		t.Errorf("expected server name 'vaultguard', got %v", serverInfo["name"])
	}
}

func TestToolsListReturnsRegisteredTools(t *testing.T) {
	server := NewServer("test", "0.0.1")
	server.RegisterTool(&echoTool{})

	resp, err := server.HandleMessage([]byte(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatal("result is not a map")
	}

	tools, ok := result["tools"].([]ToolDefinition)
	if !ok {
		t.Fatal("tools is not a []ToolDefinition")
	}

	if len(tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(tools))
	}

	if tools[0].Name != "echo" {
		t.Errorf("expected tool name 'echo', got %s", tools[0].Name)
	}
}

func TestToolCallEcho(t *testing.T) {
	server := NewServer("test", "0.0.1")
	server.RegisterTool(&echoTool{})

	input := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"echo","arguments":{"message":"hello world"}}}`
	resp, err := server.HandleMessage([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, ok := resp.Result.(ToolResult)
	if !ok {
		t.Fatal("result is not a ToolResult")
	}

	if result.IsError {
		t.Fatal("expected success, got error result")
	}

	if len(result.Content) == 0 {
		t.Fatal("expected content, got empty")
	}

	if result.Content[0].Text != "echo: hello world" {
		t.Errorf("expected 'echo: hello world', got '%s'", result.Content[0].Text)
	}
}
