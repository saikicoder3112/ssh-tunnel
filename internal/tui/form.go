package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"mytunnel/internal/config"
)

const (
	fTitle = iota
	fUser
	fHost
	fPort
	fCompression
	fForwardX
	fForwardAgent
	fGateway
	fCustom
	fPrepend
	fAppend
	fSort
	fOK
	fCancel
	fCount
)

type formModel struct {
	draft  config.Connection
	inputs map[int]textinput.Model
	focus  int
	isNew  bool
}

func newForm(c config.Connection, isNew bool) formModel {
	f := formModel{
		draft:  c,
		inputs: make(map[int]textinput.Model),
		focus:  fTitle,
		isNew:  isNew,
	}
	f.inputs[fTitle] = newInput(c.Title, "demo-ssh")
	f.inputs[fUser] = newInput(c.User, "root")
	f.inputs[fHost] = newInput(c.Host, "example.com")
	port := ""
	if c.Port > 0 {
		port = strconv.Itoa(c.Port)
	} else {
		port = "22"
	}
	f.inputs[fPort] = newInput(port, "22")
	f.inputs[fCustom] = newInput(c.Custom, "-L 3307:127.0.0.1:3306")
	f.inputs[fPrepend] = newInput(c.Prepend, "")
	f.inputs[fAppend] = newInput(c.Append, "")
	sortVal := "0"
	if c.Sort != 0 {
		sortVal = strconv.Itoa(c.Sort)
	}
	f.inputs[fSort] = newInput(sortVal, "0")
	f.syncFocus()
	return f
}

func newInput(value, placeholder string) textinput.Model {
	ti := textinput.New()
	ti.SetValue(value)
	ti.Placeholder = placeholder
	ti.CharLimit = 256
	ti.Width = 42
	ti.Prompt = ""
	ti.PlaceholderStyle = mutedStyle
	return ti
}

func (f *formModel) syncFocus() {
	for i, ti := range f.inputs {
		if i == f.focus {
			ti.Focus()
			ti.TextStyle = focusLabel
			f.inputs[i] = ti
		} else {
			ti.Blur()
			ti.TextStyle = labelStyle
			f.inputs[i] = ti
		}
	}
}

func (f *formModel) next() {
	f.focus = (f.focus + 1) % fCount
	f.syncFocus()
}

func (f *formModel) prev() {
	f.focus--
	if f.focus < 0 {
		f.focus = fCount - 1
	}
	f.syncFocus()
}

func (m Model) updateForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.screen = screenList
			m.status = "Cancelled"
			return m, nil
		case "tab", "down":
			m.form.next()
			return m, textinput.Blink
		case "shift+tab", "up":
			m.form.prev()
			return m, textinput.Blink
		case " ":
			if m.toggleCheck() {
				return m, nil
			}
		case "enter":
			switch m.form.focus {
			case fOK:
				return m.saveForm()
			case fCancel:
				m.screen = screenList
				m.status = "Cancelled"
				return m, nil
			case fCompression, fForwardX, fForwardAgent, fGateway:
				m.toggleCheck()
				return m, nil
			default:
				m.form.next()
				return m, textinput.Blink
			}
		}
	}

	if ti, ok := m.form.inputs[m.form.focus]; ok {
		var cmd tea.Cmd
		ti, cmd = ti.Update(msg)
		m.form.inputs[m.form.focus] = ti
		return m, cmd
	}
	return m, nil
}

func (m *Model) toggleCheck() bool {
	switch m.form.focus {
	case fCompression:
		m.form.draft.Compression = !m.form.draft.Compression
	case fForwardX:
		m.form.draft.ForwardX = !m.form.draft.ForwardX
	case fForwardAgent:
		m.form.draft.ForwardAgent = !m.form.draft.ForwardAgent
	case fGateway:
		m.form.draft.GatewayPorts = !m.form.draft.GatewayPorts
	default:
		return false
	}
	return true
}

func (m Model) saveForm() (tea.Model, tea.Cmd) {
	d := m.form.draft
	d.Title = strings.TrimSpace(m.form.inputs[fTitle].Value())
	d.User = strings.TrimSpace(m.form.inputs[fUser].Value())
	d.Host = strings.TrimSpace(m.form.inputs[fHost].Value())
	d.Custom = strings.TrimSpace(m.form.inputs[fCustom].Value())
	d.Prepend = strings.TrimSpace(m.form.inputs[fPrepend].Value())
	d.Append = strings.TrimSpace(m.form.inputs[fAppend].Value())

	if d.Title == "" || d.Host == "" {
		m.status = "Title and host are required"
		return m, nil
	}

	portStr := strings.TrimSpace(m.form.inputs[fPort].Value())
	if portStr == "" {
		d.Port = 0
	} else {
		p, err := strconv.Atoi(portStr)
		if err != nil || p < 1 || p > 65535 {
			m.status = "Port must be 1–65535"
			return m, nil
		}
		d.Port = p
	}

	sortStr := strings.TrimSpace(m.form.inputs[fSort].Value())
	if sortStr == "" {
		d.Sort = 0
	} else {
		s, err := strconv.Atoi(sortStr)
		if err != nil {
			m.status = "Sort order must be a number"
			return m, nil
		}
		d.Sort = s
	}

	if m.form.isNew {
		m.cfg.Connections = append(m.cfg.Connections, d)
	} else if m.editIndex >= 0 && m.editIndex < len(m.cfg.Connections) {
		m.cfg.Connections[m.editIndex] = d
	}

	if err := m.cfg.Save(m.path); err != nil {
		m.status = fmt.Sprintf("Save failed: %v", err)
		return m, nil
	}

	m.cfg.Sort()
	for i, c := range m.cfg.Connections {
		if c.Title == d.Title && c.Host == d.Host {
			m.cursor = i
			break
		}
	}
	m.screen = screenList
	if m.form.isNew {
		m.status = "Added " + d.Title
	} else {
		m.status = "Updated " + d.Title
	}
	return m, nil
}

func (m Model) openAdd() (tea.Model, tea.Cmd) {
	m.screen = screenForm
	m.editIndex = -1
	m.status = ""
	m.form = newForm(config.Connection{Port: 22}, true)
	return m, textinput.Blink
}

func (m Model) openEdit() (tea.Model, tea.Cmd) {
	c := m.selected()
	if c == nil {
		m.status = "No connection selected"
		return m, nil
	}
	m.screen = screenForm
	m.editIndex = m.cursor
	m.status = ""
	m.form = newForm(*c, false)
	return m, textinput.Blink
}
