# Build-Time Loop Conditionals - Best Practices

**Last Updated:** 2026-01-29

This guide explains how conditionals work inside build-time expanded loops and how to avoid common runtime errors.

---

## The Problem

When loops expand at build-time, **simple conditionals** are expanded, but **complex conditionals** may fall back to runtime Alpine templates, causing "variable is not defined" errors.

### Example of the Issue

```html
---
export let allContent
---

{for post in allContent}
  {if post.fields.textItems && post.fields.textItems.length > 0}
    <p>{post.fields.textItems[0].paragraph}</p>
  {/if}
{/for}
```

**What you expect**: Loop expands, conditional resolves for each post

**What actually happens**:
1. Loop expands ✅ (3 `<li>` items created)
2. Conditional does NOT expand ❌ (becomes `<template x-if="post.fields...">`)
3. Runtime error: `post is not defined` 💥

**Generated Output (WRONG)**:
```html
<li>
  <h3>Post Title</h3>
  <template x-if="post.fields.textItems && post.fields.textItems.length > 0">
    <p x-text="post.fields.textItems[0].paragraph"></p>
  </template>
</li>
```

**Error**: `post` doesn't exist at runtime - it was only available during build-time loop expansion!

---

## Why This Happens

### Build-Time Loop Expansion Phases

1. **Loop Expansion** - `{for post in allContent}` expands to 3 items
2. **Variable Resolution** - Each iteration has `post` with actual data
3. **Conditional Evaluation** - Checks if conditional can be resolved:
   - ✅ **Simple**: `{if post.fields.title}` → Resolved
   - ✅ **Comparison**: `{if post.type === "news"}` → Resolved
   - ❌ **Complex**: `{if post.fields.items && post.fields.items.length > 0}` → Falls back to runtime

### Why Complex Conditionals Fall Back

The `&&` operator creates a compound expression that the build-time evaluator may not fully resolve:

- `post.fields.textItems` → Can resolve ✅
- `post.fields.textItems.length` → Can resolve ✅
- `post.fields.textItems.length > 0` → Can resolve ✅
- `A && B` → Compound expression, may fall back to runtime ❌

---

## Solutions

### Solution 1: Remove the Conditional (If Data Always Exists)

**Best when**: All items in your collection have the required field

```html
{for post in allContent}
  <!-- Just access the data directly -->
  <p>{post.fields.textItems[0].paragraph}</p>
{/for}
```

**Why it works**: No conditional = no runtime Alpine template

---

### Solution 2: Use Nested Simple Conditionals

**Best when**: Some items might not have the field

```html
{for post in allContent}
  {if post.fields.textItems}
    <p>{post.fields.textItems[0].paragraph}</p>
  {else}
    <p>No content available</p>
  {/if}
{/for}
```

**Why it works**: Simple existence check resolves at build-time

**Don't do this** (too complex):
```html
{if post.fields.textItems && post.fields.textItems.length > 0}
```

**Do this instead** (nested simple):
```html
{if post.fields.textItems}
  {if post.fields.textItems.length > 0}
    <p>{post.fields.textItems[0].paragraph}</p>
  {/if}
{/if}
```

---

### Solution 3: Use Fallback Operators

**Best when**: You want a default value

```html
{for post in allContent}
  <p>{post.fields.textItems[0].paragraph || 'No content available'}</p>
{/for}
```

**Why it works**: Fallback operator `||` is resolved at build-time

---

### Solution 4: Pre-filter the Collection

**Best when**: You only want items that meet certain criteria

```html
---
export let allContent

// Filter in fence section before rendering
let newsWithText = allContent.filter(post =>
  post.type === "news" &&
  post.fields.textItems &&
  post.fields.textItems.length > 0
)
---

{for post in newsWithText}
  <!-- Now you know post.fields.textItems[0] exists -->
  <p>{post.fields.textItems[0].paragraph}</p>
{/for}
```

**Note**: This only works if the filter runs at build-time. If `allContent` is passed as a prop, JavaScript filtering in the fence section may not work as expected.

---

## Best Practices

### ✅ DO: Use Simple Conditionals

```html
<!-- Simple existence check -->
{if post.fields.title}
  <h3>{post.fields.title}</h3>
{/if}

<!-- Simple equality comparison -->
{if post.type === "news"}
  <article>...</article>
{/if}

<!-- Simple inequality -->
{if post.fields.published !== false}
  <div>Published content</div>
{/if}
```

### ✅ DO: Use Nested Conditionals for Complex Logic

```html
{if post.fields.items}
  {if post.fields.items.length > 0}
    <ul>
      {for item in post.fields.items}
        <li>{item}</li>
      {/for}
    </ul>
  {/if}
{/if}
```

### ✅ DO: Use Fallback Operators

```html
<h3>{post.fields.title || 'Untitled'}</h3>
<p>{post.fields.description || 'No description available'}</p>
```

### ❌ DON'T: Use Complex Boolean Expressions

```html
<!-- ❌ BAD: Complex AND expression -->
{if post.fields.items && post.fields.items.length > 0}

<!-- ❌ BAD: Complex OR expression -->
{if post.fields.title || post.fields.fallbackTitle}

<!-- ❌ BAD: Multiple conditions -->
{if post.published && post.featured && post.type === "news"}
```

### ❌ DON'T: Use Method Calls in Conditionals

```html
<!-- ❌ BAD: Method calls -->
{if post.fields.items.includes('featured')}

<!-- ❌ BAD: Complex property access -->
{if post.fields.author?.profile?.verified}
```

---

## Debugging Tips

### Check if Conditionals are Expanding

**Test**: View page source and search for `<template x-if`

**If you find** `<template x-if="post.fields...">`:
- ❌ The conditional did NOT expand at build-time
- ❌ You'll get runtime errors about `post` not defined
- ✅ Simplify your conditional

**If you don't find it**:
- ✅ The conditional expanded successfully
- ✅ No runtime Alpine templates created

### Console Error Patterns

**Error**: `Uncaught ReferenceError: post is not defined`

**Means**: A conditional inside your build-time loop is falling back to runtime

**Fix**: Simplify the conditional or remove it

---

**Error**: `[Alpine] post.fields.textItems && post.fields.textItems.length > 0`

**Means**: Alpine is trying to evaluate a conditional with loop variables

**Fix**: This exact pattern is too complex - use nested conditionals instead

---

## Summary Table

| Pattern | Build-Time | Runtime Alpine | Result |
|---------|------------|----------------|--------|
| `{if post.title}` | ✅ Expands | ❌ N/A | ✅ Works |
| `{if post.type === "news"}` | ✅ Expands | ❌ N/A | ✅ Works |
| `{if post.items && post.items.length > 0}` | ❌ Falls back | ✅ Creates template | ❌ Error! |
| Nested: `{if post.items} {if post.items.length > 0}` | ✅ Both expand | ❌ N/A | ✅ Works |
| `{post.title \|\| 'Default'}` | ✅ Resolved | ❌ N/A | ✅ Works |

---

## Real-World Examples

### Example 1: News Posts

**Problem**:
```html
{for post in allContent}
  {if post.fields.textItems && post.fields.textItems.length > 0}
    <p>{post.fields.textItems[0].paragraph}</p>
  {/if}
{/for}
```

**Solution**:
```html
{for post in allContent}
  {if post.fields.textItems}
    <p>{post.fields.textItems[0].paragraph}</p>
  {/if}
{/for}
```

Or even simpler (if all posts have textItems):
```html
{for post in allContent}
  <p>{post.fields.textItems[0].paragraph}</p>
{/for}
```

---

### Example 2: Optional Images

**Problem**:
```html
{for post in allContent}
  {if post.fields.image && post.fields.image.src}
    <img src="{post.fields.image.src}" />
  {/if}
{/for}
```

**Solution**:
```html
{for post in allContent}
  {if post.fields.image}
    <img src="{post.fields.image.src}" />
  {/if}
{/for}
```

---

### Example 3: Author Information

**Problem**:
```html
{if post.fields.author && post.fields.author.name}
  <span>By {post.fields.author.name}</span>
{/if}
```

**Solution**:
```html
{if post.fields.author}
  <span>By {post.fields.author.name || 'Anonymous'}</span>
{/if}
```

---

## See Also

- [BUILD_TIME_LOOP_EXPANSION.md](./BUILD_TIME_LOOP_EXPANSION.md) - Main loop expansion documentation
- [DEVELOPER_GUIDE.md](./DEVELOPER_GUIDE.md) - Complete developer guide
- [ARCHITECTURE.md](./ARCHITECTURE.md) - System architecture overview
