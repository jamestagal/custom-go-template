// Package builder provides component registry and bundle generation utilities.
package builder

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

// BundleManifest contains metadata about all generated bundles.
// Used for runtime bundle loading and build tooling.
//
// Pattern: Data Transfer Object [Load: 6]
type BundleManifest struct {
	// Version is a unique identifier for this build
	Version string `json:"version"`

	// Generated is the timestamp when the manifest was created
	Generated string `json:"generated"`

	// Fingerprint is the Plenti-compatible fingerprint directory
	Fingerprint string `json:"fingerprint"`

	// Common contains metadata about the shared component chunk
	Common *CommonChunkManifest `json:"common,omitempty"`

	// Bundles maps page paths to their bundle metadata
	Bundles map[string]*PageBundleManifest `json:"bundles"`

	// Fallback contains the path to the full registry (fallback)
	Fallback *FallbackManifest `json:"fallback"`

	// Stats contains aggregate statistics about the bundles
	Stats *BundleStats `json:"stats"`
}

// CommonChunkManifest contains metadata about the common chunk.
//
// Pattern: DTO [Load: 3]
type CommonChunkManifest struct {
	// Path is the URL path to the common chunk
	Path string `json:"path"`

	// Components is the list of component names in the common chunk
	Components []string `json:"components"`

	// Size is the file size in bytes
	Size int `json:"size"`
}

// PageBundleManifest contains metadata about a page-specific bundle.
//
// Pattern: DTO [Load: 4]
type PageBundleManifest struct {
	// Path is the URL path to the bundle
	Path string `json:"path"`

	// Components is the list of unique component names in this bundle
	Components []string `json:"components"`

	// IncludesCommon indicates if this bundle imports the common chunk
	IncludesCommon bool `json:"includes_common"`

	// TotalSize is the total component size (including common)
	TotalSize int `json:"total_size"`

	// UniqueSize is the size of unique components only
	UniqueSize int `json:"unique_size"`
}

// FallbackManifest contains metadata about the fallback registry.
//
// Pattern: DTO [Load: 2]
type FallbackManifest struct {
	// Path is the URL path to the full registry
	Path string `json:"path"`

	// Size is the file size in bytes
	Size int `json:"size"`
}

// BundleStats contains aggregate statistics about the bundles.
//
// Pattern: DTO [Load: 4]
type BundleStats struct {
	// TotalComponents is the total number of components
	TotalComponents int `json:"total_components"`

	// CommonComponents is the number of shared components
	CommonComponents int `json:"common_components"`

	// OrphanComponents is the number of unused components
	OrphanComponents int `json:"orphan_components"`

	// AverageBundleSize is the average bundle size in bytes
	AverageBundleSize int `json:"average_bundle_size"`

	// AverageSavingsPercent is the average savings compared to full registry
	AverageSavingsPercent float64 `json:"average_savings_percent"`
}

// GenerateBundleManifest creates a manifest from bundle generation results.
//
// Pattern: Builder [Load: 8]
func GenerateBundleManifest(
	fingerprint string,
	commonResult *CommonChunkResult,
	pageBundles map[string]*PageBundleResult,
	fallbackPath string,
	fallbackSize int,
	totalComponents int,
	orphanComponents int,
) (*BundleManifest, error) {
	// Generate version hash from all bundle hashes
	bundleHashes := make(map[string]string)
	if commonResult != nil && len(commonResult.Components) > 0 {
		bundleHashes["common"] = commonResult.Hash
	}
	for page, bundle := range pageBundles {
		bundleHashes[page] = bundle.Hash
	}
	version := GenerateVersionHash(bundleHashes)

	manifest := &BundleManifest{
		Version:     version,
		Generated:   time.Now().UTC().Format(time.RFC3339),
		Fingerprint: fingerprint,
		Bundles:     make(map[string]*PageBundleManifest),
		Fallback: &FallbackManifest{
			Path: fallbackPath,
			Size: fallbackSize,
		},
	}

	// Build path prefix (handle empty fingerprint)
	pathPrefix := "/generated"
	if fingerprint != "" {
		pathPrefix = "/" + fingerprint + "/generated"
	}

	// Add common chunk if present
	if commonResult != nil && len(commonResult.Components) > 0 {
		manifest.Common = &CommonChunkManifest{
			Path:       fmt.Sprintf("%s/bundles/common.%s.js", pathPrefix, commonResult.Hash),
			Components: commonResult.Components,
			Size:       commonResult.Size,
		}
	}

	// Add page bundles
	var totalBundleSize int
	for page, bundle := range pageBundles {
		manifest.Bundles[page] = &PageBundleManifest{
			Path:           fmt.Sprintf("%s/bundles/%s.%s.js", pathPrefix, page, bundle.Hash),
			Components:     bundle.UniqueComponents,
			IncludesCommon: bundle.IncludesCommon,
			TotalSize:      bundle.TotalSize,
			UniqueSize:     bundle.UniqueSize,
		}
		totalBundleSize += bundle.TotalSize
	}

	// Calculate stats
	numBundles := len(pageBundles)
	avgBundleSize := 0
	if numBundles > 0 {
		avgBundleSize = totalBundleSize / numBundles
	}

	avgSavings := 0.0
	if fallbackSize > 0 && numBundles > 0 {
		avgSavings = (1.0 - float64(avgBundleSize)/float64(fallbackSize)) * 100
	}

	commonCount := 0
	if commonResult != nil {
		commonCount = len(commonResult.Components)
	}

	manifest.Stats = &BundleStats{
		TotalComponents:       totalComponents,
		CommonComponents:      commonCount,
		OrphanComponents:      orphanComponents,
		AverageBundleSize:     avgBundleSize,
		AverageSavingsPercent: avgSavings,
	}

	return manifest, nil
}

// ToJSON serializes the manifest to JSON with indentation.
//
// Pattern: Serializer [Load: 2]
func (m *BundleManifest) ToJSON() ([]byte, error) {
	return json.MarshalIndent(m, "", "  ")
}

// GetBundlePath returns the bundle path for a given page.
// Returns the fallback path if page not found.
//
// Pattern: Accessor [Load: 3]
func (m *BundleManifest) GetBundlePath(page string) string {
	if bundle, exists := m.Bundles[page]; exists {
		return bundle.Path
	}
	return m.Fallback.Path
}

// GetCommonPath returns the common chunk path, or empty string if not present.
//
// Pattern: Accessor [Load: 2]
func (m *BundleManifest) GetCommonPath() string {
	if m.Common != nil {
		return m.Common.Path
	}
	return ""
}

// GetAllBundlePaths returns all bundle paths (for file generation).
//
// Pattern: Query [Load: 4]
func (m *BundleManifest) GetAllBundlePaths() []string {
	var paths []string

	if m.Common != nil {
		paths = append(paths, m.Common.Path)
	}

	// Sort pages for deterministic output
	pages := make([]string, 0, len(m.Bundles))
	for page := range m.Bundles {
		pages = append(pages, page)
	}
	sort.Strings(pages)

	for _, page := range pages {
		paths = append(paths, m.Bundles[page].Path)
	}

	return paths
}

// GetScriptTagAttributes returns the data attributes for the script tag.
//
// Pattern: Formatter [Load: 4]
func (m *BundleManifest) GetScriptTagAttributes(page string) map[string]string {
	attrs := make(map[string]string)

	attrs["data-bundle"] = m.GetBundlePath(page)
	attrs["data-fallback"] = m.Fallback.Path

	if bundle, exists := m.Bundles[page]; exists && bundle.IncludesCommon {
		if m.Common != nil {
			attrs["data-common"] = m.Common.Path
		}
	}

	return attrs
}
