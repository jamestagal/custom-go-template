package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	// Import the new renderer package
	"github.com/jimafisk/custom_go_template/renderer"
)

func main() {
	// Ensure public directory exists
	publicDir := "./public" // Use a variable for clarity
	err := os.MkdirAll(publicDir, os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create public directory: %v", err)
	}

	// Define initial props
	props := map[string]any{"name": "John", "age": 2, "animals": []string{"cat", "dog", "pig"}}

	// Render the main template using the renderer package
	// Assuming views are relative to the execution directory (where go run ./cmd/server is executed)
	// Adjust path if needed, e.g., "../views/home.html" if running from cmd/server/
	// entrypoint := "views/test_basic.html" // Changed to render the test file
	entrypoint := "views/comprehensive_test.html"
	markup, script, style := renderer.Render(entrypoint, props)

	// Write the output files to the public directory
	err = os.WriteFile(publicDir+"/script.js", []byte(script), 0644) // Use standard file permissions
	if err != nil {
		log.Fatalf("Failed to write script.js: %v", err)
	}
	err = os.WriteFile(publicDir+"/style.css", []byte(style), 0644)
	if err != nil {
		log.Fatalf("Failed to write style.css: %v", err)
	}
	err = os.WriteFile(publicDir+"/index.html", []byte(markup), 0644)
	if err != nil {
		log.Fatalf("Failed to write index.html: %v", err)
	}

	// Set up the static file server
	fs := http.FileServer(http.Dir(publicDir))
	http.Handle("/", fs)

	// Start the server
	port := ":3000"
	fmt.Printf("Server starting on http://localhost%s\n", port)
	err = http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
