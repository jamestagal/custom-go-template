package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/jimafisk/custom_go_template/loader"
	"github.com/jimafisk/custom_go_template/transformer"
)

func main() {
	// Load the content JSON
	contentData, err := loader.LoadContentForRoute("/")
	if err != nil {
		log.Fatalf("Failed to load content: %v", err)
	}

	// Extract fields from first component
	var contentFields map[string]interface{}
	if loader.IsCollectionType(contentData) {
		if componentsRaw, ok := contentData["components"]; ok {
			if components, ok := componentsRaw.([]interface{}); ok && len(components) > 0 {
				if firstComp, ok := components[0].(map[string]interface{}); ok {
					if fields, ok := firstComp["fields"].(map[string]interface{}); ok {
						contentFields = fields
						fmt.Printf("✓ Extracted fields from first component: %d fields\n", len(fields))
						fmt.Println("Keys in contentFields map:")
						for key := range fields {
							fmt.Printf("  - %q\n", key)
						}
					}
				}
			}
		}
	}

	// Build the nested structure like renderWithWrapper does
	contentWithFields := map[string]interface{}{
		"fields": contentFields,
	}
	for key, val := range contentData {
		if key != "fields" {
			contentWithFields[key] = val
		}
	}

	fmt.Println("\nKeys in contentWithFields['fields'] map:")
	if fields, ok := contentWithFields["fields"].(map[string]interface{}); ok {
		for key := range fields {
			fmt.Printf("  - %q\n", key)
		}
	}

	// Build props like renderWithWrapper does
	props := map[string]interface{}{
		"content": contentWithFields,
	}

	// Add top-level props from fields (the workaround)
	if components, ok := contentData["components"].([]interface{}); ok && len(components) > 0 {
		if firstComp, ok := components[0].(map[string]interface{}); ok {
			if fields, ok := firstComp["fields"].(map[string]interface{}); ok {
				for key, value := range fields {
					props[key] = value
				}
				fmt.Printf("\n✓ Added %d fields as top-level props\n", len(fields))
			}
		}
	}

	fmt.Println("\nKeys in props map:")
	for key := range props {
		fmt.Printf("  - %q\n", key)
	}

	// Now format the props like buildXDataFromProps does
	fmt.Println("\n=== Testing FormatGoValueToJS ===")

	// Test formatting the content object
	contentFormatted := transformer.FormatGoValueToJS(props["content"])
	fmt.Println("\nFormatted content object:")
	fmt.Println(contentFormatted)

	// Test JSON marshal to see raw structure
	fmt.Println("\n=== JSON Marshal (for comparison) ===")
	jsonBytes, _ := json.MarshalIndent(props, "", "  ")
	fmt.Println(string(jsonBytes))
}
