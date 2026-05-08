package rotation

// TODO: Implement check_rotation tool
//
// This tool reads a secrets manifest (YAML) and flags credentials
// that are past their rotation deadline.
//
// Key design decisions to make:
// - Define the manifest YAML schema (see testdata/manifests/ for examples)
// - Accept time.Now as a parameter (not global) for testability
// - Return findings with severity: healthy, due_soon, overdue, critical
//
// Start here after you finish scanner and the MCP server is solid.
