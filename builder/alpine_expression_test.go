package builder

import (
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
)

// TestComponentRegistryWithXDataExpressions tests the critical bug where x-data
// attributes containing {expression} syntax produce invalid JavaScript
//
// This test reproduces the bug from KNOWN_ISSUE_COMPONENT_REGISTRY_SYNTAX.md
func TestComponentRegistryWithXDataExpressions(t *testing.T) {
	// This component mimics the problematic pattern:
	// <div x-data="{ count: {count}, message: '{message}' }">
	component := ComponentTemplate{
		Name: "ScriptTest",
		AST: &ast.Template{
			RootNodes: []ast.Node{
				&ast.Element{
					TagName: "div",
					Attributes: []ast.Attribute{
						{
							Name:       "x-data",
							Value:      "{ count: {count}, message: '{message}' }",
							IsAlpine:   true,
							AlpineType: "data",
						},
					},
					Children: []ast.Node{
						&ast.Element{
							TagName: "p",
							Children: []ast.Node{
								&ast.TextNode{Content: "Count: "},
								&ast.Element{
									TagName: "span",
									Attributes: []ast.Attribute{
										{Name: "x-text", Value: "count"},
									},
									Children: []ast.Node{},
								},
							},
						},
					},
				},
			},
		},
	}

	result := GenerateComponentRegistry([]ComponentTemplate{component})

	t.Logf("Generated JavaScript:\n%s", result)

	// Should NOT contain ${props. inside x-data (this was the bug)
	if strings.Contains(result, `x-data="${props.`) {
		t.Errorf("Generated registry contains incorrect ${props. substitution in x-data attribute:\n%s", result)
	}

	// Should preserve x-data attribute with expressions intact
	// The attribute value should be preserved as-is
	if !strings.Contains(result, `x-data="{`) {
		t.Errorf("Generated registry missing x-data attribute:\n%s", result)
	}

	// Should have proper opening <div tag
	if !strings.Contains(result, `<div`) {
		t.Errorf("Generated registry missing <div opening tag:\n%s", result)
	}

	// Should NOT have syntax errors like "div x-data=" (missing <)
	if strings.Contains(result, `div x-data=`) && !strings.Contains(result, `<div x-data=`) {
		t.Errorf("Generated registry has malformed div tag (missing <):\n%s", result)
	}
}

// TestXDataWithExpressionNodes tests when the parser creates ExpressionNodes
// within an attribute value (the actual structural problem)
func TestXDataWithExpressionNodes(t *testing.T) {
	// This represents what the PARSER actually creates when it encounters:
	// <div x-data="{ count: {count} }">
	//
	// The parser might create:
	// - Element: div
	// - Attribute: x-data="{ count: "
	// - Child: ExpressionNode{Expression: "count"}
	// - Additional text: " }"
	//
	// This is a structural mismatch!

	component := ComponentTemplate{
		Name: "ParsedWithExpression",
		AST: &ast.Template{
			RootNodes: []ast.Node{
				&ast.Element{
					TagName: "div",
					Attributes: []ast.Attribute{
						{
							Name:  "x-data",
							Value: "{ count: ", // Parser might split here
						},
					},
					Children: []ast.Node{
						&ast.ExpressionNode{Expression: "count"},
						&ast.TextNode{Content: " }"},
					},
				},
			},
		},
	}

	result := GenerateComponentRegistry([]ComponentTemplate{component})

	t.Logf("Generated JavaScript with split expression:\n%s", result)

	// The fix should prevent ${props.count} from appearing in this context
	// Because the expression is conceptually part of the x-data attribute

	// For now, this test documents the structural problem
	// The proper fix requires refactoring how attributes are parsed/stored
}

// TestLiteralContentPreservation verifies that expressions in <style> and <script>
// are NOT converted to ${props.} (this should already work)
func TestLiteralContentPreservation(t *testing.T) {
	component := ComponentTemplate{
		Name: "StyleScript",
		AST: &ast.Template{
			RootNodes: []ast.Node{
				&ast.Element{
					TagName: "style",
					Children: []ast.Node{
						&ast.TextNode{Content: ".class { color: "},
						&ast.ExpressionNode{Expression: "primaryColor"},
						&ast.TextNode{Content: "; }"},
					},
				},
				&ast.Element{
					TagName: "script",
					Children: []ast.Node{
						&ast.TextNode{Content: "const x = "},
						&ast.ExpressionNode{Expression: "value"},
						&ast.TextNode{Content: ";"},
					},
				},
			},
		},
	}

	result := GenerateComponentRegistry([]ComponentTemplate{component})

	t.Logf("Generated JavaScript for style/script:\n%s", result)

	// Style content should NOT have ${props.}
	if strings.Contains(result, "${props.primaryColor}") {
		t.Errorf("Style content should NOT convert expressions to ${props.}:\n%s", result)
	}

	// Script content should NOT have ${props.}
	if strings.Contains(result, "${props.value}") {
		t.Errorf("Script content should NOT convert expressions to ${props.}:\n%s", result)
	}

	// Should contain literal expressions
	if !strings.Contains(result, ".class { color: primaryColor; }") {
		t.Errorf("Style content missing literal expression:\n%s", result)
	}

	if !strings.Contains(result, "const x = value;") {
		t.Errorf("Script content missing literal expression:\n%s", result)
	}
}
