package disposable

import (
	"io"
	"log"
	"time"

	"github.com/rezmoss/go-is-disposable-email/data"
)

// Mode determines how the checker operates regarding network access.
type Mode int

const (
	// ModeOnline auto-downloads fresh data on first use, caches locally.
	// This is the only supported mode - data.bin is always external.
	ModeOnline Mode = iota
)

// String returns the string representation of the Mode.
func (m Mode) String() string {
	switch m {
	case ModeOnline:
		return "online"
	default:
		return "unknown"
	}
}

// Logger is the interface for logging operations.
type Logger interface {
	Printf(format string, v ...any)
}

// Config holds all configuration for a Checker.
type Config struct {
	// Mode controls online/offline behavior. Default: ModeOnline
	Mode Mode

	// AutoRefresh enables automatic periodic updates.
	AutoRefresh bool

	// RefreshInterval sets how often to auto-refresh. Default: 24h
	RefreshInterval time.Duration

	// CacheDir specifies where to cache downloaded data.
	// Default: os.UserCacheDir()/disposable
	CacheDir string

	// HTTPTimeout for download operations. Default: 30s
	HTTPTimeout time.Duration

	// CustomBlocklist adds extra domains to block at initialization.
	CustomBlocklist []string

	// CustomAllowlist adds extra domains to allow at initialization.
	CustomAllowlist []string

	// Logger for diagnostic output. Default: discards logs
	Logger Logger

	// DataURL is the URL to download data.bin from for updates.
	// Default: GitHub releases URL
	DataURL string
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		Mode:            ModeOnline,
		AutoRefresh:     false,
		RefreshInterval: 24 * time.Hour,
		CacheDir:        "",
		HTTPTimeout:     30 * time.Second,
		CustomBlocklist: nil,
		CustomAllowlist: nil,
		Logger:          log.New(io.Discard, "", 0),
		DataURL:         data.DefaultDataURL,
	}
}

// Option configures a Checker.
type Option func(*Config)

// WithMode sets the operating mode.
func WithMode(mode Mode) Option {
	return func(c *Config) {
		c.Mode = mode
	}
}

// WithAutoRefresh enables automatic updates at the specified interval.
func WithAutoRefresh(interval time.Duration) Option {
	return func(c *Config) {
		c.AutoRefresh = true
		if interval > 0 {
			c.RefreshInterval = interval
		}
	}
}

// WithCacheDir sets the cache directory for downloaded data.
func WithCacheDir(dir string) Option {
	return func(c *Config) {
		c.CacheDir = dir
	}
}

// WithHTTPTimeout sets the timeout for HTTP operations.
func WithHTTPTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.HTTPTimeout = timeout
	}
}

// WithCustomBlocklist adds custom domains to block.
func WithCustomBlocklist(domains ...string) Option {
	return func(c *Config) {
		c.CustomBlocklist = append(c.CustomBlocklist, domains...)
	}
}

// WithCustomAllowlist adds custom domains to allow.
func WithCustomAllowlist(domains ...string) Option {
	return func(c *Config) {
		c.CustomAllowlist = append(c.CustomAllowlist, domains...)
	}
}

// WithLogger sets a custom logger.
func WithLogger(logger Logger) Option {
	return func(c *Config) {
		c.Logger = logger
	}
}

// WithDataURL sets a custom URL for downloading data.bin updates.
func WithDataURL(url string) Option {
	return func(c *Config) {
		c.DataURL = url
	}
}

// Statistics contains information about the current database state.
type Statistics struct {
	BlocklistCount int       // Number of blocked domains
	AllowlistCount int       // Number of allowlisted domains
	LastUpdated    time.Time // When the database was last updated
	Mode           Mode      // Current operating mode
	Version        string    // Version of the data
}
