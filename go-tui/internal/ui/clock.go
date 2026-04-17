package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rfuller25/kaktovik/go-tui/internal/ktv"
)

type clockModel struct {
	now       time.Time
	tz        *time.Location
	glyphMode bool // true = large Unicode cells; false = ASCII style
}

func newClock(tz *time.Location, glyphMode bool) clockModel {
	if tz == nil {
		tz = time.Local
	}
	return clockModel{now: time.Now().In(tz), tz: tz, glyphMode: glyphMode}
}

func (c clockModel) update(msg tea.Msg) clockModel {
	switch msg := msg.(type) {
	case tickMsg:
		c.now = time.Now().In(c.tz)
	case tea.KeyMsg:
		if msg.String() == "g" {
			c.glyphMode = !c.glyphMode
		}
	}
	return c
}

func (c clockModel) view(width int) string {
	kt := ktv.FromTime(c.now)

	var digits string
	if c.glyphMode {
		digits = renderGlyphMode(kt, width)
	} else {
		digits = renderBigDigits(kt, width)
	}

	h, m, s, _ := kt.ToHMS()
	normalStr := fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	ktvDotted := kt.Dotted()

	tzName, offset := c.now.Zone()
	tzStr := fmt.Sprintf("%s (UTC%+d)", tzName, offset/3600)

	modeHint := "ascii"
	if c.glyphMode {
		modeHint = "glyph"
	}

	lines := []string{
		stylePanelTitle.Render("KAKTOVIK CLOCK"),
		"",
		digits,
		"",
		styleNormalTime.Render(fmt.Sprintf("  %s  ·  %s", ktvDotted, normalStr)),
		styleNormalTime.Render(fmt.Sprintf("  %s", tzStr)),
		"",
		styleHelp.Render(fmt.Sprintf("  Tab/←/→ switch views · g toggle glyph/ascii [%s] · q quit", modeHint)),
	}
	return strings.Join(lines, "\n")
}

// renderGlyphMode renders each Kaktovik digit as a large rounded-border cell
// containing the Unicode glyph prominently centred.
func renderGlyphMode(kt ktv.Time, width int) string {
	components := []struct {
		label string
		value int
	}{
		{"ikarraq", kt.Ikarraq},
		{"mein", kt.Mein},
		{"tick", kt.Tick},
		{"kick", kt.Kick},
	}

	cells := make([]string, 4)
	for i, c := range components {
		char := ktv.Digit(c.value)
		inner := lipgloss.JoinVertical(lipgloss.Center,
			"",
			styleBigTime.Copy().Render(char),
			"",
			styleNormalTime.Copy().Render(fmt.Sprintf("%s  %d", c.label, c.value)),
		)
		cells[i] = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorAccent).
			Padding(0, 4).
			Align(lipgloss.Center).
			Render(inner)
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, cells...)
	return lipgloss.NewStyle().
		Width(width - 4).
		Align(lipgloss.Center).
		Render(row)
}

// renderBigDigits is the original display: Unicode char + number, no borders.
func renderBigDigits(kt ktv.Time, width int) string {
	components := []struct {
		label string
		value int
	}{
		{"ikarraq", kt.Ikarraq},
		{"mein", kt.Mein},
		{"tick", kt.Tick},
		{"kick", kt.Kick},
	}

	cells := make([]string, 4)
	for i, c := range components {
		char := ktv.Digit(c.value)
		num := fmt.Sprintf("%d", c.value)
		cell := lipgloss.NewStyle().
			Width(12).
			Align(lipgloss.Center).
			Render(
				styleBigTime.Copy().Render(fmt.Sprintf("  %s  ", char)) + "\n" +
					styleNormalTime.Copy().Render(fmt.Sprintf(" (%s) ", num)),
			)
		cells[i] = cell
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, cells...)
	return lipgloss.NewStyle().
		Width(width - 4).
		Align(lipgloss.Center).
		Render(row)
}
