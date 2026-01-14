package disposable

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	// Set up test data before running tests
	if err := setupTestData(); err != nil {
		// Log but don't fail - tests will fail with more specific errors
		os.Stderr.WriteString("Warning: failed to setup test data: " + err.Error() + "\n")
	}

	os.Exit(m.Run())
}

// setupTestData copies data/data.bin to the cache directory for tests.
// This ensures tests work in CI where there's no pre-existing cache.
func setupTestData() error {
	// Find the data.bin file relative to the package
	srcPath := filepath.Join("data", "data.bin")

	// Check if source file exists
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		// Try parent directory (in case tests run from subdirectory)
		srcPath = filepath.Join("..", "data", "data.bin")
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			return nil // No local data.bin, will download from GitHub
		}
	}

	// Get cache directory
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return err
	}

	destDir := filepath.Join(cacheDir, "disposable-email")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	destPath := filepath.Join(destDir, "data.bin")

	// Check if dest already exists and is newer/same
	srcInfo, _ := os.Stat(srcPath)
	if destInfo, err := os.Stat(destPath); err == nil {
		if destInfo.ModTime().After(srcInfo.ModTime()) || destInfo.ModTime().Equal(srcInfo.ModTime()) {
			return nil // Destination is up to date
		}
	}

	// Copy file
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}

	return os.WriteFile(destPath, data, 0644)
}
