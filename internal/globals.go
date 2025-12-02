//nolint:gochecknoglobals // These are all intended as constants (and are private).
package internal

import (
	"io"
	"log"
)

var debug = log.New(io.Discard, "", 0)
