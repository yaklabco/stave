package update

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeJSON is a test helper that writes a JSON response body to an
// httptest.ResponseWriter. It fails the test if the write errors.
func writeJSON(t *testing.T, rw http.ResponseWriter, body string) {
	t.Helper()

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)

	_, err := rw.Write([]byte(body))
	require.NoError(t, err)
}

func TestFetchLatestRelease_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, `{
			"tag_name": "v0.10.10",
			"html_url": "https://github.com/yaklabco/stave/releases/tag/v0.10.10",
			"body": "## What's Changed\n- Bug fixes"
		}`)
	}))
	defer server.Close()

	client := NewGitHubClient(WithBaseURL(server.URL))
	release, err := client.FetchLatestRelease(context.Background())

	require.NoError(t, err)
	assert.Equal(t, "v0.10.10", release.TagName)
	assert.Equal(t, "https://github.com/yaklabco/stave/releases/tag/v0.10.10", release.HTMLURL)
	assert.Equal(t, "## What's Changed\n- Bug fixes", release.Body)
}

func TestFetchLatestRelease_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewGitHubClient(WithBaseURL(server.URL))
	release, err := client.FetchLatestRelease(context.Background())

	require.Error(t, err)
	assert.Nil(t, release)
	assert.Contains(t, err.Error(), "unexpected status 404")
}

func TestFetchLatestRelease_RateLimited(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	client := NewGitHubClient(WithBaseURL(server.URL))
	release, err := client.FetchLatestRelease(context.Background())

	require.Error(t, err)
	assert.Nil(t, release)
	assert.Contains(t, err.Error(), "unexpected status 403")
}

func TestFetchLatestRelease_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, `not valid json at all`)
	}))
	defer server.Close()

	client := NewGitHubClient(WithBaseURL(server.URL))
	release, err := client.FetchLatestRelease(context.Background())

	require.Error(t, err)
	assert.Nil(t, release)
	assert.Contains(t, err.Error(), "decoding response")
}

func TestFetchLatestRelease_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(1 * time.Second):
			w.WriteHeader(http.StatusOK)
		case <-r.Context().Done():
			return
		}
	}))
	defer server.Close()

	httpClient := &http.Client{Timeout: 100 * time.Millisecond}
	client := NewGitHubClient(
		WithBaseURL(server.URL),
		WithHTTPClient(httpClient),
	)

	release, err := client.FetchLatestRelease(context.Background())

	require.Error(t, err)
	assert.Nil(t, release)
}

func TestFetchLatestRelease_SetsAcceptHeader(t *testing.T) {
	var receivedAccept string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAccept = r.Header.Get("Accept")
		writeJSON(t, w, `{"tag_name": "v1.0.0", "html_url": "", "body": ""}`)
	}))
	defer server.Close()

	client := NewGitHubClient(WithBaseURL(server.URL))
	_, err := client.FetchLatestRelease(context.Background())

	require.NoError(t, err)
	assert.Equal(t, "application/vnd.github+json", receivedAccept)
}

func TestFetchLatestRelease_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	client := NewGitHubClient(WithBaseURL(server.URL))
	release, err := client.FetchLatestRelease(ctx)

	require.Error(t, err)
	assert.Nil(t, release)
}

func TestFetchLatestRelease_RequestsCorrectPath(t *testing.T) {
	var receivedPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		writeJSON(t, w, `{"tag_name": "v1.0.0", "html_url": "", "body": ""}`)
	}))
	defer server.Close()

	client := NewGitHubClient(WithBaseURL(server.URL))
	_, err := client.FetchLatestRelease(context.Background())

	require.NoError(t, err)
	assert.Equal(t, "/repos/yaklabco/stave/releases/latest", receivedPath)
}
