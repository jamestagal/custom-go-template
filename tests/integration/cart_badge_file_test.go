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
// to ensure we're not missing something in our test cases
func TestCartBadge_ActualFile(t *testing.T) {
	// Read the actual CartBadge.html file
	content, err := os.ReadFile("../../examples/components/CartBadge.html")
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
	markup, _, _ := renderer.RenderWithStores(tmpl, transformed, storeDefinitions, "CartBadge.html", "", "")

	// Print the rendered markup for manual inspection
	t.Logf("Rendered CartBadge.html:\n%s", markup)

	// Look for the x-text attribute
	expectedAttr := `x-text="'$' + $store.cart.total.toFixed(2)"`
	if !strings.Contains(markup, expectedAttr) {
		// Find what we actually got
		if idx := strings.Index(markup, `x-text="`); idx != -1 {
			// Find the closing quote
			endIdx := idx + len(`x-text="`)
			closeIdx := strings.Index(markup[endIdx:], `"`)
			if closeIdx != -1 {
				actualAttr := markup[idx : endIdx+closeIdx+1]
				t.Errorf("x-text attribute was modified!\nExpected: %s\nGot:      %s", expectedAttr, actualAttr)
			}
		} else {
			t.Error("x-text attribute not found in rendered output!")
		}
	}
}
