package disposable

import (
	"strings"
)

// ExtractDomain extracts the domain from an email address or returns the input
// if it's already a domain. Returns empty string for invalid input.
func ExtractDomain(emailOrDomain string) string {
	emailOrDomain = strings.TrimSpace(emailOrDomain)
	emailOrDomain = strings.ToLower(emailOrDomain)

	if emailOrDomain == "" {
		return ""
	}

	// Check if it's an email address
	if idx := strings.LastIndex(emailOrDomain, "@"); idx != -1 {
		domain := emailOrDomain[idx+1:]
		if domain == "" {
			return ""
		}
		return domain
	}

	// Assume it's already a domain
	return emailOrDomain
}

// GetDomainHierarchy returns all domain levels to check.
// For "mail.tempmail.com", it returns ["mail.tempmail.com", "tempmail.com", "com"]
// We skip single-part TLDs (like "com") as they're not useful for checking.
func GetDomainHierarchy(domain string) []string {
	if domain == "" {
		return nil
	}

	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return nil
	}

	var hierarchy []string
	for i := 0; i < len(parts)-1; i++ {
		subdomain := strings.Join(parts[i:], ".")
		if subdomain != "" {
			hierarchy = append(hierarchy, subdomain)
		}
	}

	return hierarchy
}

// IsValidDomain performs basic domain validation.
func IsValidDomain(domain string) bool {
	if domain == "" {
		return false
	}

	// Must have at least one dot
	if !strings.Contains(domain, ".") {
		return false
	}

	// Check each part
	parts := strings.Split(domain, ".")
	for _, part := range parts {
		if part == "" {
			return false
		}
		// Check for invalid characters (basic check)
		for _, c := range part {
			if !isValidDomainChar(c) {
				return false
			}
		}
	}

	return true
}

// isValidDomainChar checks if a character is valid in a domain name.
func isValidDomainChar(c rune) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= '0' && c <= '9') ||
		c == '-' ||
		c == '_'
}

// NormalizeDomain normalizes a domain for consistent storage and lookup.
func NormalizeDomain(domain string) string {
	return strings.ToLower(strings.TrimSpace(domain))
}
