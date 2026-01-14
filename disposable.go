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
// For production systems that need error handling, use the error-returning variants:
//
//	isDisposable, err := disposable.CheckEmail("user@tempmail.com")
//	if err != nil {
//	    // Handle initialization error (network, cache, etc.)
//	}
//
// For custom configuration, use the Checker type:
//
//	checker, err := disposable.New(
//	    disposable.WithAutoRefresh(24 * time.Hour),
//	    disposable.WithCustomBlocklist("internal-blocked.com"),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer checker.Close()
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
//
// Note: This function returns false on initialization errors (network failure, cache issues).
// For production systems that need to distinguish between "not disposable" and "error",
// use CheckEmail instead.
func IsDisposable(emailOrDomain string) bool {
	checker, err := getDefaultChecker()
	if err != nil {
		return false
	}
	return checker.IsDisposable(emailOrDomain)
}

// IsDisposableWithContext is like IsDisposable but accepts a context for cancellation.
//
// Note: Returns false on initialization errors. Use CheckEmailWithContext for error handling.
func IsDisposableWithContext(ctx context.Context, emailOrDomain string) bool {
	checker, err := getDefaultChecker()
	if err != nil {
		return false
	}
	return checker.IsDisposableWithContext(ctx, emailOrDomain)
}

// CheckEmail checks if an email address or domain is from a disposable email service.
// Unlike IsDisposable, this function returns an error if the checker fails to initialize.
//
// Returns:
//   - (true, nil) if the domain is disposable
//   - (false, nil) if the domain is not disposable
//   - (false, error) if the checker failed to initialize
func CheckEmail(emailOrDomain string) (bool, error) {
	checker, err := getDefaultChecker()
	if err != nil {
		return false, err
	}
	return checker.IsDisposable(emailOrDomain), nil
}

// CheckEmailWithContext is like CheckEmail but accepts a context for cancellation.
func CheckEmailWithContext(ctx context.Context, emailOrDomain string) (bool, error) {
	checker, err := getDefaultChecker()
	if err != nil {
		return false, err
	}
	return checker.IsDisposableWithContext(ctx, emailOrDomain), nil
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
//
// Note: Silently fails if the checker is not initialized. Use IsReady() to check status.
func AddDomains(domains ...string) {
	checker, err := getDefaultChecker()
	if err != nil {
		return
	}
	checker.AddDomains(domains...)
}

// AddAllowlist adds domains to the allowlist at runtime.
// Allowlisted domains will never be reported as disposable.
//
// Note: Silently fails if the checker is not initialized. Use IsReady() to check status.
func AddAllowlist(domains ...string) {
	checker, err := getDefaultChecker()
	if err != nil {
		return
	}
	checker.AddAllowlist(domains...)
}

// GetBlocklist returns a copy of all blocked domains.
//
// Note: Returns nil if the checker is not initialized. Use IsReady() to check status.
func GetBlocklist() []string {
	checker, err := getDefaultChecker()
	if err != nil {
		return nil
	}
	return checker.GetBlocklist()
}

// GetAllowlist returns a copy of all allowlisted domains.
//
// Note: Returns nil if the checker is not initialized. Use IsReady() to check status.
func GetAllowlist() []string {
	checker, err := getDefaultChecker()
	if err != nil {
		return nil
	}
	return checker.GetAllowlist()
}

// Stats returns statistics about the current database.
//
// Note: Returns empty Statistics if the checker is not initialized. Use IsReady() to check status.
func Stats() Statistics {
	checker, err := getDefaultChecker()
	if err != nil {
		return Statistics{}
	}
	return checker.Stats()
}

// IsReady returns true if the default checker is initialized and ready for use.
// This can be used to verify that the domain database was loaded successfully.
func IsReady() bool {
	checker, err := getDefaultChecker()
	return err == nil && checker != nil
}

// InitError returns the initialization error if the default checker failed to initialize.
// Returns nil if the checker is ready or hasn't been initialized yet.
func InitError() error {
	_, err := getDefaultChecker()
	return err
}
