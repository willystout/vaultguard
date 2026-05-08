package scanner

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/willystout/vaultguard/pkg/mcp"
)

// --- Secret detection patterns ---
// Each pattern has a name, a regex, and a description.
// This is where you'll expand as you learn more about what leaks look like.
type secretPattern struct {
	Name        string
	Pattern     *regexp.Regexp
	Description string
}

// defaultPatterns returns the built-in secret detection rules.
// Start with a small, high-confidence set. You can always add more.
func defaultPatterns() []secretPattern {
	return []secretPattern{
		{
			Name:        "AWS Access Key",
			Pattern:     regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
			Description: "AWS access key ID",
		},
		{
			Name:        "Generic API Key",
			Pattern:     regexp.MustCompile(`(?i)(api[_-]?key|apikey)\s*[=:]\s*['\"]?([a-zA-Z0-9_\-]{20,})['\"]?`),
			Description: "Generic API key assignment",
		},
		{
			Name:        "Generic Secret",
			Pattern:     regexp.MustCompile(`(?i)(secret|password|passwd|token)\s*[=:]\s*['\"]?([a-zA-Z0-9_\-!@#$%^&*]{8,})['\"]?`),
			Description: "Generic secret/password/token assignment",
		},
		{
			Name:        "Private Key Header",
			Pattern:     regexp.MustCompile(`-----BEGIN (RSA |EC |DSA |OPENSSH )?PRIVATE KEY-----`),
			Description: "Private key file content",
		},
		{
			Name:        "GitHub Token",
			Pattern:     regexp.MustCompile(`gh[pousr]_[A-Za-z0-9_]{36,}`),
			Description: "GitHub personal access token",
		},
	}
}

// --- Finding represents a single detected secret ---

// Finding is a single secret detection result.
type Finding struct {
	File        string  `json:"file"`
	Line        int     `json:"line"`
	RuleName    string  `json:"rule_name"`
	Description string  `json:"description"`
	Confidence  string  `json:"confidence"` // "high", "medium", "low"
	Entropy     float64 `json:"entropy,omitempty"`
}

// --- Shannon entropy ---

// ShannonEntropy calculates the Shannon entropy of a string.
// High entropy strings (> 4.5) are likely to be secrets/tokens.
// Low entropy strings are likely to be normal text.
// This is a useful heuristic, not a guarantee.
func ShannonEntropy(s string) float64 {
	if len(s) == 0 {
		return 0
	}

	freq := make(map[rune]float64)
	for _, c := range s {
		freq[c]++
	}

	length := float64(len(s))
	entropy := 0.0
	for _, count := range freq {
		p := count / length
		if p > 0 {
			entropy -= p * math.Log2(p)
		}
	}
	return entropy
}

// --- File scanner ---

// skipExtensions are file types we should never scan (binary, media, etc.)
var skipExtensions = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true,
	".ico": true, ".svg": true, ".woff": true, ".woff2": true,
	".ttf": true, ".eot": true, ".mp3": true, ".mp4": true,
	".zip": true, ".tar": true, ".gz": true, ".exe": true,
	".dll": true, ".so": true, ".dylib": true, ".bin": true,
	".pdf": true, ".lock": true,
}

// skipDirs are directories we should never descend into.
var skipDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"vendor":       true,
	".venv":        true,
	"__pycache__":  true,
	".idea":        true,
	".vscode":      true,
}

// ScanDirectory walks a directory and scans all text files for secrets.
func ScanDirectory(root string) ([]Finding, error) {
	patterns := defaultPatterns()
	var findings []Finding

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip files we can't access
		}

		// Skip directories we don't care about
		if info.IsDir() {
			if skipDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip binary/media files
		ext := strings.ToLower(filepath.Ext(path))
		if skipExtensions[ext] {
			return nil
		}

		// Skip large files (> 1MB) — likely not source code
		if info.Size() > 1024*1024 {
			return nil
		}

		// Scan the file
		fileFindings, err := scanFile(path, root, patterns)
		if err != nil {
			return nil // skip files we can't read
		}
		findings = append(findings, fileFindings...)

		return nil
	})

	return findings, err
}

// scanFile reads a single file and checks each line against patterns.
func scanFile(path, root string, patterns []secretPattern) ([]Finding, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Get relative path for cleaner output
	relPath, err := filepath.Rel(root, path)
	if err != nil {
		relPath = path
	}

	var findings []Finding
	lines := strings.Split(string(data), "\n")

	for i, line := range lines {
		// Check each pattern against this line
		for _, p := range patterns {
			if p.Pattern.MatchString(line) {
				findings = append(findings, Finding{
					File:        relPath,
					Line:        i + 1,
					RuleName:    p.Name,
					Description: p.Description,
					Confidence:  "high",
				})
			}
		}

		// Check for high-entropy strings that might be secrets
		// Look for assignment-like patterns with high entropy values
		if finding, ok := checkEntropy(line, relPath, i+1); ok {
			findings = append(findings, finding)
		}
	}

	return findings, nil
}

// checkEntropy looks for high-entropy values in assignment patterns.
func checkEntropy(line, file string, lineNum int) (Finding, bool) {
	// Match patterns like: KEY = "value" or KEY: "value"
	re := regexp.MustCompile(`(?i)[a-z_]+\s*[=:]\s*['\"]?([a-zA-Z0-9+/=_\-]{20,})['\"]?`)
	matches := re.FindStringSubmatch(line)
	if len(matches) < 2 {
		return Finding{}, false
	}

	value := matches[1]
	entropy := ShannonEntropy(value)

	// Threshold of 4.5 is a reasonable starting point
	// Too low = false positives on normal strings
	// Too high = misses some real secrets
	if entropy > 4.5 {
		return Finding{
			File:        file,
			Line:        lineNum,
			RuleName:    "High Entropy String",
			Description: fmt.Sprintf("Possible secret (entropy: %.2f)", entropy),
			Confidence:  "medium",
			Entropy:     entropy,
		}, true
	}

	return Finding{}, false
}

// --- MCP Tool implementation ---

// ScanRepoTool implements the mcp.Tool interface for scan_repo.
type ScanRepoTool struct{}

// NewScanRepoTool creates a new scan_repo tool instance.
func NewScanRepoTool() *ScanRepoTool {
	return &ScanRepoTool{}
}

// Definition returns the MCP tool definition for scan_repo.
func (s *ScanRepoTool) Definition() mcp.ToolDefinition {
	return mcp.ToolDefinition{
		Name:        "scan_repo",
		Description: "Scan a directory for hardcoded secrets, API keys, tokens, and other sensitive data using pattern matching and entropy analysis.",
		InputSchema: mcp.InputSchema{
			Type: "object",
			Properties: map[string]mcp.Property{
				"path": {
					Type:        "string",
					Description: "Path to the directory to scan",
				},
			},
			Required: []string{"path"},
		},
	}
}

// Execute runs the scan_repo tool.
func (s *ScanRepoTool) Execute(args json.RawMessage) mcp.ToolResult {
	var input struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return mcp.ErrorResult("invalid input: " + err.Error())
	}

	if input.Path == "" {
		return mcp.ErrorResult("path is required")
	}

	findings, err := ScanDirectory(input.Path)
	if err != nil {
		return mcp.ErrorResult("scan failed: " + err.Error())
	}

	if len(findings) == 0 {
		return mcp.TextResult("No secrets detected. Clean scan!")
	}

	// Format results as JSON for structured output
	out, err := json.MarshalIndent(findings, "", "  ")
	if err != nil {
		return mcp.ErrorResult("failed to format results: " + err.Error())
	}

	summary := fmt.Sprintf("Found %d potential secret(s):\n\n%s", len(findings), string(out))
	return mcp.TextResult(summary)
}
