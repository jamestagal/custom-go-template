// Package builder provides component registry and bundle generation utilities.
package builder

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
)

// GenerateBundleHash generates an 8-character content hash for cache busting.
// Uses SHA256 truncated to 8 hex characters (~4 billion combinations).
//
// Pattern: Pure Function [Load: 2]
func GenerateBundleHash(content []byte) string {
	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:])[:8]
}

// GenerateBundleHashFromComponents generates a deterministic hash from component names.
// Components are sorted to ensure consistent hash output regardless of input order.
//
// Pattern: Pure Function [Load: 3]
func GenerateBundleHashFromComponents(components []string) string {
	// Sort for deterministic output
	sorted := make([]string, len(components))
	copy(sorted, components)
	sort.Strings(sorted)

	// Concatenate and hash
	var combined []byte
	for _, comp := range sorted {
		combined = append(combined, []byte(comp)...)
		combined = append(combined, '\n')
	}

	return GenerateBundleHash(combined)
}

// GenerateVersionHash generates a version hash for the entire bundle manifest.
// This hash changes when any bundle content changes.
//
// Pattern: Pure Function [Load: 3]
func GenerateVersionHash(bundleHashes map[string]string) string {
	// Sort keys for deterministic output
	keys := make([]string, 0, len(bundleHashes))
	for k := range bundleHashes {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Concatenate bundle path:hash pairs
	var combined []byte
	for _, k := range keys {
		combined = append(combined, []byte(k)...)
		combined = append(combined, ':')
		combined = append(combined, []byte(bundleHashes[k])...)
		combined = append(combined, '\n')
	}

	// Generate 12-char version hash
	hash := sha256.Sum256(combined)
	return hex.EncodeToString(hash[:])[:12]
}
