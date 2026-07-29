// disposable-update downloads disposable email domain lists from sources defined
// in data/sources.txt, merges them, and generates a compressed binary data file.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/rezmoss/go-is-disposable-email/internal/trie"
)

// UpdateStats tracks changes between updates
type UpdateStats struct {
	OldBlocklistCount int
	NewBlocklistCount int
	OldAllowlistCount int
	NewAllowlistCount int
	FailedSources     []string

	// AddedBlocklist and RemovedBlocklist are the exact domains that changed in
	// the blocklist this run, sorted. Populated when previous data is available.
	AddedBlocklist   []string
	RemovedBlocklist []string
}

// changelogPreviewLimit caps how many domains are listed inline in the release
// notes; the full lists are written to added.txt / removed.txt.
const changelogPreviewLimit = 50

// Summary returns a short summary of changes
func (s *UpdateStats) Summary() string {
	blocklistDiff := s.NewBlocklistCount - s.OldBlocklistCount
	allowlistDiff := s.NewAllowlistCount - s.OldAllowlistCount

	var parts []string

	if blocklistDiff > 0 {
		parts = append(parts, fmt.Sprintf("%d domains added to blocklist", blocklistDiff))
	} else if blocklistDiff < 0 {
		parts = append(parts, fmt.Sprintf("%d domains removed from blocklist", -blocklistDiff))
	}

	if allowlistDiff > 0 {
		parts = append(parts, fmt.Sprintf("%d domains added to allowlist", allowlistDiff))
	} else if allowlistDiff < 0 {
		parts = append(parts, fmt.Sprintf("%d domains removed from allowlist", -allowlistDiff))
	}

	if len(parts) == 0 {
		return "no changes"
	}

	return strings.Join(parts, ", ")
}

// ReleaseNotes renders Markdown release notes describing the exact domains that
// changed this run. The inline lists are capped to changelogPreviewLimit; the
// full lists live in the added.txt / removed.txt release assets.
func (s *UpdateStats) ReleaseNotes() string {
	var b strings.Builder

	fmt.Fprintf(&b, "## Disposable email data update\n\n")
	fmt.Fprintf(&b, "**Blocklist:** +%d / -%d (total: %d)\n\n",
		len(s.AddedBlocklist), len(s.RemovedBlocklist), s.NewBlocklistCount)

	allowlistDiff := s.NewAllowlistCount - s.OldAllowlistCount
	fmt.Fprintf(&b, "**Allowlist:** %+d (total: %d)\n\n", allowlistDiff, s.NewAllowlistCount)

	writeChangeSection(&b, "Added domains", "+", s.AddedBlocklist, "added.txt")
	writeChangeSection(&b, "Removed domains", "-", s.RemovedBlocklist, "removed.txt")

	if len(s.FailedSources) > 0 {
		fmt.Fprintf(&b, "> ⚠️ Sources unavailable this run: %s\n\n", strings.Join(s.FailedSources, ", "))
	}

	fmt.Fprintf(&b, "### Installation\n\n```bash\ngo get github.com/rezmoss/go-is-disposable-email@latest\n```\n")

	return b.String()
}

// writeChangeSection appends a capped diff block for one set of changed domains.
func writeChangeSection(b *strings.Builder, title, marker string, domains []string, asset string) {
	if len(domains) == 0 {
		return
	}

	limit := len(domains)
	if limit > changelogPreviewLimit {
		limit = changelogPreviewLimit
	}

	fmt.Fprintf(b, "### %s (%d)\n\n```diff\n", title, len(domains))
	for _, d := range domains[:limit] {
		fmt.Fprintf(b, "%s %s\n", marker, d)
	}
	if len(domains) > limit {
		fmt.Fprintf(b, "  … and %d more (see %s)\n", len(domains)-limit, asset)
	}
	fmt.Fprintf(b, "```\n\n")
}

// writeChangelog writes the release artifacts describing this update: the full
// added/removed domain lists plus Markdown release notes.
func (s *UpdateStats) writeChangelog(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	if err := writeDomainList(filepath.Join(dir, "added.txt"), s.AddedBlocklist); err != nil {
		return err
	}
	if err := writeDomainList(filepath.Join(dir, "removed.txt"), s.RemovedBlocklist); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "notes.md"), []byte(s.ReleaseNotes()), 0644)
}

// writeDomainList writes one domain per line, sorted by the caller.
func writeDomainList(path string, domains []string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, d := range domains {
		if _, err := fmt.Fprintln(f, d); err != nil {
			return err
		}
	}
	return nil
}

// diffSorted returns the sorted keys present in a but not in b.
func diffSorted(a, b map[string]struct{}) []string {
	var diff []string
	for domain := range a {
		if _, ok := b[domain]; !ok {
			diff = append(diff, domain)
		}
	}
	sort.Strings(diff)
	return diff
}

func main() {
	outputDir := flag.String("o", "./data", "Output directory for data.bin")
	sourcesFile := flag.String("sources", "", "Path to sources.txt file (default: <output-dir>/sources.txt)")
	manualFile := flag.String("manual", "", "Path to manual additions file")
	verbose := flag.Bool("v", false, "Verbose output")
	timeout := flag.Duration("timeout", 60*time.Second, "HTTP timeout for downloads")
	summaryFile := flag.String("summary", "", "Write update summary to file (for CI)")
	changelogDir := flag.String("changelog", "", "Directory to write release artifacts (added.txt, removed.txt, notes.md)")
	flag.Parse()

	// Default sources file location
	if *sourcesFile == "" {
		*sourcesFile = filepath.Join(*outputDir, "sources.txt")
	}

	if err := run(*outputDir, *sourcesFile, *manualFile, *verbose, *timeout, *summaryFile, *changelogDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(outputDir, sourcesFile, manualFile string, verbose bool, timeout time.Duration, summaryFile, changelogDir string) error {
	log := func(format string, args ...any) {
		if verbose {
			fmt.Printf(format+"\n", args...)
		}
	}
	logError := func(format string, args ...any) {
		fmt.Fprintf(os.Stderr, "ERROR: "+format+"\n", args...)
	}

	// Load existing data to compare changes
	stats := &UpdateStats{}
	var oldBlocklistSet map[string]struct{}
	outputPath := filepath.Join(outputDir, "data.bin")
	if existingData, err := os.ReadFile(outputPath); err == nil {
		if oldBlocklist, oldAllowlist, _, err := trie.Deserialize(existingData); err == nil {
			stats.OldBlocklistCount = oldBlocklist.Size()
			stats.OldAllowlistCount = oldAllowlist.Size()
			oldBlocklistSet = make(map[string]struct{}, stats.OldBlocklistCount)
			for _, d := range oldBlocklist.GetAll() {
				oldBlocklistSet[d] = struct{}{}
			}
			log("Existing data: %d blocklist, %d allowlist domains", stats.OldBlocklistCount, stats.OldAllowlistCount)
		}
	}

	// Load sources from file
	log("Loading sources from %s...", sourcesFile)
	sources, err := LoadSourcesFromFile(sourcesFile)
	if err != nil {
		return fmt.Errorf("failed to load sources: %w", err)
	}
	log("Loaded %d sources", len(sources))

	client := &http.Client{Timeout: timeout}

	blocklist := make(map[string]struct{})
	allowlist := make(map[string]struct{})
	successfulSources := 0

	// Download from all sources
	for _, src := range sources {
		log("Downloading %s...", src.Name)

		domains, err := downloadSource(client, src.URL)
		if err != nil {
			logError("Failed to download %s: %v (skipping)", src.Name, err)
			stats.FailedSources = append(stats.FailedSources, src.Name)
			continue
		}

		// Validate: skip empty sources
		if len(domains) == 0 {
			logError("Source %s returned empty data (skipping)", src.Name)
			stats.FailedSources = append(stats.FailedSources, src.Name)
			continue
		}

		log("  Downloaded %d domains from %s", len(domains), src.Name)
		successfulSources++

		for _, domain := range domains {
			domain = normalizeDomain(domain)
			if domain == "" || !isValidDomain(domain) {
				continue
			}

			switch src.Type {
			case SourceTypeBlocklist:
				blocklist[domain] = struct{}{}
			case SourceTypeAllowlist:
				allowlist[domain] = struct{}{}
			}
		}
	}

	// Check if we have any successful sources
	if successfulSources == 0 {
		return fmt.Errorf("all sources failed, not updating data.bin to preserve existing data")
	}

	// Load manual additions if provided
	if manualFile != "" {
		log("Loading manual additions from %s...", manualFile)
		manualDomains, err := loadManualFile(manualFile)
		if err != nil {
			log("  Warning: could not load manual file: %v", err)
		} else {
			log("  Loaded %d manual domains", len(manualDomains))
			for _, domain := range manualDomains {
				domain = normalizeDomain(domain)
				if domain != "" && isValidDomain(domain) {
					blocklist[domain] = struct{}{}
				}
			}
		}
	}

	// Check for manual.txt in output directory
	manualPath := filepath.Join(outputDir, "manual.txt")
	if _, err := os.Stat(manualPath); err == nil {
		log("Loading manual additions from %s...", manualPath)
		manualDomains, err := loadManualFile(manualPath)
		if err != nil {
			log("  Warning: could not load manual file: %v", err)
		} else {
			log("  Loaded %d manual domains", len(manualDomains))
			for _, domain := range manualDomains {
				domain = normalizeDomain(domain)
				if domain != "" && isValidDomain(domain) {
					blocklist[domain] = struct{}{}
				}
			}
		}
	}

	// Remove allowlisted domains from blocklist
	for domain := range allowlist {
		delete(blocklist, domain)
	}

	log("Total unique blocklist domains: %d", len(blocklist))
	log("Total unique allowlist domains: %d", len(allowlist))

	// Validate: don't save if we ended up with an empty blocklist
	if len(blocklist) == 0 {
		return fmt.Errorf("blocklist is empty after processing, not updating data.bin to preserve existing data")
	}

	// Build tries
	blocklistTrie := trie.New()
	for domain := range blocklist {
		blocklistTrie.Insert(domain)
	}

	allowlistTrie := trie.New()
	for domain := range allowlist {
		allowlistTrie.Insert(domain)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Serialize and write to file
	log("Writing %s...", outputPath)

	data, err := trie.Serialize(blocklistTrie, allowlistTrie)
	if err != nil {
		return fmt.Errorf("failed to serialize: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Update stats
	stats.NewBlocklistCount = blocklistTrie.Size()
	stats.NewAllowlistCount = allowlistTrie.Size()

	// Compute the exact blocklist changes against the previous data.
	stats.AddedBlocklist = diffSorted(blocklist, oldBlocklistSet)
	stats.RemovedBlocklist = diffSorted(oldBlocklistSet, blocklist)

	// Print stats
	fmt.Printf("Successfully generated %s\n", outputPath)
	fmt.Printf("  Blocklist domains: %d\n", blocklistTrie.Size())
	fmt.Printf("  Allowlist domains: %d\n", allowlistTrie.Size())
	fmt.Printf("  File size: %d bytes (%.2f KB)\n", len(data), float64(len(data))/1024)
	fmt.Printf("  Summary: %s\n", stats.Summary())

	if len(stats.FailedSources) > 0 {
		fmt.Printf("  Failed sources: %s\n", strings.Join(stats.FailedSources, ", "))
	}

	// Write summary to file if requested (for CI)
	if summaryFile != "" {
		if err := os.WriteFile(summaryFile, []byte(stats.Summary()), 0644); err != nil {
			log("Warning: could not write summary file: %v", err)
		}
	}

	// Write release changelog artifacts if requested (for CI)
	if changelogDir != "" {
		if err := stats.writeChangelog(changelogDir); err != nil {
			log("Warning: could not write changelog: %v", err)
		}
	}

	// Also write a text version of the lists for reference
	if verbose {
		if err := writeTextList(filepath.Join(outputDir, "blocklist.txt"), blocklist); err != nil {
			log("  Warning: could not write blocklist.txt: %v", err)
		}
		if err := writeTextList(filepath.Join(outputDir, "allowlist.txt"), allowlist); err != nil {
			log("  Warning: could not write allowlist.txt: %v", err)
		}
	}

	return nil
}

func downloadSource(client *http.Client, url string) ([]string, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return parseLines(resp.Body)
}

func parseLines(r io.Reader) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			lines = append(lines, line)
		}
	}
	return lines, scanner.Err()
}

func loadManualFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return parseLines(f)
}

func normalizeDomain(domain string) string {
	domain = strings.TrimSpace(domain)
	domain = strings.ToLower(domain)
	return domain
}

func isValidDomain(domain string) bool {
	if domain == "" {
		return false
	}

	if !strings.Contains(domain, ".") {
		return false
	}

	parts := strings.Split(domain, ".")
	for _, part := range parts {
		if part == "" {
			return false
		}
		for _, c := range part {
			if !isValidDomainChar(c) {
				return false
			}
		}
	}

	return true
}

func isValidDomainChar(c rune) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= '0' && c <= '9') ||
		c == '-' ||
		c == '_'
}

func writeTextList(path string, domains map[string]struct{}) error {
	sorted := make([]string, 0, len(domains))
	for domain := range domains {
		sorted = append(sorted, domain)
	}
	sort.Strings(sorted)

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintf(f, "# Generated by disposable-update\n")
	fmt.Fprintf(f, "# Date: %s\n", time.Now().UTC().Format(time.RFC3339))
	fmt.Fprintf(f, "# Count: %d\n", len(sorted))
	fmt.Fprintln(f, "#")

	for _, domain := range sorted {
		fmt.Fprintln(f, domain)
	}

	return nil
}
