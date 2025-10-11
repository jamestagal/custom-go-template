package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
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

	// Set up the HTTP server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Serve static files from organized directories
		if r.URL.Path != "/" {
			serveStaticFile(w, r)
			return
		}

		// Render home page using Plenti-style component iteration
		renderPlentiPage(w, r)
	})

	// Add comprehensive-simple page route (WORKING - no multi-line vars)
	http.HandleFunc("/comprehensive-simple", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate("layouts/content/comprehensive.html", w, r)
	})

	// Add comprehensive page route (HAS BUGS - multi-line var extraction broken)
	http.HandleFunc("/comprehensive", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate("layouts/content/comprehensive.html", w, r)
	})

	// Add store-test-minimal page route (Testing Task 2.1: Store Expression Transformer - no definitions)
	http.HandleFunc("/store-test-minimal", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate("layouts/content/store-test-minimal.html", w, r)
	})

	// Add store-test-with-theme page route (Testing visual theme switching with stores)
	http.HandleFunc("/store-test-with-theme", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate("layouts/content/store-test-with-theme.html", w, r)
	})

	// TASK 4.1 & 4.2: Store components demo page with content loading
	http.HandleFunc("/store-components-demo", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate("layouts/content/store-demo.html", w, r)
	})
	http.HandleFunc("/store-demo", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate("layouts/content/store-demo.html", w, r)
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
// Routes: /scripts/* → ./scripts/, /styles/* → ./styles/, /images/* → ./images/, /* → ./public/
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
	log.Println("Content cache invalidated")
}

// renderPlentiPage renders a page using Plenti-style component iteration
// This mimics Svelte's {#each components as {name, fields}} pattern but at the server level.
// No template file is needed - components are rendered directly from the registry based on
// the components array in the content JSON.
//
// Pattern: Plenti Component Iteration [Load: 25]
// Cognitive Load: 25 (content load: 3, component iteration: 5, component rendering: 8, assembly: 5, output: 4)
func renderPlentiPage(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	routePath := r.URL.Path

	// Load content JSON for this route
	contentData, err := loadContentWithCache(routePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load content: %v", err), http.StatusInternalServerError)
		return
	}

	// Check if content has components array (Plenti format)
	componentsRaw, hasComponents := contentData["components"]
	if !hasComponents {
		http.Error(w, "Content must have a 'components' array for Plenti-style rendering", http.StatusInternalServerError)
		return
	}

	// Type assert to array
	components, ok := componentsRaw.([]interface{})
	if !ok {
		http.Error(w, fmt.Sprintf("Invalid components structure: expected array, got %T", componentsRaw), http.StatusInternalServerError)
		return
	}

	log.Printf("renderPlentiPage: found %d components to render", len(components))

	// Render each component and collect HTML/CSS/JS
	var allMarkup strings.Builder
	var allStyles strings.Builder
	var allScripts strings.Builder
	var allStores = make(map[string]string)

	for i, compRaw := range components {
		comp, ok := compRaw.(map[string]interface{})
		if !ok {
			log.Printf("Warning: component %d is not a map, skipping", i)
			continue
		}

		// Extract component name
		nameRaw, ok := comp["name"]
		if !ok {
			log.Printf("Warning: component %d has no name, skipping", i)
			continue
		}

		name, ok := nameRaw.(string)
		if !ok {
			log.Printf("Warning: component %d name is not a string, skipping", i)
			continue
		}

		// Extract component fields
		fieldsRaw, ok := comp["fields"]
		if !ok {
			log.Printf("Warning: component %s has no fields, using empty map", name)
			fieldsRaw = make(map[string]interface{})
		}

		fields, ok := fieldsRaw.(map[string]interface{})
		if !ok {
			log.Printf("Warning: component %s fields is not a map, using empty map", name)
			fields = make(map[string]interface{})
		}

		log.Printf("Rendering component %s with %d fields", name, len(fields))

		// Render the component with its fields
		markup, script, style, stores := renderComponentWithFields(name, fields)

		allMarkup.WriteString(markup)
		allMarkup.WriteString("\n")

		if style != "" {
			allStyles.WriteString(style)
			allStyles.WriteString("\n")
		}

		if script != "" {
			allScripts.WriteString(script)
			allScripts.WriteString("\n")
		}

		// Merge stores
		for storeName, storeDef := range stores {
			if _, exists := allStores[storeName]; !exists {
				allStores[storeName] = storeDef
			}
		}
	}

	// Build final HTML page
	finalHTML := buildHTMLPage(allMarkup.String(), allStyles.String(), allScripts.String(), allStores)

	// Add build time comment
	totalBuildTime := time.Since(startTime)
	htmlComment := fmt.Sprintf("<!-- Build time: %v -->\n", totalBuildTime)
	finalHTML = htmlComment + finalHTML

	// Send response
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(finalHTML))
}

// renderComponentWithFields renders a single component with given fields as props
// Returns: markup, script, style, stores
func renderComponentWithFields(componentName string, fields map[string]interface{}) (string, string, string, map[string]string) {
	// Look up component template
	componentTemplate, exists := transformer.GetComponentTemplate(componentName)
	if !exists {
		log.Printf("Warning: Component %s not found in registry", componentName)
		return fmt.Sprintf("<!-- Component %s not found -->", componentName), "", "", nil
	}

	// Build props map from fields
	props := make(map[string]interface{})
	for key, value := range fields {
		props[key] = value
	}

	// Transform the component AST with props
	transformed := transformer.TransformAST(componentTemplate.Template, props)

	// Get stores from transformer
	referencedStores, allDefinitions := transformer.GetTrackedStores(transformed)
	referencedStoreDefs := transformer.GetReferencedStoreDefinitions(allDefinitions, referencedStores)

	// Merge with external stores if needed
	finalStores := make(map[string]string)
	for name, def := range referencedStoreDefs {
		finalStores[name] = def
	}
	for _, storeName := range referencedStores {
		if _, exists := finalStores[storeName]; !exists {
			if externalDef, exists := storeRegistry[storeName]; exists {
				finalStores[storeName] = externalDef
			}
		}
	}

	// Render the component
	markup, script, style := renderer.RenderWithStores(componentTemplate.Template, transformed, finalStores, "")

	return markup, script, style, finalStores
}

// buildHTMLPage assembles the final HTML page from components
func buildHTMLPage(markup, styles, scripts string, stores map[string]string) string {
	var html strings.Builder

	html.WriteString("<!DOCTYPE html>\n")
	html.WriteString("<html lang=\"en\">\n")
	html.WriteString("<head>\n")
	html.WriteString("  <meta charset=\"UTF-8\">\n")
	html.WriteString("  <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n")
	html.WriteString("  <title>Page</title>\n")
	html.WriteString("  <link rel=\"stylesheet\" href=\"/styles/style.css\">\n")

	// Add styles
	if styles != "" {
		html.WriteString("  <style>\n")
		html.WriteString(styles)
		html.WriteString("  </style>\n")
	}

	// Add Alpine.js CDN
	html.WriteString("  <script defer src=\"https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js\"></script>\n")

	html.WriteString("</head>\n")
	html.WriteString("<body>\n")

	// Add component markup
	html.WriteString(markup)

	// Add scripts (including store initializations)
	if scripts != "" || len(stores) > 0 {
		html.WriteString("  <script>\n")

		// Add store initializations
		if len(stores) > 0 {
			html.WriteString("    // Initialize Alpine.js stores\n")
			html.WriteString("    document.addEventListener('alpine:init', () => {\n")
			for storeName, storeDef := range stores {
				html.WriteString(fmt.Sprintf("      Alpine.store('%s', %s);\n", storeName, storeDef))
			}
			html.WriteString("    });\n")
		}

		if scripts != "" {
			html.WriteString(scripts)
		}

		html.WriteString("  </script>\n")
	}

	html.WriteString("</body>\n")
	html.WriteString("</html>\n")

	return html.String()
}

// renderTemplate is a unified handler for rendering template files with store support
// Now integrates with the global store system (Task 3.5)
// UPDATED: Now supports content injection (Task 4)
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
			props[variable.Name] = parseValue(variable.Value)
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
	markup, script, style := renderer.RenderWithStores(template, transformed, finalStores, entrypoint)

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

// extractFunctionsFromFence extracts function declarations from fence section content
// Returns a map of function name -> function body
//
// DEPRECATED: This function is kept for backward compatibility but should not be needed
// now that FenceSection.Functions exists. It will be removed in a future refactor.
//
// Pattern: Regex Extraction Pattern [Load: 10]
// Cognitive Load: 10 (regex compile: 2, find all: 3, extract names: 2, map construction: 3)
func extractFunctionsFromFence(content string) map[string]string {
	functions := make(map[string]string)

	// Regex to match function declarations:
	// function name(...) { ... }
	// Handles nested braces correctly
	funcRegex := regexp.MustCompile(`function\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*\([^)]*\)\s*\{`)

	// Find all function declarations
	matches := funcRegex.FindAllStringSubmatchIndex(content, -1)

	for _, match := range matches {
		// match[0], match[1] are the full match indices
		// match[2], match[3] are the function name indices
		funcStart := match[0]
		nameStart := match[2]
		nameEnd := match[3]

		funcName := content[nameStart:nameEnd]

		// Find the matching closing brace for this function
		braceStart := match[1] // End of the initial match (right after opening {)
		braceDepth := 1
		braceEnd := braceStart

		for braceEnd < len(content) && braceDepth > 0 {
			if content[braceEnd] == '{' {
				braceDepth++
			} else if content[braceEnd] == '}' {
				braceDepth--
			}
			braceEnd++
		}

		if braceDepth == 0 {
			// Successfully found the full function
			funcBody := content[funcStart:braceEnd]
			functions[funcName] = funcBody
			log.Printf("[extractFunctionsFromFence] Found function: %s (%d chars)", funcName, len(funcBody))
		} else {
			log.Printf("[extractFunctionsFromFence] WARNING: Could not find closing brace for function %s", funcName)
		}
	}

	return functions
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
			formattedValue = transformer.FormatGoValueToJS(v)		}

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
	// Escape HTML entities that would break the attribute
	// Note: We're using double quotes for the attribute, so escape double quotes
	value = strings.ReplaceAll(value, `&`, `&amp;`)
	value = strings.ReplaceAll(value, `"`, `&quot;`)
	value = strings.ReplaceAll(value, `<`, `&lt;`)
	value = strings.ReplaceAll(value, `>`, `&gt;`)

	return value
}

// registerComponents scans the components directory and registers each component
// Now accepts storeRegistry to parse component fence sections with store imports
// Also registers global layout components from layouts/global/
//
// Pattern: File Discovery Pattern with Store Integration [Load: 15]
// Cognitive Load: 15 (read 2 dirs: 4, iterate: 2, read file: 2, parse: 2, fence parsing: 2, register: 3)
func registerComponents(storeRegistry map[string]string) {
	// Register regular components from layouts/components
	componentDir := "layouts/components"
	registerComponentsFromDir(componentDir, "../components/", storeRegistry)

	// Register global layout components from layouts/global
	globalDir := "layouts/global"
	registerComponentsFromDir(globalDir, "../global/", storeRegistry)
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
