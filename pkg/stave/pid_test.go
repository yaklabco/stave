package stave

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMainFilePathFromExePath(t *testing.T) {
	dir := "/tmp"
	exePath := "/path/to/myexe"
	path := mainFilePathFromExePath(dir, exePath)

	assert.True(t, strings.HasPrefix(filepath.Base(path), mainFileBase+"_"))
	assert.True(t, strings.HasSuffix(path, ".go"))

	// Now it should HAVE the PID
	pid := strconv.Itoa(os.Getpid())
	assert.Contains(t, filepath.Base(path), "_"+pid)
}
