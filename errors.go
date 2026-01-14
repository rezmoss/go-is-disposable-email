package disposable

import (
	"errors"
	"fmt"
)

// Error types for programmatic error handling.
// Use errors.Is() or errors.As() to check error types.

// ErrNotInitialized is returned when operations are attempted before initialization.
var ErrNotInitialized = errors.New("checker not initialized")

// DownloadError represents an error that occurred while downloading data.
type DownloadError struct {
	URL        string
	StatusCode int    // HTTP status code, 0 if not an HTTP error
	Err        error  // underlying error
}

func (e *DownloadError) Error() string {
	if e.StatusCode != 0 {
		return fmt.Sprintf("download failed from %s: HTTP %d", e.URL, e.StatusCode)
	}
	return fmt.Sprintf("download failed from %s: %v", e.URL, e.Err)
}

func (e *DownloadError) Unwrap() error {
	return e.Err
}

// CacheError represents an error related to cache operations.
type CacheError struct {
	Path      string
	Operation string // "read", "write", "create"
	Err       error
}

func (e *CacheError) Error() string {
	return fmt.Sprintf("cache %s failed for %s: %v", e.Operation, e.Path, e.Err)
}

func (e *CacheError) Unwrap() error {
	return e.Err
}

// DeserializationError represents an error deserializing the data file.
type DeserializationError struct {
	Source string // "cache" or "download"
	Err    error
}

func (e *DeserializationError) Error() string {
	return fmt.Sprintf("failed to deserialize data from %s: %v", e.Source, e.Err)
}

func (e *DeserializationError) Unwrap() error {
	return e.Err
}

// InitializationError represents an error during checker initialization.
type InitializationError struct {
	Reason string
	Err    error
}

func (e *InitializationError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("initialization failed: %s: %v", e.Reason, e.Err)
	}
	return fmt.Sprintf("initialization failed: %s", e.Reason)
}

func (e *InitializationError) Unwrap() error {
	return e.Err
}

// IsDownloadError returns true if the error is a download error.
func IsDownloadError(err error) bool {
	var downloadErr *DownloadError
	return errors.As(err, &downloadErr)
}

// IsCacheError returns true if the error is a cache error.
func IsCacheError(err error) bool {
	var cacheErr *CacheError
	return errors.As(err, &cacheErr)
}

// IsDeserializationError returns true if the error is a deserialization error.
func IsDeserializationError(err error) bool {
	var deserErr *DeserializationError
	return errors.As(err, &deserErr)
}

// IsInitializationError returns true if the error is an initialization error.
func IsInitializationError(err error) bool {
	var initErr *InitializationError
	return errors.As(err, &initErr)
}
