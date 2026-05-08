package mcp

import "encoding/json"

// --- JSON-RPC base types ---

// Request represents an incoming JSON-RPC 2.0 request.
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response represents an outgoing JSON-RPC 2.0 response.
type Response struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id"`
	Result  any    `json:"result,omitempty"`
	Error   *Error `json:"error,omitempty"`
}

// Error represents a JSON-RPC error object.
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// --- MCP-specific types ---

// ToolDefinition describes a tool that can be called via MCP.
// This is what gets returned in the tools/list response.
type ToolDefinition struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

// InputSchema describes the expected input parameters for a tool.
// Uses JSON Schema format as required by the MCP spec.
type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required,omitempty"`
}

// Property describes a single input parameter.
type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

// ToolCallParams is what we receive when tools/call is invoked.
type ToolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// ToolResult is returned from a tool execution.
type ToolResult struct {
	Content []ContentBlock `json:"content"`
	IsError bool           `json:"isError,omitempty"`
}

// ContentBlock represents a piece of content in a tool result.
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// --- Helper constructors ---

// TextResult creates a successful tool result with a text response.
func TextResult(text string) ToolResult {
	return ToolResult{
		Content: []ContentBlock{
			{Type: "text", Text: text},
		},
	}
}

// ErrorResult creates an error tool result.
func ErrorResult(errMsg string) ToolResult {
	return ToolResult{
		Content: []ContentBlock{
			{Type: "text", Text: errMsg},
		},
		IsError: true,
	}
}
