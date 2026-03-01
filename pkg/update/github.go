// Package update provides version checking against GitHub releases.
package update

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	defaultOwner      = "yaklabco"
	defaultRepo       = "stave"
	defaultTimeout    = 5 * time.Second
	latestReleasePath = "/repos/%s/%s/releases/latest"
	githubAPIBase     = "https://api.github.com"
)

// Release represents a GitHub release with the fields needed for version checking.
type Release struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
	Body    string `json:"body"`
}

// GitHubClient fetches release information from the GitHub API.
type GitHubClient struct {
	httpClient *http.Client
	baseURL    string
	owner      string
	repo       string
}

// GitHubClientOption configures a GitHubClient.
type GitHubClientOption func(*GitHubClient)

// WithBaseURL sets the base URL for API requests. This is primarily useful
// for testing with httptest servers.
func WithBaseURL(url string) GitHubClientOption {
	return func(c *GitHubClient) {
		c.baseURL = url
	}
}

// WithHTTPClient sets the HTTP client used for API requests. This allows
// callers to configure custom timeouts or transports.
func WithHTTPClient(hc *http.Client) GitHubClientOption {
	return func(c *GitHubClient) {
		c.httpClient = hc
	}
}

// NewGitHubClient creates a GitHubClient with sensible defaults for querying
// the yaklabco/stave repository. Use functional options to override defaults.
func NewGitHubClient(opts ...GitHubClientOption) *GitHubClient {
	gc := &GitHubClient{
		httpClient: &http.Client{Timeout: defaultTimeout},
		baseURL:    githubAPIBase,
		owner:      defaultOwner,
		repo:       defaultRepo,
	}

	for _, opt := range opts {
		opt(gc)
	}

	return gc
}

// FetchLatestRelease queries the GitHub Releases API for the latest release
// of the configured repository. It returns the release metadata or an error
// if the request fails, the server returns a non-200 status, or the response
// body cannot be decoded.
func (c *GitHubClient) FetchLatestRelease(ctx context.Context) (*Release, error) {
	url := c.baseURL + fmt.Sprintf(latestReleasePath, c.owner, c.repo)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.httpClient.Do(req) //nolint:gosec // URL built from trusted owner/repo config
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &release, nil
}
