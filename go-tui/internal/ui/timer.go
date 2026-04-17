package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rfuller25/kaktovik/go-tui/internal/ktv"
	"github.com/rfuller25/kaktovik/go-tui/internal/notify"
)

type timerState int

const (
	timerIdle timerState = iota
	timerRunning
	timerPaused
	timerDone
)

type timerModel struct {
	state     timerState
	duration  time.Duration
	remaining time.Duration
	lastTick  time.Time
	input     textinput.Model
	inputFmt  string // "normal" or "ktv"
	err       string
}

func newTimer(preset time.Duration) timerModel {
	inp := textinput.New()
	inp.Placeholder = "5m30s · 1.2.3.0 · 𝋅𝋃𝋉𝋂"
	inp.CharLimit = 20
	inp.Width = 26
	inp.Focus()

	m := timerModel{
		state:    timerIdle,
		input:    inp,
		inputFmt: "normal",
	}
	if preset > 0 {
		m.duration = preset
		m.remaining = preset
		m.state = timerRunning
		m.lastTick = time.Now()
	}
	return m
}

func (m timerModel) update(msg tea.Msg) (timerModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		if m.state == timerRunning {
			now := time.Time(msg)
			elapsed := now.Sub(m.lastTick)
			m.lastTick = now
			if elapsed > m.remaining {
				m.remaining = 0
				m.state = timerDone
				go func() {
					notify.SendUrgent("Kaktovik Timer", "Your timer has finished!", "critical", "")
					notify.TerminalAttention()
					notify.PlaySound(true, "")
				}()
			} else {
				m.remaining -= elapsed
			}
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.state == timerIdle {
				d, err := parseDuration(m.input.Value())
				if err != nil {
					m.err = err.Error()
					return m, nil
				}
				m.err = ""
				m.duration = d
				m.remaining = d
				m.state = timerRunning
				m.lastTick = time.Now()
				return m, nil
			}
		case " ":
			switch m.state {
			case timerRunning:
				m.state = timerPaused
			case timerPaused:
				m.state = timerRunning
				m.lastTick = time.Now()
			case timerDone:
				m.state = timerIdle
				m.remaining = 0
				m.input.SetValue("")
			}
			return m, nil
		case "r", "ctrl+r":
			m.state = timerIdle
			m.remaining = m.duration
			m.input.SetValue("")
			return m, nil
		}
	}

	if m.state == timerIdle {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m timerModel) view(width int) string {
	var sb strings.Builder
	sb.WriteString(stylePanelTitle.Render("TIMER"))
	sb.WriteString("\n\n")

	switch m.state {
	case timerIdle:
		sb.WriteString(styleValue.Render("Duration: ") + m.input.View())
		sb.WriteString("\n")
		if m.err != "" {
			sb.WriteString(styleError.Render(m.err) + "\n")
		}
		sb.WriteString("\n")
		sb.WriteString(styleHelp.Render("Enter to start · formats: 5m30s · 1h5m · 1.2.3.0 (base-20) · 𝋅𝋃𝋉𝋂 (KTV chars)"))

	case timerRunning, timerPaused:
		kt := ktv.FromDuration(m.remaining)
		h, min, s, _ := formatDur(m.remaining)

		statusIcon := "▶ RUNNING"
		if m.state == timerPaused {
			statusIcon = "⏸ PAUSED"
		}
		sb.WriteString(styleWarn.Render(statusIcon) + "\n\n")
		sb.WriteString(styleBigTime.Render(fmt.Sprintf("  %s  ", kt.Spaced())) + "\n")
		sb.WriteString(styleNormalTime.Render(fmt.Sprintf("  %02d:%02d:%02d remaining  (%s)", h, min, s, kt.Dotted())) + "\n\n")

		pct := float64(m.remaining) / float64(m.duration)
		sb.WriteString(progressBar(pct, width-8) + "\n\n")
		sb.WriteString(styleHelp.Render("Space pause/resume · r reset"))

	case timerDone:
		sb.WriteString(styleSuccess.Render("TIMER COMPLETE!") + "\n\n")
		sb.WriteString(styleHelp.Render("Space to dismiss · r to restart"))
	}

	return sb.String()
}

func formatDur(d time.Duration) (h, m, s, ms int) {
	h = int(d.Hours())
	m = int(d.Minutes()) % 60
	s = int(d.Seconds()) % 60
	ms = int(d.Milliseconds()) % 1000
	return
}

func parseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("enter a duration")
	}
	// KTV numeral characters (𝋅𝋃𝋉𝋂) or dotted base-20 (5.3.9.2)
	if kt, err := ktv.ParseAny(s); err == nil {
		d := kt.ToDuration()
		if d <= 0 {
			return 0, fmt.Errorf("KTV time 0.0.0.0 has zero duration")
		}
		return d, nil
	}
	// Standard Go duration (5m30s, 1h, etc.)
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("invalid duration %q — use 5m30s, 1.2.3.0 (base-20), or 𝋅𝋃𝋉𝋂 (KTV chars)", s)
	}
	if d <= 0 {
		return 0, fmt.Errorf("duration must be positive")
	}
	return d, nil
}

func progressBar(pct float64, width int) string {
	if width < 4 {
		return ""
	}
	filled := int(pct * float64(width))
	if filled > width {
		filled = width
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return styleAccent.Render("[") + styleWarn.Render(bar) + styleAccent.Render("]")
}
