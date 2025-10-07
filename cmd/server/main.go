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
	"time"

	// Import the new renderer package
	"github.com/jimafisk/custom_go_template/ast"
	"github.com/jimafisk/custom_go_template/parser"
	"github.com/jimafisk/custom_go_template/renderer"
	"github.com/jimafisk/custom_go_template/transformer"
)

// Global store registry loaded at startup
// Pattern: Package-level State [Load: 2]
var storeRegistry map[string]string

func main() {
	log.Println("Starting server...")

	// Create the public directory if it doesn't exist
	publicDir := "./public" // Use a variable for clarity
	err := os.MkdirAll(publicDir, 0755)
	if err != nil {
		log.Fatalf("Failed to create public directory: %v", err)
	}

	// Register components
	registerComponents()

	// Register stores (now stored in package-level variable)
	storeRegistry = registerStores()
	log.Printf("Registered %d store(s)", len(storeRegistry))

	// Set up the HTTP server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Serve static files from the public directory
		if r.URL.Path != "/" {
			http.ServeFile(w, r, publicDir+r.URL.Path)
			return
		}

		// Render home page
		renderTemplate("examples/pages/home.html", w, r)
	})

	// Add comprehensive-simple page route (WORKING - no multi-line vars)
	http.HandleFunc("/comprehensive-simple", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate("examples/pages/comprehensive-simple.html", w, r)
	})

	// Add comprehensive page route (HAS BUGS - multi-line var extraction broken)
	http.HandleFunc("/comprehensive", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate("examples/pages/comprehensive.html", w, r)
	})

	// Add store-test-minimal page route (Testing Task 2.1: Store Expression Transformer - no definitions)
	http.HandleFunc("/store-test-minimal", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate("examples/pages/store-test-minimal.html", w, r)
	})

	// Add store-test-with-theme page route (Testing visual theme switching with stores)
	http.HandleFunc("/store-test-with-theme", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate("examples/pages/store-test-with-theme.html", w, r)
	})

	// Start the server
	port := ":3333"
	fmt.Printf("Server starting on http://localhost%s\n", port)
	err = http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// renderTemplate is a unified handler for rendering template files with store support
// Now integrates with the global store system (Task 3.5)
//
// Pattern: Service Implementation Pattern with Store Integration [Load: 18]
// Cognitive Load: 18 (read: 2, parse: 3, fence parsing: 3, transform: 3, store merge: 3, render: 2, inject: 2)
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

	// Extract fence section and parse with store registry (Task 3.5 integration)
	var fenceWithStores *ast.FenceSection
	for i, node := range template.RootNodes {
		if fence, ok := node.(*ast.FenceSection); ok {
			// Parse fence content with store registry to resolve store imports
			fenceWithStores = parser.ParseFenceContentWithStores(fence.RawContent, storeRegistry)
			// Replace the fence section in template
			template.RootNodes[i] = fenceWithStores
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

		// CRITICAL: Extract functions from RawContent
		// The parser doesn't extract function declarations into Variables
		// So we need to manually parse them from the raw fence content
		extractedFunctions := extractFunctionsFromFence(fenceWithStores.RawContent)
		for name, funcBody := range extractedFunctions {
			props[name] = funcBody
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
	referencedStoreDefs := transformer.GetReferencedStoreDefinitions(allDefinitions, referencedStores)

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
			// Complex types (arrays, objects) - marshal to JSON
			jsonBytes, err := json.Marshal(v)
			if err != nil {
				// Fallback to string representation
				escaped := escapeStringForJS(fmt.Sprintf("%v", v))
				formattedValue = fmt.Sprintf(`'%s'`, escaped)
			} else {
				formattedValue = string(jsonBytes)
			}
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
	// Escape HTML entities that would break the attribute
	// Note: We're using double quotes for the attribute, so escape double quotes
	value = strings.ReplaceAll(value, `&`, `&amp;`)
	value = strings.ReplaceAll(value, `"`, `&quot;`)
	value = strings.ReplaceAll(value, `<`, `&lt;`)
	value = strings.ReplaceAll(value, `>`, `&gt;`)

	return value
}

func registerComponents() {
	// Register components with the transformer
	componentDir := "examples/components"
	files, err := os.ReadDir(componentDir)
	if err != nil {
		log.Fatalf("Failed to read component directory: %v", err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".html") {
			componentName := strings.TrimSuffix(file.Name(), ".html")
			componentPath := fmt.Sprintf("%s/%s", componentDir, file.Name())
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

			// Extract props from the component template
			componentProps := extractComponentProps(componentAST)

			// Register the component with the transformer - both by name and by path
			transformer.RegisterComponent(componentName, componentAST, componentProps)

			// Also register with path for import resolution
			pathWithPrefix := fmt.Sprintf("./components/%s.html", componentName)
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

	// CRITICAL: Check if it's a function BEFORE trying JSON parsing
	// Functions start with "function " or contain "=>"
	if strings.HasPrefix(value, "function ") || strings.Contains(value, "=>") {
		// Return the function as-is (as a string, but buildXDataFromProps will detect it)
		return value
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
	if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
		(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
		return value[1 : len(value)-1]
	}

	// Default: return as string
	return value
}
