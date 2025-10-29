package alpine

import (
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
	"github.com/jimafisk/custom_go_template/parser"
	"github.com/jimafisk/custom_go_template/transformer"
)

// TestDynamicComponentParsing tests the parsing of dynamic component syntax (<= syntax)
// This is Jim's innovative feature for dynamic component path selection
func TestDynamicComponentParsing(t *testing.T) {
	tests := []struct {
		name       string
		template   string
		wantPath   string
		wantProps  int
		shouldFail bool
	}{
		{
			name:       "static path with single quotes",
			template:   `<='./Card.html' />`,
			wantPath:   "./Card.html",
			wantProps:  0,
			shouldFail: false,
		},
		{
			name:       "static path with double quotes",
			template:   `<="./Button.html" />`,
			wantPath:   "./Button.html",
			wantProps:  0,
			shouldFail: false,
		},
		{
			name:       "path with variable interpolation",
			template:   `<='./views/{comp}.html' />`,
			wantPath:   "./views/{comp}.html",
			wantProps:  0,
			shouldFail: false,
		},
		{
			name:       "with static props",
			template:   `<='path' title="Test" count="5" />`,
			wantPath:   "path",
			wantProps:  2,
			shouldFail: false,
		},
		{
			name:       "with dynamic props",
			template:   `<='./Card.html' title={pageTitle} count={itemCount} />`,
			wantPath:   "./Card.html",
			wantProps:  2,
			shouldFail: false,
		},
		{
			name:       "with shorthand props",
			template:   `<='./Profile.html' {age} {name} />`,
			wantPath:   "./Profile.html",
			wantProps:  2,
			shouldFail: false,
		},
		{
			name:       "with mixed props",
			template:   `<='./views/{comp}.html' title="Welcome" {age} count={total} />`,
			wantPath:   "./views/{comp}.html",
			wantProps:  3,
			shouldFail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseTemplate(tt.template)

			if tt.shouldFail {
				if err == nil {
					t.Errorf("Expected parsing to fail, but it succeeded")
				}
				return
			}

			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			if len(result.RootNodes) == 0 {
				t.Fatal("Expected at least one root node")
			}

			dynComp, ok := result.RootNodes[0].(*ast.DynamicComponentNode)
			if !ok {
				t.Fatalf("Expected *ast.DynamicComponentNode, got %T", result.RootNodes[0])
			}

			if dynComp.PathExpression != tt.wantPath {
				t.Errorf("PathExpression = %q, want %q", dynComp.PathExpression, tt.wantPath)
			}

			if len(dynComp.Props) != tt.wantProps {
				t.Errorf("Props count = %d, want %d", len(dynComp.Props), tt.wantProps)
			}
		})
	}
}

// TestDynamicComponentTransformation tests the transformation of dynamic components
func TestDynamicComponentTransformation(t *testing.T) {
	// Register test components
	registerDynamicTestComponents()

	tests := []struct {
		name      string
		node      *ast.DynamicComponentNode
		dataScope map[string]any
	}{
		{
			name: "static path - component registered",
			node: &ast.DynamicComponentNode{
				PathExpression: "./Card.html",
				Props:          []ast.ComponentProp{},
			},
			dataScope: map[string]any{},
		},
		{
			name: "static path with props - component registered",
			node: &ast.DynamicComponentNode{
				PathExpression: "./Card.html",
				Props: []ast.ComponentProp{
					{Name: "title", Value: "Hello", IsDynamic: false},
					{Name: "count", Value: "5", IsDynamic: false},
				},
			},
			dataScope: map[string]any{},
		},
		{
			name: "path with variable - variable resolved",
			node: &ast.DynamicComponentNode{
				PathExpression: "./views/{comp}.html",
				Props:          []ast.ComponentProp{},
			},
			dataScope: map[string]any{
				"comp": "Card",
			},
		},
		{
			name: "path with variable - variable not resolved",
			node: &ast.DynamicComponentNode{
				PathExpression: "./views/{comp}.html",
				Props:          []ast.ComponentProp{},
			},
			dataScope: map[string]any{},
		},
		{
			name: "unresolved with props",
			node: &ast.DynamicComponentNode{
				PathExpression: "./views/{comp}.html",
				Props: []ast.ComponentProp{
					{Name: "title", Value: "Test", IsDynamic: false},
					{Name: "count", Value: "10", IsDynamic: false},
				},
			},
			dataScope: map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a template with the dynamic component node
			template := &ast.Template{
				RootNodes: []ast.Node{tt.node},
			}

			// Transform the template - just check it doesn't error
			result := transformer.TransformAST(template, tt.dataScope)

			if len(result.RootNodes) == 0 {
				t.Fatal("Expected at least one root node in result")
			}

			// Success - transformation completed without errors
		})
	}
}


// TestDynamicComponentBuildTimeOptimization tests build-time path resolution
func TestDynamicComponentBuildTimeOptimization(t *testing.T) {
	// Register component with resolved path
	cardTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.TextNode{Content: "Dynamic Card Component"},
		},
	}
	transformer.RegisterComponent("./views/Card.html", cardTemplate, []string{"title"})

	node := &ast.DynamicComponentNode{
		PathExpression: "./views/{comp}.html",
		Props: []ast.ComponentProp{
			{Name: "title", Value: "Welcome", IsDynamic: false},
		},
	}

	// Provide the variable value so path can be resolved at build time
	dataScope := map[string]any{
		"comp": "Card",
	}

	template := &ast.Template{
		RootNodes: []ast.Node{node},
	}

	result := transformer.TransformAST(template, dataScope)

	if len(result.RootNodes) == 0 {
		t.Fatal("Expected at least one root node")
	}

	// Should be resolved to actual component (with x-data)
	element, ok := result.RootNodes[0].(*ast.Element)
	if !ok {
		t.Fatalf("Expected *ast.Element, got %T", result.RootNodes[0])
	}

	hasXData := false
	for _, attr := range element.Attributes {
		if attr.Name == "x-data" {
			hasXData = true
			break
		}
	}

	if !hasXData {
		t.Error("Expected build-time resolved component to have x-data attribute (not placeholder)")
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

// registerDynamicTestComponents registers test component templates for dynamic component tests
func registerDynamicTestComponents() {
	// Card component
	cardTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.TextNode{Content: "Card Component"},
		},
	}
	transformer.RegisterComponent("./Card.html", cardTemplate, []string{"title", "count"})

	// Button component
	buttonTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.TextNode{Content: "Button Component"},
		},
	}
	transformer.RegisterComponent("./Button.html", buttonTemplate, []string{})
}
