// Example: Basic usage of the disposable email checker
package main

import (
	"fmt"

	disposable "github.com/rezmoss/go-is-disposable-email"
)

func main() {
	// Test various email addresses
	emails := []string{
		"user@gmail.com",
		"user@tempmail.com",
		"user@10minutemail.com",
		"user@outlook.com",
		"user@guerrillamail.com",
		"user@yahoo.com",
		"user@mailinator.com",
	}

	fmt.Println("Checking emails for disposable domains:")
	fmt.Println("========================================")

	for _, email := range emails {
		isDisposable := disposable.IsDisposable(email)
		status := "legitimate"
		if isDisposable {
			status = "DISPOSABLE"
		}
		fmt.Printf("%-35s %s\n", email, status)
	}

	// Print statistics
	fmt.Println()
	stats := disposable.Stats()
	fmt.Printf("Database statistics:\n")
	fmt.Printf("  Blocklist domains: %d\n", stats.BlocklistCount)
	fmt.Printf("  Allowlist domains: %d\n", stats.AllowlistCount)
	fmt.Printf("  Version: %s\n", stats.Version)
	fmt.Printf("  Mode: %s\n", stats.Mode)
}
