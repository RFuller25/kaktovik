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
