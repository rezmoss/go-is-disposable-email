package disposable

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rezmoss/go-is-disposable-email/data"
	"github.com/rezmoss/go-is-disposable-email/internal/trie"
)

// Checker performs disposable email detection with custom configuration.
type Checker struct {
	config    *Config
	blocklist *trie.Trie
	allowlist *trie.Trie

	mu          sync.RWMutex
	initialized bool
	lastUpdated time.Time
	version     string

	cancelFunc context.CancelFunc
	wg         sync.WaitGroup
}

// New creates a new Checker with the given options.
func New(opts ...Option) (*Checker, error) {
	config := DefaultConfig()
	for _, opt := range opts {
		opt(config)
	}

	// Set default cache directory if not specified
	if config.CacheDir == "" {
		cacheDir, err := getDefaultCacheDir()
		if err != nil {
			return nil, &InitializationError{Reason: "failed to get cache directory", Err: err}
		}
		config.CacheDir = cacheDir
	}

	// Ensure cache directory exists
	if err := os.MkdirAll(config.CacheDir, 0755); err != nil {
		return nil, &CacheError{Path: config.CacheDir, Operation: "create", Err: err}
	}

	c := &Checker{
		config:    config,
		blocklist: trie.New(),
		allowlist: trie.New(),
	}

	// Initialize - download data if needed
	if err := c.init(context.Background()); err != nil {
		return nil, err
	}

	// Start auto-refresh if configured
	if config.AutoRefresh {
		ctx, cancel := context.WithCancel(context.Background())
		c.cancelFunc = cancel
		c.wg.Add(1)
		go c.autoRefreshWorker(ctx)
	}

	return c, nil
}

// getDefaultCacheDir returns the default cache directory for storing data.bin.
func getDefaultCacheDir() (string, error) {
	// Try user cache directory first
	cacheDir, err := os.UserCacheDir()
	if err == nil {
		return filepath.Join(cacheDir, "disposable-email"), nil
	}

	// Fall back to temp directory
	return filepath.Join(os.TempDir(), "disposable-email"), nil
}

// getDataFilePath returns the path to the data.bin file.
func (c *Checker) getDataFilePath() string {
	return filepath.Join(c.config.CacheDir, data.DataFileName)
}

// init initializes the checker by loading data.
func (c *Checker) init(ctx context.Context) error {
	// Try to load from cache first
	if err := c.loadFromCache(); err == nil {
		c.config.Logger.Printf("Loaded data from cache: %s", c.getDataFilePath())
		c.applyCustomDomains()
		return nil
	}

	// Download fresh data
	c.config.Logger.Printf("Downloading data from %s...", c.config.DataURL)
	if err := c.downloadAndLoad(ctx); err != nil {
		return &InitializationError{Reason: "no cached data and download failed", Err: err}
	}

	c.applyCustomDomains()
	return nil
}

// applyCustomDomains adds custom blocklist/allowlist domains.
func (c *Checker) applyCustomDomains() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, domain := range c.config.CustomBlocklist {
		c.blocklist.Insert(NormalizeDomain(domain))
	}
	for _, domain := range c.config.CustomAllowlist {
		c.allowlist.Insert(NormalizeDomain(domain))
	}
}

// loadFromCache loads data from the cached data.bin file.
func (c *Checker) loadFromCache() error {
	dataPath := c.getDataFilePath()

	fileData, err := os.ReadFile(dataPath)
	if err != nil {
		return &CacheError{Path: dataPath, Operation: "read", Err: err}
	}

	blocklist, allowlist, dataFile, err := trie.Deserialize(fileData)
	if err != nil {
		return &DeserializationError{Source: "cache", Err: err}
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.blocklist = blocklist
	c.allowlist = allowlist
	c.initialized = true
	c.lastUpdated = dataFile.CreatedAt
	c.version = dataFile.Version

	return nil
}

// downloadAndLoad downloads fresh data and loads it.
func (c *Checker) downloadAndLoad(ctx context.Context) error {
	// Download data
	fileData, err := c.downloadData(ctx)
	if err != nil {
		return err
	}

	// Deserialize to validate
	blocklist, allowlist, dataFile, err := trie.Deserialize(fileData)
	if err != nil {
		return &DeserializationError{Source: "download", Err: err}
	}

	// Save to cache
	dataPath := c.getDataFilePath()
	if err := os.WriteFile(dataPath, fileData, 0644); err != nil {
		c.config.Logger.Printf("Warning: failed to save to cache: %v", err)
		// Continue anyway - we have the data in memory
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.blocklist = blocklist
	c.allowlist = allowlist
	c.initialized = true
	c.lastUpdated = dataFile.CreatedAt
	c.version = dataFile.Version

	c.config.Logger.Printf("Loaded %d blocklist and %d allowlist domains (version: %s)",
		blocklist.Size(), allowlist.Size(), dataFile.Version)

	return nil
}

// downloadData downloads fresh data from the configured URL.
func (c *Checker) downloadData(ctx context.Context) ([]byte, error) {
	client := &http.Client{
		Timeout: c.config.HTTPTimeout,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.config.DataURL, nil)
	if err != nil {
		return nil, &DownloadError{URL: c.config.DataURL, Err: err}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, &DownloadError{URL: c.config.DataURL, Err: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &DownloadError{URL: c.config.DataURL, StatusCode: resp.StatusCode}
	}

	fileData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &DownloadError{URL: c.config.DataURL, Err: err}
	}

	return fileData, nil
}

// autoRefreshWorker periodically refreshes the data.
func (c *Checker) autoRefreshWorker(ctx context.Context) {
	defer c.wg.Done()

	ticker := time.NewTicker(c.config.RefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := c.RefreshWithContext(ctx); err != nil {
				c.config.Logger.Printf("Auto-refresh failed: %v", err)
			} else {
				c.config.Logger.Printf("Auto-refresh completed successfully")
			}
		}
	}
}

// IsDisposable checks if an email address or domain is from a disposable email service.
func (c *Checker) IsDisposable(emailOrDomain string) bool {
	return c.IsDisposableWithContext(context.Background(), emailOrDomain)
}

// IsDisposableWithContext is like IsDisposable but accepts a context for cancellation.
func (c *Checker) IsDisposableWithContext(ctx context.Context, emailOrDomain string) bool {
	domain := ExtractDomain(emailOrDomain)
	if domain == "" {
		return false
	}

	domain = NormalizeDomain(domain)

	c.mu.RLock()
	defer c.mu.RUnlock()

	// Check allowlist first (takes precedence)
	if c.allowlist.ContainsHierarchical(domain) {
		return false
	}

	// Check blocklist with hierarchical matching
	return c.blocklist.ContainsHierarchical(domain)
}

// Refresh updates the domain database by downloading fresh data.
func (c *Checker) Refresh() error {
	return c.RefreshWithContext(context.Background())
}

// RefreshWithContext is like Refresh but accepts a context for cancellation/timeout.
func (c *Checker) RefreshWithContext(ctx context.Context) error {
	if err := c.downloadAndLoad(ctx); err != nil {
		return err // Already a typed error (DownloadError or DeserializationError)
	}
	c.applyCustomDomains()
	return nil
}

// AddDomains adds custom domains to the blocklist at runtime.
func (c *Checker) AddDomains(domains ...string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, domain := range domains {
		c.blocklist.Insert(NormalizeDomain(domain))
	}
}

// AddAllowlist adds domains to the allowlist at runtime.
func (c *Checker) AddAllowlist(domains ...string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, domain := range domains {
		c.allowlist.Insert(NormalizeDomain(domain))
	}
}

// GetBlocklist returns a copy of all blocked domains.
func (c *Checker) GetBlocklist() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.blocklist.GetAll()
}

// GetAllowlist returns a copy of all allowlisted domains.
func (c *Checker) GetAllowlist() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.allowlist.GetAll()
}

// Stats returns statistics about the current database.
func (c *Checker) Stats() Statistics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return Statistics{
		BlocklistCount: c.blocklist.Size(),
		AllowlistCount: c.allowlist.Size(),
		LastUpdated:    c.lastUpdated,
		Mode:           c.config.Mode,
		Version:        c.version,
	}
}

// Close releases resources held by the Checker and stops the auto-refresh goroutine.
//
// Close MUST be called when you are done using a Checker that was created with
// WithAutoRefresh. Failing to call Close will result in a goroutine leak.
// It is safe to call Close multiple times; subsequent calls are no-ops.
//
// For Checkers without auto-refresh, calling Close is optional but recommended
// for consistency.
func (c *Checker) Close() error {
	if c.cancelFunc != nil {
		c.cancelFunc()
	}
	c.wg.Wait()
	return nil
}
