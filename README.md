# go-is-disposable-email

[![Go Reference](https://pkg.go.dev/badge/github.com/rezmoss/go-is-disposable-email.svg)](https://pkg.go.dev/github.com/rezmoss/go-is-disposable-email)
[![CI](https://github.com/rezmoss/go-is-disposable-email/actions/workflows/ci.yml/badge.svg)](https://github.com/rezmoss/go-is-disposable-email/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/rezmoss/go-is-disposable-email)](https://goreportcard.com/report/github.com/rezmoss/go-is-disposable-email)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A high-performance Go package for detecting disposable/temporary email addresses. Uses a trie data structure for efficient lookups and supports hierarchical domain matching.

## Features

- **High Performance**: Trie-based data structure for O(m) lookups where m = domain length
- **Auto-Download**: Downloads data on first use, caches locally (~450KB compressed)
- **72,000+ Domains**: Merged from multiple trusted sources, updated daily
- **Hierarchical Matching**: Detects subdomains of known disposable domains (e.g., `mail.tempmail.com`)
- **Runtime Extensible**: Add custom domains to blocklist/allowlist at runtime
- **Zero Dependencies**: Uses only Go standard library
- **Thread-Safe**: Safe for concurrent use with race-tested code
- **Error Handling**: Typed errors for programmatic error handling (`DownloadError`, `CacheError`, etc.)

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
    disposable.WithAutoRefresh(24 * time.Hour),    // Enable auto-refresh
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
| `WithAutoRefresh(interval)` | Enable automatic background data updates (requires `Close()`) |
| `WithCacheDir(dir)` | Set cache directory for downloaded data |
| `WithHTTPTimeout(timeout)` | Set HTTP timeout for downloads |
| `WithCustomBlocklist(domains...)` | Add domains to block |
| `WithCustomAllowlist(domains...)` | Add domains to allow |
| `WithDataURL(url)` | Set custom URL for data.bin downloads |
| `WithLogger(logger)` | Set custom logger |

### Error Handling

For production systems, use the error-returning variants to distinguish between "not disposable" and "initialization failed":

```go
// Check with error handling
isDisposable, err := disposable.CheckEmail("user@tempmail.com")
if err != nil {
    // Handle error (network failure, cache issue, etc.)
    log.Printf("Check failed: %v", err)
}

// Check initialization status
if !disposable.IsReady() {
    err := disposable.InitError()
    log.Printf("Initialization failed: %v", err)
}

// Programmatic error type checking
if disposable.IsDownloadError(err) {
    // Handle download-specific error
}
if disposable.IsCacheError(err) {
    // Handle cache-specific error
}
```

## Data Sources

We gather and compile disposable email domains from various trusted sources on a daily basis. The database is automatically updated every day at 2 AM UTC via GitHub Actions.

**Total unique domains: 72,000+**

To add custom domains, edit `data/manual.txt` (one domain per line).

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

- [FGRibreau/mailchecker](https://github.com/FGRibreau/mailchecker)
- [disposable-email-domains](https://github.com/disposable-email-domains/disposable-email-domains)
- [disposable/disposable-email-domains](https://github.com/disposable/disposable-email-domains)
- [7c/fakefilter](https://github.com/7c/fakefilter)
