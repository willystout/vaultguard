package mcp

import (
	"encoding/json"
	"fmt"
)

// Tool is the interface every VaultGuard tool must implement.
// This is a key design pattern — adding a new tool means implementing
// this interface and registering it with the server. The server itself
// never needs to change.
type Tool interface {
	// Definition returns the MCP tool definition (name, description, schema).
	Definition() ToolDefinition

	// Execute runs the tool with the given arguments and returns a result.
	Execute(args json.RawMessage) ToolResult
}

// Server handles MCP protocol communication.
type Server struct {
	name    string
	version string
	tools   map[string]Tool
}

// NewServer creates a new MCP server with the given name and version.
func NewServer(name, version string) *Server {
	return &Server{
		name:    name,
		version: version,
		tools:   make(map[string]Tool),
	}
}

// RegisterTool adds a tool to the server. The tool's name (from its
// Definition) is used as the lookup key for tools/call routing.
func (s *Server) RegisterTool(t Tool) {
	def := t.Definition()
	s.tools[def.Name] = t
}

// HandleMessage processes a single JSON-RPC message and returns a response.
// This is the main routing function — it reads the method field and
// dispatches to the appropriate handler.
func (s *Server) HandleMessage(data []byte) (*Response, error) {
	var req Request
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("invalid JSON-RPC request: %w", err)
	}

	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolsCall(req)
	case "notifications/initialized":
		// Client acknowledgment — no response needed for notifications
		return nil, nil
	default:
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &Error{
				Code:    -32601,
				Message: fmt.Sprintf("unknown method: %s", req.Method),
			},
		}, nil
	}
}

// handleInitialize responds to the MCP initialize handshake.
// The client sends this first to learn what the server supports.
func (s *Server) handleInitialize(req Request) (*Response, error) {
	return &Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]any{
				"tools": map[string]any{},
			},
			"serverInfo": map[string]any{
				"name":    s.name,
				"version": s.version,
			},
		},
	}, nil
}

// handleToolsList returns all registered tool definitions.
// Clients use this to discover what tools are available.
func (s *Server) handleToolsList(req Request) (*Response, error) {
	tools := make([]ToolDefinition, 0, len(s.tools))
	for _, t := range s.tools {
		tools = append(tools, t.Definition())
	}

	return &Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]any{
			"tools": tools,
		},
	}, nil
}

// handleToolsCall routes a tool call to the correct handler.
func (s *Server) handleToolsCall(req Request) (*Response, error) {
	var params ToolCallParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &Error{
				Code:    -32602,
				Message: fmt.Sprintf("invalid tool call params: %v", err),
			},
		}, nil
	}

	tool, ok := s.tools[params.Name]
	if !ok {
		return &Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  ErrorResult(fmt.Sprintf("unknown tool: %s", params.Name)),
		}, nil
	}

	result := tool.Execute(params.Arguments)
	return &Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}, nil
}
