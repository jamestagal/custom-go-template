package alpine

import (
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
	"github.com/jimafisk/custom_go_template/transformer"
)

func TestDynamicComponentTransformation(t *testing.T) {
	tests := []struct {
		name     string
		input    ast.Node
		props    map[string]any
		expected string
	}{
		{
			name: "dynamic_component_variable_name",
			input: &ast.ComponentNode{
				Name:    "componentType",
				Props:   []ast.ComponentProp{},
				Dynamic: true,
			},
			props: map[string]any{
				"componentType": "Button",
			},
			expected: `<div x-component="componentType"></div>`,
		},
		{
			name: "path_based_component",
			input: &ast.ComponentNode{
				Name:    "./components/Card",
				Props:   []ast.ComponentProp{},
				Dynamic: false,
			},
			props:    map[string]any{},
			expected: `<div x-component="./components/Card"></div>`,
		},
		{
			name: "dynamic_path_based_component",
			input: &ast.ComponentNode{
				Name:    "componentPath",
				Props:   []ast.ComponentProp{},
				Dynamic: true,
			},
			props: map[string]any{
				"componentPath": "./components/UserProfile",
			},
			expected: `<div x-component="componentPath"></div>`,
		},
		{
			name: "dynamic_component_with_props",
			input: &ast.ComponentNode{
				Name: "componentType",
				Props: []ast.ComponentProp{
					{
						Name:        "title",
						Value:       "Welcome",
						IsShorthand: false,
						IsDynamic:   false,
					},
					{
						Name:        "user",
						Value:       "currentUser",
						IsShorthand: false,
						IsDynamic:   true,
					},
				},
				Dynamic: true,
			},
			props: map[string]any{
				"componentType": "UserCard",
				"currentUser": map[string]any{
					"name": "John Doe",
				},
			},
			expected: `<div x-component="componentType" data-prop-title="Welcome" data-prop-user="currentUser"></div>`,
		},
		{
			name: "computed_component_path",
			input: &ast.ComponentNode{
				Name:    "getComponentPath()",
				Props:   []ast.ComponentProp{},
				Dynamic: true,
			},
			props:    map[string]any{},
			expected: `<div x-component="getComponentPath()"></div>`,
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
			renderDynamicComponentNode(&sb, componentNode)
			output := sb.String()

			if output != tt.expected {
				t.Errorf("Expected output to be %q, but got %q", tt.expected, output)
			}
		})
	}
}

// Helper function to render a node to string for testing
func renderDynamicComponentNode(sb *strings.Builder, node ast.Node) {
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
			renderDynamicComponentNode(sb, child)
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
