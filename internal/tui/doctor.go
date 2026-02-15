package tui

import (
	"fmt"
	"strings"

	"github.com/lu-zhengda/lanchr/internal/agent"
)

// renderDoctorView renders the doctor report.
func renderDoctorView(findings []agent.Finding, scrollOffset int, height int) string {
	var b strings.Builder

	b.WriteString(titleBar.Render("Doctor Report"))
	b.WriteString("\n\n")

	if len(findings) == 0 {
		b.WriteString(findingOK.Render("  All services passed health checks."))
		b.WriteString("\n")
		return b.String()
	}

	critical, warning, _ := agent.CountBySeverity(findings)

	// Summary.
	if critical > 0 {
		b.WriteString(findingCritical.Render(fmt.Sprintf("  CRITICAL: %d", critical)))
		b.WriteString("\n")
	}
	if warning > 0 {
		b.WriteString(findingWarning.Render(fmt.Sprintf("  WARNING:  %d", warning)))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Build the lines for findings.
	var lines []string
	lastSeverity := agent.Severity(-1)

	for _, f := range findings {
		if f.Severity != lastSeverity {
			if lastSeverity != agent.Severity(-1) {
				lines = append(lines, "")
			}
			lines = append(lines, severityHeader(f.Severity))
			lastSeverity = f.Severity
		}

		indicator := f.Severity.Indicator()
		line := fmt.Sprintf("  %s %s: %s", indicator, f.Label, f.Message)

		switch f.Severity {
		case agent.SeverityCritical:
			lines = append(lines, findingCritical.Render(line))
		case agent.SeverityWarning:
			lines = append(lines, findingWarning.Render(line))
		default:
			lines = append(lines, findingOK.Render(line))
		}

		if f.Suggestion != "" {
			lines = append(lines, dimStyle.Render(fmt.Sprintf("      Suggestion: %s", f.Suggestion)))
		}
	}

	// Apply scrolling.
	viewHeight := height - 8
	if viewHeight < 5 {
		viewHeight = 5
	}

	end := scrollOffset + viewHeight
	if end > len(lines) {
		end = len(lines)
	}
	if scrollOffset >= len(lines) {
		scrollOffset = 0
	}

	for i := scrollOffset; i < end; i++ {
		b.WriteString(lines[i])
		b.WriteString("\n")
	}

	if len(lines) > viewHeight {
		b.WriteString(dimStyle.Render(fmt.Sprintf("\n  [%d-%d of %d lines]", scrollOffset+1, end, len(lines))))
	}

	return b.String()
}

func severityHeader(s agent.Severity) string {
	switch s {
	case agent.SeverityCritical:
		return findingCritical.Render("CRITICAL")
	case agent.SeverityWarning:
		return findingWarning.Render("WARNING")
	default:
		return findingOK.Render("OK")
	}
}
