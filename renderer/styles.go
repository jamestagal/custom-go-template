package renderer

import (
	"crypto/sha256"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/jimafisk/custom_go_template/ast"
	"github.com/jimafisk/custom_go_template/transformer"
)

// Cache management for style aggregation
var (
	// componentStyleCache stores aggregated styles per component
	componentStyleCache = make(map[string]string)

	// styleCacheMutex protects concurrent access to cache
	styleCacheMutex sync.RWMutex

	// cacheEnabled allows disabling cache for testing
	cacheEnabled = true
)

// StyleBlock represents a style block with metadata for aggregation
type StyleBlock struct {
	Content string // CSS content
	Source  string // Component name that contributed this style
	Hash    string // SHA256 hash for deduplication
}

// extractComponentNameFromPath extracts component name from a path
// Example: "./components/UserProfile.html" -> "UserProfile"
// Example: "{path}" -> "" (skip dynamic runtime paths)
func extractComponentNameFromPath(path string) string {
	// Skip paths with runtime variables like {path}
	if strings.Contains(path, "{") {
		return ""
	}

	// Remove quotes if present
	path = strings.Trim(path, "\"'")

	// Extract filename from path
	parts := strings.Split(path, "/")
	filename := parts[len(parts)-1]

	// Remove extension
	if idx := strings.LastIndex(filename, "."); idx != -1 {
		return filename[:idx]
	}

	return filename
}

// findComponentNodes recursively finds all component nodes in the AST tree
// Returns a list of component names that should have their styles aggregated
func findComponentNodes(nodes []ast.Node) []string {
	var componentNames []string

	for _, node := range nodes {
		switch n := node.(type) {
		case *ast.ComponentNode:
			// Regular component: <MyComponent />
			componentNames = append(componentNames, n.Name)

		case *ast.DynamicComponentNode:
			// Dynamic component: <='./components/Foo.html' />
			// Extract component name from path (skip runtime variable paths)
			componentName := extractComponentNameFromPath(n.PathExpression)
			if componentName != "" {
				componentNames = append(componentNames, componentName)
			}

		case *ast.Element:
			// Recursively check children
			componentNames = append(componentNames, findComponentNodes(n.Children)...)

		case *ast.Conditional:
			// Check conditional branches
			componentNames = append(componentNames, findComponentNodes(n.IfContent)...)
			for _, elseIfContent := range n.ElseIfContent {
				componentNames = append(componentNames, findComponentNodes(elseIfContent)...)
			}
			componentNames = append(componentNames, findComponentNodes(n.ElseContent)...)

		case *ast.Loop:
			// Check loop content
			componentNames = append(componentNames, findComponentNodes(n.Content)...)
		}
	}

	return componentNames
}

// AggregateComponentStyles traverses the component tree and aggregates styles
//
// Algorithm:
// 1. Use depth-first traversal to collect styles (dependencies first)
// 2. Track visited components to prevent infinite loops (circular dependency protection)
// 3. Deduplicate based on content hash (SHA256)
// 4. Return aggregated CSS string with source comments
//
// Example:
//   template := &ast.Template{
//     RootNodes: []ast.Node{
//       &ast.StyleSection{Content: ".header { color: blue; }"},
//     },
//   }
//   result := AggregateComponentStyles(template, "Header")
//   // Returns:
//   // /* Styles from: Header */
//   // .header { color: blue; }
func AggregateComponentStyles(rootTemplate *ast.Template, componentName string) string {
	// Handle nil template gracefully
	if rootTemplate == nil {
		return ""
	}

	// Track visited components to prevent infinite loops
	visited := make(map[string]bool)

	// Store style blocks by hash for deduplication
	styleBlocks := make(map[string]*StyleBlock)

	// Preserve insertion order for deterministic output
	var orderedHashes []string

	// Recursive collection function
	var collectStyles func(template *ast.Template, name string)
	collectStyles = func(template *ast.Template, name string) {
		// Prevent infinite loops from circular dependencies
		if visited[name] {
			return
		}
		visited[name] = true

		// Handle nil or empty RootNodes
		if template == nil || template.RootNodes == nil {
			return
		}

		// PHASE 1: Collect all component names from the tree
		// This includes both FenceSection.Imports AND ComponentNode/DynamicComponentNode usage
		var allComponentNames []string

		// Find components used in the template body (ComponentNode and DynamicComponentNode)
		allComponentNames = append(allComponentNames, findComponentNodes(template.RootNodes)...)

		// Also add components from FenceSection imports
		for _, node := range template.RootNodes {
			if fence, ok := node.(*ast.FenceSection); ok {
				if fence.Imports != nil {
					for _, imp := range fence.Imports {
						allComponentNames = append(allComponentNames, imp.Name)
					}
				}
			}
		}

		// Deduplicate component names
		seenComponents := make(map[string]bool)
		var uniqueComponentNames []string
		for _, compName := range allComponentNames {
			if !seenComponents[compName] {
				seenComponents[compName] = true
				uniqueComponentNames = append(uniqueComponentNames, compName)
			}
		}

		// Process all discovered components recursively (dependencies first)
		for _, compName := range uniqueComponentNames {
			if compTemplate, exists := transformer.GetComponentTemplate(compName); exists {
				// Recursively collect styles from dependency
				collectStyles(compTemplate.Template, compName)
			}
			// Gracefully skip missing components (no panic)
		}

		// PHASE 2: Process this component's styles
		for _, node := range template.RootNodes {
			if styleSection, ok := node.(*ast.StyleSection); ok {
				// Skip empty or whitespace-only style blocks
				trimmedContent := strings.TrimSpace(styleSection.Content)
				if trimmedContent == "" {
					continue
				}

				// Calculate SHA256 hash for deduplication
				hash := fmt.Sprintf("%x", sha256.Sum256([]byte(trimmedContent)))

				// Only add if not already present (deduplication)
				if _, exists := styleBlocks[hash]; !exists {
					styleBlocks[hash] = &StyleBlock{
						Content: trimmedContent,
						Source:  name,
						Hash:    hash,
					}
					orderedHashes = append(orderedHashes, hash)
				}
			}
		}
	}

	// Start collection from root template
	collectStyles(rootTemplate, componentName)

	// Build aggregated CSS with source comments
	var result strings.Builder

	for _, hash := range orderedHashes {
		block := styleBlocks[hash]

		// Add source comment
		result.WriteString(fmt.Sprintf("/* Styles from: %s */\n", block.Source))

		// Add CSS content
		result.WriteString(block.Content)
		result.WriteString("\n\n")
	}

	return result.String()
}

// GetAggregatedStyles returns aggregated styles for a component, using cache when possible.
//
// On first call for a component, performs full aggregation and caches result.
// Subsequent calls return cached result for significant performance improvement.
//
// Thread-safe for concurrent access.
//
// Example:
//   styles := GetAggregatedStyles(template, "Header")
//   // First call: cache miss, performs aggregation
//   styles2 := GetAggregatedStyles(template, "Header")
//   // Second call: cache hit, returns cached result instantly
func GetAggregatedStyles(template *ast.Template, componentName string) string {
	if !cacheEnabled {
		return AggregateComponentStyles(template, componentName)
	}

	// Try cache lookup (read lock for concurrent reads)
	styleCacheMutex.RLock()
	cached, exists := componentStyleCache[componentName]
	styleCacheMutex.RUnlock()

	if exists {
		log.Printf("[Style Cache] HIT for component: %s", componentName)
		return cached
	}

	// Cache miss - perform aggregation
	log.Printf("[Style Cache] MISS for component: %s - aggregating...", componentName)
	aggregated := AggregateComponentStyles(template, componentName)

	// Store in cache (write lock for exclusive write)
	styleCacheMutex.Lock()
	componentStyleCache[componentName] = aggregated
	styleCacheMutex.Unlock()

	return aggregated
}

// ClearStyleCache clears all cached component styles.
//
// Use this in development mode when components are modified and re-registered,
// or when you need to force re-aggregation of styles.
//
// Thread-safe for concurrent access.
//
// Example:
//   ClearStyleCache()
//   log.Println("Cache cleared - next render will re-aggregate styles")
func ClearStyleCache() {
	styleCacheMutex.Lock()
	defer styleCacheMutex.Unlock()

	count := len(componentStyleCache)
	componentStyleCache = make(map[string]string)

	log.Printf("[Style Cache] CLEARED - removed %d cached entries", count)
}

// GetCacheStats returns cache statistics for debugging
//
// Returns a map with:
//   - cached_components: number of components in cache
//   - component_names: list of cached component names
//
// Example:
//   stats := GetCacheStats()
//   fmt.Printf("Cache contains %d components\n", stats["cached_components"])
func GetCacheStats() map[string]interface{} {
	styleCacheMutex.RLock()
	defer styleCacheMutex.RUnlock()

	return map[string]interface{}{
		"cached_components": len(componentStyleCache),
		"component_names":   getKeys(componentStyleCache),
	}
}

// getKeys extracts map keys into a slice
func getKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
