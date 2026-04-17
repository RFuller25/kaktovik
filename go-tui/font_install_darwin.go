//go:build darwin

package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func fontInstallPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "Fonts", "KaktovikNumerals.ttf"), nil
}

func refreshFontCache(_ string) {
	fmt.Println("Font installed. Restart your terminal or applications to use it.")
}
