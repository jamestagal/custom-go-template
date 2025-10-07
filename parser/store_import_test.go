package parser

import (
	"testing"
)

// TestParseFenceContent_StoreImport tests parsing of store imports
// Pattern: Table-Driven Tests [Load: 5]
func TestParseFenceContent_StoreImport(t *testing.T) {
	// Mock store registry (simulates what registerStores() would return)
	mockRegistry := map[string]string{
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
	}

	tests := []struct {
		name        string
		input       string
		wantStores  int
		storeName   string
		shouldExist bool
	}{
		{
			name:        "single store import",
			input:       `import store from './stores/auth.js'`,
			wantStores:  1,
			storeName:   "auth",
			shouldExist: true,
		},
		{
			name: "multiple store imports",
			input: `import store from './stores/auth.js'
import store from './stores/cart.js'`,
			wantStores:  2,
			storeName:   "auth", // Check first store
			shouldExist: true,
		},
		{
			name: "store import with component import",
			input: `import Header from './components/Header.html'
import store from './stores/auth.js'`,
			wantStores:  1,
			storeName:   "auth",
			shouldExist: true,
		},
		{
			name: "mixed inline and imported stores",
			input: `import store from './stores/auth.js'

store theme = {
  mode: "light"
}`,
			wantStores:  2,
			storeName:   "auth",
			shouldExist: true,
		},
		{
			name: "store import with quotes variations",
			input: `import store from "./stores/auth.js"`,
			wantStores:  1,
			storeName:   "auth",
			shouldExist: true,
		},
		{
			name:        "import non-existent store (should be skipped)",
			input:       `import store from './stores/nonexistent.js'`,
			wantStores:  0,
			storeName:   "nonexistent",
			shouldExist: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse with store registry
			fence := ParseFenceContentWithStores(tt.input, mockRegistry)

			if len(fence.Stores) != tt.wantStores {
				t.Errorf("Expected %d stores, got %d", tt.wantStores, len(fence.Stores))
				return
			}

			if tt.shouldExist {
				storeDef, exists := fence.Stores[tt.storeName]
				if !exists {
					t.Errorf("Expected store '%s' not found. Available stores: %v",
						tt.storeName, getStoreKeys(fence.Stores))
					return
				}

				// Verify store content is loaded from registry
				expectedContent := mockRegistry[tt.storeName]
				if storeDef != expectedContent {
					t.Errorf("Store content mismatch for '%s'.\nGot:\n%s\n\nExpected:\n%s",
						tt.storeName, storeDef, expectedContent)
				}
			} else {
				if _, exists := fence.Stores[tt.storeName]; exists {
					t.Errorf("Store '%s' should not exist but was found", tt.storeName)
				}
			}
		})
	}
}

// TestParseFenceContent_StoreImportPathVariations tests different path formats
func TestParseFenceContent_StoreImportPathVariations(t *testing.T) {
	mockRegistry := map[string]string{
		"auth": `{ isLoggedIn: false }`,
	}

	tests := []struct {
		name      string
		input     string
		wantStore string
	}{
		{
			name:      "path with ./stores/",
			input:     `import store from './stores/auth.js'`,
			wantStore: "auth",
		},
		{
			name:      "path with stores/ (no dot)",
			input:     `import store from 'stores/auth.js'`,
			wantStore: "auth",
		},
		{
			name:      "path with ../stores/",
			input:     `import store from '../stores/auth.js'`,
			wantStore: "auth",
		},
		{
			name:      "path with /stores/",
			input:     `import store from '/stores/auth.js'`,
			wantStore: "auth",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fence := ParseFenceContentWithStores(tt.input, mockRegistry)

			if len(fence.Stores) != 1 {
				t.Errorf("Expected 1 store, got %d", len(fence.Stores))
				return
			}

			if _, exists := fence.Stores[tt.wantStore]; !exists {
				t.Errorf("Expected store '%s' not found. Available: %v",
					tt.wantStore, getStoreKeys(fence.Stores))
			}
		})
	}
}

// TestParseFenceContent_StoreImportOverride tests inline stores override imports
func TestParseFenceContent_StoreImportOverride(t *testing.T) {
	mockRegistry := map[string]string{
		"auth": `{ isLoggedIn: false }`,
	}

	input := `import store from './stores/auth.js'

store auth = {
  isLoggedIn: true
}`

	fence := ParseFenceContentWithStores(input, mockRegistry)

	if len(fence.Stores) != 1 {
		t.Errorf("Expected 1 store, got %d", len(fence.Stores))
		return
	}

	authStore := fence.Stores["auth"]
	// Inline store should override imported store
	expected := `{
  isLoggedIn: true
}`

	if authStore != expected {
		t.Errorf("Inline store should override import.\nGot:\n%s\n\nExpected:\n%s",
			authStore, expected)
	}
}

// TestParseFenceContent_MultipleStoreImports tests multiple store imports
func TestParseFenceContent_MultipleStoreImports(t *testing.T) {
	mockRegistry := map[string]string{
		"auth":  `{ isLoggedIn: false }`,
		"cart":  `{ items: [] }`,
		"theme": `{ mode: "light" }`,
	}

	input := `import store from './stores/auth.js'
import store from './stores/cart.js'
import store from './stores/theme.js'`

	fence := ParseFenceContentWithStores(input, mockRegistry)

	if len(fence.Stores) != 3 {
		t.Errorf("Expected 3 stores, got %d", len(fence.Stores))
		return
	}

	// Verify all stores are present
	for storeName := range mockRegistry {
		if _, exists := fence.Stores[storeName]; !exists {
			t.Errorf("Expected store '%s' not found", storeName)
		}
	}
}

// Helper function to get store keys for error messages
func getStoreKeys(stores map[string]string) []string {
	keys := make([]string, 0, len(stores))
	for k := range stores {
		keys = append(keys, k)
	}
	return keys
}
