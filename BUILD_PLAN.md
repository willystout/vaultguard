# VaultGuard — Build Plan

**MCP Server for Secret Health Monitoring**

**Timeline:** 3 weeks &nbsp;|&nbsp; **Language:** Go &nbsp;|&nbsp; **Protocol:** MCP (JSON-RPC)

**Target Application:** 1Password — Software Developer

---

## Project Overview

VaultGuard is an MCP (Model Context Protocol) server written in Go that enables AI assistants to audit and monitor secret hygiene across codebases. It exposes tools that scan for leaked credentials, check rotation policies, audit environment files, and generate health reports.

### Why This Project

This project is designed to demonstrate specific skills aligned with 1Password's job requirements:

| JD Requirement | How VaultGuard Demonstrates It | Where in Codebase |
|---|---|---|
| Go proficiency | Entire project in idiomatic Go | All source files |
| Testing methodologies | Table-driven tests, fuzz tests, integration tests | `*_test.go` files throughout |
| MCP servers (bonus) | Full MCP protocol implementation | `pkg/mcp/` |
| Auth/identity concepts (bonus) | Secret detection, rotation policy enforcement | `pkg/scanner/`, `pkg/rotation/` |
| Software best practices | CI/CD, linting, clean architecture, docs | `.github/`, `README.md` |

### MCP Tools Exposed

VaultGuard exposes four tools via the MCP protocol:

| Tool | Description | Key Concepts |
|---|---|---|
| `scan_repo` | Scans directories for hardcoded secrets (API keys, tokens, passwords in config/source files) using pattern matching and entropy analysis | Regex patterns, Shannon entropy, `.gitignore`-aware traversal |
| `check_rotation` | Reads a secrets manifest (YAML/JSON) and flags credentials past their rotation deadline based on configurable policies | Policy enforcement, time-based expiry, severity levels |
| `audit_env` | Analyzes `.env` files for weak patterns: short values, default/placeholder values, duplicates across environments | `.env` parsing, cross-env comparison, pattern detection |
| `health_report` | Aggregates findings from all tools into an overall secret health score with prioritized recommendations | Scoring algorithm, severity weighting, actionable output |

### Project Structure

Follow standard Go project layout conventions:

```
vaultguard/
├── cmd/vaultguard/        # main.go entry point
├── pkg/
│   ├── mcp/               # MCP protocol handler (JSON-RPC over stdio)
│   ├── scanner/           # Secret scanning engine (patterns + entropy)
│   ├── rotation/          # Rotation policy checker
│   ├── envaudit/          # .env file analyzer
│   └── report/            # Health report aggregator and scorer
├── internal/config/       # Configuration loading
├── testdata/              # Fixtures for testing (fake .env files, repos, manifests)
└── .github/workflows/     # CI pipeline (lint, test, build)
```

---

## Week-by-Week Build Plan

### Week 1: Foundation + MCP Protocol

**Goal:** Get a working MCP server that responds to tool calls, plus the first real tool (`scan_repo`).

#### Days 1–2: Go Fundamentals + Project Setup
- Initialize Go module: `go mod init github.com/yourusername/vaultguard`
- Set up project directory structure (see layout above)
- Write a basic `main.go` that reads from stdin and writes to stdout
- **Learn:** structs, interfaces, error handling, slices, maps
- **Practice:** Write a small program that reads JSON from stdin, processes it, writes JSON to stdout

#### Days 3–4: MCP Protocol Implementation
- Read the MCP spec (modelcontextprotocol.io) — focus on the "Tools" section
- Implement JSON-RPC message parsing (`initialize`, `tools/list`, `tools/call`)
- Build a `Server` struct that registers tool handlers
- Test with a dummy "echo" tool that returns its input
- **Key Go patterns:** interfaces for tool handlers, `encoding/json` for marshaling

#### Days 5–7: `scan_repo` Tool
- Build a file walker that respects `.gitignore` patterns
- Implement regex-based secret detection (AWS keys, generic API tokens, private keys)
- Add Shannon entropy analysis for high-entropy strings
- Return structured results: file path, line number, secret type, confidence level
- Write table-driven tests with `testdata/` fixtures

> **Week 1 Milestone:** You can run the server, connect an MCP client, call `scan_repo` on a directory, and get back a list of detected secrets.

---

### Week 2: Remaining Tools + Testing

**Goal:** Implement the remaining three tools with solid test coverage.

#### Days 8–9: `check_rotation` Tool
- Define a YAML manifest schema for secrets (`name`, `created_at`, `rotation_days`, `last_rotated`)
- Parse manifest and compare against current time
- Flag secrets as: `healthy`, `due_soon` (within 7 days), `overdue`, `critical` (2× past due)
- Write tests with time mocking (pass `time.Now` as dependency, don't use global)

#### Days 10–11: `audit_env` Tool
- Build `.env` file parser (handle comments, quoted values, multiline)
- Detect weak patterns: values under 8 chars, common defaults (`password`, `secret`, `changeme`, `TODO`)
- Cross-environment comparison: accept multiple `.env` paths, flag inconsistencies
- Write tests with various edge cases (empty files, malformed lines, unicode)

#### Days 12–14: `health_report` Tool + Integration Tests
- Build scoring algorithm: weight findings by severity, compute 0–100 score
- Generate prioritized recommendations list
- Write integration tests that spin up the full MCP server and make real tool calls
- Add fuzz tests for the scanner (`go test -fuzz`) to catch edge cases

> **Week 2 Milestone:** All four tools work end-to-end. Test coverage is 80%+. You can demo the full workflow: scan a repo, check rotation, audit envs, get a health report.

---

### Week 3: Polish + Ship

**Goal:** Make it production-quality and interview-ready.

#### Days 15–16: CI/CD + Code Quality
- Set up GitHub Actions: lint (`golangci-lint`), test, build
- Add a `Makefile` with targets: `build`, `test`, `lint`, `run`
- Run `go vet`, `staticcheck` — fix all warnings
- Review all code for idiomatic Go (named returns, error wrapping with `fmt.Errorf` + `%w`)

#### Days 17–18: Documentation
- Write a comprehensive README: what it does, why it exists, quickstart, architecture diagram
- Add GoDoc comments to all exported types and functions
- Create a `CONTRIBUTING.md` with dev setup instructions
- Record a short demo (asciinema or GIF) showing the tool in action

#### Days 19–21: Final Review + Stretch Goals
- Do a full self-code-review — pretend you're reviewing a PR
- **Stretch:** Add a Dockerfile for easy distribution
- **Stretch:** Add configurable rule sets (let users define custom patterns)
- **Stretch:** Add structured logging with `slog` (Go 1.21+ stdlib)
- Clean up commit history — make sure commits tell a story

> **Week 3 Milestone:** Repo is public, README is polished, CI is green, you can walk through any part of the code in an interview.

---

## Interview Talking Points

Be ready to discuss these aspects of VaultGuard in a 1Password interview:

### Technical Decisions
- **Why stdio over HTTP for MCP?** — MCP spec recommends stdio for local tools; simpler, no port management, works with Claude/Cursor natively
- **Why entropy + regex?** — Regex alone misses custom tokens; entropy alone has false positives. Combined approach balances precision and recall.
- **Why inject `time.Now`?** — Makes rotation tests deterministic without mocking libraries. Shows understanding of dependency injection in Go.
- **Why interfaces for tool handlers?** — Open/closed principle; adding a new tool doesn't modify the server core. Easy to test with mocks.

### What You'd Do Differently at Scale
- Add caching for large repo scans (hash-based invalidation)
- Support concurrent scanning with goroutines and worker pools
- Add metrics/tracing with OpenTelemetry for observability
- Integrate with actual secret managers (1Password CLI, HashiCorp Vault)
- Add SARIF output format for IDE/CI integration

### Connection to 1Password's Mission
- 1Password's core problem: helping teams manage secrets securely
- VaultGuard addresses the "last mile" — secrets that leak into code despite having a vault
- MCP integration means AI assistants become secret-hygiene enforcers
- Shows you think about security as an ongoing process, not a one-time setup

---

## Key Resources

- **MCP Spec:** [modelcontextprotocol.io](https://modelcontextprotocol.io) — read the Tools section thoroughly
- **Go by Example:** [gobyexample.com](https://gobyexample.com) — great for learning Go patterns quickly
- **Effective Go:** [go.dev/doc/effective_go](https://go.dev/doc/effective_go) — the official style guide
- **Go testing patterns:** Look up table-driven tests, testdata conventions, fuzz testing
- **Secret detection reference:** Study how tools like truffleHog and gitleaks approach the problem

## Daily Habit

Before you start coding each day, spend 15 minutes reading Go code from well-known projects (like 1Password's own open-source repos on GitHub). This builds pattern recognition faster than tutorials.
