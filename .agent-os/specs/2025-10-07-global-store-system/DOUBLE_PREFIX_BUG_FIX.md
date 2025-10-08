# Critical Bug Fix: $store.store.theme Double Prefix Issue

## Date
2025-10-08

## Severity
**CRITICAL** - Breaks all Alpine.js store functionality on pages using `:style`, `:class`, and other binding shorthands

## Reported Issue
Template source correctly uses `$store.theme.getCurrentColors()` in `:style` attributes, but rendered HTML shows `$store.store.theme.getCurrentColors()`, causing JavaScript console error: "Cannot read properties of undefined (reading 'theme')"

## Root Cause Analysis

### Investigation Steps
1. Verified parser correctly marks `:style` as `IsAlpine=true` with `AlpineType="bind"` ✅
2. Verified transformer skip logic at `transformer/stores.go:246-249` ✅  
3. Discovered the actual bug in `transformStoreExpressionsInCondition()`

### The Bug
The regex pattern `storeConditionPattern` at line 147:
```go
var storeConditionPattern = regexp.MustCompile(`\$([a-zA-Z_][a-zA-Z0-9_]*)(\.[a-zA-Z_][a-zA-Z0-9_]*)+`)
```

This pattern matches BOTH:
- Template store syntax: `$storeName.property` (needs transformation)
- **Already-transformed Alpine.js syntax: `$store.storeName.property`** ❌

When processing `$store.theme.mode`, the regex:
1. Matches the entire string
2. Captures `$1 = "store"` (the "store name")
3. Captures `$2 = ".mode"`
4. Returns `$store.store.mode` ❌

### Why This Happens
Templates like `store-components-demo.html` use already-transformed Alpine.js syntax:
```html
<body :style="`background: ${$store.theme.getCurrentColors().background};`">
```

The template author wrote `$store.theme` because:
- It's correct Alpine.js syntax
- The template already imports stores
- No transformation needed

But `transformStoreExpressionsInCondition()` doesn't distinguish between:
- `$theme.mode` → needs `$store.` prefix
- `$store.theme.mode` → already has prefix, skip!

## The Fix

### Modified Functions
1. **`transformStoreExpressionsInCondition()`** (lines 175-218)
2. **`transformStoreExpressionInCollection()`** (lines 227-271)

### Fix Logic
Before transformation, check if the captured store name is "store":
```go
storeName := parts[0]

// CRITICAL FIX: Skip if already transformed (storeName == "store")
// This prevents $store.theme.mode from becoming $store.store.theme.mode
if storeName == "store" {
    // Still track the actual store name (second part)
    if len(parts) > 1 {
        actualStoreParts := strings.SplitN(parts[1], ".", 2)
        if len(actualStoreParts) > 0 {
            TrackStoreReference(actualStoreParts[0])
        }
    }
    return match // Return unchanged
}
```

### Key Improvements
1. **Detection**: Check if pattern is already prefixed with `$store.`
2. **Skip transformation**: Return unchanged if already transformed
3. **Maintain tracking**: Still extract and track the actual store name (e.g., "theme")
4. **Zero regression**: Non-transformed patterns still work correctly

## Test Coverage

### New Regression Tests
Created `transformer/stores_double_prefix_test.go` with comprehensive tests:

#### TestTransformAlreadyTransformedStoreExpressions
- ✅ `$store.theme.mode` → `$store.theme.mode` (unchanged)
- ✅ `$store.theme.mode === 'dark'` → unchanged
- ✅ `$store.theme.mode === 'dark' && $store.auth.isLoggedIn` → unchanged
- ✅ Mixed: `$store.theme.mode === 'dark' && $auth.isLoggedIn` → second part transformed
- ✅ Method calls: `$store.theme.getCurrentColors().background` → unchanged

#### TestTransformAlreadyTransformedCollections
- ✅ `$store.cart.items` → unchanged
- ✅ `$store.user.profile.wishlist.products` → unchanged  
- ✅ `$cart.items` → `$store.cart.items` (still transforms when needed)

#### TestAlpineAttributeWithStoreReference
- ✅ `:style` with template literals and method calls
- ✅ `@click` with store methods
- ✅ `x-show` with store properties

#### TestStoreTrackingWithAlreadyTransformed
- ✅ Tracks "theme" and "auth" from `$store.theme.mode` and `$store.auth.isLoggedIn`
- ✅ Does NOT track "store" as a store name

### All Existing Tests Pass
```
go test ./transformer -v -run "Store"
=== RUN   TestTransformConditionalWithStoreExpression (10 subtests) ✅
=== RUN   TestNestedConditionalWithStoreExpressions (2 subtests) ✅
=== RUN   TestStoreExpressionInConditionalContent (2 subtests) ✅
=== RUN   TestTransformLoopWithStoreCollection (5 subtests) ✅
=== RUN   TestTransformNestedLoopsWithStores (2 subtests) ✅
=== RUN   TestTransformLoopWithStoresInConditionals (1 subtest) ✅
=== RUN   TestTransformAttributesWithStores (9 subtests) ✅
=== RUN   TestAlpineStoreTracking (4 subtests) ✅
=== RUN   TestTransformAlreadyTransformedStoreExpressions (5 subtests) ✅
=== RUN   TestTransformAlreadyTransformedCollections (3 subtests) ✅
=== RUN   TestAlpineAttributeWithStoreReference (3 subtests) ✅
=== RUN   TestStoreTrackingWithAlreadyTransformed ✅
PASS
```

## Expected Behavior After Fix

### Before (Buggy)
```html
<!-- Template source -->
<body :style="`background: ${$store.theme.getCurrentColors().background};`">

<!-- Rendered (WRONG) -->
<body :style="`background: ${$store.store.theme.getCurrentColors().background};`">
```
**Console Error**: `Cannot read properties of undefined (reading 'theme')`

### After (Fixed)
```html
<!-- Template source -->
<body :style="`background: ${$store.theme.getCurrentColors().background};`">

<!-- Rendered (CORRECT) -->
<body :style="`background: ${$store.theme.getCurrentColors().background};`">
```
**Console**: No errors, theme toggling works ✅

## Impact

### Fixed Scenarios
1. `:style` attributes with `$store.*` references
2. `:class` attributes with `$store.*` references  
3. Any Alpine binding shorthand (`:attr`) with store references
4. `x-if`, `x-show`, etc. with `$store.*` in conditions
5. Loops over `$store.*` collections
6. Event handlers (`@click`) with store method calls

### Backward Compatibility
- ✅ Templates using `$storeName.property` (without `$store.`) still transform correctly
- ✅ Templates using `$store.storeName.property` (already transformed) now pass through unchanged
- ✅ Mixed usage in same template works correctly
- ✅ Store tracking still works for both patterns

## Files Modified
1. `transformer/stores.go` - Added store name checks in 2 functions
2. `transformer/stores_double_prefix_test.go` - New comprehensive regression tests

## Validation Checklist
- [x] Parser correctly identifies Alpine bindings (`:style`, `:class`, etc.)
- [x] Transformer skips Alpine attributes containing `$store.`
- [x] `transformStoreExpressionsInCondition()` doesn't double-prefix
- [x] `transformStoreExpressionInCollection()` doesn't double-prefix
- [x] Store tracking works for already-transformed references
- [x] All existing tests pass
- [x] New regression tests pass
- [x] No `$store.store.*` patterns in output

## Cognitive Load Assessment
- **Modified functions**: Each function complexity < 15 ✅
- **New logic**: Simple string comparison (storeName == "store")
- **Pattern**: Idempotent transformation (running twice gives same result)

## References
- Issue reported: 2025-10-08
- Root cause: Regex matches already-transformed patterns
- Solution: Check for "store" prefix before transformation
- Test file: `transformer/stores_double_prefix_test.go`
