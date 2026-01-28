// Package builder provides component registry and bundle generation utilities.
package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/jimafisk/custom_go_template/types"
)

// TreeShakingConfig configures the tree-shaking bundle generator.
//
// Pattern: Configuration Object [Load: 5]
type TreeShakingConfig struct {
	// ContentDir is the path to the content directory (default: "content")
	ContentDir string

	// LayoutsDir is the path to the layouts directory (default: "layouts")
	LayoutsDir string

	// OutputDir is the path to write generated bundles (default: "generated/bundles")
	OutputDir string

	// Fingerprint is the Plenti-compatible fingerprint for cache busting
	Fingerprint string

	// CommonThreshold is the minimum number of pages for a component to be in common chunk
	// Default: 2 (component used on 2+ pages goes in common chunk)
	CommonThreshold int

	// FallbackPath is the path to the full registry (for fallback loading)
	FallbackPath string
}

// TreeShakingResult contains all generated bundles and metadata.
//
// Pattern: Result Object [Load: 6]
type TreeShakingResult struct {
	// Manifest contains metadata about all bundles
	Manifest *BundleManifest

	// CommonChunk contains the generated common chunk
	CommonChunk *CommonChunkResult

	// PageBundles maps page paths to their generated bundles
	PageBundles map[string]*PageBundleResult

	// OrphanComponents lists components not used by any page
	OrphanComponents []string

	// TotalComponents is the count of all registered components
	TotalComponents int
}

// DefaultTreeShakingConfig returns a default configuration.
//
// Pattern: Factory [Load: 2]
func DefaultTreeShakingConfig() *TreeShakingConfig {
	return &TreeShakingConfig{
		ContentDir:      "content",
		LayoutsDir:      "layouts",
		OutputDir:       "generated/bundles",
		Fingerprint:     "",
		CommonThreshold: 2,
		FallbackPath:    "/generated/layouts.js",
	}
}

// TreeShaker orchestrates the tree-shaking bundle generation process.
//
// Pattern: Facade [Load: 7]
type TreeShaker struct {
	config   *TreeShakingConfig
	registry map[string]*types.ComponentTemplate
	analyzer *UsageAnalyzer
}

// NewTreeShaker creates a new TreeShaker with the given configuration.
//
// Pattern: Constructor [Load: 3]
func NewTreeShaker(config *TreeShakingConfig, registry map[string]*types.ComponentTemplate) *TreeShaker {
	if config == nil {
		config = DefaultTreeShakingConfig()
	}

	return &TreeShaker{
		config:   config,
		registry: registry,
		analyzer: NewUsageAnalyzer(config.ContentDir, config.LayoutsDir),
	}
}

// Generate performs the complete tree-shaking process and returns all results.
//
// Pattern: Template Method [Load: 9]
func (t *TreeShaker) Generate() (*TreeShakingResult, error) {
	// Step 1: Analyze component usage
	_, err := t.analyzer.Analyze()
	if err != nil {
		return nil, fmt.Errorf("TreeShaker.Generate: analyzing usage: %w", err)
	}

	// Step 2: Identify common components (used on 2+ pages)
	commonComponents := t.analyzer.GetCommonComponents(t.config.CommonThreshold)

	// Step 3: Generate common chunk
	commonResult, err := GenerateCommonChunk(commonComponents, t.registry)
	if err != nil {
		return nil, fmt.Errorf("TreeShaker.Generate: generating common chunk: %w", err)
	}

	// Step 4: Generate page bundles
	pageBundles, err := GenerateAllPageBundles(
		t.analyzer,
		t.registry,
		commonComponents,
		commonResult.Hash,
		t.config.Fingerprint,
	)
	if err != nil {
		return nil, fmt.Errorf("TreeShaker.Generate: generating page bundles: %w", err)
	}

	// Step 5: Find orphan components
	orphanComponents := t.findOrphanComponents()

	// Step 6: Calculate fallback size
	fallbackSize := t.calculateFallbackSize()

	// Step 7: Generate manifest
	manifest, err := GenerateBundleManifest(
		t.config.Fingerprint,
		commonResult,
		pageBundles,
		t.config.FallbackPath,
		fallbackSize,
		len(t.registry),
		len(orphanComponents),
	)
	if err != nil {
		return nil, fmt.Errorf("TreeShaker.Generate: generating manifest: %w", err)
	}

	return &TreeShakingResult{
		Manifest:         manifest,
		CommonChunk:      commonResult,
		PageBundles:      pageBundles,
		OrphanComponents: orphanComponents,
		TotalComponents:  len(t.registry),
	}, nil
}

// WriteBundles writes all generated bundles to disk.
//
// Pattern: Writer [Load: 7]
func (t *TreeShaker) WriteBundles(result *TreeShakingResult) error {
	// Ensure output directory exists
	if err := os.MkdirAll(t.config.OutputDir, 0755); err != nil {
		return fmt.Errorf("TreeShaker.WriteBundles: creating output dir: %w", err)
	}

	// Write common chunk
	if result.CommonChunk != nil && len(result.CommonChunk.Components) > 0 {
		commonPath := filepath.Join(t.config.OutputDir, fmt.Sprintf("common.%s.js", result.CommonChunk.Hash))
		if err := os.WriteFile(commonPath, result.CommonChunk.Content, 0644); err != nil {
			return fmt.Errorf("TreeShaker.WriteBundles: writing common chunk: %w", err)
		}
	}

	// Write page bundles
	for page, bundle := range result.PageBundles {
		// Create subdirectory if needed
		bundleDir := filepath.Join(t.config.OutputDir, filepath.Dir(page))
		if err := os.MkdirAll(bundleDir, 0755); err != nil {
			return fmt.Errorf("TreeShaker.WriteBundles: creating bundle dir for %s: %w", page, err)
		}

		bundlePath := filepath.Join(t.config.OutputDir, fmt.Sprintf("%s.%s.js", page, bundle.Hash))
		if err := os.WriteFile(bundlePath, bundle.Content, 0644); err != nil {
			return fmt.Errorf("TreeShaker.WriteBundles: writing bundle for %s: %w", page, err)
		}
	}

	// Write manifest
	manifestPath := filepath.Join(filepath.Dir(t.config.OutputDir), "bundle-manifest.json")
	manifestJSON, err := result.Manifest.ToJSON()
	if err != nil {
		return fmt.Errorf("TreeShaker.WriteBundles: serializing manifest: %w", err)
	}
	if err := os.WriteFile(manifestPath, manifestJSON, 0644); err != nil {
		return fmt.Errorf("TreeShaker.WriteBundles: writing manifest: %w", err)
	}

	return nil
}

// findOrphanComponents returns components that are not used by any page.
//
// Pattern: Query [Load: 5]
func (t *TreeShaker) findOrphanComponents() []string {
	// Get all used components
	usedComponents := make(map[string]bool)
	for _, pu := range t.analyzer.pageUsage {
		for _, comp := range pu.Components {
			usedComponents[comp] = true
		}
	}

	// Find components in registry but not used
	var orphans []string
	for name := range t.registry {
		if !usedComponents[name] {
			orphans = append(orphans, name)
		}
	}

	sort.Strings(orphans)
	return orphans
}

// calculateFallbackSize estimates the size of the full registry.
//
// Pattern: Calculator [Load: 4]
func (t *TreeShaker) calculateFallbackSize() int {
	totalSize := 0
	for _, template := range t.registry {
		if template.AST != nil {
			content := convertToJSTemplate(template.AST)
			totalSize += len(content)
		}
	}
	// Add overhead for registry structure (~50 bytes per component)
	return totalSize + (len(t.registry) * 50)
}

// GetUsageAnalyzer returns the usage analyzer for additional queries.
//
// Pattern: Accessor [Load: 2]
func (t *TreeShaker) GetUsageAnalyzer() *UsageAnalyzer {
	return t.analyzer
}

// PrintSummary prints a summary of the tree-shaking results.
//
// Pattern: Reporter [Load: 5]
func PrintTreeShakingSummary(result *TreeShakingResult) {
	fmt.Println("\n=== Tree-Shaking Summary ===")
	fmt.Printf("Total components: %d\n", result.TotalComponents)
	fmt.Printf("Common components: %d\n", len(result.CommonChunk.Components))
	fmt.Printf("Orphan components: %d\n", len(result.OrphanComponents))
	fmt.Printf("Page bundles: %d\n", len(result.PageBundles))

	if result.Manifest.Stats != nil {
		fmt.Printf("\nStats:\n")
		fmt.Printf("  Average bundle size: %d bytes\n", result.Manifest.Stats.AverageBundleSize)
		fmt.Printf("  Average savings: %.1f%%\n", result.Manifest.Stats.AverageSavingsPercent)
	}

	if result.CommonChunk != nil && len(result.CommonChunk.Components) > 0 {
		fmt.Printf("\nCommon chunk (%d bytes):\n", result.CommonChunk.Size)
		for _, comp := range result.CommonChunk.Components {
			fmt.Printf("  - %s\n", comp)
		}
	}

	fmt.Println("\nPage bundles:")
	for page, bundle := range result.PageBundles {
		fmt.Printf("  %s:\n", page)
		fmt.Printf("    Hash: %s\n", bundle.Hash)
		fmt.Printf("    Unique: %d components (%d bytes)\n", len(bundle.UniqueComponents), bundle.UniqueSize)
		fmt.Printf("    Total: %d bytes\n", bundle.TotalSize)
		fmt.Printf("    Uses common: %v\n", bundle.IncludesCommon)
	}

	if len(result.OrphanComponents) > 0 {
		fmt.Printf("\nOrphan components (not used):\n")
		for _, comp := range result.OrphanComponents {
			fmt.Printf("  - %s\n", comp)
		}
	}
}
