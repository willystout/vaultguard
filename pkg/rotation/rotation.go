// Package rotation implements the check_rotation MCP tool: it reads
// a secrets manifest and flags credentials past their rotation policy.
package rotation

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type manifest struct {
	Secrets []secret `yaml:"secrets"`
}

type secret struct {
	Name         string    `yaml:"name"`
	Service      string    `yaml:"service"`
	LastRotated  time.Time `yaml:"last_rotated"`
	RotationDays int       `yaml:"rotation_days"`
	Owner        string    `yaml:"owner"`
}

type Finding struct {
	Name         string    `json:"name"`
	Service      string    `json:"service"`
	Owner        string    `json:"owner"`
	LastRotated  time.Time `json:"last_rotated"`
	RotationDays int       `json:"rotation_days"`
	DaysOverdue  int       `json:"days_overdue"`
	Severity     string    `json:"severity"`
}

func CheckManifest(manifestPath string, now time.Time) ([]Finding, error) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}
	var findings []Finding

	var m manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	for _, s := range m.Secrets {
		// s is a single secret struct with .name .service, .LastRotated, etc.
		var dueDate = s.LastRotated.AddDate(0, 0, s.RotationDays)
		var daysOverdue = int(now.Sub(dueDate).Hours() / 24)
		var severity string
		// Severity policy: healthy ≤ -7 days, due_soon -6..0, overdue 1..7, critical ≥ 8
		switch {
		case daysOverdue <= -7:
			severity = "healthy"
		case daysOverdue <= 0:
			severity = "due_soon"
		case daysOverdue <= 7:
			severity = "overdue"
		default:
			severity = "critical"
		}
		findings = append(findings, Finding{
			Name:         s.Name,
			Service:      s.Service,
			Owner:        s.Owner,
			LastRotated:  s.LastRotated,
			RotationDays: s.RotationDays,
			DaysOverdue:  daysOverdue,
			Severity:     severity,
		})
	}

	return findings, nil

}
