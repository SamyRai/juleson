package theme

import "github.com/charmbracelet/lipgloss"

var (
	SuccessColor = lipgloss.Color("42")
	ErrorColor   = lipgloss.Color("196")
	WarnColor    = lipgloss.Color("214")
	InfoColor    = lipgloss.Color("69")
	MutedColor   = lipgloss.Color("240")
	HeaderColor  = lipgloss.Color("205")

	SuccessStyle = lipgloss.NewStyle().Foreground(SuccessColor).Bold(true)
	ErrorStyle   = lipgloss.NewStyle().Foreground(ErrorColor).Bold(true)
	WarnStyle    = lipgloss.NewStyle().Foreground(WarnColor).Bold(true)
	InfoStyle    = lipgloss.NewStyle().Foreground(InfoColor).Bold(true)
	MutedStyle   = lipgloss.NewStyle().Foreground(MutedColor)
	HeaderStyle  = lipgloss.NewStyle().Foreground(HeaderColor).Bold(true).Underline(true)
	StepStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
)
