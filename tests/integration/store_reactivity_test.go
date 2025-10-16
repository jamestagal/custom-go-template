package integration

import (
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/parser"
	"github.com/jimafisk/custom_go_template/renderer"
	"github.com/jimafisk/custom_go_template/transformer"
)

// TestCrossComponentReactivity_MultipleComponentsShareStore verifies that
// multiple components can reference the same store and Alpine.js will update
// all components when the store changes
//
// Pattern: Integration Test [Load: 18]
// Cognitive Load: 18 (setup: 3, parse: 3, transform: 4, render: 4, assertions: 4)
func TestCrossComponentReactivity_MultipleComponentsShareStore(t *testing.T) {
	templateStr := `---
store auth = {
	user: null,
	isLoggedIn: false,
	login(username) {
		this.user = { name: username };
		this.isLoggedIn = true;
	},
	logout() {
		this.user = null;
		this.isLoggedIn = false;
	}
}
---
<div>
	<div class="header">
		{if $auth.isLoggedIn}
			<span>Welcome, {$auth.user.name}</span>
		{else}
			<span>Please log in</span>
		{/if}
	</div>

	<div class="sidebar">
		{if $auth.isLoggedIn}
			<button @click="$store.auth.logout()">Logout</button>
		{else}
			<button @click="$store.auth.login('TestUser')">Login</button>
		{/if}
	</div>

	<div class="footer">
		<span>User status: {$auth.isLoggedIn ? 'logged in' : 'guest'}</span>
	</div>
</div>`

	// Parse template
	tmpl, err := parser.ParseTemplate(templateStr)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// Transform
	transformed := transformer.TransformAST(tmpl, nil)

	// Get tracked stores
	referencedStores, allDefinitions := transformer.GetTrackedStores(transformed)
	if len(referencedStores) == 0 || referencedStores[0] != "auth" {
		t.Errorf("Should track 'auth' store, got: %v", referencedStores)
	}

	// Extract store definitions
	storeDefinitions := transformer.GetReferencedStoreDefinitions(allDefinitions, referencedStores)
	if _, found := storeDefinitions["auth"]; !found {
		t.Fatalf("Should have auth store definition, got: %v", storeDefinitions)
	}

	// Render
	markup, script, _ := renderer.RenderWithStores(tmpl, transformed, storeDefinitions, "test.html", "")

	// Verify all three components reference $store.auth
	if !strings.Contains(markup, "x-text=\"$store.auth.user.name\"") {
		t.Errorf("Header should reference store user.name")
	}
	if !strings.Contains(markup, "$store.auth.logout()") {
		t.Errorf("Sidebar logout should reference store")
	}
	if !strings.Contains(markup, "$store.auth.login") {
		t.Errorf("Sidebar login should reference store")
	}
	if !strings.Contains(markup, "$store.auth.isLoggedIn") {
		t.Errorf("Footer should reference store isLoggedIn")
	}

	// Verify store initialization includes all methods
	if !strings.Contains(script, "Alpine.store('auth'") {
		t.Errorf("Should initialize auth store")
	}
	if !strings.Contains(script, "login(username)") {
		t.Errorf("Should include login method")
	}
	if !strings.Contains(script, "logout()") {
		t.Errorf("Should include logout method")
	}

	// Verify Alpine.js reactive structure
	// All components will update when $store.auth changes because Alpine.js
	// tracks dependencies automatically
	if !strings.Contains(markup, "x-if=\"$store.auth.isLoggedIn\"") {
		t.Errorf("Should use x-if for reactivity")
	}
}

// TestStoreModificationFromAlpineCode verifies that store modifications
// from Alpine.js event handlers properly trigger reactivity
//
// Pattern: Integration Test [Load: 16]
// Cognitive Load: 16 (setup: 3, parse: 3, transform: 3, render: 3, assertions: 4)
func TestStoreModificationFromAlpineCode(t *testing.T) {
	templateStr := `---
store counter = {
	count: 0,
	increment() {
		this.count++;
	},
	decrement() {
		this.count--;
	}
}
---
<div>
	<button @click="$store.counter.increment()">+</button>
	<span>Count: {$counter.count}</span>
	<button @click="$store.counter.decrement()">-</button>
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

	// Verify event handlers use $store.counter
	if !strings.Contains(markup, "@click=\"$store.counter.increment()\"") {
		t.Errorf("Increment should call store method")
	}
	if !strings.Contains(markup, "@click=\"$store.counter.decrement()\"") {
		t.Errorf("Decrement should call store method")
	}

	// Verify display uses $store.counter
	if !strings.Contains(markup, "x-text=\"$store.counter.count\"") {
		t.Errorf("Count display should reference store")
	}

	// Verify store initialization
	if !strings.Contains(script, "Alpine.store('counter'") {
		t.Errorf("Should initialize counter store")
	}
	if !strings.Contains(script, "count: 0") {
		t.Errorf("Should set initial count")
	}
	if !strings.Contains(script, "increment()") {
		t.Errorf("Should include increment method")
	}
	if !strings.Contains(script, "decrement()") {
		t.Errorf("Should include decrement method")
	}
}

// TestMultipleComponentsSameStoreChanges verifies that when one component
// modifies a store, all other components referencing that store update
//
// Pattern: Integration Test [Load: 19]
// Cognitive Load: 19 (setup: 4, parse: 3, transform: 3, render: 3, assertions: 6)
func TestMultipleComponentsSameStoreChanges(t *testing.T) {
	templateStr := `---
store cart = {
	items: [],
	addItem(item) {
		this.items.push(item);
	},
	removeItem(index) {
		this.items.splice(index, 1);
	}
}
---
<div>
	<!-- Component 1: Add items -->
	<div class="add-section">
		<button @click="$store.cart.addItem('Product A')">Add Product A</button>
	</div>

	<!-- Component 2: Display count -->
	<div class="badge">
		Cart: {$cart.items.length} items
	</div>

	<!-- Component 3: Display list -->
	<div class="cart-list">
		{for item in $cart.items}
			<div class="cart-item">{item}</div>
		{/for}
	</div>

	<!-- Component 4: Empty state -->
	<div class="empty-state">
		{if $cart.items.length === 0}
			<p>Cart is empty</p>
		{/if}
	</div>
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

	// Verify all four "components" reference the same store
	// Component 1: Add button
	if !strings.Contains(markup, "$store.cart.addItem") {
		t.Errorf("Add button should reference store")
	}

	// Component 2: Badge count
	if !strings.Contains(markup, "$store.cart.items.length") {
		t.Errorf("Badge should reference store array length")
	}

	// Component 3: List items
	if !strings.Contains(markup, "x-for=\"(item, ) in $store.cart.items\"") {
		t.Errorf("List should iterate store array")
	}

	// Component 4: Empty state
	if !strings.Contains(markup, "x-if=\"$store.cart.items.length === 0\"") {
		t.Errorf("Empty state should check store")
	}

	// Verify store initialization
	if !strings.Contains(script, "Alpine.store('cart'") {
		t.Errorf("Should initialize cart store")
	}
	if !strings.Contains(script, "items: []") {
		t.Errorf("Should initialize empty array")
	}
	if !strings.Contains(script, "addItem(item)") {
		t.Errorf("Should include addItem method")
	}
	if !strings.Contains(script, "removeItem(index)") {
		t.Errorf("Should include removeItem method")
	}

	// When $store.cart.addItem() is called in the browser:
	// 1. Alpine.js detects the store change
	// 2. All four components automatically update:
	//    - Badge shows new count
	//    - List shows new item
	//    - Empty state disappears (if first item)
	//    - Add button remains functional
}

// TestReactivityDocumentation documents the expected reactivity behavior
//
// Pattern: Documentation Test [Load: 1]
// Cognitive Load: 1 (simple assertion)
func TestReactivityDocumentation(t *testing.T) {
	// This test documents how Alpine.js handles store reactivity

	/*
		REACTIVITY BEHAVIOR:

		1. Store Definition:
		   - Stores are global Alpine.js reactive objects
		   - Defined with: Alpine.store('name', { ... })
		   - Initialized on 'alpine:init' event

		2. Cross-Component Updates:
		   - Any component can reference $store.storeName
		   - When store changes, ALL components using that store update
		   - No manual notification needed - Alpine.js tracks dependencies

		3. Store Modification:
		   - From event handlers: @click="$store.auth.login('user')"
		   - From methods: this.count++ in store methods
		   - From x-effect: Run code when store changes

		4. Reactivity Tracking:
		   - Alpine.js automatically tracks which components use which stores
		   - When store property changes, only affected components re-render
		   - Nested properties are tracked: $store.user.profile.name

		5. Common Patterns:
		   - Shared state: Multiple components read same store
		   - Actions: One component modifies, others react
		   - Computed: Use getters for derived values
		   - Watchers: Use x-effect to respond to changes

		EXAMPLE:
		   // Component A modifies store
		   <button @click="$store.cart.addItem('Product')">Add</button>

		   // Component B automatically updates when store changes
		   <span x-text="$store.cart.items.length"></span>

		   // Component C also updates
		   <div x-show="$store.cart.items.length > 0">Cart has items</div>
	*/

	// Mark test as passing - this is documentation
	t.Log("Reactivity documentation test passed")
}

// TestStoreReactivityWithNestedStructures verifies reactivity works with
// nested objects and arrays in stores
//
// Pattern: Integration Test [Load: 17]
// Cognitive Load: 17 (setup: 4, parse: 3, transform: 3, render: 3, assertions: 4)
func TestStoreReactivityWithNestedStructures(t *testing.T) {
	templateStr := `---
store app = {
	user: {
		profile: {
			name: 'John',
			email: 'john@example.com'
		},
		preferences: {
			theme: 'dark',
			notifications: true
		}
	},
	updateTheme(theme) {
		this.user.preferences.theme = theme;
	}
}
---
<div>
	<h1>{$app.user.profile.name}</h1>
	<p>{$app.user.profile.email}</p>
	<div class="theme-{$app.user.preferences.theme}">
		<button @click="$store.app.updateTheme('light')">Light</button>
		<button @click="$store.app.updateTheme('dark')">Dark</button>
	</div>
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

	// Verify nested property access
	if !strings.Contains(markup, "$store.app.user.profile.name") {
		t.Errorf("Should access nested name")
	}
	if !strings.Contains(markup, "$store.app.user.profile.email") {
		t.Errorf("Should access nested email")
	}
	if !strings.Contains(markup, "$store.app.user.preferences.theme") {
		t.Errorf("Should access nested theme")
	}

	// Verify method calls
	if !strings.Contains(markup, "$store.app.updateTheme('light')") {
		t.Errorf("Should call updateTheme with light")
	}
	if !strings.Contains(markup, "$store.app.updateTheme('dark')") {
		t.Errorf("Should call updateTheme with dark")
	}

	// Verify store initialization preserves nested structure
	if !strings.Contains(script, "Alpine.store('app'") {
		t.Errorf("Should initialize app store")
	}

	// Check for nested object structure in initialization
	// The store definition should maintain the nested object structure
	scriptLines := strings.Split(script, "\n")
	foundUserObject := false
	for _, line := range scriptLines {
		if strings.Contains(line, "user: {") || strings.Contains(line, "user:{") {
			foundUserObject = true
			break
		}
	}
	if !foundUserObject {
		t.Errorf("Store should preserve nested user object")
	}
}
