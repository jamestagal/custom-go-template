# Breaking Changes - Build-Time Loop Expansion

## Overview

The loop transformer now expands loops at build time instead of generating Alpine.js x-for templates for all loops. This is a **BEHAVIORAL CHANGE** but improves functionality.

## What Changed

### Before (Old Behavior)

**ALL** loops generated Alpine x-for templates:

```html
Input:  {for item in items}<div>{item}</div>{/for}
Output: <template x-for="item in items"><div x-text="item"></div></template>
```

### After (New Behavior)

**Build-time resolvable** loops are expanded:

```html
Input:  {for item in items}<div>{item}</div>{/for}
With:   dataScope = {items: ["A", "B", "C"]}
Output: <div><span x-text="item">A</span></div>
        <div><span x-text="item">B</span></div>
        <div><span x-text="item">C</span></div>
```

**Runtime-only** loops still use x-for:

```html
Input:  {for item in $store.cart.items}<div>{item}</div>{/for}
Output: <template x-for="item in $store.cart.items"><div x-text="item"></div></template>
```

## Impact

### Positive Impacts

1. **Dynamic component resolution now works** - Main goal achieved
2. **Better SEO** - Fully expanded HTML in server output
3. **Matches Svelte behavior** - Familiar to Svelte developers
4. **Performance** - No runtime loop overhead for static content

### Potential Issues

1. **Template expectations** - If code explicitly expected x-for templates, behavior differs
2. **Large arrays** - Build-time expansion of very large arrays (100+ items) may increase build time
3. **Test updates** - Tests expecting x-for output need updates

## Migration Guide

### If Your Templates Use Loops

**No action needed** - Templates will work better! Component name resolution in loops now works.

### If Your Tests Expect x-for Output

Update tests to expect either:
- **Build-time expansion** (for resolvable collections)
- **Runtime x-for** (for store collections and complex expressions)

### If You Need Runtime x-for

Use store collections or complex expressions:

```html
<!-- Will still generate x-for -->
{for item in $store.dynamicItems}
  <div>{item}</div>
{/for}
```

Or use Alpine directly:

```html
<template x-for="item in items">
  <div x-text="item"></div>
</template>
```

## Timeline

- **Implemented**: 2025-10-19
- **Spec**: `.agent-os/specs/2025-10-19-build-time-loop-expansion/`
- **Version**: Effective immediately

## Questions?

See spec documentation or ask in project discussions.
