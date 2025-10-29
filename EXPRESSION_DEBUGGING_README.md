# Expression Debugging Feature - Implementation Summary

## What Was Implemented

Added **optional verbose logging** to help developers understand when expressions are resolved at build-time vs runtime. This addresses the common use case where most expressions should be build-time (from Plenti JSON content), but unexpected runtime bindings occur.

## Quick Start

```bash
# Enable debugging
DEBUG_EXPRESSIONS=true go run cmd/server/main.go

# Run demo
DEBUG_EXPRESSIONS=true go run cmd/test_expression_debug/main.go
```

## Example Output

```
[EXPR-DEBUG] Attribute 'content' expression '{description}' → BUILD-TIME
[EXPR-DEBUG]   ↳ Resolved value: "A powerful template engine"

[EXPR-DEBUG] Attribute 'class' expression '{type}' → RUNTIME
[EXPR-DEBUG]   ↳ Generated: :class="type"

[EXPR-DEBUG] Expression '{count + 10}' → RUNTIME: Complex expression (not a simple variable)
[EXPR-DEBUG] Attribute 'data-value' expression '{count + 10}' → RUNTIME
[EXPR-DEBUG]   ↳ Generated: :data-value="count + 10"
```

## What Gets Logged

### Build-Time Resolution
- Decision: BUILD-TIME
- Resolved value (the actual value interpolated)

### Runtime Binding
- Decision: RUNTIME
- Reason why (not in scope, complex expression, nil value, complex type, store)
- Generated binding syntax (`:attribute="expression"` or `x-attribute="expression"`)

### Store Expressions
- Marked as RUNTIME (store)
- Generated Alpine.js store binding

### Mixed Content
- Number of expression parts
- Generated combined expression

## Files Modified

### `/transformer/stores.go`
Added:
1. `debugExpressions` global (line ~18) - reads `DEBUG_EXPRESSIONS` env var
2. `logExpressionDebug()` function (line ~22) - conditional logging helper
3. Debug logs in `TryResolveBuildTimeValue()` (lines 206-237)
4. Debug logs in `transformAttributesWithStores()` (lines 488-671)
   - Build-time resolution logging
   - Runtime binding logging with reasons
   - Store expression logging
   - Mixed content logging

## Files Created

### `/cmd/test_expression_debug/main.go`
Demo program showing various expression types and their transformation decisions.

### `/docs/ExpressionDebugging.md`
Comprehensive documentation:
- How to enable debugging
- What gets logged and why
- Build-time vs runtime decision criteria
- Use cases and examples
- Performance impact

### `/examples/test_expression_debug.html`
Test template demonstrating various expression scenarios.

### `/test_expression_debug.sh`
Helper script to run server with debugging enabled.

## Key Design Decisions

### 1. Environment Variable Control
**Why:** Simple, standard, no code changes needed. Disabled by default (zero performance impact).

```go
var debugExpressions = os.Getenv("DEBUG_EXPRESSIONS") == "true"
```

### 2. Conditional Logging Helper
**Why:** Clean code, easy to use, consistent formatting.

```go
func logExpressionDebug(format string, args ...interface{}) {
    if debugExpressions {
        log.Printf("[EXPR-DEBUG] "+format, args...)
    }
}
```

### 3. Log at Decision Points
**Why:** Developers need to understand WHY a decision was made, not just WHAT happened.

Logged reasons:
- "Complex expression (not a simple variable)"
- "Variable not in dataScope"
- "Variable is nil (loop variable marker)"
- "Complex type %T (needs runtime evaluation)"

### 4. Show Generated Code
**Why:** Helps developers understand the transformation output.

```
[EXPR-DEBUG]   ↳ Generated: :class="'notification-' + type"
```

## Success Criteria Met

✅ **Environment variable control** - Simple on/off toggle
✅ **Build-time decisions logged** - Shows resolved values
✅ **Runtime decisions logged** - Shows reasons and generated code
✅ **Disabled by default** - Zero performance impact when off
✅ **Clear, actionable output** - Easy to understand and diagnose issues
✅ **Comprehensive documentation** - Examples, use cases, and implementation details

## Example Use Cases

### 1. Debugging Missing Build-Time Resolution

**Problem:** Expected `{description}` to be build-time, but getting runtime binding.

**Solution:**
```bash
DEBUG_EXPRESSIONS=true go run cmd/server/main.go
```

**Output:**
```
[EXPR-DEBUG] Expression '{description}' → RUNTIME: Variable not in dataScope
```

**Fix:** Add `let description = "..."` to fence section.

### 2. Understanding Loop Behavior

**Problem:** Why is `{component.name}` runtime?

**Output:**
```
[EXPR-DEBUG] Expression '{component.name}' → RUNTIME: Complex expression (not a simple variable)
```

**Explanation:** Property access requires runtime evaluation (expected for dynamic components).

### 3. Optimizing Performance

**Problem:** Page has many runtime bindings, want to see which could be build-time.

**Action:** Enable debugging, review output, move resolvable values to fence section.

## Testing

Run the demo to see the feature in action:

```bash
cd /Users/benjaminwaller/Projects/Jim Fisk/custom_go_template
DEBUG_EXPRESSIONS=true go run cmd/test_expression_debug/main.go
```

Expected output shows clear BUILD-TIME vs RUNTIME decisions for various expression types.

## Integration

The debugging system integrates cleanly with existing code:
- No changes to AST structures
- No changes to transformation logic
- Simple conditional checks (negligible performance impact)
- Log messages only during transformation (not during runtime serving)

## Future Enhancements

Potential improvements:
1. **Verbose levels**: `DEBUG_EXPRESSIONS=verbose` for even more detail
2. **Statistics**: Summary of build-time vs runtime ratio per template
3. **Suggestions**: Auto-suggest optimizations (e.g., "Move '{description}' to fence section for build-time resolution")
4. **Performance metrics**: Time spent on expression resolution

## Documentation

Full documentation: [docs/ExpressionDebugging.md](docs/ExpressionDebugging.md)

Topics covered:
- Enabling debug mode
- Understanding output
- Build-time vs runtime criteria
- Use cases and examples
- Performance impact
- Implementation details
