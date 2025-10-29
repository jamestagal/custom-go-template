package utils

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// StripExportDefault removes "export default" from ES module syntax
// Converts: "export default { ... }" to "{ ... }"
//
// This allows store files to use proper ES module syntax (for better linter
// support and IDE tooling) while still being embedded directly into
// Alpine.store() calls in the rendered output.
//
// Example:
//   Input:  "export default {\n  isLoggedIn: false\n}"
//   Output: "{\n  isLoggedIn: false\n}"
//
// Pattern: String Transformation [Load: 3]
// Cognitive Load: 3 (prefix checking: 2, trimming: 1)
func StripExportDefault(content string) string {
	content = strings.TrimSpace(content)

	// Remove "export default " prefix (with trailing space)
	if strings.HasPrefix(content, "export default ") {
		content = strings.TrimPrefix(content, "export default ")
		return strings.TrimSpace(content)
	}

	// Remove "export default\n" (with newline)
	if strings.HasPrefix(content, "export default\n") {
		content = strings.TrimPrefix(content, "export default\n")
		return strings.TrimSpace(content)
	}

	return content
}

// RegisterStores loads all store files from the stores/ directory
// and returns a map of store name to store content (with export stripped).
//
// Store files must:
//  1. Be in the stores/ directory at project root
//  2. Have .js extension
//  3. Optionally use ES module syntax (export default { ... })
//
// Example:
//   stores/auth.js → map["auth"] = "{ isLoggedIn: false, ... }"
//
// Pattern: File Discovery Pattern [Load: 8]
// Cognitive Load: 8 (read dir: 2, filter: 2, read files: 2, map building: 2)
func RegisterStores() map[string]string {
	stores := make(map[string]string)
	storeDir := "stores"

	// Read the stores directory
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

		// Read store file content
		content, err := os.ReadFile(storePath)
		if err != nil {
			log.Printf("WARNING: Failed to read store file %s: %v", storePath, err)
			continue
		}

		// Strip ES module export syntax to get just the object literal
		storeContent := StripExportDefault(string(content))
		stores[storeName] = storeContent
		log.Printf("Loaded store: %s", storeName)
	}

	log.Printf("Registered %d store(s)", len(stores))
	return stores
}
