package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	// Import your renderer package (adjust the path to match your project)
	"github.com/jimafisk/custom_go_template/renderer"
)

func main() {
	// Get the absolute path to your project root
	projectRoot, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	// Ensure we're at the project root, not in cmd/example
	if filepath.Base(projectRoot) == "example" {
		projectRoot = filepath.Dir(filepath.Dir(projectRoot))
	}

	// Paths to example files
	examplesDir := filepath.Join(projectRoot, "examples")
	componentsDir := filepath.Join(examplesDir, "components")
	pagesDir := filepath.Join(examplesDir, "pages")
	log.Printf("Components directory: %s", componentsDir)
	// Create output directory for rendered HTML
	outputDir := filepath.Join(projectRoot, "public")
	os.MkdirAll(outputDir, 0755)

	// Set up file server for static files
	fs := http.FileServer(http.Dir(outputDir))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Example data
	data := map[string]any{
		"title":      "Custom Template Showcase",
		"isLoggedIn": true,
		"user": map[string]any{
			"name":  "John Doe",
			"role":  "admin",
			"email": "john@example.com",
		},
		"products": []map[string]any{
			{"id": 1, "name": "Laptop", "price": 999.99, "inStock": true, "featured": true, "tags": []string{"electronics", "computers"},
			{"id": 2, "name": "Phone", "price": 699.99, "inStock": true, "featured": false, "tags": []string{"electronics", "mobile"},
			{"id": 3, "name": "Headphones", "price": 149.99, "inStock": false, "featured": true, "tags": []string{"electronics", "audio"},
			{"id": 4, "name": "Tablet", "price": 499.99, "inStock": true, "featured": false, "tags": []string{"electronics", "computers"},
		},
	}

	// Handle the example page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// If requesting root, serve the comprehensive example
		if r.URL.Path == "/" {
			templatePath := filepath.Join(pagesDir, "comprehensive.html")
			html, script, style := renderer.Render(templatePath, data)

			// Create Alpine.js wrapper
			fullHtml := fmt.Sprintf(`
				<!DOCTYPE html>
				<html>
				<head>
					<meta charset="UTF-8">
					<meta name="viewport" content="width=device-width, initial-scale=1.0">
					<title>%s</title>
					<script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.14.8/dist/cdn.min.js"></script>
					<style>%s</style>
				</head>
				<body>
					%s
					<script>%s</script>
				</body>
				</html>
			`, data["title"], style, html, script)
			// Debug output - write the generated HTML to a file for inspection
			err = os.WriteFile("debug-output.html", []byte(fullHtml), 0644)
			if err != nil {
				log.Printf("Warning: Failed to write debug file: %v", err)
			}
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(fullHtml))
			return
		}

		// Handle 404
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Page not found"))
	})

	// Start the server
	port := ":8080"
	fmt.Printf("Server starting at http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
