//go:build windows

package assets

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

// AddFontFile writes the embedded font to destPath and registers it with GDI.
func AddFontFile(destPath string) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(destPath, KaktovikFont, 0o644); err != nil {
		return err
	}
	gdi32 := syscall.NewLazyDLL("gdi32.dll")
	addFont := gdi32.NewProc("AddFontResourceExW")
	ptr, err := syscall.UTF16PtrFromString(destPath)
	if err != nil {
		return err
	}
	r, _, _ := addFont.Call(uintptr(unsafe.Pointer(ptr)), 0, 0)
	if r == 0 {
		return fmt.Errorf("AddFontResourceExW returned 0 for %s", destPath)
	}
	return nil
}
