package builder

import (
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
)

// TestLoopVariableExpressions verifies that expressions referencing loop variables
// are converted to Alpine.js directives instead of template literals
func TestLoopVariableExpressions(t *testing.T) {
	tests := []struct {
		name     string
		template *ast.Template
		want     string
	}{
		{
			name: "simple loop with text expression",
			template: &ast.Template{
				RootNodes: []ast.Node{
					&ast.Loop{
						Value:      "text",
						Collection: "textGroup",
						Content: []ast.Node{
							&ast.Element{
								TagName: "p",
								Children: []ast.Node{
									&ast.ExpressionNode{Expression: "text"},
								},
							},
						},
					},
				},
			},
			want: `<template x-for="text in props.textGroup"><p><span x-text="text"></span></p></template>`,
		},
		{
			name: "loop with attribute expression",
			template: &ast.Template{
				RootNodes: []ast.Node{
					&ast.Loop{
						Value:      "card",
						Collection: "cards",
						Content: []ast.Node{
							&ast.Element{
								TagName: "img",
								Attributes: []ast.Attribute{
									{Name: "src", Value: "{card.icon.src}"},
									{Name: "alt", Value: "{card.icon.alt}"},
								},
								SelfClosing: true,
							},
						},
					},
				},
			},
			want: `<template x-for="card in props.cards"><img :src="card.icon.src" :alt="card.icon.alt"></template>`,
		},
		{
			name: "loop with both text and attribute expressions",
			template: &ast.Template{
				RootNodes: []ast.Node{
					&ast.Loop{
						Value:      "item",
						Collection: "items",
						Content: []ast.Node{
							&ast.Element{
								TagName: "div",
								Children: []ast.Node{
									&ast.Element{
										TagName: "h3",
										Children: []ast.Node{
											&ast.ExpressionNode{Expression: "item.title"},
										},
									},
									&ast.Element{
										TagName: "img",
										Attributes: []ast.Attribute{
											{Name: "src", Value: "{item.image}"},
										},
										SelfClosing: true,
									},
								},
							},
						},
					},
				},
			},
			want: `<template x-for="item in props.items"><div><h3><span x-text="item.title"></span></h3><img :src="item.image"></div></template>`,
		},
		{
			name: "non-loop expression uses template literal",
			template: &ast.Template{
				RootNodes: []ast.Node{
					&ast.Element{
						TagName: "h1",
						Children: []ast.Node{
							&ast.ExpressionNode{Expression: "title"},
						},
					},
				},
			},
			want: `<h1>${props.title}</h1>`,
		},
		{
			name: "nested property access in loop",
			template: &ast.Template{
				RootNodes: []ast.Node{
					&ast.Loop{
						Value:      "product",
						Collection: "products",
						Content: []ast.Node{
							&ast.Element{
								TagName: "div",
								Children: []ast.Node{
									&ast.ExpressionNode{Expression: "product.details.name"},
								},
							},
						},
					},
				},
			},
			want: `<template x-for="product in props.products"><div><span x-text="product.details.name"></span></div></template>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertToJSTemplate(tt.template)
			// Remove whitespace for comparison
			gotClean := removeWhitespace(got)
			wantClean := removeWhitespace(tt.want)

			if gotClean != wantClean {
				t.Errorf("\nGot:\n%s\n\nWant:\n%s", got, tt.want)
			}
		})
	}
}

// TestExpressionReferencesLoopVar verifies the helper function correctly identifies loop variables
func TestExpressionReferencesLoopVar(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		loopVars map[string]bool
		want     bool
	}{
		{
			name:     "simple loop variable",
			expr:     "text",
			loopVars: map[string]bool{"text": true},
			want:     true,
		},
		{
			name:     "property access on loop variable",
			expr:     "card.icon.src",
			loopVars: map[string]bool{"card": true},
			want:     true,
		},
		{
			name:     "non-loop variable",
			expr:     "props.title",
			loopVars: map[string]bool{"card": true},
			want:     false,
		},
		{
			name:     "complex expression with loop variable",
			expr:     "item.name + ' - ' + item.price",
			loopVars: map[string]bool{"item": true},
			want:     true,
		},
		{
			name:     "partial match should not trigger",
			expr:     "discard",
			loopVars: map[string]bool{"card": true},
			want:     false,
		},
		{
			name:     "empty loop vars",
			expr:     "text",
			loopVars: map[string]bool{},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expressionReferencesLoopVar(tt.expr, tt.loopVars)
			if got != tt.want {
				t.Errorf("expressionReferencesLoopVar(%q, %v) = %v, want %v",
					tt.expr, tt.loopVars, got, tt.want)
			}
		})
	}
}

// TestAttributeReferencesLoopVar verifies attribute loop variable detection
func TestAttributeReferencesLoopVar(t *testing.T) {
	tests := []struct {
		name      string
		attrValue string
		loopVars  map[string]bool
		want      bool
	}{
		{
			name:      "simple expression with loop var",
			attrValue: "{card.icon.src}",
			loopVars:  map[string]bool{"card": true},
			want:      true,
		},
		{
			name:      "expression without loop var",
			attrValue: "{props.title}",
			loopVars:  map[string]bool{"card": true},
			want:      false,
		},
		{
			name:      "static value",
			attrValue: "static-value",
			loopVars:  map[string]bool{"card": true},
			want:      false,
		},
		{
			name:      "multiple expressions with loop var",
			attrValue: "prefix-{item.id}-{item.name}",
			loopVars:  map[string]bool{"item": true},
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := attributeReferencesLoopVar(tt.attrValue, tt.loopVars)
			if got != tt.want {
				t.Errorf("attributeReferencesLoopVar(%q, %v) = %v, want %v",
					tt.attrValue, tt.loopVars, got, tt.want)
			}
		})
	}
}

// Helper function to remove whitespace for easier comparison
func removeWhitespace(s string) string {
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\t", "")
	return s
}
