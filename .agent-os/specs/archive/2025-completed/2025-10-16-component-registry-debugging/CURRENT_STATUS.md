# Runtime Component Resolution - Current Status

**Date**: 2025-10-16 21:45 UTC
**Status**: 🔴 BLOCKED - Component registry syntax errors preventing component rendering

---

## Quick Summary

Components from JSON files are not rendering because the component registry has JavaScript syntax errors. We've fixed several issues but complex expressions still generate invalid syntax.

**Current Error**:
```
SyntaxError: Unexpected token '(' (at component-registry.js:1793:83)
```

**Problematic Code**:
```javascript
${props.(start * 1) + index + 1}  // ❌ INVALID - Can't have ( after props.
```

---

## What's Fixed ✅

### 1. Parser Quote Handling (2025-10-16)
**File**: `parser/html.go` lines 584-620

**Problem**: Parser checked inner quotes before closing quote, causing premature termination.

**Fix**: Reordered conditionals to check closing quote FIRST.

**Result**: x-data attributes now correctly extracted as complete strings.

---

### 2. Builder Expression Conversion (2025-10-16)
**File**: `builder/registry_generator.go`

**Problem**: Template expressions `{var}` in attribute values weren't converted to `${props.var}`.

**Fix**: Added regex pattern and `convertAttributeExpressions()` function.

**Result**: Simple expressions now convert correctly:
- `{count}` → `${props.count}` ✅
- `{user.name}` → `${props.user.name}` ✅
- `{items[0]}` → `${props.items[0]}` ✅

---

### 3. Component Registration (2025-10-16)
**File**: `cmd/server/main.go`

**Problem**: Content layouts weren't registered as components, causing "Component 'pages' not found" error.

**Fix**:
- Restored full registration of all content layouts
- Fixed case sensitivity (pages → Pages)

**Result**: Component lookup now works, runtime wrapper elements present in HTML.

---

### 4. Script Loading (2025-10-16)
**File**: `layouts/global/head.html`

**Problem**: Runtime components script returning 404.

**Fix**: Changed path from `/js/runtime-components.js` to `/static/js/runtime-components.js`.

**Result**: Script loads with 200 OK.

---

## What's Broken 🔴

### 1. Complex Expressions (BLOCKING)
**File**: `builder/registry_generator.go` - `convertAttributeExpressions()`

**Problem**: Regex pattern adds `props.` prefix to entire expression, including parentheses.

**Example**:
```javascript
// Template:
{(start * 1) + index + 1}

// Current Output:
${props.(start * 1) + index + 1}  // ❌ INVALID

// Expected Output:
${(props.start * 1) + index + 1}  // ✅ VALID
```

**Root Cause**: String-based regex cannot parse expression trees. It matches `(start * 1)` as a single identifier and prefixes with `props.`.

**Impact**: Component registry has syntax error at line 1793, preventing ALL components from loading.

---

### 2. Alpine Directive Variables (NOT YET ADDRESSED)
**File**: `builder/registry_generator.go`

**Problem**: Loop variables and Alpine built-ins get `props.` prefix incorrectly.

**Example**:
```html
<!-- Template: -->
<template x-for="(todo, index) in todos">
  <td>{index}</td>
</template>

<!-- Current Output: -->
<template x-for="(todo, index) in todos">
  <td>${props.index}</td>  <!-- ❌ WRONG - index is loop variable -->
</template>

<!-- Expected Output: -->
<template x-for="(todo, index) in todos">
  <td>${index}</td>  <!-- ✅ CORRECT - no props prefix -->
</template>
```

**Root Cause**: Conversion function has no context about Alpine directives or loop scopes.

**Impact**: Components with loops will have incorrect variable references.

---

### 3. 'this.' Prefix Bug (SEPARATE ISSUE)
**File**: Unknown - suspected in transformer

**Problem**: Content field names get corrupted with `'this.'` prefix in nested objects.

**Example**:
```javascript
// Current Output:
{
  'this.buttonLink': '/contact',
  'this.buttonText': 'Book A Call'
}

// Expected Output:
{
  buttonLink: '/contact',
  buttonText: 'Book A Call'
}
```

**Root Cause**: Unknown - suspected `replaceVarRefsWithThis()` function affecting map keys.

**Impact**: Component props may not be accessible correctly.

---

## Architectural Root Causes

### 1. String-Level Processing
**Current**: Builder receives attributes as strings, uses regex to find/replace expressions.

**Problem**: Cannot parse complex expressions or understand context.

**Solution**: Need AST-level processing with ExpressionNode objects in attributes.

---

### 2. No Context Awareness
**Current**: Expression converter doesn't know if identifier is a prop, loop variable, or Alpine built-in.

**Problem**: All identifiers treated the same, causing incorrect prefixing.

**Solution**: Need scope tracking during transformation (similar to existing dataScope system).

---

### 3. Ambiguous Template Syntax
**Current**: `{expression}` syntax is context-dependent but parser treats all the same.

**Problem**: Cannot distinguish between:
- Content expression: `<p>{count}</p>`
- Attribute expression: `<a href="{link}">`
- Alpine directive: `<div x-data="{ count: {count} }">`

**Solution**: Parser needs to create different AST node types based on context.

---

## Recommended Next Steps

### Option A: Quick Fix (2-4 hours)
**Approach**: Improve regex to handle common complex expression patterns.

**Implementation**:
```go
func convertComplexExpression(expr string) string {
    // Match individual identifiers, prefix each with props.
    identPattern := regexp.MustCompile(`\b([a-zA-Z_$][\w]*)\b`)

    // Skip loop variables and Alpine built-ins
    skipList := map[string]bool{
        "index": true, "item": true, "todo": true,
        "$store": true, "$el": true, "$refs": true,
    }

    return identPattern.ReplaceAllStringFunc(expr, func(match string) string {
        if skipList[match] {
            return match
        }
        return "props." + match
    })
}
```

**Pros**: Quick to implement, might unblock component rendering.

**Cons**: Still won't handle all cases, needs manual skip list maintenance.

---

### Option B: Proper Fix (1-2 days)
**Approach**: Refactor attributes to support AST nodes instead of strings.

**Implementation**:
1. Change `Attribute.Value` from `string` to `interface{}` (can be string OR `[]Node`)
2. Update parser to create `ExpressionNode` objects for `{expr}` patterns in attributes
3. Update transformer to handle attribute expression nodes with scope context
4. Update builder to render already-transformed attribute expressions

**Pros**: Solves all expression handling issues properly.

**Cons**: Significant refactoring, higher risk of regression.

---

## Files to Review

### Parser
- `parser/html.go` - Attribute parsing (parseAttributes, EnhancedAttributeParser)
- `parser/expressions.go` - Expression parsing

### Builder
- `builder/registry_generator.go` - Component registry generation
- `builder/registry_generator_test.go` - Tests

### Transformer
- `transformer/expressions.go` - Expression transformation
- `transformer/scope.go` - Scope tracking

### Runtime
- `static/js/runtime-components.js` - Client-side component loading
- `static/js/component-registry.js` - Generated registry (has syntax errors)

---

## Testing Commands

### Validate Registry Syntax
```bash
node -c static/js/component-registry.js
```

### Check Server Logs
```bash
tail -f /tmp/server.log
```

### Test Homepage
```bash
curl -s http://localhost:3333/ | grep "Welcome to Artistitch"
```

### Regenerate Registry
```bash
# Kill server
pkill -f "go run cmd/server/main.go"

# Restart (auto-generates registry)
cd /Users/benjaminwaller/Projects/Jim\ Fisk/custom_go_template
go run cmd/server/main.go
```

---

## Decision Point

**Question**: Should we implement the quick fix (Option A) or the proper fix (Option B)?

**Considerations**:
- Quick fix unblocks component rendering TODAY
- Proper fix prevents future issues and handles all cases
- Quick fix will need to be replaced eventually (technical debt)
- Proper fix is well-understood and lower risk than it appears

**Recommendation**: Implement Option A today to unblock progress, plan Option B for next sprint.

---

**Last Updated**: 2025-10-16 21:45 UTC
**Next Action**: Implement improved regex pattern for complex expressions
