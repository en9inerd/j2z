package main

import "time"

// Get the timezone, defaulting to local timezone if not provided
func getTimeZone(tzName string) *time.Location {
	if tzName != "" {
		tz, err := time.LoadLocation(tzName)
		if err != nil {
			logger.Println("Invalid timezone specified, using local timezone instead.")
			return time.Local
		}
		return tz
	}
	return time.Local
}
