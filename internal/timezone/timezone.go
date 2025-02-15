package timezone

import (
	"time"

	"github.com/en9inerd/j2z/internal/log"
)

// Get the timezone, defaulting to local timezone if not provided
func GetTimeZone(tzName string) *time.Location {
	if tzName != "" {
		tz, err := time.LoadLocation(tzName)
		if err != nil {
			log.Logger.Println("Invalid timezone specified, using local timezone instead.")
			return time.Local
		}
		return tz
	}
	return time.Local
}
