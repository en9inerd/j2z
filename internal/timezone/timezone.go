package timezone

import (
	"log/slog"
	"time"
)

// GetTimeZone returns the timezone for the given name, falling back to
// the local timezone if the name is empty or invalid.
func GetTimeZone(tzName string) *time.Location {
	if tzName != "" {
		tz, err := time.LoadLocation(tzName)
		if err != nil {
			slog.Warn("invalid timezone, using local", "tz", tzName, "err", err)
			return time.Local
		}
		return tz
	}
	return time.Local
}
