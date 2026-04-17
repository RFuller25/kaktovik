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
	now time.Time
	tz  *time.Location
}

func newClock(tz *time.Location) clockModel {
	if tz == nil {
		tz = time.Local
	}
	return clockModel{now: time.Now().In(tz), tz: tz}
}

func (c clockModel) update(msg tea.Msg) clockModel {
	if _, ok := msg.(tickMsg); ok {
		c.now = time.Now().In(c.tz)
	}
	return c
}

func (c clockModel) view(width int) string {
	kt := ktv.FromTime(c.now)

	bigDigits := renderBigAsciiDigits(kt, width)

	h, m, s, _ := kt.ToHMS()
	normalStr := fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	ktvDotted := kt.Dotted()

	tzName, offset := c.now.Zone()
	tzStr := fmt.Sprintf("%s (UTC%+d)", tzName, offset/3600)

	lines := []string{
		stylePanelTitle.Render("KAKTOVIK CLOCK"),
		"",
		bigDigits,
		"",
		styleNormalTime.Render(fmt.Sprintf("  %s  ·  %s", ktvDotted, normalStr)),
		styleNormalTime.Render(fmt.Sprintf("  %s", tzStr)),
		"",
		styleHelp.Render("  Tab/←/→ switch views · q quit"),
	}
	return strings.Join(lines, "\n")
}

// renderBigAsciiDigits renders the four Kaktovik time components as large
// box-drawing glyphs that work on any terminal regardless of font support.
func renderBigAsciiDigits(kt ktv.Time, width int) string {
	components := []struct {
		value int
		label string
	}{
		{kt.Ikarraq, "ikarraq"},
		{kt.Mein, "mein"},
		{kt.Tick, "tick"},
		{kt.Kick, "kick"},
	}

	// cellWidth gives each digit column some breathing room.
	const cellWidth = artWidth + 6

	blocks := make([]string, len(components))
	for i, c := range components {
		rows := digitLines(c.value)

		// Colour every row of the glyph in gold.
		artLines := make([]string, artHeight)
		for j, row := range rows {
			artLines[j] = styleBigTime.Render(row)
		}
		art := strings.Join(artLines, "\n")

		label := styleNormalTime.Render(fmt.Sprintf("%d · %s", c.value, c.label))

		col := lipgloss.JoinVertical(lipgloss.Center, art, label)
		blocks[i] = lipgloss.NewStyle().
			Width(cellWidth).
			Align(lipgloss.Center).
			Render(col)
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, blocks...)
	return lipgloss.NewStyle().
		Width(width - 2).
		Align(lipgloss.Center).
		Render(row)
}
