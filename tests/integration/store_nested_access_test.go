package integration

import (
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/parser"
	"github.com/jimafisk/custom_go_template/renderer"
	"github.com/jimafisk/custom_go_template/transformer"
)

// TestNestedPropertyAccess tests multi-level nested property access
//
// Pattern: Integration Test [Load: 16]
// Cognitive Load: 16 (setup: 4, parse: 3, transform: 3, render: 3, assertions: 3)
func TestNestedPropertyAccess(t *testing.T) {
	templateStr := `---
store user = {
	profile: {
		contact: {
			email: "user@example.com",
			phone: "+1-555-0100"
		},
		address: {
			city: "San Francisco",
			zip: "94105"
		}
	}
}
---
<div>
	<p>{$user.profile.contact.email}</p>
	<p>{$user.profile.contact.phone}</p>
	<p>{$user.profile.address.city}, {$user.profile.address.zip}</p>
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
	markup, script, _ := renderer.RenderWithStores(tmpl, transformed, storeDefinitions, "test.html")

	// Verify deeply nested property access
	if !strings.Contains(markup, "x-text=\"$store.user.profile.contact.email\"") {
		t.Errorf("Should access 4-level nested property: user.profile.contact.email")
	}
	if !strings.Contains(markup, "x-text=\"$store.user.profile.contact.phone\"") {
		t.Errorf("Should access 4-level nested property: user.profile.contact.phone")
	}
	if !strings.Contains(markup, "x-text=\"$store.user.profile.address.city\"") {
		t.Errorf("Should access 4-level nested property: user.profile.address.city")
	}

	// Verify store preserves nested structure
	if !strings.Contains(script, "profile:") {
		t.Errorf("Store should preserve nested structure")
	}
}

// TestArrayIndexAccess documents array index access (not yet implemented)
//
// Pattern: Documentation Test [Load: 1]
// Cognitive Load: 1 (documentation only)
func TestArrayIndexAccess(t *testing.T) {
	/*
		ARRAY INDEX ACCESS:

		Array index access in store expressions is a future feature.

		Desired behavior:
		   {$data.items[0]} → x-text="$store.data.items[0]"
		   {$data.users[0].name} → x-text="$store.data.users[0].name"

		Current workaround:
		   Use x-for loops to iterate arrays
		   {for item in $data.items}
		       <p>{item}</p>
		   {/for}

		TODO: Implement array index access in parser/expressions.go
		- Detect [ ] brackets in store property path
		- Preserve bracket notation through transformation
		- Test with nested arrays and objects
	*/

	t.Log("Array index access documentation test")
}

// TestNullUndefinedPropertyAccess documents null/undefined behavior
//
// Pattern: Documentation Test [Load: 1]
// Cognitive Load: 1 (documentation only)
func TestNullUndefinedPropertyAccess(t *testing.T) {
	/*
		NULL/UNDEFINED PROPERTY ACCESS:

		Alpine.js handles null/undefined gracefully:

		1. Null Properties:
		   - {$user.profile.name} where profile is null
		   - Alpine.js returns undefined, doesn't crash
		   - Template shows empty/nothing

		2. Undefined Properties:
		   - {$user.nonexistent.property}
		   - Alpine.js returns undefined
		   - No JavaScript errors thrown

		3. Optional Chaining (if needed):
		   - Can use: {$user?.profile?.name}
		   - Alpine.js supports ES2020 optional chaining
		   - Explicitly handles null/undefined

		4. Default Values:
		   - Use ternary: {$user.name || 'Guest'}
		   - Or nullish coalescing: {$user.name ?? 'Guest'}
		   - Provides fallback values

		5. Best Practices:
		   - Initialize stores with non-null defaults when possible
		   - Use optional chaining for uncertain paths
		   - Provide default values for user-facing display
		   - Test edge cases with null/undefined data
	*/

	t.Log("Null/undefined documentation test passed")
}

// TestNestedConditionalAccess tests nested properties in conditionals
//
// Pattern: Integration Test [Load: 17]
// Cognitive Load: 17 (setup: 4, parse: 3, transform: 3, render: 3, assertions: 4)
func TestNestedConditionalAccess(t *testing.T) {
	templateStr := `---
store app = {
	settings: {
		features: {
			darkMode: true,
			notifications: false
		}
	}
}
---
<div>
	{if $app.settings.features.darkMode}
		<p>Dark mode enabled</p>
	{/if}
	
	{if $app.settings.features.notifications}
		<p>Notifications on</p>
	{else}
		<p>Notifications off</p>
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
	markup, _, _ := renderer.RenderWithStores(tmpl, transformed, storeDefinitions, "test.html")

	// Verify nested property access in conditionals
	if !strings.Contains(markup, "x-if=\"$store.app.settings.features.darkMode\"") {
		t.Errorf("Conditional should use nested store property")
	}
	if !strings.Contains(markup, "x-if=\"$store.app.settings.features.notifications\"") {
		t.Errorf("Conditional should use nested store property")
	}
}
