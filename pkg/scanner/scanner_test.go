package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestShannonEntropy(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantMin float64
		wantMax float64
	}{
		{
			name:    "empty string",
			input:   "",
			wantMin: 0,
			wantMax: 0,
		},
		{
			name:    "single character",
			input:   "aaaaaaa",
			wantMin: 0,
			wantMax: 0.1,
		},
		{
			name:    "low entropy word",
			input:   "password",
			wantMin: 2.5,
			wantMax: 3.5,
		},
		{
			name:    "high entropy token",
			input:   "aB3xK9mP2qR7wZ4nL8vC1dF6hJ5gT0s",
			wantMin: 4.0,
			wantMax: 6.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShannonEntropy(tt.input)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("ShannonEntropy(%q) = %.2f, want between %.2f and %.2f",
					tt.input, got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestScanDirectory(t *testing.T) {
	// Create a temporary directory with test files.
	// This is better than relying on testdata/ for scan tests
	// because we control exactly what's in the directory.
	tmpDir := t.TempDir()

	// File with an AWS key — should be detected
	writeFile(t, filepath.Join(tmpDir, "config.py"),
		`# AWS config
AWS_ACCESS_KEY_ID = "AKIAIOSFODNN7EXAMPLE"
AWS_REGION = "us-east-1"
`)

	// File with a generic secret — should be detected
	writeFile(t, filepath.Join(tmpDir, "app.env"),
		`DATABASE_URL=postgres://localhost/mydb
API_KEY=sk_test_abcdefghijklmnopqrstuvwxyz1234567890
DEBUG=true
`)

	// Clean file — should produce no findings
	writeFile(t, filepath.Join(tmpDir, "clean.go"),
		`package main

import "fmt"

func main() {
    fmt.Println("Hello, world!")
}
`)

	// File with a private key header
	writeFile(t, filepath.Join(tmpDir, "id_rsa"),
		`-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA...
-----END RSA PRIVATE KEY-----
`)

	findings, err := ScanDirectory(tmpDir)
	if err != nil {
		t.Fatalf("ScanDirectory failed: %v", err)
	}

	// We expect at least 3 findings: AWS key, API key, private key
	if len(findings) < 3 {
		t.Errorf("expected at least 3 findings, got %d", len(findings))
		for _, f := range findings {
			t.Logf("  %s:%d — %s (%s)", f.File, f.Line, f.RuleName, f.Confidence)
		}
	}

	// Verify we found the AWS key specifically
	foundAWS := false
	for _, f := range findings {
		if f.RuleName == "AWS Access Key" {
			foundAWS = true
			break
		}
	}
	if !foundAWS {
		t.Error("expected to find AWS Access Key, but didn't")
	}
}

func TestScanDirectorySkipsBinaryFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a fake binary file with secret-like content
	writeFile(t, filepath.Join(tmpDir, "image.png"),
		`AKIAIOSFODNN7EXAMPLE`)

	// Create a .git directory that should be skipped
	gitDir := filepath.Join(tmpDir, ".git")

	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(gitDir, "config"),
		`AKIAIOSFODNN7EXAMPLE`)

	findings, err := ScanDirectory(tmpDir)
	if err != nil {
		t.Fatalf("ScanDirectory failed: %v", err)
	}

	if len(findings) != 0 {
		t.Errorf("expected 0 findings (binary/git files should be skipped), got %d", len(findings))
	}
}

func TestScanDirectoryEmpty(t *testing.T) {
	tmpDir := t.TempDir()

	findings, err := ScanDirectory(tmpDir)
	if err != nil {
		t.Fatalf("ScanDirectory failed: %v", err)
	}

	if len(findings) != 0 {
		t.Errorf("expected 0 findings for empty directory, got %d", len(findings))
	}
}

// writeFile is a test helper that creates a file with the given content.
// Using t.Helper() means test failures report the caller's line, not this one.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create directory %s: %v", dir, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file %s: %v", path, err)
	}
}
