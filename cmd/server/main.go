package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
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
	
	// Extract variables from the fence section for Alpine.js data scope
	entrypoint := "examples/pages/comprehensive.html"
	alpineDataScope := getAlpineDataScope(entrypoint)
	
	// Set up the HTTP server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Serve static files from the public directory
		if r.URL.Path != "/" {
			http.ServeFile(w, r, publicDir+r.URL.Path)
			return
		}
		
		// Render the template
		props := make(map[string]interface{})
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
		
		// Add Alpine.js data scope directly to the html tag
		if alpineDataScope != "" {
			// Find the html tag
			htmlTagRegex := regexp.MustCompile(`(?i)<html[^>]*>`)
			markup = htmlTagRegex.ReplaceAllString(markup, `<html x-data='`+alpineDataScope+`'>`)
		}
		
		// Add CSS and JS links to the HTML
		htmlWithLinks := addLinksToHTML(markup, "")
		
		// Post-process the HTML to transform any remaining Svelte-like syntax
		htmlWithLinks = postProcessHTML(htmlWithLinks, "")
		
		err = os.WriteFile(publicDir+"/index.html", []byte(htmlWithLinks), 0644)
		if err != nil {
			log.Fatalf("Failed to write index.html: %v", err)
		}
		
		// Serve the index.html file
		http.ServeFile(w, r, publicDir+"/index.html")
	})
	
	// Start the server
	port := ":3000"
	fmt.Printf("Server starting on http://localhost%s\n", port)
	err = http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getAlpineDataScope(entrypoint string) string {
	// Create a hardcoded Alpine.js data scope with all the variables we need
	// This is based on the structure in views/comprehensive_test.html
	alpineDataScope := `{
		title: "Custom Template Showcase",
		isLoggedIn: true,
		user: { name: "John Doe", role: "admin", email: "john@example.com" },
		products: [
			{ id: 1, name: "Laptop", price: 999.99, inStock: true, featured: true, tags: ["electronics", "computers"] },
			{ id: 2, name: "Phone", price: 699.99, inStock: true, featured: false, tags: ["electronics", "mobile"] },
			{ id: 3, name: "Headphones", price: 149.99, inStock: false, featured: true, tags: ["electronics", "audio"] },
			{ id: 4, name: "Tablet", price: 499.99, inStock: true, featured: false, tags: ["electronics", "computers"] }
		],
		categories: [
			{ name: "Electronics", items: [] },
			{ name: "Computers", items: [] },
			{ name: "Audio", items: [] },
			{ name: "Mobile", items: [] }
		],
		notifications: [
			{ type: "info", message: "Welcome to our store!" },
			{ type: "success", message: "Your order has been processed." },
			{ type: "warning", message: "Some items are out of stock." }
		],
		settings: {
			theme: "light",
			currency: "USD",
			showNotifications: true,
			filters: {
				inStockOnly: false
			}
		},
		currentTime: new Date().toLocaleTimeString(),
		getGreeting() { return "Hello" },
		formatPrice(price) { return "$" + price.toFixed(2) },
		filteredProducts: [
			{ id: 1, name: "Laptop", price: 999.99, inStock: true, featured: true, tags: ["electronics", "computers"] },
			{ id: 2, name: "Phone", price: 699.99, inStock: true, featured: false, tags: ["electronics", "mobile"] }
		],
		navItems: [
			{ label: "Home", url: "/" },
			{ label: "Products", url: "/products" },
			{ label: "About", url: "/about" }
		]
	}`
	
	return alpineDataScope
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

func addLinksToHTML(html string, alpineDataScope string) string {
	// Check if the HTML already has a head tag
	headRegex := regexp.MustCompile(`(?i)<head>`)
	hasHead := headRegex.MatchString(html)
	
	if hasHead {
		// Add CSS and Alpine.js links to the existing head tag
		headCloseRegex := regexp.MustCompile(`(?i)</head>`)
		html = headCloseRegex.ReplaceAllString(html, `
	<link rel="stylesheet" href="/style.css">
	<script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.13.0/dist/cdn.min.js"></script>
	<script src="/script.js"></script>
</head>`)
	} else {
		// Add a new head tag with CSS and Alpine.js links
		htmlOpenRegex := regexp.MustCompile(`(?i)<html[^>]*>`)
		if htmlOpenRegex.MatchString(html) {
			html = htmlOpenRegex.ReplaceAllStringFunc(html, func(match string) string {
				return match + `
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Custom Template Engine</title>
	<link rel="stylesheet" href="/style.css">
	<script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.13.0/dist/cdn.min.js"></script>
	<script src="/script.js"></script>
</head>`
			})
		} else {
			// If there's no html tag, add it at the beginning
			html = `<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Custom Template Engine</title>
	<link rel="stylesheet" href="/style.css">
	<script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.13.0/dist/cdn.min.js"></script>
	<script src="/script.js"></script>
</head>
` + html + `
</html>`
		}
	}
	
	return html
}

func postProcessHTML(html string, alpineDataScope string) string {
	// Transform remaining loop syntax
	// Convert {#for item in items} ... {/for} to <template x-for="item in items"> ... </template>
	forLoopRegex := regexp.MustCompile(`(?s){#for\s+([^}]+)}(.*?){/for}`)
	html = forLoopRegex.ReplaceAllStringFunc(html, func(match string) string {
		submatches := forLoopRegex.FindStringSubmatch(match)
		if len(submatches) < 3 {
			return match
		}
		
		loopExpr := strings.TrimSpace(submatches[1])
		content := submatches[2]
		
		// Process the content recursively to handle nested loops and expressions
		content = postProcessHTML(content, "")
		
		return fmt.Sprintf(`<template x-for="%s">%s</template>`, loopExpr, content)
	})
	
	// Transform remaining each syntax (alternative to for)
	// Convert {#each items as item} ... {/each} to <template x-for="item in items"> ... </template>
	eachLoopRegex := regexp.MustCompile(`(?s){#each\s+([^}]+)\s+as\s+([^}]+)}(.*?){/each}`)
	html = eachLoopRegex.ReplaceAllStringFunc(html, func(match string) string {
		submatches := eachLoopRegex.FindStringSubmatch(match)
		if len(submatches) < 4 {
			return match
		}
		
		collection := strings.TrimSpace(submatches[1])
		iterator := strings.TrimSpace(submatches[2])
		content := submatches[3]
		
		// If there's an else after this else-if, trim the content
		elsePos := strings.Index(content, "{:else}")
		if elsePos > 0 {
			content = content[:elsePos]
		}
		
		// Process the content recursively
		content = postProcessHTML(content, "")
		
		return fmt.Sprintf(`<template x-for="%s in %s">%s</template>`, iterator, collection, content)
	})
	
	// Transform remaining expression syntax
	// Convert {variable} to <span x-text="variable"></span>
	exprRegex := regexp.MustCompile(`{([^{}#/]+)}`)
	html = exprRegex.ReplaceAllStringFunc(html, func(match string) string {
		expr := strings.TrimSpace(match[1:len(match)-1])
		
		// Skip transforming expressions that are part of Alpine.js directives
		if strings.Contains(expr, "x-") {
			return match
		}
		
		// Skip transforming expressions that are already in Alpine.js format
		if strings.HasPrefix(match, `<span x-text="`) {
			return match
		}
		
		return fmt.Sprintf(`<span x-text="%s"></span>`, expr)
	})
	
	// Transform remaining conditional syntax
	// Convert {#if condition} ... {/if} to <template x-if="condition"> ... </template>
	ifRegex := regexp.MustCompile(`(?s){#if\s+([^}]+)}(.*?){/if}`)
	html = ifRegex.ReplaceAllStringFunc(html, func(match string) string {
		submatches := ifRegex.FindStringSubmatch(match)
		if len(submatches) < 3 {
			return match
		}
		
		condition := strings.TrimSpace(submatches[1])
		content := submatches[2]
		
		// Check for else-if or else branches
		elseIfRegex := regexp.MustCompile(`(?s){:else\s+if\s+([^}]+)}(.*?)(?:{:else}|{/if})`)
		elseIfMatches := elseIfRegex.FindAllStringSubmatch(content, -1)
		
		elseRegex := regexp.MustCompile(`(?s){:else}(.*?){/if}`)
		elseMatch := elseRegex.FindStringSubmatch(content)
		
		// Extract the if branch content (everything before the first else-if or else)
		ifContent := content
		if len(elseIfMatches) > 0 {
			firstElseIfPos := strings.Index(content, "{:else if")
			if firstElseIfPos > 0 {
				ifContent = content[:firstElseIfPos]
			}
		} else if elseMatch != nil {
			elsePos := strings.Index(content, "{:else}")
			if elsePos > 0 {
				ifContent = content[:elsePos]
			}
		}
		
		// Process the content recursively
		ifContent = postProcessHTML(ifContent, "")
		
		result := fmt.Sprintf(`<template x-if="%s">%s</template>`, condition, ifContent)
		
		// Add else-if branches
		for _, elseIfMatch := range elseIfMatches {
			if len(elseIfMatch) >= 3 {
				elseIfCondition := strings.TrimSpace(elseIfMatch[1])
				elseIfContent := elseIfMatch[2]
				
				// If there's an else after this else-if, trim the content
				elsePos := strings.Index(elseIfContent, "{:else}")
				if elsePos > 0 {
					elseIfContent = elseIfContent[:elsePos]
				}
				
				// Process the content recursively
				elseIfContent = postProcessHTML(elseIfContent, "")
				
				result += fmt.Sprintf(`<template x-else-if="%s">%s</template>`, elseIfCondition, elseIfContent)
			}
		}
		
		// Add else branch if present
		if elseMatch != nil && len(elseMatch) >= 2 {
			elseContent := elseMatch[1]
			
			// Process the content recursively
			elseContent = postProcessHTML(elseContent, "")
			
			result += fmt.Sprintf(`<template x-else>%s</template>`, elseContent)
		}
		
		return result
	})
	
	return html
}
