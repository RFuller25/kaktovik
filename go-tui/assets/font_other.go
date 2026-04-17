//go:build !windows

package assets

import (
	"os"
	"path/filepath"
)

// AddFontFile writes the embedded font to destPath.
// On non-Windows platforms the caller is responsible for running fc-cache
// or equivalent after installation.
func AddFontFile(destPath string) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(destPath, KaktovikFont, 0o644)
}
