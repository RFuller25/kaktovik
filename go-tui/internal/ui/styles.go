package ui

import "github.com/charmbracelet/lipgloss"

// themeColors defines the palette for one theme preset.
type themeColors struct {
	Accent    lipgloss.Color
	Dim       lipgloss.Color
	Muted     lipgloss.Color
	Highlight lipgloss.Color
	Green     lipgloss.Color
	Red       lipgloss.Color
	Bg        lipgloss.Color
}

var themePresets = map[string]themeColors{
	"default": {
		Accent:    "#5FD7FF",
		Dim:       "#555555",
		Muted:     "#888888",
		Highlight: "#FFD700",
		Green:     "#87FF5F",
		Red:       "#FF5F5F",
		Bg:        "#1A1A2E",
	},
	"nord": {
		Accent:    "#88C0D0",
		Dim:       "#4C566A",
		Muted:     "#616E88",
		Highlight: "#EBCB8B",
		Green:     "#A3BE8C",
		Red:       "#BF616A",
		Bg:        "#2E3440",
	},
	"solarized": {
		Accent:    "#268BD2",
		Dim:       "#073642",
		Muted:     "#586E75",
		Highlight: "#B58900",
		Green:     "#859900",
		Red:       "#DC322F",
		Bg:        "#002B36",
	},
}

// active color vars — mutated by ApplyTheme.
var (
	colorAccent    lipgloss.Color
	colorDim       lipgloss.Color
	colorMuted     lipgloss.Color
	colorHighlight lipgloss.Color
	colorGreen     lipgloss.Color
	colorRed       lipgloss.Color
	colorBg        lipgloss.Color
)

// Style vars — recreated by ApplyTheme.
var (
	styleTab          lipgloss.Style
	styleTabActive    lipgloss.Style
	stylePanelTitle   lipgloss.Style
	styleBigTime      lipgloss.Style
	styleNormalTime   lipgloss.Style
	styleLabel        lipgloss.Style
	styleValue        lipgloss.Style
	styleHelp         lipgloss.Style
	styleSuccess      lipgloss.Style
	styleWarn         lipgloss.Style
	styleError        lipgloss.Style
	styleBorder       lipgloss.Style
	styleInputFocused lipgloss.Style
	styleInputBlurred lipgloss.Style
	styleAccent       lipgloss.Style
	styleMuted        lipgloss.Style
)

func init() {
	ApplyTheme("default")
}

// ApplyTheme sets colors and reinitializes all style vars for the named theme.
// Falls back to "default" for unknown names.
func ApplyTheme(name string) {
	t, ok := themePresets[name]
	if !ok {
		t = themePresets["default"]
	}
	colorAccent = t.Accent
	colorDim = t.Dim
	colorMuted = t.Muted
	colorHighlight = t.Highlight
	colorGreen = t.Green
	colorRed = t.Red
	colorBg = t.Bg

	styleTab = lipgloss.NewStyle().Padding(0, 2).Foreground(colorMuted)
	styleTabActive = lipgloss.NewStyle().Padding(0, 2).Foreground(colorAccent).Bold(true).Underline(true)
	stylePanelTitle = lipgloss.NewStyle().Foreground(colorAccent).Bold(true).MarginBottom(1)
	styleBigTime = lipgloss.NewStyle().Foreground(colorHighlight).Bold(true)
	styleNormalTime = lipgloss.NewStyle().Foreground(colorMuted)
	styleLabel = lipgloss.NewStyle().Foreground(colorAccent).Width(14)
	styleValue = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	styleHelp = lipgloss.NewStyle().Foreground(colorDim).MarginTop(1)
	styleSuccess = lipgloss.NewStyle().Foreground(colorGreen).Bold(true)
	styleWarn = lipgloss.NewStyle().Foreground(colorHighlight)
	styleError = lipgloss.NewStyle().Foreground(colorRed)
	styleBorder = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colorAccent).Padding(1, 2)
	styleInputFocused = lipgloss.NewStyle().Foreground(colorHighlight)
	styleInputBlurred = lipgloss.NewStyle().Foreground(colorMuted)
	styleAccent = lipgloss.NewStyle().Foreground(colorAccent)
	styleMuted = lipgloss.NewStyle().Foreground(colorMuted)
}
