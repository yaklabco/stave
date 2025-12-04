package st

import (
	"os"

	cblog "github.com/charmbracelet/log"
	"github.com/yaklabco/stave/pkg/stave/prettylog"
)

func init() {
	cbLogger := prettylog.SetupPrettyLogger(os.Stdout)
	if Debug() {
		cbLogger.SetLevel(cblog.DebugLevel)
	}

	cbLogger.SetLevel(cblog.InfoLevel)
}
