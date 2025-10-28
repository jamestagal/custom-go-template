package integration

import (
	"os"
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/parser"
	"github.com/jimafisk/custom_go_template/renderer"
	"github.com/jimafisk/custom_go_template/transformer"
)

// TestCartBadge_ActualFile tests the ACTUAL CartBadge.html component file
// to ensure store expressions are properly transformed
func TestCartBadge_ActualFile(t *testing.T) {
	// Read the actual CartBadge.html file (it's in layouts/components, not examples/components)
	content, err := os.ReadFile("../../layouts/components/CartBadge.html")
	if err != nil {
		t.Fatalf("Failed to read CartBadge.html: %v", err)
	}

	templateStr := string(content)

	// Parse
	tmpl, err := parser.ParseTemplate(templateStr)
	if err != nil {
		t.Fatalf("Failed to parse CartBadge.html: %v", err)
	}

	// Transform
	transformed := transformer.TransformAST(tmpl, nil)

	// Get stores
	referencedStores, allDefinitions := transformer.GetTrackedStores(transformed)
	storeDefinitions := transformer.GetReferencedStoreDefinitions(allDefinitions, referencedStores)

	// Render
	markup, _, _ := renderer.RenderWithStores(tmpl, transformed, storeDefinitions, "CartBadge.html", "")

	// Print the rendered markup for manual inspection
	t.Logf("Rendered CartBadge.html:\n%s", markup)

	// UPDATED: The CartBadge component now uses formattedTotal from the store
	// (line 23: <strong x-text="$store.cart.formattedTotal"></strong>)
	// not .toFixed(2) method call
	expectedAttr := `x-text="$store.cart.formattedTotal"`
	if !strings.Contains(markup, expectedAttr) {
		t.Errorf("Expected x-text attribute not found!\nExpected: %s\nMarkup:\n%s", expectedAttr, markup)
	}

	// Verify store references are properly transformed
	if !strings.Contains(markup, "$store.cart.items.length") {
		t.Error("Store reference $store.cart.items.length not found")
	}

	// Verify the cart store is tracked
	found := false
	for _, store := range referencedStores {
		if store == "cart" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Cart store should be tracked, got: %v", referencedStores)
	}
}
