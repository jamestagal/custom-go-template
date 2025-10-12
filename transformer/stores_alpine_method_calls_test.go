package transformer

import (
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
)

// TestAlpineAttributesWithMethodCalls tests that Alpine attributes with method calls are preserved unchanged
// BUG: x-text="'$' + $store.cart.total.toFixed(2)" was becoming x-text="$store.cart.total.toFixed"
func TestAlpineAttributesWithMethodCalls(t *testing.T) {
	tests := []struct {
		name           string
		inputAttr      ast.Attribute
		expectedValue  string
		expectedName   string
		expectedStores []string
	}{
		{
			name: "x-text with toFixed(2)",
			inputAttr: ast.Attribute{
				Name:       "x-text",
				Value:      "'$' + $store.cart.total.toFixed(2)",
				IsAlpine:   true,
				AlpineType: "text",
			},
			expectedValue:  "'$' + $store.cart.total.toFixed(2)", // Should be UNCHANGED
			expectedName:   "x-text",
			expectedStores: []string{"cart"},
		},
		{
			name: "x-text with complex expression",
			inputAttr: ast.Attribute{
				Name:       "x-text",
				Value:      "`Total: $${Number($store.cart.total).toFixed(2)}`",
				IsAlpine:   true,
				AlpineType: "text",
			},
			expectedValue:  "`Total: $${Number($store.cart.total).toFixed(2)}`", // Should be UNCHANGED
			expectedName:   "x-text",
			expectedStores: []string{"cart"},
		},
		{
			name: "@click with method call",
			inputAttr: ast.Attribute{
				Name:       "@click",
				Value:      "$store.theme.setDark()",
				IsAlpine:   true,
				AlpineType: "click",
			},
			expectedValue:  "$store.theme.setDark()", // Should be UNCHANGED
			expectedName:   "@click",
			expectedStores: []string{"theme"},
		},
		{
			name: "x-show with comparison",
			inputAttr: ast.Attribute{
				Name:       "x-show",
				Value:      "$store.cart.items.length > 0",
				IsAlpine:   true,
				AlpineType: "show",
			},
			expectedValue:  "$store.cart.items.length > 0", // Should be UNCHANGED
			expectedName:   "x-show",
			expectedStores: []string{"cart"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset tracker
			InitStoreTracking(map[string]string{})

			// Transform attributes
			attributes := []ast.Attribute{tt.inputAttr}
			dataScope := make(map[string]any)

			result := transformAttributesWithStores(attributes, dataScope)

			// Should have exactly 1 attribute
			if len(result) != 1 {
				t.Fatalf("Expected 1 attribute, got %d", len(result))
			}

			attr := result[0]

			// Check attribute name is unchanged
			if attr.Name != tt.expectedName {
				t.Errorf("Attribute name changed: got %q, want %q", attr.Name, tt.expectedName)
			}

			// CRITICAL: Check attribute value is COMPLETELY UNCHANGED
			if attr.Value != tt.expectedValue {
				t.Errorf("Attribute value was modified:\ngot:  %q\nwant: %q", attr.Value, tt.expectedValue)
			}

			// Verify stores are tracked
			for _, storeName := range tt.expectedStores {
				if !storeTracker.referencedStores[storeName] {
					t.Errorf("Expected store %q to be tracked, but it wasn't", storeName)
				}
			}

			// Check flags are preserved
			if !attr.IsAlpine {
				t.Error("IsAlpine flag should be true")
			}
			if attr.AlpineType != tt.inputAttr.AlpineType {
				t.Errorf("AlpineType changed: got %q, want %q", attr.AlpineType, tt.inputAttr.AlpineType)
			}
		})
	}
}

// TestNonAlpineAttributesWithStoreExpressions tests that non-Alpine attributes are transformed
// This ensures we're not breaking the existing store transformation logic
func TestNonAlpineAttributesWithStoreExpressions(t *testing.T) {
	tests := []struct {
		name          string
		inputAttr     ast.Attribute
		expectedValue string
		expectedName  string
	}{
		{
			name: "class with store expression",
			inputAttr: ast.Attribute{
				Name:     "class",
				Value:    "{$theme.mode}",
				IsAlpine: false,
			},
			expectedValue: "$store.theme.mode",
			expectedName:  ":class", // Should get : prefix
		},
		{
			name: "title with mixed content",
			inputAttr: ast.Attribute{
				Name:     "title",
				Value:    "User: {$user.name}",
				IsAlpine: false,
			},
			expectedValue: "'User: ' + $store.user.name",
			expectedName:  ":title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			InitStoreTracking(map[string]string{})

			attributes := []ast.Attribute{tt.inputAttr}
			dataScope := make(map[string]any)

			result := transformAttributesWithStores(attributes, dataScope)

			if len(result) != 1 {
				t.Fatalf("Expected 1 attribute, got %d", len(result))
			}

			attr := result[0]

			if attr.Name != tt.expectedName {
				t.Errorf("Attribute name: got %q, want %q", attr.Name, tt.expectedName)
			}

			if attr.Value != tt.expectedValue {
				t.Errorf("Attribute value:\ngot:  %q\nwant: %q", attr.Value, tt.expectedValue)
			}
		})
	}
}
