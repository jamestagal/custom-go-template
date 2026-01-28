# Bug Fix: Dynamic Component Style Aggregation

## Problem

**User Report:** "All imported components styles are being applied except for dynamic components using the `<=` syntax."

Dynamic components using the `<=` syntax in templates were not having their styles aggregated correctly. For example:

```html
<="./components/UserProfile.html" user={user1} showRole={true} />
```

The `UserProfile.html` component has extensive styles (lines 39-220), but they were NOT being included in the aggregated output.

## Root Cause

The `AggregateComponentStyles()` function in `renderer/styles.go` only traversed `FenceSection.Imports` to find component dependencies:

```go
// PHASE 1: Process imported components first
for _, node := range template.RootNodes {
    if fence, ok := node.(*ast.FenceSection); ok {
        if fence.Imports != nil {
            for _, imp := range fence.Imports {
                // Process imported component...
            }
        }
    }
}
```

**Missing:** It did NOT traverse the AST tree looking for:
- `ComponentNode` - Regular component usage like `<MyComponent />`
- `DynamicComponentNode` - Dynamic components with `<=` syntax

## Solution

Enhanced `AggregateComponentStyles()` to perform full AST tree traversal to discover ALL component usage, not just imports.

### New Functions Added

1. **`extractComponentNameFromPath(path string) string`**
   - Extracts component name from static paths
   - Example: `"./components/UserProfile.html"` → `"UserProfile"`
   - Skips runtime variable paths like `"{path}"` or `"./components/{comp}.html"`

2. **`findComponentNodes(nodes []ast.Node) []string`**
   - Recursively traverses entire AST tree
   - Finds all `ComponentNode` and `DynamicComponentNode` instances
   - Searches in:
     - Element children
     - Conditional branches (if/else if/else)
     - Loop content
     - Nested structures

### Updated Algorithm

```go
// PHASE 1: Collect ALL component names from the tree
var allComponentNames []string

// Find components used in template body (ComponentNode and DynamicComponentNode)
allComponentNames = append(allComponentNames, findComponentNodes(template.RootNodes)...)

// Also add components from FenceSection imports
for _, node := range template.RootNodes {
    if fence, ok := node.(*ast.FenceSection); ok {
        if fence.Imports != nil {
            for _, imp := range fence.Imports {
                allComponentNames = append(allComponentNames, imp.Name)
            }
        }
    }
}

// Deduplicate and process all discovered components
// ... (existing recursion logic)
```

## Files Modified

### Primary Changes
- **`renderer/styles.go`**
  - Added `extractComponentNameFromPath()` helper
  - Added `findComponentNodes()` recursive tree traversal
  - Enhanced `AggregateComponentStyles()` to use full tree traversal

### Test Coverage
- **`renderer/styles_test.go`**
  - Added `TestAggregateComponentStyles_DynamicComponentNode`
  - Added `TestAggregateComponentStyles_RegularComponentNode`
  - Added `TestAggregateComponentStyles_ComponentsInConditionals`
  - Added `TestAggregateComponentStyles_ComponentsInLoops`
  - Added `TestAggregateComponentStyles_MixedComponentUsage`
  - Added `TestExtractComponentNameFromPath`
  - Added tests for runtime path skipping

- **`tests/components/style_aggregation_integration_test.go`**
  - Added `TestRenderTemplate_DynamicComponentStyles`
  - Added `TestRenderTemplate_MixedImportsAndDynamicComponents`
  - Added `TestRenderTemplate_DynamicComponentInConditional`
  - Added `TestRenderTemplate_DynamicComponentInLoop`
  - Added `TestRenderTemplate_RuntimeDynamicComponentsSkipped`
  - Added `TestRenderTemplate_VariableInPathSkipped`

## Test Results

All new tests pass:
```
✅ TestAggregateComponentStyles_DynamicComponentNode
✅ TestAggregateComponentStyles_RegularComponentNode  
✅ TestAggregateComponentStyles_ComponentsInConditionals
✅ TestAggregateComponentStyles_ComponentsInLoops
✅ TestAggregateComponentStyles_MixedComponentUsage
✅ TestExtractComponentNameFromPath (6 sub-tests)
✅ TestRenderTemplate_DynamicComponentStyles
✅ TestRenderTemplate_MixedImportsAndDynamicComponents
✅ TestRenderTemplate_DynamicComponentInConditional
✅ TestRenderTemplate_DynamicComponentInLoop
✅ TestRenderTemplate_RuntimeDynamicComponentsSkipped
✅ TestRenderTemplate_VariableInPathSkipped
```

All existing style aggregation tests continue to pass (no regressions).

## Behavior Changes

### Before Fix
- Only `FenceSection.Imports` were processed
- Dynamic components (`<=` syntax) styles were NOT aggregated
- Regular inline `<ComponentNode />` usage was NOT aggregated

### After Fix
- Full AST tree traversal discovers ALL component usage
- Dynamic components with static paths are aggregated correctly
- Regular inline component usage is aggregated correctly
- Runtime variable paths are gracefully skipped (can't resolve at build time)

## Edge Cases Handled

1. **Static paths** - `<="./components/UserProfile.html" />` ✅ Aggregated
2. **Runtime variables** - `<='{path}' />` ✅ Skipped (can't resolve)
3. **Path with variables** - `<="./components/{comp}.html" />` ✅ Skipped
4. **Components in conditionals** - `{if x}<Component />{/if}` ✅ Aggregated
5. **Components in loops** - `{for x}<Component />{/for}` ✅ Aggregated
6. **Mixed imports + inline** - Import Header, use `<=Footer />` ✅ Both aggregated
7. **Deduplication** - Still works correctly ✅
8. **Dependency order** - Still correct (children before parents) ✅

## Success Criteria Met

✅ `AggregateComponentStyles()` traverses entire AST tree
✅ Regular `ComponentNode` instances found and processed
✅ `DynamicComponentNode` instances found and processed  
✅ Helper function extracts component names from paths correctly
✅ Skips dynamic runtime paths (e.g., `{path}`, `{comp}`)
✅ Test for dynamic components passes
✅ All existing tests still pass
✅ No duplicate styles (deduplication still works)
✅ Dependency order preserved (children before parents)

## Impact

This fix ensures that **all component styles are properly aggregated**, regardless of how the component is referenced:
- Traditional imports in `FenceSection`
- Dynamic components with `<=` syntax
- Regular inline component usage
- Components nested in conditionals and loops

The fix maintains backward compatibility and all existing behavior.
