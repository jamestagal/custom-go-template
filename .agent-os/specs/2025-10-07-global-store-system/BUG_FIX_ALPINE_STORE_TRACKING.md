# Bug Fix: Alpine Store Tracking for @click Handlers

## Problem

The theme store was not being initialized in Alpine.js, causing runtime errors when theme toggle buttons were clicked.

### Symptoms
- **Error**: Console errors "Cannot read properties of undefined (reading 'mode')" and "Cannot read properties of undefined (reading 'setLight')"
- **Root Cause**: Theme store was defined but not tracked as "referenced"
- **Debug Evidence**:
  ```
  [DEBUG] Referenced stores: [auth cart]      ← Only 2 stores!
  [DEBUG] All definitions keys: [cart theme auth]  ← 3 stores defined
  [DEBUG] Final stores keys: [cart auth]      ← Theme missing!
  ```

### Analysis

The store tracking logic had a critical gap:

1. **What was tracked**: Store references in template syntax like `{$theme.mode}`
2. **What was NOT tracked**: Store references already in Alpine format like `@click="$store.theme.setLight()"`

The issue was in `transformer/stores.go` line 226-229:
```go
// Skip Alpine directives - they're already handled
if attr.IsAlpine {
    transformedAttributes = append(transformedAttributes, attr)
    continue  // ← Skipped WITHOUT tracking!
}
```

Alpine directives with `$store.*` references were being passed through unchanged (correct behavior) but WITHOUT tracking the store names (bug).

## Solution

### Code Changes

**File**: `transformer/stores.go`

1. **Added Alpine store pattern** (line 153):
   ```go
   var alpineStorePattern = regexp.MustCompile(`\$store\.([a-zA-Z_][a-zA-Z0-9_]*)`)
   ```

2. **Added tracking function** (lines 155-166):
   ```go
   func trackAlpineStoreReferences(value string) {
       matches := alpineStorePattern.FindAllStringSubmatch(value, -1)
       for _, match := range matches {
           if len(match) > 1 {
               storeName := match[1]
               TrackStoreReference(storeName)
           }
       }
   }
   ```

3. **Updated transformAttributesWithStores** (lines 244-250):
   ```go
   // CRITICAL FIX: Track Alpine store references before skipping
   if attr.IsAlpine && strings.Contains(attr.Value, "$store.") {
       trackAlpineStoreReferences(attr.Value)
       transformedAttributes = append(transformedAttributes, attr)
       continue
   }
   ```

### Test Coverage

**File**: `transformer/stores_alpine_tracking_test.go`

Added comprehensive tests:
- `TestTrackAlpineStoreReferences` - Tests the tracking function
- `TestTransformAttributesWithAlpineStores` - Tests attribute transformation with tracking
- `TestAlpineStorePattern` - Tests the regex pattern
- `TestBugFix_ThemeStoreNotTracked` - Regression test for this exact bug

All tests pass ✓

## Verification

### Before Fix
```bash
curl -s http://localhost:3333/store-components-demo | grep -c "Alpine.store"
# Output: 2  (auth and cart only)

# Debug logs showed:
[DEBUG] Referenced stores: [auth cart]       ← theme missing
[DEBUG] Final stores keys: [cart auth]       ← theme missing
```

### After Fix
```bash
curl -s http://localhost:3333/store-components-demo | grep -c "Alpine.store"
# Output: 3  (auth, cart, AND theme)

# Debug logs now show:
[DEBUG] Referenced stores: [cart theme auth]  ← All 3 tracked!
[DEBUG] Final stores keys: [cart theme auth]  ← All 3 included!
```

### Browser Verification
- Theme store properly initialized: `Alpine.store('theme', { mode: 'light', ... })`
- Click handlers intact: `@click="$store.theme.setLight()"`
- No console errors when clicking theme buttons

## Impact

### Fixed
✅ Theme store now initializes correctly
✅ Theme toggle buttons work without errors
✅ All `$store.theme.*` references tracked
✅ Pattern applies to ALL Alpine directives with store references

### Pattern Recognition
This fix handles any Alpine directive with `$store.*` syntax:
- `@click="$store.cart.clear()"`
- `x-show="$store.auth.isLoggedIn"`
- `x-text="$store.theme.mode"`
- Any combination of stores in Alpine expressions

## Cognitive Load

**File Total**: 72 (within limits)
- Original: 41
- Task 2.4 additions: 22
- Bug fix additions: 9
  - New pattern: 4
  - Tracking function: 5

**Individual Functions**: All < 15 ✓
- `trackAlpineStoreReferences`: 5
- Pattern update in `transformAttributesWithStores`: +2 to existing

## Related Files
- `transformer/stores.go` - Core fix
- `transformer/stores_alpine_tracking_test.go` - Test coverage
- `cmd/server/main.go` - Store merging logic (unchanged, working correctly)

## Prevention
- Regression test ensures this pattern continues to work
- Pattern is now documented and tested
- Any new Alpine directive patterns will follow this approach
