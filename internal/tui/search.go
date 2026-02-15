package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// SearchModel handles the search overlay / filter bar.
type SearchModel struct {
	input   textinput.Model
	active  bool
}

// NewSearchModel creates a new search model.
func NewSearchModel() SearchModel {
	ti := textinput.New()
	ti.Placeholder = "Search services..."
	ti.CharLimit = 100
	ti.Width = 40

	return SearchModel{
		input: ti,
	}
}

// Focus activates the search input.
func (s *SearchModel) Focus() {
	s.active = true
	s.input.Focus()
}

// Blur deactivates the search input.
func (s *SearchModel) Blur() {
	s.active = false
	s.input.Blur()
}

// Reset clears the search query and deactivates.
func (s *SearchModel) Reset() {
	s.input.SetValue("")
	s.Blur()
}

// Query returns the current search query.
func (s *SearchModel) Query() string {
	return s.input.Value()
}

// SetQuery sets the search query.
func (s *SearchModel) SetQuery(q string) {
	s.input.SetValue(q)
}

// IsActive returns whether the search input is focused.
func (s *SearchModel) IsActive() bool {
	return s.active
}

// Update handles key events for the search input.
func (s *SearchModel) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)
	return cmd
}

// View renders the search bar.
func (s *SearchModel) View() string {
	return searchStyle.Render("/ " + s.input.View())
}

// MatchesService returns true if the service label or program matches the query.
func MatchesService(query, label, program string) bool {
	if query == "" {
		return true
	}
	q := strings.ToLower(query)
	return strings.Contains(strings.ToLower(label), q) ||
		strings.Contains(strings.ToLower(program), q)
}
