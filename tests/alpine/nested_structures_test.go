package alpine

import (
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
	"github.com/jimafisk/custom_go_template/transformer"
)

func TestNestedStructuresTransformation(t *testing.T) {
	tests := []struct {
		name      string
		template  *ast.Template
		props     map[string]any
		contains  []string
		notContains []string
	}{
		{
			name: "deeply_nested_conditionals",
			template: &ast.Template{
				RootNodes: []ast.Node{
					&ast.Element{
						TagName: "div",
						Attributes: []ast.Attribute{
							{
								Name:  "class",
								Value: "nested-conditionals",
							},
						},
						Children: []ast.Node{
							&ast.Conditional{
								IfCondition: "level1",
								IfContent: []ast.Node{
									&ast.TextNode{Content: "Level 1 True"},
									&ast.Conditional{
										IfCondition: "level2",
										IfContent: []ast.Node{
											&ast.TextNode{Content: "Level 2 True"},
											&ast.Conditional{
												IfCondition: "level3",
												IfContent: []ast.Node{
													&ast.TextNode{Content: "Level 3 True"},
												},
												ElseContent: []ast.Node{
													&ast.TextNode{Content: "Level 3 False"},
												},
											},
										},
										ElseContent: []ast.Node{
											&ast.TextNode{Content: "Level 2 False"},
										},
									},
								},
								ElseContent: []ast.Node{
									&ast.TextNode{Content: "Level 1 False"},
								},
							},
						},
					},
				},
			},
			props: map[string]any{
				"level1": true,
				"level2": true,
				"level3": false,
			},
			contains: []string{
				`<div class="nested-conditionals">`,
				`<template x-if="level1">`,
				`Level 1 True`,
				`<template x-if="level2">`,
				`Level 2 True`,
				`<template x-if="level3">`,
				`Level 3 True`,
				`<template x-else>`,
				`Level 3 False`,
				`</template>`,
				`</template>`,
				`<template x-else>`,
				`Level 2 False`,
				`</template>`,
				`</template>`,
				`<template x-else>`,
				`Level 1 False`,
				`</template>`,
				`</div>`,
			},
		},
		{
			name: "nested_loops_with_conditionals",
			template: &ast.Template{
				RootNodes: []ast.Node{
					&ast.Element{
						TagName: "div",
						Attributes: []ast.Attribute{
							{
								Name:  "class",
								Value: "nested-loops-conditionals",
							},
						},
						Children: []ast.Node{
							&ast.Loop{
								Iterator:   "category",
								Collection: "categories",
								Content: []ast.Node{
									&ast.Element{
										TagName: "div",
										Attributes: []ast.Attribute{
											{
												Name:  "class",
												Value: "category",
											},
										},
										Children: []ast.Node{
											&ast.Element{
												TagName: "h3",
												Children: []ast.Node{
													&ast.ExpressionNode{Expression: "category.name"},
												},
											},
											&ast.Conditional{
												IfCondition: "category.items.length > 0",
												IfContent: []ast.Node{
													&ast.Loop{
														Iterator:   "item",
														Collection: "category.items",
														Content: []ast.Node{
															&ast.Element{
																TagName: "div",
																Attributes: []ast.Attribute{
																	{
																		Name:  "class",
																		Value: "item",
																	},
																},
																Children: []ast.Node{
																	&ast.ExpressionNode{Expression: "item.name"},
																	&ast.Conditional{
																		IfCondition: "item.featured",
																		IfContent: []ast.Node{
																			&ast.Element{
																				TagName: "span",
																				Attributes: []ast.Attribute{
																					{
																						Name:  "class",
																						Value: "featured",
																					},
																				},
																				Children: []ast.Node{
																					&ast.TextNode{Content: "Featured!"},
																				},
																			},
																		},
																	},
																},
															},
														},
													},
												},
												ElseContent: []ast.Node{
													&ast.TextNode{Content: "No items in this category"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			props: map[string]any{
				"categories": []map[string]any{
					{
						"name": "Electronics",
						"items": []map[string]any{
							{"name": "Laptop", "featured": true},
							{"name": "Phone", "featured": false},
						},
					},
					{
						"name": "Books",
						"items": []map[string]any{},
					},
				},
			},
			contains: []string{
				`<div class="nested-loops-conditionals">`,
				`<template x-for="category in categories">`,
				`<div class="category">`,
				`<h3><span x-text="category.name"></span></h3>`,
				`<template x-if="category.items.length > 0">`,
				`<template x-for="item in category.items">`,
				`<div class="item">`,
				`<span x-text="item.name"></span>`,
				`<template x-if="item.featured">`,
				`<span class="featured">Featured!</span>`,
				`</template>`,
				`</div>`,
				`</template>`,
				`</template>`,
				`<template x-else>`,
				`No items in this category`,
				`</template>`,
				`</div>`,
				`</template>`,
				`</div>`,
			},
		},
		{
			name: "conditionals_inside_loops_with_variable_access",
			template: &ast.Template{
				RootNodes: []ast.Node{
					&ast.Element{
						TagName: "ul",
						Attributes: []ast.Attribute{
							{
								Name:  "class",
								Value: "user-list",
							},
						},
						Children: []ast.Node{
							&ast.Loop{
								Iterator:   "index",
								Value:      "user",
								Collection: "users",
								Content: []ast.Node{
									&ast.Element{
										TagName: "li",
										Children: []ast.Node{
											&ast.TextNode{Content: "{index + 1}. "},
											&ast.ExpressionNode{Expression: "user.name"},
											&ast.Conditional{
												IfCondition: "user.isAdmin",
												IfContent: []ast.Node{
													&ast.TextNode{Content: " (Admin)"},
													&ast.Conditional{
														IfCondition: "user.superAdmin",
														IfContent: []ast.Node{
															&ast.TextNode{Content: " - Super Admin"},
														},
													},
												},
												ElseContent: []ast.Node{
													&ast.Conditional{
														IfCondition: "user.verified",
														IfContent: []ast.Node{
															&ast.TextNode{Content: " (Verified)"},
														},
														ElseContent: []ast.Node{
															&ast.TextNode{Content: " (Unverified)"},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			props: map[string]any{
				"users": []map[string]any{
					{"name": "John", "isAdmin": true, "superAdmin": true, "verified": true},
					{"name": "Jane", "isAdmin": true, "superAdmin": false, "verified": true},
					{"name": "Bob", "isAdmin": false, "superAdmin": false, "verified": true},
					{"name": "Alice", "isAdmin": false, "superAdmin": false, "verified": false},
				},
			},
			contains: []string{
				`<ul class="user-list">`,
				`<template x-for="(index, user) in users">`,
				`<li>`,
				`<span x-text="index + 1"></span>. <span x-text="user.name"></span>`,
				`<template x-if="user.isAdmin">`,
				` (Admin)`,
				`<template x-if="user.superAdmin">`,
				` - Super Admin`,
				`</template>`,
				`</template>`,
				`<template x-else>`,
				`<template x-if="user.verified">`,
				` (Verified)`,
				`</template>`,
				`<template x-else>`,
				` (Unverified)`,
				`</template>`,
				`</template>`,
				`</li>`,
				`</template>`,
				`</ul>`,
			},
		},
		{
			name: "nested_loops_with_shared_variable_names",
			template: &ast.Template{
				RootNodes: []ast.Node{
					&ast.Element{
						TagName: "div",
						Children: []ast.Node{
							&ast.Loop{
								Iterator:   "item",
								Collection: "outerItems",
								Content: []ast.Node{
									&ast.Element{
										TagName: "div",
										Children: []ast.Node{
											&ast.TextNode{Content: "Outer: {item.name}"},
											&ast.Element{
												TagName: "div",
												Children: []ast.Node{
													&ast.Loop{
														Iterator:   "item",
														Collection: "item.children",
														Content: []ast.Node{
															&ast.Element{
																TagName: "div",
																Children: []ast.Node{
																	&ast.TextNode{Content: "Inner: {item.name}"},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			props: map[string]any{
				"outerItems": []map[string]any{
					{
						"name": "Parent 1",
						"children": []map[string]any{
							{"name": "Child 1.1"},
							{"name": "Child 1.2"},
						},
					},
					{
						"name": "Parent 2",
						"children": []map[string]any{
							{"name": "Child 2.1"},
						},
					},
				},
			},
			contains: []string{
				`<div>`,
				`<template x-for="item in outerItems">`,
				`<div>`,
				`Outer: <span x-text="item.name"></span>`,
				`<div>`,
				`<template x-for="item in item.children">`,
				`<div>`,
				`Inner: <span x-text="item.name"></span>`,
				`</div>`,
				`</template>`,
				`</div>`,
				`</div>`,
				`</template>`,
				`</div>`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Transform the template
			transformed := transformer.TransformAST(tt.template, tt.props)
			
			// Render the transformed template
			var sb strings.Builder
			for _, node := range transformed.RootNodes {
				renderNestedStructuresNode(&sb, node)
			}
			result := sb.String()
			
			// Check that output contains expected strings
			for _, s := range tt.contains {
				if !strings.Contains(result, s) {
					t.Errorf("Expected output to contain %q, but it doesn't.\nOutput: %s", s, result)
				}
			}
			
			// Check that output doesn't contain unwanted strings
			for _, s := range tt.notContains {
				if strings.Contains(result, s) {
					t.Errorf("Expected output not to contain %q, but it does.\nOutput: %s", s, result)
				}
			}
		})
	}
}

// Helper function to render a node to string for testing
func renderNestedStructuresNode(sb *strings.Builder, node ast.Node) {
	switch n := node.(type) {
	case *ast.Element:
		// Special handling for template elements with Alpine directives
		if n.TagName == "template" {
			// Check for Alpine directives
			var isAlpineDirective bool
			var directiveName, directiveValue string
			
			for _, attr := range n.Attributes {
				if attr.IsAlpine {
					isAlpineDirective = true
					directiveName = attr.Name
					directiveValue = attr.Value
					break
				}
			}
			
			// Handle x-if with negated conditions as x-else
			if isAlpineDirective && directiveName == "x-if" && strings.HasPrefix(directiveValue, "!(") {
				// This is an else condition
				sb.WriteString("<template x-else>")
				
				// Render children
				for _, child := range n.Children {
					renderNestedStructuresNode(sb, child)
				}
				
				sb.WriteString("</template>")
				return
			}
		}
		
		// Default element rendering
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
			renderNestedStructuresNode(sb, child)
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
