package disposable

import (
	"context"
	"testing"
)

func TestIsDisposable(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Known disposable domains (verified in the blocklist)
		{"10minutemail.com", "user@10minutemail.com", true},
		{"guerrillamail.com", "guerrillamail.com", true},
		{"mailinator.com", "test@mailinator.com", true},
		{"yopmail.com", "yopmail.com", true},
		{"trashmail.com", "user@trashmail.com", true},

		// Legitimate emails
		{"gmail.com", "user@gmail.com", false},
		{"outlook.com", "user@outlook.com", false},
		{"yahoo.com", "user@yahoo.com", false},
		{"protonmail.com", "user@protonmail.com", false},

		// Edge cases
		{"empty string", "", false},
		{"no domain", "user@", false},
		{"uppercase legitimate", "USER@GMAIL.COM", false},
		{"mixed case disposable", "USER@10MINUTEMAIL.COM", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsDisposable(tt.input)
			if result != tt.expected {
				t.Errorf("IsDisposable(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestAddDomains(t *testing.T) {
	// Add a custom domain
	AddDomains("custom-test-domain.com")

	// It should now be detected as disposable
	if !IsDisposable("user@custom-test-domain.com") {
		t.Error("Custom domain should be detected as disposable after AddDomains")
	}
}

func TestAddAllowlist(t *testing.T) {
	// First verify a known disposable domain
	if !IsDisposable("10minutemail.com") {
		t.Skip("10minutemail.com not in blocklist, skipping allowlist test")
	}

	// Add to allowlist
	AddAllowlist("10minutemail.com")

	// It should no longer be detected as disposable
	if IsDisposable("10minutemail.com") {
		t.Error("Allowlisted domain should not be detected as disposable")
	}
}

func TestStats(t *testing.T) {
	stats := Stats()

	if stats.BlocklistCount == 0 {
		t.Error("BlocklistCount should be greater than 0")
	}

	t.Logf("Blocklist count: %d", stats.BlocklistCount)
	t.Logf("Allowlist count: %d", stats.AllowlistCount)
	t.Logf("Version: %s", stats.Version)
	t.Logf("Mode: %s", stats.Mode)
}

func TestGetBlocklist(t *testing.T) {
	blocklist := GetBlocklist()

	if len(blocklist) == 0 {
		t.Error("Blocklist should not be empty")
	}

	t.Logf("Blocklist has %d domains", len(blocklist))
}

func TestGetAllowlist(t *testing.T) {
	allowlist := GetAllowlist()
	t.Logf("Allowlist has %d domains", len(allowlist))
}

func BenchmarkIsDisposable(b *testing.B) {
	domains := []string{
		"user@gmail.com",
		"user@tempmail.com",
		"user@10minutemail.com",
		"user@company.co.uk",
		"user@outlook.com",
		"user@yopmail.com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsDisposable(domains[i%len(domains)])
	}
}

func BenchmarkIsDisposable_Parallel(b *testing.B) {
	domains := []string{
		"user@gmail.com",
		"user@tempmail.com",
		"user@10minutemail.com",
		"user@company.co.uk",
		"user@outlook.com",
		"user@yopmail.com",
	}

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			IsDisposable(domains[i%len(domains)])
			i++
		}
	})
}

func TestCheckEmail(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"disposable domain", "user@guerrillamail.com", true},
		{"legitimate domain", "user@gmail.com", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CheckEmail(tt.input)
			if err != nil {
				t.Fatalf("CheckEmail(%q) returned error: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("CheckEmail(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCheckEmailWithContext(t *testing.T) {
	ctx := context.Background()

	// Test with disposable domain
	result, err := CheckEmailWithContext(ctx, "user@mailinator.com")
	if err != nil {
		t.Fatalf("CheckEmailWithContext returned error: %v", err)
	}
	if !result {
		t.Error("Expected mailinator.com to be disposable")
	}

	// Test with legitimate domain
	result, err = CheckEmailWithContext(ctx, "user@gmail.com")
	if err != nil {
		t.Fatalf("CheckEmailWithContext returned error: %v", err)
	}
	if result {
		t.Error("Expected gmail.com to not be disposable")
	}

	// Test with cancelled context (should still work since data is cached)
	cancelledCtx, cancel := context.WithCancel(ctx)
	cancel()
	result, err = CheckEmailWithContext(cancelledCtx, "user@gmail.com")
	if err != nil {
		t.Fatalf("CheckEmailWithContext with cancelled context returned error: %v", err)
	}
	if result {
		t.Error("Expected gmail.com to not be disposable even with cancelled context")
	}
}

func TestIsReady(t *testing.T) {
	// Since the checker is initialized by other tests, it should be ready
	if !IsReady() {
		t.Error("Expected IsReady() to return true after initialization")
	}
}

func TestInitError(t *testing.T) {
	// Since the checker is initialized successfully by other tests, there should be no error
	if err := InitError(); err != nil {
		t.Errorf("Expected InitError() to return nil, got: %v", err)
	}
}
