# Quick Unblock Implementation Plan

**Date**: 2025-10-16 22:00 UTC
**Goal**: Fix component registry syntax errors to unblock component rendering
**Approach**: Improved regex with identifier-level prefixing + skip list (Option 1)
**Estimated Time**: 2-4 hours
**MANDATORY: Use go-backend agent for all Go implementation**
---

## Current Blocker

**Error**:
```
SyntaxError: Unexpected token '(' (at component-registry.js:1793:83)
```

**Problematic Code**:
```javascript
${props.(start * 1) + index + 1}  // ❌ INVALID - ( immediately after props.
```

**Root Cause**: Regex matches entire expression `(start * 1)` and prefixes it with `props.`, creating invalid syntax.

---

## Solution: Identifier-Level Prefixing

Instead of prefixing the **entire expression**, prefix **each identifier** within the expression.

### Example Conversions

```javascript
// Input: {(start * 1) + index + 1}
// Current (BROKEN): ${props.(start * 1) + index + 1}
// Fixed (CORRECT):  ${(props.start * 1) + index + 1}

// Input: {count + 1}
// Current (BROKEN): ${props.count + 1}  // Actually works for simple cases
// Fixed (CORRECT):  ${props.count + 1}  // Same result

// Input: {user.name}
// Current (WORKS): ${props.user.name}
// Fixed (CORRECT): ${props.user.name}  // Same result

// Input: {items[0]}
// Current (WORKS): ${props.items[0]}
// Fixed (CORRECT): ${props.items[0]}  // Same result

// Input: {getName()}
// Current (WORKS): ${props.getName()}
// Fixed (CORRECT): ${props.getName()}  // Same result
```

### Skip List

These identifiers should **NOT** get `props.` prefix:

**Loop Variables**:
- `index`, `item`, `todo`, `component`, `value`, `key`

**Alpine.js Built-ins**:
- `$store`, `$el`, `$refs`, `$watch`, `$dispatch`, `$nextTick`

**JavaScript Built-ins**:
- `window`, `document`, `console`, `Math`, `Date`, `JSON`

---

## Implementation Steps

### Step 1: Update `convertAttributeExpressions()` Function

**File**: `builder/registry_generator.go`

**Current Implementation** (Lines ~30-40):
```go
var expressionPattern = regexp.MustCompile(`\{([a-zA-Z_$][\w.$]*(?:\[[^\]]+\])?(?:\([^)]*\))?)\}`)

func convertAttributeExpressions(attrValue string) string {
    return expressionPattern.ReplaceAllString(attrValue, "${props.$1}")
}
```

**New Implementation**:
```go
// Skip list for identifiers that should NOT get props. prefix
var skipIdentifiers = map[string]bool{
    // Loop variables (common in x-for)
    "index":     true,
    "item":      true,
    "todo":      true,
    "component": true,
    "value":     true,
    "key":       true,

    // Alpine.js built-ins (magic properties)
    "$store":    true,
    "$el":       true,
    "$refs":     true,
    "$watch":    true,
    "$dispatch": true,
    "$nextTick": true,
    "$data":     true,
    "$root":     true,

    // JavaScript built-ins
    "window":   true,
    "document": true,
    "console":  true,
    "Math":     true,
    "Date":     true,
    "JSON":     true,
    "Array":    true,
    "Object":   true,
    "String":   true,
    "Number":   true,
    "Boolean":  true,
}

// Match template expressions: {anything}
var expressionPattern = regexp.MustCompile(`\{([^{}]+)\}`)

// Match JavaScript identifiers (but not property access, arrays, or calls)
var identifierPattern = regexp.MustCompile(`\b([a-zA-Z_$][\w]*)\b`)

func convertAttributeExpressions(attrValue string) string {
    // Find all {expression} patterns
    return expressionPattern.ReplaceAllStringFunc(attrValue, func(match string) string {
        // Extract the expression without braces
        expr := match[1 : len(match)-1] // Remove { and }

        // Check if this is an Alpine object literal: { key: value }
        // Simple heuristic: if it contains : and no surrounding { }, it's likely an object literal
        // We want to preserve these but convert expressions inside them
        if isAlpineObjectLiteral(expr) {
            // Process object literal carefully
            return "${" + convertObjectLiteralExpressions(expr) + "}"
        }

        // For other expressions, prefix each identifier with props.
        converted := identifierPattern.ReplaceAllStringFunc(expr, func(id string) string {
            // Skip if in skip list
            if skipIdentifiers[id] {
                return id
            }
            // Skip if already prefixed with props.
            // (Check if preceded by "props." in the original expr)
            // This is a simple check; more robust would track position
            if strings.Contains(expr, "props."+id) {
                return id
            }
            return "props." + id
        })

        return "${" + converted + "}"
    })
}

// isAlpineObjectLiteral checks if expression looks like an Alpine object literal
// Examples: "{ count: 0, items: [] }" or "{ count: {count}, message: '{message}' }"
func isAlpineObjectLiteral(expr string) bool {
    trimmed := strings.TrimSpace(expr)
    // Must start with { and end with }
    if !strings.HasPrefix(trimmed, "{") || !strings.HasSuffix(trimmed, "}") {
        return false
    }
    // Must contain : (key: value syntax)
    return strings.Contains(trimmed, ":")
}

// convertObjectLiteralExpressions processes object literals to convert nested expressions
// Example: "{ count: {count}, message: '{message}' }" → "{ count: ${props.count}, message: '${props.message}' }"
func convertObjectLiteralExpressions(objLiteral string) string {
    // This is a simplified version; full implementation would need proper JS parsing
    // For now, we'll use a conservative approach:
    // 1. Find {identifier} patterns that are NOT part of the object literal braces
    // 2. Convert them to ${props.identifier}

    // Match {identifier} patterns but NOT at the start/end (those are object literal braces)
    nestedExprPattern := regexp.MustCompile(`(\{)([a-zA-Z_$][\w]*)(\})`)

    return nestedExprPattern.ReplaceAllStringFunc(objLiteral, func(match string) string {
        // Extract identifier
        id := match[1 : len(match)-1] // Remove { and }

        // Skip if in skip list
        if skipIdentifiers[id] {
            return match
        }

        // Convert to ${props.identifier}
        return "${props." + id + "}"
    })
}
```

**Cognitive Load**: 12 (regex patterns: 5, skip list: 2, object literal handling: 5)

---

### Step 2: Skip Alpine Directive Headers

**Problem**: We don't want to convert expressions in Alpine directive **names/headers**, only in their **values**.

**Example**:
```html
<!-- x-for HEADER should stay as-is -->
<template x-for="(todo, index) in todos">
  <!-- Content expressions SHOULD be converted -->
  <td>{index}</td>  <!-- But index should NOT get props. prefix because it's a loop var -->
</template>
```

**Solution**: The attribute value is what we're converting, not the attribute name. This should already be handled by the attribute parsing, but we need to ensure expressions inside attribute values are handled correctly.

**Implementation**: Already handled by `convertAttributeExpressions()` which only processes attribute **values**, not names.

---

### Step 3: Add Tests

**File**: `builder/registry_generator_test.go`

```go
func TestConvertAttributeExpressions_ComplexExpressions(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {
            name:     "Parenthesized expression",
            input:    "{(start * 1) + index + 1}",
            expected: "${(props.start * 1) + index + 1}", // index skipped (loop var)
        },
        {
            name:     "Multiple operators",
            input:    "{count + total - discount}",
            expected: "${props.count + props.total - props.discount}",
        },
        {
            name:     "Alpine store access",
            input:    "{$store.cart.count}",
            expected: "${$store.cart.count}", // $store skipped (Alpine built-in)
        },
        {
            name:     "Mixed loop var and prop",
            input:    "{start + index}",
            expected: "${props.start + index}", // index skipped
        },
        {
            name:     "Already prefixed",
            input:    "{props.count}",
            expected: "${props.count}", // Don't double-prefix
        },
        {
            name:     "Alpine object literal with expressions",
            input:    "{ count: {count}, message: '{message}' }",
            expected: "{ count: ${props.count}, message: '${props.message}' }",
        },
        {
            name:     "Alpine object literal plain",
            input:    "{ count: 0, items: [] }",
            expected: "{ count: 0, items: [] }", // No expressions to convert
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := convertAttributeExpressions(tt.input)
            if result != tt.expected {
                t.Errorf("convertAttributeExpressions() = %v, want %v", result, tt.expected)
            }
        })
    }
}

func TestSkipIdentifiers(t *testing.T) {
    // Test that skip list identifiers are not prefixed
    tests := []struct {
        input    string
        expected string
    }{
        {"{index}", "${index}"},                    // Loop var
        {"{item.name}", "${item.name}"},            // Loop var with property
        {"{$store.cart}", "${$store.cart}"},        // Alpine built-in
        {"{Math.floor(x)}", "${Math.floor(props.x)}"}, // JS built-in function, prop var
        {"{window.location}", "${window.location}"}, // JS built-in
    }

    for _, tt := range tests {
        result := convertAttributeExpressions(tt.input)
        if result != tt.expected {
            t.Errorf("convertAttributeExpressions(%q) = %v, want %v", tt.input, result, tt.expected)
        }
    }
}
```

---

### Step 4: Regenerate Registry and Validate

```bash
# Kill existing server
pkill -f "go run cmd/server/main.go"

# Rebuild and regenerate registry
cd /Users/benjaminwaller/Projects/Jim\ Fisk/custom_go_template
go run cmd/server/main.go > /tmp/server.log 2>&1 &

# Wait for registry generation
sleep 5

# Validate JavaScript syntax
node -c static/js/component-registry.js

# Check for invalid patterns
echo "Checking for invalid \${props.( patterns:"
grep -n '\${props\.(' static/js/component-registry.js

echo "Checking for invalid \${props.\[ patterns:"
grep -n '\${props\.\[' static/js/component-registry.js

echo "Checking for invalid \${props.\$ patterns (Alpine built-ins):"
grep -n '\${props\.\$' static/js/component-registry.js
```

**Expected Results**:
- `node -c` passes with no errors
- All three grep commands return **no matches**

---

### Step 5: Test Component Rendering

```bash
# Test homepage
curl -s http://localhost:3333/ | grep -i "welcome to artistitch"

# Should output: (match found)
```

**Expected**: "Welcome to Artistitch" appears in output, indicating hero2436 component rendered.

---

## Edge Cases to Handle

### 1. Nested Expressions
```javascript
// Input: {items.filter(x => x.count > {threshold})}
// Expected: ${props.items.filter(x => x.count > ${props.threshold})}
```

**Status**: Current implementation may struggle with nested `{}`. Consider this a known limitation for now.

**Workaround**: Avoid nested expressions in templates. Use fence section functions instead.

---

### 2. String Interpolation in Alpine
```javascript
// Input: { message: '{greeting}, {name}!' }
// Expected: { message: '${props.greeting}, ${props.name}!' }
```

**Status**: Should be handled by `convertObjectLiteralExpressions()`.

---

### 3. Ternary Operators
```javascript
// Input: {isActive ? 'active' : 'inactive'}
// Expected: ${props.isActive ? 'active' : 'inactive'}
```

**Status**: Should work with identifier-level prefixing.

---

## Validation Checklist

After implementation, verify:

- [ ] `node -c static/js/component-registry.js` passes
- [ ] No `${props.(` patterns in registry
- [ ] No `${props.[` patterns in registry
- [ ] No `${props.$` patterns in registry (Alpine built-ins)
- [ ] Simple expressions work: `{count}` → `${props.count}`
- [ ] Complex expressions work: `{(start * 1) + index}` → `${(props.start * 1) + index}`
- [ ] Loop variables not prefixed: `index` stays as `index`
- [ ] Alpine built-ins not prefixed: `$store` stays as `$store`
- [ ] Homepage loads without console errors
- [ ] Components render: "Welcome to Artistitch" visible
- [ ] All tests pass: `go test ./builder -v`

---

## Rollback Plan

If this fix causes regressions:

1. Revert changes to `builder/registry_generator.go`
2. Restore original `convertAttributeExpressions()` function
3. Document specific failing cases
4. Proceed with proper AST-level fix (Option 2) instead

---

## Next Steps After Quick Fix

Once components are rendering:

1. **Document limitations** of regex-based approach
2. **Plan proper AST-level fix** (Option 2) for next sprint
3. **Add regression tests** for all edge cases
4. **Fix 'this.' prefix bug** (separate investigation)

---

## Estimated Timeline

| Task | Time | Status |
|------|------|--------|
| Implement improved regex | 1-2 hours | Pending |
| Add skip list | 30 min | Pending |
| Write tests | 30 min | Pending |
| Test and validate | 30 min | Pending |
| Documentation | 30 min | Pending |
| **Total** | **2-4 hours** | **Pending** |

---

## Success Criteria

✅ Components from `/content/pages/_index.json` render on homepage
✅ No console errors about component registry
✅ JavaScript syntax validates
✅ All tests pass
✅ No regression in existing components

---

**Created**: 2025-10-16 22:00 UTC
**Status**: 📋 READY TO IMPLEMENT
**Next Action**: Implement improved regex in `builder/registry_generator.go`
