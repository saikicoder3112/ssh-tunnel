package tui

import "github.com/charmbracelet/lipgloss"

var (
	fg       = lipgloss.Color("252")
	muted    = lipgloss.Color("244")
	borderFg = lipgloss.Color("245")
	accent   = lipgloss.Color("15")

	titleOnBorder = lipgloss.NewStyle().Bold(true).Foreground(accent)
	mutedStyle    = lipgloss.NewStyle().Foreground(muted)
	selectedRow   = lipgloss.NewStyle().Reverse(true)
	labelStyle    = lipgloss.NewStyle().Foreground(fg)
	focusLabel    = lipgloss.NewStyle().Bold(true).Foreground(accent)
	okStyle       = lipgloss.NewStyle().Reverse(true).Bold(true)
	errStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	keyStyle      = lipgloss.NewStyle().Bold(true).Foreground(accent)
)
