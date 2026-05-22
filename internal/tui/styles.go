package tui

import "github.com/charmbracelet/lipgloss"

var (
	AppStyle = lipgloss.NewStyle().Padding(1, 2)

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("170")).
			MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	HighlightStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212"))

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)

	TabActiveStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("170")).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(lipgloss.Color("170")).
			Padding(0, 2)

	TabInactiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				Padding(0, 2)

	DetailLabelStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("39"))

	DetailValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))
)
