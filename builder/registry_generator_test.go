package builder

import (
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
)


// TestGenerateComponentRegistry_SingleComponent tests converting a single component to JS template
// Cognitive Load: 8 (Table-driven test with simple component)
func TestGenerateComponentRegistry_SingleComponent(t *testing.T) {
	tests := []struct {
		name           string
		component      ComponentTemplate
		expectContains []string // Strings that must appear in output
	}{
		{
			name: "simple component with text and expressions",
			component: ComponentTemplate{
				Name: "Hero2436",
				AST: &ast.Template{
					RootNodes: []ast.Node{
						&ast.Element{
							TagName: "section",
							Attributes: []ast.Attribute{
								{Name: "id", Value: "hero-2436"},
							},
							Children: []ast.Node{
								&ast.Element{
									TagName: "h1",
									Children: []ast.Node{
										&ast.ExpressionNode{Expression: "title"},
									},
								},
								&ast.Element{
									TagName: "p",
									Children: []ast.Node{
										&ast.ExpressionNode{Expression: "description"},
									},
								},
							},
						},
					},
				},
			},
			expectContains: []string{
				"export default {",
				"'Hero2436': (props) =>",
				"${props.title}",
				"${props.description}",
				"<section",
				"id=\"hero-2436\"",
				"<h1>",
				"<p>",
			},
		},
		{
			name: "component with nested elements",
			component: ComponentTemplate{
				Name: "Card",
				AST: &ast.Template{
					RootNodes: []ast.Node{
						&ast.Element{
							TagName: "div",
							Attributes: []ast.Attribute{
								{Name: "class", Value: "card"},
							},
							Children: []ast.Node{
								&ast.Element{
									TagName: "div",
									Attributes: []ast.Attribute{
										{Name: "class", Value: "card-header"},
									},
									Children: []ast.Node{
										&ast.ExpressionNode{Expression: "heading"},
									},
								},
								&ast.Element{
									TagName: "div",
									Attributes: []ast.Attribute{
										{Name: "class", Value: "card-body"},
									},
									Children: []ast.Node{
										&ast.TextNode{Content: "Static content"},
									},
								},
							},
						},
					},
				},
			},
			expectContains: []string{
				"'Card': (props) =>",
				"${props.heading}",
				"class=\"card\"",
				"class=\"card-header\"",
				"class=\"card-body\"",
				"Static content",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateComponentRegistry([]ComponentTemplate{tt.component})

			for _, expected := range tt.expectContains {
				if !strings.Contains(result, expected) {
					t.Errorf("GenerateComponentRegistry() output missing expected string:\nwant: %q\ngot: %s", expected, result)
				}
			}

			// Verify ES module structure
			if !strings.HasPrefix(result, "export default {") {
				t.Errorf("GenerateComponentRegistry() should start with 'export default {', got: %s", result[:50])
			}

			if !strings.HasSuffix(strings.TrimSpace(result), "};") {
				t.Errorf("GenerateComponentRegistry() should end with '};'")
			}
		})
	}
}

// TestGenerateComponentRegistry_MultipleComponents tests registry with multiple components
// Cognitive Load: 6 (Simple iteration verification)
func TestGenerateComponentRegistry_MultipleComponents(t *testing.T) {
	components := []ComponentTemplate{
		{
			Name: "Hero2436",
			AST: &ast.Template{
				RootNodes: []ast.Node{
					&ast.Element{
						TagName: "h1",
						Children: []ast.Node{
							&ast.ExpressionNode{Expression: "title"},
						},
					},
				},
			},
		},
		{
			Name: "Services2437",
			AST: &ast.Template{
				RootNodes: []ast.Node{
					&ast.Element{
						TagName: "h2",
						Children: []ast.Node{
							&ast.ExpressionNode{Expression: "heading"},
						},
					},
				},
			},
		},
		{
			Name: "Footer",
			AST: &ast.Template{
				RootNodes: []ast.Node{
					&ast.Element{
						TagName: "footer",
						Children: []ast.Node{
							&ast.TextNode{Content: "Copyright 2025"},
						},
					},
				},
			},
		},
	}

	result := GenerateComponentRegistry(components)

	// All components should be in registry
	expectedComponents := []string{
		"'Hero2436': (props) =>",
		"'Services2437': (props) =>",
		"'Footer': (props) =>",
	}

	for _, expected := range expectedComponents {
		if !strings.Contains(result, expected) {
			t.Errorf("GenerateComponentRegistry() missing component definition: %q", expected)
		}
	}

	// All expressions should be converted
	expectedExpressions := []string{
		"${props.title}",
		"${props.heading}",
		"Copyright 2025",
	}

	for _, expected := range expectedExpressions {
		if !strings.Contains(result, expected) {
			t.Errorf("GenerateComponentRegistry() missing expected content: %q", expected)
		}
	}
}

// TestGenerateComponentRegistry_ExpressionConversion tests {var} -> ${props.var} transformation
// Cognitive Load: 7 (Multiple expression types)
func TestGenerateComponentRegistry_ExpressionConversion(t *testing.T) {
	tests := []struct {
		name          string
		expression    string
		expectedInJS  string
		unexpectedInJS string
	}{
		{
			name:          "simple variable",
			expression:    "title",
			expectedInJS:  "${props.title}",
			unexpectedInJS: "{title}",
		},
		{
			name:          "nested property",
			expression:    "user.name",
			expectedInJS:  "${props.user.name}",
			unexpectedInJS: "{user.name}",
		},
		{
			name:          "array index",
			expression:    "items[0]",
			expectedInJS:  "${props.items[0]}",
			unexpectedInJS: "{items[0]}",
		},
		{
			name:          "method call",
			expression:    "getName()",
			expectedInJS:  "${props.getName()}",
			unexpectedInJS: "{getName()}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component := ComponentTemplate{
				Name: "TestComp",
				AST: &ast.Template{
					RootNodes: []ast.Node{
						&ast.Element{
							TagName: "div",
							Children: []ast.Node{
								&ast.ExpressionNode{Expression: tt.expression},
							},
						},
					},
				},
			}

			result := GenerateComponentRegistry([]ComponentTemplate{component})

			if !strings.Contains(result, tt.expectedInJS) {
				t.Errorf("GenerateComponentRegistry() should convert {%s} to %s\ngot: %s",
					tt.expression, tt.expectedInJS, result)
			}

			if strings.Contains(result, tt.unexpectedInJS) {
				t.Errorf("GenerateComponentRegistry() should NOT contain %s (unconverted expression)",
					tt.unexpectedInJS)
			}
		})
	}
}

// TestGenerateComponentRegistry_AlpineDirectives tests preservation of Alpine.js directives
// Cognitive Load: 9 (Multiple directive types and contexts)
func TestGenerateComponentRegistry_AlpineDirectives(t *testing.T) {
	tests := []struct {
		name           string
		attributes     []ast.Attribute
		expectContains []string
	}{
		{
			name: "x-text directive",
			attributes: []ast.Attribute{
				{Name: "x-text", Value: "message", IsAlpine: true, AlpineType: "text"},
			},
			expectContains: []string{
				"x-text=\"message\"",
			},
		},
		{
			name: "x-bind directive",
			attributes: []ast.Attribute{
				{
					Name:       "x-bind:class",
					Value:      "isActive ? 'active' : ''",
					IsAlpine:   true,
					AlpineType: "bind",
					AlpineKey:  "class",
				},
			},
			expectContains: []string{
				"x-bind:class=\"isActive ? 'active' : ''\"",
			},
		},
		{
			name: "x-if directive on template",
			attributes: []ast.Attribute{
				{Name: "x-if", Value: "isVisible", IsAlpine: true, AlpineType: "if"},
			},
			expectContains: []string{
				"x-if=\"isVisible\"",
			},
		},
		{
			name: "event handler",
			attributes: []ast.Attribute{
				{Name: "@click", Value: "handleClick()", IsAlpine: true, AlpineType: "on"},
			},
			expectContains: []string{
				"@click=\"handleClick()\"",
			},
		},
		{
			name: "x-data directive",
			attributes: []ast.Attribute{
				{Name: "x-data", Value: "{count: 0}", IsAlpine: true, AlpineType: "data"},
			},
			expectContains: []string{
				"x-data=\"{count: 0}\"",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component := ComponentTemplate{
				Name: "AlpineComp",
				AST: &ast.Template{
					RootNodes: []ast.Node{
						&ast.Element{
							TagName:    "div",
							Attributes: tt.attributes,
							Children: []ast.Node{
								&ast.TextNode{Content: "Content"},
							},
						},
					},
				},
			}

			result := GenerateComponentRegistry([]ComponentTemplate{component})

			for _, expected := range tt.expectContains {
				if !strings.Contains(result, expected) {
					t.Errorf("GenerateComponentRegistry() should preserve Alpine directive:\nwant: %q\ngot: %s",
						expected, result)
				}
			}
		})
	}
}

// TestGenerateComponentRegistry_TemplateLiteralEscaping tests proper escaping for JS template literals
// Cognitive Load: 8 (Multiple escaping scenarios)
func TestGenerateComponentRegistry_TemplateLiteralEscaping(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		shouldEscape   bool
		expectContains string
	}{
		{
			name:           "backticks should be escaped",
			content:        "Use backticks `like this` in code",
			shouldEscape:   true,
			expectContains: "\\`like this\\`",
		},
		{
			name:           "dollar brace should be escaped",
			content:        "Template literal: ${variable}",
			shouldEscape:   true,
			expectContains: "\\${variable}",
		},
		{
			name:           "backslash should be escaped",
			content:        "Path: C:\\Users\\Name",
			shouldEscape:   true,
			expectContains: "C:\\\\Users\\\\Name",
		},
		{
			name:           "normal quotes preserved",
			content:        `She said "hello" and I said 'hi'`,
			shouldEscape:   false,
			expectContains: `She said "hello" and I said 'hi'`,
		},
		{
			name:           "newlines preserved",
			content:        "Line 1\nLine 2",
			shouldEscape:   false,
			expectContains: "Line 1\nLine 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component := ComponentTemplate{
				Name: "EscapeTest",
				AST: &ast.Template{
					RootNodes: []ast.Node{
						&ast.Element{
							TagName: "div",
							Children: []ast.Node{
								&ast.TextNode{Content: tt.content},
							},
						},
					},
				},
			}

			result := GenerateComponentRegistry([]ComponentTemplate{component})

			if !strings.Contains(result, tt.expectContains) {
				t.Errorf("GenerateComponentRegistry() escaping issue:\nwant: %q\ngot: %s",
					tt.expectContains, result)
			}
		})
	}
}

// TestGenerateComponentRegistry_MixedContent tests complex components with mixed content types
// Cognitive Load: 10 (Complex structure with multiple node types)
func TestGenerateComponentRegistry_MixedContent(t *testing.T) {
	component := ComponentTemplate{
		Name: "MixedComponent",
		AST: &ast.Template{
			RootNodes: []ast.Node{
				&ast.Element{
					TagName: "div",
					Attributes: []ast.Attribute{
						{Name: "class", Value: "container"},
						{Name: "x-data", Value: "{open: false}", IsAlpine: true},
					},
					Children: []ast.Node{
						&ast.TextNode{Content: "Static text before"},
						&ast.Element{
							TagName: "h2",
							Children: []ast.Node{
								&ast.ExpressionNode{Expression: "title"},
							},
						},
						&ast.TextNode{Content: " - "},
						&ast.Element{
							TagName: "span",
							Attributes: []ast.Attribute{
								{Name: "x-text", Value: "subtitle", IsAlpine: true},
							},
							Children: []ast.Node{},
						},
						&ast.TextNode{Content: "Static text after"},
						&ast.Element{
							TagName: "button",
							Attributes: []ast.Attribute{
								{Name: "@click", Value: "open = !open", IsAlpine: true},
							},
							Children: []ast.Node{
								&ast.TextNode{Content: "Toggle"},
							},
						},
					},
				},
			},
		},
	}

	result := GenerateComponentRegistry([]ComponentTemplate{component})

	expectedElements := []string{
		// Structure
		"export default {",
		"'MixedComponent': (props) =>",
		"<div",
		"class=\"container\"",
		"x-data=\"{open: false}\"",

		// Content
		"Static text before",
		"<h2>",
		"${props.title}",
		" - ",
		"<span",
		"x-text=\"subtitle\"",
		"Static text after",
		"<button",
		"@click=\"open = !open\"",
		"Toggle",
		"</button>",
		"</div>",
	}

	for _, expected := range expectedElements {
		if !strings.Contains(result, expected) {
			t.Errorf("GenerateComponentRegistry() missing expected element:\nwant: %q\ngot: %s",
				expected, result)
		}
	}
}

// TestGenerateComponentRegistry_EmptyComponent tests edge case with empty component
// Cognitive Load: 4 (Simple edge case)
func TestGenerateComponentRegistry_EmptyComponent(t *testing.T) {
	component := ComponentTemplate{
		Name: "Empty",
		AST: &ast.Template{
			RootNodes: []ast.Node{},
		},
	}

	result := GenerateComponentRegistry([]ComponentTemplate{component})

	// Should still generate valid structure
	if !strings.Contains(result, "'Empty': (props) =>") {
		t.Errorf("GenerateComponentRegistry() should handle empty component")
	}

	// Output should be valid JS
	if !strings.HasPrefix(result, "export default {") {
		t.Errorf("GenerateComponentRegistry() should maintain valid ES module structure")
	}
}

// TestGenerateComponentRegistry_SelfClosingElements tests handling of self-closing tags
// Cognitive Load: 6 (Void element handling)
func TestGenerateComponentRegistry_SelfClosingElements(t *testing.T) {
	component := ComponentTemplate{
		Name: "ImageCard",
		AST: &ast.Template{
			RootNodes: []ast.Node{
				&ast.Element{
					TagName: "div",
					Children: []ast.Node{
						&ast.Element{
							TagName: "img",
							Attributes: []ast.Attribute{
								{Name: "src", Value: "/image.jpg"},
								{Name: "alt", Value: "Description"},
							},
							SelfClosing: true,
						},
						&ast.Element{
							TagName: "br",
							SelfClosing: true,
						},
						&ast.Element{
							TagName: "input",
							Attributes: []ast.Attribute{
								{Name: "type", Value: "text"},
							},
							SelfClosing: true,
						},
					},
				},
			},
		},
	}

	result := GenerateComponentRegistry([]ComponentTemplate{component})

	expectedElements := []string{
		"<img",
		"src=\"/image.jpg\"",
		"alt=\"Description\"",
		"<br",
		"<input",
		"type=\"text\"",
	}

	for _, expected := range expectedElements {
		if !strings.Contains(result, expected) {
			t.Errorf("GenerateComponentRegistry() missing self-closing element:\nwant: %q",
				expected)
		}
	}
}

// TestGenerateComponentRegistry_AttributeQuoting tests proper attribute value quoting
// Cognitive Load: 7 (Quote style handling)
func TestGenerateComponentRegistry_AttributeQuoting(t *testing.T) {
	tests := []struct {
		name           string
		attributes     []ast.Attribute
		expectContains string
	}{
		{
			name: "double quotes for simple value",
			attributes: []ast.Attribute{
				{Name: "class", Value: "btn btn-primary"},
			},
			expectContains: `class="btn btn-primary"`,
		},
		{
			name: "value with single quotes",
			attributes: []ast.Attribute{
				{Name: "data-message", Value: "It's working"},
			},
			expectContains: `data-message="It's working"`,
		},
		{
			name: "value with double quotes inside",
			attributes: []ast.Attribute{
				{Name: "title", Value: `Say "hello"`},
			},
			expectContains: `title="Say \"hello\""`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component := ComponentTemplate{
				Name: "QuoteTest",
				AST: &ast.Template{
					RootNodes: []ast.Node{
						&ast.Element{
							TagName:    "div",
							Attributes: tt.attributes,
						},
					},
				},
			}

			result := GenerateComponentRegistry([]ComponentTemplate{component})

			if !strings.Contains(result, tt.expectContains) {
				t.Errorf("GenerateComponentRegistry() attribute quoting issue:\nwant: %q\ngot: %s",
					tt.expectContains, result)
			}
		})
	}
}

// TestGenerateComponentRegistry_ValidJSOutput tests that output is valid JavaScript
// Cognitive Load: 5 (Syntax validation)
func TestGenerateComponentRegistry_ValidJSOutput(t *testing.T) {
	components := []ComponentTemplate{
		{
			Name: "TestComp1",
			AST: &ast.Template{
				RootNodes: []ast.Node{
					&ast.Element{
						TagName: "div",
						Children: []ast.Node{
							&ast.TextNode{Content: "Test"},
						},
					},
				},
			},
		},
		{
			Name: "TestComp2",
			AST: &ast.Template{
				RootNodes: []ast.Node{
					&ast.Element{
						TagName: "span",
						Children: []ast.Node{
							&ast.ExpressionNode{Expression: "value"},
						},
					},
				},
			},
		},
	}

	result := GenerateComponentRegistry(components)

	// Check ES module structure
	if !strings.HasPrefix(result, "export default {") {
		t.Errorf("GenerateComponentRegistry() should start with ES module export")
	}

	// Check proper closing
	if !strings.HasSuffix(strings.TrimSpace(result), "};") {
		t.Errorf("GenerateComponentRegistry() should end with '};'")
	}

	// Check arrow function syntax
	if !strings.Contains(result, "(props) =>") {
		t.Errorf("GenerateComponentRegistry() should use arrow function syntax")
	}

	// Check no syntax errors (basic checks)
	openBraces := strings.Count(result, "{")
	closeBraces := strings.Count(result, "}")
	if openBraces != closeBraces {
		t.Errorf("GenerateComponentRegistry() has mismatched braces: %d open, %d close",
			openBraces, closeBraces)
	}

	// Check proper comma separation between components
	componentCount := len(components)
	// Should have at least componentCount-1 commas for separation
	// (last component should not have trailing comma for cleaner output)
	commaCount := strings.Count(result, "',\n") + strings.Count(result, "'\n")
	if commaCount < componentCount-1 {
		t.Logf("Warning: Expected at least %d component separators, found %d", componentCount-1, commaCount)
	}
}

// TestGenerateComponentRegistry_LiteralContentElements tests that expressions inside <style> and <script>
// are NOT converted to template literals (CRITICAL BUG FIX)
// Cognitive Load: 9 (Testing literal context tracking)
func TestGenerateComponentRegistry_LiteralContentElements(t *testing.T) {
	tests := []struct {
		name              string
		component         ComponentTemplate
		expectContains    []string
		mustNotContain    []string
	}{
		{
			name: "style tag with expression should NOT convert to props",
			component: ComponentTemplate{
				Name: "StyledComponent",
				AST: &ast.Template{
					RootNodes: []ast.Node{
						&ast.Element{
							TagName: "style",
							Children: []ast.Node{
								&ast.TextNode{Content: "body "},
								&ast.ExpressionNode{Expression: "color"},
								&ast.TextNode{Content: " {\n  font-family: system-ui;\n}"},
							},
						},
					},
				},
			},
			expectContains: []string{
				"<style>",
				"body color",
				"font-family: system-ui",
				"</style>",
			},
			mustNotContain: []string{
				"${props.color}",  // Should NOT convert expression in style
			},
		},
		{
			name: "script tag with expression should NOT convert to props",
			component: ComponentTemplate{
				Name: "ScriptComponent",
				AST: &ast.Template{
					RootNodes: []ast.Node{
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
			},
			expectContains: []string{
				"<script>",
				"const x = value;",
				"</script>",
			},
			mustNotContain: []string{
				"${props.value}",  // Should NOT convert expression in script
			},
		},
		{
			name: "normal div with expression SHOULD convert to props",
			component: ComponentTemplate{
				Name: "NormalComponent",
				AST: &ast.Template{
					RootNodes: []ast.Node{
						&ast.Element{
							TagName: "div",
							Children: []ast.Node{
								&ast.ExpressionNode{Expression: "title"},
							},
						},
					},
				},
			},
			expectContains: []string{
				"<div>",
				"${props.title}",  // SHOULD convert in normal elements
				"</div>",
			},
			mustNotContain: []string{},
		},
		{
			name: "nested elements inside style should NOT convert",
			component: ComponentTemplate{
				Name: "NestedStyleComponent",
				AST: &ast.Template{
					RootNodes: []ast.Node{
						&ast.Element{
							TagName: "style",
							Children: []ast.Node{
								&ast.TextNode{Content: ".class {\n  color: "},
								&ast.ExpressionNode{Expression: "primaryColor"},
								&ast.TextNode{Content: ";\n}"},
							},
						},
					},
				},
			},
			expectContains: []string{
				"<style>",
				".class {",
				"color: primaryColor;",
				"</style>",
			},
			mustNotContain: []string{
				"${props.primaryColor}",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateComponentRegistry([]ComponentTemplate{tt.component})

			// Check expected content
			for _, expected := range tt.expectContains {
				if !strings.Contains(result, expected) {
					t.Errorf("GenerateComponentRegistry() missing expected content:\nwant: %q\ngot: %s",
						expected, result)
				}
			}

			// Check prohibited content
			for _, prohibited := range tt.mustNotContain {
				if strings.Contains(result, prohibited) {
					t.Errorf("GenerateComponentRegistry() contains prohibited content:\nshould NOT have: %q\ngot: %s",
						prohibited, result)
				}
			}
		})
	}
}

// TestGenerateComponentRegistry_MixedLiteralAndNormalContent tests components with both literal and normal content
// Cognitive Load: 10 (Complex mixed context)
func TestGenerateComponentRegistry_MixedLiteralAndNormalContent(t *testing.T) {
	component := ComponentTemplate{
		Name: "MixedLiteralComponent",
		AST: &ast.Template{
			RootNodes: []ast.Node{
				&ast.Element{
					TagName: "div",
					Children: []ast.Node{
						&ast.Element{
							TagName: "h1",
							Children: []ast.Node{
								&ast.ExpressionNode{Expression: "title"},  // Should convert
							},
						},
						&ast.Element{
							TagName: "style",
							Children: []ast.Node{
								&ast.TextNode{Content: ".heading { color: "},
								&ast.ExpressionNode{Expression: "blue"},  // Should NOT convert
								&ast.TextNode{Content: "; }"},
							},
						},
						&ast.Element{
							TagName: "p",
							Children: []ast.Node{
								&ast.ExpressionNode{Expression: "description"},  // Should convert
							},
						},
					},
				},
			},
		},
	}

	result := GenerateComponentRegistry([]ComponentTemplate{component})

	// Expressions in normal elements should be converted
	expectedConversions := []string{
		"${props.title}",
		"${props.description}",
	}

	for _, expected := range expectedConversions {
		if !strings.Contains(result, expected) {
			t.Errorf("GenerateComponentRegistry() should convert normal element expressions:\nwant: %q\ngot: %s",
				expected, result)
		}
	}

	// Expressions in style should NOT be converted
	if strings.Contains(result, "${props.blue}") {
		t.Errorf("GenerateComponentRegistry() should NOT convert style expressions:\ngot: %s", result)
	}

	// Style content should contain literal expression
	if !strings.Contains(result, ".heading { color: blue; }") {
		t.Errorf("GenerateComponentRegistry() style content incorrect:\ngot: %s", result)
	}
}

// TestConvertAttributeExpressions tests the conversion of {expression} patterns in attribute values
// Cognitive Load: 8 (Table-driven test with multiple cases)
// CRITICAL FIX TEST: This test verifies the fix for attribute expression conversion
func TestConvertAttributeExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple expression",
			input:    "{count}",
			expected: "${props.count}",
		},
		{
			name:     "x-data with single expression",
			input:    "{ count: {count} }",
			expected: "{ count: ${props.count} }",
		},
		{
			name:     "x-data with multiple expressions",
			input:    "{ count: {count}, message: '{message}' }",
			expected: "{ count: ${props.count}, message: '${props.message}' }",
		},
		{
			name:     "href with expression",
			input:    "{buttonLink}",
			expected: "${props.buttonLink}",
		},
		{
			name:     "no expressions",
			input:    "{ count: 0, message: 'hello' }",
			expected: "{ count: 0, message: 'hello' }",
		},
		{
			name:     "nested braces in strings",
			input:    "{ message: 'Value: {value}' }",
			expected: "{ message: 'Value: ${props.value}' }",
		},
		{
			name:     "multiple expressions in one value",
			input:    "{firstName} {lastName}",
			expected: "${props.firstName} ${props.lastName}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertAttributeExpressions(tt.input, nil)
			if result != tt.expected {
				t.Errorf("convertAttributeExpressions() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestConvertAttributeExpressions_ComplexExpressions tests identifier-level prefixing fix
// Cognitive Load: 10 (Complex expression patterns from IMPLEMENTATION_PLAN.md)
func TestConvertAttributeExpressions_ComplexExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Parenthesized expression",
			input:    "{(start * 1) + index + 1}",
			expected: "${(props.start * 1) + index + 1}", // index skipped (loop var)
		},
		{
			name:     "Multiple operators",
			input:    "{count + total - discount}",
			expected: "${props.count + props.total - props.discount}",
		},
		{
			name:     "Alpine store access",
			input:    "{$store.cart.count}",
			expected: "${$store.cart.count}", // $store skipped (Alpine built-in)
		},
		{
			name:     "Mixed loop var and prop",
			input:    "{start + index}",
			expected: "${props.start + index}", // index skipped
		},
		{
			name:     "Already prefixed",
			input:    "{props.count}",
			expected: "${props.count}", // Don't double-prefix
		},
		{
			name:     "Alpine object literal with expressions",
			input:    "{ count: {count}, message: '{message}' }",
			expected: "{ count: ${props.count}, message: '${props.message}' }",
		},
		{
			name:     "Alpine object literal plain",
			input:    "{ count: 0, items: [] }",
			expected: "{ count: 0, items: [] }", // No expressions to convert
		},
		{
			name:     "Arrow function with two parameters",
			input:    "{products.reduce((sum, p) => sum + p.price, 0)}",
			expected: "${props.products.reduce((sum, p) => sum + p.price, 0)}", // sum and p NOT prefixed
		},
		{
			name:     "Arrow function with single parameter",
			input:    "{items.filter(x => x > 10)}",
			expected: "${props.items.filter(x => x > 10)}", // x NOT prefixed
		},
		{
			name:     "Arrow function using props inside",
			input:    "{products.map(p => p.price * multiplier)}",
			expected: "${props.products.map(p => p.price * props.multiplier)}", // IMPROVED: Method call arguments ARE now prefixed (multiplier is a prop, not arrow param)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertAttributeExpressions(tt.input, nil)
			if result != tt.expected {
				t.Errorf("convertAttributeExpressions() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestSkipIdentifiers tests that skip list identifiers are not prefixed
// Cognitive Load: 6 (Skip list validation)
func TestSkipIdentifiers(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"{index}", "${index}"},                    // Loop var
		{"{item.name}", "${item.name}"},            // Loop var with property
		{"{$store.cart}", "${$store.cart}"},        // Alpine built-in
		{"{Math.floor(x)}", "${Math.floor(props.x)}"}, // IMPROVED: Arguments inside method calls ARE now prefixed (x is a prop)
		{"{window.location}", "${window.location}"}, // JS built-in
	}

	for _, tt := range tests {
		result := convertAttributeExpressions(tt.input, nil)
		if result != tt.expected {
			t.Errorf("convertAttributeExpressions(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

// TestRenderAttributeToJS tests attribute rendering with expression conversion
// Cognitive Load: 12 (Integration test with string building)
// CRITICAL FIX TEST: Verifies the complete attribute rendering pipeline
func TestRenderAttributeToJS(t *testing.T) {
	tests := []struct {
		name     string
		attr     ast.Attribute
		expected string
	}{
		{
			name: "x-data with expressions",
			attr: ast.Attribute{
				Name:  "x-data",
				Value: "{ count: {count}, message: '{message}' }",
			},
			expected: `x-data="{ count: ${props.count}, message: '${props.message}' }"`,
		},
		{
			name: "href with expression",
			attr: ast.Attribute{
				Name:  "href",
				Value: "{buttonLink}",
			},
			expected: `href="${props.buttonLink}"`,
		},
		{
			name: "class with no expressions",
			attr: ast.Attribute{
				Name:  "class",
				Value: "btn btn-primary",
			},
			expected: `class="btn btn-primary"`,
		},
		{
			name: "x-bind with expression",
			attr: ast.Attribute{
				Name:  "x-bind:disabled",
				Value: "{isDisabled}",
			},
			expected: `x-bind:disabled="${props.isDisabled}"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sb strings.Builder
			ctx := &RenderContext{}
			renderAttributeToJS(tt.attr, &sb, ctx, nil)
			result := sb.String()

			if result != tt.expected {
				t.Errorf("renderAttributeToJS() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestGenerateComponentRegistry_WithExpressions tests the full registry generation with attribute expressions
// Cognitive Load: 15 (Integration test with complex AST)
// CRITICAL FIX TEST: This is the end-to-end test that would have caught the original bug
func TestGenerateComponentRegistry_WithExpressions(t *testing.T) {
	// Create a component with x-data attribute containing expressions
	component := ComponentTemplate{
		Name: "TestComponent",
		AST: &ast.Template{
			RootNodes: []ast.Node{
				&ast.Element{
					TagName: "div",
					Attributes: []ast.Attribute{
						{
							Name:  "x-data",
							Value: "{ count: {count}, message: '{message}' }",
						},
					},
					Children: []ast.Node{
						&ast.TextNode{Content: "Test content"},
					},
				},
			},
		},
	}

	result := GenerateComponentRegistry([]ComponentTemplate{component})

	// Verify the generated registry
	expectedSubstring := `x-data="{ count: ${props.count}, message: '${props.message}' }"`
	if !strings.Contains(result, expectedSubstring) {
		t.Errorf("Generated registry does not contain expected x-data attribute conversion\nGot: %s", result)
	}

	// Verify no invalid ${props. syntax appears
	if strings.Contains(result, "${props. ") {
		t.Errorf("Generated registry contains invalid ${props. syntax with space\nGot: %s", result)
	}

	// Verify the component function is properly formatted
	expectedPrefix := `export default {
  'TestComponent': (props) => `
	if !strings.HasPrefix(result, expectedPrefix) {
		t.Errorf("Generated registry does not have expected prefix\nGot: %s", result)
	}
}

// TestEscapeTemplateLiteral tests template literal escaping
// Cognitive Load: 6 (Table-driven test)
func TestEscapeTemplateLiteral(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "backtick",
			input:    "Hello `world`",
			expected: "Hello \\`world\\`",
		},
		{
			name:     "template expression",
			input:    "Hello ${name}",
			expected: "Hello \\${name}",
		},
		{
			name:     "backslash",
			input:    "Path\\to\\file",
			expected: "Path\\\\to\\\\file",
		},
		{
			name:     "mixed special chars",
			input:    "Text with `backtick` and ${expr} and \\backslash",
			expected: "Text with \\`backtick\\` and \\${expr} and \\\\backslash",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeTemplateLiteral(tt.input)
			if result != tt.expected {
				t.Errorf("escapeTemplateLiteral() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestRenderElementToJS_CompleteElement tests rendering a complete element with attributes and children
// Cognitive Load: 14 (Integration test with nested structure)
func TestRenderElementToJS_CompleteElement(t *testing.T) {
	elem := &ast.Element{
		TagName: "div",
		Attributes: []ast.Attribute{
			{Name: "class", Value: "container"},
			{Name: "x-data", Value: "{ count: {count} }"},
		},
		Children: []ast.Node{
			&ast.TextNode{Content: "Count: "},
			&ast.Element{
				TagName: "span",
				Attributes: []ast.Attribute{
					{Name: "x-text", Value: "count"},
				},
			},
		},
	}

	var sb strings.Builder
	ctx := &RenderContext{}
	renderElementToJS(elem, &sb, ctx)

	result := sb.String()
	expected := `<div class="container" x-data="{ count: ${props.count} }">Count: <span x-text="count"></span></div>`

	if result != expected {
		t.Errorf("renderElementToJS() =\n%q\nwant:\n%q", result, expected)
	}
}

// TestExtractArrowFunctionParams tests arrow function parameter extraction
// Cognitive Load: 8 (Arrow function parameter detection)
func TestExtractArrowFunctionParams(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string // Expected parameter names
	}{
		{
			name:     "Two parameters",
			input:    "products.reduce((sum, p) => sum + p.price, 0)",
			expected: []string{"sum", "p"},
		},
		{
			name:     "Single parameter with parens",
			input:    "items.map((x) => x * 2)",
			expected: []string{"x"},
		},
		{
			name:     "Single parameter without parens",
			input:    "items.filter(x => x > 10)",
			expected: []string{"x"},
		},
		{
			name:     "Multiple arrow functions",
			input:    "data.map(x => x.items.filter(y => y > 0))",
			expected: []string{"x", "y"},
		},
		{
			name:     "No arrow functions",
			input:    "count + total",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractArrowFunctionParams(tt.input)

			// Check that all expected params are present
			for _, expected := range tt.expected {
				if !result[expected] {
					t.Errorf("extractArrowFunctionParams() missing expected param %q\nGot: %v", expected, result)
				}
			}

			// Check that we don't have extra params
			if len(result) != len(tt.expected) {
				t.Errorf("extractArrowFunctionParams() returned %d params, want %d\nGot: %v", len(result), len(tt.expected), result)
			}
		})
	}
}

// TestArrowFunctionBugFix tests the specific bug case from the issue
// Bug: .reduce((sum, p) => was incorrectly extracting "props.sum" instead of "sum"
// Cognitive Load: 6 (Specific regression test)
func TestArrowFunctionBugFix(t *testing.T) {
	// This is the EXACT case from the bug report
	input := "{formatPrice(products.reduce((sum, p) => sum + p.price, 0) / products.length)}"

	// Expected: Arrow function params (sum, p) should NOT get props. prefix
	expected := "${props.formatPrice(props.products.reduce((sum, p) => sum + p.price, 0) / props.products.length)}"

	result := convertAttributeExpressions(input, nil)

	if result != expected {
		t.Errorf("Arrow function parameter extraction failed!\nInput:    %s\nExpected: %s\nGot:      %s",
			input, expected, result)

		// Extract params to show what went wrong
		params := extractArrowFunctionParams("products.reduce((sum, p) => sum + p.price, 0)")
		t.Logf("Extracted arrow params: %v", params)
	}

	// Verify the specific parameters were correctly identified
	params := extractArrowFunctionParams("products.reduce((sum, p) => sum + p.price, 0)")
	if !params["sum"] {
		t.Error("Failed to extract 'sum' parameter from arrow function")
	}
	if !params["p"] {
		t.Error("Failed to extract 'p' parameter from arrow function")
	}

	// Verify they're not prefixed in the output
	if !contains(result, "(sum, p)") {
		t.Errorf("Arrow function parameters were incorrectly modified. Expected '(sum, p)' in output, got: %s", result)
	}
}

// Helper function for test
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		len(s) > len(substr)+1 && strings.Contains(s, substr)))
}

// TestStringLiteralHandling tests the string literal fix (NEW)
// Cognitive Load: 8 (String literal detection and preservation)
// CRITICAL FIX TEST: This test verifies the fix for string literals inside expressions
func TestStringLiteralHandling(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		expected string
	}{
		{
			name:     "single quoted string",
			expr:     "task.completed ? '✓ Done' : '○ Pending'",
			expected: "props.task.completed ? '✓ Done' : '○ Pending'",
		},
		{
			name:     "double quoted string",
			expr:     `title + " - " + description`,
			expected: `props.title + " - " + props.description`,
		},
		{
			name:     "escaped quotes",
			expr:     `message + ' can\'t ' + suffix`,
			expected: `props.message + ' can\'t ' + props.suffix`,
		},
		{
			name:     "nested quotes",
			expr:     `outer + ' "inner" ' + end`,
			expected: `props.outer + ' "inner" ' + props.end`,
		},
		{
			name:     "ternary with strings",
			expr:     "status === 'active' ? 'Yes' : 'No'",
			expected: "props.status === 'active' ? 'Yes' : 'No'",
		},
		{
			name:     "string with numbers",
			expr:     "count + ' items'",
			expected: "props.count + ' items'",
		},
		{
			name:     "empty string",
			expr:     "name || ''",
			expected: "props.name || ''",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := prefixIdentifiersInExpression(tt.expr, nil, nil)
			if result != tt.expected {
				t.Errorf("String literal handling failed\nExpression: %s\nExpected:   %s\nGot:        %s",
					tt.expr, tt.expected, result)
			}
		})
	}
}

// TestStringLiteralInAttributes tests string literals in attribute expressions (NEW)
// Cognitive Load: 9 (End-to-end test with attribute conversion)
func TestStringLiteralInAttributes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "ternary with string literals",
			input:    "{task.completed ? '✓ Done' : '○ Pending'}",
			expected: "${props.task.completed ? '✓ Done' : '○ Pending'}",
		},
		{
			name:     "string concatenation",
			input:    "{firstName + ' ' + lastName}",
			expected: "${props.firstName + ' ' + props.lastName}",
		},
		{
			name:     "string comparison",
			input:    "{status === 'active' ? 'Yes' : 'No'}",
			expected: "${props.status === 'active' ? 'Yes' : 'No'}",
		},
		{
			name:     "empty string fallback",
			input:    "{message || ''}",
			expected: "${props.message || ''}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertAttributeExpressions(tt.input, nil)
			if result != tt.expected {
				t.Errorf("String literal in attribute failed\nInput:    %s\nExpected: %s\nGot:      %s",
					tt.input, tt.expected, result)
			}
		})
	}
}

// TestStringLiteralInComponentRegistry tests the complete pipeline (NEW)
// Cognitive Load: 12 (Full integration test)
func TestStringLiteralInComponentRegistry(t *testing.T) {
	component := ComponentTemplate{
		Name: "TodoItem",
		AST: &ast.Template{
			RootNodes: []ast.Node{
				&ast.Element{
					TagName: "div",
					Attributes: []ast.Attribute{
						{
							Name:  "x-text",
							Value: "{task.completed ? '✓ Done' : '○ Pending'}",
						},
					},
					Children: []ast.Node{},
				},
			},
		},
	}

	result := GenerateComponentRegistry([]ComponentTemplate{component})

	// Should contain the ternary with string literals preserved
	expected := `x-text="${props.task.completed ? '✓ Done' : '○ Pending'}"`
	if !strings.Contains(result, expected) {
		t.Errorf("String literal handling in component registry failed\nExpected to contain: %s\nGot: %s",
			expected, result)
	}

	// Should NOT contain broken syntax like "props.'✓"
	if strings.Contains(result, "props.'") {
		t.Errorf("Found broken syntax with props. prefix inside string literal:\n%s", result)
	}
}

// TestMethodChainingBugFix tests the Bug #2: Method chaining after method calls
// Cognitive Load: 8 (Method chain preservation)
// BUG: .split('').reverse().join('') was producing invalid syntax
func TestMethodChainingBugFix(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		expected string
	}{
		{
			name:     "method chain with empty args",
			expr:     "animal.split('').reverse().join('')",
			expected: "props.animal.split('').reverse().join('')",
		},
		{
			name:     "method chain with args",
			expr:     "str.replace('a', 'b').toUpperCase()",
			expected: "props.str.replace('a', 'b').toUpperCase()",
		},
		{
			name:     "nested method chains",
			expr:     "data.items.filter(x => x.active).map(x => x.id)",
			expected: "props.data.items.filter(x => x.active).map(x => x.id)",
		},
		{
			name:     "triple method chain",
			expr:     "text.trim().toLowerCase().split(' ')",
			expected: "props.text.trim().toLowerCase().split(' ')",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := prefixIdentifiersInExpression(tt.expr, nil, nil)
			if result != tt.expected {
				t.Errorf("Method chaining failed\nExpression: %s\nExpected:   %s\nGot:        %s",
					tt.expr, tt.expected, result)
			}
		})
	}
}

// TestQuoteEscapingBugFix tests Bug #1: Unescaped quotes in attribute values
// Cognitive Load: 9 (Quote escaping in attributes)
// BUG: x-if="animal == "cat"" was causing JavaScript syntax errors
func TestQuoteEscapingBugFix(t *testing.T) {
	tests := []struct {
		name     string
		attr     string
		expected string
	}{
		{
			name:     "double quotes in condition",
			attr:     `{animal == "cat"}`,
			expected: `{animal == "cat"}`, // Quotes inside {} should NOT be escaped
		},
		{
			name:     "single quotes preserved",
			attr:     `{animal == 'cat'}`,
			expected: `{animal == 'cat'}`,
		},
		{
			name:     "quotes outside expression",
			attr:     `Say "hello" to {name}`,
			expected: `Say \"hello\" to {name}`, // Quotes OUTSIDE {} should be escaped
		},
		{
			name:     "mixed quotes",
			attr:     `{status == "active"} and "text"`,
			expected: `{status == "active"} and \"text\"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeQuotesInAttributeValue(tt.attr)
			if result != tt.expected {
				t.Errorf("Quote escaping failed\nInput:    %s\nExpected: %s\nGot:      %s",
					tt.attr, tt.expected, result)
			}
		})
	}
}

// TestQuoteEscapingInFullPipeline tests Bug #1 through the complete attribute rendering pipeline
// Cognitive Load: 12 (Integration test)
func TestQuoteEscapingInFullPipeline(t *testing.T) {
	tests := []struct {
		name     string
		attr     ast.Attribute
		expected string
	}{
		{
			name: "x-if with double quotes in condition",
			attr: ast.Attribute{
				Name:  "x-if",
				Value: `{animal == "cat"}`,
			},
			expected: `x-if="${props.animal == "cat"}"`, // Quotes preserved inside ${}
		},
		{
			name: "x-if with single quotes",
			attr: ast.Attribute{
				Name:  "x-if",
				Value: `{animal == 'cat'}`,
			},
			expected: `x-if="${props.animal == 'cat'}"`,
		},
		{
			name: "complex condition",
			attr: ast.Attribute{
				Name:  "x-if",
				Value: `{status == "active" && count > 0}`,
			},
			expected: `x-if="${props.status == "active" && props.count > 0}"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sb strings.Builder
			ctx := &RenderContext{}
			renderAttributeToJS(tt.attr, &sb, ctx, nil)
			result := sb.String()

			if result != tt.expected {
				t.Errorf("Full pipeline quote escaping failed\nInput:    %s\nExpected: %s\nGot:      %s",
					tt.attr.Value, tt.expected, result)
			}
		})
	}
}

// TestMethodChainingInComponentRegistry tests Bug #2 through component registry generation
// Cognitive Load: 13 (Full integration test)
func TestMethodChainingInComponentRegistry(t *testing.T) {
	component := ComponentTemplate{
		Name: "StringManipulator",
		AST: &ast.Template{
			RootNodes: []ast.Node{
				&ast.Element{
					TagName: "div",
					Attributes: []ast.Attribute{
						{
							Name:  "x-text",
							Value: `{animal.split('').reverse().join('')}`,
						},
					},
					Children: []ast.Node{},
				},
			},
		},
	}

	result := GenerateComponentRegistry([]ComponentTemplate{component})

	// Should contain properly chained methods
	expected := `x-text="${props.animal.split('').reverse().join('')}"`
	if !strings.Contains(result, expected) {
		t.Errorf("Method chaining in component registry failed\nExpected to contain: %s\nGot: %s",
			expected, result)
	}

	// Should NOT contain broken syntax
	broken := []string{
		"props..reverse()",       // Double dot
		"split('')props.",        // props. in wrong place
		"props.reverse()",        // Missing chain
	}

	for _, bad := range broken {
		if strings.Contains(result, bad) {
			t.Errorf("Found broken method chain syntax: %s\nFull output: %s", bad, result)
		}
	}
}
