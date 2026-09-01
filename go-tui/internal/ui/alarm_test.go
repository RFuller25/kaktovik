package ui

import (
	"testing"
	"time"
)

// isolateAlarmEnv redirects alarmstore's file path to a temp directory and
// stubs out the real systemd-run call, so tests never touch the user's
// actual ~/.config/kaktovik/alarms.json or schedule real systemd timers.
func isolateAlarmEnv(t *testing.T) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	orig := scheduleAlarmUnit
	scheduleAlarmUnit = func(time.Time) string { return "" }
	t.Cleanup(func() { scheduleAlarmUnit = orig })
}

func TestIsCapturingInput(t *testing.T) {
	isolateAlarmEnv(t)
	m := newAlarm(time.Time{})
	if m.IsCapturingInput() {
		t.Error("new alarm (list mode) should not be capturing input")
	}

	m.mode = alarmAdd
	if !m.IsCapturingInput() {
		t.Error("alarm in alarmAdd mode should be capturing input")
	}

	m.mode = alarmList
	if m.IsCapturingInput() {
		t.Error("alarm back in list mode should not be capturing input")
	}
}

func TestEnterOnLastFieldSavesAlarm(t *testing.T) {
	isolateAlarmEnv(t)
	m := newAlarm(time.Time{})
	m.mode = alarmAdd
	m.ktvMode = false
	m.inputs = makeNormalInputs()

	m.inputs[0].SetValue("10")
	m.inputs[1].SetValue("30")
	m.inputs[2].SetValue("00")
	m.inputs[3].SetValue("Wake up")

	m.inputs[m.focus].Blur()
	m.focus = 3
	m.inputs[m.focus].Focus()

	if m.focus != 3 {
		t.Fatalf("expected focus=3, got %d", m.focus)
	}
	saved := m.saveAlarm()
	if saved.mode != alarmList {
		t.Errorf("after save, expected alarmList mode, got %d", saved.mode)
	}
	if len(saved.alarms) != 1 {
		t.Errorf("expected 1 alarm saved, got %d", len(saved.alarms))
	}
	if saved.alarms[0].label != "Wake up" {
		t.Errorf("expected label 'Wake up', got %q", saved.alarms[0].label)
	}
}

func TestEnterOnNonLastFieldAdvancesFocus(t *testing.T) {
	isolateAlarmEnv(t)
	m := newAlarm(time.Time{})
	m.mode = alarmAdd
	m.ktvMode = false
	m.inputs = makeNormalInputs()
	m.focus = 0

	next := alarmEnterNextFocus(m.focus, m.ktvMode)
	if next == m.focus {
		t.Errorf("Enter on field 0 should advance focus, not stay at %d", m.focus)
	}
	if next != 1 {
		t.Errorf("Enter on field 0 (normal mode) should go to 1, got %d", next)
	}
}

func TestKTVModeEnterOnLabelSaves(t *testing.T) {
	isolateAlarmEnv(t)
	m := newAlarm(time.Time{})
	m.mode = alarmAdd
	m.ktvMode = true
	m.inputs = makeKTVInputs()
	m.inputs[0].SetValue("5.3.9.2")
	m.inputs[3].SetValue("Lunch")
	m.focus = 3

	saved := m.saveAlarm()
	if saved.mode != alarmList {
		t.Errorf("KTV mode: after save, expected alarmList, got %d", saved.mode)
	}
	if len(saved.alarms) != 1 {
		t.Errorf("expected 1 alarm, got %d", len(saved.alarms))
	}
}
