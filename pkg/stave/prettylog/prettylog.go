package prettylog

import (
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/charmbracelet/log"
)

func SetupPrettyLogger(writerForLogger io.Writer) *log.Logger {
	setupTERMRBG()

	logHandler := log.NewWithOptions(
		writerForLogger,
		log.Options{
			// Default level. Callers can use SetLevel on the returned handler to change.
			Level:           log.InfoLevel,
			ReportTimestamp: true,
			ReportCaller:    true,
		},
	)
	logger := slog.New(logHandler)
	slog.SetDefault(logger)

	return logHandler
}

func setupTERMRBG() {
	origTERM := os.Getenv("TERM")
	switch {
	case strings.HasPrefix(origTERM, "screen"):
		fallthrough
	case strings.HasPrefix(origTERM, "tmux"):
		fallthrough
	case strings.HasPrefix(origTERM, "dumb"):
	// no-op
	default:
		_ = os.Setenv("TERM", "screen-256color")
	}

	if os.Getenv("COLORFGBG") == "" {
		_ = os.Setenv("COLORFGBG", "15;0")
	}
}
