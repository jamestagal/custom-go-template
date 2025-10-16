package integration

import (
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/parser"
	"github.com/jimafisk/custom_go_template/renderer"
	"github.com/jimafisk/custom_go_template/transformer"
)

// TestPropsAndStoresSeparation verifies that props and stores don't interfere
//
// Pattern: Integration Test [Load: 18]
// Cognitive Load: 18 (setup: 4, parse: 3, transform: 3, render: 4, assertions: 4)
func TestPropsAndStoresSeparation(t *testing.T) {
	templateStr := `---
prop userName = "DefaultUser"
prop userAge = 25

store session = {
	isActive: true,
	timeout: 3600
}
---
<div>
	<p>Name: {userName}</p>
	<p>Age: {userAge}</p>
	<p>Session Active: {$session.isActive}</p>
	<p>Timeout: {$session.timeout}</p>
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
	markup, script, _ := renderer.RenderWithStores(tmpl, transformed, storeDefinitions, "test.html", "")

	// Verify props use x-text without $store prefix
	if !strings.Contains(markup, "x-text=\"userName\"") {
		t.Errorf("Props should use variable name directly, not $store")
	}
	if !strings.Contains(markup, "x-text=\"userAge\"") {
		t.Errorf("Props should use variable name directly")
	}

	// Verify stores use $store prefix
	if !strings.Contains(markup, "x-text=\"$store.session.isActive\"") {
		t.Errorf("Stores should use $store.session prefix")
	}
	if !strings.Contains(markup, "x-text=\"$store.session.timeout\"") {
		t.Errorf("Stores should use $store.session prefix")
	}

	// Verify store initialization
	if !strings.Contains(script, "Alpine.store('session'") {
		t.Errorf("Should initialize session store")
	}

	// Props should be in x-data, not in Alpine.store
	if strings.Contains(script, "Alpine.store('userName'") {
		t.Errorf("Props should NOT be in Alpine.store")
	}
}

// TestPropNamedSameAsStore verifies behavior when prop name matches store name
//
// Pattern: Integration Test [Load: 16]
// Cognitive Load: 16 (setup: 3, parse: 3, transform: 3, render: 3, assertions: 4)
func TestPropNamedSameAsStore(t *testing.T) {
	templateStr := `---
prop config = { version: "1.0" }

store config = {
	apiKey: "abc123",
	endpoint: "https://api.example.com"
}
---
<div>
	<p>Local Config: {config.version}</p>
	<p>Store Config: {$config.apiKey}</p>
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
	markup, script, _ := renderer.RenderWithStores(tmpl, transformed, storeDefinitions, "test.html", "")

	// Verify prop uses local scope
	if !strings.Contains(markup, "x-text=\"config.version\"") {
		t.Errorf("Prop should use local scope: config.version")
	}

	// Verify store uses $store prefix
	if !strings.Contains(markup, "x-text=\"$store.config.apiKey\"") {
		t.Errorf("Store should use $store.config prefix")
	}

	// Verify store initialization
	if !strings.Contains(script, "Alpine.store('config'") {
		t.Errorf("Should initialize config store")
	}
	if !strings.Contains(script, "apiKey") {
		t.Errorf("Store should contain apiKey")
	}
}

// TestScopingDocumentation documents the scoping behavior
//
// Pattern: Documentation Test [Load: 1]
// Cognitive Load: 1 (documentation only)
func TestScopingDocumentation(t *testing.T) {
	/*
		SCOPING BEHAVIOR:

		1. Props (Component-Local Scope):
		   - Defined with: prop name = value
		   - Accessed with: {name} or {name.property}
		   - Rendered as: x-text="name" or x-text="name.property"
		   - Stored in: Component's x-data object
		   - Scope: Local to component instance

		2. Stores (Global Scope):
		   - Defined with: store name = { ... }
		   - Accessed with: {$name.property}
		   - Rendered as: x-text="$store.name.property"
		   - Stored in: Alpine.store('name', { ... })
		   - Scope: Global across all components

		3. Name Collision:
		   - If prop and store have same name
		   - Prop: Use {name} → x-text="name"
		   - Store: Use {$name} → x-text="$store.name"
		   - The $ prefix disambiguates

		4. Best Practices:
		   - Use props for component-specific data
		   - Use stores for shared/global state
		   - Avoid naming collisions when possible
		   - Use $ prefix consistently for stores
	*/

	t.Log("Scoping documentation test passed")
}

// TestMixedPropsStoresConditionals verifies props and stores in conditionals
//
// Pattern: Integration Test [Load: 19]
// Cognitive Load: 19 (setup: 4, parse: 3, transform: 4, render: 4, assertions: 4)
func TestMixedPropsStoresConditionals(t *testing.T) {
	templateStr := `---
prop isAdmin = false

store auth = {
	isLoggedIn: true,
	userRole: "guest"
}
---
<div>
	{if isAdmin}
		<p>Admin Panel</p>
	{/if}
	
	{if $auth.isLoggedIn}
		<p>Welcome, {$auth.userRole}</p>
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
	markup, _, _ := renderer.RenderWithStores(tmpl, transformed, storeDefinitions, "test.html", "")

	// Verify prop conditional uses local scope
	if !strings.Contains(markup, "x-if=\"isAdmin\"") {
		t.Errorf("Prop conditional should use local scope")
	}

	// Verify store conditional uses $store prefix
	if !strings.Contains(markup, "x-if=\"$store.auth.isLoggedIn\"") {
		t.Errorf("Store conditional should use $store prefix")
	}

	// Verify store property access in text
	if !strings.Contains(markup, "x-text=\"$store.auth.userRole\"") {
		t.Errorf("Store property should use $store prefix")
	}
}
