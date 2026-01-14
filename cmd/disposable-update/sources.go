package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Source represents a data source for disposable email domains.
type Source struct {
	Name string
	URL  string
	Type SourceType
}

// SourceType indicates whether a source is a blocklist or allowlist.
type SourceType int

const (
	SourceTypeBlocklist SourceType = iota
	SourceTypeAllowlist
)

// LoadSourcesFromFile reads data sources from a text file.
// Format: type|name|url
// Lines starting with # are comments, empty lines are ignored.
func LoadSourcesFromFile(path string) ([]Source, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open sources file: %w", err)
	}
	defer f.Close()

	var sources []Source
	scanner := bufio.NewScanner(f)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "|", 3)
		if len(parts) != 3 {
			return nil, fmt.Errorf("invalid format at line %d: expected 'type|name|url', got %q", lineNum, line)
		}

		sourceType := strings.TrimSpace(strings.ToLower(parts[0]))
		name := strings.TrimSpace(parts[1])
		url := strings.TrimSpace(parts[2])

		if name == "" || url == "" {
			return nil, fmt.Errorf("invalid source at line %d: name and url cannot be empty", lineNum)
		}

		var stype SourceType
		switch sourceType {
		case "blocklist":
			stype = SourceTypeBlocklist
		case "allowlist":
			stype = SourceTypeAllowlist
		default:
			return nil, fmt.Errorf("invalid source type at line %d: expected 'blocklist' or 'allowlist', got %q", lineNum, sourceType)
		}

		sources = append(sources, Source{
			Name: name,
			URL:  url,
			Type: stype,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading sources file: %w", err)
	}

	if len(sources) == 0 {
		return nil, fmt.Errorf("no sources found in file")
	}

	return sources, nil
}
