package disposable

import (
	"context"
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
