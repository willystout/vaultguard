package rotation

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCheckManifestSeverity(t *testing.T) {
	tests := []struct {
		name              string
		targetDaysOverdue int
		wantSeverity      string
	}{
		{
			name:              "Healthy (-30)",
			targetDaysOverdue: -30,
			wantSeverity:      "healthy",
		},
		{
			name:              "Healthy (-7)",
			targetDaysOverdue: -7,
			wantSeverity:      "healthy",
		},
		{
			name:              "Due Soon (-3)",
			targetDaysOverdue: -3,
			wantSeverity:      "due_soon",
		},
		{
			name:              "Due Soon (0)",
			targetDaysOverdue: 0,
			wantSeverity:      "due_soon",
		},
		{
			name:              "Overdue (3)",
			targetDaysOverdue: 3,
			wantSeverity:      "overdue",
		},
		{
			name:              "Overdue (7)",
			targetDaysOverdue: 7,
			wantSeverity:      "overdue",
		},
		{
			name:              "Critical (30)",
			targetDaysOverdue: 30,
			wantSeverity:      "critical",
		},
	}

	testRotationDays := 90
	now := time.Date(2025, time.May, 15, 0, 0, 0, 0, time.UTC)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lastRotated := now.AddDate(0, 0, -(testRotationDays + tt.targetDaysOverdue))
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

			if findings[0].Severity != tt.wantSeverity {
				t.Errorf("Severity: got %q, want %q", findings[0].Severity, tt.wantSeverity)
			}

			if findings[0].DaysOverdue != tt.targetDaysOverdue {
				t.Errorf("Days: %d, want %d ", findings[0].DaysOverdue, tt.targetDaysOverdue)
			}
		})
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file %s: %v", path, err)
	}
}
