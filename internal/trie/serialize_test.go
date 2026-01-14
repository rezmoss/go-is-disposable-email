package trie

import (
	"bytes"
	"testing"
)

func TestSerializeDeserialize(t *testing.T) {
	// Create test tries
	blocklist := New()
	blocklist.Insert("tempmail.com")
	blocklist.Insert("guerrillamail.com")
	blocklist.Insert("10minutemail.com")

	allowlist := New()
	allowlist.Insert("gmail.com")
	allowlist.Insert("outlook.com")

	// Serialize
	data, err := Serialize(blocklist, allowlist)
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("Serialized data is empty")
	}

	t.Logf("Serialized size: %d bytes", len(data))

	// Deserialize
	restoredBlocklist, restoredAllowlist, dataFile, err := Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	// Verify blocklist
	if restoredBlocklist.Size() != blocklist.Size() {
		t.Errorf("Blocklist size mismatch: got %d, want %d", restoredBlocklist.Size(), blocklist.Size())
	}

	if !restoredBlocklist.Contains("tempmail.com") {
		t.Error("Restored blocklist should contain tempmail.com")
	}
	if !restoredBlocklist.Contains("guerrillamail.com") {
		t.Error("Restored blocklist should contain guerrillamail.com")
	}

	// Verify allowlist
	if restoredAllowlist.Size() != allowlist.Size() {
		t.Errorf("Allowlist size mismatch: got %d, want %d", restoredAllowlist.Size(), allowlist.Size())
	}

	if !restoredAllowlist.Contains("gmail.com") {
		t.Error("Restored allowlist should contain gmail.com")
	}

	// Verify metadata
	if dataFile.Version != "1.0" {
		t.Errorf("Version mismatch: got %s, want 1.0", dataFile.Version)
	}
	if dataFile.DomainCount != 3 {
		t.Errorf("DomainCount mismatch: got %d, want 3", dataFile.DomainCount)
	}
}

func TestSerializeToWriter(t *testing.T) {
	blocklist := New()
	blocklist.Insert("test.com")

	allowlist := New()

	var buf bytes.Buffer
	err := SerializeToWriter(blocklist, allowlist, &buf)
	if err != nil {
		t.Fatalf("SerializeToWriter failed: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("Buffer should not be empty")
	}
}

func TestDeserializeFromReader(t *testing.T) {
	blocklist := New()
	blocklist.Insert("test.com")

	allowlist := New()

	data, err := Serialize(blocklist, allowlist)
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	restoredBlocklist, restoredAllowlist, _, err := DeserializeFromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("DeserializeFromReader failed: %v", err)
	}

	if !restoredBlocklist.Contains("test.com") {
		t.Error("Restored blocklist should contain test.com")
	}

	if restoredAllowlist.Size() != 0 {
		t.Error("Allowlist should be empty")
	}
}

func TestDeserializeInvalidData(t *testing.T) {
	// Test with invalid data
	invalidData := []byte("this is not valid gzip data")
	_, _, _, err := Deserialize(invalidData)
	if err == nil {
		t.Error("Expected error for invalid data")
	}
}

func BenchmarkSerialize(b *testing.B) {
	blocklist := New()
	for i := 0; i < 1000; i++ {
		blocklist.Insert("domain" + string(rune(i)) + ".com")
	}

	allowlist := New()
	for i := 0; i < 100; i++ {
		allowlist.Insert("allowed" + string(rune(i)) + ".com")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Serialize(blocklist, allowlist)
	}
}

func BenchmarkDeserialize(b *testing.B) {
	blocklist := New()
	for i := 0; i < 1000; i++ {
		blocklist.Insert("domain" + string(rune(i)) + ".com")
	}

	allowlist := New()

	data, _ := Serialize(blocklist, allowlist)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _, _ = Deserialize(data)
	}
}
