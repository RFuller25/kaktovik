package ui

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rfuller25/kaktovik/go-tui/internal/alarmstore"
	"github.com/rfuller25/kaktovik/go-tui/internal/config"
	"github.com/rfuller25/kaktovik/go-tui/internal/ktv"
	"github.com/rfuller25/kaktovik/go-tui/internal/notify"
)

type alarm struct {
	label    string
	target   time.Time
	ktv      ktv.Time
	fired    bool
	enabled  bool
	unitName string // systemd transient timer unit, empty if not scheduled
}

type alarmMode int

const (
	alarmList alarmMode = iota
	alarmAdd
)

type alarmModel struct {
	mode    alarmMode
	ktvMode bool             // when true, add form uses KTV time input
	alarms  []alarm
	cursor  int
	inputs  []textinput.Model // [0]=HH/KTV, [1]=MM, [2]=SS, [3]=label
	focus   int
	err     string
	now     time.Time
}

func newAlarm(presetTime time.Time) alarmModel {
	m := alarmModel{
		mode:   alarmList,
		now:    time.Now(),
		inputs: makeNormalInputs(),
	}
	// Load alarms persisted from previous sessions.
	if saved, err := alarmstore.Load(); err == nil {
		for _, s := range saved {
			m.alarms = append(m.alarms, alarmFromStore(s))
		}
	}
	if !presetTime.IsZero() {
		m.mode = alarmAdd
		m.inputs[0].SetValue(fmt.Sprintf("%02d", presetTime.Hour()))
		m.inputs[1].SetValue(fmt.Sprintf("%02d", presetTime.Minute()))
		m.inputs[2].SetValue(fmt.Sprintf("%02d", presetTime.Second()))
	}
	return m
}

func alarmFromStore(s alarmstore.Alarm) alarm {
	return alarm{
		label:    s.Label,
		target:   s.Target,
		ktv:      ktv.FromHMS(s.Target.Hour(), s.Target.Minute(), s.Target.Second(), 0),
		fired:    s.Fired,
		enabled:  s.Enabled,
		unitName: s.UnitName,
	}
}

func (a alarm) toStore() alarmstore.Alarm {
	return alarmstore.Alarm{
		Label:    a.label,
		Target:   a.target,
		Enabled:  a.enabled,
		Fired:    a.fired,
		UnitName: a.unitName,
	}
}

// scheduleAlarmUnit creates a systemd transient timer that fires kaktovik alarm
// --headless --immediate at the given wall-clock time, even if the TUI is closed.
// Returns the unit name on success, empty string on failure.
func scheduleAlarmUnit(target time.Time) string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	unitName := fmt.Sprintf("kaktovik-alarm-%d", target.Unix())
	calSpec := target.Local().Format("2006-01-02 15:04:05")
	err = exec.Command("systemd-run", "--user",
		"--unit="+unitName,
		"--timer-property=AccuracySec=1s",
		"--timer-property=RemainAfterElapse=no",
		"--on-calendar="+calSpec,
		"--",
		exe, "alarm", "--headless", "--immediate", target.Format("15:04:05"),
	).Run()
	if err != nil {
		return ""
	}
	return unitName
}

// cancelAlarmUnit stops a pending systemd timer unit so it does not fire.
func cancelAlarmUnit(unitName string) {
	if unitName == "" {
		return
	}
	exec.Command("systemctl", "--user", "stop", unitName+".timer").Run() //nolint:errcheck
}

// persistAlarms saves the current alarm list to disk in a background goroutine.
func persistAlarms(alarms []alarm) {
	saved := make([]alarmstore.Alarm, len(alarms))
	for i, a := range alarms {
		saved[i] = a.toStore()
	}
	alarmstore.Save(saved) //nolint:errcheck
}

// IsCapturingInput reports whether the alarm model is actively editing a text field.
// Used by the root model to suppress global hotkeys during text entry.
func (m alarmModel) IsCapturingInput() bool {
	return m.mode == alarmAdd
}

// alarmEnterNextFocus returns the next focus index when Enter is pressed.
// Returns cur if on the last field (triggering save), otherwise advances.
func alarmEnterNextFocus(cur int, ktvMode bool) int {
	if cur == 3 {
		return 3 // same → caller triggers saveAlarm
	}
	if ktvMode {
		return 3 // KTV mode: 0 → 3 (label)
	}
	return cur + 1
}

func makeNormalInputs() []textinput.Model {
	placeholders := []string{"HH (0-23)", "MM (0-59)", "SS (0-59)", "Label (optional)"}
	inputs := make([]textinput.Model, 4)
	for i := range inputs {
		t := textinput.New()
		t.Placeholder = placeholders[i]
		t.CharLimit = 24
		t.Width = 22
		inputs[i] = t
	}
	inputs[0].Focus()
	return inputs
}

func makeKTVInputs() []textinput.Model {
	inputs := make([]textinput.Model, 4)
	for i := range inputs {
		t := textinput.New()
		t.CharLimit = 24
		t.Width = 28
		inputs[i] = t
	}
	inputs[0].Placeholder = "𝋅𝋃𝋉𝋂 or 5.3.9.2 (base-20)"
	inputs[3].Placeholder = "Label (optional)"
	inputs[0].Focus()
	return inputs
}

func (m alarmModel) update(msg tea.Msg, cfg config.Config) (alarmModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		m.now = time.Time(msg)
		m.checkAlarms(cfg)

	case tea.KeyMsg:
		if m.mode == alarmList {
			switch msg.String() {
			case "a", "n":
				m.mode = alarmAdd
				m.ktvMode = false
				m.inputs = makeNormalInputs()
				m.focus = 0
				m.err = ""
				return m, nil
			case "d", "delete":
				if len(m.alarms) > 0 {
					unitName := m.alarms[m.cursor].unitName
					m.alarms = append(m.alarms[:m.cursor], m.alarms[m.cursor+1:]...)
					if m.cursor >= len(m.alarms) && m.cursor > 0 {
						m.cursor--
					}
					alarms := append([]alarm(nil), m.alarms...)
					go func() {
						cancelAlarmUnit(unitName)
						persistAlarms(alarms)
					}()
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
			}
		} else { // alarmAdd
			switch msg.String() {
			case "esc":
				m.mode = alarmList
				return m, nil
			case "ctrl+k":
				// Toggle between normal and KTV input mode
				m.ktvMode = !m.ktvMode
				if m.ktvMode {
					m.inputs = makeKTVInputs()
				} else {
					m.inputs = makeNormalInputs()
				}
				m.focus = 0
				m.err = ""
				return m, nil
			case "tab", "down":
				m.inputs[m.focus].Blur()
				m.focus = nextFocus(m.focus, m.ktvMode, false)
				m.inputs[m.focus].Focus()
				return m, nil
			case "shift+tab", "up":
				m.inputs[m.focus].Blur()
				m.focus = nextFocus(m.focus, m.ktvMode, true)
				m.inputs[m.focus].Focus()
				return m, nil
			case "enter":
				next := alarmEnterNextFocus(m.focus, m.ktvMode)
				if next != m.focus {
					m.inputs[m.focus].Blur()
					m.focus = next
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

// nextFocus returns the next focused input index given the current mode.
// In normal mode: 0→1→2→3 (wrap); in KTV mode: 0→3 (wrap).
func nextFocus(cur int, ktvMode bool, reverse bool) int {
	if ktvMode {
		// only inputs 0 and 3 are used
		if cur == 0 {
			if reverse {
				return 3
			}
			return 3
		}
		return 0
	}
	if reverse {
		return (cur - 1 + 4) % 4
	}
	return (cur + 1) % 4
}

func (m *alarmModel) checkAlarms(cfg config.Config) {
	changed := false
	for i := range m.alarms {
		a := &m.alarms[i]
		if a.enabled && !a.fired && m.now.After(a.target) {
			a.fired = true
			changed = true
			unitName := a.unitName
			a.unitName = ""
			go func(label string, kv ktv.Time, urgency, icon string, soundEnabled bool, soundFile, un string) {
				cancelAlarmUnit(un) // prevent double-notification from the systemd unit
				title := "Kaktovik Alarm"
				body := fmt.Sprintf("Alarm: %s  (%s)", kv.Spaced(), label)
				notify.SendUrgent(title, body, urgency, icon)
				notify.TerminalAttention()
				notify.PlaySound(soundEnabled, soundFile)
			}(a.label, a.ktv, cfg.NotifyUrgency, cfg.NotifyIcon, cfg.SoundEnabled, cfg.SoundFile, unitName)
		}
	}
	if changed {
		alarms := append([]alarm(nil), m.alarms...)
		go persistAlarms(alarms)
	}
}

func (m alarmModel) saveAlarm() alarmModel {
	m.err = ""

	var hh, mm, ss int
	var ktt ktv.Time

	if m.ktvMode {
		raw := strings.TrimSpace(m.inputs[0].Value())
		if raw == "" {
			m.err = "enter a Kaktovik time"
			return m
		}
		var err error
		ktt, err = ktv.ParseAny(raw)
		if err != nil {
			m.err = err.Error()
			return m
		}
		var ms int
		hh, mm, ss, ms = ktt.ToHMS()
		_ = ms
	} else {
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
		var ok bool
		if hh, ok = parse(0, "hour", 23); !ok {
			return m
		}
		if mm, ok = parse(1, "minute", 59); !ok {
			return m
		}
		if ss, ok = parse(2, "second", 59); !ok {
			return m
		}
		ktt = ktv.FromHMS(hh, mm, ss, 0)
	}

	label := strings.TrimSpace(m.inputs[3].Value())
	if label == "" {
		label = fmt.Sprintf("%02d:%02d:%02d  (%s)", hh, mm, ss, ktt.Dotted())
	}

	now := time.Now()
	target := time.Date(now.Year(), now.Month(), now.Day(), hh, mm, ss, 0, now.Location())
	if target.Before(now) {
		target = target.Add(24 * time.Hour)
	}

	a := alarm{
		label:   label,
		target:  target,
		ktv:     ktt,
		enabled: true,
	}
	// Schedule a background systemd timer so the alarm fires even if the TUI is closed.
	a.unitName = scheduleAlarmUnit(target)

	m.alarms = append(m.alarms, a)
	m.mode = alarmList
	m.cursor = len(m.alarms) - 1
	alarms := append([]alarm(nil), m.alarms...)
	go persistAlarms(alarms)
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
				switch {
				case a.fired:
					remStr = "fired"
				case !a.enabled:
					remStr = "disabled"
				case remaining < 0:
					remStr = "overdue"
				default:
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
		sb.WriteString(styleHelp.Render("a add · d delete · Space enable/disable · ↑↓/k/j move"))
	} else {
		modeStr := "Normal time"
		if m.ktvMode {
			modeStr = "Kaktovik time"
		}
		sb.WriteString(styleValue.Render("New Alarm — ") +
			styleAccent.Render(modeStr) +
			styleHelp.Render("  (Ctrl+K toggle)") + "\n\n")

		if m.ktvMode {
			focusMark0, focusMark3 := "  ", "  "
			if m.focus == 0 {
				focusMark0 = "> "
			} else {
				focusMark3 = "> "
			}
			sb.WriteString(fmt.Sprintf("%s%s %s\n", focusMark0, styleLabel.Render("KTV time:"), m.inputs[0].View()))
			sb.WriteString(fmt.Sprintf("%s%s %s\n", focusMark3, styleLabel.Render("Label:"), m.inputs[3].View()))
		} else {
			labels := []string{"Hour (0–23):", "Minute (0–59):", "Second (0–59):", "Label:"}
			for i, inp := range m.inputs {
				focusMark := "  "
				if i == m.focus {
					focusMark = "> "
				}
				sb.WriteString(fmt.Sprintf("%s%s %s\n", focusMark, styleLabel.Render(labels[i]), inp.View()))
			}
		}

		if m.err != "" {
			sb.WriteString("\n" + styleError.Render(m.err) + "\n")
		}
		sb.WriteString("\n")
		sb.WriteString(styleHelp.Render("Tab/↑↓ move · Enter confirm · Ctrl+K toggle KTV/normal · Escape cancel"))
	}

	return sb.String()
}

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
