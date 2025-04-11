package renderer

import (
	"strings"
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
	"github.com/jimafisk/custom_go_template/transformer"
)

func TestNestedComponentsWithAlpine(t *testing.T) {
	// Create a complex nested component structure with Alpine.js directives
	parentComponent := &ast.Template{
		RootNodes: []ast.Node{
			&ast.Element{
				TagName: "div",
				Attributes: []ast.Attribute{
					{
						Name:       "x-data",
						Value:      "{ parentState: 'active', items: ['item1', 'item2', 'item3'] }",
						IsAlpine:   true,
						AlpineType: "data",
					},
					{
						Name:  "id",
						Value: "parent",
					},
				},
				Children: []ast.Node{
					// Child component 1
					&ast.Element{
						TagName: "div",
						Attributes: []ast.Attribute{
							{
								Name:       "x-data",
								Value:      "{ childState: 'pending', toggle() { this.childState = this.childState === 'active' ? 'pending' : 'active' } }",
								IsAlpine:   true,
								AlpineType: "data",
							},
							{
								Name:       "x-bind:class",
								Value:      "{ active: childState === 'active', pending: childState === 'pending' }",
								IsAlpine:   true,
								AlpineType: "bind",
								AlpineKey:  "class",
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
									&ast.TextNode{Content: "Toggle Child State"},
								},
							},
							&ast.Element{
								TagName: "span",
								Attributes: []ast.Attribute{
									{
										Name:       "x-text",
										Value:      "childState",
										IsAlpine:   true,
										AlpineType: "text",
									},
								},
								Children: []ast.Node{},
							},
						},
					},
					// Child component 2 with loop
					&ast.Element{
						TagName: "ul",
						Attributes: []ast.Attribute{},
						Children: []ast.Node{
							&ast.Element{
								TagName: "template",
								Attributes: []ast.Attribute{
									{
										Name:       "x-for",
										Value:      "item in items",
										IsAlpine:   true,
										AlpineType: "for",
									},
								},
								Children: []ast.Node{
									&ast.Element{
										TagName: "li",
										Attributes: []ast.Attribute{
											{
												Name:       "x-text",
												Value:      "item",
												IsAlpine:   true,
												AlpineType: "text",
											},
											{
												Name:       "x-bind:class",
												Value:      "{ highlight: parentState === 'active' }",
												IsAlpine:   true,
												AlpineType: "bind",
												AlpineKey:  "class",
											},
										},
										Children: []ast.Node{},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Test rendering the complex nested component
	t.Run("nested components with alpine directives", func(t *testing.T) {
		// Transform the template
		props := map[string]any{}
		transformed := transformer.TransformAST(parentComponent, props)
		
		// Render the transformed template
		var sb strings.Builder
		for _, node := range transformed.RootNodes {
			renderNode(&sb, node)
		}
		result := sb.String()
		
		// Check for expected Alpine.js directives
		expectedDirectives := []string{
			`x-data="{ parentState: 'active', items: ['item1', 'item2', 'item3'] }"`,
			`x-data="{ childState: 'pending', toggle() { this.childState = this.childState === 'active' ? 'pending' : 'active' } }"`,
			`x-bind:class="{ active: childState === 'active', pending: childState === 'pending' }"`,
			`x-on:click="toggle()"`,
			`x-text="childState"`,
			`x-for="item in items"`,
			`x-text="item"`,
			`x-bind:class="{ highlight: parentState === 'active' }"`,
		}
		
		for _, directive := range expectedDirectives {
			if !strings.Contains(result, directive) {
				t.Errorf("Rendered output missing expected directive: %s", directive)
			}
		}
	})
}

func TestComplexDataStructures(t *testing.T) {
	// Test with complex data structures in Alpine.js
	complexDataTemplate := &ast.Template{
		RootNodes: []ast.Node{
			&ast.Element{
				TagName: "div",
				Attributes: []ast.Attribute{
					{
						Name: "x-data",
						Value: `{
							user: {
								profile: {
									name: 'John Doe',
									age: 30,
									contact: {
										email: 'john@example.com',
										phone: '555-1234'
									}
								},
								preferences: {
									theme: 'dark',
									notifications: true
								},
								updateProfile() {
									this.profile.age++;
									this.profile.contact.email = 'updated@example.com';
								},
								get fullName() {
									return this.profile.name + ' (Age: ' + this.profile.age + ')';
								}
							},
							items: [
								{ id: 1, name: 'Item 1', selected: true },
								{ id: 2, name: 'Item 2', selected: false },
								{ id: 3, name: 'Item 3', selected: false }
							],
							selectItem(id) {
								this.items.forEach(item => {
									item.selected = item.id === id;
								});
							},
							get selectedItem() {
								return this.items.find(item => item.selected);
							},
							$refs: {},
							init() {
								console.log('Component initialized');
								this.$nextTick(() => {
									this.$refs.nameInput.focus();
								});
							}
						}`,
						IsAlpine:   true,
						AlpineType: "data",
					},
				},
				Children: []ast.Node{
					&ast.Element{
						TagName: "div",
						Attributes: []ast.Attribute{
							{
								Name:  "class",
								Value: "user-profile",
							},
						},
						Children: []ast.Node{
							&ast.Element{
								TagName: "h2",
								Attributes: []ast.Attribute{
									{
										Name:       "x-text",
										Value:      "user.fullName",
										IsAlpine:   true,
										AlpineType: "text",
									},
								},
								Children: []ast.Node{},
							},
							&ast.Element{
								TagName: "input",
								Attributes: []ast.Attribute{
									{
										Name:       "x-ref",
										Value:      "nameInput",
										IsAlpine:   true,
										AlpineType: "ref",
									},
									{
										Name:       "x-model",
										Value:      "user.profile.name",
										IsAlpine:   true,
										AlpineType: "model",
									},
								},
								Children: []ast.Node{},
							},
							&ast.Element{
								TagName: "button",
								Attributes: []ast.Attribute{
									{
										Name:       "x-on:click",
										Value:      "user.updateProfile()",
										IsAlpine:   true,
										AlpineType: "on",
										AlpineKey:  "click",
									},
								},
								Children: []ast.Node{
									&ast.TextNode{Content: "Update Profile"},
								},
							},
						},
					},
					&ast.Element{
						TagName: "div",
						Attributes: []ast.Attribute{
							{
								Name:  "class",
								Value: "items-list",
							},
						},
						Children: []ast.Node{
							&ast.Element{
								TagName: "h3",
								Attributes: []ast.Attribute{},
								Children: []ast.Node{
									&ast.TextNode{Content: "Items"},
								},
							},
							&ast.Element{
								TagName: "ul",
								Attributes: []ast.Attribute{},
								Children: []ast.Node{
									&ast.Element{
										TagName: "template",
										Attributes: []ast.Attribute{
											{
												Name:       "x-for",
												Value:      "item in items",
												IsAlpine:   true,
												AlpineType: "for",
											},
										},
										Children: []ast.Node{
											&ast.Element{
												TagName: "li",
												Attributes: []ast.Attribute{
													{
														Name:       "x-bind:class",
														Value:      "{ selected: item.selected }",
														IsAlpine:   true,
														AlpineType: "bind",
														AlpineKey:  "class",
													},
													{
														Name:       "x-on:click",
														Value:      "selectItem(item.id)",
														IsAlpine:   true,
														AlpineType: "on",
														AlpineKey:  "click",
													},
												},
												Children: []ast.Node{
													&ast.Element{
														TagName: "span",
														Attributes: []ast.Attribute{
															{
																Name:       "x-text",
																Value:      "item.name",
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
								},
							},
							&ast.Element{
								TagName: "div",
								Attributes: []ast.Attribute{
									{
										Name:       "x-show",
										Value:      "selectedItem",
										IsAlpine:   true,
										AlpineType: "show",
									},
								},
								Children: []ast.Node{
									&ast.TextNode{Content: "Selected: "},
									&ast.Element{
										TagName: "span",
										Attributes: []ast.Attribute{
											{
												Name:       "x-text",
												Value:      "selectedItem.name",
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
				},
			},
		},
	}

	t.Run("complex data structures with alpine directives", func(t *testing.T) {
		// Transform the template
		props := map[string]any{}
		transformed := transformer.TransformAST(complexDataTemplate, props)
		
		// Render the transformed template
		var sb strings.Builder
		for _, node := range transformed.RootNodes {
			renderNode(&sb, node)
		}
		result := sb.String()
		
		// Check that the complex x-data attribute is preserved
		if !strings.Contains(result, "x-data=") {
			t.Errorf("Rendered output missing x-data attribute")
		}
		
		// Check for nested object properties
		complexProperties := []string{
			"user.profile.name",
			"user.updateProfile()",
			"user.fullName",
			"item in items",
			"selectedItem.name",
		}
		
		for _, prop := range complexProperties {
			if !strings.Contains(result, prop) {
				t.Errorf("Rendered output missing complex property reference: %s", prop)
			}
		}
		
		// Check for Alpine.js magic properties
		if !strings.Contains(result, "$refs") || !strings.Contains(result, "$nextTick") {
			t.Errorf("Rendered output missing Alpine.js magic properties")
		}
	})
}

func TestEdgeCaseScenarios(t *testing.T) {
	tests := []struct {
		name     string
		jsCode   string
		expected string
	}{
		{
			name:     "empty object",
			jsCode:   "{}",
			expected: "{}",
		},
		{
			name:     "object with only methods",
			jsCode:   "{ method1() {}, method2() {} }",
			expected: "method",
		},
		{
			name:     "object with special characters",
			jsCode:   "{ 'special-key': 'value', 'another.key': 42 }",
			expected: "special-key",
		},
		{
			name:     "object with quotes in strings",
			jsCode:   "{ message: 'It\\'s a \"quoted\" string' }",
			expected: "message",
		},
		{
			name:     "object with nested arrays",
			jsCode:   "{ matrix: [[1,2], [3,4]] }",
			expected: "matrix",
		},
		{
			name:     "object with function calls",
			jsCode:   "{ result: calculateTotal() }",
			expected: "result",
		},
		{
			name:     "object with template literals",
			jsCode:   "{ greeting: `Hello ${name}!` }",
			expected: "greeting",
		},
		{
			name:     "object with computed property names",
			jsCode:   "{ [dynamicKey]: 'value' }",
			expected: "dynamicKey",
		},
		{
			name:     "object with spread operator",
			jsCode:   "{ ...baseConfig, newProp: 'value' }",
			expected: "newProp",
		},
		{
			name:     "object with shorthand properties",
			jsCode:   "{ x, y, z }",
			expected: "x",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test EvalJS with edge case
			result := EvalJS(tt.jsCode, "")
			
			// For complex objects, we expect them to be returned as strings
			resultStr, ok := result.(string)
			if !ok && isComplexJSObject(tt.jsCode) {
				t.Errorf("Expected complex object to be returned as string, got %T", result)
			}
			
			// For simple objects that can be evaluated, we don't check the exact result
			// but just make sure it doesn't cause errors
			
			// For string results, check they contain expected substrings
			if ok && !strings.Contains(resultStr, tt.expected) {
				t.Errorf("EvalJS() result doesn't contain expected substring %q in %q", tt.expected, resultStr)
			}
		})
	}
}
