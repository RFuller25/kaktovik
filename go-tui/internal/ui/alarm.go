package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rfuller25/kaktovik/go-tui/internal/ktv"
	"github.com/rfuller25/kaktovik/go-tui/internal/notify"
)

type alarm struct {
	label   string
	target  time.Time
	ktv     ktv.Time
	fired   bool
	enabled bool
}

type alarmMode int

const (
	alarmList alarmMode = iota
	alarmAdd
)

type alarmModel struct {
	mode    alarmMode
	alarms  []alarm
	cursor  int
	inputs  []textinput.Model // HH, MM, SS, label
	focus   int
	err     string
	now     time.Time
}

func newAlarm(presetTime time.Time) alarmModel {
	inputs := make([]textinput.Model, 4)
	placeholders := []string{"HH (0-23)", "MM (0-59)", "SS (0-59)", "Label (optional)"}
	for i := range inputs {
		t := textinput.New()
		t.Placeholder = placeholders[i]
		t.CharLimit = 20
		t.Width = 20
		inputs[i] = t
	}
	inputs[0].Focus()

	m := alarmModel{
		mode:  alarmList,
		now:   time.Now(),
		inputs: inputs,
	}

	if !presetTime.IsZero() {
		m.mode = alarmAdd
		inputs[0].SetValue(fmt.Sprintf("%02d", presetTime.Hour()))
		inputs[1].SetValue(fmt.Sprintf("%02d", presetTime.Minute()))
		inputs[2].SetValue(fmt.Sprintf("%02d", presetTime.Second()))
	}

	return m
}

func (m alarmModel) update(msg tea.Msg) (alarmModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		m.now = time.Time(msg)
		m.checkAlarms()

	case tea.KeyMsg:
		if m.mode == alarmList {
			switch msg.String() {
			case "a", "n":
				m.mode = alarmAdd
				for i := range m.inputs {
					m.inputs[i].SetValue("")
					m.inputs[i].Blur()
				}
				m.focus = 0
				m.inputs[0].Focus()
				m.err = ""
				return m, nil
			case "d", "delete":
				if len(m.alarms) > 0 {
					m.alarms = append(m.alarms[:m.cursor], m.alarms[m.cursor+1:]...)
					if m.cursor >= len(m.alarms) && m.cursor > 0 {
						m.cursor--
					}
				}
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.alarms)-1 {
					m.cursor++
				}
			case " ":
				if len(m.alarms) > 0 {
					m.alarms[m.cursor].enabled = !m.alarms[m.cursor].enabled
				}
			case "escape":
				// nothing
			}
		} else { // alarmAdd
			switch msg.String() {
			case "escape":
				m.mode = alarmList
				return m, nil
			case "tab", "down":
				m.inputs[m.focus].Blur()
				m.focus = (m.focus + 1) % len(m.inputs)
				m.inputs[m.focus].Focus()
				return m, nil
			case "shift+tab", "up":
				m.inputs[m.focus].Blur()
				m.focus = (m.focus - 1 + len(m.inputs)) % len(m.inputs)
				m.inputs[m.focus].Focus()
				return m, nil
			case "enter":
				if m.focus < len(m.inputs)-1 {
					m.inputs[m.focus].Blur()
					m.focus++
					m.inputs[m.focus].Focus()
				} else {
					m = m.saveAlarm()
				}
				return m, nil
			}
			var cmd tea.Cmd
			m.inputs[m.focus], cmd = m.inputs[m.focus].Update(msg)
			return m, cmd
		}
	}
	return m, nil
}

func (m *alarmModel) checkAlarms() {
	for i := range m.alarms {
		a := &m.alarms[i]
		if a.enabled && !a.fired && m.now.After(a.target) {
			a.fired = true
			go func(label string, kv ktv.Time) {
				title := "Kaktovik Alarm"
				body := fmt.Sprintf("Alarm: %s  (%s)", kv.Spaced(), label)
				notify.SendUrgent(title, body)
				notify.PlaySound()
			}(a.label, a.ktv)
		}
	}
}

func (m alarmModel) saveAlarm() alarmModel {
	m.err = ""
	parse := func(idx int, name string, max int) (int, bool) {
		v := strings.TrimSpace(m.inputs[idx].Value())
		if v == "" {
			return 0, true
		}
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 || n > max {
			m.err = fmt.Sprintf("%s must be 0–%d", name, max)
			return 0, false
		}
		return n, true
	}

	hh, ok := parse(0, "hour", 23)
	if !ok {
		return m
	}
	mm, ok := parse(1, "minute", 59)
	if !ok {
		return m
	}
	ss, ok := parse(2, "second", 59)
	if !ok {
		return m
	}
	label := strings.TrimSpace(m.inputs[3].Value())

	now := time.Now()
	target := time.Date(now.Year(), now.Month(), now.Day(), hh, mm, ss, 0, now.Location())
	if target.Before(now) {
		target = target.Add(24 * time.Hour)
	}

	kt := ktv.FromHMS(hh, mm, ss, 0)
	if label == "" {
		label = fmt.Sprintf("%02d:%02d:%02d", hh, mm, ss)
	}

	m.alarms = append(m.alarms, alarm{
		label:   label,
		target:  target,
		ktv:     kt,
		enabled: true,
	})
	m.mode = alarmList
	m.cursor = len(m.alarms) - 1
	return m
}

func (m alarmModel) view(width int) string {
	var sb strings.Builder
	sb.WriteString(stylePanelTitle.Render("ALARMS"))
	sb.WriteString("\n\n")

	if m.mode == alarmList {
		if len(m.alarms) == 0 {
			sb.WriteString(styleNormalTime.Render("No alarms set. Press 'a' to add one."))
			sb.WriteString("\n\n")
		} else {
			for i, a := range m.alarms {
				cursor := "  "
				if i == m.cursor {
					cursor = "> "
				}
				status := styleSuccess.Render("ON ")
				if !a.enabled {
					status = styleNormalTime.Render("off")
				}
				if a.fired {
					status = styleMuted.Render("FIN")
				}

				remaining := a.target.Sub(m.now)
				var remStr string
				if a.fired {
					remStr = "fired"
				} else if !a.enabled {
					remStr = "disabled"
				} else if remaining < 0 {
					remStr = "overdue"
				} else {
					remStr = fmt.Sprintf("in %s", fmtCountdown(remaining))
				}

				sb.WriteString(fmt.Sprintf("%s[%s] %s  %s  %s\n",
					cursor,
					status,
					styleValue.Render(a.ktv.Spaced()),
					styleNormalTime.Render(a.label),
					styleNormalTime.Render(remStr),
				))
			}
			sb.WriteString("\n")
		}
		sb.WriteString(styleHelp.Render("a add · d delete · Space enable/disable · ↑↓ move"))
	} else {
		sb.WriteString(styleValue.Render("New Alarm (normal time)") + "\n\n")
		labels := []string{"Hour (0–23):", "Minute (0–59):", "Second (0–59):", "Label:"}
		for i, inp := range m.inputs {
			focusMark := "  "
			if i == m.focus {
				focusMark = "> "
			}
			sb.WriteString(fmt.Sprintf("%s%s %s\n", focusMark, styleLabel.Render(labels[i]), inp.View()))
		}
		if m.err != "" {
			sb.WriteString("\n" + styleError.Render(m.err) + "\n")
		}
		sb.WriteString("\n")
		sb.WriteString(styleHelp.Render("Tab/↑↓ move · Enter confirm · Escape cancel"))
	}

	return sb.String()
}

var styleMuted = lipgloss.NewStyle().Foreground(colorMuted)

func fmtCountdown(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh%02dm%02ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm%02ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}
