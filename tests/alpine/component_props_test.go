package alpine

import (
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
	"github.com/jimafisk/custom_go_template/transformer"
)

func TestComponentPropsTransformation(t *testing.T) {
	tests := []struct {
		name     string
		input    ast.Node
		props    map[string]any
		expected string
	}{
		{
			name: "component_with_expression_props",
			input: &ast.ComponentNode{
				Name: "DataCard",
				Props: []ast.ComponentProp{
					{
						Name:        "title",
						Value:       "getTitle()",
						IsShorthand: false,
						IsDynamic:   true,
					},
					{
						Name:        "count",
						Value:       "items.length",
						IsShorthand: false,
						IsDynamic:   true,
					},
				},
				Dynamic: false,
			},
			props: map[string]any{
				"items": []any{"item1", "item2", "item3"},
			},
			expected: `<div x-component="DataCard" data-prop-title="getTitle()" data-prop-count="items.length"></div>`,
		},
		{
			name: "component_with_mixed_props",
			input: &ast.ComponentNode{
				Name: "UserProfile",
				Props: []ast.ComponentProp{
					{
						Name:        "username",
						Value:       "currentUser.name",
						IsShorthand: false,
						IsDynamic:   true,
					},
					{
						Name:        "isAdmin",
						Value:       "true",
						IsShorthand: false,
						IsDynamic:   false,
					},
					{
						Name:        "avatar",
						Value:       "'/images/default.png'",
						IsShorthand: false,
						IsDynamic:   false,
					},
				},
				Dynamic: false,
			},
			props: map[string]any{
				"currentUser": map[string]any{
					"name": "John Doe",
					"id":   "123",
				},
			},
			expected: `<div x-component="UserProfile" data-prop-username="currentUser.name" data-prop-isAdmin="true" data-prop-avatar="'/images/default.png'"></div>`,
		},
		{
			name: "component_with_complex_expression_props",
			input: &ast.ComponentNode{
				Name: "ProductCard",
				Props: []ast.ComponentProp{
					{
						Name:        "price",
						Value:       "formatCurrency(product.price)",
						IsShorthand: false,
						IsDynamic:   true,
					},
					{
						Name:        "discount",
						Value:       "product.price > 100 ? '10%' : '5%'",
						IsShorthand: false,
						IsDynamic:   true,
					},
				},
				Dynamic: false,
			},
			props: map[string]any{
				"product": map[string]any{
					"price": 149.99,
				},
			},
			expected: `<div x-component="ProductCard" data-prop-price="formatCurrency(product.price)" data-prop-discount="product.price > 100 ? '10%' : '5%'"></div>`,
		},
		{
			name: "component_with_shorthand_and_spread_props",
			input: &ast.ComponentNode{
				Name: "Form",
				Props: []ast.ComponentProp{
					{
						Name:        "formData",
						Value:       "formData",
						IsShorthand: true,
						IsDynamic:   true,
					},
					{
						Name:        "errors",
						Value:       "validationErrors",
						IsShorthand: false,
						IsDynamic:   true,
					},
				},
				Dynamic: false,
			},
			props: map[string]any{
				"formData": map[string]any{
					"name":  "",
					"email": "",
				},
				"validationErrors": map[string]any{
					"name": "Name is required",
				},
			},
			expected: `<div x-component="Form" data-prop-formData="formData" data-prop-errors="validationErrors"></div>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a template with the component node as the only root node
			template := &ast.Template{
				RootNodes: []ast.Node{tt.input},
			}
			
			// Transform the template
			result := transformer.TransformAST(template, tt.props)
			
			// Check if we have any root nodes in the result
			if len(result.RootNodes) == 0 {
				t.Fatalf("Expected at least one node in the result, but got none")
			}
			
			// Get the component node, which might be wrapped in an x-data div
			var componentNode ast.Node
			rootNode := result.RootNodes[0]
			
			// Check if the root node is an element with x-data (wrapper)
			if element, ok := rootNode.(*ast.Element); ok {
				hasXData := false
				for _, attr := range element.Attributes {
					if attr.Name == "x-data" {
						hasXData = true
						break
					}
				}
				
				// If it's an x-data wrapper, get the component from its children
				if hasXData && len(element.Children) > 0 {
					componentNode = element.Children[0]
				} else {
					componentNode = rootNode
				}
			} else {
				componentNode = rootNode
			}
			
			// Convert the component node to a string for comparison
			var sb strings.Builder
			renderComponentPropsNode(&sb, componentNode)
			output := sb.String()

			if output != tt.expected {
				t.Errorf("Expected output to be %q, but got %q", tt.expected, output)
			}
		})
	}
}

// Helper function to render a node to string for testing
func renderComponentPropsNode(sb *strings.Builder, node ast.Node) {
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
			renderComponentPropsNode(sb, child)
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
