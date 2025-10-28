package alpine

import (
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
	"github.com/jimafisk/custom_go_template/transformer"
)

func TestStaticComponentTransformation(t *testing.T) {
	// Register component templates for testing
	registerTestComponents()

	tests := []struct {
		name           string
		input          ast.Node
		props          map[string]any
		expectedProps  []string // Order-independent property checks
		expectedText   string
		allowNoXData   bool // Allow missing x-data for empty scopes
	}{
		{
			name: "basic_component_no_props",
			input: &ast.ComponentNode{
				Name:    "Button",
				Props:   []ast.ComponentProp{},
				Dynamic: false,
			},
			props:          map[string]any{},
			expectedProps:  []string{}, // No props expected
			expectedText:   "Button Component",
			allowNoXData:   true, // Empty scope optimization
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
			props: map[string]any{},
			expectedProps: []string{
				"title: 'Welcome'",
				"subtitle: 'Hello World'",
			},
			expectedText: "Card Component",
			allowNoXData: false,
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
			expectedProps: []string{
				"showDetails:",  // Can be 'true' or boolean
				"user:",         // Object value
			},
			expectedText: "UserProfile Component",
			allowNoXData: false,
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
			expectedProps: []string{
				"inStock:",
				"product:",
			},
			expectedText: "ProductCard Component",
			allowNoXData: false,
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

			// Get the root node
			rootNode := result.RootNodes[0]

			// Convert the root node to a string for comparison
			var sb strings.Builder
			renderComponentNode(&sb, rootNode)
			output := sb.String()

			// Check text content
			if !strings.Contains(output, tt.expectedText) {
				t.Errorf("Expected text content %q, but got %q", tt.expectedText, output)
			}

			// Check x-data presence
			hasXData := strings.Contains(output, "x-data=")
			if len(tt.expectedProps) > 0 && !hasXData {
				t.Errorf("Expected x-data attribute for props, but got: %q", output)
			}
			if len(tt.expectedProps) == 0 && !hasXData && !tt.allowNoXData {
				t.Errorf("Expected x-data=\"{}\" for empty scope, but got: %q", output)
			}

			// Check each expected property exists (order-independent)
			for _, expectedProp := range tt.expectedProps {
				if !strings.Contains(output, expectedProp) {
					t.Errorf("Expected property %q in x-data, but got output: %q", expectedProp, output)
				}
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

// registerTestComponents registers test component templates
func registerTestComponents() {
	// Button component
	buttonTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.TextNode{Content: "Button Component"},
		},
	}
	transformer.RegisterComponent("Button", buttonTemplate, []string{})

	// Card component
	cardTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.TextNode{Content: "Card Component"},
		},
	}
	transformer.RegisterComponent("Card", cardTemplate, []string{"title", "subtitle"})

	// UserProfile component
	userProfileTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.TextNode{Content: "UserProfile Component"},
		},
	}
	transformer.RegisterComponent("UserProfile", userProfileTemplate, []string{"user", "showDetails"})

	// ProductCard component
	productCardTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.TextNode{Content: "ProductCard Component"},
		},
	}
	transformer.RegisterComponent("ProductCard", productCardTemplate, []string{"product", "inStock"})
}
