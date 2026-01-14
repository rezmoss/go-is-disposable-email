package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rezmoss/go-is-disposable-email/internal/trie"
)

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
		result := normalizeDomain(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeDomain(%q) = %q, want %q", tt.input, result, tt.expected)
		}
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
		{"a.b.c.d.com", true},
		{"example", false},         // No TLD
		{"", false},                // Empty
		{".com", false},            // Starts with dot
		{"example.", false},        // Ends with dot
		{"exam ple.com", false},    // Contains space
		{"example..com", false},    // Double dot
	}

	for _, tt := range tests {
		t.Run(tt.domain, func(t *testing.T) {
			result := isValidDomain(tt.domain)
			if result != tt.expected {
				t.Errorf("isValidDomain(%q) = %v, want %v", tt.domain, result, tt.expected)
			}
		})
	}
}

func TestIsValidDomainChar(t *testing.T) {
	validChars := "abcdefghijklmnopqrstuvwxyz0123456789-_"
	for _, c := range validChars {
		if !isValidDomainChar(c) {
			t.Errorf("isValidDomainChar(%q) = false, want true", c)
		}
	}

	invalidChars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ!@#$%^&*()+=[]{}|\\:;\"'<>,?/"
	for _, c := range invalidChars {
		if isValidDomainChar(c) {
			t.Errorf("isValidDomainChar(%q) = true, want false", c)
		}
	}
}

func TestParseLines(t *testing.T) {
	input := `# Comment line
domain1.com
domain2.com
  domain3.com
# Another comment

domain4.com
`
	reader := strings.NewReader(input)
	lines, err := parseLines(reader)
	if err != nil {
		t.Fatalf("parseLines error: %v", err)
	}

	expected := []string{"domain1.com", "domain2.com", "domain3.com", "domain4.com"}
	if len(lines) != len(expected) {
		t.Fatalf("Expected %d lines, got %d", len(expected), len(lines))
	}

	for i, line := range lines {
		if line != expected[i] {
			t.Errorf("Line %d: got %q, want %q", i, line, expected[i])
		}
	}
}

func TestLoadSourcesFromFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sources-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a valid sources file
	sourcesPath := filepath.Join(tmpDir, "sources.txt")
	sourcesContent := `# Test sources
blocklist|Test Blocklist|https://example.com/blocklist.txt
allowlist|Test Allowlist|https://example.com/allowlist.txt
`
	if err := os.WriteFile(sourcesPath, []byte(sourcesContent), 0644); err != nil {
		t.Fatalf("Failed to write sources.txt: %v", err)
	}

	sources, err := LoadSourcesFromFile(sourcesPath)
	if err != nil {
		t.Fatalf("LoadSourcesFromFile error: %v", err)
	}

	if len(sources) != 2 {
		t.Fatalf("Expected 2 sources, got %d", len(sources))
	}

	if sources[0].Name != "Test Blocklist" {
		t.Errorf("Expected name 'Test Blocklist', got %q", sources[0].Name)
	}
	if sources[0].Type != SourceTypeBlocklist {
		t.Errorf("Expected type blocklist, got %v", sources[0].Type)
	}

	if sources[1].Name != "Test Allowlist" {
		t.Errorf("Expected name 'Test Allowlist', got %q", sources[1].Name)
	}
	if sources[1].Type != SourceTypeAllowlist {
		t.Errorf("Expected type allowlist, got %v", sources[1].Type)
	}
}

func TestLoadSourcesFromFileInvalid(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sources-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name    string
		content string
	}{
		{"invalid format", "invalid line without pipes"},
		{"invalid type", "invalid|Test|https://example.com"},
		{"empty name", "blocklist||https://example.com"},
		{"empty url", "blocklist|Test|"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourcesPath := filepath.Join(tmpDir, "sources.txt")
			if err := os.WriteFile(sourcesPath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write sources.txt: %v", err)
			}

			_, err := LoadSourcesFromFile(sourcesPath)
			if err == nil {
				t.Error("Expected error for invalid sources file")
			}
		})
	}
}

func TestUpdateStatsSummary(t *testing.T) {
	tests := []struct {
		name     string
		stats    UpdateStats
		expected string
	}{
		{
			name: "domains added",
			stats: UpdateStats{
				OldBlocklistCount: 100,
				NewBlocklistCount: 110,
			},
			expected: "10 domains added to blocklist",
		},
		{
			name: "domains removed",
			stats: UpdateStats{
				OldBlocklistCount: 100,
				NewBlocklistCount: 95,
			},
			expected: "5 domains removed from blocklist",
		},
		{
			name: "no changes",
			stats: UpdateStats{
				OldBlocklistCount: 100,
				NewBlocklistCount: 100,
			},
			expected: "no changes",
		},
		{
			name: "mixed changes",
			stats: UpdateStats{
				OldBlocklistCount: 100,
				NewBlocklistCount: 110,
				OldAllowlistCount: 50,
				NewAllowlistCount: 48,
			},
			expected: "10 domains added to blocklist, 2 domains removed from allowlist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.stats.Summary()
			if result != tt.expected {
				t.Errorf("Summary() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestRunWithSourcesFile(t *testing.T) {
	// Create a temporary directory for test
	tmpDir, err := os.MkdirTemp("", "disposable-update-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create sources.txt pointing to real sources
	sourcesPath := filepath.Join(tmpDir, "sources.txt")
	sourcesContent := `# Test sources - using real URLs
blocklist|FGRibreau/mailchecker|https://raw.githubusercontent.com/FGRibreau/mailchecker/master/list.txt
blocklist|disposable-email-domains|https://raw.githubusercontent.com/disposable-email-domains/disposable-email-domains/main/disposable_email_blocklist.conf
allowlist|disposable-email-domains|https://raw.githubusercontent.com/disposable-email-domains/disposable-email-domains/main/allowlist.conf
`
	if err := os.WriteFile(sourcesPath, []byte(sourcesContent), 0644); err != nil {
		t.Fatalf("Failed to write sources.txt: %v", err)
	}

	// Create output directory
	outputDir := filepath.Join(tmpDir, "output")

	// Run the update
	err = run(outputDir, sourcesPath, "", true, 60*time.Second, "")
	if err != nil {
		t.Fatalf("run() error: %v", err)
	}

	// Verify output file exists
	dataPath := filepath.Join(outputDir, "data.bin")
	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		t.Error("Expected data.bin to be created")
	}
}

func TestGeneratedDataBinIsReadable(t *testing.T) {
	// This test verifies that data/data.bin (if it exists) is readable
	dataPath := filepath.Join("..", "..", "data", "data.bin")
	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		t.Skip("data/data.bin not found")
	}

	data, err := os.ReadFile(dataPath)
	if err != nil {
		t.Fatalf("Failed to read data.bin: %v", err)
	}

	blocklist, allowlist, dataFile, err := trie.Deserialize(data)
	if err != nil {
		t.Fatalf("Failed to deserialize data.bin: %v", err)
	}

	// Verify data is valid
	if blocklist.Size() == 0 {
		t.Error("Blocklist should not be empty")
	}

	if dataFile.Version == "" {
		t.Error("Version should not be empty")
	}

	t.Logf("data.bin stats: blocklist=%d, allowlist=%d, version=%s",
		blocklist.Size(), allowlist.Size(), dataFile.Version)

	// Verify known disposable domains are present
	knownDisposable := []string{
		"10minutemail.com",
		"guerrillamail.com",
		"mailinator.com",
	}

	for _, domain := range knownDisposable {
		if !blocklist.Contains(domain) {
			t.Errorf("Expected %s to be in blocklist", domain)
		}
	}

	// Verify legitimate domains are NOT present
	legitimate := []string{
		"gmail.com",
		"outlook.com",
		"yahoo.com",
	}

	for _, domain := range legitimate {
		if blocklist.Contains(domain) {
			t.Errorf("Expected %s to NOT be in blocklist", domain)
		}
	}
}

func TestWriteTextList(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "disposable-textlist-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	domains := map[string]struct{}{
		"alpha.com": {},
		"beta.com":  {},
		"gamma.com": {},
	}

	outPath := filepath.Join(tmpDir, "test.txt")
	if err := writeTextList(outPath, domains); err != nil {
		t.Fatalf("writeTextList error: %v", err)
	}

	// Read and verify
	content, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	contentStr := string(content)

	// Should have header comments
	if !strings.Contains(contentStr, "# Generated by disposable-update") {
		t.Error("Expected header comment")
	}

	// Should have all domains
	for domain := range domains {
		if !strings.Contains(contentStr, domain) {
			t.Errorf("Expected %s in output", domain)
		}
	}
}

func TestSourcesFileLocation(t *testing.T) {
	// Verify that data/sources.txt exists
	sourcesPath := filepath.Join("..", "..", "data", "sources.txt")
	if _, err := os.Stat(sourcesPath); os.IsNotExist(err) {
		t.Error("Expected data/sources.txt to exist")
	}

	// Verify it can be parsed
	sources, err := LoadSourcesFromFile(sourcesPath)
	if err != nil {
		t.Fatalf("Failed to load sources.txt: %v", err)
	}

	if len(sources) == 0 {
		t.Error("Expected at least one source in sources.txt")
	}

	t.Logf("Loaded %d sources from sources.txt", len(sources))
	for _, src := range sources {
		t.Logf("  - %s (%v)", src.Name, src.Type)
	}
}
