package disposable

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCheckerNew(t *testing.T) {
	checker, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer checker.Close()

	if !checker.IsDisposable("10minutemail.com") {
		t.Error("Expected 10minutemail.com to be disposable")
	}

	if checker.IsDisposable("gmail.com") {
		t.Error("Expected gmail.com to not be disposable")
	}
}

func TestCheckerWithCustomBlocklist(t *testing.T) {
	checker, err := New(
		WithCustomBlocklist("my-custom-domain.com", "another-custom.org"),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer checker.Close()

	// Custom domains should be blocked
	if !checker.IsDisposable("user@my-custom-domain.com") {
		t.Error("Expected my-custom-domain.com to be disposable")
	}
	if !checker.IsDisposable("another-custom.org") {
		t.Error("Expected another-custom.org to be disposable")
	}
}

func TestCheckerWithCustomAllowlist(t *testing.T) {
	checker, err := New(
		WithCustomAllowlist("10minutemail.com"), // Override a known disposable
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer checker.Close()

	// Allowlisted domain should NOT be detected as disposable
	if checker.IsDisposable("10minutemail.com") {
		t.Error("Expected 10minutemail.com to be allowed (not disposable)")
	}
}

func TestCheckerWithCacheDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "disposable-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// First copy data.bin to the temp cache dir
	srcPath := filepath.Join("data", "data.bin")
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		t.Skip("data/data.bin not found, skipping test")
	}

	data, err := os.ReadFile(srcPath)
	if err != nil {
		t.Fatalf("Failed to read data.bin: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "data.bin"), data, 0644); err != nil {
		t.Fatalf("Failed to write data.bin: %v", err)
	}

	checker, err := New(WithCacheDir(tmpDir))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer checker.Close()

	// Should work with the custom cache dir
	if !checker.IsDisposable("10minutemail.com") {
		t.Error("Expected 10minutemail.com to be disposable")
	}
}

func TestCheckerAddDomains(t *testing.T) {
	checker, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer checker.Close()

	// Add domain at runtime
	checker.AddDomains("runtime-added-domain.com")

	if !checker.IsDisposable("runtime-added-domain.com") {
		t.Error("Expected runtime-added-domain.com to be disposable after AddDomains")
	}
}

func TestCheckerAddAllowlist(t *testing.T) {
	checker, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer checker.Close()

	// First verify domain is blocked
	if !checker.IsDisposable("mailinator.com") {
		t.Skip("mailinator.com not in blocklist")
	}

	// Add to allowlist
	checker.AddAllowlist("mailinator.com")

	// Should no longer be detected
	if checker.IsDisposable("mailinator.com") {
		t.Error("Expected mailinator.com to be allowed after AddAllowlist")
	}
}

func TestCheckerStats(t *testing.T) {
	checker, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer checker.Close()

	stats := checker.Stats()

	if stats.BlocklistCount == 0 {
		t.Error("Expected BlocklistCount > 0")
	}
	if stats.Mode != ModeOnline {
		t.Errorf("Expected Mode = ModeOnline, got %v", stats.Mode)
	}
	if stats.Version == "" {
		t.Error("Expected Version to be set")
	}

	t.Logf("Stats: Blocklist=%d, Allowlist=%d, Version=%s",
		stats.BlocklistCount, stats.AllowlistCount, stats.Version)
}

func TestCheckerGetBlocklist(t *testing.T) {
	checker, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer checker.Close()

	blocklist := checker.GetBlocklist()
	if len(blocklist) == 0 {
		t.Error("Expected blocklist to not be empty")
	}
}

func TestCheckerGetAllowlist(t *testing.T) {
	checker, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer checker.Close()

	allowlist := checker.GetAllowlist()
	t.Logf("Allowlist has %d domains", len(allowlist))
}

func TestCheckerIsDisposableWithContext(t *testing.T) {
	checker, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer checker.Close()

	ctx := context.Background()

	if !checker.IsDisposableWithContext(ctx, "10minutemail.com") {
		t.Error("Expected 10minutemail.com to be disposable")
	}

	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Should still work (context only affects initialization)
	result := checker.IsDisposableWithContext(ctx, "10minutemail.com")
	if !result {
		t.Error("Expected 10minutemail.com to be disposable even with cancelled context")
	}
}

func TestCheckerHierarchicalMatching(t *testing.T) {
	checker, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer checker.Close()

	// Add a parent domain
	checker.AddDomains("custom-disposable.com")

	// Subdomains should also be blocked
	tests := []struct {
		domain   string
		expected bool
	}{
		{"custom-disposable.com", true},
		{"sub.custom-disposable.com", true},
		{"deep.sub.custom-disposable.com", true},
		{"gmail.com", false},
	}

	for _, tt := range tests {
		result := checker.IsDisposable(tt.domain)
		if result != tt.expected {
			t.Errorf("IsDisposable(%q) = %v, want %v", tt.domain, result, tt.expected)
		}
	}
}

func TestCheckerClose(t *testing.T) {
	checker, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Close should not error
	if err := checker.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Calling Close again should be safe
	if err := checker.Close(); err != nil {
		t.Errorf("Second Close() error = %v", err)
	}
}

func TestCheckerWithHTTPTimeout(t *testing.T) {
	// Just verify the option is accepted
	checker, err := New(
		WithHTTPTimeout(10 * time.Second),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer checker.Close()

	// Should still work
	if !checker.IsDisposable("10minutemail.com") {
		t.Error("Expected 10minutemail.com to be disposable")
	}
}

func TestCheckerWithAutoRefresh(t *testing.T) {
	checker, err := New(
		WithAutoRefresh(100 * time.Millisecond), // Very short interval for testing
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Verify it works initially
	if !checker.IsDisposable("mailinator.com") {
		t.Error("Expected mailinator.com to be disposable")
	}

	// Wait a bit to allow auto-refresh to potentially run
	time.Sleep(150 * time.Millisecond)

	// Should still work after potential refresh
	if !checker.IsDisposable("mailinator.com") {
		t.Error("Expected mailinator.com to be disposable after potential auto-refresh")
	}

	// Close should stop the auto-refresh goroutine
	if err := checker.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestCheckerCloseStopsAutoRefresh(t *testing.T) {
	checker, err := New(
		WithAutoRefresh(50 * time.Millisecond),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Close should return promptly (not block indefinitely)
	done := make(chan struct{})
	go func() {
		checker.Close()
		close(done)
	}()

	select {
	case <-done:
		// Success - Close returned
	case <-time.After(2 * time.Second):
		t.Error("Close() did not return within timeout - possible goroutine leak")
	}
}

func TestCheckerCorruptedCache(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "disposable-corrupt-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write corrupted data to cache
	corruptedData := []byte("this is not valid gob/gzip data")
	if err := os.WriteFile(filepath.Join(tmpDir, "data.bin"), corruptedData, 0644); err != nil {
		t.Fatalf("Failed to write corrupted data: %v", err)
	}

	// Creating a checker should fail or fall back to download
	// Since we're in a test environment with real network, it should download
	checker, err := New(WithCacheDir(tmpDir))
	if err != nil {
		// If download fails too, that's OK for this test
		var initErr *InitializationError
		if IsInitializationError(err) {
			t.Logf("Expected error type: InitializationError: %v", err)
		} else {
			t.Logf("Got error (download also failed, which is OK): %v", initErr)
		}
		return
	}
	defer checker.Close()

	// If we got here, the checker recovered by downloading fresh data
	if !checker.IsDisposable("mailinator.com") {
		t.Error("Expected mailinator.com to be disposable after recovering from corrupted cache")
	}
}

func TestCheckerRefresh(t *testing.T) {
	checker, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer checker.Close()

	// Get initial stats
	initialStats := checker.Stats()

	// Call Refresh
	if err := checker.Refresh(); err != nil {
		// Network errors are acceptable in test environments
		if IsDownloadError(err) {
			t.Skipf("Skipping refresh test due to network error: %v", err)
		}
		t.Errorf("Refresh() error = %v", err)
	}

	// Stats should still be valid after refresh
	afterStats := checker.Stats()
	if afterStats.BlocklistCount == 0 {
		t.Error("Expected BlocklistCount > 0 after refresh")
	}

	t.Logf("Before refresh: %d blocklist, After refresh: %d blocklist",
		initialStats.BlocklistCount, afterStats.BlocklistCount)
}

func TestCheckerRefreshWithContext(t *testing.T) {
	checker, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer checker.Close()

	// Test with a context that has a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := checker.RefreshWithContext(ctx); err != nil {
		if IsDownloadError(err) {
			t.Skipf("Skipping refresh test due to network error: %v", err)
		}
		t.Errorf("RefreshWithContext() error = %v", err)
	}
}

func TestCheckerConcurrentAccess(t *testing.T) {
	checker, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer checker.Close()

	// Run multiple goroutines accessing the checker concurrently
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				checker.IsDisposable("test@mailinator.com")
				checker.IsDisposable("test@gmail.com")
				_ = checker.Stats()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestCheckerInvalidDataURL(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "disposable-invalid-url-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Try with invalid URL and no cache
	_, err = New(
		WithCacheDir(tmpDir),
		WithDataURL("https://invalid.example.com/nonexistent/data.bin"),
		WithHTTPTimeout(2*time.Second),
	)

	// Should fail since there's no cache and URL is invalid
	if err == nil {
		t.Error("Expected error when using invalid URL with no cache")
	}

	// Error should be an InitializationError wrapping a DownloadError
	if !IsInitializationError(err) {
		t.Errorf("Expected InitializationError, got %T: %v", err, err)
	}

	var initErr *InitializationError
	if errors.As(err, &initErr) && initErr.Err != nil {
		if !IsDownloadError(initErr.Err) {
			t.Logf("Underlying error is not DownloadError: %T: %v", initErr.Err, initErr.Err)
		}
	}
}

func TestCheckerEmptyDomain(t *testing.T) {
	checker, err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer checker.Close()

	// Empty strings should return false, not panic
	if checker.IsDisposable("") {
		t.Error("Expected empty string to not be disposable")
	}

	if checker.IsDisposable("@") {
		t.Error("Expected '@' to not be disposable")
	}

	if checker.IsDisposable("user@") {
		t.Error("Expected 'user@' to not be disposable")
	}
}
