package transformer

import (
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
)

func TestTransformConditional(t *testing.T) {
	tests := []struct {
		name      string
		condition *ast.Conditional
		dataScope map[string]any
		contains  []string
		notContains []string
	}{
		{
			name: "simple if condition",
			condition: &ast.Conditional{
				IfCondition: "isActive",
				IfContent: []ast.Node{
					&ast.TextNode{Content: "This is active"},
				},
			},
			// NOTE: Don't include isActive in dataScope to force runtime x-if generation
			// If isActive is in dataScope with a value, the conditional is resolved at build-time
			dataScope: map[string]any{},
			contains: []string{
				`<template x-if="isActive"`,
				`This is active`,
			},
		},
		{
			name: "if-else condition",
			condition: &ast.Conditional{
				IfCondition: "isActive",
				IfContent: []ast.Node{
					&ast.TextNode{Content: "This is active"},
				},
				ElseContent: []ast.Node{
					&ast.TextNode{Content: "This is inactive"},
				},
			},
			// NOTE: Don't include isActive in dataScope to force runtime x-if generation
			// If isActive is in dataScope, the conditional is resolved at build-time
			dataScope: map[string]any{},
			contains: []string{
				`<template x-if="isActive"`,
				`This is active`,
				// Alpine.js 3.x: else uses negated x-if
				`<template x-if="!(isActive)"`,
				`This is inactive`,
			},
			notContains: []string{
				// Alpine.js 3.x doesn't support x-else directive
				`<template x-else`,
			},
		},
		{
			name: "if-else-if-else condition",
			condition: &ast.Conditional{
				IfCondition: "status === 'active'",
				IfContent: []ast.Node{
					&ast.TextNode{Content: "Active"},
				},
				ElseIfConditions: []string{"status === 'pending'"},
				ElseIfContent: [][]ast.Node{
					{
						&ast.TextNode{Content: "Pending"},
					},
				},
				ElseContent: []ast.Node{
					&ast.TextNode{Content: "Inactive"},
				},
			},
			// NOTE: Don't include status in dataScope to force runtime evaluation
			dataScope: map[string]any{},
			contains: []string{
				`<template x-if="status === 'active'"`,
				`Active`,
				// Alpine.js 3.x: else-if uses combined negated x-if
				`<template x-if="(!(status === 'active')) && (status === 'pending')"`,
				`Pending`,
				// Alpine.js 3.x: else uses negated all previous conditions
				// Note: output format is !(x) && !(y) without extra parens
				`!(status === 'active') && !(status === 'pending')`,
				`Inactive`,
			},
			notContains: []string{
				// Alpine.js 3.x doesn't support x-else-if or x-else directives
				`x-else-if`,
				`<template x-else`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := transformConditional(tt.condition, tt.dataScope)

			// Render to string
			var sb strings.Builder
			for _, node := range result {
				renderTestNode(&sb, node)
			}
			output := sb.String()

			// Check that output contains expected strings
			for _, expected := range tt.contains {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %q\nGot: %s", expected, output)
				}
			}

			// Check that output doesn't contain unexpected strings
			for _, unexpected := range tt.notContains {
				if strings.Contains(output, unexpected) {
					t.Errorf("Expected output to NOT contain %q\nGot: %s", unexpected, output)
				}
			}
		})
	}
}

func TestTransformConditionalEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		condition *ast.Conditional
		dataScope map[string]any
		contains  []string
	}{
		{
			name: "conditional with complex expression",
			condition: &ast.Conditional{
				IfCondition: "count > 0 && isValid",
				IfContent: []ast.Node{
					&ast.TextNode{Content: "Valid items"},
				},
			},
			dataScope: map[string]any{
				"count":   5,
				"isValid": true,
			},
			contains: []string{
				`<template x-if="count > 0 && isValid"`,
			},
		},
		{
			name: "conditional with nested elements",
			condition: &ast.Conditional{
				IfCondition: "showContent",
				IfContent: []ast.Node{
					&ast.Element{
						TagName: "div",
						Children: []ast.Node{
							&ast.TextNode{Content: "Nested content"},
						},
					},
				},
			},
			// NOTE: Don't include showContent in dataScope to force runtime x-if generation
			dataScope: map[string]any{},
			contains: []string{
				`<template x-if="showContent"`,
				`<div>`,
				`Nested content`,
				`</div>`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := transformConditional(tt.condition, tt.dataScope)

			// Render to string
			var sb strings.Builder
			for _, node := range result {
				renderTestNode(&sb, node)
			}
			output := sb.String()

			// Check that output contains expected strings
			for _, expected := range tt.contains {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %q\nGot: %s", expected, output)
				}
			}
		})
	}
}

// Helper function to render a node to string for testing
func renderTestNode(sb *strings.Builder, node ast.Node) {
	switch n := node.(type) {
	case *ast.Element:
		sb.WriteString("<")
		sb.WriteString(n.TagName)

		// Render attributes
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

		// Render children
		for _, child := range n.Children {
			renderTestNode(sb, child)
		}

		sb.WriteString("</")
		sb.WriteString(n.TagName)
		sb.WriteString(">")

	case *ast.TextNode:
		sb.WriteString(n.Content)

	case *ast.ExpressionNode:
		sb.WriteString("<span x-text=\"")
		sb.WriteString(n.Expression)
		sb.WriteString("\"></span>")
	}
}
