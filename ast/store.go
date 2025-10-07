package ast

import "fmt"

// StoreExpressionNode represents a global store reference in template syntax.
// Syntax: {$storeName.property} or {$storeName.nested.property}
// Example: {$auth.isLoggedIn}, {$cart.items.length}
//
// This transforms to Alpine.js: <span x-text="$store.storeName.property"></span>
type StoreExpressionNode struct {
	StoreName string // The name of the store (e.g., "auth", "cart", "theme")
	Property  string // The property path within the store (e.g., "isLoggedIn", "items.length")
}

// NodeType returns the node type identifier for StoreExpressionNode
func (s *StoreExpressionNode) NodeType() string {
	return "StoreExpression"
}

// String returns a debug representation of the store expression
// Format: $storeName.property
func (s *StoreExpressionNode) String() string {
	if s.Property == "" {
		return fmt.Sprintf("$%s", s.StoreName)
	}
	return fmt.Sprintf("$%s.%s", s.StoreName, s.Property)
}
