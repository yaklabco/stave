package update

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCacheRoundTrip(t *testing.T) {
	cacheDir := filepath.Join(t.TempDir(), "stave")

	entry := &CacheEntry{
		CheckedAt:     time.Now().Truncate(time.Second),
		LatestVersion: "v0.11.0",
		ReleaseURL:    "https://github.com/yaklabco/stave/releases/tag/v0.11.0",
		ReleaseBody:   "## What's Changed\n- New feature",
	}

	err := WriteCache(cacheDir, entry)
	require.NoError(t, err)

	got := ReadCache(cacheDir)
	require.NotNil(t, got)

	assert.Equal(t, entry.CheckedAt.UTC(), got.CheckedAt.UTC())
	assert.Equal(t, entry.LatestVersion, got.LatestVersion)
	assert.Equal(t, entry.ReleaseURL, got.ReleaseURL)
	assert.Equal(t, entry.ReleaseBody, got.ReleaseBody)
}

func TestCacheExpired(t *testing.T) {
	cacheDir := filepath.Join(t.TempDir(), "stave")

	entry := &CacheEntry{
		CheckedAt:     time.Now().Add(-25 * time.Hour),
		LatestVersion: "v0.11.0",
		ReleaseURL:    "https://github.com/yaklabco/stave/releases/tag/v0.11.0",
	}

	err := WriteCache(cacheDir, entry)
	require.NoError(t, err)

	got := ReadCache(cacheDir)
	require.NotNil(t, got)

	assert.True(t, got.IsExpired(24*time.Hour))
}

func TestCacheNotExpired(t *testing.T) {
	cacheDir := filepath.Join(t.TempDir(), "stave")

	entry := &CacheEntry{
		CheckedAt:     time.Now(),
		LatestVersion: "v0.11.0",
		ReleaseURL:    "https://github.com/yaklabco/stave/releases/tag/v0.11.0",
	}

	err := WriteCache(cacheDir, entry)
	require.NoError(t, err)

	got := ReadCache(cacheDir)
	require.NotNil(t, got)

	assert.False(t, got.IsExpired(24*time.Hour))
}

func TestReadCache_FileNotFound(t *testing.T) {
	cacheDir := filepath.Join(t.TempDir(), "nonexistent")

	got := ReadCache(cacheDir)
	assert.Nil(t, got)
}

func TestReadCache_CorruptFile(t *testing.T) {
	cacheDir := filepath.Join(t.TempDir(), "stave")

	require.NoError(t, os.MkdirAll(cacheDir, 0o755))

	corruptPath := filepath.Join(cacheDir, cacheFileName)
	require.NoError(t, os.WriteFile(corruptPath, []byte(`{not valid json`), 0o600))

	got := ReadCache(cacheDir)
	assert.Nil(t, got)
}

func TestWriteCache_CreatesDir(t *testing.T) {
	cacheDir := filepath.Join(t.TempDir(), "deep", "nested", "dir")

	entry := &CacheEntry{
		CheckedAt:     time.Now(),
		LatestVersion: "v0.11.0",
		ReleaseURL:    "https://github.com/yaklabco/stave/releases/tag/v0.11.0",
	}

	err := WriteCache(cacheDir, entry)
	require.NoError(t, err)

	// Verify the directory was created.
	info, err := os.Stat(cacheDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	// Verify the file was written.
	filePath := filepath.Join(cacheDir, cacheFileName)
	_, err = os.Stat(filePath)
	require.NoError(t, err)
}

func TestCachePath(t *testing.T) {
	got := cachePath("/foo/bar")
	assert.Equal(t, "/foo/bar/update-check.json", got)
}

func TestCacheNotifiedVersion(t *testing.T) {
	t.Run("with notified version", func(t *testing.T) {
		cacheDir := filepath.Join(t.TempDir(), "stave")

		entry := &CacheEntry{
			CheckedAt:       time.Now().Truncate(time.Second),
			LatestVersion:   "v0.11.0",
			ReleaseURL:      "https://github.com/yaklabco/stave/releases/tag/v0.11.0",
			NotifiedVersion: "v0.11.0",
		}

		err := WriteCache(cacheDir, entry)
		require.NoError(t, err)

		got := ReadCache(cacheDir)
		require.NotNil(t, got)
		assert.Equal(t, "v0.11.0", got.NotifiedVersion)
	})

	t.Run("without notified version", func(t *testing.T) {
		cacheDir := filepath.Join(t.TempDir(), "stave")

		entry := &CacheEntry{
			CheckedAt:     time.Now().Truncate(time.Second),
			LatestVersion: "v0.11.0",
			ReleaseURL:    "https://github.com/yaklabco/stave/releases/tag/v0.11.0",
		}

		err := WriteCache(cacheDir, entry)
		require.NoError(t, err)

		got := ReadCache(cacheDir)
		require.NotNil(t, got)
		assert.Empty(t, got.NotifiedVersion)
	})
}
