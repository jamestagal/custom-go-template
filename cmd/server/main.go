package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	// Import the new renderer package
	"github.com/jimafisk/custom_go_template/ast"
	"github.com/jimafisk/custom_go_template/loader"
	"github.com/jimafisk/custom_go_template/parser"
	"github.com/jimafisk/custom_go_template/renderer"
	"github.com/jimafisk/custom_go_template/transformer"
)

// Global store registry loaded at startup
// Pattern: Package-level State [Load: 2]
var storeRegistry map[string]string

// TASK 4.4: Content cache for performance
// Pattern: In-Memory Cache [Load: 5]
var (
	contentCache   = make(map[string]map[string]interface{})
	contentCacheMu sync.RWMutex
)

// TASK 5.3: Cache for getAllContent to avoid repeated directory walks
var (
	allContentCache   map[string]interface{}
	allContentCacheMu sync.RWMutex
	allContentCached  bool
)

func main() {
	log.Println("Starting server...")

	// Create asset directories if they don't exist
	assetDirs := []string{"./scripts", "./styles", "./images", "./public"}
	for _, dir := range assetDirs {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			log.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}
	log.Println("Asset directories initialized")

	// Register stores FIRST (before components need them)
	storeRegistry = registerStores()
	log.Printf("Registered %d store(s)", len(storeRegistry))

	// Register components (now with store registry available)
	registerComponents(storeRegistry)

	// Dynamically register routes for all content layouts
	registerContentRoutes()

	// Set up the HTTP server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Serve static files from organized directories
		if r.URL.Path != "/" {
			serveStaticFile(w, r)
			return
		}

		// TEST: Use pages.html layout to test dynamic component iteration
		if err := renderWithWrapper("pages", w, r); err != nil {
			log.Printf("Error rendering home page: %v", err)
			http.Error(w, "Failed to render page", http.StatusInternalServerError)
		}
	})

	// Start the server
	port := ":3333"
	fmt.Printf("Server starting on http://localhost%s\n", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// serveStaticFile handles serving static files from organized asset directories
// Routes: /scripts/* → ./scripts/, /styles/* → ./styles/, /images/* → ./images/, /static/* → ./static/, /* → ./public/
func serveStaticFile(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	var filePath string

	// Route to appropriate directory based on path prefix
	switch {
	case strings.HasPrefix(path, "/scripts/"):
		filePath = "." + path
	case strings.HasPrefix(path, "/styles/"):
		filePath = "." + path
	case strings.HasPrefix(path, "/images/"):
		filePath = "." + path
	case strings.HasPrefix(path, "/static/"):
		filePath = "." + path
	default:
		// Everything else goes to public directory
		filePath = "./public" + path
	}

	// Serve the file
	http.ServeFile(w, r, filePath)
}

// loadContentWithCache loads content JSON with caching for performance
// TASK 4.4: Content caching implementation
//
// Pattern: Cache-Aside Pattern [Load: 12]
// Cognitive Load: 12 (cache lookup: 3, load on miss: 3, cache update: 3, error handling: 3)
func loadContentWithCache(routePath string) (map[string]interface{}, error) {
	// Check cache first (read lock for concurrent access)
	contentCacheMu.RLock()
	cached, exists := contentCache[routePath]
	contentCacheMu.RUnlock()

	if exists {
		log.Printf("loadContentWithCache: cache hit for %s", routePath)
		return cached, nil
	}

	// Cache miss - load from file
	log.Printf("loadContentWithCache: cache miss for %s, loading from file", routePath)
	contentData, err := loader.LoadContentForRoute(routePath)
	if err != nil {
		return nil, fmt.Errorf("loadContentWithCache: %w", err)
	}

	// Update cache (write lock)
	contentCacheMu.Lock()
	contentCache[routePath] = contentData
	contentCacheMu.Unlock()

	return contentData, nil
}

// invalidateContentCache clears the content cache (useful for development)
// Can be called via HTTP endpoint or during file watching
func invalidateContentCache() {
	contentCacheMu.Lock()
	defer contentCacheMu.Unlock()

	contentCache = make(map[string]map[string]interface{})

	// Also invalidate allContent cache
	allContentCacheMu.Lock()
	allContentCached = false
	allContentCache = nil
	allContentCacheMu.Unlock()

	log.Println("Content cache invalidated")
}

// TASK 5.3: getAllContent loads all content JSON files from content/ directory
// Returns map indexed by relative path: "pages/_index", "blog/post-1", etc.
//
// Pattern: File Discovery Pattern with Caching [Load: 15]
// Cognitive Load: 15 (directory walk: 5, file reading: 3, JSON parsing: 3, path formatting: 2, caching: 2)
func getAllContent() map[string]interface{} {
	// Check cache first
	allContentCacheMu.RLock()
	if allContentCached {
		cached := allContentCache
		allContentCacheMu.RUnlock()
		return cached
	}
	allContentCacheMu.RUnlock()

	// Build fresh cache
	result := make(map[string]interface{})
	contentDir := "content"

	// Walk content directory recursively (COGNITIVE LOAD RULE: wrapped error)
	err := filepath.Walk(contentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// If content directory doesn't exist, just return empty map
			return nil
		}

		// Skip directories and non-JSON files
		if info.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}

		// Load and parse JSON (COGNITIVE LOAD RULE: wrapped error)
		data, err := os.ReadFile(path)
		if err != nil {
			log.Printf("getAllContent: Warning: Failed to read %s: %v", path, err)
			return nil // Continue walking
		}

		var content map[string]interface{}
		if err := json.Unmarshal(data, &content); err != nil {
			log.Printf("getAllContent: Warning: Invalid JSON in %s: %v", path, err)
			return nil // Continue walking
		}

		// Calculate relative key: "content/pages/about.json" -> "pages/about"
		relPath := strings.TrimPrefix(path, contentDir+string(filepath.Separator))
		relPath = strings.TrimSuffix(relPath, ".json")
		// Normalize path separators for cross-platform compatibility
		relPath = filepath.ToSlash(relPath)

		result[relPath] = content
		return nil
	})

	if err != nil {
		log.Printf("getAllContent: Warning: Error walking content directory: %v", err)
	}

	// Update cache
	allContentCacheMu.Lock()
	allContentCache = result
	allContentCached = true
	allContentCacheMu.Unlock()

	log.Printf("getAllContent: Loaded %d content files", len(result))
	return result
}

// renderWithWrapper renders a page with the html.html wrapper (Nav + Content + Footer)
// This is the new unified rendering function that wraps all pages.
//
// UPDATED: Now extracts components array as top-level prop for pages.html iteration
//
// Pattern: Wrapper Pattern with Props Injection [Load: 20]
// Cognitive Load: 20 (content loading: 3, allContent: 3, allLayouts: 3, props building: 5, template rendering: 6)
func renderWithWrapper(layoutName string, w http.ResponseWriter, r *http.Request) error {
	log.Printf("[renderWithWrapper] Starting wrapper render for layout: %s, route: %s", layoutName, r.URL.Path)

	// Step 1: Load content for this route (COGNITIVE LOAD RULE: wrapped error)
	routePath := r.URL.Path
	contentData, err := loadContentWithCache(routePath)
	if err != nil {
		// Content loading failure is not fatal - log warning and continue
		log.Printf("[renderWithWrapper] Warning: failed to load content for route %s: %v", routePath, err)
		contentData = make(map[string]interface{}) // Empty content
	} else if len(contentData) > 0 {
		log.Printf("[renderWithWrapper] Loaded content for route %s: %d top-level keys", routePath, len(contentData))
	}

	// Step 2: Generate allContent list (all available content files)
	allContent := getAllContent()
	log.Printf("[renderWithWrapper] Generated allContent: %d files", len(allContent))

	// Step 3: Generate allLayouts list (all available layout components)
	allLayoutsMap := transformer.GetAllComponentNames()
	allLayouts := make([]string, 0, len(allLayoutsMap))
	for name := range allLayoutsMap {
		allLayouts = append(allLayouts, name)
	}
	sort.Strings(allLayouts) // Consistent ordering
	log.Printf("[renderWithWrapper] Generated allLayouts: %d components", len(allLayouts))

	// Step 4: Extract content.fields for wrapper to pass to dynamic component
	// This handles the Plenti collection type structure
	var contentFields map[string]interface{}
	if loader.IsCollectionType(contentData) {
		// For collection types, extract fields from first component
		if componentsRaw, ok := contentData["components"]; ok {
			if components, ok := componentsRaw.([]interface{}); ok && len(components) > 0 {
				if firstComp, ok := components[0].(map[string]interface{}); ok {
					if fields, ok := firstComp["fields"].(map[string]interface{}); ok {
						contentFields = fields
						log.Printf("[renderWithWrapper] Extracted fields from first component: %d fields", len(fields))
					}
				}
			}
		}
	} else {
		// For single type, use all content as fields
		contentFields = contentData
		log.Printf("[renderWithWrapper] Using full content as fields: %d keys", len(contentFields))
	}

	// If no fields extracted, use empty map
	if contentFields == nil {
		contentFields = make(map[string]interface{})
	}

	// Step 5: Build props map for wrapper
	// IMPORTANT: These are offered to the wrapper, but wrapper's export let controls which are actually used
	// The opt-in filtering happens inside renderTemplateWithProps
	props := map[string]interface{}{
		"layout":        layoutName, // Name of the layout to render (e.g., "_index")
		"content":       contentData, // Full content object
		"allContent":    allContent,  // All site content
		"allLayouts":    allLayouts,  // All available layouts
		"env":           make(map[string]interface{}), // Environment vars (TODO: populate if needed)
		"user":          make(map[string]interface{}), // User data (TODO: populate if needed)
		"shadowContent": make(map[string]interface{}), // Shadow content (TODO: populate if needed)
	}

	// Step 5.1: Extract components array as top-level prop
	// This is needed for pages.html which uses: {for component in components}
	if componentsRaw, ok := contentData["components"]; ok {
		props["components"] = componentsRaw
		log.Printf("[renderWithWrapper] Extracted components array as top-level prop")
	}

	// Add content.fields as a separate prop for easier access
	if len(contentFields) > 0 {
		// Create a nested content structure with fields
		contentWithFields := map[string]interface{}{
			"fields": contentFields,
		}
		// Merge in other top-level content keys (like components)
		for key, val := range contentData {
			if key != "fields" {
				contentWithFields[key] = val
			}
		}
		props["content"] = contentWithFields
	}

	// Step 5.5: TEMPORARY WORKAROUND - Extract first component fields as top-level props
	// This allows _index.html to receive Hero2436 props directly until full Component:dynamic iteration is implemented
	// TODO: Remove this workaround when implementing .agent-os/specs/2025-10-12-dynamic-component-iteration/
	if components, ok := contentData["components"].([]interface{}); ok && len(components) > 0 {
		if firstComp, ok := components[0].(map[string]interface{}); ok {
			if fields, ok := firstComp["fields"].(map[string]interface{}); ok {
				// Add each field as a top-level prop
				for key, value := range fields {
					props[key] = value
				}
				log.Printf("[renderWithWrapper] TEMPORARY WORKAROUND: Added %d fields from first component as top-level props", len(fields))
			}
		}
	}

	log.Printf("[renderWithWrapper] Built props map with %d keys (offered to wrapper)", len(props))
	log.Printf("[renderWithWrapper] Props keys: %v", getKeys(props))

	// Step 6: Call renderTemplate with html.html wrapper and props
	// renderTemplateWithProps will filter these based on wrapper's export let declarations
	wrapperPath := "layouts/global/html.html"
	log.Printf("[renderWithWrapper] Rendering wrapper template: %s", wrapperPath)

	// Use renderTemplateWithProps (with opt-in filtering)
	err = renderTemplateWithProps(wrapperPath, props, w, r)
	if err != nil {
		return fmt.Errorf("renderWithWrapper: failed to render wrapper: %w", err)
	}

	log.Printf("[renderWithWrapper] Successfully rendered wrapper for layout: %s", layoutName)
	return nil
}

// renderTemplateWithProps renders a template with explicitly provided props
// This is a variant of renderTemplate that accepts pre-built props instead of extracting them from fence
//
// UPDATED: Now implements magic variables opt-in system (commit 760ccb1)
// Only props declared in fence's "export let" will be added to x-data scope
//
// Pattern: Template Rendering with Props Injection [Load: 22]
// Cognitive Load: 22 (read: 2, parse: 3, props merge: 4, fence processing: 3, transform: 3, store merge: 3, render: 2, inject: 2)
func renderTemplateWithProps(entrypoint string, explicitProps map[string]interface{}, w http.ResponseWriter, r *http.Request) error {
	startTime := time.Now()

	// Read template file (COGNITIVE LOAD RULE: wrapped error)
	templateContent, err := os.ReadFile(entrypoint)
	if err != nil {
		return fmt.Errorf("renderTemplateWithProps: failed to read template %s: %w", entrypoint, err)
	}

	// Parse template to extract fence data (COGNITIVE LOAD RULE: wrapped error)
	template, err := parser.ParseTemplate(string(templateContent))
	if err != nil {
		return fmt.Errorf("renderTemplateWithProps: failed to parse template %s: %w", entrypoint, err)
	}

	// Extract fence section and parse with store registry if needed
	var fenceWithStores *ast.FenceSection
	for i, node := range template.RootNodes {
		if fence, ok := node.(*ast.FenceSection); ok {
			// Only re-parse if fence contains store imports
			if strings.Contains(fence.RawContent, "import store from") {
				// Parse fence content with store registry to resolve store imports
				fenceWithStores = parser.ParseFenceContentWithStores(fence.RawContent, storeRegistry)
				// Replace the fence section in template
				template.RootNodes[i] = fenceWithStores
			} else {
				// No store imports, use the already-parsed fence as-is
				fenceWithStores = fence
				template.RootNodes[i] = fenceWithStores
			}
			break
		}
	}

	// Build exportedPropNames map for opt-in filtering (Magic Variables Opt-In System)
	exportedPropNames := make(map[string]bool)
	if fenceWithStores != nil {
		for _, propName := range fenceWithStores.ExportedProps {
			exportedPropNames[propName] = true
		}
		log.Printf("[renderTemplateWithProps] Template %s exports: %v", entrypoint, fenceWithStores.ExportedProps)
	}

	// Filter explicit props based on export let declarations (OPT-IN ONLY)
	// This prevents data bloat from unused magic variables
	// IMPORTANT: Props from content.fields are allowed to pass through even if not in wrapper's export let
	// because they're meant for child components (via {...content.fields} spread)
	props := make(map[string]interface{})

	// Check if any props are from content.fields (for passthrough to children)
	contentFieldsProps := make(map[string]bool)
	if contentProp, ok := explicitProps["content"].(map[string]interface{}); ok {
		if fields, ok := contentProp["fields"].(map[string]interface{}); ok {
			for fieldName := range fields {
				contentFieldsProps[fieldName] = true
			}
		}
	}

	for k, v := range explicitProps {
		// Allow prop if:
		// 1. It's declared in export let, OR
		// 2. It's from content.fields (passthrough for child components)
		if exportedPropNames[k] {
			props[k] = v
			log.Printf("[renderTemplateWithProps] Added prop '%s' (declared in export let)", k)
		} else if contentFieldsProps[k] {
			props[k] = v
			log.Printf("[renderTemplateWithProps] Added prop '%s' (from content.fields, passthrough to children)", k)
		} else {
			log.Printf("[renderTemplateWithProps] Skipped prop '%s' (not in export let or content.fields)", k)
		}
	}

	// Extract props from fence section (as defaults if not in explicitProps)
	if fenceWithStores != nil {
		// Process variables
		for _, variable := range fenceWithStores.Variables {
			if _, exists := props[variable.Name]; !exists {
				props[variable.Name] = variable.Value
			}
		}

		// Process props with default values
		for _, prop := range fenceWithStores.Props {
			if _, exists := props[prop.Name]; !exists && prop.DefaultValue != "" {
				props[prop.Name] = parseValue(prop.DefaultValue)
			}
		}

		// Extract functions from FenceSection.Functions field
		for _, function := range fenceWithStores.Functions {
			if _, exists := props[function.Name]; !exists {
				props[function.Name] = function.Body
			}
		}
	}

	// Add build time as a prop
	buildTime := time.Since(startTime)
	buildTimeMs := float64(buildTime.Microseconds()) / 1000.0
	props["buildTime"] = fmt.Sprintf("%.2fms", buildTimeMs)

	log.Printf("[renderTemplateWithProps] Final props for %s: %v", entrypoint, getKeys(props))

	// Transform template (this tracks store references)
	transformed := transformer.TransformAST(template, props)

	// Get tracked stores from transformer
	referencedStores, allDefinitions := transformer.GetTrackedStores(transformed)
	log.Printf("[renderTemplateWithProps] Referenced stores: %v", referencedStores)

	referencedStoreDefs := transformer.GetReferencedStoreDefinitions(allDefinitions, referencedStores)

	// Merge with external stores if referenced but not defined
	// Priority: Inline > Imported > External
	finalStores := make(map[string]string)

	// Add all referenced stores (inline + imported from fence)
	for name, def := range referencedStoreDefs {
		finalStores[name] = def
	}

	// Add external stores if referenced but not yet in finalStores
	for _, storeName := range referencedStores {
		if _, exists := finalStores[storeName]; !exists {
			if externalDef, exists := storeRegistry[storeName]; exists {
				finalStores[storeName] = externalDef
				log.Printf("[renderTemplateWithProps] Added external store: %s", storeName)
			}
		}
	}

	// Render with stores
	// CRITICAL: Pass original template AST and path for component style aggregation
	// Extract layout name from props for CSS aggregation
	layoutName := ""
	if layoutProp, ok := props["layout"].(string); ok {
		layoutName = layoutProp
	}

	markup, script, style := renderer.RenderWithStores(template, transformed, finalStores, entrypoint, layoutName)

	// Build x-data from props
	xDataValue := buildXDataFromProps(props)

	// Build final HTML with x-data injected
	finalHTML := markup

	// Inject x-data into body tag (or html tag as fallback)
	bodyTagRegex := regexp.MustCompile(`(?i)<body[^>]*>`)
	if bodyTagRegex.MatchString(finalHTML) {
		finalHTML = bodyTagRegex.ReplaceAllStringFunc(finalHTML, func(match string) string {
			// Check if x-data already exists
			if strings.Contains(match, "x-data") {
				return match
			}
			// Remove the closing > and add x-data
			tagWithoutClose := strings.TrimSuffix(match, ">")
			return fmt.Sprintf(`%s x-data="%s">`, tagWithoutClose, escapeXDataForAttr(xDataValue))
		})
	} else {
		// Fallback: add x-data to html tag if no body tag
		htmlTagRegex := regexp.MustCompile(`(?i)<html[^>]*>`)
		if htmlTagRegex.MatchString(finalHTML) {
			finalHTML = htmlTagRegex.ReplaceAllStringFunc(finalHTML, func(match string) string {
				if strings.Contains(match, "x-data") {
					return match
				}
				tagWithoutClose := strings.TrimSuffix(match, ">")
				return fmt.Sprintf(`%s x-data="%s">`, tagWithoutClose, escapeXDataForAttr(xDataValue))
			})
		}
	}

	// Inject styles into <head>
	if style != "" {
		headEndRegex := regexp.MustCompile(`(?i)</head>`)
		finalHTML = headEndRegex.ReplaceAllString(finalHTML, fmt.Sprintf("<style>\n%s\n</style></head>", style))
	}

	// Inject scripts before </body>
	if script != "" {
		bodyEndRegex := regexp.MustCompile(`(?i)</body>`)
		finalHTML = bodyEndRegex.ReplaceAllString(finalHTML, fmt.Sprintf("<script>\n%s\n</script></body>", script))
	}

	// Add Alpine.js CDN if not already present
	if !strings.Contains(finalHTML, "alpinejs") {
		headEndRegex := regexp.MustCompile(`(?i)</head>`)
		finalHTML = headEndRegex.ReplaceAllString(finalHTML,
			`<script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script></head>`)
	}

	// Add build time comment
	totalBuildTime := time.Since(startTime)
	htmlComment := fmt.Sprintf("<!-- Build time: %v -->\n", totalBuildTime)
	finalHTML = htmlComment + finalHTML

	// Send response
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(finalHTML))

	return nil
}

// getKeys extracts keys from a map for logging
// Pattern: Helper Function [Load: 3]
func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// renderTemplate is a unified handler for rendering template files with store support
// Now integrates with the global store system (Task 3.5)
// UPDATED: Now supports content injection (Task 4)
// UPDATED: Now passes magic variables (Task 5.4)
//
// Pattern: Service Implementation Pattern with Store Integration [Load: 20]
// Cognitive Load: 20 (read: 2, parse: 3, fence parsing: 3, content loading: 3, transform: 3, store merge: 3, render: 2, inject: 1)
func renderTemplate(entrypoint string, w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Read template file (COGNITIVE LOAD RULE: wrapped error)
	templateContent, err := os.ReadFile(entrypoint)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read template: %v", err), http.StatusInternalServerError)
		return
	}

	// Parse template to extract fence data (COGNITIVE LOAD RULE: wrapped error)
	template, err := parser.ParseTemplate(string(templateContent))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse template: %v", err), http.StatusInternalServerError)
		return
	}

	// TASK 4.1 & 4.2: Load content JSON for this route
	// Extract route path from request URL
	routePath := r.URL.Path
	contentData, err := loadContentWithCache(routePath)
	if err != nil {
		// Content loading failure is not fatal - log warning and continue
		log.Printf("Warning: failed to load content for route %s: %v", routePath, err)
		contentData = nil // No content injection
	} else if len(contentData) > 0 {
		log.Printf("Loaded content for route %s: %d top-level keys", routePath, len(contentData))
	}

	// Extract fence section and parse with store registry ONLY if needed (Task 3.5 integration)
	var fenceWithStores *ast.FenceSection
	for i, node := range template.RootNodes {
		if fence, ok := node.(*ast.FenceSection); ok {
			// TASK 4: Inject content into exported props BEFORE store parsing
			if contentData != nil && len(fence.ExportedProps) > 0 {
				// Check if this is a collection type - extract component fields
				if loader.IsCollectionType(contentData) {
					// For collection types, we need to know which component we're rendering
					// For now, use a simple heuristic: first component in the array
					// TODO: In future, route could specify which component to use
					componentsRaw, ok := contentData["components"]
					if ok {
						if components, ok := componentsRaw.([]interface{}); ok && len(components) > 0 {
							if firstComp, ok := components[0].(map[string]interface{}); ok {
								if fields, ok := firstComp["fields"].(map[string]interface{}); ok {
									// Use the extracted fields for injection
									contentData = fields
									log.Printf("Extracted fields from first component for injection: %d fields", len(fields))
								}
							}
						}
					}
				}

				// Inject content props
				injectedFence, err := renderer.InjectContentProps(fence, contentData)
				if err != nil {
					log.Printf("Warning: content injection failed: %v", err)
					// Continue with original fence
					fenceWithStores = fence
				} else {
					fence = injectedFence
					log.Printf("Content injection successful: %d exported props injected", len(fence.ExportedProps))
				}
			}

			// Only re-parse if fence contains store imports
			if strings.Contains(fence.RawContent, "import store from") {
				// Parse fence content with store registry to resolve store imports
				fenceWithStores = parser.ParseFenceContentWithStores(fence.RawContent, storeRegistry)
				// Replace the fence section in template
				template.RootNodes[i] = fenceWithStores
			} else {
				// No store imports, use the already-parsed fence as-is (possibly with injected content)
				fenceWithStores = fence
				template.RootNodes[i] = fenceWithStores
			}
			break
		}
	}

	// Extract initial props from fence section (for buildTime)
	props := make(map[string]interface{})
	if fenceWithStores != nil {
		// Process variables
		for _, variable := range fenceWithStores.Variables {
			props[variable.Name] = variable.Value
		}

		// Process props with default values
		for _, prop := range fenceWithStores.Props {
			if _, exists := props[prop.Name]; !exists && prop.DefaultValue != "" {
				props[prop.Name] = parseValue(prop.DefaultValue)
			}
		}

		// UPDATED: Extract functions from FenceSection.Functions field (Task 2.3)
		// This replaces the manual regex-based extraction
		for _, function := range fenceWithStores.Functions {
			props[function.Name] = function.Body
		}
	}

	// TASK 5.4: Add magic variables to props (OPT-IN via export let)
	// Only add magic variables if the template explicitly declares them in export let
	exportedPropNames := make(map[string]bool)
	if fenceWithStores != nil {
		for _, propName := range fenceWithStores.ExportedProps {
			exportedPropNames[propName] = true
		}
	}

	if contentData != nil {
		// Magic variable 1: components array from content JSON
		if exportedPropNames["components"] {
			if componentsRaw, ok := contentData["components"]; ok {
				props["components"] = componentsRaw
				log.Printf("Magic variable 'components' added to props (requested via export let)")
			}
		}

		// Magic variable 2: content - full content object for this page
		if exportedPropNames["content"] {
			props["content"] = contentData
			log.Printf("Magic variable 'content' added to props (requested via export let)")
		}
	}

	// Magic variable 3: allContent - all site content (OPT-IN only)
	if exportedPropNames["allContent"] {
		props["allContent"] = getAllContent()
		log.Printf("Magic variable 'allContent' added to props (requested via export let)")
	}

	// Magic variable 4: allLayouts - component registry names (OPT-IN only)
	if exportedPropNames["allLayouts"] {
		// Convert to array of strings for proper JSON serialization
		layoutNames := make([]string, 0)
		for name := range transformer.GetAllComponentNames() {
			layoutNames = append(layoutNames, name)
		}
		props["allLayouts"] = layoutNames
		log.Printf("Magic variable 'allLayouts' added to props (%d components, requested via export let)", len(layoutNames))
	}

	// Add build time as a prop
	buildTime := time.Since(startTime)
	buildTimeMs := float64(buildTime.Microseconds()) / 1000.0
	props["buildTime"] = fmt.Sprintf("%.2fms", buildTimeMs)

	// Transform template (this tracks store references)
	transformed := transformer.TransformAST(template, props)

	// Get tracked stores from transformer (Task 3.5: Store merging)
	referencedStores, allDefinitions := transformer.GetTrackedStores(transformed)
	log.Printf("[DEBUG] Referenced stores: %v", referencedStores)

	allDefKeys := make([]string, 0, len(allDefinitions))
	for k := range allDefinitions {
		allDefKeys = append(allDefKeys, k)
	}
	log.Printf("[DEBUG] All definitions keys: %v", allDefKeys)

	referencedStoreDefs := transformer.GetReferencedStoreDefinitions(allDefinitions, referencedStores)

	refDefKeys := make([]string, 0, len(referencedStoreDefs))
	for k := range referencedStoreDefs {
		refDefKeys = append(refDefKeys, k)
	}
	log.Printf("[DEBUG] Referenced store defs keys: %v", refDefKeys)

	// Merge with external stores if referenced but not defined (Task 3.5: Priority system)
	// Priority: Inline > Imported > External
	finalStores := make(map[string]string)

	// Add all referenced stores (inline + imported from fence)
	for name, def := range referencedStoreDefs {
		finalStores[name] = def
	}

	// Add external stores if referenced but not yet in finalStores
	for _, storeName := range referencedStores {
		if _, exists := finalStores[storeName]; !exists {
			if externalDef, exists := storeRegistry[storeName]; exists {
				finalStores[storeName] = externalDef
				log.Printf("[renderTemplate] Added external store: %s", storeName)
			}
		}
	}

	finalKeys := make([]string, 0, len(finalStores))
	for k := range finalStores {
		finalKeys = append(finalKeys, k)
	}
	log.Printf("[DEBUG] Final stores keys: %v", finalKeys)

	// Render with stores (Task 3.5: Use RenderWithStores instead of Render)
	// CRITICAL: Pass original template AST and path for component style aggregation
	markup, script, style := renderer.RenderWithStores(template, transformed, finalStores, entrypoint, "")

	// CRITICAL: Generate x-data using transformer's alpineDataFormatter
	// This function is not exported, so we need to call Transform to get the data scope
	// and then format it ourselves
	xDataValue := buildXDataFromProps(props)

	// Build final HTML with x-data injected
	finalHTML := markup

	// Inject x-data into body tag (or html tag as fallback)
	bodyTagRegex := regexp.MustCompile(`(?i)<body[^>]*>`)
	if bodyTagRegex.MatchString(finalHTML) {
		finalHTML = bodyTagRegex.ReplaceAllStringFunc(finalHTML, func(match string) string {
			// Check if x-data already exists
			if strings.Contains(match, "x-data") {
				return match
			}
			// Remove the closing > and add x-data
			tagWithoutClose := strings.TrimSuffix(match, ">")
			return fmt.Sprintf(`%s x-data="%s">`, tagWithoutClose, escapeXDataForAttr(xDataValue))
		})
	} else {
		// Fallback: add x-data to html tag if no body tag
		htmlTagRegex := regexp.MustCompile(`(?i)<html[^>]*>`)
		if htmlTagRegex.MatchString(finalHTML) {
			finalHTML = htmlTagRegex.ReplaceAllStringFunc(finalHTML, func(match string) string {
				if strings.Contains(match, "x-data") {
					return match
				}
				tagWithoutClose := strings.TrimSuffix(match, ">")
				return fmt.Sprintf(`%s x-data="%s">`, tagWithoutClose, escapeXDataForAttr(xDataValue))
			})
		}
	}

	// Inject styles into <head>
	if style != "" {
		headEndRegex := regexp.MustCompile(`(?i)</head>`)
		finalHTML = headEndRegex.ReplaceAllString(finalHTML, fmt.Sprintf("<style>\n%s\n</style></head>", style))
	}

	// Inject scripts before </body>
	if script != "" {
		bodyEndRegex := regexp.MustCompile(`(?i)</body>`)
		finalHTML = bodyEndRegex.ReplaceAllString(finalHTML, fmt.Sprintf("<script>\n%s\n</script></body>", script))
	}

	// Add Alpine.js CDN if not already present
	if !strings.Contains(finalHTML, "alpinejs") {
		headEndRegex := regexp.MustCompile(`(?i)</head>`)
		finalHTML = headEndRegex.ReplaceAllString(finalHTML,
			`<script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script></head>`)
	}

	// Add build time comment
	totalBuildTime := time.Since(startTime)
	htmlComment := fmt.Sprintf("<!-- Build time: %v -->\n", totalBuildTime)
	finalHTML = htmlComment + finalHTML

	// Send response
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(finalHTML))
}

// buildXDataFromProps creates an Alpine.js x-data attribute value from props
// This builds a JavaScript object literal (NOT JSON) to support functions
//
// CRITICAL: Functions must NOT be quoted. Example output:
// {buildTime:'20ms',formatPrice:function formatPrice(price){return "$"+price.toFixed(2);}}
//
// Pattern: Data Formatting Pattern [Load: 12]
// Cognitive Load: 12 (iterate props: 2, detect functions: 3, format values: 5, join: 2)
func buildXDataFromProps(props map[string]interface{}) string {
	if len(props) == 0 {
		return "{}"
	}

	// COGNITIVE LOAD RULE: preallocate slice
	parts := make([]string, 0, len(props))

	// Sort keys for consistent output (easier debugging)
	keys := make([]string, 0, len(props))
	for key := range props {
		if !strings.HasPrefix(key, "$") {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := props[key]

		// Format value as JavaScript (NOT JSON)
		var formattedValue string
		switch v := value.(type) {
		case string:
			// Check if it's a function definition
			if strings.HasPrefix(v, "function ") || strings.Contains(v, "=>") {
				// Function - keep as-is, just minify whitespace for HTML attribute
				formattedValue = minifyFunction(v)
			} else {
				// Regular string - quote with single quotes for HTML attribute safety
				escaped := escapeStringForJS(v)
				formattedValue = fmt.Sprintf(`'%s'`, escaped)
			}
		case bool:
			formattedValue = fmt.Sprintf("%t", v)
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			formattedValue = fmt.Sprintf("%d", v)
		case float32, float64:
			formattedValue = fmt.Sprintf("%v", v)
		case nil:
			formattedValue = "null"
		default:
			// Complex types (arrays, objects) - use transformer's formatter
			// This correctly formats maps as JavaScript object literals, not JSON strings
			formattedValue = transformer.FormatGoValueToJS(v)
		}

		// Build key:value pair for JavaScript object literal
		parts = append(parts, fmt.Sprintf("%s:%s", key, formattedValue))
	}

	return "{" + strings.Join(parts, ",") + "}"
}

// minifyFunction removes unnecessary whitespace from function definitions
// to make them more compact for HTML attributes
func minifyFunction(fn string) string {
	// Remove leading/trailing whitespace
	fn = strings.TrimSpace(fn)

	// Replace multiple spaces with single space
	multiSpace := regexp.MustCompile(`\s+`)
	fn = multiSpace.ReplaceAllString(fn, " ")

	// Remove spaces around operators and braces (careful not to break keywords)
	fn = strings.ReplaceAll(fn, " {", "{")
	fn = strings.ReplaceAll(fn, "{ ", "{")
	fn = strings.ReplaceAll(fn, " }", "}")
	fn = strings.ReplaceAll(fn, "} ", "}")
	fn = strings.ReplaceAll(fn, " (", "(")
	fn = strings.ReplaceAll(fn, "( ", "(")
	fn = strings.ReplaceAll(fn, " )", ")")
	fn = strings.ReplaceAll(fn, ") ", ")")
	fn = strings.ReplaceAll(fn, "; ", ";")

	return fn
}

// escapeStringForJS escapes a string for use in JavaScript string literal
func escapeStringForJS(s string) string {
	// Escape backslashes first (must be first to avoid double-escaping)
	s = strings.ReplaceAll(s, `\`, `\\`)
	// Escape single quotes (we use single quotes for strings)
	s = strings.ReplaceAll(s, `'`, `\'`)
	// Escape newlines
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	// Escape tabs
	s = strings.ReplaceAll(s, "\t", `\t`)

	return s
}

// escapeXDataForAttr escapes an x-data value for use in HTML attribute
// CRITICAL: We need to escape characters that would break the attribute
//
// Pattern: String Escaping Pattern [Load: 4]
// Cognitive Load: 4 (replace operations: 4)
func escapeXDataForAttr(value string) string {
	// CRITICAL FIX: Do NOT HTML-escape JavaScript code in x-data!
	// The JavaScript object literal uses single quotes for strings,
	// so only the double quote needs escaping for the HTML attribute delimiter.
	//
	// Example: x-data="{ name: 'John & Jane' }"
	//   - The & stays as & (not &amp;) because it's inside JavaScript
	//   - Only " needs to become &quot; to not break the attribute
	//
	// WRONG: x-data="{ name: 'John &amp; Jane' }"  ← JavaScript syntax error!
	// RIGHT: x-data="{ name: 'John & Jane' }"      ← Valid JavaScript

	// Only escape double quotes (the HTML attribute delimiter)
	value = strings.ReplaceAll(value, `"`, `&quot;`)

	return value
}

// registerComponents scans the components directory and registers each component
// Now accepts storeRegistry to parse component fence sections with store imports
// Also registers global layout components from layouts/global/
// UPDATED: Now also registers content layouts from layouts/content/
//
// Pattern: File Discovery Pattern with Store Integration [Load: 15]
// Cognitive Load: 15 (read 3 dirs: 6, iterate: 2, read file: 2, parse: 2, fence parsing: 2, register: 3)
func registerComponents(storeRegistry map[string]string) {
	// Register regular components from layouts/components
	componentDir := "layouts/components"
	registerComponentsFromDir(componentDir, "../components/", storeRegistry)

	// Register global layout components from layouts/global
	globalDir := "layouts/global"
	registerComponentsFromDir(globalDir, "../global/", storeRegistry)

	// FIXED: Register content layouts from layouts/content (for _index, pages, etc.)
	contentDir := "layouts/content"
	registerComponentsFromDir(contentDir, "../content/", storeRegistry)
}

// registerComponentsFromDir registers all components from a directory
// Pattern: Component Registration Helper [Load: 12]
func registerComponentsFromDir(dir string, pathPrefix string, storeRegistry map[string]string) {
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("Warning: Failed to read directory %s: %v", dir, err)
		return
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".html") {
			// Extract base name and capitalize first letter (matches Plenti/Svelte convention)
			// e.g., "footer.html" -> "Footer", matching: import Footer from "./footer.html"
			baseName := strings.TrimSuffix(file.Name(), ".html")
			componentName := strings.ToUpper(baseName[:1]) + baseName[1:]
			componentPath := fmt.Sprintf("%s/%s", dir, file.Name())
			log.Printf("Registering component: %s from %s", componentName, componentPath)

			// Read component file
			componentContent, err := os.ReadFile(componentPath)
			if err != nil {
				log.Fatalf("Error reading component: %v", err)
			}

			// Parse the component template
			componentAST, err := parser.ParseTemplate(string(componentContent))
			if err != nil {
				log.Fatalf("Error parsing component: %v", err)
			}

			// TASK 2.2 FIX: Only re-parse fence if component has store imports
			// This preserves functions that were already parsed by ParseTemplate
			for i, node := range componentAST.RootNodes {
				if fence, ok := node.(*ast.FenceSection); ok {
					// Only re-parse if component has store imports
					if strings.Contains(fence.RawContent, "import store from") {
						// Parse fence content with store registry to resolve imports
						fenceWithStores := parser.ParseFenceContentWithStores(fence.RawContent, storeRegistry)
						// Replace the fence section in component AST
						componentAST.RootNodes[i] = fenceWithStores
						log.Printf("[registerComponents] Re-parsed fence with stores for %s (stores: %d, functions: %d)",
							componentName, len(fenceWithStores.Stores), len(fenceWithStores.Functions))
					} else {
						// No store imports - keep the already-parsed fence with functions intact
						log.Printf("[registerComponents] Preserved original fence for %s (functions: %d)",
							componentName, len(fence.Functions))
					}
					break
				}
			}

			// Extract props from the component template
			componentProps := extractComponentProps(componentAST)

			// Register the component with the transformer - both by name and by path
			transformer.RegisterComponent(componentName, componentAST, componentProps)

			// Also register with path prefix for import resolution (using lowercase filename)
			pathWithPrefix := fmt.Sprintf("%s%s", pathPrefix, file.Name())
			transformer.RegisterComponent(pathWithPrefix, componentAST, componentProps)
		}
	}
}

// registerStores scans the stores/ directory for .js files and loads them
// Returns a map of store name (filename without .js) to store content
//
// Pattern: File Discovery Pattern [Load: 8]
// Cognitive Load: 8 (read dir: 2, filter: 2, read files: 2, map building: 2)
func registerStores() map[string]string {
	stores := make(map[string]string)
	storeDir := "stores"

	// Read the stores directory (COGNITIVE LOAD RULE: wrapped error)
	files, err := os.ReadDir(storeDir)
	if err != nil {
		// Directory not existing is not an error - just log and return empty map
		log.Printf("Stores directory not found (this is OK): %s", storeDir)
		return stores
	}

	// Process each .js file in the directory
	for _, file := range files {
		// Skip directories and non-.js files
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".js") {
			continue
		}

		// Extract store name from filename (e.g., "auth.js" → "auth")
		storeName := strings.TrimSuffix(file.Name(), ".js")
		storePath := fmt.Sprintf("%s/%s", storeDir, file.Name())

		// Read store file content (COGNITIVE LOAD RULE: wrapped error)
		content, err := os.ReadFile(storePath)
		if err != nil {
			log.Printf("WARNING: Failed to read store file %s: %v", storePath, err)
			continue
		}

		// Store the content
		stores[storeName] = string(content)
		log.Printf("Registered store: %s from %s", storeName, storePath)
	}

	return stores
}

func extractComponentProps(template *ast.Template) []string {
	var props []string

	// Look for prop declarations in fence sections
	for _, node := range template.RootNodes {
		if fence, ok := node.(*ast.FenceSection); ok {
			// Extract props directly from the Props field
			for _, prop := range fence.Props {
				props = append(props, prop.Name)
			}

			// Also check the raw content for any props that might not have been parsed
			content := fence.RawContent
			// Look for lines starting with "prop"
			lines := strings.Split(content, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "prop ") {
					// Extract the prop name
					parts := strings.SplitN(line, "=", 2)
					if len(parts) > 0 {
						propDecl := strings.TrimSpace(parts[0])
						propName := strings.TrimPrefix(propDecl, "prop ")
						propName = strings.TrimSpace(propName)

						// Check if this prop is already in our list
						found := false
						for _, existingProp := range props {
							if existingProp == propName {
								found = true
								break
							}
						}

						// Add it if it's not already in the list
						if !found {
							props = append(props, propName)
						}
					}
				}
			}
			break
		}
	}

	return props
}


// convertJSToJSON converts JavaScript object syntax to valid JSON
// Handles unquoted keys: {name: "value"} → {"name": "value"}
func convertJSToJSON(js string) string {
	js = strings.TrimSpace(js)

	// Only convert if it looks like a JS object or array
	if !(strings.HasPrefix(js, "{") || strings.HasPrefix(js, "[")) {
		return js
	}

	// Simple regex to quote unquoted object keys
	// Matches: word characters followed by colon (but not inside quotes)
	// This is a simplified approach - for production use a proper JS parser
	re := regexp.MustCompile(`([{,]\s*)([a-zA-Z_$][a-zA-Z0-9_$]*)\s*:`)
	result := re.ReplaceAllString(js, `$1"$2":`)

	return result
}

func parseValue(value string) interface{} {
	value = strings.TrimSpace(value)

	// Handle empty values
	if value == "" {
		return ""
	}

	// CRITICAL FIX: Convert JavaScript object syntax to JSON before unmarshaling
	// JavaScript: {name: "value"} → JSON: {"name": "value"}
	// This allows fence section objects to be parsed as structured data
	jsonValue := convertJSToJSON(value)

	// Try to parse as JSON first (handles arrays, objects, numbers, booleans, null)
	var parsedValue interface{}
	if err := json.Unmarshal([]byte(jsonValue), &parsedValue); err == nil {
		// Successfully parsed as JSON
		return parsedValue
	}

	// If JSON parsing failed, try specific type conversions

	// Handle booleans
	if value == "true" {
		return true
	}
	if value == "false" {
		return false
	}

	// Handle null
	if value == "null" {
		return nil
	}

	// Handle numbers (integers)
	if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
		return intVal
	}

	// Handle numbers (floats)
	if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
		return floatVal
	}

	// Handle quoted strings - remove quotes
	if (strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`)) ||
		(strings.HasPrefix(value, `'`) && strings.HasSuffix(value, `'`)) {
		value = value[1 : len(value)-1]
	}

	// Default to string
	return value
}

// registerContentRoutes dynamically registers HTTP routes for all .html files in layouts/content/
// This eliminates the need to manually add routes for each page.
//
// Pattern: Dynamic Route Registration [Load: 12]
// Cognitive Load: 12 (directory read: 3, file filtering: 2, route creation: 4, logging: 3)
func registerContentRoutes() {
	contentDir := "layouts/content"

	// Read all files in the content directory
	files, err := os.ReadDir(contentDir)
	if err != nil {
		log.Printf("Warning: Failed to read content directory %s: %v", contentDir, err)
		return
	}

	routeCount := 0
	for _, file := range files {
		// Skip directories and non-HTML files
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".html") {
			continue
		}

		// Extract route name from filename (e.g., "store-demo.html" → "store-demo")
		routeName := strings.TrimSuffix(file.Name(), ".html")

		// Skip _index.html (handled separately by the root "/" route)
		if routeName == "_index" {
			continue
		}

		// Build file path
		filePath := filepath.Join(contentDir, file.Name())

		// Register route (capture filePath in closure)
		route := "/" + routeName
		currentFilePath := filePath // Capture for closure
		http.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
			renderTemplate(currentFilePath, w, r)
		})

		routeCount++
		log.Printf("Registered route: %s → %s", route, filePath)
	}

	log.Printf("Registered %d dynamic content routes", routeCount)
}
