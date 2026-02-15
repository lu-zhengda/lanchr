package tui

import (
	"fmt"
	"strings"

	"github.com/zhengda-lu/lanchr/internal/agent"
)

// renderLogView renders the inline log viewer for a service.
func renderLogView(svc *agent.Service, logLines []string, scrollOffset int, height int) string {
	var b strings.Builder

	b.WriteString(titleBar.Render(fmt.Sprintf("Logs: %s", svc.Label)))
	b.WriteString("\n")

	if svc.StandardOutPath != "" {
		b.WriteString(dimStyle.Render(fmt.Sprintf("  stdout: %s", svc.StandardOutPath)))
		b.WriteString("\n")
	}
	if svc.StandardErrorPath != "" {
		b.WriteString(dimStyle.Render(fmt.Sprintf("  stderr: %s", svc.StandardErrorPath)))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	if len(logLines) == 0 {
		b.WriteString(dimStyle.Render("  No log output available."))
		b.WriteString("\n")
		return b.String()
	}

	viewHeight := height - 8
	if viewHeight < 5 {
		viewHeight = 5
	}

	end := scrollOffset + viewHeight
	if end > len(logLines) {
		end = len(logLines)
	}
	if scrollOffset >= len(logLines) {
		scrollOffset = 0
	}

	for i := scrollOffset; i < end; i++ {
		b.WriteString("  ")
		b.WriteString(logLines[i])
		b.WriteString("\n")
	}

	if len(logLines) > viewHeight {
		b.WriteString(dimStyle.Render(fmt.Sprintf("\n  [%d-%d of %d lines]", scrollOffset+1, end, len(logLines))))
	}

	return b.String()
}
