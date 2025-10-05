# Conditional Wrapper Bug Fix

## Problem Description

When a conditional (`{if}...{/if}`) was placed inside a loop, sibling nodes following the conditional were incorrectly being nested inside the conditional's template wrapper instead of being siblings at the loop level.

### Example

**Input Template:**
```html
{for animal in animals}
{if animal === 'cat'}
Hi cat!
{/if}
<div>likes: {animal}s</div>
<br>
{/for}
```

**Expected Output:**
```html
<template x-for="...">
  <div>
    <template x-if="animal === 'cat'">
      Hi cat!
    </template>
    <div>likes: cats</div>  <!-- Sibling -->
    <br>                      <!-- Sibling -->
  </div>
</template>
```

**Buggy Output (Before Fix):**
```html
<template x-for="...">
  <div>
    <template x-if="animal === 'cat'">
      Hi cat!
      <div>likes: cats</div>  <!-- WRONG: Inside conditional -->
      <br>                      <!-- WRONG: Inside conditional -->
    </template>
  </div>
</template>
```

## Root Cause

The bug was in `transformer/template_nesting.go` in the `ensureProperNesting` function (lines 118-124).

When processing transformed nodes:
1. It encountered a `<template x-if>` element
2. Set `currentTemplate = element` to collect subsequent content
3. Incorrectly assumed ALL following non-template nodes should be nested inside the conditional
4. Added sibling elements to the conditional's children via the `contentBuffer`

**The issue:** `transformConditional` already correctly wraps conditional content. The `ensureProperNesting` function was incorrectly collecting additional siblings into conditionals that already had their proper content.

## The Fix

Added a check in `ensureProperNesting` at lines 118-125:

```go
// FIX: Conditional templates already have their content wrapped by transformConditional
// We should NOT collect additional siblings into them
if isConditionalTemplate && len(element.Children) > 0 {
    // Conditional already has its content, don't collect more
    result = append(result, element)
    currentTemplate = nil
    continue
}
```

**Logic:** If a conditional template already has children (meaning `transformConditional` properly wrapped its content), add it to the result immediately and don't set it as `currentTemplate` to collect more content.

## Verification

Created a test case that verifies:
1. The loop wrapper has the correct number of children (conditional + siblings)
2. The conditional template has only 1 child (its own content)
3. The sibling `<div>` and `<br>` elements are at the loop level, not inside the conditional

**Result:** ✅ FIX VERIFIED - Conditional and siblings are correctly at the same level

## Impact

- Fixes conditional-in-loop nesting issues
- Preserves existing functionality for conditionals outside loops
- Does not affect loop-only transformations
- Maintains compatibility with wrapper div logic in `transformConditional`

## Related Files

- `transformer/template_nesting.go` - The fix location
- `transformer/conditionals.go` - Creates properly wrapped conditionals
- `transformer/loops.go` - Processes loop content
