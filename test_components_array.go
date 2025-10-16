package main

import (
	"fmt"
	"strings"

	"github.com/jimafisk/custom_go_template/loader"
	"github.com/jimafisk/custom_go_template/transformer"
)

func main() {
	// Load content
	contentData, _ := loader.LoadContentForRoute("/")
	
	// Build contentWithFields like renderWithWrapper
	var contentFields map[string]interface{}
	if loader.IsCollectionType(contentData) {
		if componentsRaw, ok := contentData["components"]; ok {
			if components, ok := componentsRaw.([]interface{}); ok && len(components) > 0 {
				if firstComp, ok := components[0].(map[string]interface{}); ok {
					if fields, ok := firstComp["fields"].(map[string]interface{}); ok {
						contentFields = fields
					}
				}
			}
		}
	}
	
	contentWithFields := map[string]interface{}{
		"fields": contentFields,
	}
	// This is where components array gets added
	for key, val := range contentData {
		if key != "fields" {
			contentWithFields[key] = val
		}
	}
	
	// Check the keys in components[0].fields
	fmt.Println("=== Keys in contentWithFields['components'][0]['fields'] ===")
	if componentsRaw, ok := contentWithFields["components"]; ok {
		if components, ok := componentsRaw.([]interface{}); ok && len(components) > 0 {
			if firstComp, ok := components[0].(map[string]interface{}); ok {
				if fields, ok := firstComp["fields"].(map[string]interface{}); ok {
					for key := range fields {
						fmt.Printf("  - %q\n", key)
					}
				}
			}
		}
	}
	
	// Format it
	formatted := transformer.FormatGoValueToJS(contentWithFields)
	
	// Look for the bug pattern
	if strings.Contains(formatted, "'this.") {
		fmt.Println("\n❌ BUG FOUND!")
		// Find the section with the bug
		lines := strings.Split(formatted, ",")
		for _, line := range lines {
			if strings.Contains(line, "'this.") {
				fmt.Println(line)
			}
		}
	} else {
		fmt.Println("\n✓ No 'this.' prefix in keys")
	}
	
	// Show a preview
	fmt.Println("\n=== Formatted output (first 500 chars) ===")
	if len(formatted) > 500 {
		fmt.Println(formatted[:500] + "...")
	} else {
		fmt.Println(formatted)
	}
}
