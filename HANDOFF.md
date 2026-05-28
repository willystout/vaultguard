# VaultGuard — Session Handoff

> Paste this into a new chat to bring a fresh session up to speed. Last updated: 2026-05-27.

## What this is

**VaultGuard** is an MCP (Model Context Protocol) server in Go for secret health monitoring. It's a **portfolio project targeting a Software Developer role at 1Password** — it demonstrates Go proficiency, MCP protocol knowledge, testing practices, and security domain thinking.

Full 3-week build plan lives in [`BUILD_PLAN.md`](./BUILD_PLAN.md). Architecture and quickstart are in [`README.md`](./README.md). This doc is the *live state* layer on top of those.

## How to help me (READ THIS FIRST)

I'm an experienced full-stack dev (React, PHP/Symfony, MySQL, CS degree) **learning Go through this project**. Treat me as a peer developer learning a new language, not a beginner.

- **I use the "Socratic Tutor" output style, calibrated to: mostly questions, skeletons OK, no filled-in answers.** That means: lead with questions and concepts; if I'm stuck you may give a code *skeleton with holes* (e.g. `switch { case ___: ___ }`) but **never the finished, filled-in answer.** I need to write every real line myself to retain it for interviews.
  - (This is enforced by `outputStyle: "Socratic Tutor"` in `.claude/settings.local.json` — already set, no action needed.)
- **Ground every explanation in the VaultGuard codebase** — show me where a pattern already appears (e.g. "like `scanner.go:181`").
- **Tell me directly when I write non-idiomatic Go**, and show the idiomatic shape (as a skeleton, not a solution).
- **Surface the *why*, interview-style.** Every design choice, ask: "how would I defend this to an interviewer?"
- **Push back if I'm overcomplicating something.**

## Where we are RIGHT NOW

**Week 2, Days 8–9: `check_rotation` tool.** (Week 1 + `scan_repo` are done.)

### `pkg/rotation/rotation.go` — `CheckManifest()` is functionally complete ✅

Signature: `func CheckManifest(manifestPath string, now time.Time) ([]Finding, error)`

What it does now:
- Reads the manifest file, unmarshals YAML (`gopkg.in/yaml.v3`) into internal `manifest`/`secret` types
- Loops secrets, computes `dueDate = LastRotated.AddDate(0,0,RotationDays)` and `daysOverdue = int(now.Sub(dueDate).Hours()/24)`
- Classifies severity via a tagless `switch` ladder, appends a public `Finding` (with all 7 fields populated) to the result slice

Severity policy as currently coded (note: `daysOverdue > 0` means PAST due):
- `daysOverdue <= -7` → `healthy`
- `daysOverdue <= 0` → `due_soon` (range -6..0)
- `daysOverdue <= 7` → `overdue` (range 1..7)
- else → `critical` (8+)

### Concepts I worked through this session (don't re-teach unless I ask)
- Go identifier capitalization = export visibility (caught `DaysOverdue` → `daysOverdue` for a local)
- `:=` vs `=` and variable shadowing (the `findings :=` foot-gun)
- `append` returns a new slice — must reassign
- Severity is **classification, not arithmetic** — it's a `string` category derived purely from `daysOverdue`, not from subtracting dates
- Tagless `switch` with monotonic `<` ladder + `default` makes cases mutually exclusive by construction (fixed an unreachable-case + fall-through bug)
- The `time.Time` injection (`now` as a param) is decision #6 — makes tests deterministic without mock libraries

## Immediate next step

**Write `pkg/rotation/rotation_test.go`** (does not exist yet). I chose tests-first over wiring into the MCP server, for fast feedback.

The plan we agreed on:
1. Read `pkg/scanner/scanner_test.go` as the project's test template — note it uses **`t.TempDir()`** for isolation (architecture decision #5: build fixtures inline rather than depend on shared `testdata/`).
2. List 6–10 test cases in plain English first (happy path; each severity boundary at `-7`, `0`, `7`; empty manifest; malformed YAML; missing file; `RotationDays = 0`).
3. Translate them into `Test*` functions **one at a time** — I write them, the assistant reviews.

There's an existing fixture at `testdata/manifests/secrets.yaml`, but the idiomatic move here is likely writing YAML to a temp file per-test with `t.TempDir()`.

## Open items / TODOs (not blocking, but don't lose them)

1. **Severity thresholds are magic numbers in `rotation.go` ~line 51.** The in-code comment now matches the code (`healthy ≤ -7, due_soon -6..0, overdue 1..7, critical ≥ 8`), so the drift is resolved. Optional improvement: extract named threshold constants so the `switch` is self-documenting and the comment can't drift again.
2. **Module path.** `go.mod` / imports may still reference `github.com/willysecuritydev/vaultguard` (or similar placeholder) — needs to become my real GitHub username before pushing. Grep imports.
3. **Per-service severity tolerance.** Severity thresholds are currently global. Real secrets aren't equally risky (prod DB cred vs dev sandbox key). Deferred — but it's a strong *"what I'd do differently at scale"* interview answer. File as a TODO.
4. **After tests:** wire `check_rotation` into `pkg/mcp/server.go` (follow the `scan_repo` registration pattern — implement the `mcp.Tool` interface, `Definition()` + `Execute(json.RawMessage)`, then `server.RegisterTool()`; never modify `server.go` core). Then an end-to-end JSON-RPC smoke test.

## Key architecture decisions (for grounding explanations)

1. **`Tool` interface** — every tool implements `Definition()` + `Execute(args json.RawMessage)`. Adding a tool is additive; server core never changes (open/closed principle).
2. **`json.RawMessage` for tool args** — server passes raw JSON; each tool unmarshals into its own struct.
3. **Entropy + regex combined** in the scanner — regex for known patterns (precision), entropy for unknown tokens (recall).
4. **Stdio transport** — MCP spec default for local tools; no port management.
5. **`t.TempDir()` for test isolation** — tests build their own fixtures.
6. **Dependency injection for time** — pass `time.Time`, never call `time.Now()` internally.
7. **`pkg/` vs `internal/`** — `pkg/` reusable, `internal/` project-private (compiler-enforced).

## Environment notes

- Project root: `/Users/williamstout/Projects/vaultguard`
- `make test` / `make cover` / `make vet` / `make fmt` / `make lint` all defined
- CI (`.github/workflows/ci.yml`): fmt check, vet, race-detector tests, 70% coverage threshold, golangci-lint
- No `pandoc` or `defusedxml` installed locally (relevant only for docx tooling)
