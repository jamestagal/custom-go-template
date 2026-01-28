# Spec Requirements Document

> Spec: Loop Rendering & Integration Fix
> Created: 2025-10-02

## Overview

Investigate and fix loop rendering issues in the transformer, ensuring that loops work correctly in isolation, within conditionals, nested within other loops, and when containing components.

## User Stories

### Template Developer Using Loops

As a template developer, I want to iterate over collections using `{for}` syntax, so that I can render dynamic lists of items with proper Alpine.js reactivity.

When I write:
```
{for item in items}
  <div>{item.name}</div>
{/for}
```

The transformer should generate:
```html
<template x-for="item in items">
  <div><span x-text="item.name"></span></div>
</template>
```

And Alpine.js should properly iterate and render each item.

### Developer Using Nested Structures

As a developer, I want to use loops within conditionals and vice versa, so that I can create complex template logic.

Complex structures should work:
```
{if hasItems}
  {for item in items}
    <div>{item.name}</div>
  {/for}
{/if}
```

And:
```
{for category in categories}
  {if category.items.length > 0}
    <div>{category.name}</div>
  {/if}
{/for}
```

### Component Developer

As a component developer, I want to use components within loops, so that I can render dynamic lists of components.

```
{for product in products}
  <ProductCard product={product} />
{/for}
```

Should correctly pass each product to its own ProductCard instance.

## Spec Scope

1. **Loop Transformation Review** - Analyze `transformer/loops.go` to identify issues with loop transformation logic.

2. **Scope Handling** - Ensure loop iterator variables are properly scoped and don't leak to parent scope.

3. **Nested Loops** - Fix any issues with loops nested within loops (ensure unique iterator names don't conflict).

4. **Loops in Conditionals** - Ensure loops within conditional blocks transform correctly.

5. **Components in Loops** - Verify that component transformation works correctly when components are inside loop bodies (may already be fixed by Spec 1).

## Out of Scope

- Component transformation (covered in Spec 1)
- Function expression handling (covered in Spec 2)
- Parser changes to loop syntax
- New loop features (like loop.index, loop.first, etc.)
- Performance optimizations

## Expected Deliverable

1. Tests `TestAlpineIntegration/loop_rendering` and `TestAlpineIntegration/nested_conditionals_and_loops` pass successfully.

2. Loops correctly render as `<template x-for>` elements with proper Alpine.js syntax.

3. Iterator variables are properly scoped and don't cause variable conflicts in nested structures.
