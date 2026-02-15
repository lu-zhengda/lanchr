package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Status indicators.
	statusRunning = lipgloss.NewStyle().Foreground(lipgloss.Color("42")) // green
	statusStopped = lipgloss.NewStyle().Foreground(lipgloss.Color("245")) // gray
	statusError   = lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // red
	statusDisabled = lipgloss.NewStyle().Foreground(lipgloss.Color("208")) // orange

	// Domain headers.
	domainHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("99")).
			MarginTop(1)

	// Selected row.
	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Bold(true)

	// Count badge.
	countBadge = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Italic(true)

	// Help bar at bottom.
	helpBar = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Background(lipgloss.Color("235")).
		Padding(0, 1)

	// Title bar.
	titleBar = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("255")).
			Background(lipgloss.Color("62")).
			Padding(0, 1)

	// Doctor severity colors.
	findingCritical = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	findingWarning  = lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
	findingOK       = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))

	// General dim text.
	dimStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	// Error message style.
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)

	// Success message style.
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)

	// Search input style.
	searchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Background(lipgloss.Color("236")).
			Padding(0, 1)

	// Detail label style.
	detailLabel = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Width(22)

	// Detail value style.
	detailValue = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))
)
