package log

import (
	"log/slog"
	"os"
)

// Level is a package-level LevelVar so callers can adjust the log level
// at runtime (e.g. via --verbose / --quiet flags).
var Level = new(slog.LevelVar)

func init() {
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: Level,
	})
	slog.SetDefault(slog.New(handler))
}
