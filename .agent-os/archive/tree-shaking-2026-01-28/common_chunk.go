// Package builder provides component registry and bundle generation utilities.
package builder

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jimafisk/custom_go_template/types"
)

// CommonChunkResult contains the generated common chunk and metadata.
//
// Pattern: Result Object [Load: 4]
type CommonChunkResult struct {
	// Content is the generated JavaScript content
	Content []byte

	// Components is the list of components included in the common chunk
	Components []string

	// Hash is the content hash for cache busting
	Hash string

	// Size is the content size in bytes
	Size int
}

// GenerateCommonChunk generates a JavaScript chunk containing shared components.
// Components used on 2+ pages are extracted to enable caching across pages.
//
// Pattern: Generator [Load: 7]
func GenerateCommonChunk(
	commonComponents []string,
	registry map[string]*types.ComponentTemplate,
) (*CommonChunkResult, error) {
	if len(commonComponents) == 0 {
		return &CommonChunkResult{
			Content:    []byte("// Empty common chunk - no shared components\nexport default {};\n"),
			Components: []string{},
			Hash:       GenerateBundleHash([]byte("empty")),
			Size:       0,
		}, nil
	}

	// Sort for deterministic output
	sorted := make([]string, len(commonComponents))
	copy(sorted, commonComponents)
	sort.Strings(sorted)

	var sb strings.Builder

	// Header
	sb.WriteString("// Auto-generated common component chunk\n")
	sb.WriteString("// Components used on 2+ pages for shared caching\n")
	sb.WriteString(fmt.Sprintf("// Components: %s\n\n", strings.Join(sorted, ", ")))

	sb.WriteString("const common = {\n")

	// Generate each component
	for i, compName := range sorted {
		template := findComponent(compName, registry)
		if template == nil {
			// Component not found - skip with warning comment
			sb.WriteString(fmt.Sprintf("  // WARNING: Component '%s' not found in registry\n", compName))
			continue
		}

		// Convert AST to JavaScript template literal
		templateContent := convertToJSTemplate(template.AST)

		sb.WriteString(fmt.Sprintf("  '%s': (props) => `", compName))
		sb.WriteString(templateContent)
		sb.WriteString("`")

		if i < len(sorted)-1 {
			sb.WriteString(",\n")
		} else {
			sb.WriteString("\n")
		}
	}

	sb.WriteString("};\n\n")
	sb.WriteString("export default common;\n")

	content := []byte(sb.String())
	hash := GenerateBundleHash(content)

	return &CommonChunkResult{
		Content:    content,
		Components: sorted,
		Hash:       hash,
		Size:       len(content),
	}, nil
}

// findComponent looks up a component by name in the registry.
// Tries multiple lookup strategies:
// 1. Exact key match
// 2. Signature format (layouts_components_{name}_html)
// 3. Match by component's Name field
// 4. Case-insensitive match
//
// Pattern: Lookup [Load: 5]
func findComponent(name string, registry map[string]*types.ComponentTemplate) *types.ComponentTemplate {
	// Try exact match first
	if template, exists := registry[name]; exists {
		return template
	}

	// Try Plenti signature format
	signature := types.ShortNameToSignature(name, "html")
	if template, exists := registry[signature]; exists {
		return template
	}

	// Try matching by component's Name field
	lowerName := strings.ToLower(name)
	for _, template := range registry {
		if template.Name == name || strings.ToLower(template.Name) == lowerName {
			return template
		}
	}

	// Try case-insensitive key match
	for regName, template := range registry {
		if strings.ToLower(regName) == lowerName {
			return template
		}
	}

	return nil
}

// CalculateCommonChunkStats calculates size statistics for the common chunk.
//
// Pattern: Calculator [Load: 3]
func CalculateCommonChunkStats(
	commonComponents []string,
	registry map[string]*types.ComponentTemplate,
) (totalSize int, componentSizes map[string]int) {
	componentSizes = make(map[string]int)

	for _, compName := range commonComponents {
		template := findComponent(compName, registry)
		if template == nil {
			continue
		}

		// Estimate size from AST (rough approximation)
		content := convertToJSTemplate(template.AST)
		size := len(content)
		componentSizes[compName] = size
		totalSize += size
	}

	return totalSize, componentSizes
}
