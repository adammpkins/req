// Package tui provides an interactive terminal user interface for req.
package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/adammpkins/req/internal/tui/views"
)

// Launch starts the TUI application.
func Launch() error {
	p := tea.NewProgram(NewModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}
	return nil
}

// Model represents the application state.
type Model struct {
	view View
}

// NewModel creates a new TUI model.
func NewModel() Model {
	return Model{
		view: views.NewBuilderView(),
	}
}

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	// WindowSizeMsg will be sent automatically by bubbletea
	return m.view.Init()
}

// Update handles messages and updates the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		// Pass window size to view
		var cmd tea.Cmd
		m.view, cmd = m.view.Update(msg)
		return m, cmd
	}

	var cmd tea.Cmd
	m.view, cmd = m.view.Update(msg)
	return m, cmd
}

// View renders the current view.
func (m Model) View() string {
	return m.view.View()
}

// View represents a TUI view (exported from views package).
type View = views.View

