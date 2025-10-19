# Technical Specification

This is the technical specification for the spec detailed in [..agent-os/specs/2025-10-19-build-time-loop-expansion/spec.md](../.agent-os/specs/2025-10-19-build-time-loop-expansion/spec.md)

## Technical Requirements

### Core Transformation Logic

**File:** `transformer/loops.go`

**Current Behavior (Broken):**
```go
// Creates Alpine x-for template that executes at runtime
func TransformFor(node *ast.Loop, dataScope map[string]interface{}, ...) []ast.Node {
    // Transforms loop into <template x-for="...">
    return []ast.Node{
        &ast.Element{
            TagName: "template",
            Attributes: []ast.Attribute{
                {Name: "x-for", Value: fmt.Sprintf("%s in %s", loopVar, collection)},
            },
            Children: node.Body, // ❌ component doesn't exist in dataScope here
        },
    }
}
```

**Problem:** When the transformer tries to resolve `component.name` inside the loop body, it looks up `component` in dataScope and fails because `component` is a runtime variable that only exists when Alpine.js executes the x-for directive.

**Required Behavior (Fixed):**
```go
// Expands loop at build time, like Svelte
func TransformFor(node *ast.Loop, dataScope map[string]interface{}, ...) []ast.Node {
    // 1. Resolve the collection array from dataScope
    collection, ok := resolveCollectionFromScope(node.Collection, dataScope)
    if !ok {
        return []ast.Node{} // or error handling
    }

    // 2. Expand loop by iterating in Go
    var expandedNodes []ast.Node

    for _, item := range collection {
        // 3. Create iteration scope (clone + add loop var)
        iterationScope := cloneScope(dataScope)
        iterationScope[node.Variable] = item // ✅ Now component exists!

        // 4. Transform loop body with iteration scope
        transformedBody := transformNodes(node.Body, iterationScope, false, false)

        // 5. Append to result
        expandedNodes = append(expandedNodes, transformedBody...)
    }

    return expandedNodes // ✅ Fully expanded HTML, no x-for
}
```

### Scope Cloning

**File:** `transformer/scope.go` (may need to create or add to existing file)

**Requirement:** Safe deep cloning of dataScope map for each loop iteration

```go
// cloneScope creates a deep copy of dataScope for loop iteration
// Pattern: Safe Map Cloning [Cognitive Load: 8]
func cloneScope(dataScope map[string]interface{}) map[string]interface{} {
    clone := make(map[string]interface{}, len(dataScope))

    for key, value := range dataScope {
        // Deep copy depends on value type
        // For now, shallow copy is acceptable since we're adding loop vars
        clone[key] = value
    }

    return clone
}
```

**Note:** Shallow copy is acceptable because:
- Loop variables are new keys (not modifying existing values)
- Values are typically strings, numbers, or data structures from JSON
- Deep mutation of scope values doesn't happen during transformation

### Collection Resolution

**Requirement:** Resolve collection name to actual array from dataScope

```go
// resolveCollectionFromScope looks up collection in dataScope
// Pattern: Map Lookup with Type Assertion [Cognitive Load: 5]
func resolveCollectionFromScope(collectionName string, dataScope map[string]interface{}) ([]interface{}, bool) {
    value, exists := dataScope[collectionName]
    if !exists {
        return nil, false
    }

    // Type assert to array
    array, ok := value.([]interface{})
    return array, ok
}
```

**Edge Cases:**
- Collection doesn't exist in dataScope → return empty/error
- Collection is not an array → return empty/error
- Collection is empty array → return empty result (valid)

### Integration with Component Resolution

**File:** `transformer/dynamic_component_by_name.go`

**Current Issue:** When this transformer tries to resolve `component.name`:
```go
// Tries to get component from scope
name := evaluateExpression("component.name", dataScope)
// ❌ Fails: component not in dataScope (it's in Alpine x-for runtime scope)
```

**After Fix:** Component resolution will work automatically because:
```go
// Loop has already added component to dataScope
iterationScope["component"] = {name: "Hero2436", fields: {...}}

// Now resolution succeeds
name := evaluateExpression("component.name", iterationScope)
// ✅ Returns "Hero2436"
```

**No changes needed** in `dynamic_component_by_name.go` - it will work once loop variables exist in scope.

### Output Comparison

**Input Template:**
```html
---
export let components
---

{for component in components}
  <Component:dynamic name={component.name} {...component.fields} />
{/for}
```

**Input Data (from JSON):**
```json
{
  "components": [
    {"name": "Hero2436", "fields": {"title": "Welcome"}},
    {"name": "Services2437", "fields": {"title": "Our Services"}}
  ]
}
```

**Current Output (Broken):**
```html
<template x-for="component in components">
  <!-- ERROR: component.name cannot be resolved -->
</template>
```

**Expected Output (Fixed):**
```html
<div x-data='{"title":"Welcome"}' class="hero">
  <h1 x-text="title">Welcome</h1>
</div>

<div x-data='{"title":"Our Services"}' class="services">
  <h2 x-text="title">Our Services</h2>
</div>
```

### Performance Considerations

**Loop Expansion Cost:**
- Array size: Acceptable for typical use cases (<100 components per page)
- Each iteration: Clone scope (O(n) where n = scope size) + transform body
- Total: O(items × scope_size × body_complexity)

**Optimization Opportunities (Future):**
- Cache transformed body if all iterations use same structure
- Lazy scope cloning (only clone if body modifies scope)
- Compile-time constant folding

**Acceptable Trade-off:**
Build-time cost is acceptable because:
- Happens once during build, not on every page load
- Produces fully static HTML (better runtime performance)
- Eliminates need for client-side component registry

### Testing Requirements

**Unit Tests:**
```go
func TestTransformFor_BuildTimeExpansion(t *testing.T) {
    dataScope := map[string]interface{}{
        "components": []interface{}{
            map[string]interface{}{"name": "Hero", "title": "Test"},
        },
    }

    loopNode := &ast.Loop{
        Variable: "component",
        Collection: "components",
        Body: []ast.Node{/* component template */},
    }

    result := TransformFor(loopNode, dataScope)

    // Should return expanded nodes, not x-for template
    assert.NotContains(t, result, "x-for")
    assert.Contains(t, result, "Hero") // Component rendered
}
```

**Integration Tests:**
- Load JSON with components array
- Transform template with `{for component in components}`
- Verify each component in array produces rendered HTML
- Verify no x-for templates in output

### Files to Modify

1. **`transformer/loops.go`** (Main changes)
   - Modify `TransformFor` function to expand at build time
   - Add collection resolution logic
   - Add iteration scope management

2. **`transformer/scope.go`** (Helper functions)
   - Add `cloneScope` function
   - Add `resolveCollectionFromScope` function

3. **`transformer/transformer.go`** (Integration)
   - Ensure `transformNodes` is accessible for recursive body transformation
   - May need to expose or refactor existing internal functions

4. **`transformer/loops_test.go`** (Testing)
   - Add unit tests for build-time expansion
   - Add tests for scope cloning
   - Add tests for collection resolution

### Migration Notes

**Breaking Changes:**
- Templates relying on Alpine x-for for loops will now get expanded HTML
- This is actually the desired behavior (matches Svelte)
- No user-facing breaking changes if templates use recommended syntax

**Backwards Compatibility:**
- Static loops (non-dynamic) continue to work
- Dynamic component resolution now works (previously broken)
- Alpine x-for can still be used directly in templates if needed (via literal `<template x-for="...">`)

## External Dependencies

None - this implementation uses only Go standard library:
- `fmt` for string formatting
- Standard map/slice operations
- Existing transformer utilities
