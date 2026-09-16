package tui

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"

	"mytunnel/internal/ssh"
)

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "loading…"
	}
	w := max(m.width, 40)
	h := max(m.height, 12)

	switch m.screen {
	case screenForm:
		return m.renderBox("Connection Details", m.renderFormBody(w-2, h-2), w, h)
	case screenConfirm:
		body := fmt.Sprintf("Delete “%s”?\n\nY confirm    N cancel", m.deleteName)
		return m.renderBox("Connections", padBody(body, w-2, h-4), w, h)
	default:
		return m.renderBox("Connections", m.renderListBody(w-2, h-2), w, h)
	}
}

func (m Model) renderListBody(innerW, innerH int) string {
	footer := m.renderListFooter()
	preview := m.renderPreview(innerW)
	sep := strings.Repeat("─", innerW)

	used := 1 + lipgloss.Height(preview) + 1 + lipgloss.Height(footer)
	listH := max(innerH-used, 3)

	n := 0
	if m.cfg != nil {
		n = len(m.cfg.Connections)
	}

	offset := 0
	if m.cursor >= listH {
		offset = m.cursor - listH + 1
	}

	var lines []string
	if n == 0 {
		lines = []string{mutedStyle.Render("No connections  ·  press A to add")}
	} else {
		end := min(n, offset+listH)
		for i := offset; i < end; i++ {
			title := m.cfg.Connections[i].Title
			if title == "" {
				title = m.cfg.Connections[i].Host
			}
			row := " " + title + " "
			if i == m.cursor {
				row = selectedRow.Render(padRight(row, innerW))
			} else {
				row = padRight(row, innerW)
			}
			lines = append(lines, row)
		}
	}
	for len(lines) < listH {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n") + "\n" + sep + "\n" + preview + "\n" + sep + "\n" + footer
}

func (m Model) renderPreview(width int) string {
	if m.status != "" {
		return mutedStyle.Render(truncate(m.status, width))
	}
	c := m.selected()
	if c == nil {
		return mutedStyle.Render(" ")
	}
	return truncate(ssh.CommandLine(*c), width)
}

func (m Model) renderListFooter() string {
	item := func(key, rest string) string {
		return "< " + keyStyle.Render(key) + rest + " >"
	}
	parts := []string{
		item("[A]", "dd"),
		item("C[o]", "py"),
		item("[E]", "dit"),
		item("[D]", "elete"),
		item("[C]", "onnect"),
		item("[Q]", "uit"),
	}
	return mutedStyle.Render(" " + strings.Join(parts, "   ") + " ")
}

func (m Model) renderFormBody(innerW, innerH int) string {
	d := m.form.draft
	row := func(id int, label, value string) string {
		l := fmt.Sprintf("%-18s", label)
		if m.form.focus == id {
			l = focusLabel.Render(l)
			value = selectedRow.Render(value + " ")
		} else {
			l = labelStyle.Render(l)
		}
		return l + value
	}
	check := func(id int, label, flag string, on bool) string {
		mark := "[ ]"
		if on {
			mark = "[√]"
		}
		l := fmt.Sprintf("%-22s", label)
		r := mark + " " + mutedStyle.Render(flag)
		if m.form.focus == id {
			l = focusLabel.Render(l)
			r = selectedRow.Render(mark) + " " + mutedStyle.Render(flag)
		} else {
			l = labelStyle.Render(l)
		}
		return l + r
	}
	val := func(id int) string {
		if ti, ok := m.form.inputs[id]; ok {
			return ti.View()
		}
		return ""
	}

	sep := mutedStyle.Render(strings.Repeat("─", innerW))
	ok := " < OK > "
	cancel := " < Cancel > "
	if m.form.focus == fOK {
		ok = okStyle.Render(" < OK > ")
	}
	if m.form.focus == fCancel {
		cancel = okStyle.Render(" < Cancel > ")
	}
	gap := max(innerW-lipgloss.Width(ok)-lipgloss.Width(cancel), 2)
	buttons := ok + strings.Repeat(" ", gap) + cancel

	lines := []string{
		row(fTitle, "Title:", val(fTitle)),
		row(fUser, "User:", val(fUser)),
		row(fHost, "Host:", val(fHost)),
		row(fPort, "Port:", val(fPort)),
		sep,
		check(fCompression, "Compression", "-C", d.Compression),
		check(fForwardX, "Forward X", "-X", d.ForwardX),
		check(fForwardAgent, "Forward Agent", "-A", d.ForwardAgent),
		check(fGateway, "Allow Remote Port Conn", "-g", d.GatewayPorts),
		sep,
		row(fCustom, "Custom options:", val(fCustom)),
		row(fPrepend, "Prepend command:", val(fPrepend)),
		row(fAppend, "Append command:", val(fAppend)),
		sep,
		row(fSort, "Sort order:", val(fSort)),
		sep,
		buttons,
	}
	if m.status != "" {
		lines = append(lines, "", errStyle.Render(m.status))
	}

	body := strings.Join(lines, "\n")
	return padBody(body, innerW, innerH)
}

func (m Model) renderBox(title, body string, width, height int) string {
	innerW := width - 2
	innerH := height - 2
	body = padBody(body, innerW, innerH)

	top := titledBorder(title, innerW)
	bot := "└" + strings.Repeat("─", innerW) + "┘"

	var b strings.Builder
	b.WriteString(top)
	b.WriteByte('\n')
	for _, line := range strings.Split(body, "\n") {
		b.WriteString("│")
		b.WriteString(padRight(line, innerW))
		b.WriteString("│\n")
	}
	b.WriteString(bot)
	return b.String()
}

func titledBorder(title string, innerW int) string {
	label := " " + title + " "
	if utf8.RuneCountInString(label) >= innerW {
		label = title
	}
	lw := lipgloss.Width(label)
	left := max((innerW-lw)/2, 0)
	right := max(innerW-left-lw, 0)
	return "┌" + strings.Repeat("─", left) + titleOnBorder.Render(label) + strings.Repeat("─", right) + "┐"
}

func padBody(body string, width, height int) string {
	lines := strings.Split(body, "\n")
	for i := range lines {
		lines[i] = padRight(lines[i], width)
	}
	for len(lines) < height {
		lines = append(lines, strings.Repeat(" ", width))
	}
	if len(lines) > height {
		lines = lines[:height]
	}
	return strings.Join(lines, "\n")
}

func padRight(s string, width int) string {
	w := lipgloss.Width(s)
	if w >= width {
		return truncate(s, width)
	}
	return s + strings.Repeat(" ", width-w)
}

func truncate(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}
	runes := []rune(s)
	out := string(runes)
	for lipgloss.Width(out) > width {
		runes = runes[:len(runes)-1]
		out = string(runes)
	}
	return out
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
