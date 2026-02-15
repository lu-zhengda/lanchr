package tui

import "strings"

// helpText returns the full help text for the TUI.
func helpText() string {
	var b strings.Builder

	b.WriteString("NAVIGATION\n")
	b.WriteString("  j / Down      Move cursor down\n")
	b.WriteString("  k / Up        Move cursor up\n")
	b.WriteString("  g             Jump to top\n")
	b.WriteString("  G             Jump to bottom\n")
	b.WriteString("  Tab           Cycle through domain groups\n")
	b.WriteString("  Enter / i     Open detail view for selected service\n")
	b.WriteString("  Esc           Go back to list view\n")
	b.WriteString("\n")
	b.WriteString("ACTIONS\n")
	b.WriteString("  e             Enable selected service\n")
	b.WriteString("  x             Disable selected service\n")
	b.WriteString("  r             Restart selected service (kickstart -k)\n")
	b.WriteString("  l             View logs for selected service\n")
	b.WriteString("  L             Load a plist (prompts for path)\n")
	b.WriteString("  U             Unload selected service\n")
	b.WriteString("\n")
	b.WriteString("FILTERING\n")
	b.WriteString("  /             Open search bar\n")
	b.WriteString("  d             Cycle domain filter: all -> user -> global -> system\n")
	b.WriteString("  s             Cycle status filter: all -> running -> stopped -> error\n")
	b.WriteString("  a             Toggle show/hide Apple services (com.apple.*)\n")
	b.WriteString("\n")
	b.WriteString("VIEWS\n")
	b.WriteString("  D             Run doctor and show report\n")
	b.WriteString("  R             Refresh service list\n")
	b.WriteString("  ?             Show/hide this help\n")
	b.WriteString("  q / Ctrl+C    Quit\n")

	return b.String()
}

// shortHelp returns the single-line help bar for the bottom of the screen.
func shortHelp(mode viewMode) string {
	switch mode {
	case viewDetail:
		return "esc:back | j/k:scroll | e:enable | x:disable | r:restart | l:logs"
	case viewLogs:
		return "esc:back | j/k:scroll"
	case viewDoctor:
		return "esc:back | j/k:scroll"
	default:
		return "e:enable | x:disable | r:restart | l:logs | D:doctor | /:search | ?:help | q:quit"
	}
}
