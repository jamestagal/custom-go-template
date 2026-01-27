package loader

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Common top-level content files that live in content/ root (not pages/)
var topLevelContentPages = map[string]bool{
	"about":   true,
	"contact": true,
	"privacy": true,
	"terms":   true,
}

// LoadContentJSON loads JSON from a file path and returns it as a map.
// Returns empty map if file doesn't exist (not an error).
// Returns error only for invalid JSON or read failures.
//
// Cognitive Load: 8
func LoadContentJSON(filePath string) (map[string]interface{}, error) {
	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Missing file is not an error - return empty map
			log.Printf("LoadContentJSON: file not found %s (returning empty map)", filePath)
			return make(map[string]interface{}), nil
		}
		return nil, fmt.Errorf("LoadContentJSON: failed to read file %s: %w", filePath, err)
	}

	// Parse JSON
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("LoadContentJSON: invalid JSON in %s: %w", filePath, err)
	}

	return result, nil
}

// RoutePathToFilePath converts a route path to a content file path.
// Examples:
//   - "/store-demo" -> "content/pages/store-demo.json"
//   - "/" -> "content/_index.json"
//   - "/about" -> "content/about.json" (common top-level page)
//   - "/pages/example" -> "content/pages/example.json"
//
// Cognitive Load: 7
func RoutePathToFilePath(routePath string) string {
	// Remove leading slash
	routePath = strings.TrimPrefix(routePath, "/")

	// Handle root route
	if routePath == "" {
		return "content/pages/_index.json"
	}

	// Check if route already contains "pages/"
	if strings.HasPrefix(routePath, "pages/") {
		return fmt.Sprintf("content/%s.json", routePath)
	}

	// Check if this is a top-level route (no slashes)
	if !strings.Contains(routePath, "/") {
		// Check if it's a common top-level page (about, contact, etc.)
		if topLevelContentPages[routePath] {
			return fmt.Sprintf("content/%s.json", routePath)
		}

		// Otherwise default to pages/ subdirectory
		return fmt.Sprintf("content/pages/%s.json", routePath)
	}

	// For nested routes, use as-is under content/
	return fmt.Sprintf("content/%s.json", routePath)
}

// ExtractComponentFields extracts fields for a specific component from collection type JSON.
// If the JSON has a "components" array, it searches for a component with matching name
// and returns its "fields" object.
// Returns empty map if component not found or if not a collection type.
//
// Example:
//
//	data := map[string]interface{}{
//	  "components": []interface{}{
//	    map[string]interface{}{
//	      "name": "page_header",
//	      "fields": map[string]interface{}{"title": "Hello"},
//	    },
//	  },
//	}
//	fields := ExtractComponentFields(data, "page_header")
//	// returns: map[string]interface{}{"title": "Hello"}
//
// Cognitive Load: 12
func ExtractComponentFields(data map[string]interface{}, componentName string) map[string]interface{} {
	// Check if this is a collection type
	componentsRaw, ok := data["components"]
	if !ok {
		return make(map[string]interface{})
	}

	// Type assert to array
	components, ok := componentsRaw.([]interface{})
	if !ok {
		log.Printf("ExtractComponentFields: 'components' is not an array, got %T", componentsRaw)
		return make(map[string]interface{})
	}

	// Search for component by name
	for _, compRaw := range components {
		comp, ok := compRaw.(map[string]interface{})
		if !ok {
			log.Printf("ExtractComponentFields: component item is not a map, got %T", compRaw)
			continue
		}

		// Check name
		nameRaw, ok := comp["name"]
		if !ok {
			continue
		}

		name, ok := nameRaw.(string)
		if !ok {
			log.Printf("ExtractComponentFields: component name is not a string, got %T", nameRaw)
			continue
		}

		if name != componentName {
			continue
		}

		// Found matching component - extract fields
		fieldsRaw, ok := comp["fields"]
		if !ok {
			return make(map[string]interface{})
		}

		fields, ok := fieldsRaw.(map[string]interface{})
		if !ok {
			log.Printf("ExtractComponentFields: fields is not a map, got %T", fieldsRaw)
			return make(map[string]interface{})
		}

		return fields
	}

	// Component not found
	return make(map[string]interface{})
}

// IsCollectionType checks if the JSON data represents a collection type (has "components" array)
// or a single type (flat fields).
//
// Cognitive Load: 3
func IsCollectionType(data map[string]interface{}) bool {
	_, ok := data["components"]
	return ok
}

// LoadContentForRoute loads content JSON for a given route path.
// Combines RoutePathToFilePath and LoadContentJSON for convenience.
//
// Example:
//
//	data, err := LoadContentForRoute("/store-demo")
//	// Loads from content/pages/store-demo.json
//
// Cognitive Load: 5
func LoadContentForRoute(routePath string) (map[string]interface{}, error) {
	filePath := RoutePathToFilePath(routePath)
	data, err := LoadContentJSON(filePath)
	if err != nil {
		return nil, fmt.Errorf("LoadContentForRoute %s: %w", routePath, err)
	}
	return data, nil
}

// LoadComponentContent loads content for a specific component from a route.
// Convenience function that combines LoadContentForRoute and ExtractComponentFields.
//
// Example:
//
//	fields, err := LoadComponentContent("/store-demo", "page_header")
//	// Loads content/pages/store-demo.json and extracts page_header fields
//
// Cognitive Load: 7
func LoadComponentContent(routePath, componentName string) (map[string]interface{}, error) {
	data, err := LoadContentForRoute(routePath)
	if err != nil {
		return nil, fmt.Errorf("LoadComponentContent: %w", err)
	}

	if !IsCollectionType(data) {
		// For single type, return all data as component fields
		return data, nil
	}

	fields := ExtractComponentFields(data, componentName)
	return fields, nil
}

// ExtractAllComponentNames extracts all component names from JSON content data.
// If the JSON has a "components" array, it returns all component names.
// Returns empty slice if no components found or if not a collection type.
//
// Example:
//
//	data := map[string]interface{}{
//	  "components": []interface{}{
//	    map[string]interface{}{"name": "hero2436", "fields": {...}},
//	    map[string]interface{}{"name": "services2437", "fields": {...}},
//	  },
//	}
//	names := ExtractAllComponentNames(data)
//	// returns: []string{"hero2436", "services2437"}
//
// Cognitive Load: 8
func ExtractAllComponentNames(data map[string]interface{}) []string {
	if data == nil {
		return []string{}
	}

	// Check if this is a collection type
	componentsRaw, ok := data["components"]
	if !ok {
		return []string{}
	}

	// Type assert to array
	components, ok := componentsRaw.([]interface{})
	if !ok {
		log.Printf("ExtractAllComponentNames: 'components' is not an array, got %T", componentsRaw)
		return []string{}
	}

	// Collect all component names
	names := make([]string, 0, len(components))
	for _, compRaw := range components {
		comp, ok := compRaw.(map[string]interface{})
		if !ok {
			log.Printf("ExtractAllComponentNames: component item is not a map, got %T", compRaw)
			continue
		}

		// Get name
		nameRaw, ok := comp["name"]
		if !ok {
			continue
		}

		name, ok := nameRaw.(string)
		if !ok {
			log.Printf("ExtractAllComponentNames: component name is not a string, got %T", nameRaw)
			continue
		}

		names = append(names, name)
	}

	return names
}

// GetAbsoluteContentPath converts a relative content path to absolute path.
// Useful for resolving paths from the project root.
//
// Cognitive Load: 4
func GetAbsoluteContentPath(relativePath string) (string, error) {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("GetAbsoluteContentPath: failed to get working directory: %w", err)
	}

	// Join with relative path
	absPath := filepath.Join(cwd, relativePath)
	return absPath, nil
}
