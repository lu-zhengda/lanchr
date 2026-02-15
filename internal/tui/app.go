package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zhengda-lu/lanchr/internal/agent"
	"github.com/zhengda-lu/lanchr/internal/logs"
	"github.com/zhengda-lu/lanchr/internal/platform"
)

// viewMode represents which view is currently active.
type viewMode int

const (
	viewList   viewMode = iota
	viewDetail
	viewLogs
	viewDoctor
)

// Messages sent by background operations.
type servicesScanResult struct {
	services []agent.Service
	err      error
}

type doctorResult struct {
	findings []agent.Finding
	err      error
}

type logLinesResult struct {
	lines []string
	err   error
}

type actionResult struct {
	msg string
	err error
}

// Model is the main Bubbletea model for the lanchr TUI.
type Model struct {
	// Core dependencies.
	scanner *agent.Scanner
	manager *agent.Manager
	doctor  *agent.Doctor
	version string

	// Data.
	services []agent.Service
	filtered []agent.Service
	groups   []domainGroup

	// View state.
	mode         viewMode
	cursor       int
	scrollOffset int
	width        int
	height       int

	// Detail view state.
	detailService *agent.Service
	detailScroll  int

	// Log view state.
	logService *agent.Service
	logLines   []string
	logScroll  int

	// Doctor view state.
	findings    []agent.Finding
	doctorScroll int

	// Filters.
	domainFilter *platform.Domain
	statusFilter *agent.Status
	hideApple    bool
	searchModel  SearchModel

	// State.
	showSearch bool
	showHelp   bool
	loading    bool
	statusMsg  string
	err        error

	// Spinner.
	spinner spinner.Model
}

// New creates a new TUI model.
func New(scanner *agent.Scanner, manager *agent.Manager, doctor *agent.Doctor, version string) Model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("99"))

	return Model{
		scanner:     scanner,
		manager:     manager,
		doctor:      doctor,
		version:     version,
		loading:     true,
		searchModel: NewSearchModel(),
		spinner:     sp,
	}
}

// Init starts the spinner and kicks off the initial service scan.
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.scanServices())
}

// scanServices returns a command that scans all services in the background.
func (m Model) scanServices() tea.Cmd {
	return func() tea.Msg {
		services, err := m.scanner.ScanAll()
		return servicesScanResult{services: services, err: err}
	}
}

// runDoctor returns a command that runs the doctor check in the background.
func (m Model) runDoctor() tea.Cmd {
	return func() tea.Msg {
		findings, err := m.doctor.Check()
		return doctorResult{findings: findings, err: err}
	}
}

// loadLogs returns a command that loads log lines for a service.
func (m Model) loadLogs(svc *agent.Service) tea.Cmd {
	return func() tea.Msg {
		tailer := logs.NewTailer()
		unified := logs.NewUnifiedLog()

		var allLines []string

		// Try stdout file first.
		if svc.StandardOutPath != "" {
			lines, err := tailer.Tail(svc.StandardOutPath, 100)
			if err == nil {
				allLines = append(allLines, lines...)
			}
		}

		// Try stderr file.
		if svc.StandardErrorPath != "" {
			lines, err := tailer.Tail(svc.StandardErrorPath, 100)
			if err == nil {
				allLines = append(allLines, lines...)
			}
		}

		// Fall back to unified logging if no file logs.
		if len(allLines) == 0 {
			binary := svc.BinaryPath()
			if binary != "" {
				processName := filepath.Base(binary)
				lines, err := unified.Show(processName, time.Hour, 50)
				if err == nil {
					allLines = lines
				}
			}
		}

		return logLinesResult{lines: allLines}
	}
}

// performAction returns a command that performs a service action.
func (m Model) performAction(action string, label string) tea.Cmd {
	return func() tea.Msg {
		var err error
		var msg string

		switch action {
		case "enable":
			err = m.manager.Enable(label)
			msg = fmt.Sprintf("Enabled %s", label)
		case "disable":
			err = m.manager.Disable(label)
			msg = fmt.Sprintf("Disabled %s", label)
		case "restart":
			err = m.manager.Restart(label)
			msg = fmt.Sprintf("Restarted %s", label)
		case "unload":
			err = m.manager.Unload(label)
			msg = fmt.Sprintf("Unloaded %s", label)
		}

		return actionResult{msg: msg, err: err}
	}
}

// Update handles all messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil

	case servicesScanResult:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			m.statusMsg = fmt.Sprintf("Error: %v", msg.err)
			return m, nil
		}
		m.services = msg.services
		m.applyFilters()
		m.statusMsg = fmt.Sprintf("Loaded %d services", len(m.services))
		return m, nil

	case doctorResult:
		m.loading = false
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("Doctor error: %v", msg.err)
			return m, nil
		}
		m.findings = msg.findings
		m.doctorScroll = 0
		m.mode = viewDoctor
		return m, nil

	case logLinesResult:
		m.loading = false
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("Log error: %v", msg.err)
			return m, nil
		}
		m.logLines = msg.lines
		m.logScroll = 0
		m.mode = viewLogs
		return m, nil

	case actionResult:
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("Error: %v", msg.err)
		} else {
			m.statusMsg = msg.msg
		}
		// Rescan after action.
		m.loading = true
		return m, tea.Batch(m.scanServices(), m.spinner.Tick)

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

// handleKey processes keyboard input.
func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Global quit.
	if key == keyCtrlC {
		return m, tea.Quit
	}

	// Help overlay.
	if m.showHelp {
		if key == keyHelp || key == keyEsc || key == keyQuit {
			m.showHelp = false
		}
		return m, nil
	}

	// Search input active.
	if m.showSearch {
		switch key {
		case keyEsc:
			m.showSearch = false
			m.searchModel.Blur()
			m.applyFilters()
			return m, nil
		case keyEnter:
			m.showSearch = false
			m.searchModel.Blur()
			m.applyFilters()
			return m, nil
		default:
			cmd := m.searchModel.Update(msg)
			m.applyFilters()
			return m, cmd
		}
	}

	// Mode-specific handling.
	switch m.mode {
	case viewList:
		return m.handleListKey(key)
	case viewDetail:
		return m.handleDetailKey(key)
	case viewLogs:
		return m.handleLogKey(key)
	case viewDoctor:
		return m.handleDoctorKey(key)
	}

	return m, nil
}

// handleListKey handles key events in the list view.
func (m Model) handleListKey(key string) (tea.Model, tea.Cmd) {
	totalServices := len(m.filtered)

	switch key {
	case keyQuit:
		return m, tea.Quit
	case keyJ, keyDown:
		if m.cursor < totalServices-1 {
			m.cursor++
		}
	case keyK, keyUp:
		if m.cursor > 0 {
			m.cursor--
		}
	case keyG:
		m.cursor = 0
		m.scrollOffset = 0
	case keyShiftG:
		m.cursor = totalServices - 1
	case keyTab:
		m.jumpToNextGroup()
	case keyEnter:
		return m.openDetail()
	case keyEnable:
		return m.enableSelected()
	case keyDisable:
		return m.disableSelected()
	case keyRestart:
		return m.restartSelected()
	case keyLogs:
		return m.openLogs()
	case keyUnload:
		return m.unloadSelected()
	case keySearch:
		m.showSearch = true
		m.searchModel.Focus()
		return m, nil
	case keyDomain:
		m.cycleDomainFilter()
		m.applyFilters()
	case keyStatus:
		m.cycleStatusFilter()
		m.applyFilters()
	case keyApple:
		m.hideApple = !m.hideApple
		m.applyFilters()
		if m.hideApple {
			m.statusMsg = "Hiding Apple services"
		} else {
			m.statusMsg = "Showing all services"
		}
	case keyDoctor:
		m.loading = true
		return m, tea.Batch(m.runDoctor(), m.spinner.Tick)
	case keyHelp:
		m.showHelp = true
	case keyRefresh:
		m.loading = true
		m.statusMsg = "Refreshing..."
		return m, tea.Batch(m.scanServices(), m.spinner.Tick)
	}

	return m, nil
}

// handleDetailKey handles key events in the detail view.
func (m Model) handleDetailKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case keyEsc, keyBackspace:
		m.mode = viewList
		m.detailService = nil
	case keyJ, keyDown:
		m.detailScroll++
	case keyK, keyUp:
		if m.detailScroll > 0 {
			m.detailScroll--
		}
	case keyEnable:
		if m.detailService != nil {
			return m, m.performAction("enable", m.detailService.Label)
		}
	case keyDisable:
		if m.detailService != nil {
			return m, m.performAction("disable", m.detailService.Label)
		}
	case keyRestart:
		if m.detailService != nil {
			return m, m.performAction("restart", m.detailService.Label)
		}
	case keyLogs:
		if m.detailService != nil {
			return m.openLogsForService(m.detailService)
		}
	case keyQuit:
		return m, tea.Quit
	}
	return m, nil
}

// handleLogKey handles key events in the log view.
func (m Model) handleLogKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case keyEsc, keyBackspace:
		m.mode = viewList
		m.logService = nil
		m.logLines = nil
	case keyJ, keyDown:
		m.logScroll++
	case keyK, keyUp:
		if m.logScroll > 0 {
			m.logScroll--
		}
	case keyQuit:
		return m, tea.Quit
	}
	return m, nil
}

// handleDoctorKey handles key events in the doctor view.
func (m Model) handleDoctorKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case keyEsc, keyBackspace:
		m.mode = viewList
	case keyJ, keyDown:
		m.doctorScroll++
	case keyK, keyUp:
		if m.doctorScroll > 0 {
			m.doctorScroll--
		}
	case keyQuit:
		return m, tea.Quit
	}
	return m, nil
}

// openDetail opens the detail view for the selected service.
func (m Model) openDetail() (tea.Model, tea.Cmd) {
	if m.cursor >= len(m.filtered) {
		return m, nil
	}

	svc := m.filtered[m.cursor]

	// Try to get enriched info.
	enriched, err := m.manager.Info(svc.Label)
	if err == nil {
		m.detailService = enriched
	} else {
		m.detailService = &svc
	}

	m.detailScroll = 0
	m.mode = viewDetail
	return m, nil
}

// openLogs opens the log view for the selected service.
func (m Model) openLogs() (tea.Model, tea.Cmd) {
	if m.cursor >= len(m.filtered) {
		return m, nil
	}
	svc := m.filtered[m.cursor]
	return m.openLogsForService(&svc)
}

// openLogsForService opens the log view for a specific service.
func (m Model) openLogsForService(svc *agent.Service) (tea.Model, tea.Cmd) {
	m.logService = svc
	m.loading = true
	return m, tea.Batch(m.loadLogs(svc), m.spinner.Tick)
}

// enableSelected enables the currently selected service.
func (m Model) enableSelected() (tea.Model, tea.Cmd) {
	if m.cursor >= len(m.filtered) {
		return m, nil
	}
	svc := m.filtered[m.cursor]
	return m, m.performAction("enable", svc.Label)
}

// disableSelected disables the currently selected service.
func (m Model) disableSelected() (tea.Model, tea.Cmd) {
	if m.cursor >= len(m.filtered) {
		return m, nil
	}
	svc := m.filtered[m.cursor]
	return m, m.performAction("disable", svc.Label)
}

// restartSelected restarts the currently selected service.
func (m Model) restartSelected() (tea.Model, tea.Cmd) {
	if m.cursor >= len(m.filtered) {
		return m, nil
	}
	svc := m.filtered[m.cursor]
	return m, m.performAction("restart", svc.Label)
}

// unloadSelected unloads the currently selected service.
func (m Model) unloadSelected() (tea.Model, tea.Cmd) {
	if m.cursor >= len(m.filtered) {
		return m, nil
	}
	svc := m.filtered[m.cursor]
	return m, m.performAction("unload", svc.Label)
}

// jumpToNextGroup moves the cursor to the first service of the next domain group.
func (m *Model) jumpToNextGroup() {
	if len(m.groups) == 0 || len(m.filtered) == 0 {
		return
	}

	// Find which group the cursor is currently in.
	currentIdx := 0
	offset := 0
	for i, group := range m.groups {
		if m.cursor < offset+len(group.services) {
			currentIdx = i
			break
		}
		offset += len(group.services)
	}

	// Move to the next group that has services.
	for i := 1; i <= len(m.groups); i++ {
		nextIdx := (currentIdx + i) % len(m.groups)
		if len(m.groups[nextIdx].services) > 0 {
			offset := 0
			for j := 0; j < nextIdx; j++ {
				offset += len(m.groups[j].services)
			}
			m.cursor = offset
			return
		}
	}
}

// applyFilters rebuilds the filtered service list based on current filters.
func (m *Model) applyFilters() {
	m.filtered = nil

	for _, svc := range m.services {
		// Domain filter.
		if m.domainFilter != nil && svc.Domain != *m.domainFilter {
			continue
		}

		// Status filter.
		if m.statusFilter != nil && svc.Status != *m.statusFilter {
			continue
		}

		// Hide Apple services.
		if m.hideApple && svc.IsApple() {
			continue
		}

		// Search query.
		if m.searchModel.Query() != "" {
			if !MatchesService(m.searchModel.Query(), svc.Label, svc.BinaryPath()) {
				continue
			}
		}

		m.filtered = append(m.filtered, svc)
	}

	m.groups = groupServices(m.filtered)

	// Clamp cursor.
	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
}

// cycleDomainFilter cycles through domain filter options: all -> user -> global -> system.
func (m *Model) cycleDomainFilter() {
	if m.domainFilter == nil {
		d := platform.DomainUser
		m.domainFilter = &d
		m.statusMsg = "Filter: user"
	} else {
		switch *m.domainFilter {
		case platform.DomainUser:
			d := platform.DomainGlobal
			m.domainFilter = &d
			m.statusMsg = "Filter: global"
		case platform.DomainGlobal:
			d := platform.DomainSystem
			m.domainFilter = &d
			m.statusMsg = "Filter: system"
		case platform.DomainSystem:
			m.domainFilter = nil
			m.statusMsg = "Filter: all domains"
		}
	}
}

// cycleStatusFilter cycles through status filter options: all -> running -> stopped -> error.
func (m *Model) cycleStatusFilter() {
	if m.statusFilter == nil {
		s := agent.StatusRunning
		m.statusFilter = &s
		m.statusMsg = "Filter: running"
	} else {
		switch *m.statusFilter {
		case agent.StatusRunning:
			s := agent.StatusStopped
			m.statusFilter = &s
			m.statusMsg = "Filter: stopped"
		case agent.StatusStopped:
			s := agent.StatusError
			m.statusFilter = &s
			m.statusMsg = "Filter: error"
		case agent.StatusError:
			m.statusFilter = nil
			m.statusMsg = "Filter: all statuses"
		}
	}
}

// View renders the TUI.
func (m Model) View() string {
	if m.loading {
		return titleBar.Render(fmt.Sprintf("lanchr %s", m.version)) + "\n\n" +
			m.spinner.View() + " Loading services...\n"
	}

	var b strings.Builder

	// Title bar.
	title := fmt.Sprintf("lanchr %s", m.version)
	if m.showHelp {
		title += "  [?] help"
	}
	b.WriteString(titleBar.Width(m.width).Render(title))
	b.WriteString("\n")

	// Filter bar.
	filterParts := []string{}
	if m.showSearch {
		filterParts = append(filterParts, m.searchModel.View())
	} else {
		filterParts = append(filterParts, "[/] Search")
	}
	filterParts = append(filterParts, "[Tab] Switch pane")
	if m.domainFilter != nil {
		filterParts = append(filterParts, fmt.Sprintf("[d] %s", m.domainFilter.String()))
	} else {
		filterParts = append(filterParts, "[d] Domain")
	}
	if m.statusFilter != nil {
		filterParts = append(filterParts, fmt.Sprintf("[s] %s", m.statusFilter.String()))
	} else {
		filterParts = append(filterParts, "[s] Status")
	}
	if m.hideApple {
		filterParts = append(filterParts, "[a] hiding apple")
	} else {
		filterParts = append(filterParts, "[a] Apple")
	}
	b.WriteString(dimStyle.Render(strings.Join(filterParts, "  ")))
	b.WriteString("\n")

	// Help overlay.
	if m.showHelp {
		b.WriteString("\n")
		b.WriteString(helpText())
		b.WriteString("\n")
		b.WriteString(helpBar.Width(m.width).Render("Press ? or Esc to close help"))
		return b.String()
	}

	// Main content.
	switch m.mode {
	case viewList:
		b.WriteString(renderListView(m.filtered, m.groups, m.cursor, m.scrollOffset, m.width, m.height))
	case viewDetail:
		if m.detailService != nil {
			b.WriteString(renderDetailView(m.detailService, m.detailScroll, m.height))
		}
	case viewLogs:
		if m.logService != nil {
			b.WriteString(renderLogView(m.logService, m.logLines, m.logScroll, m.height))
		}
	case viewDoctor:
		b.WriteString(renderDoctorView(m.findings, m.doctorScroll, m.height))
	}

	// Status message.
	if m.statusMsg != "" {
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("  " + m.statusMsg))
	}

	// Help bar.
	b.WriteString("\n")
	b.WriteString(helpBar.Width(m.width).Render(shortHelp(m.mode)))

	return b.String()
}
