// Package builder provides component registry and bundle generation utilities.
package builder

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jimafisk/custom_go_template/types"
)

// PageBundleResult contains the generated page bundle and metadata.
//
// Pattern: Result Object [Load: 5]
type PageBundleResult struct {
	// Page is the page identifier (e.g., "pages/_index")
	Page string

	// Content is the generated JavaScript content
	Content []byte

	// UniqueComponents is the list of components unique to this page
	UniqueComponents []string

	// IncludesCommon indicates if this bundle imports the common chunk
	IncludesCommon bool

	// Hash is the content hash for cache busting
	Hash string

	// TotalSize is the total size including common chunk components
	TotalSize int

	// UniqueSize is the size of unique components only
	UniqueSize int
}

// GeneratePageBundle generates a JavaScript bundle for a specific page.
// Imports the common chunk if the page uses shared components.
//
// Pattern: Generator [Load: 8]
func GeneratePageBundle(
	page string,
	uniqueComponents []string,
	commonComponents []string,
	registry map[string]*types.ComponentTemplate,
	commonChunkPath string, // Relative path to common chunk (e.g., "../common.abc123.js")
) (*PageBundleResult, error) {
	// Sort for deterministic output
	sortedUnique := make([]string, len(uniqueComponents))
	copy(sortedUnique, uniqueComponents)
	sort.Strings(sortedUnique)

	// Determine if we need to import the common chunk
	includesCommon := len(commonComponents) > 0 && commonChunkPath != ""

	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("// Auto-generated bundle for %s\n", page))
	if len(sortedUnique) > 0 {
		sb.WriteString(fmt.Sprintf("// Unique components: %s\n", strings.Join(sortedUnique, ", ")))
	}
	if includesCommon {
		sb.WriteString(fmt.Sprintf("// Imports common chunk with: %s\n", strings.Join(commonComponents, ", ")))
	}
	sb.WriteString("\n")

	// Import common chunk if needed
	if includesCommon {
		sb.WriteString(fmt.Sprintf("import common from '%s';\n\n", commonChunkPath))
	}

	// Generate unique components
	if len(sortedUnique) > 0 {
		sb.WriteString("const unique = {\n")

		for i, compName := range sortedUnique {
			template := findComponent(compName, registry)
			if template == nil {
				sb.WriteString(fmt.Sprintf("  // WARNING: Component '%s' not found in registry\n", compName))
				continue
			}

			templateContent := convertToJSTemplate(template.AST)

			sb.WriteString(fmt.Sprintf("  '%s': (props) => `", compName))
			sb.WriteString(templateContent)
			sb.WriteString("`")

			if i < len(sortedUnique)-1 {
				sb.WriteString(",\n")
			} else {
				sb.WriteString("\n")
			}
		}

		sb.WriteString("};\n\n")
	}

	// Export merged registry
	if includesCommon && len(sortedUnique) > 0 {
		sb.WriteString("export default { ...common, ...unique };\n")
	} else if includesCommon {
		sb.WriteString("export default common;\n")
	} else if len(sortedUnique) > 0 {
		sb.WriteString("export default unique;\n")
	} else {
		sb.WriteString("export default {};\n")
	}

	content := []byte(sb.String())
	hash := GenerateBundleHash(content)

	// Calculate sizes
	uniqueSize := 0
	for _, compName := range sortedUnique {
		template := findComponent(compName, registry)
		if template != nil {
			uniqueSize += len(convertToJSTemplate(template.AST))
		}
	}

	totalSize := uniqueSize
	for _, compName := range commonComponents {
		template := findComponent(compName, registry)
		if template != nil {
			totalSize += len(convertToJSTemplate(template.AST))
		}
	}

	return &PageBundleResult{
		Page:             page,
		Content:          content,
		UniqueComponents: sortedUnique,
		IncludesCommon:   includesCommon,
		Hash:             hash,
		TotalSize:        totalSize,
		UniqueSize:       uniqueSize,
	}, nil
}

// GenerateAllPageBundles generates bundles for all pages.
//
// Pattern: Batch Generator [Load: 6]
func GenerateAllPageBundles(
	analyzer *UsageAnalyzer,
	registry map[string]*types.ComponentTemplate,
	commonComponents []string,
	commonChunkHash string,
	fingerprint string,
) (map[string]*PageBundleResult, error) {
	results := make(map[string]*PageBundleResult)

	pages := analyzer.GetAllPages()

	for _, page := range pages {
		// Get unique components for this page
		uniqueComponents := analyzer.GetPageComponents(page, commonComponents)

		// Determine which common components this page uses
		pageCommonComponents := getPageCommonComponents(analyzer, page, commonComponents)

		// Build common chunk path
		var commonChunkPath string
		if len(pageCommonComponents) > 0 {
			// Calculate relative path from page bundle to common chunk
			// bundles/pages/about.hash.js -> ../common.hash.js
			depth := strings.Count(page, "/")
			prefix := strings.Repeat("../", depth+1)
			commonChunkPath = fmt.Sprintf("%scommon.%s.js", prefix, commonChunkHash)
		}

		result, err := GeneratePageBundle(
			page,
			uniqueComponents,
			pageCommonComponents,
			registry,
			commonChunkPath,
		)
		if err != nil {
			return nil, fmt.Errorf("GenerateAllPageBundles: generating bundle for %s: %w", page, err)
		}

		results[page] = result
	}

	return results, nil
}

// getPageCommonComponents returns the common components used by a specific page.
//
// Pattern: Filter [Load: 4]
func getPageCommonComponents(analyzer *UsageAnalyzer, page string, commonComponents []string) []string {
	pageUsage, exists := analyzer.pageUsage[page]
	if !exists {
		return nil
	}

	commonSet := make(map[string]bool)
	for _, c := range commonComponents {
		commonSet[c] = true
	}

	var pageCommon []string
	for _, comp := range pageUsage.Components {
		if commonSet[comp] {
			pageCommon = append(pageCommon, comp)
		}
	}

	sort.Strings(pageCommon)
	return pageCommon
}
