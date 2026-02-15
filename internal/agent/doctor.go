package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Severity indicates how serious a finding is.
type Severity int

const (
	SeverityOK       Severity = iota
	SeverityWarning
	SeverityCritical
)

// String returns a human-readable severity name.
func (s Severity) String() string {
	switch s {
	case SeverityOK:
		return "OK"
	case SeverityWarning:
		return "WARNING"
	case SeverityCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// Indicator returns a display character for the severity.
func (s Severity) Indicator() string {
	switch s {
	case SeverityCritical:
		return "[!]"
	case SeverityWarning:
		return "[~]"
	default:
		return "[ok]"
	}
}

// Finding represents a single diagnostic result.
type Finding struct {
	Severity   Severity
	Label      string
	PlistPath  string
	Message    string
	Suggestion string
}

// Doctor runs health checks across all services.
type Doctor struct {
	scanner *Scanner
}

// NewDoctor creates a new doctor instance.
func NewDoctor(scanner *Scanner) *Doctor {
	return &Doctor{scanner: scanner}
}

// Check runs all health checks and returns findings sorted by severity.
func (d *Doctor) Check() ([]Finding, error) {
	services, err := d.scanner.ScanAll()
	if err != nil {
		return nil, fmt.Errorf("failed to scan services: %w", err)
	}

	var findings []Finding

	findings = append(findings, d.checkMissingBinaries(services)...)
	findings = append(findings, d.checkPermissions(services)...)
	findings = append(findings, d.checkDuplicateLabels(services)...)
	findings = append(findings, d.checkFilenameMismatch(services)...)
	findings = append(findings, d.checkCrashedServices(services)...)
	findings = append(findings, d.checkStaleLogPaths(services)...)
	findings = append(findings, d.checkMissingLabels(services)...)

	// Sort by severity (critical first).
	sortFindings(findings)

	return findings, nil
}

// checkMissingBinaries reports services whose program path does not exist on disk.
func (d *Doctor) checkMissingBinaries(services []Service) []Finding {
	var findings []Finding
	for _, svc := range services {
		binary := svc.BinaryPath()
		if binary == "" {
			continue
		}
		if _, err := os.Stat(binary); os.IsNotExist(err) {
			findings = append(findings, Finding{
				Severity:   SeverityCritical,
				Label:      svc.Label,
				PlistPath:  svc.PlistPath,
				Message:    fmt.Sprintf("binary not found at %s", binary),
				Suggestion: "Remove or update the plist to point to a valid binary",
			})
		}
	}
	return findings
}

// checkPermissions reports world-writable plists and ownership issues.
func (d *Doctor) checkPermissions(services []Service) []Finding {
	var findings []Finding
	for _, svc := range services {
		if svc.PlistPath == "" {
			continue
		}
		info, err := os.Stat(svc.PlistPath)
		if err != nil {
			continue
		}
		mode := info.Mode()
		// Check for world-writable.
		if mode&0002 != 0 {
			findings = append(findings, Finding{
				Severity:   SeverityWarning,
				Label:      svc.Label,
				PlistPath:  svc.PlistPath,
				Message:    "world-writable plist",
				Suggestion: fmt.Sprintf("Fix permissions: chmod 644 %s", svc.PlistPath),
			})
		}
	}
	return findings
}

// checkDuplicateLabels reports multiple plists across domains with the same label.
func (d *Doctor) checkDuplicateLabels(services []Service) []Finding {
	labelCount := make(map[string][]string)
	for _, svc := range services {
		if svc.PlistPath != "" {
			labelCount[svc.Label] = append(labelCount[svc.Label], svc.PlistPath)
		}
	}

	var findings []Finding
	for label, paths := range labelCount {
		if len(paths) > 1 {
			findings = append(findings, Finding{
				Severity:   SeverityWarning,
				Label:      label,
				Message:    fmt.Sprintf("duplicate label found in %d plists: %s", len(paths), strings.Join(paths, ", ")),
				Suggestion: "Remove duplicate plists or use unique labels",
			})
		}
	}
	return findings
}

// checkFilenameMismatch reports plists whose filename does not match the Label key.
func (d *Doctor) checkFilenameMismatch(services []Service) []Finding {
	var findings []Finding
	for _, svc := range services {
		if svc.PlistPath == "" || svc.Label == "" {
			continue
		}
		filename := strings.TrimSuffix(filepath.Base(svc.PlistPath), ".plist")
		if filename != svc.Label {
			findings = append(findings, Finding{
				Severity:   SeverityWarning,
				Label:      svc.Label,
				PlistPath:  svc.PlistPath,
				Message:    fmt.Sprintf("filename %q does not match label %q", filepath.Base(svc.PlistPath), svc.Label),
				Suggestion: "Rename the plist to match its Label key",
			})
		}
	}
	return findings
}

// checkCrashedServices reports services with non-zero last exit status.
func (d *Doctor) checkCrashedServices(services []Service) []Finding {
	var findings []Finding
	for _, svc := range services {
		if svc.LastExitStatus != 0 && svc.Status != StatusRunning {
			findings = append(findings, Finding{
				Severity:   SeverityCritical,
				Label:      svc.Label,
				PlistPath:  svc.PlistPath,
				Message:    fmt.Sprintf("last exit status: %d", svc.LastExitStatus),
				Suggestion: "Check logs for the service to diagnose the crash",
			})
		}
	}
	return findings
}

// checkStaleLogPaths reports services whose log paths point to non-existent directories.
func (d *Doctor) checkStaleLogPaths(services []Service) []Finding {
	var findings []Finding
	for _, svc := range services {
		for _, logPath := range []string{svc.StandardOutPath, svc.StandardErrorPath} {
			if logPath == "" {
				continue
			}
			dir := filepath.Dir(logPath)
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				findings = append(findings, Finding{
					Severity:   SeverityWarning,
					Label:      svc.Label,
					PlistPath:  svc.PlistPath,
					Message:    fmt.Sprintf("log directory does not exist: %s", dir),
					Suggestion: fmt.Sprintf("Create the directory: mkdir -p %s", dir),
				})
			}
		}
	}
	return findings
}

// checkMissingLabels reports plists that have no Label key set.
func (d *Doctor) checkMissingLabels(services []Service) []Finding {
	var findings []Finding
	for _, svc := range services {
		if svc.Label == "" && svc.PlistPath != "" {
			findings = append(findings, Finding{
				Severity:   SeverityCritical,
				Label:      filepath.Base(svc.PlistPath),
				PlistPath:  svc.PlistPath,
				Message:    "plist missing required Label key",
				Suggestion: "Add a Label key to the plist",
			})
		}
	}
	return findings
}

// sortFindings sorts findings by severity (critical first, then warning, then ok).
func sortFindings(findings []Finding) {
	for i := 1; i < len(findings); i++ {
		key := findings[i]
		j := i - 1
		for j >= 0 && findings[j].Severity < key.Severity {
			findings[j+1] = findings[j]
			j--
		}
		findings[j+1] = key
	}
}

// CountBySeverity returns counts grouped by severity.
func CountBySeverity(findings []Finding) (critical, warning, ok int) {
	for _, f := range findings {
		switch f.Severity {
		case SeverityCritical:
			critical++
		case SeverityWarning:
			warning++
		default:
			ok++
		}
	}
	return
}
