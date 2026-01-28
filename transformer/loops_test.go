package transformer

import (
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
)

func TestTransformLoop(t *testing.T) {
	tests := []struct {
		name        string
		loop        *ast.Loop
		dataScope   map[string]any
		contains    []string
		notContains []string
	}{
		{
			name: "simple array loop",
			loop: &ast.Loop{
				Value:      "item", // FIXED: item variable goes in Value field
				Iterator:   "",     // FIXED: no index variable
				Collection: "items",
				Content: []ast.Node{
					&ast.ExpressionNode{Expression: "item"}, // Use ExpressionNode, not TextNode with {item}
				},
				IsOf: false, // "in" loop
			},
			dataScope: map[string]any{
				// Don't provide "items" data - force runtime x-for
			},
			contains: []string{
				`<template x-for="item in items"`,
				`<span x-text="item"></span>`,
			},
		},
		{
			name: "array loop with index",
			loop: &ast.Loop{
				Value:      "item",  // FIXED: item variable is in Value
				Iterator:   "index", // FIXED: index variable is in Iterator
				Collection: "items",
				Content: []ast.Node{
					&ast.ExpressionNode{Expression: "index"},
					&ast.TextNode{Content: ": "},
					&ast.ExpressionNode{Expression: "item"},
				},
				IsOf: false, // "in" loop
			},
			dataScope: map[string]any{
				// Don't provide "items" data - force runtime x-for
			},
			contains: []string{
				`<template x-for="(item, index) in items"`,
				`<span x-text="index"></span>: <span x-text="item"></span>`,
			},
		},
		{
			name: "object loop",
			loop: &ast.Loop{
				Value:      "value", // FIXED: value variable
				Iterator:   "key",   // FIXED: key (iterator) for object loops
				Collection: "user",
				Content: []ast.Node{
					&ast.ExpressionNode{Expression: "key"},
					&ast.TextNode{Content: ": "},
					&ast.ExpressionNode{Expression: "value"},
				},
				IsOf: true, // "of" loop
			},
			dataScope: map[string]any{
				// Don't provide "user" data - force runtime x-for
			},
			contains: []string{
				`<template x-for="(value, key) in user"`,
				`<span x-text="key"></span>: <span x-text="value"></span>`,
			},
		},
		{
			name: "nested element in loop",
			loop: &ast.Loop{
				Value:      "item", // FIXED
				Iterator:   "",     // FIXED: no index
				Collection: "items",
				Content: []ast.Node{
					&ast.Element{
						TagName: "li",
						Attributes: []ast.Attribute{
							{
								Name:  "class",
								Value: "item",
							},
						},
						Children: []ast.Node{
							&ast.ExpressionNode{Expression: "item"},
						},
					},
				},
				IsOf: false, // "in" loop
			},
			dataScope: map[string]any{
				// Don't provide "items" data - force runtime x-for
			},
			contains: []string{
				`<template x-for="item in items"`,
				`<li class="item"><span x-text="item"></span></li>`,
			},
		},
		{
			name: "loop with conditional",
			loop: &ast.Loop{
				Value:      "item", // FIXED
				Iterator:   "",     // FIXED: no index
				Collection: "items",
				Content: []ast.Node{
					&ast.Conditional{
						IfCondition: "item.active",
						IfContent: []ast.Node{
							&ast.ExpressionNode{Expression: "item.name"},
							&ast.TextNode{Content: " (Active)"},
						},
						ElseContent: []ast.Node{
							&ast.ExpressionNode{Expression: "item.name"},
						},
					},
				},
				IsOf: false, // "in" loop
			},
			dataScope: map[string]any{
				// Don't provide "items" data - force runtime x-for
			},
			contains: []string{
				`<template x-for="item in items"`,
				`<template x-if="item.active">`,
				`<span x-text="item.name"></span> (Active)`,
				// Alpine.js 3.x: else uses negated x-if
				`<template x-if="!(item.active)"`,
				`<span x-text="item.name"></span>`,
			},
			notContains: []string{
				// Alpine.js 3.x doesn't support x-else directive
				`<template x-else`,
			},
		},
		{
			name: "loop with complex expression",
			loop: &ast.Loop{
				Value:      "i", // FIXED
				Iterator:   "",  // FIXED: no index
				Collection: "Array(count)",
				Content: []ast.Node{
					&ast.TextNode{Content: "Item "},
					&ast.ExpressionNode{Expression: "i+1"},
				},
				IsOf: false, // "in" loop
			},
			dataScope: map[string]any{
				"count": 3,
				// Array(count) can't be resolved at build-time, so runtime x-for
			},
			contains: []string{
				`<template x-for="i in Array(count)"`,
				`Item <span x-text="i+1"></span>`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Transform the loop
			result := transformLoop(tt.loop, tt.dataScope)

			// Convert to string for easier testing
			var sb strings.Builder

			// Render the nodes returned by transformLoop
			for _, node := range result {
				renderTestNode(&sb, node)
			}

			output := sb.String()

			// Check that output contains expected strings
			for _, s := range tt.contains {
				if !strings.Contains(output, s) {
					t.Errorf("Expected output to contain %q, but it doesn't.\nOutput: %s", s, output)
				}
			}

			// Check that output doesn't contain unwanted strings
			for _, s := range tt.notContains {
				if strings.Contains(output, s) {
					t.Errorf("Expected output to NOT contain %q, but it does.\nOutput: %s", s, output)
				}
			}
		})
	}
}
