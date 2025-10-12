package transformer

import (
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
)

// TestLiteralContentInPreTag tests that content inside <pre> tags is not transformed
// This is a regression test for the bug where ${$store.theme} was incorrectly transformed to ${$store.store.theme}
func TestLiteralContentInPreTag(t *testing.T) {
	tests := []struct {
		name     string
		nodes    []ast.Node
		expected string
	}{
		{
			name: "pre tag with store expression should not be transformed",
			nodes: []ast.Node{
				&ast.Element{
					TagName: "pre",
					Children: []ast.Node{
						&ast.TextNode{
							Content: ":style=\"`background: ${$store.theme.getCurrentColors().background};`\"",
						},
					},
				},
			},
			expected: "<pre>:style=\"`background: ${$store.theme.getCurrentColors().background};`\"</pre>",
		},
		{
			name: "code tag with variable expressions should not be transformed",
			nodes: []ast.Node{
				&ast.Element{
					TagName: "code",
					Children: []ast.Node{
						&ast.TextNode{
							Content: "{user.name} + {user.age}",
						},
					},
				},
			},
			expected: `<code>{user.name} + {user.age}</code>`,
		},
		{
			name: "textarea with expressions should not be transformed",
			nodes: []ast.Node{
				&ast.Element{
					TagName: "textarea",
					Children: []ast.Node{
						&ast.TextNode{
							Content: "Default value: {count}",
						},
					},
				},
			},
			expected: `<textarea>Default value: {count}</textarea>`,
		},
		{
			name: "normal div with expressions should still be transformed",
			nodes: []ast.Node{
				&ast.Element{
					TagName: "div",
					Children: []ast.Node{
						&ast.TextNode{
							Content: "Count: {count}",
						},
					},
				},
			},
			expected: `<div>Count: <span x-text="count"></span></div>`,
		},
		{
			name: "pre tag with nested code tag",
			nodes: []ast.Node{
				&ast.Element{
					TagName: "pre",
					Children: []ast.Node{
						&ast.Element{
							TagName: "code",
							Children: []ast.Node{
								&ast.TextNode{
									Content: "@click=\"$store.auth.login()\"",
								},
							},
						},
					},
				},
			},
			expected: `<pre><code>@click="$store.auth.login()"</code></pre>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dataScope := make(map[string]any)

			// Transform nodes with literal context flag = false (root level)
			result := transformNodes(tt.nodes, dataScope, false, false)

			// Render to string for comparison (using existing helper from loops_stores_test.go)
			rendered := renderNodesToString(result)

			if rendered != tt.expected {
				t.Errorf("\nExpected: %s\nGot:      %s", tt.expected, rendered)
			}
		})
	}
}

// TestStoreExpressionNotInLiteralContext tests that store expressions outside literal contexts are still transformed
func TestStoreExpressionNotInLiteralContext(t *testing.T) {
	nodes := []ast.Node{
		&ast.Element{
			TagName: "div",
			Attributes: []ast.Attribute{
				{
					Name:  ":style",
					Value: "`background: ${$store.theme.getCurrentColors().background};`",
				},
			},
		},
	}

	dataScope := make(map[string]any)
	result := transformNodes(nodes, dataScope, false, false)

	rendered := renderNodesToString(result)

	// The :style attribute should have store expressions transformed in attribute values
	if !strings.Contains(rendered, "$store.theme") {
		t.Error("Store expression should be present in :style attribute")
	}
}
