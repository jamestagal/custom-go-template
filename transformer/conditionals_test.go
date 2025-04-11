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
			dataScope: map[string]any{
				"isActive": true,
			},
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
			dataScope: map[string]any{
				"isActive": false,
			},
			contains: []string{
				`<template x-if="isActive"`,
				`This is active`,
				`<template x-else`,
				`This is inactive`,
			},
			notContains: []string{
				`<template x-if="!(isActive)"`,
			},
		},
		{
			name: "if-else-if-else condition",
			condition: &ast.Conditional{
				IfCondition: "status === 'active'",
				IfContent: []ast.Node{
					&ast.TextNode{Content: "This is active"},
				},
				ElseIfConditions: []string{"status === 'pending'"},
				ElseIfContent: [][]ast.Node{
					{
						&ast.TextNode{Content: "This is pending"},
					},
				},
				ElseContent: []ast.Node{
					&ast.TextNode{Content: "This is inactive"},
				},
			},
			dataScope: map[string]any{
				"status": "pending",
			},
			contains: []string{
				`<template x-if="status === 'active'"`,
				`This is active`,
				`<template x-else-if="status === 'pending'"`,
				`This is pending`,
				`<template x-else`,
				`This is inactive`,
			},
			notContains: []string{
				`<template x-if="(!(status === 'active')) && (status === 'pending')"`,
				`<template x-if="!(status === 'active') && !(status === 'pending')"`,
			},
		},
		{
			name: "nested elements in condition",
			condition: &ast.Conditional{
				IfCondition: "isVisible",
				IfContent: []ast.Node{
					&ast.Element{
						TagName: "div",
						Attributes: []ast.Attribute{
							{
								Name:  "class",
								Value: "active",
							},
						},
						Children: []ast.Node{
							&ast.TextNode{Content: "Visible content"},
						},
					},
				},
			},
			dataScope: map[string]any{
				"isVisible": true,
			},
			contains: []string{
				`<template x-if="isVisible"`,
				`<div class="active"`,
				`Visible content`,
			},
		},
		{
			name: "condition with expression",
			condition: &ast.Conditional{
				IfCondition: "count > 0",
				IfContent: []ast.Node{
					&ast.TextNode{Content: "Count is {count}"},
				},
			},
			dataScope: map[string]any{
				"count": 5,
			},
			contains: []string{
				`<template x-if="count > 0"`,
				`Count is <span x-text="count"></span>`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Transform the conditional
			result := transformConditional(tt.condition, tt.dataScope)
			
			// Convert to string for easier testing
			var sb strings.Builder
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

// Helper function to render a node to string for testing
func renderTestNode(sb *strings.Builder, node ast.Node) {
	switch n := node.(type) {
	case *ast.Element:
		sb.WriteString("<")
		sb.WriteString(n.TagName)
		
		// Render attributes
		for _, attr := range n.Attributes {
			sb.WriteString(" ")
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
