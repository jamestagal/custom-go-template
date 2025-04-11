package transformer

import (
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
)

func TestNestedStructures(t *testing.T) {
	tests := []struct {
		name        string
		nodes       []ast.Node
		dataScope   map[string]any
		contains    []string
		notContains []string
	}{
		{
			name: "nested_if_inside_if",
			nodes: []ast.Node{
				&ast.Conditional{
					IfCondition: "outerCondition",
					IfContent: []ast.Node{
						&ast.TextNode{Content: "Outer content"},
						&ast.Conditional{
							IfCondition: "innerCondition",
							IfContent: []ast.Node{
								&ast.TextNode{Content: "Inner content"},
							},
						},
					},
				},
			},
			dataScope: map[string]any{
				"outerCondition": true,
				"innerCondition": true,
			},
			contains: []string{
				"x-if=\"outerCondition\"",
				"Outer content",
				"x-if=\"innerCondition\"",
				"Inner content",
			},
		},
		{
			name: "loop_inside_conditional",
			nodes: []ast.Node{
				&ast.Conditional{
					IfCondition: "showList",
					IfContent: []ast.Node{
						&ast.TextNode{Content: "Items:"},
						&ast.Loop{
							Iterator:   "item",
							Collection: "items",
							Content: []ast.Node{
								&ast.TextNode{Content: "- "},
								&ast.ExpressionNode{Expression: "item"},
							},
						},
					},
				},
			},
			dataScope: map[string]any{
				"showList": true,
				"items":    []string{"Item 1", "Item 2"},
			},
			contains: []string{
				"x-if=\"showList\"",
				"Items:",
				"x-for=\"item in items\"",
				"x-text=\"item\"",
			},
		},
		{
			name: "conditional_inside_loop",
			nodes: []ast.Node{
				&ast.Loop{
					Iterator:   "item",
					Collection: "items",
					Content: []ast.Node{
						&ast.Conditional{
							IfCondition: "item.completed",
							IfContent: []ast.Node{
								&ast.TextNode{Content: "✓ "},
							},
							ElseContent: []ast.Node{
								&ast.TextNode{Content: "✗ "},
							},
						},
						&ast.ExpressionNode{Expression: "item.title"},
					},
				},
			},
			dataScope: map[string]any{
				"items": []map[string]any{
					{"title": "Task 1", "completed": true},
					{"title": "Task 2", "completed": false},
				},
			},
			contains: []string{
				"x-for=\"item in items\"",
				"x-if=\"item.completed\"",
				"✓",
				"x-else",
				"✗",
				"x-text=\"item.title\"",
			},
		},
		{
			name: "nested_loops",
			nodes: []ast.Node{
				&ast.Loop{
					Iterator:   "category",
					Collection: "categories",
					Content: []ast.Node{
						&ast.Element{
							TagName: "h2",
							Children: []ast.Node{
								&ast.ExpressionNode{Expression: "category.name"},
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
												&ast.ExpressionNode{Expression: "item"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			dataScope: map[string]any{
				"categories": []map[string]any{
					{
						"name":  "Fruits",
						"items": []string{"Apple", "Banana"},
					},
					{
						"name":  "Vegetables",
						"items": []string{"Carrot", "Broccoli"},
					},
				},
			},
			contains: []string{
				"x-for=\"category in categories\"",
				"<h2",
				"x-text=\"category.name\"",
				"<ul",
				"x-for=\"item in category.items\"",
				"<li",
				"x-text=\"item\"",
			},
		},
		{
			name: "complex_nesting_with_mixed_structures",
			nodes: []ast.Node{
				&ast.Conditional{
					IfCondition: "hasData",
					IfContent: []ast.Node{
						&ast.Loop{
							Iterator:   "section",
							Collection: "sections",
							Content: []ast.Node{
								&ast.Element{
									TagName: "div",
									Attributes: []ast.Attribute{
										{Name: "class", Value: "section"},
									},
									Children: []ast.Node{
										&ast.Element{
											TagName: "h3",
											Children: []ast.Node{
												&ast.ExpressionNode{Expression: "section.title"},
											},
										},
										&ast.Conditional{
											IfCondition: "section.items.length > 0",
											IfContent: []ast.Node{
												&ast.Loop{
													Iterator:   "item",
													Collection: "section.items",
													Content: []ast.Node{
														&ast.Element{
															TagName: "div",
															Attributes: []ast.Attribute{
																{Name: "class", Value: "item"},
															},
															Children: []ast.Node{
																&ast.ExpressionNode{Expression: "item.name"},
																&ast.Conditional{
																	IfCondition: "item.isSpecial",
																	IfContent: []ast.Node{
																		&ast.TextNode{Content: " (Special)"},
																	},
																},
															},
														},
													},
												},
											},
											ElseContent: []ast.Node{
												&ast.TextNode{Content: "No items available"},
											},
										},
									},
								},
							},
						},
					},
					ElseContent: []ast.Node{
						&ast.TextNode{Content: "No data available"},
					},
				},
			},
			dataScope: map[string]any{
				"hasData": true,
				"sections": []map[string]any{
					{
						"title": "Section 1",
						"items": []map[string]any{
							{"name": "Item 1", "isSpecial": true},
							{"name": "Item 2", "isSpecial": false},
						},
					},
					{
						"title": "Section 2",
						"items": []map[string]any{},
					},
				},
			},
			contains: []string{
				"x-if=\"hasData\"",
				"x-for=\"section in sections\"",
				"class=\"section\"",
				"x-text=\"section.title\"",
				"x-if=\"section.items.length > 0\"",
				"x-for=\"item in section.items\"",
				"class=\"item\"",
				"x-text=\"item.name\"",
				"x-if=\"item.isSpecial\"",
				"(Special)",
				"x-else",
				"No items available",
				"x-else",
				"No data available",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Transform the nodes
			result := transformNodes(tt.nodes, tt.dataScope, false)
			
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
