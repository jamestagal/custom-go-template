# Bug: Attribute Expressions Transform Incorrectly

**Date Discovered**: 2025-10-07
**Severity**: High
**Status**: Open
**Affected Component**: Transformer (Expression handling in attributes)

## Summary

Expressions inside HTML attributes (e.g., `{!inStock ? 'disabled' : ''}`) are incorrectly transformed to `<span x-text="...">` wrappers instead of Alpine.js attribute bindings like `:disabled="..."`.

## Problem

When a template contains an expression within an HTML attribute position:

```html
<button class="add-to-cart" {!inStock ? 'disabled' : ''}>
  Add to Cart
</button>
```

The transformer converts it to:

```html
<button class="add-to-cart" <span x-text="!inStock ? 'disabled' : ''"></span>>
  Add to Cart
</button>
```

This is **invalid HTML** and causes the expression to evaluate in the wrong scope (global instead of component scope), leading to `ReferenceError: inStock is not defined`.

## Expected Behavior

The expression should be transformed to an Alpine.js attribute binding:

**Option 1** (Conditional attribute):
```html
<button class="add-to-cart" :disabled="!inStock">
  Add to Cart
</button>
```

**Option 2** (Dynamic attribute):
```html
<button class="add-to-cart" x-bind:disabled="!inStock">
  Add to Cart
</button>
```

**Option 3** (String attribute if needed):
```html
<button class="add-to-cart" :disabled="!inStock ? 'disabled' : null">
  Add to Cart
</button>
```

## Root Cause

The transformer's expression parser doesn't distinguish between:
1. **Text content expressions**: `<p>{value}</p>` → `<p><span x-text="value"></span></p>` ✅
2. **Attribute value expressions**: `<button {expr}>` → Should use Alpine bindings ❌

All expressions are currently treated as text content and wrapped in `<span x-text>`.

## Impact

- ✅ **Text expressions work**: `<p>{user.name}</p>` correctly becomes `<p><span x-text="user.name"></span></p>`
- ❌ **Attribute expressions broken**: Cannot use conditional attributes, dynamic classes, computed props, etc.
- ❌ **Ternary operators in attributes fail**: Common pattern `{condition ? 'value' : ''}` unusable
- ❌ **Component prop scoping breaks**: Expressions evaluate in wrong scope

## Workaround

Use conditional rendering instead of conditional attributes:

```html
<!-- BEFORE (broken) -->
<button {!inStock ? 'disabled' : ''}>Add to Cart</button>

<!-- AFTER (working) -->
{if inStock}
  <button>Add to Cart</button>
{else}
  <button disabled>Sold Out</button>
{/if}
```

## Test Case

**File**: `examples/components/ProductCard.html` (line 127, now fixed with workaround)

**Original Code**:
```html
<button class="add-to-cart" {!inStock ? 'disabled' : ''}>
  {if inStock}
    Add to Cart
  {else}
    Sold Out
  {/if}
</button>
```

**Console Error**:
```
Uncaught ReferenceError: inStock is not defined
[Alpine] inStock @ VM3397:3
```

**Current Workaround**:
```html
{if inStock}
  <button class="add-to-cart">Add to Cart</button>
{else}
  <button class="add-to-cart" disabled>Sold Out</button>
{/if}
```

## Fix Required

### Location
`transformer/expressions.go` or similar expression transformation logic

### Required Changes

1. **Detect attribute context**: When parsing expressions, determine if the expression is inside an HTML attribute position
2. **Use Alpine bindings for attributes**: Transform to `:attribute="expression"` instead of wrapping in `<span x-text>`
3. **Handle different attribute types**:
   - Boolean attributes: `:disabled="expr"`
   - String attributes: `:class="expr"`, `:href="expr"`
   - Event handlers: `@click="expr"`
4. **Preserve scope**: Ensure expressions evaluate in correct Alpine.js scope

### Pseudo-code

```go
func transformExpression(expr string, context ExpressionContext) string {
    if context.IsAttributePosition {
        // Extract attribute name from context
        attrName := context.AttributeName

        // Transform to Alpine binding
        return fmt.Sprintf(":%s=\"%s\"", attrName, expr)
    }

    // Text content - use x-text wrapper
    return fmt.Sprintf("<span x-text=\"%s\"></span>", expr)
}
```

## Related Issues

- **Bug #1**: Server manually builds x-data (different issue, but related to scope)
- **Bug #2**: Component props in loops don't evaluate (scope evaluation issue)

## References

- Alpine.js docs on `x-bind`: https://alpinejs.dev/directives/bind
- Test file: `examples/pages/comprehensive-simple.html`
- Component: `examples/components/ProductCard.html`

## Priority Justification

**High Priority** because:
1. Common pattern in templates (conditional attributes, dynamic classes)
2. Blocks testing of realistic component patterns
3. Required for Plenti compatibility (Svelte uses similar patterns)
4. Current workaround is verbose and limits expressiveness
