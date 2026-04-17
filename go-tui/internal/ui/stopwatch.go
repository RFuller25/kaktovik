package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rfuller25/kaktovik/go-tui/internal/ktv"
)

type stopwatchState int

const (
	swIdle stopwatchState = iota
	swRunning
	swPaused
)

type lapEntry struct {
	index   int
	elapsed time.Duration
	split   time.Duration
}

type stopwatchModel struct {
	state    stopwatchState
	elapsed  time.Duration
	lastTick time.Time
	laps     []lapEntry
}

func newStopwatch() stopwatchModel {
	return stopwatchModel{state: swIdle}
}

func (m stopwatchModel) update(msg tea.Msg) stopwatchModel {
	switch msg := msg.(type) {
	case tickMsg:
		if m.state == swRunning {
			now := time.Time(msg)
			m.elapsed += now.Sub(m.lastTick)
			m.lastTick = now
		}

	case tea.KeyMsg:
		switch msg.String() {
		case " ", "enter":
			switch m.state {
			case swIdle:
				m.state = swRunning
				m.lastTick = time.Now()
			case swRunning:
				m.state = swPaused
			case swPaused:
				m.state = swRunning
				m.lastTick = time.Now()
			}
		case "l", "L":
			if m.state == swRunning {
				var split time.Duration
				if len(m.laps) > 0 {
					split = m.elapsed - m.laps[len(m.laps)-1].elapsed
				} else {
					split = m.elapsed
				}
				m.laps = append(m.laps, lapEntry{
					index:   len(m.laps) + 1,
					elapsed: m.elapsed,
					split:   split,
				})
			}
		case "r", "ctrl+r":
			m.state = swIdle
			m.elapsed = 0
			m.laps = nil
		}
	}
	return m
}

func (m stopwatchModel) view(width int) string {
	var sb strings.Builder
	sb.WriteString(stylePanelTitle.Render("STOPWATCH"))
	sb.WriteString("\n\n")

	kt := ktv.FromDuration(m.elapsed)
	h, min, s, ms := formatDur(m.elapsed)

	statusIcon := "⏹ STOPPED"
	switch m.state {
	case swRunning:
		statusIcon = "▶ RUNNING"
	case swPaused:
		statusIcon = "⏸ PAUSED"
	}
	sb.WriteString(styleWarn.Render(statusIcon) + "\n\n")
	sb.WriteString(styleBigTime.Render(fmt.Sprintf("  %s  ", kt.Spaced())) + "\n")
	sb.WriteString(styleNormalTime.Render(fmt.Sprintf("  %02d:%02d:%02d.%03d  (%s)", h, min, s, ms, kt.Dotted())) + "\n\n")

	if len(m.laps) > 0 {
		sb.WriteString(styleLabel.Render("Laps:") + "\n")
		start := 0
		if len(m.laps) > 8 {
			start = len(m.laps) - 8
		}
		for _, lap := range m.laps[start:] {
			lh, lm, ls, lms := formatDur(lap.elapsed)
			sh, sm, ss2, sms := formatDur(lap.split)
			sb.WriteString(styleNormalTime.Render(fmt.Sprintf(
				"  #%d  total %02d:%02d:%02d.%03d  split %02d:%02d:%02d.%03d\n",
				lap.index, lh, lm, ls, lms, sh, sm, ss2, sms,
			)))
		}
		sb.WriteString("\n")
	}

	keyHelp := "Space/Enter start · l lap · r reset"
	if m.state == swRunning {
		keyHelp = "Space pause · l lap · r reset"
	} else if m.state == swPaused {
		keyHelp = "Space resume · r reset"
	}
	sb.WriteString(styleHelp.Render(keyHelp))
	return sb.String()
}

// formatDurStr returns a human string for a duration.
func formatDurStr(d time.Duration) string {
	h, m, s, ms := formatDur(d)
	return fmt.Sprintf("%02d:%02d:%02d.%03d", h, m, s, ms)
}

var _ = formatDurStr // keep for possible future use
