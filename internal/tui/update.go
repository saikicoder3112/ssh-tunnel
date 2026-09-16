package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"mytunnel/internal/ssh"
)

type execFinishedMsg struct{ err error }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case execFinishedMsg:
		if msg.err != nil {
			m.status = fmt.Sprintf("SSH ended: %v", msg.err)
		} else {
			m.status = "Disconnected"
		}
		return m, nil
	}

	switch m.screen {
	case screenForm:
		return m.updateForm(msg)
	case screenConfirm:
		return m.updateConfirm(msg)
	default:
		return m.updateList(msg)
	}
}

func (m Model) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cfg != nil && m.cursor < len(m.cfg.Connections)-1 {
			m.cursor++
		}
	case "a":
		return m.openAdd()
	case "e":
		return m.openEdit()
	case "o":
		return m.copyCmd()
	case "d":
		return m.openDelete()
	case "c", "enter":
		return m.connect()
	}
	return m, nil
}

func (m Model) updateConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "y", "enter":
		return m.deleteSelected()
	case "n", "esc", "q":
		m.screen = screenList
		m.status = "Delete cancelled"
		return m, nil
	}
	return m, nil
}

func (m Model) connect() (tea.Model, tea.Cmd) {
	c := m.selected()
	if c == nil {
		m.status = "No connection selected"
		return m, nil
	}
	if c.Host == "" {
		m.status = "Host is empty"
		return m, nil
	}
	m.status = ssh.CommandLine(*c)
	cmd := ssh.Command(*c)
	return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
		return execFinishedMsg{err}
	})
}

func (m Model) copyCmd() (tea.Model, tea.Cmd) {
	c := m.selected()
	if c == nil {
		m.status = "No connection selected"
		return m, nil
	}
	if err := ssh.CopyCommand(*c); err != nil {
		m.status = err.Error()
		return m, nil
	}
	m.status = "Copied command"
	return m, nil
}

func (m Model) openDelete() (tea.Model, tea.Cmd) {
	c := m.selected()
	if c == nil {
		m.status = "No connection to delete"
		return m, nil
	}
	m.screen = screenConfirm
	m.deleteName = c.Title
	m.deleteIdx = m.cursor
	m.status = ""
	return m, nil
}

func (m Model) deleteSelected() (tea.Model, tea.Cmd) {
	if m.deleteIdx < 0 || m.deleteIdx >= len(m.cfg.Connections) {
		m.screen = screenList
		return m, nil
	}
	name := m.cfg.Connections[m.deleteIdx].Title
	ssh.ReapConnection(m.cfg.Connections[m.deleteIdx])
	m.cfg.Connections = append(m.cfg.Connections[:m.deleteIdx], m.cfg.Connections[m.deleteIdx+1:]...)
	if err := m.cfg.Save(m.path); err != nil {
		m.status = fmt.Sprintf("Save failed: %v", err)
		m.screen = screenList
		return m, nil
	}
	m.clampCursor()
	m.screen = screenList
	m.status = "Deleted " + name
	return m, nil
}
