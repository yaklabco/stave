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
			Level:           log.InfoLevel, // This is the default; called can grab the returned logHandler and call the SetLevel method on it to set it to something else.
			ReportTimestamp: true,
			ReportCaller:    true,
		},
	)
	logger := slog.New(logHandler)
	slog.SetDefault(logger)

	return logHandler
}
