// Package renderer provides HTML rendering utilities.
// This file contains Plenti-compatible HTML generation helpers.
package renderer

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jimafisk/custom_go_template/builder"
)

// PlentiRenderOptions configures Plenti-compatible HTML output.
type PlentiRenderOptions struct {
	// Fingerprint is the build fingerprint for cache-busting (e.g., "aQwupMmCDl")
	Fingerprint string

	// ContentFilePath is the content JSON file path (e.g., "content/pages/about.json")
	// This is added as data-content-filepath on the <html> tag.
	ContentFilePath string

	// BaseHref is the base URL for assets (default: "/")
	BaseHref string
}

// AddPlentiAttributes adds Plenti-compatible attributes to HTML.
// This includes:
//   - data-content-filepath on <html> tag
//   - <base href="/"> in <head>
//   - Fingerprinted asset paths
//
// Example:
//
//	html := AddPlentiAttributes(html, PlentiRenderOptions{
//	    Fingerprint:     "aQwupMmCDl",
//	    ContentFilePath: "content/pages/about.json",
//	})
func AddPlentiAttributes(html string, opts PlentiRenderOptions) string {
	// Set defaults
	if opts.BaseHref == "" {
		opts.BaseHref = "/"
	}

	// Add data-content-filepath to <html> tag
	if opts.ContentFilePath != "" {
		html = addDataContentFilepath(html, opts.ContentFilePath)
	}

	// Add <base href="/"> if not present
	html = ensureBaseHref(html, opts.BaseHref)

	// Update script/stylesheet paths with fingerprint if provided
	if opts.Fingerprint != "" {
		html = fingerprintAssetPaths(html, opts.Fingerprint)
	}

	return html
}

// addDataContentFilepath adds data-content-filepath attribute to the <html> tag.
func addDataContentFilepath(html, contentFilePath string) string {
	htmlTagRegex := regexp.MustCompile(`(?i)<html([^>]*)>`)
	return htmlTagRegex.ReplaceAllStringFunc(html, func(match string) string {
		// Check if attribute already exists
		if strings.Contains(match, "data-content-filepath") {
			return match
		}

		// Insert attribute before closing >
		tagWithoutClose := strings.TrimSuffix(match, ">")
		return fmt.Sprintf(`%s data-content-filepath="%s">`, tagWithoutClose, contentFilePath)
	})
}

// ensureBaseHref ensures a <base href="/"> tag is present in <head>.
func ensureBaseHref(html, baseHref string) string {
	// Check if base tag already exists
	if strings.Contains(html, "<base") {
		return html
	}

	// Insert base tag after <head>
	headRegex := regexp.MustCompile(`(?i)<head>`)
	return headRegex.ReplaceAllString(html, fmt.Sprintf(`<head><base href="%s">`, baseHref))
}

// fingerprintAssetPaths updates asset paths to include the fingerprint.
// This updates:
//   - <script src="..."> tags
//   - <link href="..."> tags
//
// Only paths that start with "/" or are relative are updated.
// CDN URLs (https://, http://) are left unchanged.
func fingerprintAssetPaths(html, fingerprint string) string {
	// Update script src attributes
	scriptRegex := regexp.MustCompile(`(<script[^>]*\ssrc=")(/?)([^"]+)(")`)
	html = scriptRegex.ReplaceAllStringFunc(html, func(match string) string {
		// Skip external URLs
		if strings.Contains(match, "://") {
			return match
		}

		// Extract parts and rewrite
		parts := scriptRegex.FindStringSubmatch(match)
		if len(parts) < 5 {
			return match
		}

		prefix := parts[1]       // `<script ... src="`
		slash := parts[2]        // "/" or ""
		path := parts[3]         // path after optional /
		suffix := parts[4]       // closing "

		// Only fingerprint certain paths (js/, core/, generated/)
		if shouldFingerprintPath(path) {
			newPath := builder.GetFingerprintedPath(fingerprint, path)
			return fmt.Sprintf("%s%s%s%s", prefix, slash, newPath, suffix)
		}

		return match
	})

	// Update link href attributes (for stylesheets)
	linkRegex := regexp.MustCompile(`(<link[^>]*\shref=")(/?)([^"]+)(")`)
	html = linkRegex.ReplaceAllStringFunc(html, func(match string) string {
		// Skip external URLs
		if strings.Contains(match, "://") {
			return match
		}

		// Extract parts and rewrite
		parts := linkRegex.FindStringSubmatch(match)
		if len(parts) < 5 {
			return match
		}

		prefix := parts[1]       // `<link ... href="`
		slash := parts[2]        // "/" or ""
		path := parts[3]         // path after optional /
		suffix := parts[4]       // closing "

		// Only fingerprint certain paths (css/, core/, generated/)
		if shouldFingerprintPath(path) {
			newPath := builder.GetFingerprintedPath(fingerprint, path)
			return fmt.Sprintf("%s%s%s%s", prefix, slash, newPath, suffix)
		}

		return match
	})

	return html
}

// shouldFingerprintPath checks if a path should be fingerprinted.
// We fingerprint paths under: js/, css/, core/, generated/, layouts/
func shouldFingerprintPath(path string) bool {
	prefixes := []string{
		"js/",
		"css/",
		"core/",
		"generated/",
		"layouts/",
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	return false
}

// GeneratePlentiScriptTag generates a script tag with fingerprinted path.
// Example:
//
//	GeneratePlentiScriptTag("aQwupMmCDl", "core/main.js", true)
//	// Returns: <script defer src="/aQwupMmCDl/core/main.js"></script>
func GeneratePlentiScriptTag(fingerprint, path string, defer_ bool) string {
	fullPath := builder.GetFingerprintedPath(fingerprint, path)
	if defer_ {
		return fmt.Sprintf(`<script defer src="/%s"></script>`, fullPath)
	}
	return fmt.Sprintf(`<script src="/%s"></script>`, fullPath)
}

// GeneratePlentiStyleTag generates a link tag with fingerprinted path.
// Example:
//
//	GeneratePlentiStyleTag("aQwupMmCDl", "generated/styles.css")
//	// Returns: <link rel="stylesheet" href="/aQwupMmCDl/generated/styles.css">
func GeneratePlentiStyleTag(fingerprint, path string) string {
	fullPath := builder.GetFingerprintedPath(fingerprint, path)
	return fmt.Sprintf(`<link rel="stylesheet" href="/%s">`, fullPath)
}
