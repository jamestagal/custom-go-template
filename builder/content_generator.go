// Package builder provides utilities for building Plenti-compatible output.
// This file implements content.js generation for Plenti runtime.
package builder

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ContentEntry represents a single content item in the allContent array.
// This matches Plenti's content format for runtime lookup.
type ContentEntry struct {
	// Type is the content type (folder name, e.g., "pages", "news", "blog")
	Type string `json:"type"`

	// Path is the URL path without leading slash or extension (e.g., "about")
	Path string `json:"path"`

	// Filepath is the full content file path (e.g., "content/pages/about.json")
	Filepath string `json:"filepath"`

	// Filename is just the filename (e.g., "about.json")
	Filename string `json:"filename"`

	// Fields contains the parsed JSON content from the file
	Fields map[string]interface{} `json:"fields"`
}

// GenerateContentJS scans a content directory and generates Plenti-compatible
// content.js module with all content entries.
//
// Directory structure expected:
//
//	content/
//	  pages/
//	    about.json
//	    _index.json
//	  news/
//	    article.json
//
// Output format:
//
//	const allContent = [
//	  { type: "pages", path: "about", filepath: "content/pages/about.json", ... },
//	  ...
//	];
//	export default allContent;
//
// Returns the JavaScript module source, or an error if scanning fails.
func GenerateContentJS(contentDir string) (string, error) {
	entries, err := ScanContentDirectory(contentDir)
	if err != nil {
		return "", fmt.Errorf("GenerateContentJS: %w", err)
	}

	return FormatContentJS(entries), nil
}

// ScanContentDirectory recursively scans a content directory and returns
// all content entries in Plenti-compatible format.
//
// Handles:
// - Nested directories (content type from first-level folder)
// - JSON files at any depth
// - Special files like _index.json and _defaults.json
func ScanContentDirectory(contentDir string) ([]ContentEntry, error) {
	var entries []ContentEntry

	// Check if directory exists
	info, err := os.Stat(contentDir)
	if err != nil {
		if os.IsNotExist(err) {
			// Empty content directory is valid - return empty list
			return entries, nil
		}
		return nil, fmt.Errorf("ScanContentDirectory: failed to stat %s: %w", contentDir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("ScanContentDirectory: %s is not a directory", contentDir)
	}

	// Walk the content directory
	err = filepath.WalkDir(contentDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Skip inaccessible directories
			return nil
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Only process JSON files
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".json") {
			return nil
		}

		// Parse content file
		entry, err := ParseContentFile(contentDir, path)
		if err != nil {
			// Log warning but continue processing other files
			fmt.Fprintf(os.Stderr, "Warning: failed to parse %s: %v\n", path, err)
			return nil
		}

		entries = append(entries, *entry)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("ScanContentDirectory: walk failed: %w", err)
	}

	// Sort entries for deterministic output
	sort.Slice(entries, func(i, j int) bool {
		// Sort by type first, then by path
		if entries[i].Type != entries[j].Type {
			return entries[i].Type < entries[j].Type
		}
		return entries[i].Path < entries[j].Path
	})

	return entries, nil
}

// ParseContentFile reads a JSON content file and creates a ContentEntry.
//
// Parameters:
//   - contentDir: The base content directory (e.g., "content")
//   - filePath: The full path to the JSON file
//
// Returns a ContentEntry with type, path, filepath, filename, and fields.
func ParseContentFile(contentDir, filePath string) (*ContentEntry, error) {
	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("ParseContentFile: failed to read %s: %w", filePath, err)
	}

	// Parse JSON
	var fields map[string]interface{}
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, fmt.Errorf("ParseContentFile: failed to parse JSON in %s: %w", filePath, err)
	}

	// Get relative path from content directory
	relPath, err := filepath.Rel(contentDir, filePath)
	if err != nil {
		return nil, fmt.Errorf("ParseContentFile: failed to get relative path: %w", err)
	}

	// Use forward slashes for filepath (Plenti convention)
	relPath = filepath.ToSlash(relPath)

	// Extract type from first directory component
	// e.g., "pages/about.json" → "pages"
	// e.g., "pages/nested/deep.json" → "pages"
	contentType := extractContentType(relPath)

	// Extract path (URL path without extension)
	// e.g., "pages/about.json" → "about"
	// e.g., "pages/_index.json" → "" (root)
	// e.g., "pages/news/article.json" → "news/article"
	urlPath := extractURLPath(relPath, contentType)

	// Get filename
	filename := filepath.Base(filePath)

	// Build filepath with "content" prefix (Plenti convention)
	// Always use "content" regardless of actual directory name for Plenti compatibility
	contentFilepath := "content/" + relPath

	return &ContentEntry{
		Type:     contentType,
		Path:     urlPath,
		Filepath: contentFilepath,
		Filename: filename,
		Fields:   fields,
	}, nil
}

// extractContentType extracts the content type from a relative path.
// The type is the first directory component.
//
// Examples:
//
//	"pages/about.json" → "pages"
//	"news/article.json" → "news"
//	"blog/2024/post.json" → "blog"
//	"index.json" → "" (root level)
func extractContentType(relPath string) string {
	parts := strings.Split(relPath, "/")
	if len(parts) <= 1 {
		// File at root level - no type
		return ""
	}
	return parts[0]
}

// extractURLPath extracts the URL path from a content file path.
// Removes the type prefix and .json extension.
//
// Special handling for _index.json: becomes the type root.
//
// Examples:
//
//	"pages/about.json", "pages" → "about"
//	"pages/_index.json", "pages" → ""
//	"news/2024/article.json", "news" → "2024/article"
//	"blog/_defaults.json", "blog" → "_defaults"
func extractURLPath(relPath, contentType string) string {
	// Remove type prefix
	path := relPath
	if contentType != "" {
		path = strings.TrimPrefix(relPath, contentType+"/")
	}

	// Remove .json extension
	path = strings.TrimSuffix(path, ".json")

	// Special handling for _index
	if path == "_index" {
		return ""
	}

	return path
}

// FormatContentJS formats content entries as a JavaScript ES module.
//
// Output format:
//
//	// Auto-generated Plenti-compatible content module
//	// Generated by Go template engine
//
//	const allContent = [
//	  {
//	    type: "pages",
//	    path: "about",
//	    filepath: "content/pages/about.json",
//	    filename: "about.json",
//	    fields: { ... }
//	  },
//	  ...
//	];
//
//	export default allContent;
func FormatContentJS(entries []ContentEntry) string {
	var sb strings.Builder

	// Header
	sb.WriteString("// Auto-generated Plenti-compatible content module\n")
	sb.WriteString("// Generated by Go template engine\n\n")

	sb.WriteString("const allContent = [\n")

	for i, entry := range entries {
		// Format fields as JSON
		fieldsJSON, err := json.MarshalIndent(entry.Fields, "    ", "  ")
		if err != nil {
			// Fallback to empty object on error
			fieldsJSON = []byte("{}")
		}

		sb.WriteString("  {\n")
		sb.WriteString(fmt.Sprintf("    type: %q,\n", entry.Type))
		sb.WriteString(fmt.Sprintf("    path: %q,\n", entry.Path))
		sb.WriteString(fmt.Sprintf("    filepath: %q,\n", entry.Filepath))
		sb.WriteString(fmt.Sprintf("    filename: %q,\n", entry.Filename))
		sb.WriteString(fmt.Sprintf("    fields: %s\n", string(fieldsJSON)))
		sb.WriteString("  }")

		if i < len(entries)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}

	sb.WriteString("];\n\n")
	sb.WriteString("export default allContent;\n")

	return sb.String()
}

// WriteContentJS generates and writes content.js to the specified path.
// This is a convenience function that combines GenerateContentJS with file writing.
func WriteContentJS(contentDir, outputPath string) error {
	js, err := GenerateContentJS(contentDir)
	if err != nil {
		return fmt.Errorf("WriteContentJS: %w", err)
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("WriteContentJS: failed to create output directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(outputPath, []byte(js), 0644); err != nil {
		return fmt.Errorf("WriteContentJS: failed to write file: %w", err)
	}

	return nil
}
