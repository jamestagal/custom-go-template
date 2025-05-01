package alpine

import (
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
	"github.com/jimafisk/custom_go_template/tests/testutils"
	"github.com/jimafisk/custom_go_template/transformer"
)

func TestConditionalTransformation(t *testing.T) {
	tests := []struct {
		name      string
		template  *ast.Template
		props     map[string]any
		contains  []string
		notContains []string
	}{
		{
			name: "simple if condition",
			template: &ast.Template{
				RootNodes: []ast.Node{
					&ast.Element{
						TagName: "div",
						Attributes: []ast.Attribute{
							{
								Name:  "class",
								Value: "container",
							},
						},
						Children: []ast.Node{
							&ast.Conditional{
								IfCondition: "isActive",
								IfContent: []ast.Node{
									&ast.TextNode{Content: "Active state"},
								},
							},
						},
					},
				},
			},
			props: map[string]any{
				"isActive": true,
			},
			contains: []string{
				`<div x-data=`,
				`<div class="container">`,
				`<template x-if="isActive">`,
				`Active state`,
				`</template>`,
				`</div>`,
				`</div>`,
			},
		},
		{
			name: "if-else condition",
			template: &ast.Template{
				RootNodes: []ast.Node{
					&ast.Element{
						TagName: "div",
						Attributes: []ast.Attribute{
							{
								Name:  "class",
								Value: "status-indicator",
							},
						},
						Children: []ast.Node{
							&ast.Conditional{
								IfCondition: "status === 'active'",
								IfContent: []ast.Node{
									&ast.Element{
										TagName: "span",
										Attributes: []ast.Attribute{
											{
												Name:  "class",
												Value: "active",
											},
										},
										Children: []ast.Node{
											&ast.TextNode{Content: "Active"},
										},
									},
								},
								ElseContent: []ast.Node{
									&ast.Element{
										TagName: "span",
										Attributes: []ast.Attribute{
											{
												Name:  "class",
												Value: "inactive",
											},
										},
										Children: []ast.Node{
											&ast.TextNode{Content: "Inactive"},
										},
									},
								},
							},
						},
					},
				},
			},
			props: map[string]any{
				"status": "inactive",
			},
			contains: []string{
				`<div x-data=`,
				`<div class="status-indicator">`,
				`<template x-if="status === 'active'">`,
				`<span class="active">Active</span>`,
				`</template>`,
				`<template x-else>`,
				`<span class="inactive">Inactive</span>`,
				`</template>`,
				`</div>`,
				`</div>`,
			},
			notContains: []string{
				`<template x-if="!(status === 'active')">`,
			},
		},
		{
			name: "if-else-if-else condition",
			template: &ast.Template{
				RootNodes: []ast.Node{
					&ast.Element{
						TagName: "div",
						Attributes: []ast.Attribute{
							{
								Name:  "class",
								Value: "status-display",
							},
						},
						Children: []ast.Node{
							&ast.Conditional{
								IfCondition: "status === 'active'",
								IfContent: []ast.Node{
									&ast.Element{
										TagName: "div",
										Attributes: []ast.Attribute{
											{
												Name:  "class",
												Value: "active-status",
											},
										},
										Children: []ast.Node{
											&ast.TextNode{Content: "Status: Active"},
										},
									},
								},
								ElseIfConditions: []string{"status === 'pending'"},
								ElseIfContent: [][]ast.Node{
									{
										&ast.Element{
											TagName: "div",
											Attributes: []ast.Attribute{
												{
													Name:  "class",
													Value: "pending-status",
												},
											},
											Children: []ast.Node{
												&ast.TextNode{Content: "Status: Pending"},
											},
										},
									},
								},
								ElseContent: []ast.Node{
									&ast.Element{
										TagName: "div",
										Attributes: []ast.Attribute{
											{
												Name:  "class",
												Value: "inactive-status",
											},
										},
										Children: []ast.Node{
											&ast.TextNode{Content: "Status: Inactive"},
										},
									},
								},
							},
						},
					},
				},
			},
			props: map[string]any{
				"status": "pending",
			},
			contains: []string{
				`<div x-data=`,
				`<div class="status-display">`,
				`<template x-if="status === 'active'">`,
				`<div class="active-status">Status: Active</div>`,
				`</template>`,
				`<template x-else-if="status === 'pending'">`,
				`<div class="pending-status">Status: Pending</div>`,
				`</template>`,
				`<template x-else>`,
				`<div class="inactive-status">Status: Inactive</div>`,
				`</template>`,
				`</div>`,
				`</div>`,
			},
			notContains: []string{
				`<template x-if="(!(status === 'active')) && (status === 'pending')">`,
				`<template x-if="!(status === 'active') && !(status === 'pending')">`,
			},
		},
		{
			name: "nested conditionals",
			template: &ast.Template{
				RootNodes: []ast.Node{
					&ast.Element{
						TagName: "div",
						Attributes: []ast.Attribute{
							{
								Name:  "class",
								Value: "user-profile",
							},
						},
						Children: []ast.Node{
							&ast.Conditional{
								IfCondition: "isLoggedIn",
								IfContent: []ast.Node{
									&ast.TextNode{Content: "Welcome, {username}!"},
									&ast.Conditional{
										IfCondition: "isAdmin",
										IfContent: []ast.Node{
											&ast.TextNode{Content: "You have admin privileges."},
										},
										ElseContent: []ast.Node{
											&ast.TextNode{Content: "You have user privileges."},
										},
									},
								},
								ElseContent: []ast.Node{
									&ast.TextNode{Content: "Please log in."},
								},
							},
						},
					},
				},
			},
			props: map[string]any{
				"isLoggedIn": true,
				"isAdmin": false,
				"username": "JohnDoe",
			},
			contains: []string{
				`<div x-data=`,
				`<div class="user-profile">`,
				`<template x-if="isLoggedIn">`,
				`Welcome, <span x-text="username"></span>!`,
				`<template x-if="isAdmin">`,
				`You have admin privileges.`,
				`</template>`,
				`<template x-else>`,
				`You have user privileges.`,
				`</template>`,
				`</template>`,
				`<template x-else>`,
				`Please log in.`,
				`</template>`,
				`</div>`,
				`</div>`,
			},
			notContains: []string{
				`<template x-if="!(isAdmin)">`,
				`<template x-if="!(isLoggedIn)">`,
			},
		},
		{
			name: "conditionals with expressions",
			template: &ast.Template{
				RootNodes: []ast.Node{
					&ast.Element{
						TagName: "div",
						Attributes: []ast.Attribute{
							{
								Name:  "class",
								Value: "cart",
							},
						},
						Children: []ast.Node{
							&ast.Conditional{
								IfCondition: "items.length > 0",
								IfContent: []ast.Node{
									&ast.TextNode{Content: "You have {items.length} items in your cart."},
									&ast.TextNode{Content: "Total: ${total}"},
								},
								ElseContent: []ast.Node{
									&ast.TextNode{Content: "Your cart is empty."},
								},
							},
						},
					},
				},
			},
			props: map[string]any{
				"items": []string{"item1", "item2", "item3"},
				"total": 42.99,
			},
			contains: []string{
				`<div x-data=`,
				`<div class="cart">`,
				`<template x-if="items.length > 0">`,
				`You have <span x-text="items.length"></span> items in your cart.`,
				`Total: $<span x-text="total"></span>`,
				`</template>`,
				`<template x-else>`,
				`Your cart is empty.`,
				`</template>`,
				`</div>`,
				`</div>`,
			},
			notContains: []string{
				`<template x-if="!(items.length > 0)">`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Transform the template
			transformed := transformer.TransformAST(tt.template, tt.props)
			
			// Render the transformed template
			var result string
			for _, node := range transformed.RootNodes {
				result += testutils.RenderNode(node)
			}
			
			// Normalize and compare
			resultNormalized := testutils.NormalizeWhitespace(result)
			
			// Check that output contains expected strings
			for _, s := range tt.contains {
				normalizedSubstr := testutils.NormalizeWhitespace(s)
				if !strings.Contains(resultNormalized, normalizedSubstr) {
					t.Errorf("Expected output to contain (normalized):\n%s\n\nBut it doesn't. Full output (normalized):\n%s", normalizedSubstr, resultNormalized)
				}
			}
			
			// Check that output doesn't contain unwanted strings
			for _, s := range tt.notContains {
				normalizedSubstr := testutils.NormalizeWhitespace(s)
				if strings.Contains(resultNormalized, normalizedSubstr) {
					t.Errorf("Expected output NOT to contain (normalized):\n%s\n\nBut it does. Full output (normalized):\n%s", normalizedSubstr, resultNormalized)
				}
			}
		})
	}
}
