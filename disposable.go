// Package disposable provides functionality to detect disposable/temporary email addresses.
//
// The package maintains a database of known disposable email domains that is:
//   - Downloaded on first use from GitHub releases
//   - Cached locally for fast subsequent access
//   - Automatically updated daily from multiple sources
//   - Extensible with custom domains at runtime
//
// Basic usage:
//
//	if disposable.IsDisposable("user@tempmail.com") {
//	    // Handle disposable email
//	}
//
// For custom configuration, use the Checker type:
//
//	checker, _ := disposable.New(
//	    disposable.WithAutoRefresh(24 * time.Hour),
//	    disposable.WithCustomBlocklist("internal-blocked.com"),
//	)
//	result := checker.IsDisposable("user@example.com")
package disposable

import (
	"context"
	"sync"
)

var (
	defaultChecker     *Checker
	defaultCheckerOnce sync.Once
	defaultCheckerErr  error
)

// getDefaultChecker returns the default checker, initializing it if needed.
// On first use, it downloads the data.bin file from GitHub releases.
func getDefaultChecker() (*Checker, error) {
	defaultCheckerOnce.Do(func() {
		// Download data on first use, cache locally
		defaultChecker, defaultCheckerErr = New()
	})
	return defaultChecker, defaultCheckerErr
}

// IsDisposable checks if an email address or domain is from a disposable email service.
// It accepts either a full email address ("user@tempmail.com") or just a domain ("tempmail.com").
// Returns true if the domain is disposable, false otherwise.
//
// On first call, this function downloads the domain database if not already cached.
// Subsequent calls use the cached data.
func IsDisposable(emailOrDomain string) bool {
	checker, err := getDefaultChecker()
	if err != nil {
		return false
	}
	return checker.IsDisposable(emailOrDomain)
}

// IsDisposableWithContext is like IsDisposable but accepts a context for cancellation.
func IsDisposableWithContext(ctx context.Context, emailOrDomain string) bool {
	checker, err := getDefaultChecker()
	if err != nil {
		return false
	}
	return checker.IsDisposableWithContext(ctx, emailOrDomain)
}

// Refresh updates the domain database by downloading fresh data from the source.
func Refresh() error {
	checker, err := getDefaultChecker()
	if err != nil {
		return err
	}
	return checker.Refresh()
}

// RefreshWithContext is like Refresh but accepts a context for cancellation/timeout.
func RefreshWithContext(ctx context.Context) error {
	checker, err := getDefaultChecker()
	if err != nil {
		return err
	}
	return checker.RefreshWithContext(ctx)
}

// AddDomains adds custom domains to the blocklist at runtime.
// These additions are not persisted and will be lost on program restart.
func AddDomains(domains ...string) {
	checker, err := getDefaultChecker()
	if err != nil {
		return
	}
	checker.AddDomains(domains...)
}

// AddAllowlist adds domains to the allowlist at runtime.
// Allowlisted domains will never be reported as disposable.
func AddAllowlist(domains ...string) {
	checker, err := getDefaultChecker()
	if err != nil {
		return
	}
	checker.AddAllowlist(domains...)
}

// GetBlocklist returns a copy of all blocked domains.
func GetBlocklist() []string {
	checker, err := getDefaultChecker()
	if err != nil {
		return nil
	}
	return checker.GetBlocklist()
}

// GetAllowlist returns a copy of all allowlisted domains.
func GetAllowlist() []string {
	checker, err := getDefaultChecker()
	if err != nil {
		return nil
	}
	return checker.GetAllowlist()
}

// Stats returns statistics about the current database.
func Stats() Statistics {
	checker, err := getDefaultChecker()
	if err != nil {
		return Statistics{}
	}
	return checker.Stats()
}
