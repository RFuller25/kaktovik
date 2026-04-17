package ui

import "github.com/charmbracelet/lipgloss"

var (
	colorAccent    = lipgloss.Color("#5FD7FF")
	colorDim       = lipgloss.Color("#555555")
	colorMuted     = lipgloss.Color("#888888")
	colorHighlight = lipgloss.Color("#FFD700")
	colorGreen     = lipgloss.Color("#87FF5F")
	colorRed       = lipgloss.Color("#FF5F5F")
	colorBg        = lipgloss.Color("#1A1A2E")

	styleTab = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(colorMuted)

	styleTabActive = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(colorAccent).
			Bold(true).
			Underline(true)

	stylePanelTitle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true).
			MarginBottom(1)

	styleBigTime = lipgloss.NewStyle().
			Foreground(colorHighlight).
			Bold(true)

	styleNormalTime = lipgloss.NewStyle().
			Foreground(colorMuted)

	styleLabel = lipgloss.NewStyle().
			Foreground(colorAccent).
			Width(14)

	styleValue = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	styleHelp = lipgloss.NewStyle().
			Foreground(colorDim).
			MarginTop(1)

	styleSuccess = lipgloss.NewStyle().
			Foreground(colorGreen).
			Bold(true)

	styleWarn = lipgloss.NewStyle().
			Foreground(colorHighlight)

	styleError = lipgloss.NewStyle().
			Foreground(colorRed)

	styleBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorAccent).
			Padding(1, 2)

	styleInputFocused = lipgloss.NewStyle().
				Foreground(colorHighlight)

	styleInputBlurred = lipgloss.NewStyle().
				Foreground(colorMuted)

	styleAccent = lipgloss.NewStyle().Foreground(colorAccent)
)
