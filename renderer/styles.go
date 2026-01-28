package renderer

import (
	"crypto/sha256"
	"fmt"
	"log"
	"sort"
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
// Example: "./components/userprofile.html" -> "Userprofile" (capitalized to match registration)
// Example: "{path}" -> "" (skip dynamic runtime paths)
//
// CRITICAL FIX: Must capitalize first letter to match component registration convention
// Components are registered with capitalized names (e.g., "userprofile.html" -> "Userprofile")
// This function must return the same capitalization for lookups to work.
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
	baseName := filename
	if idx := strings.LastIndex(filename, "."); idx != -1 {
		baseName = filename[:idx]
	}

	// CRITICAL FIX: Capitalize first letter to match registration
	// This matches the logic in cmd/server/main.go registerComponentsFromDir()
	if len(baseName) == 0 {
		return ""
	}

	// Capitalize first letter: "userprofile" -> "Userprofile"
	capitalizedName := strings.ToUpper(baseName[:1]) + baseName[1:]

	log.Printf("[extractComponentNameFromPath] Extracted component name: %q from path: %q", capitalizedName, path)

	return capitalizedName
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

		case *ast.DynamicComponentByNameNode:
			// BUGFIX: Handle <Component:dynamic name={var} /> nodes
			// These appear in the ORIGINAL AST before transformation
			// We can't resolve the name here (it's a variable), but we log it
			log.Printf("[findComponentNodes] Found DynamicComponentByNameNode with nameExpr=%q (will be resolved in transformed AST)",
				n.NameExpression)
			// Don't add to componentNames - will be picked up from transformed AST

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

// AggregateComponentStylesWithTransformed aggregates styles from both original and transformed ASTs
//
// This function solves the "missing component CSS" problem by collecting components from:
// 1. Original AST - for FenceSection imports (Hero2436, Services2437 in _index.html)
// 2. Transformed AST - for dynamically resolved components (Component:dynamic → _index)
// 3. Dynamic layout - when Component:dynamic resolves to a layout, collect that layout's imports
// 4. JSON components - component names specified in JSON content (e.g., content/pages/_index.json)
//
// Pattern: Style Aggregation Pattern [Load: 24]
// Cognitive Load: 24 (traverse original: 5, traverse transformed: 5, dynamic layout: 4, json components: 3, merge: 3, dedupe: 4)
//
// Example:
//
//	html.html (original) imports: Nav, Head, Footer
//	html.html (transformed) has: Nav, Head, Footer, + resolved _index component
//	dynamicLayoutName: "_index"
//	jsonComponentNames: ["hero2436", "services2437", "whyChoose2425"]
//	Result: Styles from Nav, Head, Footer, Hero2436, Services2437, WhyChoose2425, _index
func AggregateComponentStylesWithTransformed(originalAST *ast.Template, transformedAST *ast.Template, componentName string, dynamicLayoutName string, jsonComponentNames []string) string {
	// Handle nil template gracefully
	if originalAST == nil && transformedAST == nil {
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
		// This includes both FenceSection.Imports AND ComponentNode usage
		var allComponentNames []string

		// Find components used in the template body (ComponentNode)
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

		log.Printf("[collectStyles] Component '%s' references: %v", name, uniqueComponentNames)

		// Process all discovered components recursively (dependencies first)
		for _, compName := range uniqueComponentNames {
			if compTemplate, exists := transformer.GetComponentTemplate(compName); exists {
				log.Printf("[collectStyles] Found component template for '%s', collecting styles recursively", compName)
				// Recursively collect styles from dependency
				collectStyles(compTemplate.AST, compName)
			} else {
				log.Printf("[collectStyles] WARNING: Component '%s' referenced but not found in registry", compName)
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
					log.Printf("[collectStyles] Added style block from '%s' (%d bytes, hash: %s...)", name, len(trimmedContent), hash[:8])
				}
			}
		}
	}

	// CRITICAL FIX: Collect from BOTH original and transformed ASTs
	// Step 1: Collect from original AST (gets FenceSection imports)
	log.Printf("[AggregateComponentStyles] Phase 1: Collecting from original AST for: %s", componentName)
	collectStyles(originalAST, componentName)

	// Step 2: Collect from transformed AST (gets dynamically resolved components)
	if transformedAST != nil {
		log.Printf("[AggregateComponentStyles] Phase 2: Collecting from transformed AST for: %s", componentName)
		// Find additional components in transformed AST (like resolved Component:dynamic)
		transformedComponents := findComponentNodes(transformedAST.RootNodes)
		log.Printf("[AggregateComponentStyles] Found %d components in transformed AST: %v",
			len(transformedComponents), transformedComponents)

		// Collect styles from these components
		for _, compName := range transformedComponents {
			if compTemplate, exists := transformer.GetComponentTemplate(compName); exists {
				collectStyles(compTemplate.AST, compName)
			}
		}
	}

	// Step 3: NEW FIX - If a dynamic layout was resolved, collect its imports too
	// This handles the case where Component:dynamic → _index, and _index imports Hero2436, Services2437
	if dynamicLayoutName != "" {
		log.Printf("[AggregateComponentStyles] Phase 3: Collecting from dynamic layout: %s", dynamicLayoutName)

		// Get the layout's component template
		if layoutTemplate, exists := transformer.GetComponentTemplate(dynamicLayoutName); exists {
			log.Printf("[AggregateComponentStyles] Found dynamic layout template: %s", dynamicLayoutName)

			// Collect styles from the layout itself and its imports
			collectStyles(layoutTemplate.AST, dynamicLayoutName)
		} else {
			log.Printf("[AggregateComponentStyles] Warning: dynamic layout %s not found in component registry", dynamicLayoutName)
		}
	}

	// Step 4: Collect styles from JSON-specified components
	// This handles the case where components are specified in content JSON files (e.g., content/pages/_index.json)
	// These components are NOT imported in the template but are rendered via {for component in content.components}
	if len(jsonComponentNames) > 0 {
		log.Printf("[AggregateComponentStyles] Phase 4: Collecting from %d JSON-specified components: %v", len(jsonComponentNames), jsonComponentNames)

		for _, compName := range jsonComponentNames {
			// Capitalize first letter to match component registration convention
			// Components are registered with capitalized names (e.g., "hero2436" → "Hero2436")
			capitalizedName := compName
			if len(compName) > 0 {
				capitalizedName = strings.ToUpper(compName[:1]) + compName[1:]
			}

			if compTemplate, exists := transformer.GetComponentTemplate(capitalizedName); exists {
				log.Printf("[AggregateComponentStyles] Found JSON component template: %s", capitalizedName)
				collectStyles(compTemplate.AST, capitalizedName)
			} else {
				log.Printf("[AggregateComponentStyles] Warning: JSON component %s (capitalized: %s) not found in component registry", compName, capitalizedName)
			}
		}
	}

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

	finalResult := result.String()
	log.Printf("[AggregateComponentStyles] Final aggregated CSS: %d bytes from %d unique style blocks", len(finalResult), len(orderedHashes))

	return finalResult
}

// AggregateComponentStyles traverses the component tree and aggregates styles
//
// DEPRECATED: Use AggregateComponentStylesWithTransformed instead
// This function only looks at the original AST and misses dynamically resolved components
//
// Algorithm:
// 1. Use depth-first traversal to collect styles (dependencies first)
// 2. Track visited components to prevent infinite loops (circular dependency protection)
// 3. Deduplicate based on content hash (SHA256)
// 4. Return aggregated CSS string with source comments
//
// Example:
//
//	template := &ast.Template{
//	  RootNodes: []ast.Node{
//	    &ast.StyleSection{Content: ".header { color: blue; }"},
//	  },
//	}
//	result := AggregateComponentStyles(template, "Header")
//	// Returns:
//	// /* Styles from: Header */
//	// .header { color: blue; }
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
				collectStyles(compTemplate.AST, compName)
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
// UPDATED: Now accepts both original and transformed ASTs to handle dynamic components
// UPDATED: Now accepts dynamicLayoutName to collect styles from dynamically resolved layouts
// UPDATED: Now accepts jsonComponentNames to collect styles from JSON-specified components
//
// On first call for a component, performs full aggregation and caches result.
// Subsequent calls return cached result for significant performance improvement.
//
// Thread-safe for concurrent access.
//
// Example:
//
//	styles := GetAggregatedStyles(originalAST, transformedAST, "Header", "_index", []string{"hero2436", "services2437"})
//	// First call: cache miss, performs aggregation including _index layout and JSON component imports
//	styles2 := GetAggregatedStyles(originalAST, transformedAST, "Header", "_index", []string{"hero2436", "services2437"})
//	// Second call: cache hit, returns cached result instantly
func GetAggregatedStyles(originalAST *ast.Template, transformedAST *ast.Template, componentName string, dynamicLayoutName string, jsonComponentNames []string) string {
	// Build cache key that includes dynamic layout name and JSON component names
	cacheKey := componentName
	if dynamicLayoutName != "" {
		cacheKey = fmt.Sprintf("%s:layout=%s", componentName, dynamicLayoutName)
	}
	// Include JSON component names in cache key for proper invalidation
	if len(jsonComponentNames) > 0 {
		sortedNames := make([]string, len(jsonComponentNames))
		copy(sortedNames, jsonComponentNames)
		sort.Strings(sortedNames)
		cacheKey = fmt.Sprintf("%s:json=%s", cacheKey, strings.Join(sortedNames, ","))
	}

	if !cacheEnabled {
		return AggregateComponentStylesWithTransformed(originalAST, transformedAST, componentName, dynamicLayoutName, jsonComponentNames)
	}

	// Try cache lookup (read lock for concurrent reads)
	styleCacheMutex.RLock()
	cached, exists := componentStyleCache[cacheKey]
	styleCacheMutex.RUnlock()

	if exists {
		log.Printf("[Style Cache] HIT for component: %s (dynamic layout: %s, json components: %v)", componentName, dynamicLayoutName, jsonComponentNames)
		return cached
	}

	// Cache miss - perform aggregation
	log.Printf("[Style Cache] MISS for component: %s (dynamic layout: %s, json components: %v) - aggregating...", componentName, dynamicLayoutName, jsonComponentNames)
	aggregated := AggregateComponentStylesWithTransformed(originalAST, transformedAST, componentName, dynamicLayoutName, jsonComponentNames)

	// Store in cache (write lock for exclusive write)
	styleCacheMutex.Lock()
	componentStyleCache[cacheKey] = aggregated
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
//
//	ClearStyleCache()
//	log.Println("Cache cleared - next render will re-aggregate styles")
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
//
//	stats := GetCacheStats()
//	fmt.Printf("Cache contains %d components\n", stats["cached_components"])
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
