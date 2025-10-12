package transformer

import (
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
)

// TestTrackStoreReferences tests that store references are tracked during transformation
// Cognitive Load: 8 (test setup + assertions)
func TestTrackStoreReferences(t *testing.T) {
	tests := []struct {
		name              string
		template          *ast.Template
		expectedStores    []string
		expectedDefs      map[string]string
	}{
		{
			name: "single store in text",
			template: &ast.Template{
				RootNodes: []ast.Node{
					&ast.FenceSection{
						Stores: map[string]string{
							"auth": "{ isLoggedIn: false, user: null }",
						},
					},
					&ast.StoreExpressionNode{
						StoreName: "auth",
						Property:  "isLoggedIn",
					},
				},
			},
			expectedStores: []string{"auth"},
			expectedDefs: map[string]string{
				"auth": "{ isLoggedIn: false, user: null }",
			},
		},
		{
			name: "multiple stores in text",
			template: &ast.Template{
				RootNodes: []ast.Node{
					&ast.FenceSection{
						Stores: map[string]string{
							"auth":  "{ isLoggedIn: false }",
							"cart":  "{ items: [], total: 0 }",
							"theme": "{ mode: 'light' }",
						},
					},
					&ast.StoreExpressionNode{
						StoreName: "auth",
						Property:  "isLoggedIn",
					},
					&ast.StoreExpressionNode{
						StoreName: "cart",
						Property:  "items",
					},
				},
			},
			expectedStores: []string{"auth", "cart"},
			expectedDefs: map[string]string{
				"auth":  "{ isLoggedIn: false }",
				"cart":  "{ items: [], total: 0 }",
				"theme": "{ mode: 'light' }",
			},
		},
		{
			name: "store in conditional",
			template: &ast.Template{
				RootNodes: []ast.Node{
					&ast.FenceSection{
						Stores: map[string]string{
							"auth": "{ isLoggedIn: false }",
						},
					},
					&ast.Conditional{
						IfCondition: "$auth.isLoggedIn",
						IfContent: []ast.Node{
							&ast.TextNode{Content: "Welcome!"},
						},
					},
				},
			},
			expectedStores: []string{"auth"},
			expectedDefs: map[string]string{
				"auth": "{ isLoggedIn: false }",
			},
		},
		{
			name: "store in loop collection",
			template: &ast.Template{
				RootNodes: []ast.Node{
					&ast.FenceSection{
						Stores: map[string]string{
							"cart": "{ items: [] }",
						},
					},
					&ast.Loop{
						Iterator:   "item",
						Collection: "$cart.items",
						Content: []ast.Node{
							&ast.TextNode{Content: "Item"},
						},
					},
				},
			},
			expectedStores: []string{"cart"},
			expectedDefs: map[string]string{
				"cart": "{ items: [] }",
			},
		},
		{
			name: "store in loop body",
			template: &ast.Template{
				RootNodes: []ast.Node{
					&ast.FenceSection{
						Stores: map[string]string{
							"cart":  "{ items: [] }",
							"theme": "{ mode: 'light' }",
						},
					},
					&ast.Loop{
						Iterator:   "item",
						Collection: "$cart.items",
						Content: []ast.Node{
							&ast.Element{
								TagName: "div",
								Attributes: []ast.Attribute{
									{Name: "class", Value: "{$theme.mode}"},
								},
							},
						},
					},
				},
			},
			expectedStores: []string{"cart", "theme"},
			expectedDefs: map[string]string{
				"cart":  "{ items: [] }",
				"theme": "{ mode: 'light' }",
			},
		},
		{
			name: "nested conditionals with stores",
			template: &ast.Template{
				RootNodes: []ast.Node{
					&ast.FenceSection{
						Stores: map[string]string{
							"auth": "{ isLoggedIn: false, isPremium: false }",
							"cart": "{ items: [] }",
						},
					},
					&ast.Conditional{
						IfCondition: "$auth.isLoggedIn",
						IfContent: []ast.Node{
							&ast.Conditional{
								IfCondition: "$auth.isPremium",
								IfContent: []ast.Node{
									&ast.StoreExpressionNode{
										StoreName: "cart",
										Property:  "items.length",
									},
								},
							},
						},
					},
				},
			},
			expectedStores: []string{"auth", "cart"},
			expectedDefs: map[string]string{
				"auth": "{ isLoggedIn: false, isPremium: false }",
				"cart": "{ items: [] }",
			},
		},
		{
			name: "store in attributes",
			template: &ast.Template{
				RootNodes: []ast.Node{
					&ast.FenceSection{
						Stores: map[string]string{
							"theme": "{ mode: 'light', color: 'blue' }",
						},
					},
					&ast.Element{
						TagName: "div",
						Attributes: []ast.Attribute{
							{Name: "class", Value: "{$theme.mode}"},
							{Name: "data-color", Value: "{$theme.color}"},
						},
					},
				},
			},
			expectedStores: []string{"theme"},
			expectedDefs: map[string]string{
				"theme": "{ mode: 'light', color: 'blue' }",
			},
		},
		{
			name: "no stores referenced",
			template: &ast.Template{
				RootNodes: []ast.Node{
					&ast.FenceSection{
						Stores: map[string]string{
							"auth":  "{ isLoggedIn: false }",
							"cart":  "{ items: [] }",
							"theme": "{ mode: 'light' }",
						},
					},
					&ast.TextNode{Content: "Hello World"},
				},
			},
			expectedStores: []string{},
			expectedDefs: map[string]string{
				"auth":  "{ isLoggedIn: false }",
				"cart":  "{ items: [] }",
				"theme": "{ mode: 'light' }",
			},
		},
		{
			name: "mixed regular and store expressions",
			template: &ast.Template{
				RootNodes: []ast.Node{
					&ast.FenceSection{
						Stores: map[string]string{
							"auth": "{ user: null }",
						},
						Variables: []ast.VariableNode{
							{Name: "greeting", Value: "'Hello'"},
						},
					},
					&ast.ExpressionNode{Expression: "greeting"},
					&ast.StoreExpressionNode{
						StoreName: "auth",
						Property:  "user.name",
					},
				},
			},
			expectedStores: []string{"auth"},
			expectedDefs: map[string]string{
				"auth": "{ user: null }",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Transform the AST
			transformed := TransformAST(tt.template, map[string]any{})

			// Get tracked stores (this will be the new API)
			referencedStores, storeDefinitions := GetTrackedStores(transformed)

			// Check referenced stores count
			if len(referencedStores) != len(tt.expectedStores) {
				t.Errorf("expected %d referenced stores, got %d", len(tt.expectedStores), len(referencedStores))
			}

			// Check each expected store is tracked
			for _, expectedStore := range tt.expectedStores {
				found := false
				for _, store := range referencedStores {
					if store == expectedStore {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected store %q to be tracked, but it wasn't", expectedStore)
				}
			}

			// Check store definitions match
			if len(storeDefinitions) != len(tt.expectedDefs) {
				t.Errorf("expected %d store definitions, got %d", len(tt.expectedDefs), len(storeDefinitions))
			}

			for name, expectedDef := range tt.expectedDefs {
				actualDef, ok := storeDefinitions[name]
				if !ok {
					t.Errorf("expected store definition for %q, but not found", name)
					continue
				}
				if actualDef != expectedDef {
					t.Errorf("store definition for %q mismatch:\nexpected: %s\ngot:      %s", name, expectedDef, actualDef)
				}
			}
		})
	}
}

// TestGetReferencedStoreDefinitions tests extracting only referenced store definitions
// Cognitive Load: 6 (focused test with clear assertions)
func TestGetReferencedStoreDefinitions(t *testing.T) {
	tests := []struct {
		name                 string
		allDefinitions       map[string]string
		referencedStores     []string
		expectedDefinitions  map[string]string
	}{
		{
			name: "get only referenced stores",
			allDefinitions: map[string]string{
				"auth":  "{ isLoggedIn: false }",
				"cart":  "{ items: [] }",
				"theme": "{ mode: 'light' }",
			},
			referencedStores: []string{"auth", "cart"},
			expectedDefinitions: map[string]string{
				"auth": "{ isLoggedIn: false }",
				"cart": "{ items: [] }",
			},
		},
		{
			name: "no stores referenced",
			allDefinitions: map[string]string{
				"auth": "{ isLoggedIn: false }",
				"cart": "{ items: [] }",
			},
			referencedStores:    []string{},
			expectedDefinitions: map[string]string{},
		},
		{
			name: "all stores referenced",
			allDefinitions: map[string]string{
				"auth":  "{ isLoggedIn: false }",
				"cart":  "{ items: [] }",
				"theme": "{ mode: 'light' }",
			},
			referencedStores: []string{"auth", "cart", "theme"},
			expectedDefinitions: map[string]string{
				"auth":  "{ isLoggedIn: false }",
				"cart":  "{ items: [] }",
				"theme": "{ mode: 'light' }",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the function that filters definitions
			result := GetReferencedStoreDefinitions(tt.allDefinitions, tt.referencedStores)

			// Check the result
			if len(result) != len(tt.expectedDefinitions) {
				t.Errorf("expected %d definitions, got %d", len(tt.expectedDefinitions), len(result))
			}

			for name, expectedDef := range tt.expectedDefinitions {
				actualDef, ok := result[name]
				if !ok {
					t.Errorf("expected definition for %q, but not found", name)
					continue
				}
				if actualDef != expectedDef {
					t.Errorf("definition for %q mismatch:\nexpected: %s\ngot:      %s", name, expectedDef, actualDef)
				}
			}
		})
	}
}

// TestStoreTrackingInComplexTemplate tests store tracking in a realistic complex template
// Cognitive Load: 10 (complex scenario with multiple contexts)
func TestStoreTrackingInComplexTemplate(t *testing.T) {
	template := &ast.Template{
		RootNodes: []ast.Node{
			&ast.FenceSection{
				Stores: map[string]string{
					"auth":  "{ isLoggedIn: false, user: null }",
					"cart":  "{ items: [], total: 0 }",
					"theme": "{ mode: 'light' }",
					"ui":    "{ sidebarOpen: false }",
				},
			},
			&ast.Element{
				TagName: "div",
				Attributes: []ast.Attribute{
					{Name: "class", Value: "{$theme.mode}"},
				},
				Children: []ast.Node{
					&ast.Conditional{
						IfCondition: "$auth.isLoggedIn",
						IfContent: []ast.Node{
							&ast.StoreExpressionNode{
								StoreName: "auth",
								Property:  "user.name",
							},
							&ast.Loop{
								Iterator:   "item",
								Collection: "$cart.items",
								Content: []ast.Node{
									&ast.TextNode{Content: "Item"},
								},
							},
						},
					},
				},
			},
		},
	}

	// Transform
	transformed := TransformAST(template, map[string]any{})

	// Get tracked stores
	referencedStores, storeDefinitions := GetTrackedStores(transformed)

	// Expected: auth, cart, theme (ui is not referenced)
	expectedStores := []string{"auth", "cart", "theme"}

	if len(referencedStores) != len(expectedStores) {
		t.Errorf("expected %d referenced stores, got %d", len(expectedStores), len(referencedStores))
	}

	for _, expectedStore := range expectedStores {
		found := false
		for _, store := range referencedStores {
			if store == expectedStore {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected store %q to be tracked, but it wasn't", expectedStore)
		}
	}

	// Store definitions should contain all 4 stores (not filtered)
	if len(storeDefinitions) != 4 {
		t.Errorf("expected 4 store definitions, got %d", len(storeDefinitions))
	}
}

// Confidence Score: 95%
// - Central validation passed: ✓ +40%
//   - GO-ERROR-CONTEXT: Not applicable for tests ✓
//   - No defer in loops ✓
//   - Slices checked with len() ✓
// - Pattern Completeness: ✓ +30%
//   - All tracking contexts tested (text, conditionals, loops, attributes) ✓
//   - Nested scenarios covered ✓
//   - Edge cases included (no stores, mixed expressions) ✓
// - Agent patterns followed: ✓ +25%
//   - Table-driven tests ✓
//   - Clear test names ✓
//   - Cognitive load documented ✓
// - Tests would pass: ⏳ (pending implementation)
