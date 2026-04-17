package ui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rfuller25/kaktovik/go-tui/internal/config"
)

type tabID int

const (
	TabClock     tabID = iota
	TabConvert
	TabTimer
	TabStopwatch
	TabAlarm
	TabSettings
	tabCount
)

var tabNames = []string{"[C]lock", "Con[V]ert", "[T]imer", "Stop[W]atch", "[A]larm", "[S]ettings"}

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Options carries CLI-driven initial state.
type Options struct {
	InitialTab  tabID
	TimerPreset time.Duration
	AlarmPreset time.Time
	Timezone    *time.Location
	Cfg         config.Config
}

// Model is the root Bubbletea model.
type Model struct {
	activeTab tabID
	clock     clockModel
	converter convertModel
	timer     timerModel
	stopwatch stopwatchModel
	alarm     alarmModel
	settings  settingsModel
	cfg       config.Config
	width     int
	height    int
	quitting  bool
}

// New creates the root model with optional CLI-driven defaults.
func New(opts Options) Model {
	return Model{
		activeTab: opts.InitialTab,
		clock:     newClock(opts.Timezone, opts.Cfg.ClockMode == "glyph"),
		converter: newConverter(),
		timer:     newTimer(opts.TimerPreset),
		stopwatch: newStopwatch(),
		alarm:     newAlarm(opts.AlarmPreset),
		settings:  newSettings(opts.Cfg),
		cfg:       opts.Cfg,
		width:     80,
		height:    24,
	}
}

func (m Model) Init() tea.Cmd {
	return tickCmd()
}

// isCapturingInput returns true when a sub-model has text input focus,
// meaning global hotkeys (tab switching, quit) should be suppressed.
func (m Model) isCapturingInput() bool {
	return m.alarm.IsCapturingInput() ||
		(m.activeTab == TabSettings && m.settings.IsCapturingInput()) ||
		(m.activeTab == TabTimer && m.timer.IsCapturingInput()) ||
		(m.activeTab == TabConvert && m.converter.IsCapturingInput())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case configChangedMsg:
		m.cfg = config.Config(msg)
		m.clock.glyphMode = (m.cfg.ClockMode == "glyph")
		ApplyTheme(m.cfg.Theme)
		go config.Save(m.cfg) //nolint:errcheck
		return m, nil

	case tea.KeyMsg:
		if !m.isCapturingInput() {
			switch msg.String() {
			case "ctrl+c", "q":
				m.quitting = true
				return m, tea.Quit
			case "right", "ctrl+right":
				m.activeTab = (m.activeTab + 1) % tabCount
				return m, nil
			case "left", "ctrl+left":
				m.activeTab = (m.activeTab - 1 + tabCount) % tabCount
				return m, nil
			case "c":
				m.activeTab = TabClock
				return m, nil
			case "v":
				m.activeTab = TabConvert
				return m, nil
			case "t":
				m.activeTab = TabTimer
				return m, nil
			case "w":
				m.activeTab = TabStopwatch
				return m, nil
			case "a":
				if m.activeTab != TabAlarm {
					m.activeTab = TabAlarm
					return m, nil
				}
			case "s":
				m.activeTab = TabSettings
				return m, nil
			}
		}

		// Clock handles 'g' toggle itself; persist the change.
		if m.activeTab == TabClock {
			m.clock = m.clock.update(msg)
			newMode := "ascii"
			if m.clock.glyphMode {
				newMode = "glyph"
			}
			if newMode != m.cfg.ClockMode {
				m.cfg.ClockMode = newMode
				m.settings.cfg.ClockMode = newMode
				go config.Save(m.cfg) //nolint:errcheck
			}
			return m, nil
		}
	}

	// Route tick to all sub-models that need it.
	if _, ok := msg.(tickMsg); ok {
		m.clock = m.clock.update(msg)
		m.timer, _ = m.timer.update(msg, m.cfg)
		m.stopwatch = m.stopwatch.update(msg)
		m.alarm, _ = m.alarm.update(msg, m.cfg)
		return m, tickCmd()
	}

	// Route other messages to the active tab.
	var cmd tea.Cmd
	switch m.activeTab {
	case TabClock:
		m.clock = m.clock.update(msg)
	case TabConvert:
		m.converter, cmd = m.converter.update(msg)
	case TabTimer:
		m.timer, cmd = m.timer.update(msg, m.cfg)
	case TabStopwatch:
		m.stopwatch = m.stopwatch.update(msg)
	case TabAlarm:
		m.alarm, cmd = m.alarm.update(msg, m.cfg)
	case TabSettings:
		m.settings, cmd = m.settings.update(msg)
	}
	return m, cmd
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	tabs := m.renderTabs()
	content := m.renderContent()
	footer := styleHelp.Render("c/v/t/w/a/s jump tab · ←/→ cycle · q quit")

	inner := lipgloss.JoinVertical(lipgloss.Left, tabs, "", content, "", footer)

	return lipgloss.NewStyle().
		Width(m.width).
		MaxWidth(m.width).
		Render(inner)
}

func (m Model) renderTabs() string {
	parts := make([]string, len(tabNames))
	for i, name := range tabNames {
		if tabID(i) == m.activeTab {
			parts[i] = styleTabActive.Render(name)
		} else {
			parts[i] = styleTab.Render(name)
		}
	}
	line := strings.Join(parts, styleNormalTime.Render("│"))
	return line + "\n" + strings.Repeat("─", m.width)
}

func (m Model) renderContent() string {
	contentWidth := m.width
	switch m.activeTab {
	case TabClock:
		return m.clock.view(contentWidth)
	case TabConvert:
		return m.converter.view(contentWidth)
	case TabTimer:
		return m.timer.view(contentWidth)
	case TabStopwatch:
		return m.stopwatch.view(contentWidth)
	case TabAlarm:
		return m.alarm.view(contentWidth)
	case TabSettings:
		return m.settings.view(contentWidth)
	}
	return ""
}
