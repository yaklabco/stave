package stave

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/stave/config"
	"github.com/yaklabco/stave/pkg/st"
)

// hermeticCacheEnv pins every input of cache-directory resolution to
// temporary directories so tests never read or delete the developer's real
// caches. It returns the configured (XDG-derived) cache dir followed by the
// legacy st.CacheDir location, both as resolved under the pinned environment.
func hermeticCacheEnv(t *testing.T) (string, string) {
	t.Helper()

	home := t.TempDir()
	t.Setenv("HOME", home)
	if runtime.GOOS == "windows" {
		// st.CacheDir derives the legacy location from HOMEDRIVE/HOMEPATH on
		// Windows; everything else in the resolution chain reads HOME or the
		// XDG variables pinned below.
		volume := filepath.VolumeName(home)
		t.Setenv("HOMEDRIVE", volume)
		t.Setenv("HOMEPATH", home[len(volume):])
	}
	t.Setenv("XDG_CACHE_HOME", filepath.Join(home, "xdg-cache"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, "xdg-config"))
	t.Setenv(st.CacheEnv, "")

	return config.ResolveXDGPaths().CacheDir(), st.CacheDir()
}

// writeCacheFile creates dir if needed and drops a file in it, standing in
// for a cached compiled binary.
func writeCacheFile(t *testing.T, dir, name string) string {
	t.Helper()

	require.NoError(t, os.MkdirAll(dir, 0o700))
	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, []byte("stale binary"), 0o700))

	return path
}

// writeProjectConfig writes a stave.yaml in projectDir pointing cache_dir at
// cacheDir. The path is single-quoted so Windows backslashes survive YAML.
func writeProjectConfig(t *testing.T, projectDir, cacheDir string) {
	t.Helper()

	require.NoError(t, os.WriteFile(
		filepath.Join(projectDir, "stave.yaml"),
		fmt.Appendf(nil, "cache_dir: '%s'\n", cacheDir),
		0o600,
	))
}

func cleanRunParams(t *testing.T) RunParams {
	t.Helper()

	return RunParams{
		BaseCtx: t.Context(),
		Dir:     t.TempDir(),
		Stdout:  io.Discard,
		Stderr:  io.Discard,
		Clean:   true,
	}
}

func TestCleanRemovesConfiguredCacheDirContents(t *testing.T) {
	configured, _ := hermeticCacheEnv(t)

	stale := writeCacheFile(t, configured, "0123abcd")
	kept := writeCacheFile(t, filepath.Join(configured, "subdir"), "kept")

	require.NoError(t, Run(cleanRunParams(t)))

	assert.NoFileExists(t, stale, "clean must remove binaries from the configured cache dir")
	assert.FileExists(t, kept, "clean must not descend into subdirectories")
}

func TestCleanSweepsLegacyCacheDir(t *testing.T) {
	configured, legacy := hermeticCacheEnv(t)

	staleConfigured := writeCacheFile(t, configured, "configured-binary")
	staleLegacy := writeCacheFile(t, legacy, "legacy-binary")

	require.NoError(t, Run(cleanRunParams(t)))

	assert.NoFileExists(t, staleConfigured)
	assert.NoFileExists(t, staleLegacy,
		"clean must also sweep the legacy cache dir left behind by older versions")
}

func TestCleanIgnoresLegacyPathThatIsNotADirectory(t *testing.T) {
	configured, legacy := hermeticCacheEnv(t)

	require.NoError(t, os.WriteFile(legacy, []byte("not a cache dir"), 0o600))
	stale := writeCacheFile(t, configured, "configured-binary")

	require.NoError(t, Run(cleanRunParams(t)))

	assert.NoFileExists(t, stale)
	assert.FileExists(t, legacy, "a stray file at the legacy path must be left alone")
}

func TestCleanHonorsProjectConfigCacheDir(t *testing.T) {
	hermeticCacheEnv(t)

	projectDir := t.TempDir()
	projectCache := filepath.Join(projectDir, "project-cache")
	writeProjectConfig(t, projectDir, projectCache)
	stale := writeCacheFile(t, projectCache, "project-binary")

	params := cleanRunParams(t)
	params.Dir = projectDir
	require.NoError(t, Run(params))

	assert.NoFileExists(t, stale, "clean must honor cache_dir from project config")
}

func TestCleanHonorsProjectConfigWithStavefilesLayout(t *testing.T) {
	hermeticCacheEnv(t)

	// stave.yaml lives in the project root; preprocessRunParams redirects
	// Dir into stavefiles/, and config resolution must still find it.
	projectDir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(projectDir, StavefilesDirName), 0o700))
	projectCache := filepath.Join(projectDir, "project-cache")
	writeProjectConfig(t, projectDir, projectCache)
	stale := writeCacheFile(t, projectCache, "project-binary")

	params := cleanRunParams(t)
	params.Dir = projectDir
	require.NoError(t, Run(params))

	assert.NoFileExists(t, stale,
		"clean must honor project cache_dir when stavefiles live in a stavefiles/ directory")
}

func TestResolveCacheDir(t *testing.T) {
	t.Run("empty falls back to configured default", func(t *testing.T) {
		configured, _ := hermeticCacheEnv(t)

		got, err := resolveCacheDir(RunParams{Dir: t.TempDir()})
		require.NoError(t, err)

		assert.Equal(t, configured, got)
	})

	t.Run("explicit value wins over env and config", func(t *testing.T) {
		hermeticCacheEnv(t)
		t.Setenv(st.CacheEnv, t.TempDir())

		explicit := t.TempDir()
		got, err := resolveCacheDir(RunParams{Dir: t.TempDir(), CacheDir: explicit})
		require.NoError(t, err)

		assert.Equal(t, explicit, got)
	})

	t.Run("cache env variable wins over default", func(t *testing.T) {
		hermeticCacheEnv(t)

		envDir := t.TempDir()
		t.Setenv(st.CacheEnv, envDir)
		got, err := resolveCacheDir(RunParams{Dir: t.TempDir()})
		require.NoError(t, err)

		assert.Equal(t, envDir, got)
	})

	t.Run("cache env variable wins over project config", func(t *testing.T) {
		hermeticCacheEnv(t)

		projectDir := t.TempDir()
		writeProjectConfig(t, projectDir, filepath.Join(projectDir, "project-cache"))
		envDir := t.TempDir()
		t.Setenv(st.CacheEnv, envDir)

		got, err := resolveCacheDir(RunParams{Dir: projectDir})
		require.NoError(t, err)

		assert.Equal(t, envDir, got)
	})

	t.Run("stavefiles dir with trailing separator resolves project config", func(t *testing.T) {
		hermeticCacheEnv(t)

		projectDir := t.TempDir()
		require.NoError(t, os.Mkdir(filepath.Join(projectDir, StavefilesDirName), 0o700))
		projectCache := filepath.Join(projectDir, "project-cache")
		writeProjectConfig(t, projectDir, projectCache)

		dirWithSeparator := filepath.Join(projectDir, StavefilesDirName) + string(os.PathSeparator)
		got, err := resolveCacheDir(RunParams{Dir: dirWithSeparator})
		require.NoError(t, err)

		assert.Equal(t, projectCache, got,
			"-C proj/stavefiles/ must read the same project config as -C proj/stavefiles")
	})

	t.Run("unparseable project config is an error", func(t *testing.T) {
		hermeticCacheEnv(t)

		projectDir := t.TempDir()
		require.NoError(t, os.WriteFile(
			filepath.Join(projectDir, "stave.yaml"),
			[]byte("cache_dir: [not, a, string"),
			0o600,
		))

		_, err := resolveCacheDir(RunParams{Dir: projectDir})
		assert.Error(t, err)
	})

	t.Run("cosmetic validation error does not block resolution", func(t *testing.T) {
		hermeticCacheEnv(t)

		projectDir := t.TempDir()
		projectCache := filepath.Join(projectDir, "project-cache")
		require.NoError(t, os.WriteFile(
			filepath.Join(projectDir, "stave.yaml"),
			fmt.Appendf(nil, "cache_dir: '%s'\ntarget_color: purple\n", projectCache),
			0o600,
		))

		got, err := resolveCacheDir(RunParams{Dir: projectDir})
		require.NoError(t, err,
			"an invalid target_color must not prevent builds or cleaning")

		assert.Equal(t, projectCache, got)
	})
}

// TestRunThenCleanUseTheSameCacheDir pins the original bug end to end: the
// directory a target run compiles into must be the directory --clean empties.
func TestRunThenCleanUseTheSameCacheDir(t *testing.T) {
	// Capture the real Go caches before pinning HOME, then restore them so
	// the compile stays warm inside an otherwise hermetic environment.
	goEnv, err := exec.Command("go", "env", "GOCACHE", "GOMODCACHE").Output()
	require.NoError(t, err)
	goCaches := strings.Fields(string(goEnv))
	require.Len(t, goCaches, 2)

	configured, _ := hermeticCacheEnv(t)
	t.Setenv("GOCACHE", goCaches[0])
	t.Setenv("GOMODCACHE", goCaches[1])

	stderr := &bytes.Buffer{}
	params := RunParams{
		BaseCtx: t.Context(),
		Dir:     testDataDir,
		Stdout:  io.Discard,
		Stderr:  stderr,
		Args:    []string{"testverbose"},
	}
	require.NoError(t, Run(params), "stderr was: %s", stderr.String())

	entries, err := os.ReadDir(configured)
	require.NoError(t, err, "run must compile the stavefile binary into the configured cache dir")
	require.NotEmpty(t, entries)

	cleanParams := cleanRunParams(t)
	require.NoError(t, Run(cleanParams))

	entries, err = os.ReadDir(configured)
	require.NoError(t, err)
	for _, entry := range entries {
		assert.True(t, entry.IsDir(),
			"clean must remove the binaries that a target run compiled, found %q", entry.Name())
	}
}
