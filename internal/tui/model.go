package tui

import (
	"mytunnel/internal/config"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type screen int

const (
	screenList screen = iota
	screenForm
	screenConfirm
)

type Model struct {
	cfg    *config.Config
	path   string
	cursor int
	width  int
	height int
	status string
	screen screen

	form       formModel
	deleteName string
	deleteIdx  int
	editIndex  int
}

func New(cfg *config.Config, path string) Model {
	if cfg == nil {
		cfg = &config.Config{}
	}
	return Model{
		cfg:       cfg,
		path:      path,
		editIndex: -1,
		status:    "",
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) selected() *config.Connection {
	if m.cfg == nil || m.cursor < 0 || m.cursor >= len(m.cfg.Connections) {
		return nil
	}
	return &m.cfg.Connections[m.cursor]
}

func (m *Model) clampCursor() {
	n := 0
	if m.cfg != nil {
		n = len(m.cfg.Connections)
	}
	if n == 0 {
		m.cursor = 0
		return
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= n {
		m.cursor = n - 1
	}
}
