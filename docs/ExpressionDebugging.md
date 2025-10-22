# Expression Transformation Debugging

## Overview

The template engine supports **optional verbose logging** to help developers understand when expressions are resolved at build-time vs runtime. This is particularly useful for debugging unexpected runtime bindings when build-time resolution was expected.

## Enabling Debug Mode

Set the `DEBUG_EXPRESSIONS` environment variable to `true`:

```bash
DEBUG_EXPRESSIONS=true go run cmd/server/main.go
```

Or for testing:

```bash
DEBUG_EXPRESSIONS=true go test ./transformer -v
```

## What Gets Logged

When enabled, the system logs detailed information about expression transformation decisions:

### Build-Time Resolution

When an expression can be resolved at build time:

```
[EXPR-DEBUG] Attribute 'content' expression '{description}' → BUILD-TIME
[EXPR-DEBUG]   ↳ Resolved value: "A powerful template engine"
```

### Runtime Binding

When an expression requires runtime Alpine.js binding:

```
[EXPR-DEBUG] Attribute 'class' expression '{type}' → RUNTIME
[EXPR-DEBUG]   ↳ Generated: :class="type"
```

With reason:

```
[EXPR-DEBUG] Expression '{items}' → RUNTIME: Variable not in dataScope
[EXPR-DEBUG] Expression '{count + 10}' → RUNTIME: Complex expression (not a simple variable)
[EXPR-DEBUG] Expression '{component}' → RUNTIME: Variable is nil (loop variable marker)
[EXPR-DEBUG] Expression '{user.orders}' → RUNTIME: Complex type map[string]interface {} (needs runtime evaluation)
```

### Store Expressions

For Alpine.js store expressions:

```
[EXPR-DEBUG] Attribute 'theme' expression '{$theme.mode}' → RUNTIME (store)
[EXPR-DEBUG]   ↳ Generated: :theme="$store.theme.mode"
```

### Mixed Content

For attributes with mixed static and dynamic content:

```
[EXPR-DEBUG] Attribute 'class' mixed content (static + expressions) → RUNTIME
[EXPR-DEBUG]   ↳ Generated: :class="'notification notification-' + type"
[EXPR-DEBUG]   ↳ Expression parts: 3
```

## When Expressions Are Build-Time vs Runtime

### Build-Time Resolution ✅

Expressions are resolved at build time when:
- The expression is a simple variable name (no dots, operators, or brackets)
- The variable exists in the dataScope (from fence section or props)
- The variable has a primitive value (string, number, boolean)
- The variable is NOT a loop variable (value is not nil)

**Examples:**
```html
---
let title = "Home Page"
let count = 42
let enabled = true
---

<!-- All build-time: values inserted statically -->
<h1 content="{title}">Home Page</h1>
<div data-count="{count}">42</div>
<input checked="{enabled}">
```

**Output:**
```html
<h1 content="Home Page">Home Page</h1>
<div data-count="42">42</div>
<input checked="true">
```

### Runtime Binding ⚙️

Expressions require runtime binding when:
- Complex expression (contains operators, dots, brackets)
- Variable not in dataScope
- Loop variable (marked as nil in dataScope)
- Complex type (object, array) - needs Alpine.js evaluation
- Store expression (requires `$store.*` runtime access)

**Examples:**
```html
---
let count = 42
prop type = "info"  // Has default, but might come from parent at runtime
---

{for item in items}
  <!-- RUNTIME: Loop variables -->
  <div name="{item.name}">

  <!-- RUNTIME: Complex expression -->
  <div data-value="{count + 10}">

  <!-- RUNTIME: Store expression -->
  <div theme="{$theme.mode}">
{/for}

<!-- RUNTIME: Variable not in dataScope -->
<div class="{dynamicClass}">
```

**Output:**
```html
<div :name="item.name">
<div :data-value="count + 10">
<div :theme="$store.theme.mode">
<div :class="dynamicClass">
```

## Use Cases

### 1. Debugging Unexpected Runtime Bindings

**Problem:** You expect `{description}` to be interpolated at build time, but it's creating a runtime binding.

**Debug Output:**
```
[EXPR-DEBUG] Expression '{description}' → RUNTIME: Variable not in dataScope
```

**Solution:** Check that `description` is declared in the fence section:
```html
---
let description = "My content"  // Add this
---
```

### 2. Understanding Loop Variable Behavior

**Problem:** Why is `{component.name}` runtime but `{layoutName}` is build-time?

**Debug Output:**
```
[EXPR-DEBUG] Expression '{component.name}' → RUNTIME: Complex expression (not a simple variable)
[EXPR-DEBUG] Expression '{layoutName}' → BUILD-TIME
[EXPR-DEBUG]   ↳ Resolved value: "Hero"
```

**Explanation:** `component.name` is a property access (complex), while `layoutName` is a simple variable that's resolvable.

### 3. Diagnosing Performance Issues

**Problem:** Page seems slow, want to see which expressions are causing runtime work.

**Action:** Enable debugging and look for:
- Many RUNTIME bindings that could be BUILD-TIME
- Complex expressions that could be simplified
- Store accesses that could be cached

## Performance Impact

**When Disabled (default):** Zero performance impact. The debug logging code is a simple conditional check that evaluates to false.

**When Enabled:** Minimal impact. Logging only occurs during template transformation (build/server start), not during runtime page serving.

## Example Demo

Run the included demo:

```bash
cd /path/to/custom_go_template
DEBUG_EXPRESSIONS=true go run cmd/test_expression_debug/main.go
```

This demonstrates build-time vs runtime decisions for various expression types.

## Implementation Details

### Location

The debugging system is implemented in `/transformer/stores.go`:

- `debugExpressions` - Environment variable flag (line ~18)
- `logExpressionDebug()` - Logging helper function (line ~22)
- Debug logs in `TryResolveBuildTimeValue()` (line ~206-237)
- Debug logs in `transformAttributesWithStores()` (line ~488-671)

### Architecture

```
Environment Variable → debugExpressions flag → logExpressionDebug() → log.Printf()
```

The system integrates seamlessly with existing transformation logic without affecting the transformation pipeline itself.

## Future Enhancements

Potential improvements:

1. **Verbose Levels**: `DEBUG_EXPRESSIONS=verbose` for even more detail
2. **Statistics Summary**: Count of build-time vs runtime decisions per template
3. **Suggestions**: Auto-suggest moving expressions to fence section for build-time resolution
4. **Performance Metrics**: Time spent on expression resolution

## Related Documentation

- [Build-Time Loop Expansion](../specs/2025-10-19-build-time-loop-expansion/SPECIFICATION.md)
- [Runtime Component Resolution](../specs/2025-10-15-runtime-component-resolution/SPECIFICATION.md)
- [Expression Transformation](./RecursiveComponentTranformationChecklist.md)
