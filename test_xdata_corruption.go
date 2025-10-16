package main

import (
	"fmt"
	"strings"

	"github.com/jimafisk/custom_go_template/loader"
	"github.com/jimafisk/custom_go_template/parser"
	"github.com/jimafisk/custom_go_template/transformer"
)

func main() {
	// Load content
	contentData, _ := loader.LoadContentForRoute("/")
	
	// Build props like renderWithWrapper does
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
	for key, val := range contentData {
		if key != "fields" {
			contentWithFields[key] = val
		}
	}
	
	props := map[string]interface{}{
		"content": contentWithFields,
	}
	
	fmt.Println("=== BEFORE TransformAST ===")
	fmt.Println("Keys in props['content']['fields']:")
	if content, ok := props["content"].(map[string]interface{}); ok {
		if fields, ok := content["fields"].(map[string]interface{}); ok {
			for key := range fields {
				fmt.Printf("  - %q\n", key)
			}
		}
	}
	
	// Parse a simple template
	templateContent := `---
export let content
---
<div>test</div>`
	template, _ := parser.ParseTemplate(templateContent)
	
	// Transform (this is where corruption might happen)
	_ = transformer.TransformAST(template, props)
	
	fmt.Println("\n=== AFTER TransformAST ===")
	fmt.Println("Keys in props['content']['fields']:")
	if content, ok := props["content"].(map[string]interface{}); ok {
		if fields, ok := content["fields"].(map[string]interface{}); ok {
			for key := range fields {
				fmt.Printf("  - %q\n", key)
			}
		}
	}
	
	// Now format for x-data
	fmt.Println("\n=== FormatGoValueToJS Output ===")
	formatted := transformer.FormatGoValueToJS(props["content"])
	
	// Check if any keys have 'this.' prefix
	if strings.Contains(formatted, "'this.") {
		fmt.Println("❌ BUG FOUND: Keys contain 'this.' prefix!")
		fmt.Println(formatted)
	} else {
		fmt.Println("✓ No corruption detected")
		fmt.Println(formatted[:200] + "...")
	}
}
