// Package notify wraps desktop notification and audio alert tools.
package notify

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
// if empty, the system default is found via XDG_DATA_DIRS. If enabled is false, this is a no-op.
// Tries pw-play (PipeWire), paplay (PulseAudio), aplay (ALSA) in order.
func PlaySound(enabled bool, soundFile string) {
	if !enabled {
		return
	}
	if soundFile == "" {
		soundFile = findDefaultSoundFile()
	} else {
		soundFile = expandPath(soundFile)
	}
	if soundFile == "" {
		_ = exec.Command("bash", "-c", `printf '\a'`).Start()
		return
	}
	script := fmt.Sprintf(
		`pw-play %q 2>/dev/null || paplay %q 2>/dev/null || aplay %q 2>/dev/null || printf '\a'`,
		soundFile, soundFile, soundFile,
	)
	_ = exec.Command("bash", "-c", script).Start()
}

// expandPath resolves ~ and $VAR references in a file path so that paths like
// "~/Music/alert.wav" work when stored in the config file.
func expandPath(p string) string {
	if strings.HasPrefix(p, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			p = filepath.Join(home, p[2:])
		}
	}
	return os.ExpandEnv(p)
}

// findDefaultSoundFile locates the freedesktop complete.oga alert sound by searching
// XDG_DATA_DIRS, which on NixOS points at /run/current-system/sw/share rather than /usr/share.
func findDefaultSoundFile() string {
	const rel = "sounds/freedesktop/stereo/complete.oga"
	for _, dir := range strings.Split(os.Getenv("XDG_DATA_DIRS"), ":") {
		if dir == "" {
			continue
		}
		p := filepath.Join(dir, rel)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	// Fallback to standard FHS path for non-NixOS systems.
	if p := "/usr/share/" + rel; fileExists(p) {
		return p
	}
	return ""
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
