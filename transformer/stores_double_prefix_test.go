package transformer

import (
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
)

// TestTransformAlreadyTransformedStoreExpressions tests the critical bug fix
// where $store.theme.mode patterns should NOT be transformed again
// This prevents the bug: $store.theme.mode -> $store.store.theme.mode
func TestTransformAlreadyTransformedStoreExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "already transformed condition - simple",
			input:    "$store.theme.mode",
			expected: "$store.theme.mode",
		},
		{
			name:     "already transformed condition - with comparison",
			input:    "$store.theme.mode === 'dark'",
			expected: "$store.theme.mode === 'dark'",
		},
		{
			name:     "already transformed condition - complex expression",
			input:    "$store.theme.mode === 'dark' && $store.auth.isLoggedIn",
			expected: "$store.theme.mode === 'dark' && $store.auth.isLoggedIn",
		},
		{
			name:     "mixed - some transformed, some not",
			input:    "$store.theme.mode === 'dark' && $auth.isLoggedIn",
			expected: "$store.theme.mode === 'dark' && $store.auth.isLoggedIn",
		},
		{
			name:     "already transformed - method call",
			input:    "$store.theme.getCurrentColors().background",
			expected: "$store.theme.getCurrentColors().background",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := transformStoreExpressionsInCondition(tt.input)
			
			if result != tt.expected {
				t.Errorf("transformStoreExpressionsInCondition() failed\nInput:    %q\nExpected: %q\nGot:      %q", 
					tt.input, tt.expected, result)
			}
			
			// Extra validation: ensure we never create $store.store.* patterns
			if strings.Contains(result, "$store.store.") {
				t.Errorf("BUG: Created double prefix $store.store.* in result: %q", result)
			}
		})
	}
}

// TestTransformAlreadyTransformedCollections tests collection transformation
// with already-transformed $store.* patterns
func TestTransformAlreadyTransformedCollections(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "already transformed collection",
			input:    "$store.cart.items",
			expected: "$store.cart.items",
		},
		{
			name:     "already transformed nested collection",
			input:    "$store.user.profile.wishlist.products",
			expected: "$store.user.profile.wishlist.products",
		},
		{
			name:     "not transformed collection",
			input:    "$cart.items",
			expected: "$store.cart.items",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := transformStoreExpressionInCollection(tt.input)
			
			if result != tt.expected {
				t.Errorf("transformStoreExpressionInCollection() failed\nInput:    %q\nExpected: %q\nGot:      %q", 
					tt.input, tt.expected, result)
			}
			
			// Extra validation: ensure we never create $store.store.* patterns
			if strings.Contains(result, "$store.store.") {
				t.Errorf("BUG: Created double prefix $store.store.* in result: %q", result)
			}
		})
	}
}

// TestAlpineAttributeWithStoreReference tests the real-world scenario
// from the bug report: :style attributes with $store.theme.* references
func TestAlpineAttributeWithStoreReference(t *testing.T) {
	tests := []struct {
		name     string
		attr     ast.Attribute
		expected string // Expected attribute value after transformation
	}{
		{
			name: ":style with $store.theme reference",
			attr: ast.Attribute{
				Name:       ":style",
				Value:      "`background-color: ${$store.theme.getCurrentColors().background};`",
				Dynamic:    false,
				IsAlpine:   true,
				AlpineType: "bind",
				AlpineKey:  "style",
			},
			expected: "`background-color: ${$store.theme.getCurrentColors().background};`",
		},
		{
			name: "@click with $store.theme reference",
			attr: ast.Attribute{
				Name:       "@click",
				Value:      "$store.theme.setLight()",
				Dynamic:    false,
				IsAlpine:   true,
				AlpineType: "on",
				AlpineKey:  "click",
			},
			expected: "$store.theme.setLight()",
		},
		{
			name: "x-show with $store.auth reference",
			attr: ast.Attribute{
				Name:       "x-show",
				Value:      "$store.auth.isLoggedIn",
				Dynamic:    false,
				IsAlpine:   true,
				AlpineType: "show",
			},
			expected: "$store.auth.isLoggedIn",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dataScope := make(map[string]any)
			attributes := []ast.Attribute{tt.attr}
			
			result := transformAttributesWithStores(attributes, dataScope)
			
			if len(result) != 1 {
				t.Fatalf("Expected 1 attribute, got %d", len(result))
			}
			
			if result[0].Value != tt.expected {
				t.Errorf("Attribute value changed unexpectedly\nOriginal: %q\nExpected: %q\nGot:      %q",
					tt.attr.Value, tt.expected, result[0].Value)
			}
			
			// Extra validation: ensure we never create $store.store.* patterns
			if strings.Contains(result[0].Value, "$store.store.") {
				t.Errorf("BUG: Created double prefix $store.store.* in attribute value: %q", result[0].Value)
			}
		})
	}
}

// TestStoreTrackingWithAlreadyTransformed tests that already-transformed
// store references are still tracked correctly
func TestStoreTrackingWithAlreadyTransformed(t *testing.T) {
	// Reset tracker
	InitStoreTracking(map[string]string{
		"theme": "{}",
		"auth":  "{}",
		"cart":  "{}",
	})

	// Test condition with already-transformed references
	condition := "$store.theme.mode === 'dark' && $store.auth.isLoggedIn"
	transformStoreExpressionsInCondition(condition)

	// Get tracked stores
	referencedStores, _ := GetTrackedStores(&ast.Template{})

	// Should have tracked both theme and auth
	storeMap := make(map[string]bool)
	for _, storeName := range referencedStores {
		storeMap[storeName] = true
	}

	if !storeMap["theme"] {
		t.Error("Expected 'theme' store to be tracked")
	}
	if !storeMap["auth"] {
		t.Error("Expected 'auth' store to be tracked")
	}
	if storeMap["store"] {
		t.Error("Should NOT track 'store' as a store name")
	}
}
