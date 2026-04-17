# Kaktovik Enhancements Design

**Date:** 2026-04-17
**Scope:** Four feature areas: notifications/sounds/settings, alarm input capture fix, large clock glyph mode, embedded font portability.

---

## A — Sounds, Notifications, Terminal Attention, Settings Screen

### Terminal Attention
When a timer or alarm fires, write to stdout in order:
1. `\033]777;urgent\007` — dunst/xterm urgency escape (raises window, flashes taskbar on supported terminals)
2. `\a` — BEL fallback (universal; triggers taskbar blink on most systems)

Both are written unconditionally (the escape is silently ignored by terminals that don't support it).

### Dunst Notifications
`notify.go`'s `SendUrgent` is upgraded to pass:
- `--icon=<path>` if `Config.NotifyIcon` is set
- `--hint=string:category:x-kaktovik` for dunst rule matching
- `--expire-time=0` for critical alarms (persists until dismissed)
- `--urgency=<level>` driven by `Config.NotifyUrgency`

### Sound
`PlaySound()` reads `Config.SoundFile`. If non-empty, that path is used for pw-play/paplay/aplay. If empty, falls back to the current system-sound chain. Sound can be toggled via `Config.SoundEnabled`.

### Config System
New package `internal/config`. Struct:

```go
type Config struct {
    SoundEnabled    bool   `json:"sound_enabled"`
    SoundFile       string `json:"sound_file"`
    NotifyUrgency   string `json:"notify_urgency"`   // "normal" | "critical"
    NotifyIcon      string `json:"notify_icon"`
    Theme           string `json:"theme"`             // "default" | "nord" | "solarized"
    ClockMode       string `json:"clock_mode"`        // "glyph" | "ascii"
    Timezone        string `json:"timezone"`          // IANA name or "" for local
}
```

Stored as JSON at `~/.config/kaktovik/config.json`. Loaded at startup in `main.go` and threaded into `ui.Options`. Saved immediately on any settings change (no explicit save button). JSON is used over TOML to avoid adding a new vendored dependency.

### Settings Tab
New `TabSettings` added as the 6th tab (key `s`). Renders a simple form:
- Each config field gets a labeled row
- Arrow keys / Tab move between fields
- Toggle fields (bool, enum) cycle values with Enter or Space
- String fields use `textinput.Model`
- A `[S]ettings` entry is added to `tabNames`
- When the settings tab is active, the global `s` hotkey is suppressed (same pattern as alarm's `IsCapturingInput`)

Themes are implemented as swappable lipgloss color sets loaded from a lookup table in `styles.go`. Three initial presets: default (current blue/gold), Nord, Solarized Dark.

---

## B — Alarm Form: Key Capture and Enter Fix

### Problem
1. Keys `c`, `v`, `t`, `w` are handled in `model.go` before messages reach `alarmModel`, so they can never be typed into the label field.
2. `nextFocus(3, false, false)` returns `0` (wraps), so Enter on the label field never triggers save — the alarm cannot be submitted.

### Fix: IsCapturingInput
```go
func (m alarmModel) IsCapturingInput() bool {
    return m.mode == alarmAdd
}
```

In `model.go`, the `c/v/t/w/a` cases are wrapped:
```go
if !m.alarm.IsCapturingInput() {
    switch msg.String() {
    case "c": ...
    case "v": ...
    // etc.
    }
}
```

The same guard is applied to `q` / `ctrl+c` (so Ctrl+C doesn't quit mid-entry).

### Fix: Enter on last field saves
`nextFocus` is changed so that when focus is on the last meaningful field, it returns the current index (i.e. `next == m.focus`), triggering `saveAlarm()`:
- Normal mode: last field is index 3 → `nextFocus(3, false, false)` returns `3`
- KTV mode: last field is index 3 → `nextFocus(3, true, false)` returns `3`

Tab/down still wraps to 0 (for navigation); Enter on the last field saves.

---

## C — Large Unicode Glyph Clock Mode

### Two modes
Controlled by `Config.ClockMode`:
- `"glyph"` (new default): Each Kaktovik digit is rendered as a large cell with the Unicode character prominently centred inside a lipgloss box border, with generous vertical padding. Cell width ~16 chars. The four cells are laid out horizontally as now.
- `"ascii"` (current default, becomes fallback): Existing `renderBigDigits` behaviour — Unicode char + number below, no box borders.

### Glyph mode rendering
Each digit cell:
```
┌──────────────┐
│              │
│     𝋃        │
│              │
│  ikarraq (3) │
└──────────────┘
```
The Unicode char is styled `styleBigTime` (bold, gold) and the label row below is `styleNormalTime`. The box border uses `lipgloss.RoundedBorder()` with accent colour.

### ASCII art fallback
For users whose terminal font doesn't include the Kaktovik block (U+1D2C0–U+1D2D3), the `"ascii"` mode remains available and works with any monospace font. The mode can be toggled with `g` while on the clock tab, and is persisted to config.

### Help line update
The clock help line gains: `g toggle glyph/ascii`.

---

## D — Embedded Font, Cross-Platform Install

### Embedding
A Kaktovik Numerals `.ttf` (sourced from a freely licensed font covering U+1D2C0–U+1D2D3, e.g. Noto Sans Math or a dedicated Kaktovik font under OFL) is placed at `go-tui/assets/KaktovikNumerals.ttf` and embedded:

```go
//go:embed assets/KaktovikNumerals.ttf
var kaktovikFont []byte
```

### Install subcommand
`kaktovik install-font` extracts and installs the font:
- **Linux:** `~/.local/share/fonts/KaktovikNumerals.ttf`, then `fc-cache -f ~/.local/share/fonts`
- **macOS:** `~/Library/Fonts/KaktovikNumerals.ttf`
- **Windows:** `%LOCALAPPDATA%\Microsoft\Windows\Fonts\KaktovikNumerals.ttf`, then registers via a raw `syscall` call to `AddFontResourceEx` (no new vendored dep required — uses `syscall` from stdlib)

The command prints clear instructions to restart the terminal after install.

### Nix flake update
The flake's `devShell` gains a comment pointing to `kaktovik install-font`. No other flake changes are needed — the font is bundled in the binary, so NixOS users get it automatically via the existing package.

---

## File Change Summary

| File | Change |
|------|--------|
| `internal/config/config.go` | New — config struct, load/save |
| `internal/notify/notify.go` | Updated — terminal attention, dunst hints, config-driven sound |
| `internal/ui/settings.go` | New — settings tab model |
| `internal/ui/model.go` | Updated — TabSettings, IsCapturingInput guard, `s` hotkey |
| `internal/ui/clock.go` | Updated — glyph mode, `g` toggle |
| `internal/ui/styles.go` | Updated — theme presets, apply active theme |
| `internal/ui/alarm.go` | Updated — IsCapturingInput, nextFocus Enter fix |
| `go-tui/assets/KaktovikNumerals.ttf` | New — embedded font file |
| `go-tui/main.go` | Updated — load config, install-font subcommand |
| `flake.nix` | Minor — devShell comment |
