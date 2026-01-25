package integration

import (
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/parser"
	"github.com/jimafisk/custom_go_template/renderer"
	"github.com/jimafisk/custom_go_template/transformer"
)

// TestCartBadge_ToFixedBug reproduces the exact bug reported:
// x-text="'$' + $store.cart.total.toFixed(2)" becomes x-text="$store.cart.total.toFixed"
func TestCartBadge_ToFixedBug(t *testing.T) {
	// Exact template from CartBadge.html
	templateStr := `---
store cart = {
	items: [],
	total: 0
}
---
<div class="cart-badge">
  <div class="cart-icon-wrapper">
    🛒
    {if $cart.items.length > 0}
      <span class="cart-count">{$cart.items.length}</span>
    {/if}
  </div>
  <div class="cart-details">
    <div class="cart-items">
      {if $cart.items.length === 0}
        <span class="empty-cart">Cart is empty</span>
      {else}
        <span class="item-count">{$cart.items.length} items</span>
      {/if}
    </div>
    {if $cart.items.length > 0}
      <div class="cart-total">
        Total: <strong x-text="'$' + $store.cart.total.toFixed(2)"></strong>
      </div>
    {/if}
  </div>
</div>`

	// Parse
	tmpl, err := parser.ParseTemplate(templateStr)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Transform
	transformed := transformer.TransformAST(tmpl, nil)

	// Get stores
	referencedStores, allDefinitions := transformer.GetTrackedStores(transformed)
	storeDefinitions := transformer.GetReferencedStoreDefinitions(allDefinitions, referencedStores)

	// Render
	markup, _, _ := renderer.RenderWithStores(tmpl, transformed, storeDefinitions, "CartBadge.html", "", nil)

	// CRITICAL CHECKS - The bug is that toFixed(2) becomes toFixed
	criticalChecks := []struct {
		name     string
		mustHave string
		desc     string
	}{
		{
			name:     "has toFixed with argument",
			mustHave: `.toFixed(2)`,
			desc:     "toFixed must include the (2) argument",
		},
		{
			name:     "has dollar prefix",
			mustHave: `'$' +`,
			desc:     "must have the '$' + prefix",
		},
		{
			name:     "complete x-text attribute",
			mustHave: `x-text="'$' + $store.cart.total.toFixed(2)"`,
			desc:     "the complete attribute must be exactly as written",
		},
	}

	for _, check := range criticalChecks {
		if !strings.Contains(markup, check.mustHave) {
			t.Errorf("BUG CONFIRMED - %s: %s\nMissing: %q\nRendered HTML:\n%s",
				check.name, check.desc, check.mustHave, markup)
		}
	}

	// Also check what we should NOT have
	badPatterns := []struct {
		name        string
		shouldNotHave string
		desc          string
	}{
		{
			name:          "toFixed without argument",
			shouldNotHave: `x-text="$store.cart.total.toFixed"`,
			desc:          "this would indicate the (2) was stripped",
		},
		{
			name:          "missing dollar prefix",
			shouldNotHave: `x-text="$store.cart.total.toFixed(2)"`,
			desc:          "this would indicate the '$' + was stripped",
		},
	}

	for _, bad := range badPatterns {
		if strings.Contains(markup, bad.shouldNotHave) {
			t.Errorf("BUG DETECTED - %s: %s\nFound unwanted pattern: %q\nRendered HTML:\n%s",
				bad.name, bad.desc, bad.shouldNotHave, markup)
		}
	}
}

// TestCartBadge_StoreExpressionsInsideConditionals tests that store expressions
// inside conditional blocks are transformed correctly
func TestCartBadge_StoreExpressionsInsideConditionals(t *testing.T) {
	templateStr := `---
store cart = {
	items: [],
	total: 0
}
---
<div>
  {if $cart.items.length > 0}
    <span x-text="'Total: $' + $store.cart.total.toFixed(2)"></span>
  {/if}
</div>`

	tmpl, err := parser.ParseTemplate(templateStr)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	transformed := transformer.TransformAST(tmpl, nil)
	referencedStores, allDefinitions := transformer.GetTrackedStores(transformed)
	storeDefinitions := transformer.GetReferencedStoreDefinitions(allDefinitions, referencedStores)

	markup, _, _ := renderer.RenderWithStores(tmpl, transformed, storeDefinitions, "test.html", "", nil)

	// The x-text attribute should be completely preserved
	if !strings.Contains(markup, `x-text="'Total: $' + $store.cart.total.toFixed(2)"`) {
		t.Errorf("x-text attribute was modified inside conditional block\nGot: %s", markup)
	}
}
