package ktv

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Kaktovik Numerals U+1D2C0–U+1D2D3
var digits = []string{
	"𝋀", "𝋁", "𝋂", "𝋃", "𝋄",
	"𝋅", "𝋆", "𝋇", "𝋈", "𝋉",
	"𝋊", "𝋋", "𝋌", "𝋍", "𝋎",
	"𝋏", "𝋐", "𝋑", "𝋒", "𝋓",
}

var charToInt map[string]int

func init() {
	charToInt = make(map[string]int, 20)
	for i, d := range digits {
		charToInt[d] = i
	}
}

// Time represents a Kaktovik time value.
// The day is divided into 20 ikarraq, each into 20 mein, each into 20 tick, each into 20 kick.
type Time struct {
	Ikarraq int // 0–19  (~72 min each)
	Mein    int // 0–19  (~3.6 min each)
	Tick    int // 0–19  (~10.8 s each)
	Kick    int // 0–19  (~0.54 s each)
}

// FromTime converts a standard time to Kaktovik time.
func FromTime(t time.Time) Time {
	secs := float64(t.Hour()*3600+t.Minute()*60+t.Second()) + float64(t.Nanosecond())/1e9
	return fromFraction(secs / 86400.0)
}

// FromHMS converts hours/minutes/seconds/milliseconds to Kaktovik time.
func FromHMS(h, m, s, ms int) Time {
	total := float64(h*3600+m*60+s) + float64(ms)/1000.0
	return fromFraction(total / 86400.0)
}

// FromDuration treats a duration as elapsed time since midnight and converts to Kaktovik.
func FromDuration(d time.Duration) Time {
	return fromFraction(d.Seconds() / 86400.0)
}

func fromFraction(frac float64) Time {
	frac = frac - float64(int(frac)) // keep in [0,1)

	i := int(frac * 20)
	rem := frac*20 - float64(i)

	m := int(rem * 20)
	rem = rem*20 - float64(m)

	t := int(rem * 20)
	rem = rem*20 - float64(t)

	k := int(rem * 20)

	return Time{clamp(i), clamp(m), clamp(t), clamp(k)}
}

func clamp(n int) int {
	if n < 0 {
		return 0
	}
	if n > 19 {
		return 19
	}
	return n
}

// ToHMS converts a Kaktovik time back to hours, minutes, seconds, milliseconds.
func (kt Time) ToHMS() (h, m, s, ms int) {
	frac := float64(kt.Ikarraq)/20.0 +
		float64(kt.Mein)/400.0 +
		float64(kt.Tick)/8000.0 +
		float64(kt.Kick)/160000.0
	total := frac * 86400.0
	h = int(total / 3600)
	total -= float64(h * 3600)
	m = int(total / 60)
	total -= float64(m * 60)
	s = int(total)
	ms = int((total - float64(s)) * 1000)
	return
}

// ToDuration converts Kaktovik time to a duration (seconds since midnight).
func (kt Time) ToDuration() time.Duration {
	frac := float64(kt.Ikarraq)/20.0 +
		float64(kt.Mein)/400.0 +
		float64(kt.Tick)/8000.0 +
		float64(kt.Kick)/160000.0
	return time.Duration(frac*86400*1e9) * time.Nanosecond
}

// Digit returns the Kaktovik numeral character for n (0–19).
func Digit(n int) string {
	if n < 0 || n > 19 {
		return "?"
	}
	return digits[n]
}

// String returns the four Kaktovik numeral characters for this time.
func (kt Time) String() string {
	return digits[kt.Ikarraq] + digits[kt.Mein] + digits[kt.Tick] + digits[kt.Kick]
}

// Spaced returns the digits separated by spaces for readability.
func (kt Time) Spaced() string {
	return fmt.Sprintf("%s  %s  %s  %s",
		digits[kt.Ikarraq], digits[kt.Mein], digits[kt.Tick], digits[kt.Kick])
}

// Dotted returns the integer values separated by dots, e.g. "5.3.9.2".
func (kt Time) Dotted() string {
	return fmt.Sprintf("%d.%d.%d.%d", kt.Ikarraq, kt.Mein, kt.Tick, kt.Kick)
}

// ParseDotted parses a "I.M.T.K" string (integer values 0–19).
func ParseDotted(s string) (Time, error) {
	parts := strings.Split(strings.TrimSpace(s), ".")
	if len(parts) != 4 {
		return Time{}, fmt.Errorf("expected I.M.T.K format, got %q", s)
	}
	nums := make([]int, 4)
	for i, p := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil || n < 0 || n > 19 {
			return Time{}, fmt.Errorf("each component must be 0–19, got %q", p)
		}
		nums[i] = n
	}
	return Time{nums[0], nums[1], nums[2], nums[3]}, nil
}

// ParseChars parses exactly 4 Kaktovik numeral Unicode characters (e.g. "𝋅𝋃𝋉𝋂").
func ParseChars(s string) (Time, error) {
	runes := []rune(strings.TrimSpace(s))
	if len(runes) != 4 {
		return Time{}, fmt.Errorf("expected 4 Kaktovik numeral characters, got %d", len(runes))
	}
	var nums [4]int
	for i, r := range runes {
		n, ok := charToInt[string(r)]
		if !ok {
			return Time{}, fmt.Errorf("character %q is not a Kaktovik numeral (U+1D2C0–U+1D2D3)", string(r))
		}
		nums[i] = n
	}
	return Time{nums[0], nums[1], nums[2], nums[3]}, nil
}

// ParseAny accepts a KTV numeral string ("𝋅𝋃𝋉𝋂"), a dotted base-20 string ("5.3.9.2"),
// or a single base-20 integer run ("5392" parsed left-to-right as four digits).
func ParseAny(s string) (Time, error) {
	s = strings.TrimSpace(s)
	// 4 KTV Unicode chars
	if t, err := ParseChars(s); err == nil {
		return t, nil
	}
	// dotted base-20
	if strings.Contains(s, ".") {
		if t, err := ParseDotted(s); err == nil {
			return t, nil
		}
	}
	return Time{}, fmt.Errorf("cannot parse %q as Kaktovik time (use 𝋅𝋃𝋉𝋂, 5.3.9.2, etc.)", s)
}
