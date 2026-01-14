// Package trie provides an efficient trie (prefix tree) data structure
// optimized for domain name lookups.
package trie

import (
	"sync"
)

// Node represents a node in the trie.
type Node struct {
	Children map[rune]*Node
	IsEnd    bool // Marks the end of a domain
}

// NewNode creates a new trie node.
func NewNode() *Node {
	return &Node{
		Children: make(map[rune]*Node),
		IsEnd:    false,
	}
}

// Trie is a thread-safe prefix tree for domain lookups.
type Trie struct {
	mu   sync.RWMutex
	root *Node
	size int
}

// New creates a new empty trie.
func New() *Trie {
	return &Trie{
		root: NewNode(),
		size: 0,
	}
}

// Insert adds a domain to the trie.
// The domain is stored in reverse order for efficient suffix matching.
func (t *Trie) Insert(domain string) {
	if domain == "" {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	// Reverse the domain for efficient suffix matching
	reversed := reverseString(domain)

	node := t.root
	for _, char := range reversed {
		if node.Children[char] == nil {
			node.Children[char] = NewNode()
		}
		node = node.Children[char]
	}

	if !node.IsEnd {
		node.IsEnd = true
		t.size++
	}
}

// Contains checks if the exact domain exists in the trie.
func (t *Trie) Contains(domain string) bool {
	if domain == "" {
		return false
	}

	t.mu.RLock()
	defer t.mu.RUnlock()

	reversed := reverseString(domain)
	node := t.root

	for _, char := range reversed {
		if node.Children[char] == nil {
			return false
		}
		node = node.Children[char]
	}

	return node.IsEnd
}

// ContainsHierarchical checks if the domain or any of its parent domains
// exist in the trie. For example, if "tempmail.com" is in the trie,
// this returns true for "mail.tempmail.com".
func (t *Trie) ContainsHierarchical(domain string) bool {
	if domain == "" {
		return false
	}

	t.mu.RLock()
	defer t.mu.RUnlock()

	// Reverse the domain
	reversed := reverseString(domain)
	node := t.root

	for _, char := range reversed {
		if node.Children[char] == nil {
			return false
		}
		node = node.Children[char]

		// Check if this is the end of a stored domain
		// This means the current suffix matches a domain in the trie
		if node.IsEnd {
			return true
		}
	}

	return node.IsEnd
}

// Size returns the number of domains in the trie.
func (t *Trie) Size() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.size
}

// GetAll returns all domains stored in the trie.
func (t *Trie) GetAll() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var domains []string
	t.collectDomains(t.root, "", &domains)
	return domains
}

// collectDomains recursively collects all domains from the trie.
func (t *Trie) collectDomains(node *Node, prefix string, domains *[]string) {
	if node.IsEnd {
		// Reverse back to get the original domain
		*domains = append(*domains, reverseString(prefix))
	}

	for char, child := range node.Children {
		t.collectDomains(child, prefix+string(char), domains)
	}
}

// Clear removes all domains from the trie.
func (t *Trie) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.root = NewNode()
	t.size = 0
}

// GetRoot returns the root node (used for serialization).
func (t *Trie) GetRoot() *Node {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.root
}

// SetRoot sets the root node (used for deserialization).
func (t *Trie) SetRoot(root *Node, size int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.root = root
	t.size = size
}

// reverseString reverses a string.
func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
