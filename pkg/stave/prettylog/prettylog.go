package prettylog

import (
	"io"
	"log/slog"

	"github.com/charmbracelet/log"
)

func SetupPrettyLogger(writerForLogger io.Writer) *log.Logger {
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
