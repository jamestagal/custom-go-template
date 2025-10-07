package renderer

import (
	"strings"
	"testing"
)

// TestRenderStoreInitializations_Empty tests rendering with no stores
func TestRenderStoreInitializations_Empty(t *testing.T) {
	stores := map[string]string{}
	result := renderStoreInitializations(stores)

	if result != "" {
		t.Errorf("Expected empty string for no stores, got: %s", result)
	}
}

// TestRenderStoreInitializations_SingleStore tests rendering a single store
func TestRenderStoreInitializations_SingleStore(t *testing.T) {
	stores := map[string]string{
		"auth": "{ isLoggedIn: false, user: null }",
	}
	result := renderStoreInitializations(stores)

	// Expected output structure
	expected := `<script>
document.addEventListener('alpine:init', () => {
    Alpine.store('auth', { isLoggedIn: false, user: null });
});
</script>`

	if result != expected {
		t.Errorf("Single store rendering mismatch.\nExpected:\n%s\n\nGot:\n%s", expected, result)
	}
}

// TestRenderStoreInitializations_MultipleStores tests rendering multiple stores
func TestRenderStoreInitializations_MultipleStores(t *testing.T) {
	stores := map[string]string{
		"auth": "{ isLoggedIn: false, user: null }",
		"cart": "{ items: [], total: 0 }",
	}
	result := renderStoreInitializations(stores)

	// Must contain both stores
	if !strings.Contains(result, "Alpine.store('auth'") {
		t.Error("Result should contain auth store")
	}
	if !strings.Contains(result, "Alpine.store('cart'") {
		t.Error("Result should contain cart store")
	}

	// Must have proper structure
	if !strings.Contains(result, "<script>") {
		t.Error("Result should start with <script> tag")
	}
	if !strings.Contains(result, "</script>") {
		t.Error("Result should end with </script> tag")
	}
	if !strings.Contains(result, "document.addEventListener('alpine:init'") {
		t.Error("Result should contain alpine:init event listener")
	}
}

// TestRenderStoreInitializations_StoreWithMethods tests rendering store with functions
func TestRenderStoreInitializations_StoreWithMethods(t *testing.T) {
	stores := map[string]string{
		"counter": `{
			count: 0,
			increment() {
				this.count++;
			},
			decrement() {
				this.count--;
			}
		}`,
	}
	result := renderStoreInitializations(stores)

	// Check that methods are preserved
	if !strings.Contains(result, "increment()") {
		t.Error("Result should contain increment method")
	}
	if !strings.Contains(result, "decrement()") {
		t.Error("Result should contain decrement method")
	}
	if !strings.Contains(result, "this.count++") {
		t.Error("Result should contain method body")
	}
}

// TestRenderStoreInitializations_NestedObjects tests rendering store with nested objects
func TestRenderStoreInitializations_NestedObjects(t *testing.T) {
	stores := map[string]string{
		"user": `{
			profile: {
				name: 'John',
				age: 30
			},
			settings: {
				theme: 'dark'
			}
		}`,
	}
	result := renderStoreInitializations(stores)

	// Check nested structure is preserved
	if !strings.Contains(result, "profile:") {
		t.Error("Result should contain profile property")
	}
	if !strings.Contains(result, "settings:") {
		t.Error("Result should contain settings property")
	}
	if !strings.Contains(result, "theme:") {
		t.Error("Result should contain theme property")
	}
}

// TestRenderStoreInitializations_SpecialCharacters tests rendering store with special characters
func TestRenderStoreInitializations_SpecialCharacters(t *testing.T) {
	stores := map[string]string{
		"messages": `{ text: "Hello 'World' & \"Friends\"" }`,
	}
	result := renderStoreInitializations(stores)

	// The store definition should be embedded as-is (already valid JS)
	// No need to escape since it's inside a script tag
	if !strings.Contains(result, "Alpine.store('messages'") {
		t.Error("Result should contain messages store")
	}
}

// TestRenderStoreInitializations_FormatStructure tests output formatting
func TestRenderStoreInitializations_FormatStructure(t *testing.T) {
	stores := map[string]string{
		"test": "{ value: 1 }",
	}
	result := renderStoreInitializations(stores)

	// Check structure elements
	tests := []struct {
		name     string
		contains string
	}{
		{"script open tag", "<script>"},
		{"script close tag", "</script>"},
		{"event listener", "document.addEventListener('alpine:init', () => {"},
		{"listener close", "});"},
		{"store call", "Alpine.store('test', { value: 1 });"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(result, tt.contains) {
				t.Errorf("Result should contain %q\nGot:\n%s", tt.contains, result)
			}
		})
	}
}

// TestRenderStoreInitializations_Indentation tests proper indentation
func TestRenderStoreInitializations_Indentation(t *testing.T) {
	stores := map[string]string{
		"test": "{ value: 1 }",
	}
	result := renderStoreInitializations(stores)

	// Check for proper indentation (4 spaces for nested content)
	lines := strings.Split(result, "\n")

	// Verify structure (line by line)
	if len(lines) < 5 {
		t.Errorf("Expected at least 5 lines, got %d", len(lines))
	}

	// Line 0: <script>
	if !strings.HasPrefix(lines[0], "<script>") {
		t.Errorf("Line 0 should start with <script>, got: %s", lines[0])
	}

	// Line 1: document.addEventListener...
	if !strings.HasPrefix(lines[1], "document.addEventListener") {
		t.Errorf("Line 1 should start with document.addEventListener, got: %s", lines[1])
	}

	// Check for proper indentation on Alpine.store line
	foundStoreLine := false
	for _, line := range lines {
		if strings.Contains(line, "Alpine.store") {
			if !strings.HasPrefix(line, "    ") {
				t.Errorf("Alpine.store line should be indented with 4 spaces, got: %q", line)
			}
			foundStoreLine = true
			break
		}
	}
	if !foundStoreLine {
		t.Error("Should contain Alpine.store line")
	}
}

// TestRenderStoreInitializations_OrderDeterministic tests that output is deterministic
// Note: Maps in Go have non-deterministic iteration order, but we can test
// that all stores are present regardless of order
func TestRenderStoreInitializations_AllStoresPresent(t *testing.T) {
	stores := map[string]string{
		"alpha":   "{ a: 1 }",
		"beta":    "{ b: 2 }",
		"gamma":   "{ g: 3 }",
		"delta":   "{ d: 4 }",
		"epsilon": "{ e: 5 }",
	}
	result := renderStoreInitializations(stores)

	// All stores should be present
	for storeName := range stores {
		expectedCall := "Alpine.store('" + storeName + "'"
		if !strings.Contains(result, expectedCall) {
			t.Errorf("Result should contain store call for %s", storeName)
		}
	}

	// Should have exactly 5 Alpine.store calls
	count := strings.Count(result, "Alpine.store(")
	if count != 5 {
		t.Errorf("Expected 5 Alpine.store calls, got %d", count)
	}
}
