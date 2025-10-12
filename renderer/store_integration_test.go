package renderer

import (
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/parser"
	"github.com/jimafisk/custom_go_template/transformer"
)

// TestRenderWithStoresIntegration tests the full pipeline from parsing to rendering
// with store initialization included in HTML output
func TestRenderWithStoresIntegration(t *testing.T) {
	tests := []struct {
		name           string
		template       string
		props          map[string]any
		wantStoreScript bool
		wantStores     []string // Store names that should be in the script
	}{
		{
			name: "Template with inline store definition and store expression",
			template: `---
store auth = { isLoggedIn: false }
---
<div>{$auth.isLoggedIn}</div>`,
			props:          map[string]any{},
			wantStoreScript: true,
			wantStores:     []string{"auth"},
		},
		{
			name: "Template with multiple stores",
			template: `---
store auth = { isLoggedIn: false }
store cart = { items: [], total: 0 }
---
<div>{$auth.isLoggedIn}</div>
<div>{$cart.total}</div>`,
			props:          map[string]any{},
			wantStoreScript: true,
			wantStores:     []string{"auth", "cart"},
		},
		{
			name: "Template without stores",
			template: `---
let count = 0
---
<div>{count}</div>`,
			props:          map[string]any{},
			wantStoreScript: false,
			wantStores:     []string{},
		},
		{
			name: "Template with store definition but no usage",
			template: `---
store auth = { isLoggedIn: false }
---
<div>Hello</div>`,
			props:          map[string]any{},
			wantStoreScript: false, // No store script if stores not referenced
			wantStores:     []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Step 1: Parse template
			template, err := parser.ParseTemplate(tt.template)
			if err != nil {
				t.Fatalf("ParseTemplate() error = %v", err)
			}

			// Step 2: Transform AST (this tracks store references)
			transformed := transformer.TransformAST(template, tt.props)

			// Step 3: Get tracked stores
			referencedStores, allDefinitions := transformer.GetTrackedStores(transformed)
			storeDefinitions := transformer.GetReferencedStoreDefinitions(allDefinitions, referencedStores)

			// Step 4: Render with stores
			markup, script, style := RenderWithStores(template, transformed, storeDefinitions, "test.html", "")

			// Validate store script presence
			if tt.wantStoreScript {
				if !strings.Contains(script, "document.addEventListener('alpine:init'") {
					t.Errorf("Expected store initialization script, but not found in:\n%s", script)
				}

				// Check that all expected stores are initialized
				for _, storeName := range tt.wantStores {
					expectedStoreCall := "Alpine.store('" + storeName + "'"
					if !strings.Contains(script, expectedStoreCall) {
						t.Errorf("Expected Alpine.store('%s') call in script, but not found:\n%s", storeName, script)
					}
				}
			} else {
				// No store script expected
				if strings.Contains(script, "document.addEventListener('alpine:init'") {
					t.Errorf("Did not expect store initialization script, but found:\n%s", script)
				}
			}

			// Validate markup contains transformed store expressions
			if len(tt.wantStores) > 0 {
				// Should have x-text with $store.storeName.property
				if !strings.Contains(markup, "x-text=") {
					t.Errorf("Expected x-text directive in markup for store expression")
				}
			}

			// Style should be empty for these simple tests
			_ = style
		})
	}
}

// TestRenderHTMLStructureWithStores tests the complete HTML output structure
// to ensure store script is placed correctly relative to Alpine.js CDN
func TestRenderHTMLStructureWithStores(t *testing.T) {
	template := `---
store auth = { isLoggedIn: false }
---
<html>
<head>
  <title>Test</title>
</head>
<body>
  <div>{$auth.isLoggedIn}</div>
</body>
</html>`

	// Parse and transform
	parsedTemplate, err := parser.ParseTemplate(template)
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	transformed := transformer.TransformAST(parsedTemplate, map[string]any{})
	referencedStores, allDefinitions := transformer.GetTrackedStores(transformed)
	storeDefinitions := transformer.GetReferencedStoreDefinitions(allDefinitions, referencedStores)

	// Render
	markup, script, _ := RenderWithStores(parsedTemplate, transformed, storeDefinitions, "test.html", "")

	// The script should contain store initialization
	if !strings.Contains(script, "Alpine.store('auth'") {
		t.Errorf("Expected Alpine.store('auth') in script, got:\n%s", script)
	}

	// Markup should contain the HTML structure
	if !strings.Contains(markup, "<html>") {
		t.Errorf("Expected <html> in markup")
	}
}

// TestEmptyStoresNoScript tests that empty store map doesn't generate script
func TestEmptyStoresNoScript(t *testing.T) {
	template := `<div>Hello</div>`

	parsedTemplate, err := parser.ParseTemplate(template)
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	transformed := transformer.TransformAST(parsedTemplate, map[string]any{})

	// Render with empty stores
	_, script, _ := RenderWithStores(parsedTemplate, transformed, map[string]string{}, "test.html", "")

	// Script should be empty or not contain store initialization
	if strings.Contains(script, "Alpine.store") {
		t.Errorf("Expected no store initialization for empty stores, got:\n%s", script)
	}
}

// TestStoreScriptOrderInHTML tests that the full HTML assembly
// places store script after Alpine.js CDN but before component content
// This test simulates what the server does
func TestStoreScriptOrderInHTML(t *testing.T) {
	template := `---
store auth = { isLoggedIn: false }
---
<html>
<head>
</head>
<body>
<div>{$auth.isLoggedIn}</div>
</body>
</html>`

	parsedTemplate, err := parser.ParseTemplate(template)
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	transformed := transformer.TransformAST(parsedTemplate, map[string]any{})
	referencedStores, allDefinitions := transformer.GetTrackedStores(transformed)
	storeDefinitions := transformer.GetReferencedStoreDefinitions(allDefinitions, referencedStores)

	markup, script, _ := RenderWithStores(parsedTemplate, transformed, storeDefinitions, "test.html", "")

	// Build final HTML like the server does
	finalHTML := markup

	// Add Alpine.js CDN to head
	alpineCDN := `<script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>`
	if strings.Contains(finalHTML, "</head>") {
		finalHTML = strings.Replace(finalHTML, "</head>", alpineCDN+"</head>", 1)
	}

	// Add store script after Alpine CDN (in head)
	if script != "" && strings.Contains(script, "Alpine.store") {
		storeScript := "<script>\n" + script + "\n</script>"
		// Replace the first </head> (after Alpine CDN)
		finalHTML = strings.Replace(finalHTML, "</head>", storeScript+"</head>", 1)
	}

	// Verify order: Alpine CDN appears before store script
	alpineIdx := strings.Index(finalHTML, "alpinejs")
	storeIdx := strings.Index(finalHTML, "Alpine.store")

	if alpineIdx == -1 {
		t.Error("Alpine.js CDN not found in final HTML")
	}

	if storeIdx == -1 {
		t.Error("Store initialization not found in final HTML")
	}

	if alpineIdx > storeIdx {
		t.Error("Alpine.js CDN should appear before store initialization")
	}

	// Verify both are in <head>
	headEnd := strings.Index(finalHTML, "</head>")
	if storeIdx > headEnd {
		t.Error("Store initialization should be in <head>")
	}
}

// TestStoreInConditional tests store expressions in conditionals
func TestStoreInConditional(t *testing.T) {
	template := `---
store auth = { isLoggedIn: false }
---
{if $auth.isLoggedIn}
  <div>Welcome!</div>
{/if}`

	parsedTemplate, err := parser.ParseTemplate(template)
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	transformed := transformer.TransformAST(parsedTemplate, map[string]any{})
	referencedStores, allDefinitions := transformer.GetTrackedStores(transformed)
	storeDefinitions := transformer.GetReferencedStoreDefinitions(allDefinitions, referencedStores)

	markup, script, _ := RenderWithStores(parsedTemplate, transformed, storeDefinitions, "test.html", "")

	// Check store script
	if !strings.Contains(script, "Alpine.store('auth'") {
		t.Errorf("Expected Alpine.store('auth') in script")
	}

	// Check transformed conditional
	if !strings.Contains(markup, "x-if") {
		t.Errorf("Expected x-if directive in markup")
	}

	if !strings.Contains(markup, "$store.auth.isLoggedIn") {
		t.Errorf("Expected $store.auth.isLoggedIn in conditional, got:\n%s", markup)
	}
}

// TestStoreInLoop tests store expressions in loops
func TestStoreInLoop(t *testing.T) {
	template := `---
store cart = { items: [] }
---
{for item in $cart.items}
  <div>{item}</div>
{/for}`

	parsedTemplate, err := parser.ParseTemplate(template)
	if err != nil {
		t.Fatalf("ParseTemplate() error = %v", err)
	}

	transformed := transformer.TransformAST(parsedTemplate, map[string]any{})
	referencedStores, allDefinitions := transformer.GetTrackedStores(transformed)
	storeDefinitions := transformer.GetReferencedStoreDefinitions(allDefinitions, referencedStores)

	markup, script, _ := RenderWithStores(parsedTemplate, transformed, storeDefinitions, "test.html", "")

	// Check store script
	if !strings.Contains(script, "Alpine.store('cart'") {
		t.Errorf("Expected Alpine.store('cart') in script")
	}

	// Check transformed loop
	if !strings.Contains(markup, "x-for") {
		t.Errorf("Expected x-for directive in markup")
	}

	if !strings.Contains(markup, "$store.cart.items") {
		t.Errorf("Expected $store.cart.items in loop, got:\n%s", markup)
	}
}

// Cognitive Load Analysis:
// - TestRenderWithStoresIntegration: 18 (parse: 3, transform: 3, tracking: 4, render: 3, assertions: 5)
// - TestRenderHTMLStructureWithStores: 12 (parse: 3, transform: 3, render: 3, assertions: 3)
// - TestEmptyStoresNoScript: 8 (parse: 2, transform: 2, render: 2, assertion: 2)
// - TestStoreScriptOrderInHTML: 20 (parse: 3, transform: 3, render: 3, HTML building: 6, assertions: 5)
// - TestStoreInConditional: 12 (parse: 3, transform: 3, render: 3, assertions: 3)
// - TestStoreInLoop: 12 (parse: 3, transform: 3, render: 3, assertions: 3)
// Total File Load: 82 (within acceptable range for test files)
// Individual function loads: All < 25 ✓
