package bandplan

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Frequency is a frequency or frequency span in hertz.
type Frequency int64

const (
	Hz  Frequency = 1
	KHz           = 1000 * Hz
	MHz           = 1000 * KHz
	GHz           = 1000 * MHz
)

// String formats f with the largest unit that keeps the value at 1 or above,
// such as "433.92MHz".
func (f Frequency) String() string {
	unit, name := Hz, "Hz"
	switch a := max(f, -f); {
	case a >= GHz:
		unit, name = GHz, "GHz"
	case a >= MHz:
		unit, name = MHz, "MHz"
	case a >= KHz:
		unit, name = KHz, "kHz"
	}
	return strconv.FormatFloat(float64(f)/float64(unit), 'f', -1, 64) + name
}

// ParseFrequency parses a frequency such as "2.4GHz", "433.92 MHz", "500k"
// or "1000". Units are case insensitive and a bare number is in hertz.
func ParseFrequency(s string) (Frequency, error) {
	num := strings.TrimSpace(s)
	lower := strings.ToLower(num)
	unit := Hz
	for _, u := range []struct {
		suffix string
		unit   Frequency
	}{
		{"ghz", GHz}, {"mhz", MHz}, {"khz", KHz}, {"hz", Hz}, {"g", GHz}, {"m", MHz}, {"k", KHz},
	} {
		if strings.HasSuffix(lower, u.suffix) {
			unit = u.unit
			num = strings.TrimSpace(num[:len(num)-len(u.suffix)])
			break
		}
	}
	v, err := strconv.ParseFloat(num, 64)
	if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
		return 0, fmt.Errorf("bandplan: invalid frequency %q", s)
	}
	return Frequency(math.Round(v * float64(unit))), nil
}

// MarshalText implements encoding.TextMarshaler.
func (f Frequency) MarshalText() ([]byte, error) {
	return []byte(f.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (f *Frequency) UnmarshalText(b []byte) error {
	v, err := ParseFrequency(string(b))
	if err != nil {
		return err
	}
	*f = v
	return nil
}
