package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rfuller25/kaktovik/go-tui/internal/config"
)

// configChangedMsg is emitted by settingsModel when any config value changes.
// The root model saves it to disk and propagates to sub-models.
type configChangedMsg config.Config

type fieldKind int

const (
	fieldBool fieldKind = iota
	fieldEnum
	fieldText
)

type settingsField struct {
	label    string
	kind     fieldKind
	options  []string // for fieldEnum
	inputIdx int      // for fieldText: index into settingsModel.inputs
}

var settingsFields = []settingsField{
	{label: "Sound enabled", kind: fieldBool},
	{label: "Sound file", kind: fieldText, inputIdx: 0},
	{label: "Notify urgency", kind: fieldEnum, options: []string{"normal", "critical"}},
	{label: "Notify icon", kind: fieldText, inputIdx: 1},
	{label: "Theme", kind: fieldEnum, options: config.ThemeNames},
	{label: "Clock mode", kind: fieldEnum, options: []string{"glyph", "ascii"}},
	{label: "Timezone", kind: fieldText, inputIdx: 2},
}

type settingsModel struct {
	cfg        config.Config
	cursor     int
	inputs     [3]textinput.Model
	inputFocus int // -1 = no text input focused
}

func newSettings(cfg config.Config) settingsModel {
	var inputs [3]textinput.Model
	placeholders := []string{
		"path to .oga/.wav file (empty = system default)",
		"path to icon file (empty = none)",
		"IANA timezone name (empty = local)",
	}
	initialValues := []string{cfg.SoundFile, cfg.NotifyIcon, cfg.Timezone}
	for i := range inputs {
		t := textinput.New()
		t.Placeholder = placeholders[i]
		t.CharLimit = 128
		t.Width = 50
		t.SetValue(initialValues[i])
		inputs[i] = t
	}
	return settingsModel{cfg: cfg, inputs: inputs, inputFocus: -1}
}

// IsCapturingInput returns true when a text field has keyboard focus.
func (m settingsModel) IsCapturingInput() bool {
	return m.inputFocus >= 0
}

func (m settingsModel) update(msg tea.Msg) (settingsModel, tea.Cmd) {
	if m.inputFocus >= 0 {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "escape" {
				m.inputs[m.inputFocus].Blur()
				m = m.flushInput(m.inputFocus)
				m.inputFocus = -1
				return m, emitConfigChanged(m.cfg)
			}
		}
		var cmd tea.Cmd
		m.inputs[m.inputFocus], cmd = m.inputs[m.inputFocus].Update(msg)
		m = m.flushInput(m.inputFocus)
		return m, tea.Batch(cmd, emitConfigChanged(m.cfg))
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(settingsFields)-1 {
				m.cursor++
			}
		case " ", "enter":
			f := settingsFields[m.cursor]
			switch f.kind {
			case fieldBool:
				m = m.toggleBool(m.cursor)
				return m, emitConfigChanged(m.cfg)
			case fieldEnum:
				m = m.cycleEnum(m.cursor)
				return m, emitConfigChanged(m.cfg)
			case fieldText:
				m.inputFocus = f.inputIdx
				m.inputs[m.inputFocus].Focus()
				return m, nil
			}
		}
	}
	return m, nil
}

func emitConfigChanged(cfg config.Config) tea.Cmd {
	return func() tea.Msg { return configChangedMsg(cfg) }
}

func (m settingsModel) flushInput(idx int) settingsModel {
	val := m.inputs[idx].Value()
	switch idx {
	case 0:
		m.cfg.SoundFile = val
	case 1:
		m.cfg.NotifyIcon = val
	case 2:
		m.cfg.Timezone = val
	}
	return m
}

func (m settingsModel) toggleBool(row int) settingsModel {
	switch row {
	case 0:
		m.cfg.SoundEnabled = !m.cfg.SoundEnabled
	}
	return m
}

func (m settingsModel) cycleEnum(row int) settingsModel {
	f := settingsFields[row]
	switch f.label {
	case "Notify urgency":
		m.cfg.NotifyUrgency = cycleOption(m.cfg.NotifyUrgency, f.options)
	case "Theme":
		m.cfg.Theme = cycleOption(m.cfg.Theme, f.options)
	case "Clock mode":
		m.cfg.ClockMode = cycleOption(m.cfg.ClockMode, f.options)
	}
	return m
}

func cycleOption(current string, options []string) string {
	for i, o := range options {
		if o == current {
			return options[(i+1)%len(options)]
		}
	}
	return options[0]
}

func (m settingsModel) fieldValue(row int) string {
	f := settingsFields[row]
	switch f.kind {
	case fieldBool:
		if m.cfg.SoundEnabled {
			return "● on"
		}
		return "○ off"
	case fieldEnum:
		switch f.label {
		case "Notify urgency":
			return m.cfg.NotifyUrgency
		case "Theme":
			return m.cfg.Theme
		case "Clock mode":
			return m.cfg.ClockMode
		}
	case fieldText:
		v := m.inputs[f.inputIdx].Value()
		if v == "" {
			return styleNormalTime.Render("(default)")
		}
		return v
	}
	return ""
}

func (m settingsModel) view(_ int) string {
	var sb strings.Builder
	sb.WriteString(stylePanelTitle.Render("SETTINGS"))
	sb.WriteString("\n\n")

	for i, f := range settingsFields {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}

		label := styleLabel.Render(f.label + ":")

		var value string
		if f.kind == fieldText && m.inputFocus == f.inputIdx {
			value = m.inputs[f.inputIdx].View()
		} else {
			value = styleValue.Render(m.fieldValue(i))
		}

		sb.WriteString(fmt.Sprintf("%s%s %s\n", cursor, label, value))
	}

	sb.WriteString("\n")
	if m.inputFocus >= 0 {
		sb.WriteString(styleHelp.Render("Escape to confirm field · type to edit"))
	} else {
		sb.WriteString(styleHelp.Render("↑↓/k/j move · Space/Enter toggle or edit · Escape leaves field"))
	}
	sb.WriteString("\n")
	sb.WriteString(styleHelp.Render(fmt.Sprintf("Config saved to: %s", config.Path())))
	return sb.String()
}
