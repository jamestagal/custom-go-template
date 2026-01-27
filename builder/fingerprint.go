// Package builder provides utilities for building Plenti-compatible output.
package builder

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// FingerprintLength is the number of characters in a build fingerprint.
const FingerprintLength = 10

// BuildOutput represents the output paths for a fingerprinted build.
// All paths are relative to the public directory.
type BuildOutput struct {
	// Fingerprint is the 10-character build hash (e.g., "aQwupMmCDl")
	Fingerprint string

	// CoreDir is the path to core runtime files (e.g., "aQwupMmCDl/core/")
	CoreDir string

	// GeneratedDir is the path to generated files (e.g., "aQwupMmCDl/generated/")
	GeneratedDir string

	// LayoutsDir contains subdirectories for components, content, global
	LayoutsDir string

	// ComponentsDir is the path to component layouts
	ComponentsDir string

	// ContentDir is the path to content layouts
	ContentDir string

	// GlobalDir is the path to global layouts
	GlobalDir string
}

// GenerateBuildFingerprint creates a SHA-256 based fingerprint from source files.
// It hashes all .html, .js, and .css files in the source directory.
//
// Returns the first 10 characters of the hex-encoded SHA-256 hash.
// On error, returns an empty string (graceful degradation).
//
// Example:
//
//	fingerprint := GenerateBuildFingerprint("layouts")
//	// Returns something like "aQwupMmCDl"
func GenerateBuildFingerprint(sourceDir string) string {
	hash := sha256.New()

	// Collect all relevant file paths for deterministic ordering
	var filePaths []string

	err := filepath.WalkDir(sourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err // Skip directories we can't access
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Only include .html, .js, .css files
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".html" || ext == ".js" || ext == ".css" {
			filePaths = append(filePaths, path)
		}

		return nil
	})

	if err != nil {
		// Graceful degradation: return empty string on error
		return ""
	}

	// Sort paths for deterministic hash (important for consistency)
	sort.Strings(filePaths)

	// Hash each file's contents
	for _, path := range filePaths {
		if err := hashFile(hash, path); err != nil {
			// Skip files we can't read, continue with others
			continue
		}
	}

	// Get hex-encoded hash and truncate to FingerprintLength characters
	fullHash := hex.EncodeToString(hash.Sum(nil))
	if len(fullHash) >= FingerprintLength {
		return fullHash[:FingerprintLength]
	}
	return fullHash
}

// hashFile adds a file's path and contents to the hash.
// Including the path ensures renamed files produce different hashes.
func hashFile(hash io.Writer, path string) error {
	// Include the relative path in the hash (for rename detection)
	if _, err := io.WriteString(hash, path); err != nil {
		return err
	}

	// Read and hash file contents
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := io.Copy(hash, file); err != nil {
		return err
	}

	return nil
}

// CreateFingerprintedOutput creates the Plenti-compatible directory structure
// for a fingerprinted build.
//
// Directory structure:
//
//	{publicDir}/{fingerprint}/core/
//	{publicDir}/{fingerprint}/generated/
//	{publicDir}/{fingerprint}/layouts/components/
//	{publicDir}/{fingerprint}/layouts/content/
//	{publicDir}/{fingerprint}/layouts/global/
//
// Returns a BuildOutput struct with all paths, or an error if creation fails.
func CreateFingerprintedOutput(publicDir, fingerprint string) (*BuildOutput, error) {
	if fingerprint == "" {
		return nil, fmt.Errorf("CreateFingerprintedOutput: fingerprint cannot be empty")
	}

	// Build the output paths
	output := &BuildOutput{
		Fingerprint:   fingerprint,
		CoreDir:       filepath.Join(fingerprint, "core"),
		GeneratedDir:  filepath.Join(fingerprint, "generated"),
		LayoutsDir:    filepath.Join(fingerprint, "layouts"),
		ComponentsDir: filepath.Join(fingerprint, "layouts", "components"),
		ContentDir:    filepath.Join(fingerprint, "layouts", "content"),
		GlobalDir:     filepath.Join(fingerprint, "layouts", "global"),
	}

	// Create all directories
	dirs := []string{
		filepath.Join(publicDir, output.CoreDir),
		filepath.Join(publicDir, output.GeneratedDir),
		filepath.Join(publicDir, output.ComponentsDir),
		filepath.Join(publicDir, output.ContentDir),
		filepath.Join(publicDir, output.GlobalDir),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("CreateFingerprintedOutput: failed to create %s: %w", dir, err)
		}
	}

	return output, nil
}

// GetFingerprintedPath returns a fingerprinted path for an asset.
// This is used to generate script/link tags with fingerprinted URLs.
//
// Example:
//
//	GetFingerprintedPath("aQwupMmCDl", "core/main.js")
//	// Returns "aQwupMmCDl/core/main.js"
func GetFingerprintedPath(fingerprint, assetPath string) string {
	if fingerprint == "" {
		return assetPath
	}
	return filepath.ToSlash(filepath.Join(fingerprint, assetPath))
}

// CleanOldFingerprints removes old fingerprinted directories from the public directory.
// It keeps only the current fingerprint directory.
//
// This is useful for cleaning up after incremental builds.
func CleanOldFingerprints(publicDir, currentFingerprint string) error {
	entries, err := os.ReadDir(publicDir)
	if err != nil {
		return fmt.Errorf("CleanOldFingerprints: failed to read %s: %w", publicDir, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Skip current fingerprint and non-fingerprint directories
		if name == currentFingerprint {
			continue
		}

		// Only remove directories that look like fingerprints (10 hex chars)
		if len(name) == FingerprintLength && isHexString(name) {
			dirPath := filepath.Join(publicDir, name)
			if err := os.RemoveAll(dirPath); err != nil {
				// Log but don't fail - cleanup is best-effort
				continue
			}
		}
	}

	return nil
}

// isHexString checks if a string contains only hexadecimal characters.
func isHexString(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
