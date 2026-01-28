package renderer

import (
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
)

// TestRenderDynamicComponentByName tests the fallback renderer for DynamicComponentByNameNode
// This renderer should only be called when transformation fails to resolve the component
func TestRenderDynamicComponentByName(t *testing.T) {
	tests := []struct {
		name      string
		node      *ast.DynamicComponentByNameNode
		wantParts []string // Parts that should be in the output comment
	}{
		{
			name: "basic fallback - name expression only",
			node: &ast.DynamicComponentByNameNode{
				NameExpression: "component.name",
				Props:          []ast.ComponentProp{},
				SpreadProps:    []string{},
				SelfClosing:    true,
			},
			wantParts: []string{
				"<!--",
				"Component:dynamic",
				"not resolved",
				"name={component.name}",
				"-->",
			},
		},
		{
			name: "with spread props",
			node: &ast.DynamicComponentByNameNode{
				NameExpression: "Hero2436",
				Props:          []ast.ComponentProp{},
				SpreadProps:    []string{"component.fields"},
				SelfClosing:    true,
			},
			wantParts: []string{
				"<!--",
				"Component:dynamic",
				"not resolved",
				"name=Hero2436",
				"spreads=[{...component.fields}]",
				"-->",
			},
		},
		{
			name: "with regular props",
			node: &ast.DynamicComponentByNameNode{
				NameExpression: "component.type",
				Props: []ast.ComponentProp{
					{Name: "theme", Value: "dark", IsDynamic: false},
					{Name: "size", Value: "large", IsDynamic: false},
				},
				SpreadProps: []string{},
				SelfClosing: true,
			},
			wantParts: []string{
				"<!--",
				"Component:dynamic",
				"not resolved",
				"name={component.type}",
				`props=[theme="dark"`,
				`size="large"]`,
				"-->",
			},
		},
		{
			name: "with dynamic props",
			node: &ast.DynamicComponentByNameNode{
				NameExpression: "MyComponent",
				Props: []ast.ComponentProp{
					{Name: "title", Value: "title", IsDynamic: true},
					{Name: "count", Value: "count + 1", IsDynamic: true},
				},
				SpreadProps: []string{},
				SelfClosing: true,
			},
			wantParts: []string{
				"<!--",
				"Component:dynamic",
				"not resolved",
				"name=MyComponent",
				`props=[title={title}`,
				`count={count + 1}]`,
				"-->",
			},
		},
		{
			name: "complete - with spreads and props",
			node: &ast.DynamicComponentByNameNode{
				NameExpression: "component.name",
				Props: []ast.ComponentProp{
					{Name: "theme", Value: "dark", IsDynamic: false},
					{Name: "active", Value: "isActive", IsDynamic: true},
				},
				SpreadProps: []string{"component.fields", "baseProps"},
				SelfClosing: true,
			},
			wantParts: []string{
				"<!--",
				"Component:dynamic",
				"not resolved",
				"name={component.name}",
				"spreads=[{...component.fields}, {...baseProps}]",
				`theme="dark"`,
				"active={isActive}",
				"-->",
			},
		},
		{
			name: "multiple spread props",
			node: &ast.DynamicComponentByNameNode{
				NameExpression: "DynamicCard",
				Props:          []ast.ComponentProp{},
				SpreadProps:    []string{"props1", "props2", "props3"},
				SelfClosing:    true,
			},
			wantParts: []string{
				"<!--",
				"Component:dynamic",
				"not resolved",
				"name=DynamicCard",
				"spreads=[{...props1}, {...props2}, {...props3}]",
				"-->",
			},
		},
		{
			name: "empty node (minimal fallback)",
			node: &ast.DynamicComponentByNameNode{
				NameExpression: "",
				Props:          []ast.ComponentProp{},
				SpreadProps:    []string{},
				SelfClosing:    true,
			},
			wantParts: []string{
				"<!--",
				"Component:dynamic",
				"not resolved",
				"-->",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RenderDynamicComponentByName(tt.node)

			// Verify it's an HTML comment
			if !strings.HasPrefix(got, "<!--") {
				t.Errorf("RenderDynamicComponentByName() output should start with <!--, got: %s", got)
			}
			if !strings.HasSuffix(got, "-->") {
				t.Errorf("RenderDynamicComponentByName() output should end with -->, got: %s", got)
			}

			// Verify all expected parts are present
			for _, part := range tt.wantParts {
				if !strings.Contains(got, part) {
					t.Errorf("RenderDynamicComponentByName() missing part %q\nGot: %s", part, got)
				}
			}
		})
	}
}

// TestRenderDynamicComponentByName_Integration tests the integration with the main renderer
func TestRenderDynamicComponentByName_Integration(t *testing.T) {
	// Create a template with an unresolved DynamicComponentByNameNode
	// This should never happen in practice (transformer should resolve it)
	// but we test it for debugging scenarios
	template := &ast.Template{
		RootNodes: []ast.Node{
			&ast.Element{
				TagName: "div",
				Children: []ast.Node{
					&ast.DynamicComponentByNameNode{
						NameExpression: "component.name",
						Props: []ast.ComponentProp{
							{Name: "title", Value: "Hello", IsDynamic: false},
						},
						SpreadProps: []string{"fields"},
						SelfClosing: true,
					},
				},
			},
		},
	}

	// Generate markup
	markup := generateMarkup(template)

	// Should contain the diagnostic comment
	if !strings.Contains(markup, "<!-- Component:dynamic (not resolved)") {
		t.Errorf("Expected diagnostic comment in output, got: %s", markup)
	}

	// Should contain component info
	if !strings.Contains(markup, "name={component.name}") {
		t.Errorf("Expected name expression in output, got: %s", markup)
	}

	// Should contain props info
	if !strings.Contains(markup, `title="Hello"`) {
		t.Errorf("Expected props in output, got: %s", markup)
	}

	// Should contain spread info
	if !strings.Contains(markup, "{...fields}") {
		t.Errorf("Expected spread props in output, got: %s", markup)
	}
}

// TestRenderDynamicComponentByName_EdgeCases tests edge cases
func TestRenderDynamicComponentByName_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		node      *ast.DynamicComponentByNameNode
		wantParts []string
	}{
		{
			name: "complex expression in name",
			node: &ast.DynamicComponentByNameNode{
				NameExpression: "componentMap[item.type] || 'DefaultComponent'",
				Props:          []ast.ComponentProp{},
				SpreadProps:    []string{},
				SelfClosing:    true,
			},
			wantParts: []string{
				"name={componentMap[item.type] || 'DefaultComponent'}",
			},
		},
		{
			name: "special characters in prop values",
			node: &ast.DynamicComponentByNameNode{
				NameExpression: "Card",
				Props: []ast.ComponentProp{
					{Name: "title", Value: "Hello & Goodbye", IsDynamic: false},
					{Name: "data", Value: `{"key": "value"}`, IsDynamic: false},
				},
				SpreadProps: []string{},
				SelfClosing: true,
			},
			wantParts: []string{
				"name=Card",
				`title="Hello & Goodbye"`,
				// In HTML comments, we don't need to escape quotes - more readable
				`data="{"key": "value"}"`,
			},
		},
		{
			name: "very long spread expression",
			node: &ast.DynamicComponentByNameNode{
				NameExpression: "Component",
				Props:          []ast.ComponentProp{},
				SpreadProps: []string{
					"veryLongVariableName.withNestedProperties.andMore.data.fields.props",
				},
				SelfClosing: true,
			},
			wantParts: []string{
				"spreads=[{...veryLongVariableName.withNestedProperties.andMore.data.fields.props}]",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RenderDynamicComponentByName(tt.node)

			for _, part := range tt.wantParts {
				if !strings.Contains(got, part) {
					t.Errorf("RenderDynamicComponentByName() missing part %q\nGot: %s", part, got)
				}
			}
		})
	}
}
