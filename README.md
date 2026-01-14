# go-is-disposable-email

[![Go Reference](https://pkg.go.dev/badge/github.com/rezmoss/go-is-disposable-email.svg)](https://pkg.go.dev/github.com/rezmoss/go-is-disposable-email)
[![CI](https://github.com/rezmoss/go-is-disposable-email/actions/workflows/ci.yml/badge.svg)](https://github.com/rezmoss/go-is-disposable-email/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/rezmoss/go-is-disposable-email)](https://goreportcard.com/report/github.com/rezmoss/go-is-disposable-email)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A high-performance Go package for detecting disposable/temporary email addresses. Uses a trie data structure for efficient lookups and supports hierarchical domain matching.

## Features

- **High Performance**: Trie-based data structure for O(m) lookups where m = domain length
- **Offline First**: Works without network access using embedded data (~370KB compressed)
- **57,000+ Domains**: Merged from multiple trusted sources, updated daily
- **Hierarchical Matching**: Detects subdomains of known disposable domains (e.g., `mail.tempmail.com`)
- **Runtime Extensible**: Add custom domains to blocklist/allowlist at runtime
- **Zero Dependencies**: Uses only Go standard library
- **Thread-Safe**: Safe for concurrent use

## Installation

```bash
go get github.com/rezmoss/go-is-disposable-email
```

## Quick Start

```go
package main

import (
    "fmt"
    disposable "github.com/rezmoss/go-is-disposable-email"
)

func main() {
    // Check if an email is from a disposable domain
    if disposable.IsDisposable("user@tempmail.com") {
        fmt.Println("This is a disposable email!")
    }

    // Works with just domains too
    if disposable.IsDisposable("guerrillamail.com") {
        fmt.Println("This domain is disposable!")
    }

    // Legitimate emails return false
    if !disposable.IsDisposable("user@gmail.com") {
        fmt.Println("This is a legitimate email domain")
    }
}
```

## API

### Package-Level Functions

```go
// Check if email/domain is disposable
disposable.IsDisposable("user@tempmail.com") // true
disposable.IsDisposable("user@gmail.com")    // false

// Add custom domains at runtime
disposable.AddDomains("custom-disposable.com")
disposable.AddAllowlist("legitimate-domain.com")

// Get statistics
stats := disposable.Stats()
fmt.Printf("Blocklist: %d domains\n", stats.BlocklistCount)

// Get all domains (useful for debugging)
blocklist := disposable.GetBlocklist()
allowlist := disposable.GetAllowlist()
```

### Custom Checker with Options

```go
import (
    "time"
    disposable "github.com/rezmoss/go-is-disposable-email"
)

// Create a checker with custom configuration
checker, err := disposable.New(
    disposable.WithOffline(),                     // Use only embedded data
    disposable.WithCustomBlocklist("blocked.com"), // Add custom blocked domains
    disposable.WithCustomAllowlist("allowed.com"), // Add custom allowed domains
)
if err != nil {
    log.Fatal(err)
}
defer checker.Close()

// Use the checker
if checker.IsDisposable("user@tempmail.com") {
    // Handle disposable email
}
```

### Available Options

| Option | Description |
|--------|-------------|
| `WithOffline()` | Use only embedded data, no network calls |
| `WithMode(mode)` | Set operating mode (Online, Offline, Hybrid) |
| `WithAutoRefresh(interval)` | Enable automatic data updates |
| `WithCacheDir(dir)` | Set cache directory for downloaded data |
| `WithHTTPTimeout(timeout)` | Set HTTP timeout for downloads |
| `WithCustomBlocklist(domains...)` | Add domains to block |
| `WithCustomAllowlist(domains...)` | Add domains to allow |
| `WithLogger(logger)` | Set custom logger |

## Data Sources

This package aggregates disposable email domains from multiple trusted sources, which are automatically scanned and merged daily:

| Source | URL | Domains |
|--------|-----|---------|
| **FGRibreau/mailchecker** | https://github.com/FGRibreau/mailchecker | ~55,000 |
| **disposable-email-domains** | https://github.com/disposable-email-domains/disposable-email-domains | ~5,000 |
| **Manual additions** | This repository's `data/manual.txt` | Community contributed |

The data is updated automatically every day at 2 AM UTC via GitHub Actions.

**Total unique domains: 57,000+**

## How It Works

1. **Trie Data Structure**: Domains are stored in a trie (prefix tree) with domains reversed for efficient suffix matching
2. **Hierarchical Matching**: When checking `mail.tempmail.com`, the package also checks `tempmail.com`
3. **Allowlist Priority**: Allowlisted domains take precedence over blocklist
4. **Compressed Storage**: Data is serialized with gob and compressed with gzip (~370KB)

## Contributing

### Adding New Disposable Domains

1. Open an issue with the domain(s) you want to add
2. Or submit a PR adding domains to `data/manual.txt` (one domain per line)

### Development

```bash
# Clone the repository
git clone https://github.com/rezmoss/go-is-disposable-email.git
cd go-is-disposable-email

# Run tests
go test -v ./...

# Run benchmarks
go test -bench=. ./...

# Update data.bin from sources
go run ./cmd/disposable-update -o ./data -v
```

## Benchmarks

```
goos: darwin
goarch: arm64
BenchmarkIsDisposable-10              2948937    408.8 ns/op    0 B/op    0 allocs/op
BenchmarkIsDisposable_Parallel-10     5441361    219.8 ns/op    0 B/op    0 allocs/op
```

Zero allocations per lookup for maximum performance.

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

- [FGRibreau/mailchecker](https://github.com/FGRibreau/mailchecker) - Primary source of disposable domains
- [disposable-email-domains](https://github.com/disposable-email-domains/disposable-email-domains) - Additional domains and allowlist
