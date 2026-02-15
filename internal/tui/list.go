package tui

import (
	"fmt"
	"strings"

	"github.com/zhengda-lu/lanchr/internal/agent"
	"github.com/zhengda-lu/lanchr/internal/platform"
)

// domainGroup represents a group of services under a domain header.
type domainGroup struct {
	name     string
	path     string
	domain   platform.Domain
	svcType  platform.ServiceType
	services []int // indices into the filtered services slice
}

// groupServices organizes services into domain groups for display.
func groupServices(services []agent.Service) []domainGroup {
	groups := []domainGroup{
		{name: "USER AGENTS", path: "~/Library/LaunchAgents", domain: platform.DomainUser, svcType: platform.TypeAgent},
		{name: "GLOBAL AGENTS", path: "/Library/LaunchAgents", domain: platform.DomainGlobal, svcType: platform.TypeAgent},
		{name: "GLOBAL DAEMONS", path: "/Library/LaunchDaemons", domain: platform.DomainGlobal, svcType: platform.TypeDaemon},
		{name: "SYSTEM AGENTS", path: "/System/Library/LaunchAgents", domain: platform.DomainSystem, svcType: platform.TypeAgent},
		{name: "SYSTEM DAEMONS", path: "/System/Library/LaunchDaemons", domain: platform.DomainSystem, svcType: platform.TypeDaemon},
	}

	for i, svc := range services {
		for g := range groups {
			if svc.Domain == groups[g].domain && svc.Type == groups[g].svcType {
				groups[g].services = append(groups[g].services, i)
				break
			}
		}
	}

	return groups
}

// renderListView renders the service list grouped by domain.
func renderListView(services []agent.Service, groups []domainGroup, cursor int, scrollOffset int, width int, height int) string {
	var b strings.Builder

	// Build the flat list of renderable lines, tracking which line the cursor is on.
	type listLine struct {
		isHeader    bool
		headerText  string
		serviceIdx  int
		displayLine string
	}

	var lines []listLine
	flatIdx := 0

	for _, group := range groups {
		// Add domain header.
		header := fmt.Sprintf("%s (%s)%s",
			domainHeader.Render(group.name),
			group.path,
			countBadge.Render(fmt.Sprintf("  %d services", len(group.services))))

		lines = append(lines, listLine{isHeader: true, headerText: header})

		if len(group.services) == 0 {
			lines = append(lines, listLine{isHeader: true, headerText: dimStyle.Render("  (none)")})
			continue
		}

		for _, svcIdx := range group.services {
			svc := services[svcIdx]
			isCurrent := flatIdx == cursor

			line := renderServiceLine(svc, isCurrent, width)
			lines = append(lines, listLine{serviceIdx: svcIdx, displayLine: line})
			flatIdx++
		}
	}

	// Apply scrolling.
	viewHeight := height - 4 // Reserve for title bar, filter bar, and help bar.
	if viewHeight < 5 {
		viewHeight = 5
	}

	// Find the line that contains the cursor to ensure it is visible.
	cursorLineIdx := 0
	svcCount := 0
	for i, line := range lines {
		if !line.isHeader {
			if svcCount == cursor {
				cursorLineIdx = i
				break
			}
			svcCount++
		}
	}

	// Adjust scroll offset.
	if cursorLineIdx < scrollOffset {
		scrollOffset = cursorLineIdx
	}
	if cursorLineIdx >= scrollOffset+viewHeight {
		scrollOffset = cursorLineIdx - viewHeight + 1
	}
	if scrollOffset < 0 {
		scrollOffset = 0
	}

	end := scrollOffset + viewHeight
	if end > len(lines) {
		end = len(lines)
	}

	for i := scrollOffset; i < end; i++ {
		line := lines[i]
		if line.isHeader {
			b.WriteString(line.headerText)
		} else {
			b.WriteString(line.displayLine)
		}
		b.WriteString("\n")
	}

	// Scroll indicator.
	if len(lines) > viewHeight {
		b.WriteString(dimStyle.Render(fmt.Sprintf("  [showing %d-%d of %d lines]", scrollOffset+1, end, len(lines))))
		b.WriteString("\n")
	}

	return b.String()
}

// renderServiceLine renders a single service line.
func renderServiceLine(svc agent.Service, isCurrent bool, width int) string {
	// Status indicator.
	var indicator string
	switch svc.Status {
	case agent.StatusRunning:
		indicator = statusRunning.Render("*")
	case agent.StatusStopped:
		indicator = statusStopped.Render("-")
	case agent.StatusError:
		indicator = statusError.Render("!")
	case agent.StatusDisabled:
		indicator = statusDisabled.Render("x")
	default:
		indicator = statusStopped.Render("?")
	}

	// PID.
	pidStr := "-"
	if svc.PID > 0 {
		pidStr = fmt.Sprintf("%d", svc.PID)
	}

	// Binary path (truncated).
	binary := svc.BinaryPath()
	maxBinaryLen := width - 60
	if maxBinaryLen < 20 {
		maxBinaryLen = 20
	}
	if len(binary) > maxBinaryLen {
		binary = binary[:maxBinaryLen-3] + "..."
	}

	cursor := "  "
	if isCurrent {
		cursor = "> "
	}

	line := fmt.Sprintf("%s %s %6s  %-40s  %s",
		cursor, indicator, pidStr, truncateStr(svc.Label, 40), binary)

	if isCurrent {
		return selectedStyle.Render(line)
	}
	return line
}

// truncateStr truncates a string to maxLen with "..." suffix if needed.
func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen < 4 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
