# VaultGuard

An MCP (Model Context Protocol) server for secret health monitoring, written in Go.

VaultGuard lets AI assistants audit and monitor secret hygiene across codebases — scanning for leaked credentials, checking rotation policies, auditing environment files, and generating health reports.

## Why?

Secrets leak. Even teams with a password manager end up with API keys in config files, stale tokens that haven't been rotated in months, and `.env` files with `changeme` still in production. VaultGuard catches these problems by making secret hygiene checkable through any MCP-compatible AI assistant.

## Tools

| Tool | Description |
|------|-------------|
| `scan_repo` | Scans directories for hardcoded secrets using pattern matching + entropy analysis |
| `check_rotation` | Checks a secrets manifest and flags credentials past their rotation policy |
| `audit_env` | Analyzes `.env` files for weak patterns, defaults, and cross-environment issues |
| `health_report` | Aggregates all findings into a 0–100 health score with recommendations |

## Quickstart

```bash
# Build
make build

# Run tests
make test

# Run with coverage
make cover
```

### Using with an MCP client

VaultGuard communicates over stdio using JSON-RPC, following the [MCP specification](https://modelcontextprotocol.io).

```bash
# Start the server (it reads from stdin, writes to stdout)
./bin/vaultguard
```

Example: send an initialize request followed by a tool call:

```json
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}
{"jsonrpc":"2.0","id":2,"method":"tools/list"}
{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"scan_repo","arguments":{"path":"./testdata/repos/fake-project"}}}
```

### Configuring in Claude Desktop

Add to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "vaultguard": {
      "command": "/path/to/vaultguard"
    }
  }
}
```

## Project Structure

```
vaultguard/
├── cmd/vaultguard/       # Entry point
│   └── main.go
├── pkg/
│   ├── mcp/              # MCP protocol handler (JSON-RPC over stdio)
│   │   ├── server.go     # Request routing and tool dispatch
│   │   ├── types.go      # JSON-RPC and MCP type definitions
│   │   └── server_test.go
│   ├── scanner/           # Secret scanning engine
│   │   ├── scanner.go     # Pattern matching + entropy analysis
│   │   └── scanner_test.go
│   ├── rotation/          # Rotation policy checker
│   │   └── rotation.go
│   ├── envaudit/          # .env file analyzer
│   │   └── envaudit.go
│   └── report/            # Health report aggregator
│       └── report.go
├── testdata/              # Test fixtures
│   ├── repos/fake-project/
│   ├── manifests/
│   └── envfiles/
├── .github/workflows/     # CI pipeline
├── Makefile
└── go.mod
```

## Architecture

```
stdin (JSON-RPC) → Server → Router → Tool Handler → Result → stdout (JSON-RPC)
                     │
                     ├── initialize    → server info + capabilities
                     ├── tools/list    → registered tool definitions
                     └── tools/call    → dispatch to tool by name
                           │
                           ├── scan_repo      → scanner.ScanDirectory()
                           ├── check_rotation  → rotation.CheckManifest()
                           ├── audit_env       → envaudit.AuditFiles()
                           └── health_report   → report.Generate()
```

**Key design decisions:**
- **Interfaces for tools** — Adding a new tool means implementing the `mcp.Tool` interface and calling `server.RegisterTool()`. The server core never changes.
- **Dependency injection** — Time, filesystem access, and other dependencies are passed as parameters, not globals. This makes everything testable.
- **Standard library first** — No external dependencies beyond what's necessary. Shows comfort with Go's stdlib.

## Testing

```bash
# Run all tests
make test

# Run with race detector
go test -race ./...

# Run fuzz tests (scanner)
make fuzz

# Coverage report
make cover
```

## Development

```bash
# Format code
make fmt

# Run go vet
make vet

# Lint (requires golangci-lint)
make lint
```

## License

MIT
