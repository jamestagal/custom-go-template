# Loop Variable Fix - Quick Reference

## The Problem

Template functions in the component registry were trying to evaluate loop variables as JavaScript variables during function execution:

```javascript
// ❌ BROKEN
<template x-for="text in props.textGroup">
  <p>${text}</p>  // ReferenceError: text is not defined
</template>
```

## The Solution

Loop variables are now converted to Alpine.js directives:

```javascript
// ✅ FIXED
<template x-for="text in props.textGroup">
  <p><span x-text="text"></span></p>  // Alpine evaluates at DOM level
</template>
```

## When This Fix Applies

### Text Content (ExpressionNode)

**Input Template**:
```html
{for text in textGroup}
  <p>{text}</p>
{/for}
```

**Output (Component Registry)**:
```javascript
<template x-for="text in props.textGroup">
  <p><span x-text="text"></span></p>
</template>
```

### Attributes

**Input Template**:
```html
{for card in cards}
  <img src="{card.icon.src}" alt="{card.icon.alt}" />
{/for}
```

**Output (Component Registry)**:
```javascript
<template x-for="card in props.cards">
  <img :src="card.icon.src" :alt="card.icon.alt" />
</template>
```

### Nested Property Access

**Input Template**:
```html
{for item in items}
  <h3>{item.details.name}</h3>
  <img src="{item.image.url}" />
{/for}
```

**Output (Component Registry)**:
```javascript
<template x-for="item in props.items">
  <h3><span x-text="item.details.name"></span></h3>
  <img :src="item.image.url" />
</template>
```

## What Doesn't Change

### Regular Props (No Loop Context)

**Input Template**:
```html
<h1>{title}</h1>
<p>{description}</p>
```

**Output (Component Registry)**:
```javascript
<h1>${props.title}</h1>
<p>${props.description}</p>
```

**Why**: These are evaluated when the template function runs, not at DOM level, so template literals work fine.

## Implementation Details

### Key Functions

1. **`expressionReferencesLoopVar(expr, loopVars)`**
   - Detects if an expression references any loop variable
   - Uses regex with word boundaries to avoid false positives

2. **`attributeReferencesLoopVar(attrValue, loopVars)`**
   - Checks if attribute value contains expressions referencing loop vars
   - Scans all `{expression}` patterns in the attribute value

3. **`extractExpressionFromBraces(attrValue)`**
   - Extracts expression content from `{...}` braces
   - Used for converting to Alpine binding syntax

### Context Tracking

Loop variables are tracked in `RenderContext.loopVars`:

```go
type RenderContext struct {
    insideLiteral    bool
    insideAlpineAttr bool
    loopVars         map[string]bool  // "text": true, "card": true, etc.
}
```

When entering a loop, variables are added:
```go
loopCtx.loopVars[loop.Value] = true      // Iterator variable
loopCtx.loopVars[loop.Iterator] = true   // Index variable (if present)
```

## Testing

Run loop variable tests:
```bash
go test ./builder -v -run TestLoopVariable
```

All builder tests:
```bash
go test ./builder -v
```

## Files Modified

- `builder/registry_generator.go` - Core implementation
- `builder/loop_var_test.go` - New test file
- `builder/registry_generator_test.go` - Updated signatures
- `builder/debug_spread_test.go` - Updated signatures
- `builder/spread_test.go` - Updated signatures

## Common Patterns

### Simple Loop Variable
```
{text} → <span x-text="text"></span>
```

### Property Access
```
{card.title} → <span x-text="card.title"></span>
```

### Attribute Binding
```
src="{card.image}" → :src="card.image"
```

### Complex Expression
```
{item.name + ' - ' + item.price} → <span x-text="item.name + ' - ' + item.price"></span>
```

## Edge Cases Handled

✅ Partial identifier matches (e.g., "card" doesn't match "discard")
✅ Nested property access (`card.icon.src`)
✅ Multiple expressions in one attribute
✅ Mixed loop and non-loop variables in same component
✅ Empty loop vars map

## When to Use This Fix

**Use Alpine directives** (x-text, :binding) when:
- Expression references a loop variable (item, text, card, etc.)
- Inside an `x-for` template
- Variable only exists at DOM runtime

**Use template literals** (${props.var}) when:
- Expression references props passed to template function
- Outside any loop context
- Variable exists when template function executes

---

**Status**: Production Ready ✅
**Version**: 2025-10-22
**See Also**: `COMPLETION_SUMMARY.md` for full details
