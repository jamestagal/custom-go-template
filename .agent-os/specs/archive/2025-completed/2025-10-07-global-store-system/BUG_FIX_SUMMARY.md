# Bug Fix Summary: Missing Theme Store Initialization

## Issue
The theme store was not being initialized in Alpine.js, causing runtime errors when theme buttons were clicked.

## Root Cause
Store tracking logic only captured stores during **transformation** (converting `{$theme.*}` to `$store.theme.*`), but missed stores that were **already in Alpine format** in the source templates (like `@click="$store.theme.setLight()"`).

## The Gap
```
Source Template:
  - {$theme.mode}                     ← Gets transformed → TRACKED ✓
  - @click="$store.theme.setLight()"  ← Already Alpine → NOT TRACKED ✗

Result:
  - theme store defined but not initialized
  - Console errors when buttons clicked
```

## Solution

### Code Changes
**File**: `transformer/stores.go`

1. **Added pattern to detect Alpine store references**:
   ```go
   var alpineStorePattern = regexp.MustCompile(`\$store\.([a-zA-Z_][a-zA-Z0-9_]*)`)
   ```

2. **Added tracking function**:
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

3. **Updated attribute transformation to track before skipping**:
   ```go
   // CRITICAL FIX: Track Alpine store references before skipping
   if attr.IsAlpine && strings.Contains(attr.Value, "$store.") {
       trackAlpineStoreReferences(attr.Value)
       transformedAttributes = append(transformedAttributes, attr)
       continue
   }
   ```

## Test Coverage
**File**: `transformer/stores_alpine_tracking_test.go`

- `TestTrackAlpineStoreReferences` - Unit test for tracking function
- `TestTransformAttributesWithAlpineStores` - Integration test
- `TestAlpineStorePattern` - Regex pattern validation
- `TestBugFix_ThemeStoreNotTracked` - **Regression test for this exact bug**

All tests pass ✓

## Verification

### Before Fix
```bash
# Only 2 stores initialized
curl http://localhost:3333/store-components-demo | grep -c "Alpine.store"
# Output: 2

# Debug logs
[DEBUG] Referenced stores: [auth cart]       # theme missing!
[DEBUG] Final stores keys: [cart auth]       # theme missing!
```

### After Fix
```bash
# All 3 stores initialized
curl http://localhost:3333/store-components-demo | grep -c "Alpine.store"
# Output: 3

# Debug logs
[DEBUG] Referenced stores: [cart theme auth]  # All tracked!
[DEBUG] Final stores keys: [cart theme auth]  # All included!
```

## Impact
✅ Fixed theme store initialization
✅ Theme toggle buttons now work without errors
✅ Pattern applies to ANY Alpine directive with `$store.*` syntax
✅ No breaking changes to existing functionality
✅ Comprehensive test coverage prevents regression

## Files Modified
- `transformer/stores.go` - Core fix (3 additions)
- `transformer/stores_alpine_tracking_test.go` - Test coverage (new file)
- `.agent-os/specs/.../BUG_FIX_ALPINE_STORE_TRACKING.md` - Detailed documentation (new file)

## Cognitive Load Analysis
- New pattern: 4 points
- Tracking function: 5 points
- Integration update: 2 points
- **Total added**: 11 points
- **File total**: 72 points (within limits ✓)
- **All functions**: < 15 points ✓
