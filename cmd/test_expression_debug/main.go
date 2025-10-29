package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jimafisk/custom_go_template/ast"
	"github.com/jimafisk/custom_go_template/parser"
	"github.com/jimafisk/custom_go_template/transformer"
)

func main() {
	// Enable debug logging
	os.Setenv("DEBUG_EXPRESSIONS", "true")
	log.SetFlags(0) // Remove timestamp for cleaner output

	fmt.Println("=== Expression Transformation Debugging Demo ===")
	fmt.Println()

	// Test template with various expression types
	templateContent := `---
let title = "Test Page"
let description = "A powerful template engine"
let count = 42
prop type = "info"
---

<div content="{description}">BUILD-TIME: Simple variable</div>
<div class="{type}">RUNTIME: Prop not in fence variables</div>
<div data-value="{count + 10}">RUNTIME: Complex expression</div>
<div data-items="{items}">RUNTIME: Variable not in dataScope</div>
`

	// Parse template
	template, err := parser.ParseTemplate(templateContent)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

	// Extract fence data
	var fence *ast.FenceSection
	for _, node := range template.RootNodes {
		if f, ok := node.(*ast.FenceSection); ok {
			fence = f
			break
		}
	}

	// Build initial data scope
	dataScope := make(map[string]interface{})
	if fence != nil {
		for _, variable := range fence.Variables {
			dataScope[variable.Name] = variable.Value
		}
		// Add props with defaults
		for _, prop := range fence.Props {
			if prop.DefaultValue != "" {
				dataScope[prop.Name] = prop.DefaultValue
			}
		}
	}

	fmt.Println("Data Scope:")
	for k, v := range dataScope {
		fmt.Printf("  %s = %v (type: %T)\n", k, v, v)
	}
	fmt.Println()

	// Transform the template (this will trigger debug logging)
	fmt.Println("Transforming template...")
	fmt.Println()
	transformer.TransformAST(template, dataScope)

	fmt.Println()
	fmt.Println("=== Debug logging complete ===")
	fmt.Println()
	fmt.Println("Key:")
	fmt.Println("  BUILD-TIME: Expression resolved at build time (static value)")
	fmt.Println("  RUNTIME: Expression needs Alpine.js runtime binding (dynamic value)")
}
