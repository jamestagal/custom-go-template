package tests

import (
	"testing"

	"github.com/jimafisk/custom_go_template/ast"
	"github.com/jimafisk/custom_go_template/parser"
	"github.com/jimafisk/custom_go_template/renderer"
	"github.com/jimafisk/custom_go_template/transformer"
)

// TestStoreIntegrationE2E tests the complete store system flow
// from parsing through transformation to rendering
//
// Pattern: End-to-End Integration Test [Load: 22]
// Cognitive Load: 22 (setup: 4, parse: 4, transform: 4, render: 4, assertions: 6)
func TestStoreIntegrationE2E(t *testing.T) {
	// Setup: Simulate server's store registry
	storeRegistry := map[string]string{
		"auth": `{
  isLoggedIn: false,
  user: null,
  login() {
    this.isLoggedIn = true;
    this.user = { name: 'Test User' };
  }
}`,
		"cart": `{
  items: [],
  total: 0,
  addItem(item) {
    this.items.push(item);
    this.total += item.price;
  }
}`,
		"theme": `{
  mode: "light",
  toggle() {
    this.mode = this.mode === "light" ? "dark" : "light";
  }
}`,
	}

	tests := []struct {
		name           string
		template       string
		wantStores     []string
		wantInMarkup   []string
		wantInScript   []string
		wantStoreCount int
	}{
		{
			name: "inline store only",
			template: `---
store counter = {
  count: 0,
  increment() { this.count++; }
}
---
<div>{$counter.count}</div>`,
			wantStores:     []string{"counter"},
			wantInMarkup:   []string{"x-text=\"$store.counter.count\""},
			wantInScript:   []string{"Alpine.store('counter'", "count: 0"},
			wantStoreCount: 1,
		},
		{
			name: "imported store only",
			template: `---
import store from './stores/auth.js'
---
<div>{$auth.isLoggedIn}</div>`,
			wantStores:     []string{"auth"},
			wantInMarkup:   []string{"x-text=\"$store.auth.isLoggedIn\""},
			wantInScript:   []string{"Alpine.store('auth'", "isLoggedIn: false"},
			wantStoreCount: 1,
		},
		{
			name: "inline overrides imported",
			template: `---
import store from './stores/auth.js'

store auth = {
  isLoggedIn: true,
  user: { name: 'Override User' }
}
---
<div>{$auth.isLoggedIn}</div>`,
			wantStores:     []string{"auth"},
			wantInMarkup:   []string{"x-text=\"$store.auth.isLoggedIn\""},
			wantInScript:   []string{"Alpine.store('auth'", "isLoggedIn: true", "Override User"},
			wantStoreCount: 1,
		},
		{
			name: "multiple imported stores",
			template: `---
import store from './stores/auth.js'
import store from './stores/cart.js'
---
<div>
  {$auth.isLoggedIn}
  {$cart.total}
</div>`,
			wantStores:   []string{"auth", "cart"},
			wantInMarkup: []string{"x-text=\"$store.auth.isLoggedIn\"", "x-text=\"$store.cart.total\""},
			wantInScript: []string{
				"Alpine.store('auth'",
				"Alpine.store('cart'",
				"isLoggedIn: false",
				"items: []",
			},
			wantStoreCount: 2,
		},
		{
			name: "mixed inline and imported stores",
			template: `---
import store from './stores/auth.js'

store counter = {
  count: 0
}
---
<div>
  {$auth.isLoggedIn}
  {$counter.count}
</div>`,
			wantStores:   []string{"auth", "counter"},
			wantInMarkup: []string{"x-text=\"$store.auth.isLoggedIn\"", "x-text=\"$store.counter.count\""},
			wantInScript: []string{
				"Alpine.store('auth'",
				"Alpine.store('counter'",
				"isLoggedIn: false",
				"count: 0",
			},
			wantStoreCount: 2,
		},
		{
			name: "external stores added to unused imports",
			template: `---
import store from './stores/auth.js'
---
<div>
  {$auth.isLoggedIn}
  {$theme.mode}
</div>`,
			wantStores:   []string{"auth", "theme"},
			wantInMarkup: []string{"x-text=\"$store.auth.isLoggedIn\"", "x-text=\"$store.theme.mode\""},
			wantInScript: []string{
				"Alpine.store('auth'",
				"Alpine.store('theme'",
				"isLoggedIn: false",
				"mode: \"light\"",
			},
			wantStoreCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Step 1: Parse template
			template, err := parser.ParseTemplate(tt.template)
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			// Step 2: Extract fence section and parse with store registry
			var fenceSection *ast.FenceSection
			for _, node := range template.RootNodes {
				if fence, ok := node.(*ast.FenceSection); ok {
					fenceSection = fence
					break
				}
			}

			if fenceSection == nil {
				t.Fatal("No fence section found in template")
			}

			// Parse fence content with store registry (simulates server behavior)
			fenceWithStores := parser.ParseFenceContentWithStores(fenceSection.RawContent, storeRegistry)

			// Replace the fence section in the template
			for i, node := range template.RootNodes {
				if _, ok := node.(*ast.FenceSection); ok {
					template.RootNodes[i] = fenceWithStores
					break
				}
			}

			// Step 3: Transform template
			transformed := transformer.TransformAST(template, nil)

			// Step 4: Get tracked stores from transformer
			referencedStores, allDefinitions := transformer.GetTrackedStores(transformed)
			referencedStoreDefs := transformer.GetReferencedStoreDefinitions(allDefinitions, referencedStores)

			// Step 5: Merge with external stores if referenced but not defined
			finalStores := make(map[string]string)

			// Add all referenced stores (inline + imported)
			for name, def := range referencedStoreDefs {
				finalStores[name] = def
			}

			// Add external stores if referenced but not yet in finalStores
			for _, storeName := range referencedStores {
				if _, exists := finalStores[storeName]; !exists {
					if externalDef, exists := storeRegistry[storeName]; exists {
						finalStores[storeName] = externalDef
					}
				}
			}

			// Step 6: Render with stores
			markup, script, _ := renderer.RenderWithStores(transformed, finalStores)

			// Assertions
			if len(finalStores) != tt.wantStoreCount {
				t.Errorf("Expected %d stores, got %d", tt.wantStoreCount, len(finalStores))
			}

			for _, storeName := range tt.wantStores {
				if _, exists := finalStores[storeName]; !exists {
					t.Errorf("Expected store '%s' not found in final stores", storeName)
				}
			}

			for _, expectedMarkup := range tt.wantInMarkup {
				if !contains(markup, expectedMarkup) {
					t.Errorf("Expected markup to contain '%s', but it doesn't.\nMarkup: %s",
						expectedMarkup, markup)
				}
			}

			for _, expectedScript := range tt.wantInScript {
				if !contains(script, expectedScript) {
					t.Errorf("Expected script to contain '%s', but it doesn't.\nScript: %s",
						expectedScript, script)
				}
			}
		})
	}
}

// TestStorePriorityE2E tests the store priority system
// Priority: Inline > Imported > External
//
// Pattern: Priority Test [Load: 18]
func TestStorePriorityE2E(t *testing.T) {
	storeRegistry := map[string]string{
		"config": `{
  source: "external",
  value: 100
}`,
	}

	tests := []struct {
		name           string
		template       string
		expectedSource string
		expectedValue  string
	}{
		{
			name: "inline overrides imported",
			template: `---
import store from './stores/config.js'

store config = {
  source: "inline",
  value: 999
}
---
<div>{$config.source}</div>`,
			expectedSource: "inline",
			expectedValue:  "999",
		},
		{
			name: "imported overrides external",
			template: `---
import store from './stores/config.js'
---
<div>{$config.source}</div>`,
			expectedSource: "external",
			expectedValue:  "100",
		},
		{
			name: "external used if not imported",
			template: `---
---
<div>{$config.source}</div>`,
			expectedSource: "external",
			expectedValue:  "100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse template
			template, err := parser.ParseTemplate(tt.template)
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			// Extract fence section
			var fenceSection *ast.FenceSection
			for _, node := range template.RootNodes {
				if fence, ok := node.(*ast.FenceSection); ok {
					fenceSection = fence
					break
				}
			}

			if fenceSection == nil {
				t.Fatal("No fence section found in template")
			}

			// Parse fence with store registry
			fenceWithStores := parser.ParseFenceContentWithStores(fenceSection.RawContent, storeRegistry)

			// Replace fence in template
			for i, node := range template.RootNodes {
				if _, ok := node.(*ast.FenceSection); ok {
					template.RootNodes[i] = fenceWithStores
					break
				}
			}

			// Transform
			transformed := transformer.TransformAST(template, nil)

			// Get tracked stores
			referencedStores, allDefinitions := transformer.GetTrackedStores(transformed)
			referencedStoreDefs := transformer.GetReferencedStoreDefinitions(allDefinitions, referencedStores)

			// Merge with external
			finalStores := make(map[string]string)
			for name, def := range referencedStoreDefs {
				finalStores[name] = def
			}
			for _, storeName := range referencedStores {
				if _, exists := finalStores[storeName]; !exists {
					if externalDef, exists := storeRegistry[storeName]; exists {
						finalStores[storeName] = externalDef
					}
				}
			}

			// Render
			_, script, _ := renderer.RenderWithStores(transformed, finalStores)

			// Check that script contains expected values
			if !contains(script, tt.expectedSource) {
				t.Errorf("Expected script to contain source '%s', but it doesn't.\nScript: %s",
					tt.expectedSource, script)
			}

			if !contains(script, tt.expectedValue) {
				t.Errorf("Expected script to contain value '%s', but it doesn't.\nScript: %s",
					tt.expectedValue, script)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || len(s) > len(substr) && stringContains(s, substr))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
