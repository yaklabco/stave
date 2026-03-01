package update

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"regexp"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/stave/config"
)

// ansiPattern matches ANSI escape sequences.
var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// stripANSI removes all ANSI escape codes from a string.
func stripANSI(s string) string {
	return ansiPattern.ReplaceAllString(s, "")
}

// newTestServer returns an httptest.Server that always responds with a JSON
// release containing the given tagName and body.
func newTestServer(t *testing.T, tagName, body string) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		release := Release{
			TagName: tagName,
			HTMLURL: "https://github.com/yaklabco/stave/releases/tag/" + tagName,
			Body:    body,
		}
		writer.Header().Set("Content-Type", "application/json")
		assert.NoError(t, json.NewEncoder(writer).Encode(release))
	}))
}

// newFailingServer returns an httptest.Server that fails the test if any
// request is received. This is used to verify that certain code paths do
// not make HTTP calls.
func newFailingServer(t *testing.T) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fatal("unexpected HTTP request received")
	}))
}

// newCountingServer returns an httptest.Server that increments a counter on
// each request and responds with a release.
func newCountingServer(t *testing.T, tagName string, counter *atomic.Int32) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		counter.Add(1)

		release := Release{
			TagName: tagName,
			HTMLURL: "https://github.com/yaklabco/stave/releases/tag/" + tagName,
			Body:    "changelog",
		}
		writer.Header().Set("Content-Type", "application/json")
		assert.NoError(t, json.NewEncoder(writer).Encode(release))
	}))
}

// enabledConfig returns an UpdateCheckConfig with checking enabled and a long TTL.
func enabledConfig() config.UpdateCheckConfig {
	return config.UpdateCheckConfig{
		Enabled:  true,
		Interval: 24 * time.Hour,
	}
}

// --- CheckAndNotify tests ---

func TestCheckAndNotify_SkipsInCI(t *testing.T) {
	t.Setenv("CI", "true")

	server := newFailingServer(t)
	defer server.Close()

	var buf bytes.Buffer

	CheckAndNotify(context.Background(), Params{
		CurrentVersion: "v0.12.0",
		CacheDir:       t.TempDir(),
		Output:         &buf,
		Config:         enabledConfig(),
		ClientOptions:  []GitHubClientOption{WithBaseURL(server.URL)},
	})

	assert.Empty(t, buf.String())
}

func TestCheckAndNotify_SkipsWhenDisabled(t *testing.T) {
	server := newFailingServer(t)
	defer server.Close()

	var buf bytes.Buffer

	CheckAndNotify(context.Background(), Params{
		CurrentVersion: "v0.12.0",
		CacheDir:       t.TempDir(),
		Output:         &buf,
		Config: config.UpdateCheckConfig{
			Enabled:  false,
			Interval: 24 * time.Hour,
		},
		ClientOptions: []GitHubClientOption{WithBaseURL(server.URL)},
	})

	assert.Empty(t, buf.String())
}

func TestCheckAndNotify_SkipsDevVersion(t *testing.T) {
	server := newFailingServer(t)
	defer server.Close()

	var buf bytes.Buffer

	CheckAndNotify(context.Background(), Params{
		CurrentVersion: "dev",
		CacheDir:       t.TempDir(),
		Output:         &buf,
		Config:         enabledConfig(),
		ClientOptions:  []GitHubClientOption{WithBaseURL(server.URL)},
	})

	assert.Empty(t, buf.String())
}

func TestCheckAndNotify_NewerVersionAvailable(t *testing.T) {
	server := newTestServer(t, "v0.13.0", "some changelog")
	defer server.Close()

	cacheDir := filepath.Join(t.TempDir(), "stave")

	var buf bytes.Buffer

	CheckAndNotify(context.Background(), Params{
		CurrentVersion: "v0.12.0",
		CacheDir:       cacheDir,
		Output:         &buf,
		Config:         enabledConfig(),
		ClientOptions:  []GitHubClientOption{WithBaseURL(server.URL)},
	})

	assert.Contains(t, buf.String(), "v0.13.0")
	assert.Contains(t, buf.String(), "v0.12.0")

	// Verify cache has NotifiedVersion set.
	cached := ReadCache(cacheDir)
	require.NotNil(t, cached)
	assert.Equal(t, "v0.13.0", cached.NotifiedVersion)
}

func TestCheckAndNotify_SilentAfterNotified(t *testing.T) {
	server := newFailingServer(t)
	defer server.Close()

	cacheDir := filepath.Join(t.TempDir(), "stave")

	// Pre-write cache with NotifiedVersion matching LatestVersion.
	entry := &CacheEntry{
		CheckedAt:       time.Now(),
		LatestVersion:   "v0.13.0",
		ReleaseURL:      "https://github.com/yaklabco/stave/releases/tag/v0.13.0",
		NotifiedVersion: "v0.13.0",
	}
	require.NoError(t, WriteCache(cacheDir, entry))

	var buf bytes.Buffer

	CheckAndNotify(context.Background(), Params{
		CurrentVersion: "v0.12.0",
		CacheDir:       cacheDir,
		Output:         &buf,
		Config:         enabledConfig(),
		ClientOptions:  []GitHubClientOption{WithBaseURL(server.URL)},
	})

	assert.Empty(t, buf.String())
}

func TestCheckAndNotify_SilentWhenUpToDate(t *testing.T) {
	server := newTestServer(t, "v0.12.0", "changelog")
	defer server.Close()

	var buf bytes.Buffer

	CheckAndNotify(context.Background(), Params{
		CurrentVersion: "v0.12.0",
		CacheDir:       filepath.Join(t.TempDir(), "stave"),
		Output:         &buf,
		Config:         enabledConfig(),
		ClientOptions:  []GitHubClientOption{WithBaseURL(server.URL)},
	})

	assert.Empty(t, buf.String())
}

func TestCheckAndNotify_SilentOnNetworkError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	var buf bytes.Buffer

	CheckAndNotify(context.Background(), Params{
		CurrentVersion: "v0.12.0",
		CacheDir:       filepath.Join(t.TempDir(), "stave"),
		Output:         &buf,
		Config:         enabledConfig(),
		ClientOptions:  []GitHubClientOption{WithBaseURL(server.URL)},
	})

	assert.Empty(t, buf.String())
}

func TestCheckAndNotify_UsesCachedResult(t *testing.T) {
	server := newFailingServer(t)
	defer server.Close()

	cacheDir := filepath.Join(t.TempDir(), "stave")

	// Pre-write a fresh cache that is NOT newer (same version).
	entry := &CacheEntry{
		CheckedAt:     time.Now(),
		LatestVersion: "v0.12.0",
		ReleaseURL:    "https://github.com/yaklabco/stave/releases/tag/v0.12.0",
	}
	require.NoError(t, WriteCache(cacheDir, entry))

	var buf bytes.Buffer

	CheckAndNotify(context.Background(), Params{
		CurrentVersion: "v0.12.0",
		CacheDir:       cacheDir,
		Output:         &buf,
		Config:         enabledConfig(),
		ClientOptions:  []GitHubClientOption{WithBaseURL(server.URL)},
	})

	// No HTTP call was made (server would fail test), and no output because up to date.
	assert.Empty(t, buf.String())
}

func TestCheckAndNotify_WritesCache(t *testing.T) {
	server := newTestServer(t, "v0.13.0", "changelog")
	defer server.Close()

	cacheDir := filepath.Join(t.TempDir(), "stave")

	var buf bytes.Buffer

	CheckAndNotify(context.Background(), Params{
		CurrentVersion: "v0.12.0",
		CacheDir:       cacheDir,
		Output:         &buf,
		Config:         enabledConfig(),
		ClientOptions:  []GitHubClientOption{WithBaseURL(server.URL)},
	})

	cached := ReadCache(cacheDir)
	require.NotNil(t, cached)
	assert.Equal(t, "v0.13.0", cached.LatestVersion)
	assert.Equal(t, "https://github.com/yaklabco/stave/releases/tag/v0.13.0", cached.ReleaseURL)
	assert.Equal(t, "changelog", cached.ReleaseBody)
}

// --- ExplicitCheck tests ---

func TestExplicitCheck_ShowsChangelog(t *testing.T) {
	server := newTestServer(t, "v0.13.0", "## What's Changed\n- New feature")
	defer server.Close()

	var buf bytes.Buffer

	err := ExplicitCheck(context.Background(), Params{
		CurrentVersion: "v0.12.0",
		CacheDir:       filepath.Join(t.TempDir(), "stave"),
		Output:         &buf,
		Config:         enabledConfig(),
		ClientOptions:  []GitHubClientOption{WithBaseURL(server.URL)},
	})

	require.NoError(t, err)

	output := stripANSI(buf.String())
	assert.Contains(t, output, "v0.13.0")
	assert.Contains(t, output, "New feature")
}

func TestExplicitCheck_UpToDate(t *testing.T) {
	server := newTestServer(t, "v0.12.0", "changelog")
	defer server.Close()

	var buf bytes.Buffer

	err := ExplicitCheck(context.Background(), Params{
		CurrentVersion: "v0.12.0",
		CacheDir:       filepath.Join(t.TempDir(), "stave"),
		Output:         &buf,
		Config:         enabledConfig(),
		ClientOptions:  []GitHubClientOption{WithBaseURL(server.URL)},
	})

	require.NoError(t, err)
	assert.Contains(t, buf.String(), "latest version")
}

func TestExplicitCheck_ReturnsNetworkError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	var buf bytes.Buffer

	err := ExplicitCheck(context.Background(), Params{
		CurrentVersion: "v0.12.0",
		CacheDir:       filepath.Join(t.TempDir(), "stave"),
		Output:         &buf,
		Config:         enabledConfig(),
		ClientOptions:  []GitHubClientOption{WithBaseURL(server.URL)},
	})

	require.Error(t, err)
}

func TestExplicitCheck_IgnoresCacheTTL(t *testing.T) {
	var calls atomic.Int32

	server := newCountingServer(t, "v0.13.0", &calls)
	defer server.Close()

	cacheDir := filepath.Join(t.TempDir(), "stave")

	// Pre-write a fresh cache (not expired).
	entry := &CacheEntry{
		CheckedAt:     time.Now(),
		LatestVersion: "v0.12.0",
		ReleaseURL:    "https://github.com/yaklabco/stave/releases/tag/v0.12.0",
	}
	require.NoError(t, WriteCache(cacheDir, entry))

	var buf bytes.Buffer

	err := ExplicitCheck(context.Background(), Params{
		CurrentVersion: "v0.12.0",
		CacheDir:       cacheDir,
		Output:         &buf,
		Config:         enabledConfig(),
		ClientOptions:  []GitHubClientOption{WithBaseURL(server.URL)},
	})

	require.NoError(t, err)
	assert.Equal(t, int32(1), calls.Load(), "ExplicitCheck should always fetch, ignoring cache TTL")
}

// --- isNewer tests ---

func TestIsNewer(t *testing.T) {
	tests := []struct {
		name    string
		latest  string
		current string
		want    bool
	}{
		{name: "newer version", latest: "v0.13.0", current: "v0.12.0", want: true},
		{name: "same version", latest: "v0.12.0", current: "v0.12.0", want: false},
		{name: "older version", latest: "v0.11.0", current: "v0.12.0", want: false},
		{name: "no v prefix", latest: "0.13.0", current: "0.12.0", want: true},
		{name: "mixed prefix", latest: "v0.13.0", current: "0.12.0", want: true},
		{name: "invalid latest", latest: "invalid", current: "v0.12.0", want: false},
		{name: "invalid current", latest: "v0.12.0", current: "invalid", want: false},
		{name: "both empty", latest: "", current: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isNewer(tt.latest, tt.current))
		})
	}
}
