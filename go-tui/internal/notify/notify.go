// Package notify wraps NixOS/systemd desktop notification tools.
package notify

import (
	"os/exec"
)

// Send sends a desktop notification using notify-send (libnotify).
func Send(title, body string) error {
	return exec.Command("notify-send",
		"--app-name=kaktovik",
		"--urgency=normal",
		title,
		body,
	).Run()
}

// SendUrgent sends a high-urgency desktop notification.
func SendUrgent(title, body string) error {
	return exec.Command("notify-send",
		"--app-name=kaktovik",
		"--urgency=critical",
		title,
		body,
	).Run()
}

// PlaySound attempts to play the system bell / a short sound using available NixOS tools.
// Tries pw-play (PipeWire), paplay (PulseAudio), then aplay (ALSA) in order.
func PlaySound() {
	// Try terminal bell as universal fallback first; try audio players in background.
	_ = exec.Command("bash", "-c",
		`pw-play /usr/share/sounds/freedesktop/stereo/complete.oga 2>/dev/null ||
		 paplay /usr/share/sounds/freedesktop/stereo/complete.oga 2>/dev/null ||
		 aplay /usr/share/sounds/alsa/Front_Center.wav 2>/dev/null ||
		 printf '\a'`).Start()
}
