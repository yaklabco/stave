package update

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const cacheFileName = "update-check.json"

// CacheEntry stores the result of a version check so that repeated checks
// within the cache TTL do not require additional network requests.
type CacheEntry struct {
	CheckedAt       time.Time `json:"checked_at"`
	LatestVersion   string    `json:"latest_version"`
	ReleaseURL      string    `json:"release_url"`
	ReleaseBody     string    `json:"release_body,omitempty"`
	NotifiedVersion string    `json:"notified_version,omitempty"`
}

// IsExpired returns true if the cache entry is older than the given TTL.
func (e *CacheEntry) IsExpired(ttl time.Duration) bool {
	return time.Since(e.CheckedAt) > ttl
}

// CachePath returns the full path to the cache file within the given directory.
func CachePath(dir string) string {
	return filepath.Join(dir, cacheFileName)
}

// ReadCache reads and decodes the cache file from the given directory. It
// returns nil if the file does not exist, cannot be read, or contains invalid
// JSON.
func ReadCache(dir string) *CacheEntry {
	data, err := os.ReadFile(CachePath(dir))
	if err != nil {
		return nil
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil
	}

	return &entry
}

// WriteCache encodes and writes the cache entry to the given directory,
// creating the directory structure if it does not exist.
func WriteCache(dir string, entry *CacheEntry) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	return os.WriteFile(CachePath(dir), data, 0o600)
}
