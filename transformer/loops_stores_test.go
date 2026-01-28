package transformer

import (
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
)

// Helper function to render nodes to string for testing
// Cognitive Load: 8 (recursive rendering with string builder)
func renderNodesToString(nodes []ast.Node) string {
	var sb strings.Builder
	for _, node := range nodes {
		renderTestNodeForLoops(&sb, node)
	}
	return sb.String()
}

// Helper function to render a single node to string
// Cognitive Load: 10 (recursive node rendering)
func renderTestNodeForLoops(sb *strings.Builder, node ast.Node) {
	switch n := node.(type) {
	case *ast.Element:
		sb.WriteString("<")
		sb.WriteString(n.TagName)

		for _, attr := range n.Attributes {
			sb.WriteString(" ")

			// CRITICAL FIX: Add : prefix for dynamic attributes (Alpine.js bind syntax)
			// Dynamic non-Alpine attributes need : prefix for runtime binding
			if attr.Dynamic && !attr.IsAlpine {
				sb.WriteString(":")
			}

			sb.WriteString(attr.Name)
			if attr.Value != "" {
				sb.WriteString("=\"")
				sb.WriteString(attr.Value)
				sb.WriteString("\"")
			}
		}

		if n.SelfClosing {
			sb.WriteString(" />")
			return
		}

		sb.WriteString(">")

		for _, child := range n.Children {
			renderTestNodeForLoops(sb, child)
		}

		sb.WriteString("</")
		sb.WriteString(n.TagName)
		sb.WriteString(">")

	case *ast.TextNode:
		sb.WriteString(n.Content)

	case *ast.ExpressionNode:
		sb.WriteString(`<span x-text="`)
		sb.WriteString(n.Expression)
		sb.WriteString(`"></span>`)

	case *ast.StoreExpressionNode:
		sb.WriteString(`<span x-text="$store.`)
		sb.WriteString(n.StoreName)
		if n.Property != "" {
			sb.WriteString(".")
			sb.WriteString(n.Property)
		}
		sb.WriteString(`"></span>`)
	}
}

// TestTransformLoopWithStoreCollection tests loop transformation with store collections
// Pattern: Integration Test [Cognitive Load: 12]
//
// This test verifies:
// 1. Store collections ($store.cart.items) are transformed correctly
// 2. Store expressions in loop body are handled properly
// 3. Regular collections with store expressions in body work correctly
// 4. Nested store property access ($store.user.profile.wishlist.products)
func TestTransformLoopWithStoreCollection(t *testing.T) {
	tests := []struct {
		name      string
		loop      *ast.Loop
		dataScope map[string]any
		contains  []string
	}{
		{
			name: "simple store collection",
			loop: &ast.Loop{
				Value:      "item",
				Iterator:   "",
				Collection: "$cart.items",
				Content: []ast.Node{
					&ast.Element{
						TagName: "div",
						Children: []ast.Node{
							&ast.ExpressionNode{Expression: "item.name"},
						},
					},
				},
				IsOf: false,
			},
			dataScope: map[string]any{},
			contains: []string{
				`<template x-for="item in $store.cart.items"`,
				`<span x-text="item.name"></span>`,
			},
		},
		{
			name: "store collection with nested property access",
			loop: &ast.Loop{
				Value:      "product", // FIXED
				Iterator:   "",        // FIXED
				Collection: "$user.profile.wishlist.products",
				Content: []ast.Node{
					&ast.Element{
						TagName: "div",
						Children: []ast.Node{
							&ast.ExpressionNode{Expression: "product.title"},
						},
					},
				},
				IsOf: false,
			},
			dataScope: map[string]any{},
			contains: []string{
				`<template x-for="product in $store.user.profile.wishlist.products"`,
				`<span x-text="product.title"></span>`,
			},
		},
		{
			name: "loop with store collection and index",
			loop: &ast.Loop{
				Value:      "item",  // FIXED: swapped from Iterator
				Iterator:   "index", // FIXED: swapped from Value
				Collection: "$cart.items",
				Content: []ast.Node{
					&ast.Element{
						TagName: "div",
						Children: []ast.Node{
							&ast.ExpressionNode{Expression: "index"},
							&ast.TextNode{Content: ": "},
							&ast.ExpressionNode{Expression: "item.name"},
						},
					},
				},
				IsOf: false,
			},
			dataScope: map[string]any{},
			contains: []string{
				`<template x-for="(item, index) in $store.cart.items"`,
				`<span x-text="index"></span>: <span x-text="item.name"></span>`,
			},
		},
		{
			name: "loop body with store expressions",
			loop: &ast.Loop{
				Value:      "item", // FIXED
				Iterator:   "",     // FIXED
				Collection: "$cart.items",
				Content: []ast.Node{
					&ast.Element{
						TagName: "div",
						Children: []ast.Node{
							&ast.ExpressionNode{Expression: "item.name"},
							&ast.TextNode{Content: " - Buyer: "},
							&ast.StoreExpressionNode{StoreName: "auth", Property: "user.name"},
						},
					},
				},
				IsOf: false,
			},
			dataScope: map[string]any{},
			contains: []string{
				`<template x-for="item in $store.cart.items"`,
				`<span x-text="item.name"></span> - Buyer: <span x-text="$store.auth.user.name"></span>`,
			},
		},
		{
			name: "mixed: regular collection, store in body",
			loop: &ast.Loop{
				Value:      "item",  // FIXED
				Iterator:   "",      // FIXED
				Collection: "items", // Regular collection
				Content: []ast.Node{
					&ast.Element{
						TagName: "div",
						Children: []ast.Node{
							&ast.ExpressionNode{Expression: "item.name"},
							&ast.TextNode{Content: " ("},
							&ast.StoreExpressionNode{StoreName: "theme", Property: "mode"},
							&ast.TextNode{Content: ")"},
						},
					},
				},
				IsOf: false,
			},
			dataScope: map[string]any{
				// BUILD-TIME EXPANSION: Collection is resolvable, should expand at build time
				// Store expressions in body are fine - they become $store.theme.mode in output
				"items": []map[string]any{
					{"name": "a"},
					{"name": "b"},
				},
			},
			contains: []string{
				// BUILD-TIME EXPANSION: item.name resolves to actual values (a, b)
				// Store expressions remain as runtime ($store.theme.mode)
				`<div>a (<span x-text="$store.theme.mode"></span>)</div>`,
				`<div>b (<span x-text="$store.theme.mode"></span>)</div>`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := transformLoop(tt.loop, tt.dataScope)

			// Render the result to HTML string for checking
			html := renderNodesToString(result)

			// Check all expected strings are present
			for _, expected := range tt.contains {
				if !strings.Contains(html, expected) {
					t.Errorf("Expected output to contain %q\nGot:\n%s", expected, html)
				}
			}
		})
	}
}

// TestTransformNestedLoopsWithStores tests nested loops containing store expressions
// Cognitive Load: 10 (nested loop scenarios)
func TestTransformNestedLoopsWithStores(t *testing.T) {
	tests := []struct {
		name      string
		loop      *ast.Loop
		dataScope map[string]any
		contains  []string
	}{
		{
			name: "nested loops: outer store, inner regular",
			loop: &ast.Loop{
				Value:      "category", // FIXED
				Iterator:   "",         // FIXED
				Collection: "$catalog.categories",
				Content: []ast.Node{
					&ast.Element{
						TagName: "div",
						Children: []ast.Node{
							&ast.ExpressionNode{Expression: "category.name"},
							&ast.Loop{
								Value:      "product", // FIXED
								Iterator:   "",        // FIXED
								Collection: "category.products",
								Content: []ast.Node{
									&ast.Element{
										TagName: "span",
										Children: []ast.Node{
											&ast.ExpressionNode{Expression: "product.title"},
										},
									},
								},
								IsOf: false,
							},
						},
					},
				},
				IsOf: false,
			},
			dataScope: map[string]any{},
			contains: []string{
				`<template x-for="category in $store.catalog.categories"`,
				`<span x-text="category.name"></span>`,
				// Inner loop is also runtime because it references category.products (runtime variable)
				`<template x-for="product in category.products"`,
				`<span x-text="product.title"></span>`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := transformLoop(tt.loop, tt.dataScope)

			// Render the result to HTML string for checking
			html := renderNodesToString(result)

			// Check all expected strings are present
			for _, expected := range tt.contains {
				if !strings.Contains(html, expected) {
					t.Errorf("Expected output to contain %q\nGot:\n%s", expected, html)
				}
			}
		})
	}
}
