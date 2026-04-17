package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rfuller25/kaktovik/go-tui/internal/ktv"
)

type convertMode int

const (
	normalToKTV convertMode = iota
	ktvToNormal
)

type convertModel struct {
	mode   convertMode
	inputs []textinput.Model // hh, mm, ss for normal; i, m, t, k for ktv
	focus  int
	result string
	err    string
}

func newConverter() convertModel {
	inputs := make([]textinput.Model, 4)
	for i := range inputs {
		t := textinput.New()
		t.CharLimit = 5
		t.Width = 5
		inputs[i] = t
	}
	inputs[0].Placeholder = "HH"
	inputs[1].Placeholder = "MM"
	inputs[2].Placeholder = "SS"
	inputs[3].Placeholder = "MS"
	inputs[0].Focus()
	return convertModel{mode: normalToKTV, inputs: inputs, focus: 0}
}

func (c convertModel) setMode(m convertMode) convertModel {
	c.mode = m
	c.result = ""
	c.err = ""
	for i := range c.inputs {
		c.inputs[i].SetValue("")
		c.inputs[i].Blur()
	}
	c.focus = 0
	c.inputs[0].Focus()

	if m == normalToKTV {
		c.inputs[0].Placeholder = "HH"
		c.inputs[1].Placeholder = "MM"
		c.inputs[2].Placeholder = "SS"
		c.inputs[3].Placeholder = "MS"
	} else {
		c.inputs[0].Placeholder = "I (0-19)"
		c.inputs[1].Placeholder = "M (0-19)"
		c.inputs[2].Placeholder = "T (0-19)"
		c.inputs[3].Placeholder = "K (0-19)"
	}
	return c
}

// IsCapturingInput always returns true: the converter keeps a text input focused at all times.
func (c convertModel) IsCapturingInput() bool {
	return true
}

func (c convertModel) update(msg tea.Msg) (convertModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			c.inputs[c.focus].Blur()
			c.focus = (c.focus + 1) % len(c.inputs)
			c.inputs[c.focus].Focus()
			return c, nil
		case "shift+tab", "up":
			c.inputs[c.focus].Blur()
			c.focus = (c.focus - 1 + len(c.inputs)) % len(c.inputs)
			c.inputs[c.focus].Focus()
			return c, nil
		case "enter":
			c = c.calculate()
			return c, nil
		case "ctrl+r", "r":
			if c.mode == normalToKTV {
				c = c.setMode(ktvToNormal)
			} else {
				c = c.setMode(normalToKTV)
			}
			return c, nil
		}
	}

	var cmd tea.Cmd
	c.inputs[c.focus], cmd = c.inputs[c.focus].Update(msg)
	return c, cmd
}

func (c convertModel) calculate() convertModel {
	c.err = ""
	c.result = ""

	parse := func(idx int, name string, max int) (int, bool) {
		v := strings.TrimSpace(c.inputs[idx].Value())
		if v == "" {
			return 0, true
		}
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 || n > max {
			c.err = fmt.Sprintf("%s must be 0–%d", name, max)
			return 0, false
		}
		return n, true
	}

	if c.mode == normalToKTV {
		hh, ok := parse(0, "hours", 23)
		if !ok {
			return c
		}
		mm, ok := parse(1, "minutes", 59)
		if !ok {
			return c
		}
		ss, ok := parse(2, "seconds", 59)
		if !ok {
			return c
		}
		ms, ok := parse(3, "milliseconds", 999)
		if !ok {
			return c
		}
		kt := ktv.FromHMS(hh, mm, ss, ms)
		h, m, s, mst := kt.ToHMS()
		c.result = fmt.Sprintf(
			"Kaktovik: %s  (%s)\nBack to normal: %02d:%02d:%02d.%03d",
			kt.Spaced(), kt.Dotted(), h, m, s, mst,
		)
	} else {
		i, ok := parse(0, "ikarraq", 19)
		if !ok {
			return c
		}
		m, ok := parse(1, "mein", 19)
		if !ok {
			return c
		}
		t, ok := parse(2, "tick", 19)
		if !ok {
			return c
		}
		k, ok := parse(3, "kick", 19)
		if !ok {
			return c
		}
		kt := ktv.Time{Ikarraq: i, Mein: m, Tick: t, Kick: k}
		hh, mm, ss, ms := kt.ToHMS()
		c.result = fmt.Sprintf(
			"Normal: %02d:%02d:%02d.%03d\nKaktovik: %s",
			hh, mm, ss, ms, kt.Spaced(),
		)
	}
	return c
}

func (c convertModel) view(width int) string {
	var sb strings.Builder

	sb.WriteString(stylePanelTitle.Render("TIME CONVERTER"))
	sb.WriteString("\n\n")

	modeLabel := "Normal → Kaktovik"
	if c.mode == ktvToNormal {
		modeLabel = "Kaktovik → Normal"
	}
	sb.WriteString(styleValue.Render("Mode: ") + styleAccent.Render(modeLabel))
	sb.WriteString(styleHelp.Render("  (r to toggle)"))
	sb.WriteString("\n\n")

	if c.mode == normalToKTV {
		labels := []string{"Hours (0–23):", "Minutes (0–59):", "Seconds (0–59):", "Milliseconds (0–999):"}
		for i, inp := range c.inputs {
			focusMark := "  "
			if i == c.focus {
				focusMark = "> "
			}
			sb.WriteString(fmt.Sprintf("%s%s %s\n", focusMark, styleLabel.Render(labels[i]), inp.View()))
		}
	} else {
		labels := []string{"Ikarraq (0–19):", "Mein (0–19):", "Tick (0–19):", "Kick (0–19):"}
		for i, inp := range c.inputs {
			focusMark := "  "
			if i == c.focus {
				focusMark = "> "
			}
			sb.WriteString(fmt.Sprintf("%s%s %s\n", focusMark, styleLabel.Render(labels[i]), inp.View()))
		}
	}

	sb.WriteString("\n")

	if c.err != "" {
		sb.WriteString(styleError.Render("Error: " + c.err))
		sb.WriteString("\n")
	}
	if c.result != "" {
		sb.WriteString(styleSuccess.Render(c.result))
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString(styleHelp.Render("Tab/↑↓ move · Enter convert · r toggle mode"))
	return sb.String()
}

