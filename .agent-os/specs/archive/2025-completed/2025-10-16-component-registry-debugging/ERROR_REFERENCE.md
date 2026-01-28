# Component Registry Error Reference

Quick reference for debugging component registry syntax errors.

---

## Current Error (2025-10-16)

### Error Message
```
runtime-components.js:97 Failed to load component registry after 3 attempts
SyntaxError: Unexpected token '(' (at component-registry.js:1793:83)
```

### Location
**File**: `static/js/component-registry.js`
**Line**: 1793
**Column**: 83

### Source Code (Line 1791-1793)
```javascript
<template x-for="todo, index in todos.slice(start, start + number)">
    <tr>
      <td style="text-align: center; color: #6b7280; font-weight: 500;">${props.(start * 1) + index + 1}</td>
                                                                            ^
                                                                            Column 83
```

### Why It's Invalid
JavaScript syntax does NOT allow `(` immediately after a property accessor:
```javascript
${props.(start * 1) + index + 1}  // ❌ INVALID - syntax error
${(props.start * 1) + index + 1}  // ✅ VALID - parentheses around full expr
```

### Original Template
The component likely had:
```html
<td>{(start * 1) + index + 1}</td>
```

### Conversion Process (BROKEN)
1. **Parser**: Extracts attribute value as string: `"{(start * 1) + index + 1}"`
2. **Builder Regex**: Matches `(start * 1)` as identifier ❌
3. **Builder Replace**: Adds `props.` prefix: `${props.(start * 1) + index + 1}` ❌
4. **Result**: Invalid JavaScript syntax

### Expected Conversion
1. **Parser**: Identifies expression components
2. **Transformer**: Prefixes each identifier: `start` → `props.start`, `index` → keep as-is (loop var)
3. **Builder**: Renders: `${(props.start * 1) + index + 1}` ✅

---

## Previous Errors (FIXED)

### Error 1: Missing Closing Brace (Fixed 2025-10-16)

**Error Message**:
```
SyntaxError: Missing } in template expression (at component-registry.js:430:17)
```

**Location**: Line 430

**Broken Code**:
```javascript
div x-data="${props. count: {count}, message: '{message}' }">
```

**Issues**:
1. Missing opening `<` tag
2. `${props.` without property name
3. Inner `{count}` not converted
4. Closing quote position wrong

**Root Cause**: Parser's `parseComplexAlpineValue()` checked inner quotes before closing quote.

**Fix**: Reordered conditionals in `parser/html.go` lines 584-620 to check closing quote FIRST.

**Result**: ✅ FIXED - x-data attributes now parse correctly

---

### Error 2: Expression Not Converted (Fixed 2025-10-16)

**Issue**: `{buttonLink}` in attributes remained unconverted.

**Example**:
```html
<!-- Before: -->
<a href="{buttonLink}">

<!-- After: -->
<a href="${props.buttonLink}">
```

**Root Cause**: Builder had no expression conversion logic for attribute values.

**Fix**: Added `convertAttributeExpressions()` function with regex pattern.

**Result**: ✅ FIXED - Simple expressions now convert correctly

---

## Error Patterns

### Pattern 1: Complex Expressions
**Symptom**: `${props.(expr)}` or `${props.[expr]}`

**Examples**:
- `${props.(start * 1)}` ❌
- `${props.[items[0]]}` ❌

**Cause**: Regex matches entire expression as single identifier.

**Fix Needed**: Parse expression tree and prefix each identifier.

---

### Pattern 2: Loop Variables with Props Prefix
**Symptom**: `${props.index}` inside x-for loop

**Example**:
```html
<template x-for="(item, index) in items">
  <td>${props.index}</td>  <!-- ❌ WRONG - index is loop var -->
</template>
```

**Cause**: No scope context during conversion.

**Fix Needed**: Track loop scope and skip prefixing loop variables.

---

### Pattern 3: Alpine Built-ins with Props Prefix
**Symptom**: `${props.$store}`, `${props.$el}`, etc.

**Example**:
```html
<div x-text="${props.$store.cart.count}">  <!-- ❌ WRONG -->
<div x-text="$store.cart.count">           <!-- ✅ CORRECT -->
```

**Cause**: No recognition of Alpine.js magic properties.

**Fix Needed**: Skip list for Alpine built-ins (`$store`, `$el`, `$refs`, etc.).

---

### Pattern 4: Alpine Object Literal Syntax
**Symptom**: `{ key: value }` treated as expression

**Example**:
```html
<!-- Template: -->
<div x-data="{ count: 0, items: [] }">

<!-- Should Stay: -->
<div x-data="{ count: 0, items: [] }">

<!-- Should NOT Become: -->
<div x-data="${props.{ count: 0, items: [] }}">  <!-- ❌ -->
```

**Cause**: Regex cannot distinguish object literals from template expressions.

**Fix Needed**: Context-aware parsing (Alpine directive attributes).

---

## Debugging Commands

### Find Syntax Errors
```bash
node -c static/js/component-registry.js
```

### Check Specific Line
```bash
sed -n '1790,1795p' static/js/component-registry.js
```

### Search for Invalid Props Patterns
```bash
# Find ${props.( patterns (invalid syntax)
grep -n '\${props\.(' static/js/component-registry.js

# Find ${props.[ patterns (invalid syntax)
grep -n '\${props\.\[' static/js/component-registry.js

# Find ${props.$ patterns (Alpine built-ins incorrectly prefixed)
grep -n '\${props\.\$' static/js/component-registry.js
```

### Test Component Registry Loading
```bash
curl -s http://localhost:3333/js/component-registry.js | node -c
```

---

## Common Component Templates Causing Issues

### 1. Todo List Component
**Issue**: Complex expression with parentheses

**Template**:
```html
<template x-for="todo, index in todos">
  <td>{(start * 1) + index + 1}</td>
</template>
```

**Current Output**: `${props.(start * 1) + index + 1}` ❌

**Expected Output**: `${(props.start * 1) + index + 1}` ✅

---

### 2. Store Access Component
**Issue**: Alpine `$store` gets props prefix

**Template**:
```html
<div>{$store.cart.count}</div>
```

**Current Output**: `<div>${props.$store.cart.count}</div>` ❌

**Expected Output**: `<div>${$store.cart.count}</div>` ✅

---

### 3. Complex Alpine Directives
**Issue**: x-data object literal corrupted

**Template**:
```html
<div x-data="{ count: {initialCount}, items: [] }">
```

**Current Output**: Varies depending on quote handling

**Expected Output**: `<div x-data="{ count: ${props.initialCount}, items: [] }">`

---

## Resolution Strategies

### Strategy 1: Improved Regex (Quick)
**Time**: 2-4 hours
**Scope**: Fix common patterns
**Risk**: Low
**Completeness**: Partial

**Implementation**:
- Match individual identifiers, not full expressions
- Maintain skip list for loop variables and Alpine built-ins
- Handle parentheses and brackets correctly

---

### Strategy 2: AST-Level Processing (Proper)
**Time**: 1-2 days
**Scope**: Fix all expression handling
**Risk**: Medium
**Completeness**: Full

**Implementation**:
- Refactor `Attribute.Value` to support `[]Node`
- Parser creates `ExpressionNode` objects in attributes
- Transformer handles with full scope context
- Builder renders pre-transformed expressions

---

### Strategy 3: Alpine Directive Flags (Intermediate)
**Time**: 4-8 hours
**Scope**: Fix Alpine directive issues
**Risk**: Low
**Completeness**: Partial

**Implementation**:
- Add `IsAlpine` flag to `Attribute` struct
- Skip expression conversion for Alpine directives
- Preserve directive values as literals

---

## Quick Fixes to Try

### Fix 1: Add Skip List
```go
var skipIdentifiers = map[string]bool{
    // Loop variables
    "index": true,
    "item": true,
    "todo": true,
    "component": true,

    // Alpine built-ins
    "$store": true,
    "$el": true,
    "$refs": true,
    "$watch": true,
    "$dispatch": true,

    // JavaScript built-ins
    "window": true,
    "document": true,
    "console": true,
}
```

### Fix 2: Better Regex Pattern
```go
// Current (BROKEN):
var expressionPattern = regexp.MustCompile(`\{([a-zA-Z_$][\w.$]*(?:\[[^\]]+\])?(?:\([^)]*\))?)\}`)

// Proposed (BETTER):
var expressionPattern = regexp.MustCompile(`\{([^{}]+)\}`)

// Then process matched content to prefix identifiers:
func convertExpression(expr string) string {
    identPattern := regexp.MustCompile(`\b([a-zA-Z_$][\w]*)\b`)
    return identPattern.ReplaceAllStringFunc(expr, func(id string) string {
        if skipIdentifiers[id] {
            return id
        }
        return "props." + id
    })
}
```

---

## Testing Checklist

After implementing fix, verify:

- [ ] No syntax errors: `node -c static/js/component-registry.js`
- [ ] Simple expressions work: `{count}` → `${props.count}`
- [ ] Complex expressions work: `{(start * 1) + index}` → `${(props.start * 1) + index}`
- [ ] Loop variables NOT prefixed: `index` stays as `index`
- [ ] Alpine built-ins NOT prefixed: `$store` stays as `$store`
- [ ] Object literals preserved: `{ count: 0 }` stays as `{ count: 0 }`
- [ ] Components render on homepage
- [ ] No console errors about component registry

---

**Last Updated**: 2025-10-16 21:50 UTC
**Status**: 🔴 BLOCKED - Syntax error at line 1793 prevents component loading
