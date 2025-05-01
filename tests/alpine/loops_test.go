package alpine

import (
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
	"github.com/jimafisk/custom_go_template/tests/testutils"
	"github.com/jimafisk/custom_go_template/transformer"
)

func TestLoopTransformation(t *testing.T) {
	tests := []struct {
		name      string
		template  *ast.Template
		props     map[string]any
		contains  []string
		notContains []string
	}{
		{
			name: "simple list loop",
			template: &ast.Template{
				RootNodes: []ast.Node{
					&ast.Element{
						TagName: "ul",
						Attributes: []ast.Attribute{
							{
								Name:  "class",
								Value: "items-list",
							},
						},
						Children: []ast.Node{
							&ast.Loop{
								Iterator:   "item",
								Collection: "items",
								Content: []ast.Node{
									&ast.Element{
										TagName: "li",
										Children: []ast.Node{
											&ast.TextNode{Content: "{item}"},
										},
									},
								},
								IsOf: false, // "in" loop
							},
						},
					},
				},
			},
			props: map[string]any{
				"items": []string{"Item 1", "Item 2", "Item 3"},
			},
			contains: []string{
				`<ul class="items-list">`,
				`<template x-for="item in items">`,
				`<li><span x-text="item"></span></li>`,
				`</template>`,
				`</ul>`,
			},
		},
		{
			name: "table with row loop",
			template: &ast.Template{
				RootNodes: []ast.Node{
					&ast.Element{
						TagName: "table",
						Attributes: []ast.Attribute{
							{
								Name:  "class",
								Value: "data-table",
							},
						},
						Children: []ast.Node{
							&ast.Element{
								TagName: "thead",
								Children: []ast.Node{
									&ast.Element{
										TagName: "tr",
										Children: []ast.Node{
											&ast.Element{
												TagName: "th",
												Children: []ast.Node{
													&ast.TextNode{Content: "Name"},
												},
											},
											&ast.Element{
												TagName: "th",
												Children: []ast.Node{
													&ast.TextNode{Content: "Age"},
												},
											},
										},
									},
								},
							},
							&ast.Element{
								TagName: "tbody",
								Children: []ast.Node{
									&ast.Loop{
										Iterator:   "user",
										Collection: "users",
										Content: []ast.Node{
											&ast.Element{
												TagName: "tr",
												Children: []ast.Node{
													&ast.Element{
														TagName: "td",
														Children: []ast.Node{
															&ast.TextNode{Content: "{user.name}"},
														},
													},
													&ast.Element{
														TagName: "td",
														Children: []ast.Node{
															&ast.TextNode{Content: "{user.age}"},
														},
													},
												},
											},
										},
										IsOf: false, // "in" loop
									},
								},
							},
						},
					},
				},
			},
			props: map[string]any{
				"users": []map[string]any{
					{"name": "John", "age": 30},
					{"name": "Jane", "age": 25},
				},
			},
			contains: []string{
				`<table class="data-table">`,
				`<thead><tr><th>Name</th><th>Age</th></tr></thead>`,
				`<tbody>`,
				`<template x-for="user in users">`,
				`<tr><td><span x-text="user.name"></span></td><td><span x-text="user.age"></span></td></tr>`,
				`</template>`,
				`</tbody>`,
				`</table>`,
			},
		},
		{
			name: "loop with index and conditional",
			template: &ast.Template{
				RootNodes: []ast.Node{
					&ast.Element{
						TagName: "div",
						Attributes: []ast.Attribute{
							{
								Name:  "class",
								Value: "task-list",
							},
						},
						Children: []ast.Node{
							&ast.Loop{
								Iterator:   "index",
								Value:      "task",
								Collection: "tasks",
								Content: []ast.Node{
									&ast.Element{
										TagName: "div",
										Attributes: []ast.Attribute{
											{
												Name:  "class",
												Value: "task-item",
											},
										},
										Children: []ast.Node{
											&ast.TextNode{Content: "{index + 1}. "},
											&ast.Conditional{
												IfCondition: "task.completed",
												IfContent: []ast.Node{
													&ast.Element{
														TagName: "span",
														Attributes: []ast.Attribute{
															{
																Name:  "class",
																Value: "completed",
															},
														},
														Children: []ast.Node{
															&ast.TextNode{Content: "{task.title}"},
														},
													},
												},
												ElseContent: []ast.Node{
													&ast.TextNode{Content: "{task.title}"},
												},
											},
										},
									},
								},
								IsOf: false, // "in" loop
							},
						},
					},
				},
			},
			props: map[string]any{
				"tasks": []map[string]any{
					{"title": "Task 1", "completed": true},
					{"title": "Task 2", "completed": false},
				},
				"title": "Custom Template Showcase",
			},
			contains: []string{
				`<div class="task-list">`,
				`<template x-for="(index, task) in tasks">`,
				`<div class="task-item">`,
				`<span x-text="index + 1"></span>. `,
				`<template x-if="task.completed">`,
				`<span class="completed"><span x-text="task.title"></span></span>`,
				`</template>`,
				`<template x-else>`,
				`<span x-text="task.title"></span>`,
				`</template>`,
				`</div>`,
				`</template>`,
				`</div>`,
			},
			notContains: []string{
				`<template x-if="!(task.completed)">`,
			},
		},
		{
			name: "object loop with of",
			template: &ast.Template{
				RootNodes: []ast.Node{
					&ast.Element{
						TagName: "dl",
						Attributes: []ast.Attribute{
							{
								Name:  "class",
								Value: "properties",
							},
						},
						Children: []ast.Node{
							&ast.Loop{
								Iterator:   "key",
								Value:      "value",
								Collection: "product",
								Content: []ast.Node{
									&ast.Element{
										TagName: "dt",
										Children: []ast.Node{
											&ast.TextNode{Content: "{key}"},
										},
									},
									&ast.Element{
										TagName: "dd",
										Children: []ast.Node{
											&ast.TextNode{Content: "{value}"},
										},
									},
								},
								IsOf: true, // "of" loop
							},
						},
					},
				},
			},
			props: map[string]any{
				"product": map[string]any{
					"name":  "Widget",
					"price": 19.99,
					"stock": 42,
				},
			},
			contains: []string{
				`<dl class="properties">`,
				`<template x-for="key, value of Object.entries(product)">`,
				`<dt><span x-text="key"></span></dt>`,
				`<dd><span x-text="value"></span></dd>`,
				`</template>`,
				`</dl>`,
			},
		},
		{
			name: "nested loops",
			template: &ast.Template{
				RootNodes: []ast.Node{
					&ast.Element{
						TagName: "div",
						Attributes: []ast.Attribute{
							{
								Name:  "class",
								Value: "categories",
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
													&ast.TextNode{Content: "{category.name}"},
												},
											},
											&ast.Element{
												TagName: "ul",
												Children: []ast.Node{
													&ast.Loop{
														Iterator:   "item",
														Collection: "category.items",
														Content: []ast.Node{
															&ast.Element{
																TagName: "li",
																Children: []ast.Node{
																	&ast.TextNode{Content: "{item}"},
																},
															},
														},
														IsOf: false, // "in" loop
													},
												},
											},
										},
									},
								},
								IsOf: false, // "in" loop
							},
						},
					},
				},
			},
			props: map[string]any{
				"categories": []map[string]any{
					{
						"name":  "Fruits",
						"items": []string{"Apple", "Banana", "Orange"},
					},
					{
						"name":  "Vegetables",
						"items": []string{"Carrot", "Broccoli"},
					},
				},
			},
			contains: []string{
				`<div class="categories">`,
				`<template x-for="category in categories">`,
				`<div class="category">`,
				`<h3><span x-text="category.name"></span></h3>`,
				`<ul>`,
				`<template x-for="item in category.items">`,
				`<li><span x-text="item"></span></li>`,
				`</template>`,
				`</ul>`,
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
			var result string
			for _, node := range transformed.RootNodes {
				result += testutils.RenderNode(node)
			}
			
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
