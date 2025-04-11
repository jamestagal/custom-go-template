package alpine

import (
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
	"github.com/jimafisk/custom_go_template/transformer"
)

func TestStaticComponentTransformation(t *testing.T) {
	tests := []struct {
		name     string
		input    ast.Node
		props    map[string]any
		expected string
	}{
		{
			name: "basic_component_no_props",
			input: &ast.ComponentNode{
				Name:    "Button",
				Props:   []ast.ComponentProp{},
				Dynamic: false,
			},
			props:    map[string]any{},
			expected: `<div x-component="Button"></div>`,
		},
		{
			name: "component_with_static_props",
			input: &ast.ComponentNode{
				Name: "Card",
				Props: []ast.ComponentProp{
					{
						Name:        "title",
						Value:       "Welcome",
						IsShorthand: false,
						IsDynamic:   false,
					},
					{
						Name:        "subtitle",
						Value:       "Hello World",
						IsShorthand: false,
						IsDynamic:   false,
					},
				},
				Dynamic: false,
			},
			props:    map[string]any{},
			expected: `<div x-component="Card" data-prop-title="Welcome" data-prop-subtitle="Hello World"></div>`,
		},
		{
			name: "component_with_dynamic_props",
			input: &ast.ComponentNode{
				Name: "UserProfile",
				Props: []ast.ComponentProp{
					{
						Name:        "user",
						Value:       "currentUser",
						IsShorthand: false,
						IsDynamic:   true,
					},
					{
						Name:        "showDetails",
						Value:       "isAdmin",
						IsShorthand: false,
						IsDynamic:   true,
					},
				},
				Dynamic: false,
			},
			props: map[string]any{
				"currentUser": map[string]any{
					"name": "John Doe",
					"role": "Admin",
				},
				"isAdmin": true,
			},
			expected: `<div x-component="UserProfile" data-prop-user="currentUser" data-prop-showDetails="isAdmin"></div>`,
		},
		{
			name: "component_with_shorthand_props",
			input: &ast.ComponentNode{
				Name: "ProductCard",
				Props: []ast.ComponentProp{
					{
						Name:        "product",
						Value:       "product",
						IsShorthand: true,
						IsDynamic:   true,
					},
					{
						Name:        "inStock",
						Value:       "inStock",
						IsShorthand: true,
						IsDynamic:   true,
					},
				},
				Dynamic: false,
			},
			props: map[string]any{
				"product": map[string]any{
					"id":    "123",
					"name":  "Laptop",
					"price": 999.99,
				},
				"inStock": true,
			},
			expected: `<div x-component="ProductCard" data-prop-product="product" data-prop-inStock="inStock"></div>`,
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
			renderComponentNode(&sb, componentNode)
			output := sb.String()

			if output != tt.expected {
				t.Errorf("Expected output to be %q, but got %q", tt.expected, output)
			}
		})
	}
}

// Helper function to render a node to string for testing
func renderComponentNode(sb *strings.Builder, node ast.Node) {
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
			renderComponentNode(sb, child)
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
