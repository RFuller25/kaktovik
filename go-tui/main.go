package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/rfuller25/kaktovik/go-tui/assets"
	"github.com/rfuller25/kaktovik/go-tui/internal/config"
	"github.com/rfuller25/kaktovik/go-tui/internal/ktv"
	"github.com/rfuller25/kaktovik/go-tui/internal/notify"
	"github.com/rfuller25/kaktovik/go-tui/internal/ui"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "kaktovik",
	Short: "Kaktovik time TUI — display, convert, timer, stopwatch, and alarm",
	Long: `A terminal app for Kaktovik (Inupiaq base-20) time.

The day is divided into 20 ikarraq (~72 min each), each into 20 mein (~3.6 min),
each into 20 tick (~10.8 s), each into 20 kick (~0.54 s).

Run without subcommands to open the full TUI.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		tzFlag, _ := cmd.Flags().GetString("timezone")
		loc, err := parseTimezone(tzFlag)
		if err != nil {
			return err
		}
		cfg, _ := config.Load()
		if tzFlag != "" {
			cfg.Timezone = tzFlag
		}
		ui.ApplyTheme(cfg.Theme)
		return runTUI(ui.Options{Timezone: loc, Cfg: cfg})
	},
}

var timerCmd = &cobra.Command{
	Use:   "timer [duration]",
	Short: "Start a countdown timer",
	Long: `Start a countdown timer. Duration can be in Go format (5m30s, 1h) or
Kaktovik I.M.T.K format (e.g. 0.2.3.0 = 2 mein 3 tick).

With --headless, runs in background and sends a desktop notification on completion.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		headless, _ := cmd.Flags().GetBool("headless")

		var preset time.Duration
		if len(args) > 0 {
			d, err := parseDurationArg(args[0])
			if err != nil {
				return err
			}
			preset = d
		}

		cfg, _ := config.Load()
		if headless && preset > 0 {
			return runHeadlessTimer(preset, cfg)
		}
		ui.ApplyTheme(cfg.Theme)
		return runTUI(ui.Options{InitialTab: ui.TabTimer, TimerPreset: preset, Cfg: cfg})
	},
}

var alarmCmd = &cobra.Command{
	Use:   "alarm [HH:MM[:SS]]",
	Short: "Set a time alarm",
	Long: `Set an alarm for a specific time of day.
Time can be in 24-hour format (HH:MM or HH:MM:SS).

With --headless, runs in background and sends a desktop notification at alarm time.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		headless, _ := cmd.Flags().GetBool("headless")

		var preset time.Time
		if len(args) > 0 {
			t, err := parseTimeArg(args[0])
			if err != nil {
				return err
			}
			preset = t
		}

		immediate, _ := cmd.Flags().GetBool("immediate")
		cfg, _ := config.Load()
		if headless && !preset.IsZero() {
			return runHeadlessAlarm(preset, cfg, immediate)
		}
		ui.ApplyTheme(cfg.Theme)
		return runTUI(ui.Options{InitialTab: ui.TabAlarm, AlarmPreset: preset, Cfg: cfg})
	},
}

var stopwatchCmd = &cobra.Command{
	Use:   "stopwatch",
	Short: "Open the stopwatch",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := config.Load()
		ui.ApplyTheme(cfg.Theme)
		return runTUI(ui.Options{InitialTab: ui.TabStopwatch, Cfg: cfg})
	},
}

var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Open the time converter",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := config.Load()
		ui.ApplyTheme(cfg.Theme)
		return runTUI(ui.Options{InitialTab: ui.TabConvert, Cfg: cfg})
	},
}

var installFontCmd = &cobra.Command{
	Use:   "install-font",
	Short: "Install the embedded Kaktovik Numerals font for the current user",
	Long: `Write the bundled Kaktovik Numerals TTF font to your user font directory
and refresh the font cache so terminals and applications can use it.

Linux:   ~/.local/share/fonts/KaktovikNumerals.ttf  (then fc-cache -f)
macOS:   ~/Library/Fonts/KaktovikNumerals.ttf
Windows: %LOCALAPPDATA%\Microsoft\Windows\Fonts\KaktovikNumerals.ttf`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return installFont()
	},
}

var nowCmd = &cobra.Command{
	Use:   "now",
	Short: "Print current Kaktovik time to stdout (no TUI)",
	RunE: func(cmd *cobra.Command, args []string) error {
		tzFlag, _ := cmd.Flags().GetString("timezone")
		loc, err := parseTimezone(tzFlag)
		if err != nil {
			return err
		}
		now := time.Now().In(loc)
		kt := ktv.FromTime(now)
		h, m, s, _ := kt.ToHMS()
		fmt.Printf("Kaktovik: %s  (%s)\nNormal:   %02d:%02d:%02d\n",
			kt.Spaced(), kt.Dotted(), h, m, s)
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringP("timezone", "z", "", "timezone name (e.g. America/New_York) or offset (e.g. -5)")

	timerCmd.Flags().BoolP("headless", "H", false, "run without TUI, notify on completion")
	alarmCmd.Flags().BoolP("headless", "H", false, "run without TUI, notify at alarm time")
	alarmCmd.Flags().Bool("immediate", false, "fire notification immediately without sleeping (used by scheduled background alarms)")

	rootCmd.AddCommand(timerCmd, alarmCmd, stopwatchCmd, convertCmd, nowCmd, installFontCmd)
}

func runTUI(opts ui.Options) error {
	p := tea.NewProgram(
		ui.New(opts),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	_, err := p.Run()
	return err
}

func runHeadlessTimer(d time.Duration, cfg config.Config) error {
	kt := ktv.FromDuration(d)
	fmt.Printf("Timer started: %s (%s)\n", formatDurationHuman(d), kt.Dotted())
	time.Sleep(d)
	notify.SendUrgent("Kaktovik Timer", fmt.Sprintf("Timer finished after %s", formatDurationHuman(d)),
		cfg.NotifyUrgency, cfg.NotifyIcon)
	notify.TerminalAttention()
	notify.PlaySound(cfg.SoundEnabled, cfg.SoundFile)
	fmt.Println("Timer complete.")
	return nil
}

func runHeadlessAlarm(target time.Time, cfg config.Config, immediate bool) error {
	if !immediate {
		now := time.Now()
		if target.Before(now) {
			target = target.Add(24 * time.Hour)
		}
		wait := time.Until(target)
		kt := ktv.FromTime(target)
		fmt.Printf("Alarm set for %02d:%02d:%02d (%s), fires in %s\n",
			target.Hour(), target.Minute(), target.Second(),
			kt.Dotted(), formatDurationHuman(wait))
		time.Sleep(wait)
	}
	kt := ktv.FromTime(target)
	notify.SendUrgent("Kaktovik Alarm", fmt.Sprintf("Alarm: %02d:%02d:%02d  %s",
		target.Hour(), target.Minute(), target.Second(), kt.Spaced()),
		cfg.NotifyUrgency, cfg.NotifyIcon)
	notify.TerminalAttention()
	notify.PlaySound(cfg.SoundEnabled, cfg.SoundFile)
	fmt.Println("Alarm fired.")
	return nil
}

func parseDurationArg(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	// KTV chars (𝋅𝋃𝋉𝋂) or dotted base-20 (5.3.9.2)
	if kt, err := ktv.ParseAny(s); err == nil {
		return kt.ToDuration(), nil
	}
	return time.ParseDuration(s)
}

func parseTimeArg(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	now := time.Now()

	for _, layout := range []string{"15:04:05", "15:04"} {
		t, err := time.ParseInLocation(layout, s, now.Location())
		if err == nil {
			return time.Date(now.Year(), now.Month(), now.Day(),
				t.Hour(), t.Minute(), t.Second(), 0, now.Location()), nil
		}
	}

	// KTV chars (𝋅𝋃𝋉𝋂) or dotted base-20 (5.3.9.2)
	if kt, err := ktv.ParseAny(s); err == nil {
		h, m, sec, _ := kt.ToHMS()
		return time.Date(now.Year(), now.Month(), now.Day(), h, m, sec, 0, now.Location()), nil
	}

	return time.Time{}, fmt.Errorf("cannot parse time %q (use HH:MM, HH:MM:SS, or KTV I.M.T.K)", s)
}

func parseTimezone(s string) (*time.Location, error) {
	if s == "" {
		return time.Local, nil
	}
	// Try numeric offset like "-5" or "+5.5"
	if n, err := strconv.ParseFloat(s, 64); err == nil {
		secs := int(n * 3600)
		return time.FixedZone(fmt.Sprintf("UTC%+g", n), secs), nil
	}
	return time.LoadLocation(s)
}

func installFont() error {
	dest, err := fontInstallPath()
	if err != nil {
		return err
	}
	if err := assets.AddFontFile(dest); err != nil {
		return fmt.Errorf("installing font: %w", err)
	}
	fmt.Printf("Font installed to: %s\n", dest)
	refreshFontCache(dest)
	return nil
}

func formatDurationHuman(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh%02dm%02ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm%02ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}
