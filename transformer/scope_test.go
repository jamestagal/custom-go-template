package transformer

import (
	"testing"
)

// TestCloneScope verifies that scope cloning creates an independent copy
// Pattern: Basic scope cloning [Cognitive Load: 5]
func TestCloneScope(t *testing.T) {
	tests := []struct {
		name        string
		originalScope map[string]any
		modifyClone   func(map[string]any)
		verifyOriginal func(*testing.T, map[string]any)
	}{
		{
			name: "basic scope cloning creates independent copy",
			originalScope: map[string]any{
				"title": "Welcome",
				"count": 42,
				"enabled": true,
			},
			modifyClone: func(clone map[string]any) {
				clone["newKey"] = "newValue"
				clone["title"] = "Modified"
			},
			verifyOriginal: func(t *testing.T, original map[string]any) {
				if _, exists := original["newKey"]; exists {
					t.Error("newKey should not exist in original scope")
				}
				if original["title"] != "Welcome" {
					t.Errorf("title should be 'Welcome' in original, got %v", original["title"])
				}
			},
		},
		{
			name: "empty scope clones correctly",
			originalScope: map[string]any{},
			modifyClone: func(clone map[string]any) {
				clone["key1"] = "value1"
			},
			verifyOriginal: func(t *testing.T, original map[string]any) {
				if len(original) != 0 {
					t.Errorf("original scope should be empty, got %d keys", len(original))
				}
			},
		},
		{
			name: "scope with complex values",
			originalScope: map[string]any{
				"user": map[string]any{
					"name": "John",
					"age": 30,
				},
				"items": []interface{}{"a", "b", "c"},
				"count": 10,
			},
			modifyClone: func(clone map[string]any) {
				clone["count"] = 20
				clone["newField"] = "test"
			},
			verifyOriginal: func(t *testing.T, original map[string]any) {
				if original["count"] != 10 {
					t.Errorf("count should be 10, got %v", original["count"])
				}
				if _, exists := original["newField"]; exists {
					t.Error("newField should not exist in original")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clone the scope
			clonedScope := cloneScope(tt.originalScope)

			// Verify clone has same keys initially
			if len(clonedScope) != len(tt.originalScope) {
				t.Errorf("cloned scope should have %d keys, got %d", len(tt.originalScope), len(clonedScope))
			}

			// Modify the clone
			tt.modifyClone(clonedScope)

			// Verify original is unaffected
			tt.verifyOriginal(t, tt.originalScope)
		})
	}
}

// TestCloneScope_NilScope tests that cloning nil scope returns empty map
func TestCloneScope_NilScope(t *testing.T) {
	clone := cloneScope(nil)
	if clone == nil {
		t.Error("cloneScope(nil) should return empty map, not nil")
	}
	if len(clone) != 0 {
		t.Errorf("cloneScope(nil) should return empty map, got %d keys", len(clone))
	}
}

// TestResolveCollectionFromScope verifies collection lookup and type checking
// Pattern: Map lookup with type assertion [Cognitive Load: 6]
func TestResolveCollectionFromScope(t *testing.T) {
	tests := []struct {
		name           string
		dataScope      map[string]any
		collectionName string
		wantArray      []interface{}
		wantOk         bool
		description    string
	}{
		{
			name: "valid array collection",
			dataScope: map[string]any{
				"components": []interface{}{
					map[string]any{"name": "Hero", "title": "Welcome"},
					map[string]any{"name": "Footer", "title": "Contact"},
				},
			},
			collectionName: "components",
			wantArray: []interface{}{
				map[string]any{"name": "Hero", "title": "Welcome"},
				map[string]any{"name": "Footer", "title": "Contact"},
			},
			wantOk:      true,
			description: "should resolve valid array collection",
		},
		{
			name: "empty array collection",
			dataScope: map[string]any{
				"items": []interface{}{},
			},
			collectionName: "items",
			wantArray:      []interface{}{},
			wantOk:         true,
			description:    "empty array is valid collection",
		},
		{
			name: "collection not found",
			dataScope: map[string]any{
				"items": []interface{}{1, 2, 3},
			},
			collectionName: "nonexistent",
			wantArray:      nil,
			wantOk:         false,
			description:    "should return false for missing collection",
		},
		{
			name: "collection is not array - string",
			dataScope: map[string]any{
				"title": "Not an array",
			},
			collectionName: "title",
			wantArray:      nil,
			wantOk:         false,
			description:    "should return false for non-array (string)",
		},
		{
			name: "collection is not array - map",
			dataScope: map[string]any{
				"user": map[string]any{
					"name": "John",
					"age":  30,
				},
			},
			collectionName: "user",
			wantArray:      nil,
			wantOk:         false,
			description:    "should return false for non-array (map)",
		},
		{
			name: "collection is not array - number",
			dataScope: map[string]any{
				"count": 42,
			},
			collectionName: "count",
			wantArray:      nil,
			wantOk:         false,
			description:    "should return false for non-array (number)",
		},
		{
			name: "collection is nil",
			dataScope: map[string]any{
				"items": nil,
			},
			collectionName: "items",
			wantArray:      nil,
			wantOk:         false,
			description:    "should return false for nil value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			array, ok := resolveCollectionFromScope(tt.collectionName, tt.dataScope)

			// Check ok result
			if ok != tt.wantOk {
				t.Errorf("%s: got ok=%v, want ok=%v", tt.description, ok, tt.wantOk)
			}

			// Check array result
			if tt.wantOk {
				if len(array) != len(tt.wantArray) {
					t.Errorf("%s: got array length %d, want %d", tt.description, len(array), len(tt.wantArray))
				}

				// For non-empty arrays, verify content
				if len(tt.wantArray) > 0 {
					// Check first element as sample
					firstWant := tt.wantArray[0]
					firstGot := array[0]

					if wantMap, ok := firstWant.(map[string]any); ok {
						if gotMap, ok := firstGot.(map[string]any); ok {
							if gotMap["name"] != wantMap["name"] {
								t.Errorf("%s: first element name mismatch, got %v, want %v",
									tt.description, gotMap["name"], wantMap["name"])
							}
						} else {
							t.Errorf("%s: first element should be map", tt.description)
						}
					}
				}
			} else {
				if array != nil {
					t.Errorf("%s: array should be nil when ok=false, got %v", tt.description, array)
				}
			}
		})
	}
}

// TestResolveCollectionFromScope_NilScope tests handling of nil dataScope
func TestResolveCollectionFromScope_NilScope(t *testing.T) {
	array, ok := resolveCollectionFromScope("anything", nil)
	if ok {
		t.Error("should return ok=false for nil dataScope")
	}
	if array != nil {
		t.Error("should return nil array for nil dataScope")
	}
}

// TestResolveCollectionFromScope_RealWorldData tests with realistic component data
func TestResolveCollectionFromScope_RealWorldData(t *testing.T) {
	// Simulate real component data from JSON
	dataScope := map[string]any{
		"components": []interface{}{
			map[string]any{
				"name": "Hero2436",
				"fields": map[string]any{
					"title":       "Welcome to Our Site",
					"description": "We build amazing things",
					"link":        "/about",
				},
			},
			map[string]any{
				"name": "Services2437",
				"fields": map[string]any{
					"title":       "Our Services",
					"description": "Quality services for you",
				},
			},
			map[string]any{
				"name": "ContactForm2438",
				"fields": map[string]any{
					"title": "Get in Touch",
					"email": "contact@example.com",
				},
			},
		},
	}

	array, ok := resolveCollectionFromScope("components", dataScope)

	if !ok {
		t.Fatal("should resolve components collection")
	}

	if len(array) != 3 {
		t.Fatalf("should have 3 components, got %d", len(array))
	}

	// Verify first component structure
	firstComponent, ok := array[0].(map[string]any)
	if !ok {
		t.Fatal("first component should be map")
	}

	if firstComponent["name"] != "Hero2436" {
		t.Errorf("first component name should be Hero2436, got %v", firstComponent["name"])
	}

	fields, ok := firstComponent["fields"].(map[string]any)
	if !ok {
		t.Fatal("first component should have fields map")
	}

	if fields["title"] != "Welcome to Our Site" {
		t.Errorf("first component title mismatch, got %v", fields["title"])
	}
}

// TestCloneScope_PreservesCapacity tests that clone maintains similar capacity
func TestCloneScope_PreservesCapacity(t *testing.T) {
	// Create scope with several entries
	original := map[string]any{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
		"key4": "value4",
		"key5": "value5",
	}

	clone := cloneScope(original)

	// Clone should be able to hold all original keys plus new ones
	clone["key6"] = "value6"
	clone["key7"] = "value7"

	// Verify original unchanged
	if len(original) != 5 {
		t.Errorf("original should still have 5 keys, got %d", len(original))
	}

	// Verify clone has new keys
	if len(clone) != 7 {
		t.Errorf("clone should have 7 keys, got %d", len(clone))
	}
}

// TestCloneScope_LoopVariablePattern tests the specific use case for loop expansion
func TestCloneScope_LoopVariablePattern(t *testing.T) {
	// Parent scope with component array
	parentScope := map[string]any{
		"components": []interface{}{
			map[string]any{"name": "Hero", "title": "Welcome"},
			map[string]any{"name": "Footer", "title": "Contact"},
		},
		"pageTitle": "Home",
	}

	// Simulate loop iteration 1
	iteration1Scope := cloneScope(parentScope)
	iteration1Scope["component"] = map[string]any{"name": "Hero", "title": "Welcome"}

	// Simulate loop iteration 2
	iteration2Scope := cloneScope(parentScope)
	iteration2Scope["component"] = map[string]any{"name": "Footer", "title": "Contact"}

	// Verify iterations are independent
	comp1, _ := iteration1Scope["component"].(map[string]any)
	comp2, _ := iteration2Scope["component"].(map[string]any)

	if comp1["name"] == comp2["name"] {
		t.Error("iteration scopes should have different component values")
	}

	// Verify parent scope unchanged
	if _, exists := parentScope["component"]; exists {
		t.Error("parent scope should not have 'component' key")
	}

	// Verify both iterations have access to parent scope
	if iteration1Scope["pageTitle"] != "Home" {
		t.Error("iteration 1 should have access to pageTitle")
	}
	if iteration2Scope["pageTitle"] != "Home" {
		t.Error("iteration 2 should have access to pageTitle")
	}
}
