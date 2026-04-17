package ui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tabID int

const (
	TabClock     tabID = iota
	TabConvert
	TabTimer
	TabStopwatch
	TabAlarm
	tabCount
)

var tabNames = []string{"[C]lock", "Con[V]ert", "[T]imer", "Stop[W]atch", "[A]larm"}

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
}

// Model is the root Bubbletea model.
type Model struct {
	activeTab  tabID
	clock      clockModel
	converter  convertModel
	timer      timerModel
	stopwatch  stopwatchModel
	alarm      alarmModel
	width      int
	height     int
	quitting   bool
}

// New creates the root model with optional CLI-driven defaults.
func New(opts Options) Model {
	return Model{
		activeTab: opts.InitialTab,
		clock:     newClock(opts.Timezone),
		converter: newConverter(),
		timer:     newTimer(opts.TimerPreset),
		stopwatch: newStopwatch(),
		alarm:     newAlarm(opts.AlarmPreset),
		width:     80,
		height:    24,
	}
}

func (m Model) Init() tea.Cmd {
	return tickCmd()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.activeTab != TabConvert && m.activeTab != TabAlarm {
				m.quitting = true
				return m, tea.Quit
			}
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
			// Only intercept if not already on Alarm tab, so 'a' still
			// works as "add alarm" within the Alarm tab itself.
			if m.activeTab != TabAlarm {
				m.activeTab = TabAlarm
				return m, nil
			}
		}
	}

	// Route tick to all sub-models that need it.
	if _, ok := msg.(tickMsg); ok {
		m.clock = m.clock.update(msg)
		m.timer, _ = m.timer.update(msg)
		m.stopwatch = m.stopwatch.update(msg)
		m.alarm, _ = m.alarm.update(msg)
		return m, tickCmd()
	}

	// Route other messages to the active tab.
	var cmd tea.Cmd
	switch m.activeTab {
	case TabConvert:
		m.converter, cmd = m.converter.update(msg)
	case TabTimer:
		m.timer, cmd = m.timer.update(msg)
	case TabStopwatch:
		m.stopwatch = m.stopwatch.update(msg)
	case TabAlarm:
		m.alarm, cmd = m.alarm.update(msg)
	}
	return m, cmd
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	tabs := m.renderTabs()
	content := m.renderContent()
	footer := styleHelp.Render("c/v/t/w/a jump tab · ←/→ cycle · q quit")

	inner := lipgloss.JoinVertical(lipgloss.Left, tabs, "", content, "", footer)

	return lipgloss.NewStyle().
		Width(m.width).
		MaxWidth(m.width).
		Render(inner)
}

func (m Model) renderTabs() string {
	parts := make([]string, len(tabNames))
	for i, name := range tabNames {
		if tabID(i) == m.activeTab { //nolint:unconvert
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
	}
	return ""
}
