package renderer

import (
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
	"github.com/jimafisk/custom_go_template/transformer"
)

func TestAlpineJSDirectiveGeneration(t *testing.T) {
	tests := []struct {
		name       string
		attributes []ast.Attribute
		want       string
	}{
		{
			name: "x-data attribute",
			attributes: []ast.Attribute{
				{
					Name:       "x-data",
					Value:      "{ count: 0, increment() { this.count++ } }",
					IsAlpine:   true,
					AlpineType: "data",
				},
			},
			want: `x-data="{ count: 0, increment() { this.count++ } }"`,
		},
		{
			name: "x-text attribute",
			attributes: []ast.Attribute{
				{
					Name:       "x-text",
					Value:      "message",
					IsAlpine:   true,
					AlpineType: "text",
				},
			},
			want: `x-text="message"`,
		},
		{
			name: "x-bind attribute",
			attributes: []ast.Attribute{
				{
					Name:       "x-bind:class",
					Value:      "{ active: isActive }",
					IsAlpine:   true,
					AlpineType: "bind",
					AlpineKey:  "class",
				},
			},
			want: `x-bind:class="{ active: isActive }"`,
		},
		{
			name: "multiple alpine attributes",
			attributes: []ast.Attribute{
				{
					Name:       "x-data",
					Value:      "{ open: false }",
					IsAlpine:   true,
					AlpineType: "data",
				},
				{
					Name:       "x-show",
					Value:      "open",
					IsAlpine:   true,
					AlpineType: "show",
				},
				{
					Name:       "x-on:click",
					Value:      "open = !open",
					IsAlpine:   true,
					AlpineType: "on",
					AlpineKey:  "click",
				},
			},
			want: `x-data="{ open: false }" x-show="open" x-on:click="open = !open"`,
		},
		{
			name: "complex x-data with methods",
			attributes: []ast.Attribute{
				{
					Name: "x-data",
					Value: `{ 
						count: 0, 
						get doubled() { return this.count * 2 },
						increment() { this.count++ }
					}`,
					IsAlpine:   true,
					AlpineType: "data",
				},
			},
			want: `x-data="`,
		},
		{
			name: "alpine magic properties",
			attributes: []ast.Attribute{
				{
					Name: "x-data",
					Value: `{ 
						init() { this.$refs.input.focus() },
						get element() { return this.$el }
					}`,
					IsAlpine:   true,
					AlpineType: "data",
				},
			},
			want: `x-data="`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateAlpineDirectives(tt.attributes)
			if !strings.Contains(got, tt.want) {
				t.Errorf("GenerateAlpineDirectives() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRenderElementWithAlpine(t *testing.T) {
	tests := []struct {
		name     string
		element  *ast.Element
		contains []string
	}{
		{
			name: "simple x-data element",
			element: &ast.Element{
				TagName: "div",
				Attributes: []ast.Attribute{
					{
						Name:       "x-data",
						Value:      "{ count: 0 }",
						IsAlpine:   true,
						AlpineType: "data",
					},
				},
				Children: []ast.Node{
					&ast.TextNode{Content: "Counter"},
				},
			},
			contains: []string{
				`<div`,
				`x-data="{ count: 0 }"`,
				`>Counter</div>`,
			},
		},
		{
			name: "element with multiple alpine directives",
			element: &ast.Element{
				TagName: "button",
				Attributes: []ast.Attribute{
					{
						Name:       "x-on:click",
						Value:      "count++",
						IsAlpine:   true,
						AlpineType: "on",
						AlpineKey:  "click",
					},
					{
						Name:       "x-text",
						Value:      "count",
						IsAlpine:   true,
						AlpineType: "text",
					},
					{
						Name:  "class",
						Value: "btn",
					},
				},
				Children: []ast.Node{},
			},
			contains: []string{
				`<button`,
				`x-on:click="count++"`,
				`x-text="count"`,
				`class="btn"`,
				`</button>`,
			},
		},
		{
			name: "complex alpine component",
			element: &ast.Element{
				TagName: "div",
				Attributes: []ast.Attribute{
					{
						Name: "x-data",
						Value: `{ 
							open: false,
							toggle() { this.open = !this.open }
						}`,
						IsAlpine:   true,
						AlpineType: "data",
					},
				},
				Children: []ast.Node{
					&ast.Element{
						TagName: "button",
						Attributes: []ast.Attribute{
							{
								Name:       "x-on:click",
								Value:      "toggle()",
								IsAlpine:   true,
								AlpineType: "on",
								AlpineKey:  "click",
							},
						},
						Children: []ast.Node{
							&ast.TextNode{Content: "Toggle"},
						},
					},
					&ast.Element{
						TagName: "div",
						Attributes: []ast.Attribute{
							{
								Name:       "x-show",
								Value:      "open",
								IsAlpine:   true,
								AlpineType: "show",
							},
						},
						Children: []ast.Node{
							&ast.TextNode{Content: "Content"},
						},
					},
				},
			},
			contains: []string{
				`x-data=`,
				`toggle()`,
				`x-on:click="toggle()"`,
				`x-show="open"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sb strings.Builder
			renderElement(&sb, tt.element)
			result := sb.String()
			
			for _, substr := range tt.contains {
				if !strings.Contains(result, substr) {
					t.Errorf("renderElement() result does not contain %q\nGot: %s", substr, result)
				}
			}
		})
	}
}

func TestTransformAndRenderAlpine(t *testing.T) {
	// Test the full pipeline: AST -> Transform -> Render
	tests := []struct {
		name      string
		template  *ast.Template
		props     map[string]any
		contains  []string
	}{
		{
			name: "simple counter component",
			template: &ast.Template{
				RootNodes: []ast.Node{
					&ast.Element{
						TagName: "div",
						Attributes: []ast.Attribute{
							{
								Name:       "x-data",
								Value:      "{ count: 0 }",
								IsAlpine:   true,
								AlpineType: "data",
							},
						},
						Children: []ast.Node{
							&ast.Element{
								TagName: "button",
								Attributes: []ast.Attribute{
									{
										Name:       "x-on:click",
										Value:      "count++",
										IsAlpine:   true,
										AlpineType: "on",
										AlpineKey:  "click",
									},
								},
								Children: []ast.Node{
									&ast.TextNode{Content: "Increment"},
								},
							},
							&ast.Element{
								TagName: "span",
								Attributes: []ast.Attribute{
									{
										Name:       "x-text",
										Value:      "count",
										IsAlpine:   true,
										AlpineType: "text",
									},
								},
								Children: []ast.Node{},
							},
						},
					},
				},
			},
			props: map[string]any{},
			contains: []string{
				`x-data="{ count: 0 }"`,
				`x-on:click="count++"`,
				`x-text="count"`,
			},
		},
		{
			name: "component with expressions",
			template: &ast.Template{
				RootNodes: []ast.Node{
					&ast.Element{
						TagName: "div",
						Attributes: []ast.Attribute{
							{
								Name:       "x-data",
								Value:      "{ message: 'Hello' }",
								IsAlpine:   true,
								AlpineType: "data",
							},
						},
						Children: []ast.Node{
							&ast.TextNode{Content: "The message is: {message}"},
						},
					},
				},
			},
			props: map[string]any{},
			contains: []string{
				`x-data="{ message: 'Hello' }"`,
				`The message is: <span x-text="message"></span>`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Transform the template
			transformed := transformer.TransformAST(tt.template, tt.props)
			
			// Render the transformed template
			var sb strings.Builder
			for _, node := range transformed.RootNodes {
				renderNode(&sb, node)
			}
			result := sb.String()
			
			for _, substr := range tt.contains {
				if !strings.Contains(result, substr) {
					t.Errorf("Transform and render result does not contain %q\nGot: %s", substr, result)
				}
			}
		})
	}
}
