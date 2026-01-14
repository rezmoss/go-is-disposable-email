package disposable

import (
	"reflect"
	"testing"
)

func TestExtractDomain(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"user@example.com", "example.com"},
		{"test@subdomain.example.com", "subdomain.example.com"},
		{"example.com", "example.com"},
		{"EXAMPLE.COM", "example.com"},
		{"USER@EXAMPLE.COM", "example.com"},
		{"  user@example.com  ", "example.com"},
		{"", ""},
		{"user@", ""},
		{"@example.com", "example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ExtractDomain(tt.input)
			if result != tt.expected {
				t.Errorf("ExtractDomain(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetDomainHierarchy(t *testing.T) {
	tests := []struct {
		domain   string
		expected []string
	}{
		{"mail.tempmail.com", []string{"mail.tempmail.com", "tempmail.com"}},
		{"sub.mail.tempmail.com", []string{"sub.mail.tempmail.com", "mail.tempmail.com", "tempmail.com"}},
		{"example.com", []string{"example.com"}},
		{"com", nil},
		{"", nil},
	}

	for _, tt := range tests {
		t.Run(tt.domain, func(t *testing.T) {
			result := GetDomainHierarchy(tt.domain)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("GetDomainHierarchy(%q) = %v, want %v", tt.domain, result, tt.expected)
			}
		})
	}
}

func TestIsValidDomain(t *testing.T) {
	tests := []struct {
		domain   string
		expected bool
	}{
		{"example.com", true},
		{"sub.example.com", true},
		{"test-domain.com", true},
		{"test_domain.com", true},
		{"123.com", true},
		{"example", false},
		{"", false},
		{".com", false},
		{"example.", false},
		{"exam ple.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.domain, func(t *testing.T) {
			result := IsValidDomain(tt.domain)
			if result != tt.expected {
				t.Errorf("IsValidDomain(%q) = %v, want %v", tt.domain, result, tt.expected)
			}
		})
	}
}

func TestNormalizeDomain(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"EXAMPLE.COM", "example.com"},
		{"  example.com  ", "example.com"},
		{"Example.Com", "example.com"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := NormalizeDomain(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeDomain(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
