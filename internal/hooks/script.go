package hooks

import (
	"bufio"
	"bytes"
	"log/slog"
	"os"
	"strings"
	"text/template"
)

// StaveMarker is the marker comment used to identify Stave-managed hooks.
const StaveMarker = "# Installed by Stave: DO NOT EDIT BY HAND"

// ScriptParams configures hook script generation.
type ScriptParams struct {
	// HookName is the name of the Git hook (e.g., "pre-commit", "pre-push").
	HookName string
}

// hookScriptTemplate is the template for generated hook scripts.
// It follows POSIX sh conventions for maximum portability.
const hookScriptTemplate = `#!/bin/sh
# Installed by Stave: DO NOT EDIT BY HAND

# Optional user-level initialization (PATH, version managers, etc.)
init_script="${XDG_CONFIG_HOME:-$HOME/.config}/stave/hooks/init.sh"
[ -f "$init_script" ] && . "$init_script"

# Global toggle and debug controls
if [ "${STAVE_HOOKS-}" = "0" ]; then
  exit 0
fi
[ "${STAVE_HOOKS-}" = "debug" ] && set -x

if command -v stave >/dev/null 2>&1; then
  exec stave --hooks run {{.HookName}} -- "$@"
else
  echo "stave: 'stave' binary not found on PATH; skipping {{.HookName}} hook." >&2
  exit 0
fi
`

//nolint:gochecknoglobals // template is parsed once at init
var scriptTmpl = template.Must(template.New("hook").Parse(hookScriptTemplate))

// GenerateScript returns the POSIX shell script content for a hook.
// Panics if template execution fails (indicates a programming error).
func GenerateScript(params ScriptParams) string {
	slog.Debug("generating hook script",
		slog.String("hook", params.HookName))

	var buf bytes.Buffer
	if err := scriptTmpl.Execute(&buf, params); err != nil {
		// Template is compile-time constant; failure indicates a bug.
		panic("hooks: template execution failed: " + err.Error())
	}
	return buf.String()
}

// IsStaveManaged checks if a hook file was installed by Stave.
// It looks for the StaveMarker in the first few lines of the file.
func IsStaveManaged(path string) (bool, error) {
	slog.Debug("checking stave marker",
		slog.String("path", path))

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Debug("hook file does not exist",
				slog.String("path", path))
			return false, nil
		}
		return false, err
	}
	defer file.Close()

	// Check first 5 lines for the marker
	scanner := bufio.NewScanner(file)
	lineCount := 0
	for scanner.Scan() && lineCount < 5 {
		line := scanner.Text()
		if strings.Contains(line, "Installed by Stave") {
			slog.Debug("stave marker found",
				slog.String("path", path))
			return true, nil
		}
		lineCount++
	}

	if err := scanner.Err(); err != nil {
		return false, err
	}

	slog.Debug("stave marker not found",
		slog.String("path", path))
	return false, nil
}

// execPerm is the permission mode for executable scripts.
const execPerm = 0o755

// WriteHookScript writes a hook script to the specified path.
// The file is created with executable permissions (0755).
func WriteHookScript(path string, params ScriptParams) error {
	slog.Debug("writing hook script",
		slog.String("path", path),
		slog.String("hook", params.HookName))

	content := GenerateScript(params)
	// Hook scripts need to be executable, hence 0755.
	// #nosec G306 -- This is intentional: hooks must be executable
	return os.WriteFile(path, []byte(content), execPerm)
}

// RemoveHookScript removes a hook script if it was installed by Stave.
// Returns true if the file was removed, false if it wasn't Stave-managed or didn't exist.
func RemoveHookScript(path string) (bool, error) {
	managed, err := IsStaveManaged(path)
	if err != nil {
		return false, err
	}
	if !managed {
		slog.Debug("skipping removal of non-stave hook",
			slog.String("path", path))
		return false, nil
	}

	slog.Debug("removing hook script",
		slog.String("path", path))

	if err := os.Remove(path); err != nil {
		return false, err
	}
	return true, nil
}
