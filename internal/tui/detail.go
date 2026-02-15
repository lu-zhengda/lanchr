package tui

import (
	"fmt"
	"strings"

	"github.com/lu-zhengda/lanchr/internal/agent"
)

// renderDetailView renders the full info panel for a service.
func renderDetailView(svc *agent.Service, scrollOffset int, height int) string {
	var b strings.Builder

	b.WriteString(titleBar.Render(fmt.Sprintf("Service: %s", svc.Label)))
	b.WriteString("\n\n")

	rows := detailRows(svc)

	viewHeight := height - 6
	if viewHeight < 5 {
		viewHeight = 5
	}

	end := scrollOffset + viewHeight
	if end > len(rows) {
		end = len(rows)
	}
	if scrollOffset > len(rows) {
		scrollOffset = 0
	}

	for i := scrollOffset; i < end; i++ {
		b.WriteString(rows[i])
		b.WriteString("\n")
	}

	if len(rows) > viewHeight {
		b.WriteString(dimStyle.Render(fmt.Sprintf("\n  [%d-%d of %d lines]", scrollOffset+1, end, len(rows))))
	}

	return b.String()
}

// detailRows generates all the detail rows for a service.
func detailRows(svc *agent.Service) []string {
	var rows []string

	add := func(label, value string) {
		rows = append(rows, fmt.Sprintf("%s%s",
			detailLabel.Render(label+":"),
			detailValue.Render(value)))
	}

	add("Label", svc.Label)
	add("Domain", svc.Domain.String())
	add("Type", svc.Type.String())
	add("State", svc.Status.String())

	if svc.PID > 0 {
		add("PID", fmt.Sprintf("%d", svc.PID))
	} else {
		add("PID", "-")
	}

	if svc.PlistPath != "" {
		add("Plist Path", svc.PlistPath)
	} else {
		add("Plist Path", "(no plist on disk)")
	}

	if svc.Program != "" {
		add("Program", svc.Program)
	}

	if len(svc.ProgramArgs) > 0 {
		add("Arguments", strings.Join(svc.ProgramArgs, " "))
	} else {
		add("Arguments", "(none)")
	}

	add("Run At Load", fmt.Sprintf("%v", svc.RunAtLoad))

	if svc.KeepAlive != nil {
		add("Keep Alive", fmt.Sprintf("%v", svc.KeepAlive))
	} else {
		add("Keep Alive", "(none)")
	}

	if svc.StartInterval > 0 {
		add("Start Interval", fmt.Sprintf("%ds", svc.StartInterval))
	} else {
		add("Start Interval", "(none)")
	}

	if len(svc.WatchPaths) > 0 {
		add("Watch Paths", strings.Join(svc.WatchPaths, ", "))
	} else {
		add("Watch Paths", "(none)")
	}

	if svc.WorkingDirectory != "" {
		add("Working Directory", svc.WorkingDirectory)
	} else {
		add("Working Directory", "(none)")
	}

	if svc.StandardOutPath != "" {
		add("Stdout Path", svc.StandardOutPath)
	} else {
		add("Stdout Path", "(none)")
	}

	if svc.StandardErrorPath != "" {
		add("Stderr Path", svc.StandardErrorPath)
	} else {
		add("Stderr Path", "(none)")
	}

	if len(svc.EnvironmentVars) > 0 {
		var envParts []string
		for k, v := range svc.EnvironmentVars {
			envParts = append(envParts, fmt.Sprintf("%s=%s", k, v))
		}
		add("Environment", strings.Join(envParts, ", "))
	}

	if svc.ExitTimeout > 0 {
		add("Exit Timeout", fmt.Sprintf("%ds", svc.ExitTimeout))
	}

	add("Last Exit Code", fmt.Sprintf("%d", svc.LastExitStatus))
	add("Disabled", fmt.Sprintf("%v", svc.Disabled))

	if svc.BlameLine != "" {
		add("Blame", svc.BlameLine)
	}

	if svc.IsSIPProtected() {
		rows = append(rows, "")
		rows = append(rows, dimStyle.Render("  This service is SIP-protected and cannot be modified."))
	}

	return rows
}
