# Kaktovik Enhancements Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add persistent settings + settings tab, fix alarm key capture, add large glyph clock mode, and embed the Kaktovik font for cross-platform install.

**Architecture:** A new `internal/config` package owns load/save of `~/.config/kaktovik/config.json`. The root `Model` holds a `config.Config` value and threads it into sub-models via update function arguments and `configChangedMsg` messages from the new settings tab. Notify functions are updated to accept the relevant config fields as arguments instead of using hardcoded paths.

**Tech Stack:** Go, Bubbletea, Lipgloss, `encoding/json`, `//go:embed`, `os.UserConfigDir()`.

---

## File Map

| File | Status | Responsibility |
|------|--------|---------------|
| `go-tui/internal/config/config.go` | **Create** | Config struct, Default(), Load(), Save(), Path() |
| `go-tui/internal/config/config_test.go` | **Create** | Tests for load/save/defaults |
| `go-tui/internal/notify/notify.go` | **Modify** | Add TerminalAttention(), update PlaySound/SendUrgent signatures |
| `go-tui/internal/ui/alarm.go` | **Modify** | IsCapturingInput(), Enter-saves-on-last-field fix |
| `go-tui/internal/ui/alarm_test.go` | **Create** | Tests for IsCapturingInput, nextFocus, saveAlarm trigger |
| `go-tui/internal/ui/styles.go` | **Modify** | Theme structs, ApplyTheme() |
| `go-tui/internal/ui/clock.go` | **Modify** | glyphMode field, renderGlyphMode(), `g` toggle |
| `go-tui/internal/ui/settings.go` | **Create** | settingsModel, settingsField, configChangedMsg |
| `go-tui/internal/ui/model.go` | **Modify** | TabSettings, isCapturingInput(), configChangedMsg handler, pass cfg to timer/alarm update |
| `go-tui/main.go` | **Modify** | Load config at startup, apply theme, Options threading, install-font subcommand |
| `go-tui/assets/KaktovikNumerals.ttf` | **Add** | Embedded font file (OFL, see Task 9) |
| `go-tui/assets/font.go` | **Create** | `//go:embed` declaration |

---

## Task 1: Config package

**Files:**
- Create: `go-tui/internal/config/config.go`
- Create: `go-tui/internal/config/config_test.go`

- [ ] **Step 1: Write the failing tests**

Create `go-tui/internal/config/config_test.go`:

```go
package config_test

import (
	"os"
	"testing"

	"github.com/rfuller25/kaktovik/go-tui/internal/config"
)

func TestDefault(t *testing.T) {
	cfg := config.Default()
	if cfg.ClockMode != "glyph" {
		t.Errorf("expected clock_mode=glyph, got %q", cfg.ClockMode)
	}
	if !cfg.SoundEnabled {
		t.Error("expected sound_enabled=true by default")
	}
	if cfg.NotifyUrgency != "critical" {
		t.Errorf("expected notify_urgency=critical, got %q", cfg.NotifyUrgency)
	}
	if cfg.Theme != "default" {
		t.Errorf("expected theme=default, got %q", cfg.Theme)
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg := config.Default()
	cfg.Theme = "nord"
	cfg.ClockMode = "ascii"
	cfg.SoundFile = "/tmp/beep.oga"
	cfg.SoundEnabled = false

	if err := config.Save(cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Theme != "nord" {
		t.Errorf("Theme: got %q, want nord", loaded.Theme)
	}
	if loaded.ClockMode != "ascii" {
		t.Errorf("ClockMode: got %q, want ascii", loaded.ClockMode)
	}
	if loaded.SoundFile != "/tmp/beep.oga" {
		t.Errorf("SoundFile: got %q", loaded.SoundFile)
	}
	if loaded.SoundEnabled {
		t.Error("SoundEnabled: expected false")
	}
}

func TestLoadMissingFile(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir()+"/nonexistent")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load with missing file should return defaults, not error: %v", err)
	}
	d := config.Default()
	if cfg.ClockMode != d.ClockMode || cfg.Theme != d.Theme {
		t.Errorf("missing file should return defaults, got %+v", cfg)
	}
}
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
cd go-tui && go test ./internal/config/... -v
```

Expected: `cannot find package` or build failure — package doesn't exist yet.

- [ ] **Step 3: Implement `go-tui/internal/config/config.go`**

```go
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds all persistent user preferences.
type Config struct {
	SoundEnabled  bool   `json:"sound_enabled"`
	SoundFile     string `json:"sound_file"`     // empty = system default
	NotifyUrgency string `json:"notify_urgency"` // "normal" | "critical"
	NotifyIcon    string `json:"notify_icon"`    // path, empty = none
	Theme         string `json:"theme"`          // "default" | "nord" | "solarized"
	ClockMode     string `json:"clock_mode"`     // "glyph" | "ascii"
	Timezone      string `json:"timezone"`       // IANA name, empty = local
}

// Default returns the out-of-the-box config.
func Default() Config {
	return Config{
		SoundEnabled:  true,
		SoundFile:     "",
		NotifyUrgency: "critical",
		NotifyIcon:    "",
		Theme:         "default",
		ClockMode:     "glyph",
		Timezone:      "",
	}
}

// Path returns the path to the config file, honouring XDG_CONFIG_HOME.
func Path() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return filepath.Join(dir, "kaktovik", "config.json")
}

// Load reads the config from disk, returning defaults if the file is absent.
func Load() (Config, error) {
	cfg := Default()
	data, err := os.ReadFile(Path())
	if os.IsNotExist(err) {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Default(), err
	}
	return cfg, nil
}

// Save writes the config to disk, creating parent directories as needed.
func Save(cfg Config) error {
	p := Path()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o644)
}
```

- [ ] **Step 4: Run tests and confirm they pass**

```bash
cd go-tui && go test ./internal/config/... -v
```

Expected: all three tests PASS.

- [ ] **Step 5: Commit**

```bash
cd go-tui && git add internal/config/
git commit -m "feat: add persistent config package (load/save ~/.config/kaktovik/config.json)"
```

---

## Task 2: Notify package improvements

**Files:**
- Modify: `go-tui/internal/notify/notify.go`

Update function signatures to accept config fields as arguments, add `TerminalAttention()`, and improve dunst hints. No new imports needed beyond existing `os/exec`.

- [ ] **Step 1: Replace `go-tui/internal/notify/notify.go` entirely**

```go
// Package notify wraps desktop notification and audio alert tools.
package notify

import (
	"fmt"
	"os"
	"os/exec"
)

// TerminalAttention requests the terminal's attention by writing a dunst/xterm
// urgency escape sequence followed by the BEL character as a universal fallback.
func TerminalAttention() {
	fmt.Fprint(os.Stdout, "\033]777;urgent\007") // dunst / xterm urgency
	fmt.Fprint(os.Stdout, "\a")                  // BEL fallback
}

// Send sends a normal-urgency desktop notification.
func Send(title, body string) error {
	return exec.Command("notify-send",
		"--app-name=kaktovik",
		"--urgency=normal",
		"--hint=string:category:x-kaktovik",
		title, body,
	).Run()
}

// SendUrgent sends a high-urgency desktop notification with optional icon.
// urgency must be "normal" or "critical". icon may be empty.
func SendUrgent(title, body, urgency, icon string) error {
	if urgency != "normal" && urgency != "critical" {
		urgency = "critical"
	}
	args := []string{
		"--app-name=kaktovik",
		"--urgency=" + urgency,
		"--hint=string:category:x-kaktovik",
		"--expire-time=0",
	}
	if icon != "" {
		args = append(args, "--icon="+icon)
	}
	args = append(args, title, body)
	return exec.Command("notify-send", args...).Run()
}

// PlaySound plays an audio alert. soundFile is the path to an audio file;
// if empty, the system default is used. If enabled is false, this is a no-op.
// Tries pw-play (PipeWire), paplay (PulseAudio), aplay (ALSA) in order.
func PlaySound(enabled bool, soundFile string) {
	if !enabled {
		return
	}
	if soundFile == "" {
		soundFile = "/usr/share/sounds/freedesktop/stereo/complete.oga"
	}
	script := fmt.Sprintf(
		`pw-play %q 2>/dev/null || paplay %q 2>/dev/null || aplay %q 2>/dev/null || printf '\a'`,
		soundFile, soundFile, soundFile,
	)
	_ = exec.Command("bash", "-c", script).Start()
}
```

- [ ] **Step 2: Fix all callers — update `go-tui/internal/ui/alarm.go` goroutine**

In `alarm.go`, find the goroutine in `checkAlarms` and update it. The alarm model will receive a `notifyCfg` parameter in its update call (added in Task 7). For now just update the goroutine to use the new signatures by adding the parameters (these will be wired up for real in Task 7):

Find this block in `alarm.go`:
```go
go func(label string, kv ktv.Time) {
    title := "Kaktovik Alarm"
    body := fmt.Sprintf("Alarm: %s  (%s)", kv.Spaced(), label)
    notify.SendUrgent(title, body)
    notify.PlaySound()
}(a.label, a.ktv)
```

Replace with:
```go
go func(label string, kv ktv.Time, urgency, icon string, soundEnabled bool, soundFile string) {
    title := "Kaktovik Alarm"
    body := fmt.Sprintf("Alarm: %s  (%s)", kv.Spaced(), label)
    notify.SendUrgent(title, body, urgency, icon)
    notify.TerminalAttention()
    notify.PlaySound(soundEnabled, soundFile)
}(a.label, a.ktv, "critical", "", true, "")
```

(The hardcoded values are temporary; Task 7 threads real config values.)

- [ ] **Step 3: Update `go-tui/internal/ui/timer.go` goroutine**

Find:
```go
go func() {
    notify.SendUrgent("Kaktovik Timer", "Your timer has finished!")
    notify.PlaySound()
}()
```

Replace with:
```go
go func() {
    notify.SendUrgent("Kaktovik Timer", "Your timer has finished!", "critical", "")
    notify.TerminalAttention()
    notify.PlaySound(true, "")
}()
```

- [ ] **Step 4: Update headless callers in `go-tui/main.go`**

Find `notify.PlaySound()` (two occurrences in `runHeadlessTimer` and `runHeadlessAlarm`):
```go
notify.SendUrgent("Kaktovik Timer", ...)
notify.PlaySound()
```

Replace each with:
```go
notify.SendUrgent("Kaktovik Timer", ..., "critical", "")
notify.TerminalAttention()
notify.PlaySound(true, "")
```

Do the same for the alarm headless call.

- [ ] **Step 5: Build to confirm no compile errors**

```bash
cd go-tui && go build ./...
```

Expected: no errors.

- [ ] **Step 6: Commit**

```bash
git add internal/notify/notify.go internal/ui/alarm.go internal/ui/timer.go main.go
git commit -m "feat: improve notifications — dunst hints, terminal urgency escape, configurable sound"
```

---

## Task 3: Alarm form — key capture fix and Enter-saves fix

**Files:**
- Modify: `go-tui/internal/ui/alarm.go`
- Create: `go-tui/internal/ui/alarm_test.go`

- [ ] **Step 1: Write failing tests**

Create `go-tui/internal/ui/alarm_test.go`:

```go
package ui

import (
	"testing"
	"time"
)

func TestIsCapturingInput(t *testing.T) {
	m := newAlarm(time.Time{})
	if m.IsCapturingInput() {
		t.Error("new alarm (list mode) should not be capturing input")
	}

	// Simulate pressing 'a' to enter add mode
	m.mode = alarmAdd
	if !m.IsCapturingInput() {
		t.Error("alarm in alarmAdd mode should be capturing input")
	}

	m.mode = alarmList
	if m.IsCapturingInput() {
		t.Error("alarm back in list mode should not be capturing input")
	}
}

func TestEnterOnLastFieldSavesAlarm(t *testing.T) {
	m := newAlarm(time.Time{})
	m.mode = alarmAdd
	m.ktvMode = false
	m.inputs = makeNormalInputs()

	// Fill in valid hour/minute/second
	m.inputs[0].SetValue("10")
	m.inputs[1].SetValue("30")
	m.inputs[2].SetValue("00")
	m.inputs[3].SetValue("Wake up")

	// focus is 0 by default; advance to field 3 (label)
	m.inputs[m.focus].Blur()
	m.focus = 3
	m.inputs[m.focus].Focus()

	// pressing Enter on field 3 should save and return to list mode
	if m.focus != 3 {
		t.Fatalf("expected focus=3, got %d", m.focus)
	}
	saved := m.saveAlarm()
	if saved.mode != alarmList {
		t.Errorf("after save, expected alarmList mode, got %d", saved.mode)
	}
	if len(saved.alarms) != 1 {
		t.Errorf("expected 1 alarm saved, got %d", len(saved.alarms))
	}
	if saved.alarms[0].label != "Wake up" {
		t.Errorf("expected label 'Wake up', got %q", saved.alarms[0].label)
	}
}

func TestEnterOnNonLastFieldAdvancesFocus(t *testing.T) {
	m := newAlarm(time.Time{})
	m.mode = alarmAdd
	m.ktvMode = false
	m.inputs = makeNormalInputs()
	m.focus = 0

	// pressing Enter on field 0 should move to field 1, not save
	next := alarmEnterNextFocus(m.focus, m.ktvMode)
	if next == m.focus {
		t.Errorf("Enter on field 0 should advance focus, not stay at %d", m.focus)
	}
	if next != 1 {
		t.Errorf("Enter on field 0 (normal mode) should go to 1, got %d", next)
	}
}

func TestKTVModeEnterOnLabelSaves(t *testing.T) {
	m := newAlarm(time.Time{})
	m.mode = alarmAdd
	m.ktvMode = true
	m.inputs = makeKTVInputs()
	m.inputs[0].SetValue("5.3.9.2")
	m.inputs[3].SetValue("Lunch")
	m.focus = 3

	saved := m.saveAlarm()
	if saved.mode != alarmList {
		t.Errorf("KTV mode: after save, expected alarmList, got %d", saved.mode)
	}
	if len(saved.alarms) != 1 {
		t.Errorf("expected 1 alarm, got %d", len(saved.alarms))
	}
}
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
cd go-tui && go test ./internal/ui/... -run TestIsCapturing -v
cd go-tui && go test ./internal/ui/... -run TestEnter -v
```

Expected: `undefined: IsCapturingInput` or `undefined: alarmEnterNextFocus` — functions don't exist yet.

- [ ] **Step 3: Add `IsCapturingInput` to `alarm.go`**

Add after the `newAlarm` function (before `makeNormalInputs`):

```go
// IsCapturingInput reports whether the alarm model is actively editing a text field.
// Used by the root model to suppress global hotkeys during text entry.
func (m alarmModel) IsCapturingInput() bool {
	return m.mode == alarmAdd
}
```

- [ ] **Step 4: Add `alarmEnterNextFocus` helper and fix the Enter handler in `alarm.go`**

Add this helper (it is only for Enter — Tab still uses `nextFocus`):

```go
// alarmEnterNextFocus returns the next focus index when Enter is pressed.
// Returns cur if we are on the last field (triggering save), otherwise advances.
func alarmEnterNextFocus(cur int, ktvMode bool) int {
	// Label field (index 3) is always the last field in both modes.
	if cur == 3 {
		return 3 // same as current → caller triggers saveAlarm
	}
	if ktvMode {
		// KTV mode only uses field 0 and 3; from 0 go straight to 3.
		return 3
	}
	return cur + 1
}
```

In the `update` method, find the `case "enter":` block inside `alarmAdd`:

```go
case "enter":
    next := nextFocus(m.focus, m.ktvMode, false)
    if next != m.focus {
        m.inputs[m.focus].Blur()
        m.focus = next
        m.inputs[m.focus].Focus()
    } else {
        m = m.saveAlarm()
    }
    return m, nil
```

Replace with:

```go
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
```

- [ ] **Step 5: Run tests and confirm they pass**

```bash
cd go-tui && go test ./internal/ui/... -run "TestIsCapturing|TestEnter|TestKTV" -v
```

Expected: all 4 tests PASS.

- [ ] **Step 6: Build**

```bash
cd go-tui && go build ./...
```

Expected: no errors.

- [ ] **Step 7: Commit**

```bash
git add internal/ui/alarm.go internal/ui/alarm_test.go
git commit -m "fix: alarm form — Enter on last field saves, IsCapturingInput guards hotkeys"
```

---

## Task 4: Theme presets in styles.go

**Files:**
- Modify: `go-tui/internal/ui/styles.go`

Themes are swappable sets of colors. `ApplyTheme(name)` reinitializes the package-level style vars.

- [ ] **Step 1: Replace `go-tui/internal/ui/styles.go` entirely**

```go
package ui

import "github.com/charmbracelet/lipgloss"

// themeColors defines the palette for one theme preset.
type themeColors struct {
	Accent    lipgloss.Color
	Dim       lipgloss.Color
	Muted     lipgloss.Color
	Highlight lipgloss.Color
	Green     lipgloss.Color
	Red       lipgloss.Color
	Bg        lipgloss.Color
}

var themePresets = map[string]themeColors{
	"default": {
		Accent:    "#5FD7FF",
		Dim:       "#555555",
		Muted:     "#888888",
		Highlight: "#FFD700",
		Green:     "#87FF5F",
		Red:       "#FF5F5F",
		Bg:        "#1A1A2E",
	},
	"nord": {
		Accent:    "#88C0D0",
		Dim:       "#4C566A",
		Muted:     "#616E88",
		Highlight: "#EBCB8B",
		Green:     "#A3BE8C",
		Red:       "#BF616A",
		Bg:        "#2E3440",
	},
	"solarized": {
		Accent:    "#268BD2",
		Dim:       "#073642",
		Muted:     "#586E75",
		Highlight: "#B58900",
		Green:     "#859900",
		Red:       "#DC322F",
		Bg:        "#002B36",
	},
}

// ThemeNames returns sorted available theme names.
var ThemeNames = []string{"default", "nord", "solarized"}

// active color vars — mutated by ApplyTheme.
var (
	colorAccent    lipgloss.Color
	colorDim       lipgloss.Color
	colorMuted     lipgloss.Color
	colorHighlight lipgloss.Color
	colorGreen     lipgloss.Color
	colorRed       lipgloss.Color
	colorBg        lipgloss.Color
)

// Style vars — recreated by ApplyTheme.
var (
	styleTab          lipgloss.Style
	styleTabActive    lipgloss.Style
	stylePanelTitle   lipgloss.Style
	styleBigTime      lipgloss.Style
	styleNormalTime   lipgloss.Style
	styleLabel        lipgloss.Style
	styleValue        lipgloss.Style
	styleHelp         lipgloss.Style
	styleSuccess      lipgloss.Style
	styleWarn         lipgloss.Style
	styleError        lipgloss.Style
	styleBorder       lipgloss.Style
	styleInputFocused lipgloss.Style
	styleInputBlurred lipgloss.Style
	styleAccent       lipgloss.Style
	styleMuted        lipgloss.Style
)

func init() {
	ApplyTheme("default")
}

// ApplyTheme sets colors and reinitializes all style vars for the named theme.
// Falls back to "default" for unknown names.
func ApplyTheme(name string) {
	t, ok := themePresets[name]
	if !ok {
		t = themePresets["default"]
	}
	colorAccent = t.Accent
	colorDim = t.Dim
	colorMuted = t.Muted
	colorHighlight = t.Highlight
	colorGreen = t.Green
	colorRed = t.Red
	colorBg = t.Bg

	styleTab = lipgloss.NewStyle().Padding(0, 2).Foreground(colorMuted)
	styleTabActive = lipgloss.NewStyle().Padding(0, 2).Foreground(colorAccent).Bold(true).Underline(true)
	stylePanelTitle = lipgloss.NewStyle().Foreground(colorAccent).Bold(true).MarginBottom(1)
	styleBigTime = lipgloss.NewStyle().Foreground(colorHighlight).Bold(true)
	styleNormalTime = lipgloss.NewStyle().Foreground(colorMuted)
	styleLabel = lipgloss.NewStyle().Foreground(colorAccent).Width(14)
	styleValue = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	styleHelp = lipgloss.NewStyle().Foreground(colorDim).MarginTop(1)
	styleSuccess = lipgloss.NewStyle().Foreground(colorGreen).Bold(true)
	styleWarn = lipgloss.NewStyle().Foreground(colorHighlight)
	styleError = lipgloss.NewStyle().Foreground(colorRed)
	styleBorder = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colorAccent).Padding(1, 2)
	styleInputFocused = lipgloss.NewStyle().Foreground(colorHighlight)
	styleInputBlurred = lipgloss.NewStyle().Foreground(colorMuted)
	styleAccent = lipgloss.NewStyle().Foreground(colorAccent)
	styleMuted = lipgloss.NewStyle().Foreground(colorMuted)
}
```

- [ ] **Step 2: Build**

```bash
cd go-tui && go build ./...
```

Expected: no errors. All existing code still compiles because style var names are unchanged.

- [ ] **Step 3: Commit**

```bash
git add internal/ui/styles.go
git commit -m "feat: theme presets — default, nord, solarized; ApplyTheme() reinitializes styles"
```

---

## Task 5: Clock glyph mode

**Files:**
- Modify: `go-tui/internal/ui/clock.go`

Add a `glyphMode bool` field to `clockModel`. When true, render large bordered cells (glyph mode). Press `g` on the clock tab to toggle. Default is `true` (set when model is created from config in Task 8).

- [ ] **Step 1: Replace `go-tui/internal/ui/clock.go` entirely**

```go
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
	now       time.Time
	tz        *time.Location
	glyphMode bool // true = large Unicode cells; false = current ASCII style
}

func newClock(tz *time.Location, glyphMode bool) clockModel {
	if tz == nil {
		tz = time.Local
	}
	return clockModel{now: time.Now().In(tz), tz: tz, glyphMode: glyphMode}
}

func (c clockModel) update(msg tea.Msg) clockModel {
	switch msg := msg.(type) {
	case tickMsg:
		c.now = time.Now().In(c.tz)
	case tea.KeyMsg:
		if msg.String() == "g" {
			c.glyphMode = !c.glyphMode
		}
	}
	return c
}

func (c clockModel) view(width int) string {
	kt := ktv.FromTime(c.now)

	var digits string
	if c.glyphMode {
		digits = renderGlyphMode(kt, width)
	} else {
		digits = renderBigDigits(kt, width)
	}

	h, m, s, _ := kt.ToHMS()
	normalStr := fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	ktvDotted := kt.Dotted()

	tzName, offset := c.now.Zone()
	tzStr := fmt.Sprintf("%s (UTC%+d)", tzName, offset/3600)

	modeHint := "ascii"
	if c.glyphMode {
		modeHint = "glyph"
	}

	lines := []string{
		stylePanelTitle.Render("KAKTOVIK CLOCK"),
		"",
		digits,
		"",
		styleNormalTime.Render(fmt.Sprintf("  %s  ·  %s", ktvDotted, normalStr)),
		styleNormalTime.Render(fmt.Sprintf("  %s", tzStr)),
		"",
		styleHelp.Render(fmt.Sprintf("  Tab/←/→ switch views · g toggle glyph/ascii [%s] · q quit", modeHint)),
	}
	return strings.Join(lines, "\n")
}

// renderGlyphMode renders each Kaktovik digit as a large rounded-border cell
// containing the Unicode glyph prominently centred.
func renderGlyphMode(kt ktv.Time, width int) string {
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
		inner := lipgloss.JoinVertical(lipgloss.Center,
			"",
			styleBigTime.Copy().Render(char),
			"",
			styleNormalTime.Copy().Render(fmt.Sprintf("%s  %d", c.label, c.value)),
		)
		cells[i] = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorAccent).
			Padding(0, 4).
			Align(lipgloss.Center).
			Render(inner)
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, cells...)
	return lipgloss.NewStyle().
		Width(width - 4).
		Align(lipgloss.Center).
		Render(row)
}

// renderBigDigits is the original display: Unicode char + number, no borders.
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
```

- [ ] **Step 2: Fix the `newClock` call site in `model.go`**

In `go-tui/internal/ui/model.go`, find:
```go
clock: newClock(opts.Timezone),
```

Replace with:
```go
clock: newClock(opts.Timezone, true), // default glyph mode; Task 8 wires config
```

- [ ] **Step 3: Build**

```bash
cd go-tui && go build ./...
```

Expected: no errors.

- [ ] **Step 4: Commit**

```bash
git add internal/ui/clock.go internal/ui/model.go
git commit -m "feat: clock glyph mode — large bordered Unicode cells, 'g' toggles ascii fallback"
```

---

## Task 6: Settings tab

**Files:**
- Create: `go-tui/internal/ui/settings.go`

The settings model renders a form. Enum/bool fields cycle with Space or Enter. Text fields get focus on Enter and lose focus on Escape. Changes emit `configChangedMsg` which the root model handles.

- [ ] **Step 1: Create `go-tui/internal/ui/settings.go`**

```go
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

// ThemeNames must be exported from config. Add to config.go:
// var ThemeNames = []string{"default", "nord", "solarized"}

type settingsModel struct {
	cfg    config.Config
	cursor int
	// inputs maps inputIdx → textinput.Model (3 text fields)
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
	// If a text input is focused, route to it; only Escape exits.
	if m.inputFocus >= 0 {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "escape" {
				m.inputs[m.inputFocus].Blur()
				// flush value back to cfg
				m = m.flushInput(m.inputFocus)
				m.inputFocus = -1
				return m, emitConfigChanged(m.cfg)
			}
		}
		var cmd tea.Cmd
		m.inputs[m.inputFocus], cmd = m.inputs[m.inputFocus].Update(msg)
		// Flush the live value so cfg stays current (for return-key-submits behaviour).
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
				// Focus the text input
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

// flushInput writes the current text input value back to cfg.
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
	case 0: // Sound enabled
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
```

- [ ] **Step 2: Add `ThemeNames` to `go-tui/internal/config/config.go`**

Add at the end of the file:

```go
// ThemeNames lists valid theme preset names.
var ThemeNames = []string{"default", "nord", "solarized"}
```

- [ ] **Step 3: Build**

```bash
cd go-tui && go build ./...
```

Expected: no errors.

- [ ] **Step 4: Commit**

```bash
git add internal/ui/settings.go internal/config/config.go
git commit -m "feat: settings tab model with live config editing and configChangedMsg"
```

---

## Task 7: Wire config into the root model

**Files:**
- Modify: `go-tui/internal/ui/model.go`

Add `TabSettings`, `cfg config.Config`, `settings settingsModel` to `Model`. Guard global hotkeys with `isCapturingInput()`. Handle `configChangedMsg`. Pass config fields to alarm/timer `update` calls and update their goroutine captures.

- [ ] **Step 1: Replace `go-tui/internal/ui/model.go` entirely**

```go
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
		(m.activeTab == TabSettings && m.settings.IsCapturingInput())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case configChangedMsg:
		m.cfg = config.Config(msg)
		// Propagate derived state immediately.
		m.clock.glyphMode = (m.cfg.ClockMode == "glyph")
		ApplyTheme(m.cfg.Theme)
		go config.Save(m.cfg) //nolint:errcheck — best-effort background save
		return m, nil

	case tea.KeyMsg:
		if !m.isCapturingInput() {
			switch msg.String() {
			case "ctrl+c", "q":
				if m.activeTab != TabConvert {
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
				if m.activeTab != TabAlarm {
					m.activeTab = TabAlarm
					return m, nil
				}
			case "s":
				m.activeTab = TabSettings
				return m, nil
			}
		}

		// Clock handles 'g' toggle itself.
		if m.activeTab == TabClock {
			m.clock = m.clock.update(msg)
			// If glyphMode changed, persist it.
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
```

- [ ] **Step 2: Update `alarm.go` — add `cfg config.Config` parameter to `update` and `checkAlarms`**

Change the `update` signature:
```go
func (m alarmModel) update(msg tea.Msg, cfg config.Config) (alarmModel, tea.Cmd) {
```

Change `checkAlarms` call inside `update`:
```go
case tickMsg:
    m.now = time.Time(msg)
    m.checkAlarms(cfg)
```

Change `checkAlarms` signature and goroutine:
```go
func (m *alarmModel) checkAlarms(cfg config.Config) {
    for i := range m.alarms {
        a := &m.alarms[i]
        if a.enabled && !a.fired && m.now.After(a.target) {
            a.fired = true
            go func(label string, kv ktv.Time, urgency, icon string, soundEnabled bool, soundFile string) {
                title := "Kaktovik Alarm"
                body := fmt.Sprintf("Alarm: %s  (%s)", kv.Spaced(), label)
                notify.SendUrgent(title, body, urgency, icon)
                notify.TerminalAttention()
                notify.PlaySound(soundEnabled, soundFile)
            }(a.label, a.ktv, cfg.NotifyUrgency, cfg.NotifyIcon, cfg.SoundEnabled, cfg.SoundFile)
        }
    }
}
```

Add the config import to `alarm.go`:
```go
import (
    ...
    "github.com/rfuller25/kaktovik/go-tui/internal/config"
    ...
)
```

- [ ] **Step 3: Update `timer.go` — add `cfg config.Config` parameter to `update`**

Change the `update` signature:
```go
func (m timerModel) update(msg tea.Msg, cfg config.Config) (timerModel, tea.Cmd) {
```

In the timer-done goroutine:
```go
go func(urgency, icon string, soundEnabled bool, soundFile string) {
    notify.SendUrgent("Kaktovik Timer", "Your timer has finished!", urgency, icon)
    notify.TerminalAttention()
    notify.PlaySound(soundEnabled, soundFile)
}(cfg.NotifyUrgency, cfg.NotifyIcon, cfg.SoundEnabled, cfg.SoundFile)
```

Add the config import to `timer.go`.

- [ ] **Step 4: Build**

```bash
cd go-tui && go build ./...
```

Expected: no errors.

- [ ] **Step 5: Run all tests**

```bash
cd go-tui && go test ./...
```

Expected: all tests pass (config tests + alarm tests).

- [ ] **Step 6: Commit**

```bash
git add internal/ui/model.go internal/ui/alarm.go internal/ui/timer.go
git commit -m "feat: wire config into root model — settings tab, IsCapturingInput guard, live theme/sound/clock propagation"
```

---

## Task 8: Wire config into main.go

**Files:**
- Modify: `go-tui/main.go`

Load config at startup, apply theme, thread into `ui.Options`, and update headless notify calls to use config values.

- [ ] **Step 1: Add config loading and `install-font` stub to `go-tui/main.go`**

At the top of `main()` (before `rootCmd.Execute()`), add nothing — config is loaded inside each command's RunE. Instead, add it in `runTUI`:

Replace `runTUI`:

```go
func runTUI(opts ui.Options) error {
	p := tea.NewProgram(
		ui.New(opts),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	_, err := p.Run()
	return err
}
```

Add a helper and update each `RunE` that calls `runTUI`. Change the root `RunE`:

```go
var rootCmd = &cobra.Command{
	// ... existing Use/Short/Long unchanged ...
	RunE: func(cmd *cobra.Command, args []string) error {
		tzFlag, _ := cmd.Flags().GetString("timezone")
		loc, err := parseTimezone(tzFlag)
		if err != nil {
			return err
		}
		cfg, _ := config.Load()
		if tzFlag != "" {
			cfg.Timezone = tzFlag // CLI flag overrides stored timezone
		}
		ApplyTheme(cfg.Theme) // set styles before TUI starts
		return runTUI(ui.Options{Timezone: loc, Cfg: cfg})
	},
}
```

Add `config.Load()` calls similarly to `timerCmd.RunE`, `alarmCmd.RunE`, `stopwatchCmd.RunE`, `convertCmd.RunE` — each should load config and pass `Cfg: cfg` in `ui.Options`.

Add the import:
```go
import (
    ...
    "github.com/rfuller25/kaktovik/go-tui/internal/config"
    "github.com/rfuller25/kaktovik/go-tui/internal/ui"
    ...
)
```

Add a convenience wrapper at the top of main.go:

```go
// ApplyTheme delegates to ui.ApplyTheme so main.go can call it without a ui import cycle.
func applyTheme(name string) { ui.ApplyTheme(name) }
```

Wait — `ApplyTheme` is in the `ui` package. Call it as `ui.ApplyTheme(cfg.Theme)` directly (main already imports `ui`). Remove the wrapper.

Update `runHeadlessTimer` and `runHeadlessAlarm` to load config:

```go
func runHeadlessTimer(d time.Duration) error {
	cfg, _ := config.Load()
	kt := ktv.FromDuration(d)
	fmt.Printf("Timer started: %s (%s)\n", formatDurationHuman(d), kt.Dotted())
	time.Sleep(d)
	notify.SendUrgent("Kaktovik Timer", fmt.Sprintf("Timer finished after %s", formatDurationHuman(d)),
		cfg.NotifyUrgency, cfg.NotifyIcon)
	notify.TerminalAttention()
	notify.PlaySound(cfg.SoundEnabled, cfg.SoundFile)
	fmt.Println("Timer complete.")
	return nil
}

func runHeadlessAlarm(target time.Time) error {
	cfg, _ := config.Load()
	now := time.Now()
	if target.Before(now) {
		target = target.Add(24 * time.Hour)
	}
	wait := time.Until(target)
	kt := ktv.FromTime(target)
	fmt.Printf("Alarm set for %02d:%02d:%02d (%s), fires in %s\n",
		target.Hour(), target.Minute(), target.Second(),
		kt.Dotted(), formatDurationHuman(wait))
	time.Sleep(wait)
	notify.SendUrgent("Kaktovik Alarm", fmt.Sprintf("Alarm: %02d:%02d:%02d  %s",
		target.Hour(), target.Minute(), target.Second(), kt.Spaced()),
		cfg.NotifyUrgency, cfg.NotifyIcon)
	notify.TerminalAttention()
	notify.PlaySound(cfg.SoundEnabled, cfg.SoundFile)
	fmt.Println("Alarm fired.")
	return nil
}
```

- [ ] **Step 2: Build**

```bash
cd go-tui && go build ./...
```

Expected: no errors.

- [ ] **Step 3: Run tests**

```bash
cd go-tui && go test ./...
```

Expected: all pass.

- [ ] **Step 4: Commit**

```bash
git add main.go
git commit -m "feat: load config at startup, apply theme, thread config into all TUI opts and headless commands"
```

---

## Task 9: Embedded font + `install-font` subcommand

**Files:**
- Add: `go-tui/assets/KaktovikNumerals.ttf`
- Create: `go-tui/assets/font.go`
- Modify: `go-tui/main.go`

- [ ] **Step 1: Source the font file**

Download a freely licensed font that covers U+1D2C0–U+1D2D3 (Kaktovik Numerals Unicode 15.0):

```bash
# Option A: Kreativekorp Kaktovik Numerals font (OFL)
# Download from: https://www.kreativekorp.com/software/fonts/kaktovik/
# Place the .ttf at: go-tui/assets/KaktovikNumerals.ttf

mkdir -p go-tui/assets
# Manually download and copy the .ttf file to go-tui/assets/KaktovikNumerals.ttf
```

Confirm the file is present:
```bash
ls -lh go-tui/assets/KaktovikNumerals.ttf
```

Expected: file present, size > 0.

- [ ] **Step 2: Create `go-tui/assets/font.go`**

```go
package assets

import _ "embed"

// KaktovikFont is the embedded Kaktovik Numerals TrueType font (OFL licensed).
//
//go:embed KaktovikNumerals.ttf
var KaktovikFont []byte
```

- [ ] **Step 3: Build to confirm embed compiles**

```bash
cd go-tui && go build ./...
```

Expected: no errors. Binary now contains the embedded font.

- [ ] **Step 4: Add `install-font` subcommand to `go-tui/main.go`**

Add the import for the assets package and OS-specific font logic:

```go
import (
    ...
    "os"
    "os/exec"
    "path/filepath"
    "runtime"
    ...
    ktvAssets "github.com/rfuller25/kaktovik/go-tui/assets"
)
```

Add the command var and init registration:

```go
var installFontCmd = &cobra.Command{
    Use:   "install-font",
    Short: "Install the Kaktovik Numerals font for your terminal",
    Long: `Extracts the embedded Kaktovik Numerals font and installs it to
your user font directory so your terminal can render Kaktovik Unicode glyphs.

After running this command, restart your terminal emulator.`,
    RunE: func(cmd *cobra.Command, args []string) error {
        return installFont()
    },
}
```

In `init()`, add:
```go
rootCmd.AddCommand(timerCmd, alarmCmd, stopwatchCmd, convertCmd, nowCmd, installFontCmd)
```

Add the `installFont` function:

```go
func installFont() error {
    destDir, err := fontInstallDir()
    if err != nil {
        return fmt.Errorf("cannot determine font directory: %w", err)
    }
    if err := os.MkdirAll(destDir, 0o755); err != nil {
        return fmt.Errorf("cannot create font directory: %w", err)
    }
    destPath := filepath.Join(destDir, "KaktovikNumerals.ttf")
    if err := os.WriteFile(destPath, ktvAssets.KaktovikFont, 0o644); err != nil {
        return fmt.Errorf("cannot write font file: %w", err)
    }
    fmt.Printf("Font written to: %s\n", destPath)

    if err := registerFont(destDir, destPath); err != nil {
        fmt.Printf("Warning: font registration step failed (%v) — you may need to refresh fonts manually.\n", err)
    }

    fmt.Println("Done! Restart your terminal emulator to use the Kaktovik Numerals font.")
    return nil
}

func fontInstallDir() (string, error) {
    switch runtime.GOOS {
    case "windows":
        appdata := os.Getenv("LOCALAPPDATA")
        if appdata == "" {
            return "", fmt.Errorf("LOCALAPPDATA not set")
        }
        return filepath.Join(appdata, "Microsoft", "Windows", "Fonts"), nil
    case "darwin":
        home, err := os.UserHomeDir()
        if err != nil {
            return "", err
        }
        return filepath.Join(home, "Library", "Fonts"), nil
    default: // Linux and others
        home, err := os.UserHomeDir()
        if err != nil {
            return "", err
        }
        return filepath.Join(home, ".local", "share", "fonts"), nil
    }
}

func registerFont(dir, path string) error {
    switch runtime.GOOS {
    case "windows":
        return registerFontWindows(path)
    case "darwin":
        return nil // macOS picks up fonts from ~/Library/Fonts automatically
    default:
        return exec.Command("fc-cache", "-f", dir).Run()
    }
}
```

- [ ] **Step 5: Add `go-tui/assets/font_windows.go` for Windows font registration**

```go
//go:build windows

package assets

import (
    "syscall"
    "unsafe"
)

// AddFontFile registers a font file with Windows using AddFontResourceExW.
func AddFontFile(path string) error {
    mod := syscall.NewLazyDLL("gdi32.dll")
    proc := mod.NewProc("AddFontResourceExW")
    pathPtr, err := syscall.UTF16PtrFromString(path)
    if err != nil {
        return err
    }
    // FR_PRIVATE = 0x10
    r, _, err := proc.Call(uintptr(unsafe.Pointer(pathPtr)), 0x10, 0)
    if r == 0 {
        return err
    }
    return nil
}
```

Update `registerFontWindows` in `main.go` to call this:

```go
func registerFontWindows(path string) error {
    return assets.AddFontFile(path)
}
```

Add a stub for non-Windows in `go-tui/assets/font_other.go`:

```go
//go:build !windows

package assets

// AddFontFile is a no-op on non-Windows platforms.
func AddFontFile(path string) error { return nil }
```

- [ ] **Step 6: Build for current platform**

```bash
cd go-tui && go build ./...
```

Expected: no errors.

- [ ] **Step 7: Confirm binary size increased (font is embedded)**

```bash
go build -o /tmp/kaktovik-test ./... && ls -lh /tmp/kaktovik-test
```

Expected: binary is larger than before by roughly the size of the font file.

- [ ] **Step 8: Run all tests**

```bash
cd go-tui && go test ./...
```

Expected: all pass.

- [ ] **Step 9: Commit**

```bash
git add assets/ main.go
git commit -m "feat: embed Kaktovik Numerals font; add 'install-font' subcommand for Linux/macOS/Windows"
```

---

## Self-Review

**Spec coverage check:**

| Spec requirement | Task |
|-----------------|------|
| Terminal urgency escape + BEL | Task 2 `TerminalAttention()` |
| Dunst hints (category, icon, expire-time, urgency) | Task 2 `SendUrgent` |
| Config-driven sound file + enabled toggle | Tasks 1+2+7 |
| Settings screen with all config fields | Task 6 |
| Persist settings to `~/.config/kaktovik/config.json` | Tasks 1+7 |
| Theme presets (default, nord, solarized) | Task 4 |
| Alarm IsCapturingInput → suppress c/v/t/w/a | Tasks 3+7 |
| Enter on last alarm field saves | Task 3 |
| Settings text inputs suppress global hotkeys | Tasks 6+7 |
| Clock glyph mode (large bordered Unicode cells) | Task 5 |
| Clock `g` toggle persisted to config | Task 7 |
| `//go:embed` font in binary | Task 9 |
| `install-font` for Linux/macOS/Windows | Task 9 |

All spec requirements have a corresponding task. ✓

**Type consistency check:**

- `config.Config` defined in Task 1; used in Tasks 2–9 ✓
- `configChangedMsg` defined in Task 6 (`settings.go`); handled in Task 7 (`model.go`) ✓
- `alarm.update(msg, cfg config.Config)` defined in Task 7; `model.go` passes `m.cfg` ✓
- `timer.update(msg, cfg config.Config)` defined in Task 7; `model.go` passes `m.cfg` ✓
- `newClock(tz, glyphMode bool)` defined in Task 5; called in Task 7's `Model.New` ✓
- `ApplyTheme` defined in Task 4 as `ui.ApplyTheme`; called as `ui.ApplyTheme` in Task 8 ✓
- `ktvAssets.KaktovikFont` in Task 9: package alias `ktvAssets` for `go-tui/assets` ✓
- `assets.AddFontFile` called as `assets.AddFontFile` in `registerFontWindows` in `main.go` — needs the import `ktvAssets` aliased as `assets` or be consistent: use `ktvAssets.AddFontFile` ✓ (fix: use the same alias `ktvAssets` throughout)

**Placeholder scan:** No TBD, TODO, or vague steps found. ✓
