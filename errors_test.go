package disposable

import (
	"errors"
	"testing"
)

func TestDownloadError(t *testing.T) {
	// Test with underlying error
	underlyingErr := errors.New("connection refused")
	err := &DownloadError{URL: "https://example.com/data.bin", Err: underlyingErr}

	if !IsDownloadError(err) {
		t.Error("IsDownloadError should return true")
	}

	expectedMsg := "download failed from https://example.com/data.bin: connection refused"
	if err.Error() != expectedMsg {
		t.Errorf("Error() = %q, want %q", err.Error(), expectedMsg)
	}

	if !errors.Is(err, underlyingErr) {
		t.Error("errors.Is should find underlying error")
	}

	// Test with HTTP status code
	httpErr := &DownloadError{URL: "https://example.com/data.bin", StatusCode: 404}
	expectedHTTPMsg := "download failed from https://example.com/data.bin: HTTP 404"
	if httpErr.Error() != expectedHTTPMsg {
		t.Errorf("Error() = %q, want %q", httpErr.Error(), expectedHTTPMsg)
	}
}

func TestCacheError(t *testing.T) {
	underlyingErr := errors.New("permission denied")
	err := &CacheError{Path: "/tmp/cache/data.bin", Operation: "read", Err: underlyingErr}

	if !IsCacheError(err) {
		t.Error("IsCacheError should return true")
	}

	expectedMsg := "cache read failed for /tmp/cache/data.bin: permission denied"
	if err.Error() != expectedMsg {
		t.Errorf("Error() = %q, want %q", err.Error(), expectedMsg)
	}

	if !errors.Is(err, underlyingErr) {
		t.Error("errors.Is should find underlying error")
	}
}

func TestDeserializationError(t *testing.T) {
	underlyingErr := errors.New("invalid gob data")
	err := &DeserializationError{Source: "cache", Err: underlyingErr}

	if !IsDeserializationError(err) {
		t.Error("IsDeserializationError should return true")
	}

	expectedMsg := "failed to deserialize data from cache: invalid gob data"
	if err.Error() != expectedMsg {
		t.Errorf("Error() = %q, want %q", err.Error(), expectedMsg)
	}

	if !errors.Is(err, underlyingErr) {
		t.Error("errors.Is should find underlying error")
	}
}

func TestInitializationError(t *testing.T) {
	// Test with underlying error
	underlyingErr := errors.New("network unreachable")
	err := &InitializationError{Reason: "download failed", Err: underlyingErr}

	if !IsInitializationError(err) {
		t.Error("IsInitializationError should return true")
	}

	expectedMsg := "initialization failed: download failed: network unreachable"
	if err.Error() != expectedMsg {
		t.Errorf("Error() = %q, want %q", err.Error(), expectedMsg)
	}

	if !errors.Is(err, underlyingErr) {
		t.Error("errors.Is should find underlying error")
	}

	// Test without underlying error
	errNoUnderlying := &InitializationError{Reason: "cache directory not found"}
	expectedNoUnderlyingMsg := "initialization failed: cache directory not found"
	if errNoUnderlying.Error() != expectedNoUnderlyingMsg {
		t.Errorf("Error() = %q, want %q", errNoUnderlying.Error(), expectedNoUnderlyingMsg)
	}
}

func TestErrorTypeChecks(t *testing.T) {
	// Test that type check functions return false for non-matching errors
	genericErr := errors.New("some error")

	if IsDownloadError(genericErr) {
		t.Error("IsDownloadError should return false for non-download error")
	}
	if IsCacheError(genericErr) {
		t.Error("IsCacheError should return false for non-cache error")
	}
	if IsDeserializationError(genericErr) {
		t.Error("IsDeserializationError should return false for non-deserialization error")
	}
	if IsInitializationError(genericErr) {
		t.Error("IsInitializationError should return false for non-initialization error")
	}
}

func TestWrappedErrors(t *testing.T) {
	// Test that errors.As works with wrapped errors
	downloadErr := &DownloadError{URL: "https://example.com", Err: errors.New("timeout")}
	initErr := &InitializationError{Reason: "no data", Err: downloadErr}

	// Should be able to extract the download error from the initialization error
	var extractedDownloadErr *DownloadError
	if !errors.As(initErr, &extractedDownloadErr) {
		t.Error("errors.As should find wrapped DownloadError")
	}

	if extractedDownloadErr.URL != "https://example.com" {
		t.Errorf("Extracted URL = %q, want %q", extractedDownloadErr.URL, "https://example.com")
	}
}
