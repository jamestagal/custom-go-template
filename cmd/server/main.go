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
	"github.com/jimafisk/custom_go_template/builder"
	"github.com/jimafisk/custom_go_template/loader"
	"github.com/jimafisk/custom_go_template/parser"
	"github.com/jimafisk/custom_go_template/renderer"
	"github.com/jimafisk/custom_go_template/transformer"
	"github.com/jimafisk/custom_go_template/types"
	"github.com/jimafisk/custom_go_template/utils"
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
// Changed to []interface{} to match Plenti's array format
var (
	allContentCache   []interface{}
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

	// Generate component registry for runtime components
	if err := generateComponentRegistry(); err != nil {
		log.Printf("WARNING: Failed to generate component registry: %v", err)
		log.Printf("Runtime component resolution may not work correctly")
	}

	// Generate content.js from content directory
	if err := generateContentJS(); err != nil {
		log.Printf("WARNING: Failed to generate content.js: %v", err)
	}

	// Register static file handlers (must be registered first to avoid conflicts)
	http.HandleFunc("/scripts/", func(w http.ResponseWriter, r *http.Request) {
		serveStaticFile(w, r)
	})
	http.HandleFunc("/styles/", func(w http.ResponseWriter, r *http.Request) {
		serveStaticFile(w, r)
	})
	http.HandleFunc("/images/", func(w http.ResponseWriter, r *http.Request) {
		serveStaticFile(w, r)
	})
	http.HandleFunc("/js/", func(w http.ResponseWriter, r *http.Request) {
		serveStaticFile(w, r)
	})
	http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		serveStaticFile(w, r)
	})
	http.HandleFunc("/public/", func(w http.ResponseWriter, r *http.Request) {
		serveStaticFile(w, r)
	})
	// Plenti-compatible core and generated directories
	http.HandleFunc("/core/", func(w http.ResponseWriter, r *http.Request) {
		serveCoreFile(w, r)
	})
	http.HandleFunc("/generated/", func(w http.ResponseWriter, r *http.Request) {
		serveGeneratedFile(w, r)
	})
	// Serve content JSON files for CMS
	http.HandleFunc("/content/", func(w http.ResponseWriter, r *http.Request) {
		serveContentFile(w, r)
	})
	log.Println("Registered static file handlers")

	// Register routes from content/pages/*.json files (Plenti-style)
	registerContentPageRoutes()

	// Register routes from content type directories (news, blog, etc.)
	registerContentTypeRoutes()

	// Register routes from layouts/content/*.html files (custom layouts)
	registerContentRoutes()

	// Note: All routes (including "/") are now handled by:
	// - registerContentPageRoutes() for content/pages/*.json files
	// - registerContentTypeRoutes() for content/<type>/*.json files (news, blog, etc.)
	// - registerContentRoutes() for layouts/content/*.html files
	// Static files are served via serveStaticFile() when routes call it

	// Start the server
	port := ":3333"
	fmt.Printf("Server starting on http://localhost%s\n", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// serveStaticFile handles serving static files from organized asset directories
// Routes: /scripts/* → ./scripts/, /styles/* → ./styles/, /images/* → ./images/, /js/* → ./static/js/, /static/* → ./static/, /* → ./public/
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
	case strings.HasPrefix(path, "/js/"):
		// Map /js/* to ./static/js/*
		filePath = "./static" + path
		log.Printf("[serveStaticFile] /js/ route: %s → %s", path, filePath)
	case strings.HasPrefix(path, "/static/"):
		filePath = "." + path
	default:
		// Everything else goes to public directory
		filePath = "./public" + path
	}

	// Serve the file
	log.Printf("[serveStaticFile] Serving: %s from %s", path, filePath)
	http.ServeFile(w, r, filePath)
}

// serveContentFile serves JSON content files for CMS
// Route: /content/* → ./content/*
func serveContentFile(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	filePath := "." + path

	// Set JSON content type
	w.Header().Set("Content-Type", "application/json")

	log.Printf("[serveContentFile] Serving: %s from %s", path, filePath)
	http.ServeFile(w, r, filePath)
}

// serveCoreFile serves files from the core/ directory (Plenti ejectable core)
// Route: /core/* → ./core/*
func serveCoreFile(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	filePath := "." + path // /core/main.js → ./core/main.js

	// Set appropriate content type for JS files
	if strings.HasSuffix(path, ".js") {
		w.Header().Set("Content-Type", "application/javascript")
	}

	log.Printf("[serveCoreFile] Serving: %s from %s", path, filePath)
	http.ServeFile(w, r, filePath)
}

// serveGeneratedFile serves files from the generated/ directory (Plenti build output)
// Route: /generated/* → ./generated/*
func serveGeneratedFile(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	filePath := "." + path // /generated/layouts.js → ./generated/layouts.js

	// Set appropriate content type for JS files
	if strings.HasSuffix(path, ".js") {
		w.Header().Set("Content-Type", "application/javascript")
	}

	log.Printf("[serveGeneratedFile] Serving: %s from %s", path, filePath)
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
// Returns ARRAY in Plenti format: [{type, path, fields}, ...]
// This matches Plenti's allContent structure for filtering in templates.
//
// Pattern: File Discovery Pattern with Caching [Load: 15]
// Cognitive Load: 15 (directory walk: 5, file reading: 3, JSON parsing: 3, path formatting: 2, caching: 2)
func getAllContent() []interface{} {
	// Check cache first
	allContentCacheMu.RLock()
	if allContentCached {
		cached := allContentCache
		allContentCacheMu.RUnlock()
		return cached
	}
	allContentCacheMu.RUnlock()

	// Build fresh cache as array (Plenti format)
	result := make([]interface{}, 0)
	contentDir := "content"

	// Walk content directory recursively (COGNITIVE LOAD RULE: wrapped error)
	err := filepath.Walk(contentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// If content directory doesn't exist, just return empty array
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

		// Calculate relative path: "content/news/team-expansion.json" -> "news/team-expansion"
		relPath := strings.TrimPrefix(path, contentDir+string(filepath.Separator))
		relPath = strings.TrimSuffix(relPath, ".json")
		// Normalize path separators for cross-platform compatibility
		relPath = filepath.ToSlash(relPath)

		// Extract type from path: "news/team-expansion" -> "news"
		contentType := relPath
		if idx := strings.Index(relPath, "/"); idx > 0 {
			contentType = relPath[:idx]
		}

		// Build Plenti-format entry: {type, path, fields}
		entry := map[string]interface{}{
			"type":   contentType,
			"path":   "/" + relPath,
			"fields": content,
		}
		result = append(result, entry)
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
	log.Printf("[TRACE-SERVER] ========== renderWithWrapper START ==========")
	log.Printf("[TRACE-SERVER] renderWithWrapper: layout=%s, route=%s", layoutName, r.URL.Path)

	// Step 1: Load content for this route (COGNITIVE LOAD RULE: wrapped error)
	routePath := r.URL.Path
	contentData, err := loadContentWithCache(routePath)
	if err != nil {
		// Content loading failure is not fatal - log warning and continue
		log.Printf("[TRACE-SERVER] renderWithWrapper: WARNING - failed to load content for route %s: %v", routePath, err)
		contentData = make(map[string]interface{}) // Empty content
	} else if len(contentData) > 0 {
		log.Printf("[TRACE-SERVER] renderWithWrapper: ✓ Loaded content for route %s: %d top-level keys", routePath, len(contentData))
		log.Printf("[TRACE-SERVER] renderWithWrapper: content keys: %v", getKeys(contentData))

		// CRITICAL DIAGNOSTIC: Check if components array exists
		if componentsRaw, ok := contentData["components"]; ok {
			if components, ok := componentsRaw.([]interface{}); ok {
				log.Printf("[TRACE-SERVER] renderWithWrapper: ✓✓✓ Found components array with %d items", len(components))
				log.Printf("[TRACE-SERVER] renderWithWrapper: components[0] = %#v", components[0])
			}
		} else {
			log.Printf("[TRACE-SERVER] renderWithWrapper: ✗✗✗ NO 'components' key in content data!")
		}
	}

	// Step 2: allContent generation moved to opt-in only (see magic variables below)
	// NOTE: allContent is a large dataset and should only be loaded when explicitly requested via export let
	// This improves performance by reducing HTML payload size and initial page load time

	// Step 3: allLayouts generation moved to opt-in only (see magic variables below)
	// NOTE: allLayouts is only needed for Plenti compatibility and should be opt-in

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
						log.Printf("[TRACE-SERVER] renderWithWrapper: Extracted fields from first component: %d fields", len(fields))
					}
				}
			}
		}
	} else {
		// For single type, use all content as fields
		contentFields = contentData
		log.Printf("[TRACE-SERVER] renderWithWrapper: Using full content as fields: %d keys", len(contentFields))
	}

	// If no fields extracted, use empty map
	if contentFields == nil {
		contentFields = make(map[string]interface{})
	}

	// Step 5: Build props map for wrapper
	// IMPORTANT: These are offered to the wrapper, but wrapper's export let controls which are actually used
	// The opt-in filtering happens inside renderTemplateWithProps
	// NOTE: allContent and allLayouts are OPT-IN ONLY (via magic variables) to improve performance
	props := map[string]interface{}{
		"layout":        layoutName,                   // Name of the layout to render (e.g., "_index")
		"content":       contentData,                  // Full content object
		"env":           make(map[string]interface{}), // Environment vars (TODO: populate if needed)
		"user":          make(map[string]interface{}), // User data (TODO: populate if needed)
		"shadowContent": make(map[string]interface{}), // Shadow content (TODO: populate if needed)
	}

	// Step 5.1: Extract components array as top-level prop
	// This is needed for pages.html which uses: {for component in components}
	if componentsRaw, ok := contentData["components"]; ok {
		props["components"] = componentsRaw
		if components, ok := componentsRaw.([]interface{}); ok {
			log.Printf("[TRACE-SERVER] renderWithWrapper: ✓✓✓ Extracted components array as top-level prop (%d items)", len(components))
		}
	} else {
		log.Printf("[TRACE-SERVER] renderWithWrapper: ✗✗✗ NO components array to extract!")
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

	log.Printf("[TRACE-SERVER] renderWithWrapper: Built props map with %d keys", len(props))
	log.Printf("[TRACE-SERVER] renderWithWrapper: Props keys: %v", getKeys(props))

	// CRITICAL DIAGNOSTIC: Log what's being passed
	if componentsRaw, ok := props["components"]; ok {
		if components, ok := componentsRaw.([]interface{}); ok {
			log.Printf("[TRACE-SERVER] renderWithWrapper: ✓✓✓ Props INCLUDES 'components' with %d items", len(components))
		}
	} else {
		log.Printf("[TRACE-SERVER] renderWithWrapper: ✗✗✗ Props MISSING 'components'!")
	}

	log.Printf("[TRACE-SERVER] renderWithWrapper: ========== END Step 1: Props Built ==========")

	// Step 6: Call renderTemplate with html.html wrapper and props
	// renderTemplateWithProps will filter these based on wrapper's export let declarations
	wrapperPath := "layouts/global/html.html"
	log.Printf("[TRACE-SERVER] renderWithWrapper: Rendering wrapper template: %s", wrapperPath)

	// Use renderTemplateWithProps (with opt-in filtering)
	err = renderTemplateWithProps(wrapperPath, props, w, r)
	if err != nil {
		return fmt.Errorf("renderWithWrapper: failed to render wrapper: %w", err)
	}

	log.Printf("[TRACE-SERVER] renderWithWrapper: Successfully rendered wrapper for layout: %s", layoutName)
	log.Printf("[TRACE-SERVER] ========== renderWithWrapper END ==========")
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

	// PHASE 3: Reset runtime component tracking for this page
	// Each page gets fresh tracking to determine if runtime scripts are needed
	transformer.ResetRuntimeComponentTracking()

	log.Printf("[TRACE-SERVER] ========== renderTemplateWithProps START ==========")
	log.Printf("[TRACE-SERVER] renderTemplateWithProps: entrypoint=%s", entrypoint)
	log.Printf("[TRACE-SERVER] renderTemplateWithProps: explicitProps keys=%v", getKeys(explicitProps))

	// CRITICAL DIAGNOSTIC
	if componentsRaw, ok := explicitProps["components"]; ok {
		if components, ok := componentsRaw.([]interface{}); ok {
			log.Printf("[TRACE-SERVER] renderTemplateWithProps: ✓✓✓ Received 'components' prop with %d items", len(components))
		}
	} else {
		log.Printf("[TRACE-SERVER] renderTemplateWithProps: ✗✗✗ NO 'components' prop received!")
	}

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
		log.Printf("[TRACE-SERVER] renderTemplateWithProps: Template %s exports: %v", entrypoint, fenceWithStores.ExportedProps)
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
			log.Printf("[TRACE-SERVER] renderTemplateWithProps: ✓ Added prop '%s' (declared in export let)", k)
		} else if contentFieldsProps[k] {
			props[k] = v
			log.Printf("[TRACE-SERVER] renderTemplateWithProps: ✓ Added prop '%s' (from content.fields, passthrough to children)", k)
		} else {
			log.Printf("[TRACE-SERVER] renderTemplateWithProps: ✗ Skipped prop '%s' (not in export let or content.fields)", k)
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

	log.Printf("[TRACE-SERVER] renderTemplateWithProps: Final props for %s: %v", entrypoint, getKeys(props))
	log.Printf("[TRACE-SERVER] renderTemplateWithProps: ========== END Step 2: Props Filtered ==========")

	// DIAGNOSTIC: Check if components is in final props
	if componentsRaw, ok := props["components"]; ok {
		if components, ok := componentsRaw.([]interface{}); ok {
			log.Printf("[TRACE-SERVER] renderTemplateWithProps: ✓✓✓ 'components' prop PRESENT in final props (%d items)", len(components))
		}
	} else {
		log.Printf("[TRACE-SERVER] renderTemplateWithProps: ✗✗✗ 'components' prop MISSING from final props!")
	}

	// Transform template (this tracks store references)
	log.Printf("[TRACE-SERVER] renderTemplateWithProps: Calling transformer.TransformAST with props: %v", getKeys(props))
	transformed := transformer.TransformAST(template, props)
	log.Printf("[TRACE-SERVER] renderTemplateWithProps: ========== END Step 3: Transform Complete ==========")

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

	// Extract JSON component names from content data for style aggregation
	// This allows the style aggregator to find components specified in JSON content
	// (e.g., hero2436, services2437, whyChoose2425 in content/pages/_index.json)
	var jsonComponentNames []string
	if contentData, ok := explicitProps["content"].(map[string]interface{}); ok {
		jsonComponentNames = loader.ExtractAllComponentNames(contentData)
		if len(jsonComponentNames) > 0 {
			log.Printf("[renderTemplateWithProps] Extracted %d JSON component names: %v", len(jsonComponentNames), jsonComponentNames)
		}
	}

	markup, script, style := renderer.RenderWithStores(template, transformed, finalStores, entrypoint, layoutName, jsonComponentNames)

	// Build x-data from props
	// OPTIMIZATION: Filter props to only include runtime-tracked variables
	// Variables only used at build-time (like allContent for loop expansion)
	// are excluded from x-data to reduce page weight
	filteredProps := props
	if runtimeTracker := transformer.GetRuntimeTracker(); runtimeTracker != nil {
		filteredProps = runtimeTracker.FilterScope(props)
		log.Printf("[X-Data] Filtered props from %d to %d variables", len(props), len(filteredProps))
	}
	xDataValue := buildXDataFromProps(filteredProps)

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

	// Add Alpine.js CDN if not already present (always needed)
	if !strings.Contains(finalHTML, "alpinejs") {
		headEndRegex := regexp.MustCompile(`(?i)</head>`)
		finalHTML = headEndRegex.ReplaceAllString(finalHTML,
			`<script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script></head>`)
	}

	// PHASE 3: Conditionally inject runtime component scripts
	// Only pages that use runtime component resolution get these scripts
	finalHTML = injectRuntimeScripts(finalHTML)

	// Add build time comment and floating label
	totalBuildTime := time.Since(startTime)
	htmlComment := fmt.Sprintf("<!-- Build time: %v -->\n", totalBuildTime)
	finalHTML = htmlComment + finalHTML

	// Inject floating build time label for development feedback
	finalHTML = injectBuildTimeLabel(finalHTML, totalBuildTime)

	// Add data-content-filepath for CMS (Plenti pattern)
	// This tells the CMS which JSON file contains the content for this page
	contentFilePath := loader.RoutePathToFilePath(r.URL.Path)
	if contentFilePath != "" {
		// Add attribute to <html> tag
		htmlTagRegex := regexp.MustCompile(`(?i)<html([^>]*)>`)
		finalHTML = htmlTagRegex.ReplaceAllString(finalHTML, fmt.Sprintf(`<html$1 data-content-filepath="%s">`, contentFilePath))
	}

	// Send response
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(finalHTML))

	log.Printf("[TRACE-SERVER] ========== renderTemplateWithProps END ==========")
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

	// PHASE 3: Reset runtime component tracking for this page
	// Each page gets fresh tracking to determine if runtime scripts are needed
	transformer.ResetRuntimeComponentTracking()

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
		// DEBUG: Log content keys to find where this. prefix is added
		if components, ok := contentData["components"].([]interface{}); ok && len(components) > 0 {
			if comp, ok := components[0].(map[string]interface{}); ok {
				if fields, ok := comp["fields"].(map[string]interface{}); ok {
					log.Printf("DEBUG: First component fields keys: %v", getKeys(fields))
				}
			}
		}
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
			// DEBUG: Check content keys right after adding to props
			if components, ok := contentData["components"].([]interface{}); ok && len(components) > 0 {
				if comp, ok := components[0].(map[string]interface{}); ok {
					if fields, ok := comp["fields"].(map[string]interface{}); ok {
						log.Printf("DEBUG: Content fields keys AFTER adding to props: %v", getKeys(fields))
					}
				}
			}
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
	// Extract JSON component names from content data for style aggregation
	var jsonComponentNames []string
	if contentData != nil {
		jsonComponentNames = loader.ExtractAllComponentNames(contentData)
		if len(jsonComponentNames) > 0 {
			log.Printf("[renderTemplate] Extracted %d JSON component names: %v", len(jsonComponentNames), jsonComponentNames)
		}
	}
	markup, script, style := renderer.RenderWithStores(template, transformed, finalStores, entrypoint, "", jsonComponentNames)

	// CRITICAL: Generate x-data using transformer's alpineDataFormatter
	// This function is not exported, so we need to call Transform to get the data scope
	// and then format it ourselves
	// OPTIMIZATION: Filter props to only include runtime-tracked variables
	filteredProps := props
	if runtimeTracker := transformer.GetRuntimeTracker(); runtimeTracker != nil {
		filteredProps = runtimeTracker.FilterScope(props)
		log.Printf("[X-Data] Filtered props from %d to %d variables", len(props), len(filteredProps))
	}
	xDataValue := buildXDataFromProps(filteredProps)

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

	// Add Alpine.js CDN if not already present (always needed)
	if !strings.Contains(finalHTML, "alpinejs") {
		headEndRegex := regexp.MustCompile(`(?i)</head>`)
		finalHTML = headEndRegex.ReplaceAllString(finalHTML,
			`<script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script></head>`)
	}

	// PHASE 3: Conditionally inject runtime component scripts
	// Only pages that use runtime component resolution get these scripts
	finalHTML = injectRuntimeScripts(finalHTML)

	// Add build time comment and floating label
	totalBuildTime := time.Since(startTime)
	htmlComment := fmt.Sprintf("<!-- Build time: %v -->\n", totalBuildTime)
	finalHTML = htmlComment + finalHTML

	// Inject floating build time label for development feedback
	finalHTML = injectBuildTimeLabel(finalHTML, totalBuildTime)

	// Add data-content-filepath for CMS (Plenti pattern)
	contentFilePath := loader.RoutePathToFilePath(r.URL.Path)
	if contentFilePath != "" {
		htmlTagRegex := regexp.MustCompile(`(?i)<html([^>]*)>`)
		finalHTML = htmlTagRegex.ReplaceAllString(finalHTML, fmt.Sprintf(`<html$1 data-content-filepath="%s">`, contentFilePath))
	}

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
// CRITICAL FIX: Now unwraps quoted JavaScript literals (arrays/objects) before checking type
// This prevents treating `"[...]"` as a string when it should be an actual array
//
// Pattern: Data Formatting Pattern [Load: 12]
// Cognitive Load: 12 (iterate props: 2, detect functions: 3, format values: 5, join: 2)
func buildXDataFromProps(props map[string]interface{}) string {
	log.Printf("=== buildXDataFromProps CALLED with %d props ===", len(props))
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

		// DEBUG: Log the actual value type and content
		log.Printf("DEBUG buildXDataFromProps: key=%s, value=%#v, type=%T", key, value, value)

		// Format value as JavaScript (NOT JSON)
		var formattedValue string
		switch v := value.(type) {
		case string:
			// CRITICAL FIX: Apply the SAME unwrapping logic as transformer/alpine.go
			// Check if string is quoted and unwrap, then check if it's a JS literal

			trimmed := strings.TrimSpace(v)

			// CRITICAL FIX: Check if string is an UNQUOTED JavaScript literal FIRST
			// This handles multiline objects/arrays that parser stores as raw strings (no outer quotes)
			// Example: "{\n  name: \"Benjamin\",\n  role: \"admin\"\n}"
			if transformer.IsJavaScriptLiteral(trimmed) {
				log.Printf("buildXDataFromProps: Unquoted JS literal detected for key=%s: %s", key, trimmed[:min(50, len(trimmed))])
				// CRITICAL FIX: Convert double quotes to single quotes for HTML attribute safety
				// This prevents HTML entity escaping (&quot;) which breaks Alpine.js parsing
				formattedValue = strings.ReplaceAll(trimmed, `"`, `'`)
			} else if transformer.IsFunctionExpression(trimmed) {
				log.Printf("buildXDataFromProps: Unquoted function expression detected for key=%s", key)
				formattedValue = trimmed
			} else if strings.HasPrefix(trimmed, `"`) && strings.HasSuffix(trimmed, `"`) && len(trimmed) > 1 {

				// Unwrap the double quotes
				unwrapped := trimmed[1 : len(trimmed)-1]
				log.Printf("buildXDataFromProps: Unwrapped double-quoted string for key=%s: %q → %q", key, v, unwrapped)

				// CRITICAL FIX: Check if unwrapped content is a JavaScript literal
				if transformer.IsJavaScriptLiteral(unwrapped) {
					log.Printf("buildXDataFromProps: Unwrapped content is JS literal, returning as-is: %s", unwrapped[:min(50, len(unwrapped))])
					// CRITICAL FIX: Convert double quotes to single quotes for HTML attribute safety
					// This prevents HTML entity escaping (&quot;) which breaks Alpine.js parsing
					formattedValue = strings.ReplaceAll(unwrapped, `"`, `'`)
				} else if transformer.IsFunctionExpression(unwrapped) {
					// CRITICAL FIX: Check if unwrapped content is a function expression
					log.Printf("buildXDataFromProps: Unwrapped content is function expression, returning as-is")
					formattedValue = unwrapped
				} else {
					// Regular string - re-quote with single quotes
					escaped := escapeStringForJS(unwrapped)
					formattedValue = fmt.Sprintf(`'%s'`, escaped)
					log.Printf("buildXDataFromProps: Re-quoted with single quotes: %q → %s", unwrapped, formattedValue)
				}
			} else if strings.HasPrefix(v, "function ") || strings.Contains(v, "=>") {
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

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// injectRuntimeScripts conditionally injects runtime component scripts into HTML.
// Only injects if transformer.HasRuntimeComponents() returns true for this page.
//
// Scripts injected (when needed):
//   - /core/runtime-components.js - Alpine.js magic function for runtime resolution
//   - /generated/layouts.js - Component registry (loaded by runtime-components.js)
//
// Pattern: Helper Function [Load: 5]
// Cognitive Load: 5 (check: 1, regex: 2, replace: 2)
func injectRuntimeScripts(html string) string {
	if !transformer.HasRuntimeComponents() {
		log.Printf("[injectRuntimeScripts] Page has no runtime components - skipping script injection")
		return html
	}

	log.Printf("[injectRuntimeScripts] Page uses runtime components - injecting scripts")

	// Inject runtime-components.js BEFORE Alpine.js loads
	// This ensures $renderDynamicComponent is registered before Alpine.init()
	runtimeScript := `<script src="/core/runtime-components.js"></script>`

	// Find </head> and inject before it (after Alpine.js CDN which also goes in head)
	headEndRegex := regexp.MustCompile(`(?i)</head>`)
	html = headEndRegex.ReplaceAllString(html, runtimeScript+"</head>")

	return html
}

// injectBuildTimeLabel injects a floating build time label into the HTML.
// This provides visual feedback during development showing how fast pages build.
//
// Pattern: Helper Function [Load: 5]
// Cognitive Load: 5 (format: 1, regex: 2, replace: 2)
func injectBuildTimeLabel(html string, buildTime time.Duration) string {
	// Format build time nicely
	var buildTimeStr string
	if buildTime < time.Millisecond {
		buildTimeStr = fmt.Sprintf("%.2fµs", float64(buildTime.Microseconds()))
	} else if buildTime < time.Second {
		buildTimeStr = fmt.Sprintf("%.2fms", float64(buildTime.Microseconds())/1000.0)
	} else {
		buildTimeStr = fmt.Sprintf("%.2fs", buildTime.Seconds())
	}

	// Create floating label with inline styles (no external CSS needed)
	label := fmt.Sprintf(`<div id="build-time-label" style="
		position: fixed;
		top: 10px;
		left: 10px;
		background: white;
		color: #333;
		padding: 8px 16px;
		border-radius: 20px;
		font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
		font-size: 12px;
		font-weight: 600;
		box-shadow: 0 2px 12px rgba(0, 0, 0, 0.15);
		z-index: 99999;
		display: flex;
		align-items: center;
		gap: 8px;
		cursor: pointer;
		transition: transform 0.2s, opacity 0.2s;
		border: 1px solid #e5e7eb;
	" onclick="this.style.display='none'" title="Click to dismiss">
		<img src="/images/plenti.png" alt="Plenti" style="height: 20px; width: auto;">
		<span>Build: <strong>%s</strong></span>
	</div>`, buildTimeStr)

	// Inject before </body>
	bodyEndRegex := regexp.MustCompile(`(?i)</body>`)
	if bodyEndRegex.MatchString(html) {
		html = bodyEndRegex.ReplaceAllString(html, label+"</body>")
	}

	return html
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
	// Use full Plenti-compatible path prefix for signatures (layouts_components_*)
	componentDir := "layouts/components"
	registerComponentsFromDir(componentDir, "layouts/components/", storeRegistry)

	// Register global layout components from layouts/global
	globalDir := "layouts/global"
	registerComponentsFromDir(globalDir, "layouts/global/", storeRegistry)

	// Register content layouts from layouts/content
	// All layouts need to be registered (matching Plenti's architecture where pages.svelte is a component)
	// Components without fence sections won't get x-data wrapping (handled by transformer)
	contentDir := "layouts/content"
	registerComponentsFromDir(contentDir, "layouts/content/", storeRegistry)
}

// registerContentLayoutsSelectively registers content layouts from layouts/content/
// but SKIPS layouts that match the "component iterator" pattern.
//
// Component iterator layouts (like pages.html with Plenti pattern) should NOT be registered
// as components because they act as entry points that iterate over dynamic components.
//
// Pattern Detection:
// - Contains: {for component in content.components} or {for component in components}
// - Contains: <Component:dynamic
//
// Pattern: Selective Component Registration [Load: 15]
// Cognitive Load: 15 (read dir: 2, read file: 3, pattern detection: 5, registration: 5)
func registerContentLayoutsSelectively(dir string, pathPrefix string, storeRegistry map[string]string) {
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("Warning: Failed to read directory %s: %v", dir, err)
		return
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".html") {
			componentPath := fmt.Sprintf("%s/%s", dir, file.Name())

			// Read file content to detect patterns
			componentContent, err := os.ReadFile(componentPath)
			if err != nil {
				log.Printf("Warning: Failed to read %s: %v", componentPath, err)
				continue
			}

			contentStr := string(componentContent)

			// Check if this is a component iterator layout (like Plenti's pages.svelte pattern)
			isComponentIterator := isComponentIteratorLayout(contentStr)

			if isComponentIterator {
				log.Printf("Skipping component iterator layout: %s (follows Plenti pattern)", file.Name())
				continue
			}

			// Register as normal component
			baseName := strings.TrimSuffix(file.Name(), ".html")
			componentName := strings.ToUpper(baseName[:1]) + baseName[1:]
			log.Printf("Registering content layout as component: %s from %s", componentName, componentPath)

			// Parse and register (same logic as registerComponentsFromDir)
			componentAST, err := parser.ParseTemplate(contentStr)
			if err != nil {
				log.Printf("Warning: Failed to parse component %s: %v", componentPath, err)
				continue
			}

			// Handle store imports if present
			for i, node := range componentAST.RootNodes {
				if fence, ok := node.(*ast.FenceSection); ok {
					if strings.Contains(fence.RawContent, "import store from") {
						fenceWithStores := parser.ParseFenceContentWithStores(fence.RawContent, storeRegistry)
						componentAST.RootNodes[i] = fenceWithStores
						log.Printf("Re-parsed fence with stores for %s", componentName)
					}
					break
				}
			}

			// Extract props and register
			componentProps := extractComponentProps(componentAST)
			transformer.RegisterComponent(componentName, componentAST, componentProps)

			// Also register with path prefix for import resolution
			pathWithPrefix := fmt.Sprintf("%s%s", pathPrefix, file.Name())
			transformer.RegisterComponent(pathWithPrefix, componentAST, componentProps)
		}
	}
}

// isComponentIteratorLayout detects if a layout follows the "component iterator" pattern
// used in Plenti (e.g., pages.svelte that iterates over content.components)
//
// Pattern: Layout Pattern Detection [Load: 8]
// Cognitive Load: 8 (pattern matching: 5, string checks: 3)
func isComponentIteratorLayout(content string) bool {
	// Pattern 1: Contains loop over components array
	hasComponentLoop := strings.Contains(content, "{for component in content.components}") ||
		strings.Contains(content, "{for component in components}") ||
		strings.Contains(content, "{#each components as") || // Svelte syntax
		strings.Contains(content, "{#each content.components as") // Svelte syntax

	// Pattern 2: Contains dynamic component rendering
	hasDynamicComponent := strings.Contains(content, "<Component:dynamic") ||
		strings.Contains(content, "<svelte:component") // Svelte syntax

	// Must have BOTH patterns to be considered a component iterator layout
	return hasComponentLoop && hasDynamicComponent
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
			// Extract base name (without .html extension)
			// e.g., "hero2436.html" -> "hero2436"
			baseName := strings.TrimSuffix(file.Name(), ".html")
			componentPath := fmt.Sprintf("%s/%s", dir, file.Name())
			log.Printf("Registering component: %s from %s", baseName, componentPath)

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
							baseName, len(fenceWithStores.Stores), len(fenceWithStores.Functions))
					} else {
						// No store imports - keep the already-parsed fence with functions intact
						log.Printf("[registerComponents] Preserved original fence for %s (functions: %d)",
							baseName, len(fence.Functions))
					}
					break
				}
			}

			// Extract props from the component template
			componentProps := extractComponentProps(componentAST)

			// PLENTI PATTERN: Register using types.NewComponentTemplate for proper signatures
			// The path prefix follows Plenti's layout structure: layouts/{category}/{name}.html
			// Examples:
			//   - layouts/components/hero2436.html → Signature: layouts_components_hero2436_html
			//   - layouts/global/nav.html → Signature: layouts_global_nav_html
			//
			// Registration uses RegisterComponentTemplate which registers by:
			//   1. Short name (hero2436) - for backward compatibility
			//   2. Full signature (layouts_components_hero2436_html) - for Plenti lookup
			//   3. File path (layouts/components/hero2436.html) - for import resolution
			filePath := fmt.Sprintf("%s%s", pathPrefix, file.Name())
			ct := types.NewComponentTemplate(filePath, componentAST, componentProps)
			transformer.RegisterComponentTemplate(ct)

			log.Printf("✓ Registered: '%s' (signature: %s)", ct.Name, ct.Signature)
		}
	}
}

// registerStores scans the stores/ directory for .js files and loads them
// Returns a map of store name (filename without .js) to store content
//
// Pattern: File Discovery Pattern [Load: 8]
// Cognitive Load: 8 (read dir: 2, filter: 2, read files: 2, map building: 2)
// registerStores loads all store files from the stores/ directory.
// This is a wrapper around utils.RegisterStores() for backward compatibility.
//
// Pattern: Delegation to Shared Utility [Load: 1]
// Cognitive Load: 1 (simple function call delegation)
func registerStores() map[string]string {
	return utils.RegisterStores()
}

// generateComponentRegistry generates the JavaScript component registry file
// for runtime component resolution.
//
// NOTE: Registry is always generated at startup because we can't know ahead of time
// which pages will use runtime components. The optimization happens in Phase 3
// where scripts are conditionally injected based on per-page runtime component usage.
//
// Pattern: Code Generation Pattern [Load: 12]
// Cognitive Load: 12 (get components: 3, generate registry: 5, write file: 4)
func generateComponentRegistry() error {
	// Reset runtime component tracking for fresh state
	transformer.ResetRuntimeComponentTracking()

	log.Println("Generating component registry...")

	// Get all registered component keys from transformer
	componentKeys := transformer.GetAllRegisteredKeys()
	log.Printf("Found %d registered component keys", len(componentKeys))

	// Deduplicate components by signature to avoid generating the same component multiple times.
	// Components may be registered under multiple keys (short name, path, signature).
	seenSignatures := make(map[string]bool)
	builderComponents := make([]builder.ComponentTemplate, 0)
	for _, key := range componentKeys {
		tmpl, exists := transformer.GetComponentTemplate(key)
		if !exists {
			log.Printf("WARNING: Component %s not found in registry", key)
			continue
		}

		// Deduplicate by signature
		if seenSignatures[tmpl.Signature] {
			continue
		}
		seenSignatures[tmpl.Signature] = true

		// Use the component directly (types.ComponentTemplate is now the shared type)
		builderComponents = append(builderComponents, *tmpl)
	}
	log.Printf("Deduplicated to %d unique components", len(builderComponents))

	// Generate the registry JavaScript
	registryJS := builder.GenerateComponentRegistry(builderComponents)

	// Ensure the generated directory exists (Plenti structure)
	generatedDir := "generated"
	if err := os.MkdirAll(generatedDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", generatedDir, err)
	}

	// Write layouts.js (Plenti-compatible name for component registry)
	layoutsPath := filepath.Join(generatedDir, "layouts.js")
	if err := os.WriteFile(layoutsPath, []byte(registryJS), 0644); err != nil {
		return fmt.Errorf("failed to write layouts file: %w", err)
	}

	log.Printf("✓ Layouts registry generated: %s (%d components)", layoutsPath, len(builderComponents))
	return nil
}

// generateContentJS generates the content.js file from the content directory
// This creates a Plenti-compatible allContent array
func generateContentJS() error {
	contentDir := "content"
	outputPath := "generated/content.js"

	// Check if content directory exists
	if _, err := os.Stat(contentDir); os.IsNotExist(err) {
		log.Printf("Content directory not found: %s (skipping content.js generation)", contentDir)
		return nil
	}

	// Generate content.js using the builder
	if err := builder.WriteContentJS(contentDir, outputPath); err != nil {
		return fmt.Errorf("failed to generate content.js: %w", err)
	}

	log.Printf("✓ Content generated: %s", outputPath)
	return nil
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

	// Get list of JSON files in content/pages/ to avoid conflicts
	pagesDir := "content/pages"
	pageFiles, err := os.ReadDir(pagesDir)
	contentPageRoutes := make(map[string]bool)
	if err == nil {
		for _, pageFile := range pageFiles {
			if !pageFile.IsDir() && strings.HasSuffix(pageFile.Name(), ".json") {
				routeName := strings.TrimSuffix(pageFile.Name(), ".json")
				if routeName != "_defaults" {
					contentPageRoutes[routeName] = true
				}
			}
		}
	}

	routeCount := 0
	for _, file := range files {
		// Skip directories and non-HTML files
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".html") {
			continue
		}

		// Extract route name from filename (e.g., "store-demo.html" → "store-demo")
		routeName := strings.TrimSuffix(file.Name(), ".html")

		// Skip _index.html (handled by registerContentPageRoutes via _index.json)
		if routeName == "_index" {
			continue
		}

		// Skip routes that are already handled by content/pages/*.json files
		// This prevents conflicts between Pattern 1 (JSON-driven) and Pattern 2 (HTML-driven) routes
		if contentPageRoutes[routeName] {
			log.Printf("Skipping route %s - already registered from content/pages/%s.json", routeName, routeName)
			continue
		}

		// Build file path
		filePath := filepath.Join(contentDir, file.Name())

		// Register route (construct route path)
		route := "/" + routeName

		// Default handling - use direct template rendering
		currentFilePath := filePath // Capture for closure
		http.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
			renderTemplate(currentFilePath, w, r)
		})

		routeCount++
		log.Printf("Registered route: %s → %s", route, filePath)
	}

	log.Printf("Registered %d dynamic content routes", routeCount)
}

// registerContentPageRoutes registers HTTP routes for all .json files in content/pages/
// This follows the Plenti pattern where routes come from content JSON files, not layout HTML files.
// All Pattern 1 pages (with components array) use their corresponding layout via renderWithWrapper.
//
// Pattern: Dynamic Route Registration from Content [Load: 12]
// Cognitive Load: 12 (directory read: 3, file filtering: 2, route creation: 4, logging: 3)
func registerContentPageRoutes() {
	pagesDir := "content/pages"

	// Read all JSON files in content/pages/
	files, err := os.ReadDir(pagesDir)
	if err != nil {
		log.Printf("Warning: Failed to read pages directory %s: %v", pagesDir, err)
		return
	}

	routeCount := 0
	for _, file := range files {
		// Skip directories and non-JSON files
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		// Skip _defaults.json - it's not a page
		fileName := file.Name()
		if fileName == "_defaults.json" {
			continue
		}

		// Extract route name from filename
		// "jim-test.json" → "/jim-test"
		// "_index.json" → "/" (special case)
		routeName := strings.TrimSuffix(fileName, ".json")

		var route string
		if routeName == "_index" {
			route = "/" // Special case: _index.json → root route
		} else {
			route = "/" + routeName
		}

		// Determine layout name
		// All content pages use "pages" layout (the generic component loop layout)
		// except _index which uses "Pages" (capital P - note the naming convention)
		layoutName := "pages"
		if routeName == "_index" {
			layoutName = "_index"
		}

		// Register route using renderWithWrapper
		// renderWithWrapper will load content/pages/<route>.json based on URL path
		// and render with layouts/<layoutName>.html wrapper
		currentRoute := route
		currentLayoutName := layoutName
		http.HandleFunc(currentRoute, func(w http.ResponseWriter, r *http.Request) {
			log.Printf("[Handler] %s called for URL: %s", currentRoute, r.URL.Path)
			if err := renderWithWrapper(currentLayoutName, w, r); err != nil {
				log.Printf("Error rendering %s with wrapper (layout %s): %v", currentRoute, currentLayoutName, err)
				http.Error(w, "Failed to render page", http.StatusInternalServerError)
			}
		})

		routeCount++
		log.Printf("Registered content page route: %s → content/pages/%s (layout: %s)", route, fileName, layoutName)
	}

	log.Printf("Registered %d content page routes from content/pages/", routeCount)
}

// registerContentTypeRoutes registers HTTP routes for content type directories (news, blog, etc.)
// This handles the Plenti pattern where content types like "news" have:
// - Content files: content/news/*.json
// - Layout file: layouts/content/news.html
// - Routes: /news/*, /news/<slug>
//
// Pattern: Dynamic Content Type Route Registration [Load: 15]
// Cognitive Load: 15 (directory scan: 4, layout check: 3, route creation: 5, logging: 3)
func registerContentTypeRoutes() {
	contentDir := "content"
	layoutDir := "layouts/content"

	// Read all entries in content/ directory
	entries, err := os.ReadDir(contentDir)
	if err != nil {
		log.Printf("Warning: Failed to read content directory %s: %v", contentDir, err)
		return
	}

	totalRoutes := 0

	for _, entry := range entries {
		// Only process directories (content types)
		if !entry.IsDir() {
			continue
		}

		contentTypeName := entry.Name()

		// Skip "pages" - handled by registerContentPageRoutes
		if contentTypeName == "pages" {
			continue
		}

		// Check if matching layout exists
		layoutPath := filepath.Join(layoutDir, contentTypeName+".html")
		if _, err := os.Stat(layoutPath); os.IsNotExist(err) {
			log.Printf("Skipping content type %s - no matching layout at %s", contentTypeName, layoutPath)
			continue
		}

		// Register routes for this content type
		typeDir := filepath.Join(contentDir, contentTypeName)
		routeCount := registerContentTypeFiles(contentTypeName, typeDir)
		totalRoutes += routeCount

		log.Printf("Registered %d routes for content type: %s", routeCount, contentTypeName)
	}

	log.Printf("Registered %d total content type routes", totalRoutes)
}

// registerContentTypeFiles registers routes for all JSON files in a content type directory
// Pattern: Content Type File Registration [Load: 12]
func registerContentTypeFiles(contentTypeName, typeDir string) int {
	files, err := os.ReadDir(typeDir)
	if err != nil {
		log.Printf("Warning: Failed to read content type directory %s: %v", typeDir, err)
		return 0
	}

	routeCount := 0

	for _, file := range files {
		// Skip directories and non-JSON files
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		// Skip special files
		fileName := file.Name()
		if fileName == "_defaults.json" || fileName == "_schema.json" {
			continue
		}

		// Extract slug from filename: "new-product-launch.json" → "new-product-launch"
		slug := strings.TrimSuffix(fileName, ".json")

		// Create route: /news/new-product-launch
		route := "/" + contentTypeName + "/" + slug

		// Capture variables for closure
		currentRoute := route
		currentLayoutName := contentTypeName
		currentContentType := contentTypeName
		currentSlug := slug

		http.HandleFunc(currentRoute, func(w http.ResponseWriter, r *http.Request) {
			log.Printf("[ContentType Handler] %s called for URL: %s", currentRoute, r.URL.Path)
			if err := renderContentTypePage(currentLayoutName, currentContentType, currentSlug, w, r); err != nil {
				log.Printf("Error rendering %s: %v", currentRoute, err)
				http.Error(w, "Failed to render page", http.StatusInternalServerError)
			}
		})

		routeCount++
		log.Printf("Registered content type route: %s → %s/%s (layout: %s)", route, typeDir, fileName, contentTypeName)
	}

	return routeCount
}

// renderContentTypePage renders a content type page (news, blog, etc.) with the global wrapper
// Uses the same pattern as renderWithWrapper but loads content from content/<type>/<slug>.json
//
// Pattern: Content Type Page Rendering with Wrapper [Load: 18]
func renderContentTypePage(layoutName, contentType, slug string, w http.ResponseWriter, r *http.Request) error {
	log.Printf("[TRACE-SERVER] ========== renderContentTypePage START ==========")
	log.Printf("[TRACE-SERVER] renderContentTypePage: layout=%s, contentType=%s, slug=%s", layoutName, contentType, slug)

	// Step 1: Load content from content/<contentType>/<slug>.json
	contentPath := filepath.Join("content", contentType, slug+".json")
	contentData, err := loader.LoadContentJSON(contentPath)
	if err != nil {
		return fmt.Errorf("renderContentTypePage: failed to load content %s: %w", contentPath, err)
	}

	log.Printf("[TRACE-SERVER] renderContentTypePage: Loaded content from %s: %d keys", contentPath, len(contentData))
	log.Printf("[TRACE-SERVER] renderContentTypePage: content keys: %v", getKeys(contentData))

	// Step 2: Extract fields - content types use flat structure OR nested "fields"
	// Check if data has nested "fields" structure (legacy format) or flat (Plenti standard)
	var contentFields map[string]interface{}

	if fieldsRaw, hasFields := contentData["fields"]; hasFields {
		// Legacy format: data nested in "fields"
		if fields, ok := fieldsRaw.(map[string]interface{}); ok {
			log.Printf("[TRACE-SERVER] renderContentTypePage: Using nested 'fields' structure with %d fields", len(fields))
			contentFields = fields
		} else {
			contentFields = contentData
		}
	} else {
		// Plenti standard: flat structure at root (all keys are fields)
		log.Printf("[TRACE-SERVER] renderContentTypePage: Using flat structure")
		contentFields = contentData
	}

	// If no fields extracted, use empty map
	if contentFields == nil {
		contentFields = make(map[string]interface{})
	}

	// Step 3: Build props map for wrapper (same pattern as renderWithWrapper)
	props := map[string]interface{}{
		"layout":        layoutName,                   // Name of the layout to render (e.g., "news")
		"env":           make(map[string]interface{}), // Environment vars
		"user":          make(map[string]interface{}), // User data
		"shadowContent": make(map[string]interface{}), // Shadow content
	}

	// Step 4: Build content object with fields
	// The wrapper passes content.fields to the dynamic component via {...content.fields}
	contentWithFields := map[string]interface{}{
		"fields": contentFields,
	}
	// Preserve any top-level content keys that aren't "fields"
	for key, val := range contentData {
		if key != "fields" {
			contentWithFields[key] = val
		}
	}
	props["content"] = contentWithFields

	// Step 4a: Add magic variables (same pattern as renderWithWrapper)
	// allContent is needed for sidebars that list other content (e.g., Featured Posts)
	props["allContent"] = getAllContent()
	log.Printf("[TRACE-SERVER] renderContentTypePage: Added allContent magic variable (%d items)", len(getAllContent()))

	// allLayouts for dynamic component lookups
	layoutNames := make([]string, 0)
	for name := range transformer.GetAllComponentNames() {
		layoutNames = append(layoutNames, name)
	}
	props["allLayouts"] = layoutNames

	log.Printf("[TRACE-SERVER] renderContentTypePage: Built props with %d keys", len(props))
	log.Printf("[TRACE-SERVER] renderContentTypePage: Props keys: %v", getKeys(props))
	log.Printf("[TRACE-SERVER] renderContentTypePage: content.fields has %d keys: %v", len(contentFields), getKeys(contentFields))

	// Step 5: Render with html.html wrapper (same as renderWithWrapper)
	wrapperPath := "layouts/global/html.html"
	log.Printf("[TRACE-SERVER] renderContentTypePage: Rendering with wrapper: %s, layout: %s", wrapperPath, layoutName)

	err = renderTemplateWithProps(wrapperPath, props, w, r)
	if err != nil {
		return fmt.Errorf("renderContentTypePage: failed to render wrapper: %w", err)
	}

	log.Printf("[TRACE-SERVER] renderContentTypePage: Successfully rendered %s/%s with layout %s", contentType, slug, layoutName)
	log.Printf("[TRACE-SERVER] ========== renderContentTypePage END ==========")
	return nil
}
