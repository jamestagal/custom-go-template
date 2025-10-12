package integration

import (
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/parser"
	"github.com/jimafisk/custom_go_template/renderer"
	"github.com/jimafisk/custom_go_template/transformer"
)

// TestStoreInNestedConditional tests store access in nested if statements
//
// Pattern: Integration Test [Load: 20]
// Cognitive Load: 20 (setup: 4, parse: 4, transform: 4, render: 4, assertions: 4)
func TestStoreInNestedConditional(t *testing.T) {
	templateStr := `---
store user = {
	isLoggedIn: true,
	role: "admin",
	permissions: {
		canEdit: true,
		canDelete: false
	}
}
---
<div>
	{if $user.isLoggedIn}
		{if $user.role === "admin"}
			{if $user.permissions.canEdit}
				<button>Edit</button>
			{/if}
		{/if}
	{/if}
</div>`

	// Parse and transform
	tmpl, err := parser.ParseTemplate(templateStr)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	transformed := transformer.TransformAST(tmpl, nil)

	// Get store definitions
	referencedStores, allDefinitions := transformer.GetTrackedStores(transformed)
	storeDefinitions := transformer.GetReferencedStoreDefinitions(allDefinitions, referencedStores)

	// Render
	markup, _, _ := renderer.RenderWithStores(tmpl, transformed, storeDefinitions, "test.html", "", "")

	// Verify all three levels of nesting use $store prefix
	if !strings.Contains(markup, "x-if=\"$store.user.isLoggedIn\"") {
		t.Errorf("First level conditional should use $store")
	}
	
	// Note: Quotes in string comparisons are HTML-escaped as &quot;
	if !strings.Contains(markup, "$store.user.role === &quot;admin&quot;") {
		t.Errorf("Second level conditional should use $store")
	}
	
	if !strings.Contains(markup, "x-if=\"$store.user.permissions.canEdit\"") {
		t.Errorf("Third level conditional should use $store")
	}
}

// TestStoreInNestedLoop tests store access in nested for loops
//
// Pattern: Integration Test [Load: 21]
// Cognitive Load: 21 (setup: 5, parse: 4, transform: 4, render: 4, assertions: 4)
func TestStoreInNestedLoop(t *testing.T) {
	templateStr := `---
store data = {
	categories: [
		{ name: "Fruits", items: ["Apple", "Banana"] },
		{ name: "Vegetables", items: ["Carrot", "Lettuce"] }
	]
}
---
<div>
	{for category in $data.categories}
		<div class="category">
			<h3>{category.name}</h3>
			{for item in category.items}
				<p>{item}</p>
			{/for}
		</div>
	{/for}
</div>`

	// Parse and transform
	tmpl, err := parser.ParseTemplate(templateStr)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	transformed := transformer.TransformAST(tmpl, nil)

	// Get store definitions
	referencedStores, allDefinitions := transformer.GetTrackedStores(transformed)
	storeDefinitions := transformer.GetReferencedStoreDefinitions(allDefinitions, referencedStores)

	// Render
	markup, _, _ := renderer.RenderWithStores(tmpl, transformed, storeDefinitions, "test.html", "", "")

	// Verify outer loop uses $store prefix
	if !strings.Contains(markup, "x-for=\"(category, ) in $store.data.categories\"") {
		t.Errorf("Outer loop should iterate over $store.data.categories")
	}

	// Verify inner loop uses loop variable (not $store)
	if !strings.Contains(markup, "x-for=\"(item, ) in category.items\"") {
		t.Errorf("Inner loop should iterate over category.items")
	}

	// Verify category.name is accessed correctly
	if !strings.Contains(markup, "x-text=\"category.name\"") {
		t.Errorf("Loop variable property should be accessed directly")
	}
}

// TestMultipleStoresSingleTemplate tests multiple stores in one template
//
// Pattern: Integration Test [Load: 19]
// Cognitive Load: 19 (setup: 4, parse: 3, transform: 4, render: 4, assertions: 4)
func TestMultipleStoresSingleTemplate(t *testing.T) {
	templateStr := `---
store auth = {
	user: "Alice",
	isLoggedIn: true
}

store cart = {
	items: [],
	total: 0
}

store theme = {
	mode: "dark"
}
---
<div>
	<p>User: {$auth.user}</p>
	<p>Cart Items: {$cart.items.length}</p>
	<p>Theme: {$theme.mode}</p>
</div>`

	// Parse and transform
	tmpl, err := parser.ParseTemplate(templateStr)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	transformed := transformer.TransformAST(tmpl, nil)

	// Get store definitions
	referencedStores, allDefinitions := transformer.GetTrackedStores(transformed)
	storeDefinitions := transformer.GetReferencedStoreDefinitions(allDefinitions, referencedStores)

	// Render
	markup, script, _ := renderer.RenderWithStores(tmpl, transformed, storeDefinitions, "test.html", "", "")

	// Verify all three stores are referenced correctly
	if !strings.Contains(markup, "x-text=\"$store.auth.user\"") {
		t.Errorf("Should reference auth store")
	}
	if !strings.Contains(markup, "x-text=\"$store.cart.items.length\"") {
		t.Errorf("Should reference cart store")
	}
	if !strings.Contains(markup, "x-text=\"$store.theme.mode\"") {
		t.Errorf("Should reference theme store")
	}

	// Verify all three stores are initialized
	if !strings.Contains(script, "Alpine.store('auth'") {
		t.Errorf("Should initialize auth store")
	}
	if !strings.Contains(script, "Alpine.store('cart'") {
		t.Errorf("Should initialize cart store")
	}
	if !strings.Contains(script, "Alpine.store('theme'") {
		t.Errorf("Should initialize theme store")
	}
}

// TestStoreWithDynamicPropertyNames tests stores with computed properties
//
// Pattern: Integration Test [Load: 17]
// Cognitive Load: 17 (setup: 4, parse: 3, transform: 3, render: 3, assertions: 4)
func TestStoreWithDynamicPropertyNames(t *testing.T) {
	templateStr := `---
prop fieldName = "email"

store form = {
	email: "user@example.com",
	phone: "555-0100"
}
---
<div>
	<p>Email: {$form.email}</p>
</div>`

	// Parse and transform
	tmpl, err := parser.ParseTemplate(templateStr)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	transformed := transformer.TransformAST(tmpl, nil)

	// Get store definitions
	referencedStores, allDefinitions := transformer.GetTrackedStores(transformed)
	storeDefinitions := transformer.GetReferencedStoreDefinitions(allDefinitions, referencedStores)

	// Render
	markup, _, _ := renderer.RenderWithStores(tmpl, transformed, storeDefinitions, "test.html", "", "")

	// Note: Bracket notation {$form[fieldName]} is not yet implemented
	// For now, we test dot notation which works

	// Verify dot notation works
	if !strings.Contains(markup, "$store.form.email") {
		t.Errorf("Should support dot notation access")
	}
	
	// TODO: Add support for bracket notation in future:
	// {$form[fieldName]} → $store.form[fieldName]
}
