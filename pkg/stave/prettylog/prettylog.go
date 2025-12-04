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
			Level:           log.InfoLevel, // Setting this to lowest possible value, since slog will handle the actual filtering.
			ReportTimestamp: true,
			ReportCaller:    true,
		},
	)
	logger := slog.New(logHandler)
	slog.SetDefault(logger)

	return logHandler
}
