package timeutil

import (
	"fmt"
	"strings"
	"time"
)

// ReferenceTime Return the standard Golang reference time (2006-01-02T15:04:05.999999999Z07:00)
func ReferenceTime() time.Time {
	return ReferenceTimeValue
}

// FormatTimer Formats the given duration in a colon-separated timer format in the form
// [HH:]MM:SS.
func FormatTimer(d time.Duration) string {
	h, m, s := DurationHMS(d)

	out := fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	out = strings.TrimPrefix(out, `00:`)
	out = strings.TrimPrefix(out, `0`)
	return out
}

// FormatTimerf Formats the given duration using the given format string.  The string follows
// the same formatting rules as described in the fmt package, and will receive
// three integer arguments: hours, minutes, and seconds.
func FormatTimerf(format string, d time.Duration) string {
	h, m, s := DurationHMS(d)

	out := fmt.Sprintf(format, h, m, s)
	return out
}

// DurationHMS Extracts the hours, minutes, and seconds from the given duration.
func DurationHMS(d time.Duration) (int, int, int) {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	return int(h), int(m), int(s)
}
