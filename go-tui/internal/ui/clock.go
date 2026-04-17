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

	bigDigits := renderBigDigits(kt, width)

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
