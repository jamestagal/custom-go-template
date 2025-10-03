package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	// Import the new renderer package
	"github.com/jimafisk/custom_go_template/ast"
	"github.com/jimafisk/custom_go_template/parser"
	"github.com/jimafisk/custom_go_template/renderer"
	"github.com/jimafisk/custom_go_template/transformer"
)

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

	// Set up the HTTP server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Serve static files from the public directory
		if r.URL.Path != "/" {
			http.ServeFile(w, r, publicDir+r.URL.Path)
			return
		}

		// Render the template - use test-basic.html for testing
		entrypoint := "examples/pages/home.html"

		// Read the template file to extract front matter data
		templateContent, err := os.ReadFile(entrypoint)
		if err != nil {
			log.Fatalf("Failed to read template file: %v", err)
		}

		// Parse the template to extract front matter
		template, err := parser.ParseTemplate(string(templateContent))
		if err != nil {
			log.Fatalf("Failed to parse template: %v", err)
		}

		// Extract variables from the fence section for Alpine.js data scope
		props := make(map[string]interface{})
		log.Printf("DEBUG: Extracting props from template with %d root nodes", len(template.RootNodes))
		for _, node := range template.RootNodes {
			if fence, ok := node.(*ast.FenceSection); ok {
				log.Printf("DEBUG: Found fence section with %d props and %d variables", len(fence.Props), len(fence.Variables))

				// Process variables
				for _, variable := range fence.Variables {
					props[variable.Name] = parseValue(variable.Value)
					log.Printf("DEBUG: Added variable %s = %v (type: %T)", variable.Name, props[variable.Name], props[variable.Name])
				}

				// Process props with default values
				for _, prop := range fence.Props {
					if _, exists := props[prop.Name]; !exists && prop.DefaultValue != "" {
						props[prop.Name] = parseValue(prop.DefaultValue)
						log.Printf("DEBUG: Added prop %s = %v (type: %T)", prop.Name, props[prop.Name], props[prop.Name])
					}
				}

				// Extract functions from the raw content
				// Functions are typically defined as: function name() { ... }
				functionRegex := regexp.MustCompile(`function\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*\([^)]*\)\s*{([^}]*)}`)
				matches := functionRegex.FindAllStringSubmatch(fence.RawContent, -1)
				for _, match := range matches {
					if len(match) >= 3 {
						fnName := match[1]
						fnBody := match[0] // Use the full function definition
						props[fnName] = fnBody
					}
				}

				// Also look for arrow functions and method shorthand
				arrowFnRegex := regexp.MustCompile(`(const\s+)?([a-zA-Z_$][a-zA-Z0-9_$]*)\s*=\s*(\([^)]*\)|[a-zA-Z_$][a-zA-Z0-9_$]*)\s*=>\s*{([^}]*)}`)
				arrowMatches := arrowFnRegex.FindAllStringSubmatch(fence.RawContent, -1)
				for _, match := range arrowMatches {
					if len(match) >= 3 {
						fnName := match[2]
						fnBody := match[0] // Use the full function definition
						props[fnName] = fnBody
					}
				}

				// Method shorthand: methodName() { ... }
				methodRegex := regexp.MustCompile(`([a-zA-Z_$][a-zA-Z0-9_$]*)\s*\([^)]*\)\s*{([^}]*)}`)
				methodMatches := methodRegex.FindAllStringSubmatch(fence.RawContent, -1)
				for _, match := range methodMatches {
					if len(match) >= 3 {
						fnName := match[1]
						// Skip if it's already captured by the function regex
						if _, exists := props[fnName]; !exists {
							fnBody := "function " + match[0] // Convert to function syntax
							props[fnName] = fnBody
						}
					}
				}

				break
			}
		}

		// Render the template with the extracted props
		markup, script, style := renderer.Render(entrypoint, props)

		// If the style is empty, try to extract it directly from the template file
		if style == "" {
			log.Println("Style is empty, extracting directly from template file")
			templateContent, err := os.ReadFile(entrypoint)
			if err == nil {
				// Extract style content between <style> tags
				styleRegex := regexp.MustCompile(`(?s)<style>(.*?)</style>`)
				styleMatches := styleRegex.FindAllStringSubmatch(string(templateContent), -1)
				if len(styleMatches) > 0 {
					for _, match := range styleMatches {
						if len(match) > 1 {
							style += match[1]
						}
					}
					log.Println("Extracted style from template file")
				}
			}
		}

		// Convert props to JSON for inline x-data
		propsJSON, err := json.Marshal(props)
		if err != nil {
			log.Printf("Warning: Failed to convert props to JSON: %v", err)
			propsJSON = []byte("{}")
		}

		// Simple inline Alpine.js data - no complex initialization needed
		alpineDataAttr := string(propsJSON)
		alpineInitScript := ""

		// Add CSS, JS links, and Alpine.js initialization to the HTML
		htmlWithLinks := addLinksToHTML(markup, alpineInitScript)

		// Add inline x-data to the body tag with actual data
		// CRITICAL: Only add to body tag, NOT as a wrapper div
		bodyTagRegex := regexp.MustCompile(`(?i)<body[^>]*>`)
		if bodyTagRegex.MatchString(htmlWithLinks) {
			htmlWithLinks = bodyTagRegex.ReplaceAllStringFunc(htmlWithLinks, func(match string) string {
				// Remove the closing > and add x-data
				tagWithoutClose := strings.TrimSuffix(match, ">")
				return fmt.Sprintf(`%s x-data='%s'>`, tagWithoutClose, alpineDataAttr)
			})
		} else {
			// Fallback: add x-data to html tag if no body tag
			htmlTagRegex := regexp.MustCompile(`(?i)<html[^>]*>`)
			if htmlTagRegex.MatchString(htmlWithLinks) {
				htmlWithLinks = htmlTagRegex.ReplaceAllStringFunc(htmlWithLinks, func(match string) string {
					tagWithoutClose := strings.TrimSuffix(match, ">")
					return fmt.Sprintf(`%s x-data='%s'>`, tagWithoutClose, alpineDataAttr)
				})
			}
		}

		// Write the output files to the public directory
		err = os.WriteFile(publicDir+"/script.js", []byte(script), 0644) // Use standard file permissions
		if err != nil {
			log.Fatalf("Failed to write script.js: %v", err)
		}

		// Make sure the style doesn't contain style tags
		style = strings.TrimSpace(style)
		style = strings.TrimPrefix(style, "<style>")
		style = strings.TrimSuffix(style, "</style>")

		err = os.WriteFile(publicDir+"/style.css", []byte(style), 0644)
		if err != nil {
			log.Fatalf("Failed to write style.css: %v", err)
		}

		err = os.WriteFile(publicDir+"/index.html", []byte(htmlWithLinks), 0644)
		if err != nil {
			log.Fatalf("Failed to write index.html: %v", err)
		}

		// Serve the index.html file
		http.ServeFile(w, r, publicDir+"/index.html")
	})

	// Start the server
	port := ":3333"
	fmt.Printf("Server starting on http://localhost%s\n", port)
	err = http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
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

func addLinksToHTML(html string, alpineInitScript string) string {
	// Find the closing head tag
	headCloseIndex := strings.Index(strings.ToLower(html), "</head>")
	if headCloseIndex == -1 {
		// If no head tag, find the opening body tag
		bodyOpenIndex := strings.Index(strings.ToLower(html), "<body")
		if bodyOpenIndex == -1 {
			// If no body tag either, just prepend to the HTML
			return `<link rel="stylesheet" href="/style.css">
<script defer src="/script.js"></script>
<script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>` +
				alpineInitScript + html
		}

		// Insert before the body tag
		return html[:bodyOpenIndex] + `<link rel="stylesheet" href="/style.css">
<script defer src="/script.js"></script>
<script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>` +
			alpineInitScript + html[bodyOpenIndex:]
	}

	// Insert before the closing head tag
	return html[:headCloseIndex] + `<link rel="stylesheet" href="/style.css">
<script defer src="/script.js"></script>
<script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>` +
		alpineInitScript + html[headCloseIndex:]
}

func parseValue(value string) interface{} {
	value = strings.TrimSpace(value)

	// Handle empty values
	if value == "" {
		return ""
	}

	// Try to parse as JSON first (handles arrays, objects, numbers, booleans, null)
	var jsonValue interface{}
	if err := json.Unmarshal([]byte(value), &jsonValue); err == nil {
		// Successfully parsed as JSON
		return jsonValue
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
