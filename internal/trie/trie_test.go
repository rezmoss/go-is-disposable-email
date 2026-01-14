package trie

import (
	"sort"
	"testing"
)

func TestTrieInsertAndContains(t *testing.T) {
	tr := New()

	// Insert some domains
	tr.Insert("tempmail.com")
	tr.Insert("example.com")
	tr.Insert("test.org")

	// Test contains
	if !tr.Contains("tempmail.com") {
		t.Error("Expected trie to contain tempmail.com")
	}
	if !tr.Contains("example.com") {
		t.Error("Expected trie to contain example.com")
	}
	if !tr.Contains("test.org") {
		t.Error("Expected trie to contain test.org")
	}

	// Test non-existent
	if tr.Contains("notexist.com") {
		t.Error("Expected trie to not contain notexist.com")
	}
	if tr.Contains("") {
		t.Error("Expected trie to not contain empty string")
	}
}

func TestTrieContainsHierarchical(t *testing.T) {
	tr := New()

	// Insert parent domain
	tr.Insert("tempmail.com")

	// Test hierarchical matching
	if !tr.ContainsHierarchical("tempmail.com") {
		t.Error("Expected trie to match tempmail.com")
	}
	if !tr.ContainsHierarchical("mail.tempmail.com") {
		t.Error("Expected trie to match subdomain mail.tempmail.com")
	}
	if !tr.ContainsHierarchical("sub.mail.tempmail.com") {
		t.Error("Expected trie to match deep subdomain sub.mail.tempmail.com")
	}

	// Test non-matching
	if tr.ContainsHierarchical("notexist.com") {
		t.Error("Expected trie to not match notexist.com")
	}
	if tr.ContainsHierarchical("tempmail.org") {
		t.Error("Expected trie to not match tempmail.org")
	}
}

func TestTrieSize(t *testing.T) {
	tr := New()

	if tr.Size() != 0 {
		t.Errorf("Expected size 0, got %d", tr.Size())
	}

	tr.Insert("one.com")
	if tr.Size() != 1 {
		t.Errorf("Expected size 1, got %d", tr.Size())
	}

	tr.Insert("two.com")
	if tr.Size() != 2 {
		t.Errorf("Expected size 2, got %d", tr.Size())
	}

	// Duplicate should not increase size
	tr.Insert("one.com")
	if tr.Size() != 2 {
		t.Errorf("Expected size 2 after duplicate, got %d", tr.Size())
	}
}

func TestTrieGetAll(t *testing.T) {
	tr := New()

	domains := []string{"alpha.com", "beta.com", "gamma.com"}
	for _, d := range domains {
		tr.Insert(d)
	}

	result := tr.GetAll()
	sort.Strings(result)
	sort.Strings(domains)

	if len(result) != len(domains) {
		t.Errorf("Expected %d domains, got %d", len(domains), len(result))
	}

	for i, d := range domains {
		if result[i] != d {
			t.Errorf("Expected %s, got %s", d, result[i])
		}
	}
}

func TestTrieClear(t *testing.T) {
	tr := New()

	tr.Insert("test.com")
	tr.Insert("example.com")

	if tr.Size() != 2 {
		t.Errorf("Expected size 2, got %d", tr.Size())
	}

	tr.Clear()

	if tr.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", tr.Size())
	}

	if tr.Contains("test.com") {
		t.Error("Expected trie to not contain test.com after clear")
	}
}

func TestTrieEmptyDomain(t *testing.T) {
	tr := New()

	tr.Insert("")
	if tr.Size() != 0 {
		t.Error("Empty domain should not be inserted")
	}

	if tr.Contains("") {
		t.Error("Contains should return false for empty string")
	}

	if tr.ContainsHierarchical("") {
		t.Error("ContainsHierarchical should return false for empty string")
	}
}

func BenchmarkTrieInsert(b *testing.B) {
	domains := []string{
		"tempmail.com",
		"guerrillamail.com",
		"10minutemail.com",
		"example.com",
		"test.org",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr := New()
		for _, d := range domains {
			tr.Insert(d)
		}
	}
}

func BenchmarkTrieContains(b *testing.B) {
	tr := New()
	domains := []string{
		"tempmail.com",
		"guerrillamail.com",
		"10minutemail.com",
		"example.com",
		"test.org",
	}
	for _, d := range domains {
		tr.Insert(d)
	}

	testDomains := []string{
		"tempmail.com",
		"notexist.com",
		"guerrillamail.com",
		"another.com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.Contains(testDomains[i%len(testDomains)])
	}
}

func BenchmarkTrieContainsHierarchical(b *testing.B) {
	tr := New()
	domains := []string{
		"tempmail.com",
		"guerrillamail.com",
		"10minutemail.com",
		"example.com",
		"test.org",
	}
	for _, d := range domains {
		tr.Insert(d)
	}

	testDomains := []string{
		"mail.tempmail.com",
		"sub.mail.tempmail.com",
		"notexist.com",
		"guerrillamail.com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.ContainsHierarchical(testDomains[i%len(testDomains)])
	}
}
