// Package alarmstore persists alarm entries to disk.
package alarmstore

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Alarm is the on-disk representation of a single alarm entry.
type Alarm struct {
	Label    string    `json:"label"`
	Target   time.Time `json:"target"`
	Enabled  bool      `json:"enabled"`
	Fired    bool      `json:"fired"`
	UnitName string    `json:"unit_name,omitempty"`
}

// Path returns the path to the alarms file, honouring XDG_CONFIG_HOME.
func Path() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return filepath.Join(dir, "kaktovik", "alarms.json")
}

// Load reads persisted alarms from disk. Returns nil slice (no error) when the
// file does not exist yet.
func Load() ([]Alarm, error) {
	data, err := os.ReadFile(Path())
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var alarms []Alarm
	return alarms, json.Unmarshal(data, &alarms)
}

// Save writes the alarm list to disk, creating parent directories as needed.
func Save(alarms []Alarm) error {
	p := Path()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(alarms, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o644)
}
