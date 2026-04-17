package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds all persistent user preferences.
type Config struct {
	SoundEnabled  bool   `json:"sound_enabled"`
	SoundFile     string `json:"sound_file"`     // empty = system default
	NotifyUrgency string `json:"notify_urgency"` // "normal" | "critical"
	NotifyIcon    string `json:"notify_icon"`    // path, empty = none
	Theme         string `json:"theme"`          // "default" | "nord" | "solarized"
	ClockMode     string `json:"clock_mode"`     // "glyph" | "ascii"
	Timezone      string `json:"timezone"`       // IANA name, empty = local
}

// ThemeNames lists valid theme preset names.
var ThemeNames = []string{"default", "nord", "solarized"}

// Default returns the out-of-the-box config.
func Default() Config {
	return Config{
		SoundEnabled:  true,
		SoundFile:     "",
		NotifyUrgency: "critical",
		NotifyIcon:    "",
		Theme:         "default",
		ClockMode:     "glyph",
		Timezone:      "",
	}
}

// Path returns the path to the config file, honouring XDG_CONFIG_HOME.
func Path() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return filepath.Join(dir, "kaktovik", "config.json")
}

// Load reads the config from disk, returning defaults if the file is absent.
func Load() (Config, error) {
	cfg := Default()
	data, err := os.ReadFile(Path())
	if os.IsNotExist(err) {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Default(), err
	}
	return cfg, nil
}

// Save writes the config to disk, creating parent directories as needed.
func Save(cfg Config) error {
	p := Path()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o644)
}
