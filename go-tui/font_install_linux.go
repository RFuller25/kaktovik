//go:build linux

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func fontInstallPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "fonts", "KaktovikNumerals.ttf"), nil
}

func refreshFontCache(_ string) {
	out, err := exec.Command("fc-cache", "-f").CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "fc-cache: %v\n%s\n", err, out)
		return
	}
	fmt.Println("Font cache refreshed (fc-cache -f).")
}
