package trie

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"io"
	"time"
)

// DataFile represents the serialized data format.
type DataFile struct {
	Version     string    // Version identifier
	CreatedAt   time.Time // When the data was generated
	DomainCount int       // Number of domains
	Blocklist   []string  // List of blocked domains (stored as list for smaller size)
	Allowlist   []string  // List of allowed domains
}

// Serialize serializes the blocklist and allowlist tries to a compressed binary format.
func Serialize(blocklist, allowlist *Trie) ([]byte, error) {
	data := DataFile{
		Version:     "1.0",
		CreatedAt:   time.Now().UTC(),
		DomainCount: blocklist.Size(),
		Blocklist:   blocklist.GetAll(),
		Allowlist:   allowlist.GetAll(),
	}

	// Encode to gob
	var gobBuf bytes.Buffer
	encoder := gob.NewEncoder(&gobBuf)
	if err := encoder.Encode(data); err != nil {
		return nil, fmt.Errorf("gob encode failed: %w", err)
	}

	// Compress with gzip
	var gzipBuf bytes.Buffer
	gzipWriter, err := gzip.NewWriterLevel(&gzipBuf, gzip.BestCompression)
	if err != nil {
		return nil, fmt.Errorf("gzip writer creation failed: %w", err)
	}

	if _, err := gzipWriter.Write(gobBuf.Bytes()); err != nil {
		return nil, fmt.Errorf("gzip write failed: %w", err)
	}

	if err := gzipWriter.Close(); err != nil {
		return nil, fmt.Errorf("gzip close failed: %w", err)
	}

	return gzipBuf.Bytes(), nil
}

// Deserialize deserializes compressed binary data into blocklist and allowlist tries.
func Deserialize(data []byte) (*Trie, *Trie, *DataFile, error) {
	// Decompress with gzip
	gzipReader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("gzip reader creation failed: %w", err)
	}
	defer gzipReader.Close()

	decompressed, err := io.ReadAll(gzipReader)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("gzip read failed: %w", err)
	}

	// Decode from gob
	var dataFile DataFile
	decoder := gob.NewDecoder(bytes.NewReader(decompressed))
	if err := decoder.Decode(&dataFile); err != nil {
		return nil, nil, nil, fmt.Errorf("gob decode failed: %w", err)
	}

	// Build tries from domain lists
	blocklist := New()
	for _, domain := range dataFile.Blocklist {
		blocklist.Insert(domain)
	}

	allowlist := New()
	for _, domain := range dataFile.Allowlist {
		allowlist.Insert(domain)
	}

	return blocklist, allowlist, &dataFile, nil
}

// SerializeToWriter serializes the tries and writes to an io.Writer.
func SerializeToWriter(blocklist, allowlist *Trie, w io.Writer) error {
	data, err := Serialize(blocklist, allowlist)
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	return err
}

// DeserializeFromReader reads from an io.Reader and deserializes the tries.
func DeserializeFromReader(r io.Reader) (*Trie, *Trie, *DataFile, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("read failed: %w", err)
	}

	return Deserialize(data)
}
