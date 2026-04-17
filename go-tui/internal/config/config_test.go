package config_test

import (
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
