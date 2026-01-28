package integration

import (
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/parser"
	"github.com/jimafisk/custom_go_template/renderer"
	"github.com/jimafisk/custom_go_template/transformer"
)

// TestAlpineAttributePreservation tests that Alpine directive attributes are preserved unchanged
// BUG REPORT: x-text="'$' + $store.cart.total.toFixed(2)" was becoming x-text="$store.cart.total.toFixed"
func TestAlpineAttributePreservation(t *testing.T) {
	tests := []struct {
		name             string
		template         string
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name: "x-text with toFixed method",
			template: `---
store cart = {
	total: 0,
	items: []
}
---
<div>
  <strong x-text="'$' + $store.cart.total.toFixed(2)"></strong>
</div>`,
			shouldContain: []string{
				`x-text="'$' + $store.cart.total.toFixed(2)"`,
				`.toFixed(2)`,
				`'$' +`,
			},
			shouldNotContain: []string{
				`x-text="$store.cart.total.toFixed"`, // Missing (2) and '$' +
				`x-text="$store.cart.total"`,         // Too truncated
			},
		},
		{
			name: "x-text with template literal and Number()",
			template: `---
store cart = {
	total: 0
}
---
<div>
  <pre x-text="\` + "`" + `Total: $\${Number($store.cart.total).toFixed(2)}\` + "`" + `"></pre>
</div>`,
			shouldContain: []string{
				`Number($store.cart.total).toFixed(2)`,
				`.toFixed(2)`,
			},
			shouldNotContain: []string{
				`toFixed"`, // Missing (2)
			},
		},
		{
			name: "@click with method call and parameters",
			template: `---
store cart = {
	addItem(id, qty) {}
}
---
<button @click="$store.cart.addItem('item-123', 1)">Add</button>`,
			shouldContain: []string{
				`@click="$store.cart.addItem('item-123', 1)"`,
				`addItem('item-123', 1)`,
			},
			shouldNotContain: []string{
				`@click="$store.cart.addItem"`, // Missing parameters
			},
		},
		{
			name: "x-show with property access and comparison",
			template: `---
store cart = {
	items: []
}
---
<div x-show="$store.cart.items.length > 0">
  Cart has items
</div>`,
			shouldContain: []string{
				`x-show="$store.cart.items.length`,
				`.length`,
				`&gt; 0"`, // HTML entity for >
			},
			shouldNotContain: []string{
				`x-show="$store.cart.items"`, // Missing .length > 0
			},
		},
		{
			name: "x-text with complex arithmetic",
			template: `---
store cart = {
	total: 0
}
---
<span x-text="($store.cart.total * 0.15).toFixed(2)">Tax</span>`,
			shouldContain: []string{
				`x-text="($store.cart.total * 0.15).toFixed(2)"`,
				`.toFixed(2)`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the template
			tmpl, err := parser.ParseTemplate(tt.template)
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			// Transform the AST
			transformed := transformer.TransformAST(tmpl, nil)

			// Get tracked stores
			referencedStores, allDefinitions := transformer.GetTrackedStores(transformed)

			// Extract store definitions
			storeDefinitions := transformer.GetReferencedStoreDefinitions(allDefinitions, referencedStores)

			// Render to HTML
			markup, _, _ := renderer.RenderWithStores(tmpl, transformed, storeDefinitions, "test.html", "", nil)

			// Check that expected strings are in the output
			for _, expected := range tt.shouldContain {
				if !strings.Contains(markup, expected) {
					t.Errorf("Expected HTML to contain: %q\nGot: %s", expected, markup)
				}
			}

			// Check that unwanted strings are NOT in the output
			for _, unwanted := range tt.shouldNotContain {
				if strings.Contains(markup, unwanted) {
					t.Errorf("Expected HTML to NOT contain: %q\nGot: %s", unwanted, markup)
				}
			}
		})
	}
}

// TestAlpineAttributeStoreTracking verifies that store references in Alpine attributes are tracked
// even though the attributes themselves are not transformed
func TestAlpineAttributeStoreTracking(t *testing.T) {
	templateStr := `---
store cart = {
	total: 0
}
store theme = {
	mode: 'light'
}
---
<div>
  <span x-text="$store.cart.total.toFixed(2)"></span>
  <button @click="$store.theme.toggleDark()">Toggle</button>
</div>`

	// Parse
	tmpl, err := parser.ParseTemplate(templateStr)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Transform with store tracking
	transformed := transformer.TransformAST(tmpl, nil)

	// Get tracked stores
	trackedStores, _ := transformer.GetTrackedStores(transformed)

	// Verify both stores are tracked
	hasCart := false
	hasTheme := false

	for _, store := range trackedStores {
		if store == "cart" {
			hasCart = true
		}
		if store == "theme" {
			hasTheme = true
		}
	}

	if !hasCart {
		t.Error("Expected 'cart' store to be tracked")
	}
	if !hasTheme {
		t.Error("Expected 'theme' store to be tracked")
	}
}
