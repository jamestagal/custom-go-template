package transformer

import (
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
)

// TestTrackAlpineStoreReferences tests that $store.storeName patterns are tracked
// This was a critical bug: attributes with @click="$store.theme.setLight()" were not being tracked
func TestTrackAlpineStoreReferences(t *testing.T) {
	tests := []struct {
		name          string
		value         string
		expectedStore string
	}{
		{
			name:          "single store method call",
			value:         "$store.theme.setLight()",
			expectedStore: "theme",
		},
		{
			name:          "store property access",
			value:         "$store.auth.isLoggedIn",
			expectedStore: "auth",
		},
		{
			name:          "multiple stores in expression",
			value:         "$store.auth.isLoggedIn && $store.cart.itemCount > 0",
			expectedStore: "auth", // First one
		},
		{
			name:          "store in complex expression",
			value:         "count + $store.cart.total",
			expectedStore: "cart",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset tracker
			InitStoreTracking(map[string]string{})

			// Track references
			trackAlpineStoreReferences(tt.value)

			// Verify tracking
			if !storeTracker.referencedStores[tt.expectedStore] {
				t.Errorf("Expected store %q to be tracked, but it wasn't", tt.expectedStore)
			}
		})
	}
}

// TestTransformAttributesWithAlpineStores tests that Alpine directives with $store are tracked
func TestTransformAttributesWithAlpineStores(t *testing.T) {
	// Reset and init tracking
	InitStoreTracking(map[string]string{
		"theme": "{ mode: 'light' }",
		"auth":  "{ isLoggedIn: false }",
	})

	attributes := []ast.Attribute{
		{
			Name:     "@click",
			Value:    "$store.theme.setLight()",
			IsAlpine: true,
		},
		{
			Name:     "@click",
			Value:    "$store.auth.login()",
			IsAlpine: true,
		},
		{
			Name:  "class",
			Value: "button",
		},
	}

	dataScope := make(map[string]any)
	result := transformAttributeExpressions(attributes, dataScope)

	// Verify attributes unchanged (Alpine directives pass through)
	if len(result) != 3 {
		t.Errorf("Expected 3 attributes, got %d", len(result))
	}

	// Verify stores were tracked
	if !storeTracker.referencedStores["theme"] {
		t.Error("Expected theme store to be tracked")
	}
	if !storeTracker.referencedStores["auth"] {
		t.Error("Expected auth store to be tracked")
	}
}

// TestAlpineStorePattern tests the regex pattern for $store.storeName
func TestAlpineStorePattern(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedMatch bool
		expectedStore string
	}{
		{
			name:          "basic store access",
			input:         "$store.theme.mode",
			expectedMatch: true,
			expectedStore: "theme",
		},
		{
			name:          "store method call",
			input:         "$store.cart.addItem('widget')",
			expectedMatch: true,
			expectedStore: "cart",
		},
		{
			name:          "store in condition",
			input:         "$store.auth.isLoggedIn === true",
			expectedMatch: true,
			expectedStore: "auth",
		},
		{
			name:          "no store reference",
			input:         "someVariable.property",
			expectedMatch: false,
			expectedStore: "",
		},
		{
			name:          "store without property",
			input:         "$store",
			expectedMatch: false,
			expectedStore: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := alpineStorePattern.FindStringSubmatch(tt.input)

			if tt.expectedMatch {
				if len(matches) < 2 {
					t.Errorf("Expected to match store pattern, but didn't. Input: %q", tt.input)
					return
				}
				if matches[1] != tt.expectedStore {
					t.Errorf("Expected store name %q, got %q", tt.expectedStore, matches[1])
				}
			} else {
				if len(matches) > 0 {
					t.Errorf("Expected no match, but matched: %v", matches)
				}
			}
		})
	}
}

// TestBugFix_ThemeStoreNotTracked is a regression test for the critical bug where
// theme store was defined but not initialized because Alpine @click references weren't tracked
func TestBugFix_ThemeStoreNotTracked(t *testing.T) {
	// Simulate the exact scenario from store-components-demo.html
	InitStoreTracking(map[string]string{
		"auth":  "{ isLoggedIn: false }",
		"cart":  "{ items: [] }",
		"theme": "{ mode: 'light' }",
	})

	// These are the types of attributes that were missing tracking
	attributes := []ast.Attribute{
		{
			Name:     "@click",
			Value:    "$store.theme.setLight()",
			IsAlpine: true,
		},
		{
			Name:     "@click",
			Value:    "$store.theme.setDark()",
			IsAlpine: true,
		},
		{
			Name:     "@click",
			Value:    "$store.theme.toggle()",
			IsAlpine: true,
		},
	}

	dataScope := make(map[string]any)
	transformAttributeExpressions(attributes, dataScope)

	// Get tracked stores
	referencedStores, allDefs := GetTrackedStores(&ast.Template{})

	// Verify theme is tracked
	themeTracked := false
	for _, store := range referencedStores {
		if store == "theme" {
			themeTracked = true
			break
		}
	}

	if !themeTracked {
		t.Errorf("Theme store not tracked! Referenced stores: %v", referencedStores)
	}

	// Verify theme definition exists
	referencedDefs := GetReferencedStoreDefinitions(allDefs, referencedStores)
	if _, exists := referencedDefs["theme"]; !exists {
		t.Error("Theme store definition not in referenced definitions")
	}
}
