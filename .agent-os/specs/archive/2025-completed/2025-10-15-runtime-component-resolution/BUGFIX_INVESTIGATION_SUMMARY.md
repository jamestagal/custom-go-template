# Go Map Syntax Bug - Investigation Summary

**Date**: 2025-10-16  
**Status**: PARTIALLY FIXED  
**Remaining Work**: One code path still has issues

## Problem

Complex Go objects (maps, slices) passed as props appear with Go map syntax (`map[key:value]`) instead of proper JSON syntax (`{key: value}`) in HTML x-data attributes.

## Root Cause Analysis

The issue occurs in TWO separate code paths:

### Path 1: Build-Time Dynamic Components ✅ FIXED
**File**: `transformer/dynamic_component_by_name.go:661-710`
**Function**: `convertPropsMapToComponentProps()`

**Issue**: Line 662 used `fmt.Sprintf("%v", v)` for the `default:` case, which produces Go map syntax for complex objects.

**Fix Applied**: 
```go
default:
    // CRITICAL FIX: Use JSON encoding for complex types
    if isComplexObject(v) {
        jsonBytes, err := json.Marshal(v)
        if err != nil {
            log.Printf("ERROR marshaling prop %q: %v", name, err)
            valueStr = "null"
        } else {
            valueStr = string(jsonBytes)
            log.Printf("prop %q = complex object (JSON): %q", name, valueStr)
        }
    } else {
        valueStr = fmt.Sprintf("%v", v)
    }
```

**Test Result**: This path now works correctly - runtime wrappers have proper JSON syntax for complex props.

### Path 2: Regular Component Prop Transformation ❌ STILL BROKEN
**File**: `transformer/components.go:260-416`
**Function**: `formatComponentData()`

**Issue**: When components like pages.html receive complex objects as props (e.g., `content={content}`), these objects appear as strings with Go map syntax in the final x-data output.

**Example**:
```html
<!-- BROKEN OUTPUT -->
<div x-data="{ content: map[components:[map[fields:map[...]]]] }">

<!-- EXPECTED OUTPUT -->
<div x-data="{ content: {components: [{fields: {...}}]} }">
```

**Investigation Findings**:

1. **extractPropValue()** (line 657) correctly returns the map object from parentDataScope
2. **transformComponentProps()** (line 679) correctly stores the map in propScope
3. **formatComponentData()** (line 406-410) should handle maps via `default:` case calling `FormatGoValueToJS()`
4. **FormatGoValueToJS()** (alpine.go:416-450) has proper handling for `map[string]interface{}`

**Mystery**: The map is being converted to a STRING with Go map syntax BEFORE reaching `formatComponentData`, causing it to match `case string:` instead of `default:`.

**Hypothesis**: The stringification occurs in an intermediate step between prop resolution and dataScope formatting, possibly:
- During fence section processing
- When merging fence props with passed props
- When creating getter functions for reactive variables

## Test Results

### Working (Path 1):
```bash
$ curl -s http://localhost:3333/ | grep -o 'x-data="[^"]*content[^"]*"' | head -1
x-data="{allContent:{'pages/_defaults':{...},content:{components:[{fields:{...}}]}}"
```
✅ Proper JSON syntax

### Broken (Path 2):
```bash
$ curl -s http://localhost:3333/ | grep -o 'x-data="[^"]*content[^"]*"' | head -2 | tail -1
x-data="{ get allContent() { return map[pages/_defaults:map[...]] }, content: map[...] }"
```
❌ Go map syntax in getter function body AND in content prop

## Files Modified

1. ✅ `transformer/dynamic_component_by_name.go` - Fixed `convertPropsMapToComponentProps()`
   - Added `isComplexObject()` helper
   - Added JSON encoding for maps/slices in default case
   - Added comprehensive debug logging

## Next Steps for Complete Fix

1. **Add logging to transformComponent()** to trace where map→string conversion occurs
2. **Check fence section processing** for any fmt.Sprintf("%v") calls on map values  
3. **Check getter function creation** (lines 376-384 in formatComponentData)
4. **Verify dataScope merging logic** doesn't stringify values
5. **Consider adding defensive check** in `formatComponentData` case string: to detect and handle stringified maps

## Workaround

For now, avoid passing complex nested objects as props to components. Use spread props instead:

```html
<!-- Instead of this -->
<Component content={content} />

<!-- Use this -->
<Component {...content} />
```

## Confidence Score: 60%

- ✅ Build-time dynamic components fixed: +40%
- ❌ Regular component props still broken: -40%
- ✅ Root cause partially identified: +20%
- ❌ Complete solution not yet implemented: -20%

**Total**: 60% (requires additional work to reach 95%+)
