package rotation

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCheckManifest(t *testing.T) {
	testRotationDays := 90
	daysOverdue := 7
	now := time.Date(2025, time.May, 15, 0, 0, 0, 0, time.UTC)
	lastRotated := now.AddDate(0, 0, -(testRotationDays + daysOverdue))
	manifest := fmt.Sprintf(`secrets:
  - name: "test-name"
    service: "test-service"
    last_rotated: "%s"
    rotation_days: %d
    owner: "test-owner"
`, lastRotated.Format(time.RFC3339), testRotationDays)
	tmpDir := t.TempDir()

	path := filepath.Join(tmpDir, "manifest.yaml")
	writeFile(t, path, manifest)

	findings, err := CheckManifest(path, now)
	if err != nil {
		t.Fatalf("CheckManifest failed: %v", err)
	}

	if len(findings) != 1 {
		t.Fatalf("expected 1 findings, got %d", len(findings))
	}

	if findings[0].Severity != "overdue" {
		t.Errorf("Severity: got %q, want %q", findings[0].Severity, "overdue")
	}

	if findings[0].DaysOverdue != daysOverdue {
		t.Errorf("Days: %d, want %d ", findings[0].DaysOverdue, daysOverdue)
	}

}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file %s: %v", path, err)
	}
}
