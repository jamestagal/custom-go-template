package ast

import (
	"testing"
)

// TestStoreExpressionNode_NodeType tests the NodeType method
func TestStoreExpressionNode_NodeType(t *testing.T) {
	tests := []struct {
		name     string
		node     *StoreExpressionNode
		expected string
	}{
		{
			name: "simple store expression",
			node: &StoreExpressionNode{
				StoreName: "auth",
				Property:  "isLoggedIn",
			},
			expected: "StoreExpression",
		},
		{
			name: "nested property",
			node: &StoreExpressionNode{
				StoreName: "cart",
				Property:  "items.length",
			},
			expected: "StoreExpression",
		},
		{
			name: "store without property",
			node: &StoreExpressionNode{
				StoreName: "theme",
				Property:  "",
			},
			expected: "StoreExpression",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.node.NodeType()
			if result != tt.expected {
				t.Errorf("NodeType() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestStoreExpressionNode_String tests the String method for debug representation
func TestStoreExpressionNode_String(t *testing.T) {
	tests := []struct {
		name     string
		node     *StoreExpressionNode
		expected string
	}{
		{
			name: "simple store with property",
			node: &StoreExpressionNode{
				StoreName: "auth",
				Property:  "isLoggedIn",
			},
			expected: "$auth.isLoggedIn",
		},
		{
			name: "nested property path",
			node: &StoreExpressionNode{
				StoreName: "user",
				Property:  "profile.name",
			},
			expected: "$user.profile.name",
		},
		{
			name: "deeply nested property",
			node: &StoreExpressionNode{
				StoreName: "app",
				Property:  "config.settings.theme.color",
			},
			expected: "$app.config.settings.theme.color",
		},
		{
			name: "store without property",
			node: &StoreExpressionNode{
				StoreName: "cart",
				Property:  "",
			},
			expected: "$cart",
		},
		{
			name: "store with array access",
			node: &StoreExpressionNode{
				StoreName: "items",
				Property:  "list[0].name",
			},
			expected: "$items.list[0].name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.node.String()
			if result != tt.expected {
				t.Errorf("String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestStoreExpressionNode_NodeInterface tests that StoreExpressionNode implements Node interface
func TestStoreExpressionNode_NodeInterface(t *testing.T) {
	// This test ensures StoreExpressionNode implements the Node interface
	var _ Node = (*StoreExpressionNode)(nil)
}

// TestStoreExpressionNode_Creation tests creating various store expression nodes
func TestStoreExpressionNode_Creation(t *testing.T) {
	tests := []struct {
		name          string
		storeName     string
		property      string
		expectedStore string
		expectedProp  string
	}{
		{
			name:          "auth store",
			storeName:     "auth",
			property:      "loggedIn",
			expectedStore: "auth",
			expectedProp:  "loggedIn",
		},
		{
			name:          "cart store with count",
			storeName:     "cart",
			property:      "itemCount",
			expectedStore: "cart",
			expectedProp:  "itemCount",
		},
		{
			name:          "theme store",
			storeName:     "theme",
			property:      "darkMode",
			expectedStore: "theme",
			expectedProp:  "darkMode",
		},
		{
			name:          "empty property",
			storeName:     "settings",
			property:      "",
			expectedStore: "settings",
			expectedProp:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &StoreExpressionNode{
				StoreName: tt.storeName,
				Property:  tt.property,
			}

			if node.StoreName != tt.expectedStore {
				t.Errorf("StoreName = %q, want %q", node.StoreName, tt.expectedStore)
			}
			if node.Property != tt.expectedProp {
				t.Errorf("Property = %q, want %q", node.Property, tt.expectedProp)
			}
		})
	}
}

// TestStoreExpressionNode_CommonUseCases tests common real-world scenarios
func TestStoreExpressionNode_CommonUseCases(t *testing.T) {
	tests := []struct {
		name           string
		node           *StoreExpressionNode
		expectedString string
		expectedType   string
	}{
		{
			name: "authentication check",
			node: &StoreExpressionNode{
				StoreName: "auth",
				Property:  "user.isAdmin",
			},
			expectedString: "$auth.user.isAdmin",
			expectedType:   "StoreExpression",
		},
		{
			name: "shopping cart total",
			node: &StoreExpressionNode{
				StoreName: "cart",
				Property:  "total",
			},
			expectedString: "$cart.total",
			expectedType:   "StoreExpression",
		},
		{
			name: "theme preference",
			node: &StoreExpressionNode{
				StoreName: "theme",
				Property:  "current",
			},
			expectedString: "$theme.current",
			expectedType:   "StoreExpression",
		},
		{
			name: "notification count",
			node: &StoreExpressionNode{
				StoreName: "notifications",
				Property:  "unread.length",
			},
			expectedString: "$notifications.unread.length",
			expectedType:   "StoreExpression",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test String() method
			result := tt.node.String()
			if result != tt.expectedString {
				t.Errorf("String() = %q, want %q", result, tt.expectedString)
			}

			// Test NodeType() method
			nodeType := tt.node.NodeType()
			if nodeType != tt.expectedType {
				t.Errorf("NodeType() = %q, want %q", nodeType, tt.expectedType)
			}
		})
	}
}

// TestStoreExpressionNode_EdgeCases tests edge cases and boundary conditions
func TestStoreExpressionNode_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		node     *StoreExpressionNode
		testFunc func(t *testing.T, node *StoreExpressionNode)
	}{
		{
			name: "empty store name",
			node: &StoreExpressionNode{
				StoreName: "",
				Property:  "something",
			},
			testFunc: func(t *testing.T, node *StoreExpressionNode) {
				result := node.String()
				// Should handle gracefully, even with empty store name
				expected := "$.something"
				if result != expected {
					t.Errorf("String() with empty StoreName = %q, want %q", result, expected)
				}
			},
		},
		{
			name: "both empty",
			node: &StoreExpressionNode{
				StoreName: "",
				Property:  "",
			},
			testFunc: func(t *testing.T, node *StoreExpressionNode) {
				result := node.String()
				expected := "$"
				if result != expected {
					t.Errorf("String() with both empty = %q, want %q", result, expected)
				}
			},
		},
		{
			name: "property with special characters",
			node: &StoreExpressionNode{
				StoreName: "data",
				Property:  "items[0].name",
			},
			testFunc: func(t *testing.T, node *StoreExpressionNode) {
				result := node.String()
				expected := "$data.items[0].name"
				if result != expected {
					t.Errorf("String() with array access = %q, want %q", result, expected)
				}
			},
		},
		{
			name: "property with multiple dots",
			node: &StoreExpressionNode{
				StoreName: "nested",
				Property:  "a.b.c.d.e",
			},
			testFunc: func(t *testing.T, node *StoreExpressionNode) {
				result := node.String()
				expected := "$nested.a.b.c.d.e"
				if result != expected {
					t.Errorf("String() with deep nesting = %q, want %q", result, expected)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t, tt.node)
		})
	}
}
