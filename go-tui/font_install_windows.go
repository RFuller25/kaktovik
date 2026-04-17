//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func fontInstallPath() (string, error) {
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		localAppData = filepath.Join(home, "AppData", "Local")
	}
	return filepath.Join(localAppData, "Microsoft", "Windows", "Fonts", "KaktovikNumerals.ttf"), nil
}

func refreshFontCache(destPath string) {
	// AddFontResourceExW was already called in assets.AddFontFile on Windows.
	fmt.Printf("Font registered with Windows GDI. Path: %s\n", destPath)
	fmt.Println("You may need to restart applications to pick up the new font.")
}
