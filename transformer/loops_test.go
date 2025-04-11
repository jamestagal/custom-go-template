package transformer

import (
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
)

func TestTransformLoop(t *testing.T) {
	tests := []struct {
		name      string
		loop      *ast.Loop
		dataScope map[string]any
		contains  []string
		notContains []string
	}{
		{
			name: "simple array loop",
			loop: &ast.Loop{
				Iterator:   "item",
				Collection: "items",
				Content: []ast.Node{
					&ast.TextNode{Content: "Item: {item}"},
				},
				IsOf: false, // "in" loop
			},
			dataScope: map[string]any{
				"items": []string{"one", "two", "three"},
			},
			contains: []string{
				`<template x-for="item in items"`,
				`Item: <span x-text="item"></span>`,
			},
		},
		{
			name: "array loop with index",
			loop: &ast.Loop{
				Iterator:   "index",
				Value:      "item",
				Collection: "items",
				Content: []ast.Node{
					&ast.TextNode{Content: "{index}: {item}"},
				},
				IsOf: false, // "in" loop
			},
			dataScope: map[string]any{
				"items": []string{"one", "two", "three"},
			},
			contains: []string{
				`<template x-for="(index, item) in items"`,
				`<span x-text="index"></span>: <span x-text="item"></span>`,
			},
		},
		{
			name: "object loop",
			loop: &ast.Loop{
				Iterator:   "key",
				Value:      "value",
				Collection: "user",
				Content: []ast.Node{
					&ast.TextNode{Content: "{key}: {value}"},
				},
				IsOf: true, // "of" loop
			},
			dataScope: map[string]any{
				"user": map[string]any{
					"name": "John",
					"age":  30,
				},
			},
			contains: []string{
				`<template x-for="key, value of Object.entries(user)"`,
				`<span x-text="key"></span>: <span x-text="value"></span>`,
			},
		},
		{
			name: "nested element in loop",
			loop: &ast.Loop{
				Iterator:   "item",
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
							&ast.TextNode{Content: "{item}"},
						},
					},
				},
				IsOf: false, // "in" loop
			},
			dataScope: map[string]any{
				"items": []string{"one", "two", "three"},
			},
			contains: []string{
				`<template x-for="item in items"`,
				`<li class="item"><span x-text="item"></span></li>`,
			},
		},
		{
			name: "loop with conditional",
			loop: &ast.Loop{
				Iterator:   "item",
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
				"items": []map[string]any{
					{"name": "Item 1", "active": true},
					{"name": "Item 2", "active": false},
				},
			},
			contains: []string{
				`<template x-for="item in items"`,
				`<template x-if="item.active">`,
				`<span x-text="item.name"></span> (Active)`,
				`<template x-else>`,
				`<span x-text="item.name"></span>`,
			},
			notContains: []string{
				`<template x-if="!(item.active)">`,
			},
		},
		{
			name: "loop with complex expression",
			loop: &ast.Loop{
				Iterator:   "i",
				Collection: "Array(count)",
				Content: []ast.Node{
					&ast.TextNode{Content: "Item {i+1}"},
				},
				IsOf: false, // "in" loop
			},
			dataScope: map[string]any{
				"count": 3,
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
					t.Errorf("Expected output not to contain %q, but it does.\nOutput: %s", s, output)
				}
			}
		})
	}
}
