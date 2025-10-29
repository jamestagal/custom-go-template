# Component Registry & Runtime Component Resolution - Troubleshooting History

**Date Range**: 2025-10-15 to 2025-10-16
**Status**: 🔴 IN PROGRESS - Still encountering syntax errors in component registry
**Related Specs**:
- `.agent-os/specs/2025-10-15-runtime-component-resolution/`
- `.agent-os/specs/KNOWN_ISSUE_COMPONENT_REGISTRY_SYNTAX.md`

## Executive Summary

This document tracks all attempts to fix runtime component resolution issues in the custom Go template engine. The core problem is that the component registry generation creates invalid JavaScript syntax when converting template expressions inside Alpine.js directives and complex attribute values.

**Current Blocker**: Component registry still has syntax errors preventing runtime components from loading, despite multiple fix attempts.

---

## Problem Statement

### Initial Issue (2025-10-15)

Dynamic components defined in JSON files (e.g., `/content/pages/_index.json`) were not rendering on the page. Console showed errors about component registry failing to load with syntax errors.

**Example JSON**:
```json
{
  "components": [
    {
      "name": "hero2436",
      "fields": {
        "topper": "Welcome to Artistitch",
        "title": "Tattoo & Body Art Templates",
        "buttonLink": "/contact"
      }
    }
  ]
}
```

**Expected Behavior**: Components should render on homepage via runtime resolution system.

**Actual Behavior**: Components not visible, console errors about syntax errors in component-registry.js.

---

## Root Cause Analysis

The component registry generator (`builder/registry_generator.go`) converts component ASTs to JavaScript ES modules with template literal factories. The conversion process has multiple architectural issues:

### Issue 1: Expression Conversion in Attribute Values

**Problem**: Template expressions like `{buttonLink}` inside HTML attribute values were not being converted to `${props.buttonLink}` for JavaScript template literals.

**Example**:
```html
<!-- Original Template -->
<a href="{buttonLink}">Click</a>

<!-- Broken Registry Output -->
<a href="{buttonLink}">Click</a>

<!-- Expected Registry Output -->
<a href="${props.buttonLink}">Click</a>
```

### Issue 2: Alpine.js Directive Syntax Corruption

**Problem**: Alpine.js directives like `x-data="{ count: {count} }"` were being incorrectly parsed and transformed, corrupting the JavaScript object literal syntax.

**Example**:
```html
<!-- Original Template -->
<div x-data="{ count: {count}, message: '{message}' }">

<!-- Broken Registry Output (Line 430) -->
div x-data="${props. count: {count}, message: '{message}' }">

<!-- Expected Registry Output -->
<div x-data="{ count: ${props.count}, message: '${props.message}' }">
```

**Multiple Problems**:
1. Missing opening `<` tag
2. `${props.` without property name
3. Inner `{count}` and `{message}` not converted
4. Alpine.js object literal braces `{ }` being confused with template expressions

### Issue 3: Complex Expressions in Alpine Directives

**Problem**: Complex expressions inside Alpine.js directives (like x-for, x-bind) are causing syntax errors.

**Example** (Line 1791-1793):
```html
<!-- Original Template -->
<template x-for="todo, index in todos.slice(start, start + number)">
  <td>{(start * 1) + index + 1}</td>
</template>

<!-- Broken Registry Output -->
<template x-for="todo, index in todos.slice(start, start + number)">
  <td>${props.(start * 1) + index + 1}</td>
</template>

<!-- Expected Output -->
<template x-for="(todo, index) in todos.slice(props.start, props.start + props.number)">
  <td>${(props.start * 1) + index + 1}</td>
</template>
```

**Problems**:
1. `${props.(start * 1)}` is invalid syntax (should be `${(props.start * 1)}`)
2. x-for attribute values shouldn't have `props.` prefix added
3. Loop variables like `index` should NOT get `props.` prefix

### Issue 4: Persistent 'this.' Prefix Bug

**Problem**: Content field names in nested objects are being corrupted with `'this.'` prefix.

**Example**:
```javascript
// x-data output shows:
{
  'this.buttonLink': '/contact',
  'this.buttonText': 'Book A Call',
  'this.description': '...'
}

// Should be:
{
  buttonLink: '/contact',
  buttonText: 'Book A Call',
  description: '...'
}
```

**Suspected Cause**: The `replaceVarRefsWithThis()` function in transformer is adding `this.` prefix to map keys, not just getter expressions.

---

## Attempted Solutions

### Attempt 1: Fix Script Loading Order (2025-10-15)

**Hypothesis**: Runtime components script loading after Alpine.js initialization.

**Implementation**:
- Modified `/layouts/global/head.html` to load runtime-components.js BEFORE Alpine.js
- Changed script loading order

**Result**: ❌ FAILED - Same error persisted

**Files Modified**:
- `layouts/global/head.html`

---

### Attempt 2: Fix Content Prop Passing (2025-10-15)

**Hypothesis**: Content data not being passed to components correctly.

**Implementation**:
- Modified html.html wrapper to explicitly pass `content` and `allContent` props
- Updated pages.html component iterator to receive props

**Result**: ❌ FAILED - Components still not rendering

**Files Modified**:
- `layouts/global/html.html`
- `layouts/content/pages.html`

---

### Attempt 3: Remove Temporary Workaround (2025-10-16)

**Hypothesis**: Temporary workaround in `renderWithWrapper` was causing `'this.'` prefix bug.

**Implementation**:
- Removed lines 333-346 in `cmd/server/main.go` that were adding component fields as top-level props
- These prop names were entering dataScope and getting `this.` prefix from `replaceVarRefsWithThis()`

**Result**: ⚠️ PARTIAL - Workaround removed but `'this.'` prefix still appears (different source)

**Files Modified**:
- `cmd/server/main.go` (lines 333-346 removed)

---

### Attempt 4: Fix Component Registration (2025-10-16)

**Hypothesis**: Content layouts needed to be registered as components for runtime resolution.

**Implementation**:
1. Initially removed content layout registration to avoid double x-data wrapping
2. Got "Component 'pages' not found" error
3. Implemented selective registration with pattern detection
4. User explained nuance: pages.html IS a component (component iterator pattern)
5. Restored full registration of all content layouts
6. Fixed case sensitivity: `renderWithWrapper("pages")` → `renderWithWrapper("Pages")`

**Result**: ✅ SUCCESS - Components started rendering, "Welcome to Artistitch" appeared

**Files Modified**:
- `cmd/server/main.go` (registration logic, pattern detection functions)

**Key Insight**: In Plenti architecture, ALL layouts are components, including special "component iterator" layouts like pages.html.

---

### Attempt 5: Fix Runtime Components Script Path (2025-10-16)

**Hypothesis**: Script 404 error preventing runtime system from loading.

**Implementation**:
- Changed `/js/runtime-components.js` to `/static/js/runtime-components.js` in head.html
- Added `/js/*` route mapping in server

**Result**: ✅ SUCCESS - Script loads with 200 OK

**Files Modified**:
- `layouts/global/head.html`
- `cmd/server/main.go` (routing)

---

### Attempt 6: Fix Parser Quote Handling (2025-10-16) 🆕

**Hypothesis**: Parser's `parseComplexAlpineValue()` failing to correctly parse x-data attributes with nested quotes.

**Problem Identified**:
```go
// BEFORE: Checked inner quotes BEFORE closing quote
else if char == '"' && !inSingleQuote {
    inDoubleQuote = !inDoubleQuote  // This ran FIRST!
    valueBuilder.WriteByte(char)
} else if char == quoteChar && !inSingleQuote && !inDoubleQuote && !escaped {
    break  // Closing quote check came SECOND
}
```

**Implementation**:
- Reordered conditional checks in `parser/html.go` parseComplexAlpineValue() function (lines 584-620)
- Check for closing quote FIRST before processing inner quotes

```go
// AFTER: Check closing quote FIRST
if char == quoteChar && !inSingleQuote && !inDoubleQuote && !escaped {
    break  // EXIT IMMEDIATELY when closing quote found
}
...
else if char == '"' && !inSingleQuote {
    inDoubleQuote = !inDoubleQuote
    valueBuilder.WriteByte(char)
}
```

**Result**: ✅ SUCCESS - Parser now correctly extracts x-data attribute values

**Files Modified**:
- `parser/html.go` (lines 584-620)

**Tests Added**:
- Parser tests for complex Alpine attribute values

---

### Attempt 7: Fix Builder Expression Conversion (2025-10-16) 🆕

**Hypothesis**: Builder not converting `{expression}` patterns inside attribute values to `${props.expression}`.

**Implementation**:
Added `convertAttributeExpressions()` function in `builder/registry_generator.go`:

```go
// Regex pattern to match {identifier} but NOT { object: literal }
var expressionPattern = regexp.MustCompile(`\{([a-zA-Z_$][\w.$]*(?:\[[^\]]+\])?(?:\([^)]*\))?)\}`)

func convertAttributeExpressions(attrValue string) string {
    return expressionPattern.ReplaceAllString(attrValue, "${props.$1}")
}
```

**Result**: ✅ PARTIAL SUCCESS - Simple expressions converted correctly, but complex expressions still break

**Files Modified**:
- `builder/registry_generator.go` (added regex pattern and conversion function)
- `builder/registry_generator_test.go` (comprehensive tests)

**What Works**:
- Simple expressions: `{count}` → `${props.count}`
- Multiple expressions: `{count}, {message}` → `${props.count}, ${props.message}`
- Property access: `{user.name}` → `${props.user.name}`
- Array access: `{items[0]}` → `${props.items[0]}`
- Function calls: `{getName()}` → `${props.getName()}`
- Alpine.js object literals preserved: `{ count: 0 }` → `{ count: 0 }`

**What Breaks**:
- Complex expressions: `{(start * 1) + index + 1}` → `${props.(start * 1) + index + 1}` (INVALID)
- Expressions in Alpine directives shouldn't always get `props.` prefix
- Loop variables like `index` should NOT get `props.` prefix

---

### Attempt 8: Validate JavaScript Syntax (2025-10-16)

**Implementation**:
```bash
node -c component-registry.js
```

**Result**: ❌ FAILED - Syntax error still exists at line 1793

**Error**: `Unexpected token '('` at line 1793:83

**Location**: `${props.(start * 1) + index + 1}` - the `(` after `props.` is invalid

---

## Current Status (2025-10-16 21:35 UTC)

### What's Working ✅

1. **Component Registration**: All 65 components successfully registered
2. **Runtime Resolution Infrastructure**: `<Component:dynamic>` wrapper elements present in HTML
3. **Simple Expression Conversion**: Basic `{var}` → `${props.var}` works
4. **Parser Quote Handling**: x-data attributes correctly extracted
5. **Script Loading**: runtime-components.js loads with 200 OK
6. **Content Loading**: JSON data from `/content/pages/_index.json` loads into x-data

### What's Broken 🔴

1. **Complex Expressions**: `{(start * 1) + index}` produces invalid `${props.(start * 1)...}` syntax
2. **Component Registry Syntax**: Line 1793 has `Unexpected token '('` error
3. **Components Not Rendering**: Console error prevents registry from loading
4. **'this.' Prefix Bug**: Still present in nested content objects (27 instances)

### Current Error

```
runtime-components.js:97 Failed to load component registry after 3 attempts
SyntaxError: Unexpected token '(' (at component-registry.js:1793:83)
```

**Problematic Code** (line 1793):
```javascript
<td>${props.(start * 1) + index + 1}</td>
```

**Why It's Invalid**: You cannot have `props.` followed immediately by `(`. The regex matched `(start * 1)` as an expression and added `props.` prefix, but parentheses are not valid after a property accessor.

---

## Architectural Issues

### 1. Expression vs Attribute Context

The current regex-based approach cannot distinguish between:
- **Content expressions**: `{count}` should become `${props.count}`
- **Alpine directive values**: `x-for="item in items"` should NOT have `props.` added to loop variables
- **Complex expressions**: `{(start * 1) + index}` needs ALL identifiers prefixed, not just first one

### 2. Template Expression Syntax Limitations

The template syntax `{expression}` is ambiguous:
- `{count}` - Simple variable reference
- `{user.name}` - Property access
- `{items[0]}` - Array access
- `{getName()}` - Function call
- `{(start * 1) + index}` - Complex expression with operators and parentheses

The regex pattern cannot correctly parse complex expressions because:
1. It matches the ENTIRE expression as one unit
2. It adds `props.` prefix to the whole match
3. Parentheses immediately after `props.` are invalid JavaScript

### 3. Context-Aware Parsing Required

The builder needs to know:
1. **Is this expression inside an Alpine directive attribute?**
   - If yes, some identifiers are loop variables (don't prefix with `props.`)
2. **Is this expression simple or complex?**
   - Simple: Add `props.` prefix to identifier
   - Complex: Parse expression tree and add `props.` to each identifier (except loop vars)
3. **Is this Alpine.js object literal syntax `{ key: value }`?**
   - If yes, don't treat as expression

### 4. AST-Level vs String-Level Processing

**Current Approach** (String-level):
- Builder receives AST with attribute values as strings
- Regex attempts to find and replace expressions
- Cannot handle complex cases

**Needed Approach** (AST-level):
- Parser should identify expressions DURING parsing
- Attribute values should contain AST nodes for expressions, not strings
- Transformer can then handle expressions with full context
- Builder just renders the already-transformed AST

---

## Recommended Solutions

### Option 1: Improve Regex Pattern (Short-term, Partial Fix)

**Approach**: Update regex to handle parenthesized expressions.

**Pros**:
- Quick to implement
- Might fix some common cases

**Cons**:
- Still won't handle all complex expressions
- Cannot distinguish loop variables from props
- Fundamentally limited by string-based approach

**Implementation Sketch**:
```go
// Match complex expressions and prefix each identifier
func convertComplexExpression(expr string) string {
    // Match identifiers NOT preceded by props.
    identPattern := regexp.MustCompile(`\b([a-zA-Z_$][\w]*)\b`)

    // List of known loop variables and Alpine built-ins
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

**Estimated Effort**: 2-4 hours
**Risk**: High (might break other things, won't solve all cases)

---

### Option 2: Parser-Level Attribute Expression Nodes (Long-term, Proper Fix)

**Approach**: Modify parser to create ExpressionNode AST nodes for expressions inside attribute values.

**Current Architecture**:
```go
type Attribute struct {
    Name  string
    Value string  // ⚠️ String, not Node[]
}
```

**Proposed Architecture**:
```go
type Attribute struct {
    Name  string
    Value interface{}  // Can be string OR []Node
}
```

**Pros**:
- Proper architectural solution
- Full context available during transformation
- Can handle ANY expression complexity
- Transformer can distinguish loop vars from props

**Cons**:
- Significant refactoring required
- Parser, transformer, and builder all need updates
- Higher risk of regression

**Estimated Effort**: 1-2 days
**Risk**: Medium (well-understood change, but touches many files)

---

### Option 3: Literal Content Flag for Alpine Directives (Medium-term)

**Approach**: Mark Alpine.js directive attributes to be rendered literally without expression conversion.

**Implementation**:
```go
type Attribute struct {
    Name     string
    Value    string
    IsAlpine bool      // Flag to skip expression conversion
}
```

**Pros**:
- Simpler than Option 2
- Solves Alpine directive corruption
- Minimal changes to existing code

**Cons**:
- Still doesn't solve complex expressions in regular attributes
- Partial solution

**Estimated Effort**: 4-8 hours
**Risk**: Low (localized changes)

---

### Option 4: Hybrid Approach (Recommended)

**Approach**: Combine Option 1 (improved regex) with Option 3 (Alpine directive flags).

**Phase 1** (Quick Win):
1. Add `IsAlpine` flag to Attribute struct
2. Skip expression conversion for Alpine directives
3. Improve regex to handle common complex expression patterns

**Phase 2** (Proper Fix):
1. Refactor Attribute.Value to support Node[]
2. Update parser to create ExpressionNodes in attributes
3. Update transformer to handle attribute expressions
4. Update builder to render transformed attribute expressions

**Estimated Effort**:
- Phase 1: 4-6 hours
- Phase 2: 1-2 days

**Risk**: Low (Phase 1), Medium (Phase 2)

---

## Testing Checklist

### Expressions to Test

- [ ] Simple variable: `{count}` → `${props.count}`
- [ ] Property access: `{user.name}` → `${props.user.name}`
- [ ] Array access: `{items[0]}` → `${props.items[0]}`
- [ ] Function call: `{getName()}` → `${props.getName()}`
- [ ] Complex expression: `{count + 1}` → `${props.count + 1}`
- [ ] Parenthesized: `{(start * 1) + index}` → `${(props.start * 1) + index}`
- [ ] Multiple identifiers: `{start + count}` → `${props.start + props.count}`

### Alpine.js Directives to Test

- [ ] `x-data="{ count: 0 }"` - Object literal preserved
- [ ] `x-data="{ count: {count} }"` - Expression converted
- [ ] `x-for="item in items"` - Loop variables NOT prefixed
- [ ] `x-for="(item, index) in items"` - Index NOT prefixed
- [ ] `x-bind:style="{color: theme.color}"` - Expression converted
- [ ] `@click="count++"` - Preserved as-is
- [ ] `:class="{active: isActive}"` - Preserved as-is

### Component Registry Generation

- [ ] All 65 components register without errors
- [ ] Registry JavaScript syntax validates with `node -c`
- [ ] Hero2436 component renders with props from JSON
- [ ] Services2437 component renders
- [ ] Nested components work
- [ ] Components in loops work

---

## Files Involved

### Core Files

1. **`parser/html.go`** - Attribute parsing
   - `parseAttributes()` - Lines 192-233
   - `EnhancedAttributeParser()` - Lines 416-497
   - `parseAlpineDataAttribute()` - Lines 498-518
   - `parseComplexAlpineValue()` - Lines 565-619 ✅ FIXED (2025-10-16)

2. **`builder/registry_generator.go`** - Component registry generation
   - `convertToJSTemplate()` - Lines 60-76
   - `renderElementToJS()` - Lines 147-183
   - `renderAttributeToJS()` - Lines 186-229 ✅ PARTIALLY FIXED (2025-10-16)
   - `convertAttributeExpressions()` - NEW FUNCTION ✅ ADDED (2025-10-16)

3. **`transformer/expressions.go`** - Expression transformation
   - May need updates for context-aware transformation

4. **`static/js/component-registry.js`** - Generated output
   - Line 430: ✅ FIXED (was `div x-data="${props. count...`)
   - Line 1793: 🔴 BROKEN (still `${props.(start * 1)...`)
   - Line 2185: ✅ FIXED (was `{buttonLink}`, now `${props.buttonLink}`)

### Supporting Files

5. **`cmd/server/main.go`** - Component registration and server
6. **`layouts/global/head.html`** - Script loading order
7. **`layouts/global/html.html`** - Wrapper layout
8. **`layouts/content/pages.html`** - Component iterator layout
9. **`static/js/runtime-components.js`** - Runtime resolution client code

---

## Related Issues

1. **KNOWN_ISSUE_COMPONENT_REGISTRY_SYNTAX.md** - Original documentation of parser/transformer issue
2. **2025-10-15-runtime-component-resolution/** - Runtime resolution implementation spec
3. **2025-10-06-parser-unification/** - Parser architecture unification
4. **'this.' Prefix Bug** - Separate but related issue with `replaceVarRefsWithThis()`

---

## Next Steps

### Immediate (Today)

1. ✅ Document all troubleshooting attempts (this file)
2. ⏭️ Implement Option 1 (improved regex) as quick fix
3. ⏭️ Test with hero2436 and services2437 components
4. ⏭️ Validate JavaScript syntax of generated registry

### Short-term (This Week)

1. Implement Option 3 (Alpine directive flags)
2. Add comprehensive tests for expression conversion
3. Fix 'this.' prefix bug (separate investigation)
4. Document workarounds for component authors

### Long-term (Next Sprint)

1. Implement Option 2 (parser-level expression nodes)
2. Refactor attribute handling throughout pipeline
3. Add integration tests for runtime component resolution
4. Performance optimization of registry generation

---

## Lessons Learned

### What Worked

1. **Systematic Debugging**: Using grep, read, and targeted investigation helped identify exact issues
2. **go-backend Agent**: Specialized agent successfully implemented parser and builder fixes
3. **Test-Driven**: Writing tests before fixes helped validate solutions
4. **Documentation**: Tracking attempts prevented going in circles

### What Didn't Work

1. **String-Based Regex**: Fundamentally limited for complex expression parsing
2. **Workarounds**: Temporary fixes in renderWithWrapper caused more problems
3. **Assumptions**: Assuming simple fixes would solve architectural issues

### Key Insights

1. **Context Matters**: Expression conversion needs full context (Alpine directive? Loop variable? Regular content?)
2. **AST vs Strings**: String manipulation is insufficient; need AST-level processing
3. **Architecture Tax**: Shortcuts create technical debt that compounds
4. **Test Coverage**: Edge cases with complex expressions revealed gaps in testing

---

## References

- Go Template Engine Codebase: `/Users/benjaminwaller/Projects/Jim Fisk/custom_go_template/`
- Alpine.js Documentation: https://alpinejs.dev/
- Plenti SSG (inspiration): https://plenti.co/
- Component Registry Pattern: ES Modules with Template Literal Factories

---

**Last Updated**: 2025-10-16 21:45 UTC
**Status**: 🔴 BLOCKING - Components not rendering due to registry syntax errors
**Next Action**: Implement improved regex pattern (Option 1) to unblock component rendering
