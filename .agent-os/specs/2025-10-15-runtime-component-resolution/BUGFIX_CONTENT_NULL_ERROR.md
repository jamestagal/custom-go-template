# Bugfix: "Cannot read properties of null (reading 'components')" Error

**Date**: 2025-10-15
**Status**: RESOLVED
**Type**: Prop Passing Issue in Dynamic Component Resolution

---

## Problem Summary

When accessing `/pages` route, Alpine.js threw a runtime error:
```
Cannot read properties of null (reading 'components')
```

This occurred in the runtime component resolution loop in pages.html:
```html
{for component in content.components}
    <Component:dynamic name={component.name} {...component.fields} content={content} allContent={allContent} />
{/for}
```

The error indicated that `content` was `null` when Alpine.js tried to evaluate `content.components` in the x-for expression.

---

## Root Cause Analysis

The issue was a **prop passing chain break** between html.html (wrapper) and pages.html (layout):

### Flow Breakdown

1. **Server (renderWithWrapper)**: ✓ Correctly builds full `content` object
   ```go
   props := map[string]interface{}{
       "content": contentData, // Full content object with .components array
       "layout": layoutName,
       // ... other props
   }
   ```

2. **html.html Wrapper**: ✓ Declares and receives `content`
   ```html
   export let content, layout
   ```

3. **html.html → pages.html**: ✗ **ONLY passes content.fields, NOT content**
   ```html
   <!-- BEFORE (BROKEN) -->
   <Component:dynamic name={layout}
       {...content.fields}
       bind:shadowContent={shadowContent} />
   ```

4. **pages.html**: ✗ Has NO fence section, expects `content` from parent scope
   ```html
   <!-- NO fence section - relies on parent scope -->
   {for component in content.components}
   ```

5. **Runtime Wrapper**: ✗ Creates NEW x-data scope WITHOUT content
   ```html
   <!-- Emitted by emitRuntimeWrapper() -->
   <div x-data="{compName: component.name, compProps: {...}}" ...>
   ```

### Why This Breaks

When `pages.html` has no fence section, it should inherit variables from the parent scope (html.html). However:

1. The **dynamic component resolution** (`<Component:dynamic name={layout}>`) is detected as **runtime** (because `layout` variable is not a string literal)
2. This triggers `emitRuntimeWrapper()` which creates a **new x-data scope**
3. This new scope is populated from the **props passed to Component:dynamic**
4. Since html.html only passed `{...content.fields}`, the new scope has individual fields but **NOT the `content` object itself**
5. Therefore, `content.components` evaluates to `null.components` → ERROR

---

## The Fix

**File**: `layouts/global/html.html`
**Change**: Add explicit `content={content}` prop to the dynamic component invocation

```html
<!-- AFTER (FIXED) -->
<Component:dynamic name={layout}
    content={content}        <!-- ← CRITICAL FIX: Pass full content object -->
    {...content.fields}
    bind:shadowContent={shadowContent} />
```

### Why This Works

1. html.html now **explicitly passes** the `content` prop to the layout component
2. When `emitRuntimeWrapper()` builds the x-data scope, it includes `content` in the props serialization
3. The runtime wrapper's x-data becomes:
   ```html
   <div x-data="{compName: component.name, compProps: {content: content, ...}}" ...>
   ```
4. Alpine.js can now resolve `content.components` correctly in the loop expression

---

## Key Lessons

### 1. **Explicit Prop Passing for Dynamic Components**

When using `<Component:dynamic>`, **ALWAYS pass all required props explicitly**, even if they're in scope. Dynamic components that trigger runtime resolution create their own x-data scope, breaking scope inheritance.

```html
<!-- BAD: Assumes child inherits from parent scope -->
<Component:dynamic name={layout} />

<!-- GOOD: Explicitly passes all needed props -->
<Component:dynamic name={layout} content={content} allContent={allContent} />
```

### 2. **Runtime Wrappers Isolate Scope**

The runtime component resolution system (`emitRuntimeWrapper()`) creates a **new isolated x-data scope**. This is necessary for Alpine.js to work correctly, but it means:

- Props must be **explicitly passed**, not inherited
- Spread operators (`{...content.fields}`) spread individual fields, not the parent object
- Always include the parent object itself if children need nested access

### 3. **No Fence Section = Scope Dependency**

Templates without fence sections (like pages.html) rely entirely on:
- Props passed from parent component
- Parent scope (for static components)
- Runtime wrapper scope (for dynamic components)

This dependency must be satisfied explicitly in the parent component.

---

## Testing Verification

After the fix:

1. **Server renders correctly**:
   ```bash
   curl http://localhost:3333/pages
   ```
   Returns HTML with:
   ```html
   <template x-for="component in content.components">
       <div x-data="{compName: component.name, compProps: {content: content, allContent: allContent}}"
            class="dyn-comp-runtime"
            x-init="$renderDynamicComponent($el, compName, compProps)">
       </div>
   </template>
   ```

2. **Alpine.js evaluates successfully**:
   - `content.components` resolves to the components array
   - Loop iterates over components
   - No runtime errors in browser console

---

## Related Files

- `layouts/global/html.html` - Wrapper template (FIXED)
- `layouts/content/pages.html` - Layout template (uses content.components)
- `transformer/dynamic_component_by_name.go` - Runtime wrapper emission
- `cmd/server/main.go` - renderWithWrapper() function
- `.agent-os/specs/2025-10-15-runtime-component-resolution/` - Full spec

---

## Diff

```diff
--- a/layouts/global/html.html
+++ b/layouts/global/html.html
@@ -50,8 +50,10 @@
 		<!-- The server will inject the appropriate layout based on the route -->

 		<!-- Slot for dynamic content - replaced by server with specific layout -->
+		<!-- CRITICAL FIX: Pass content explicitly to layout component -->
+		<!-- Pages.html needs content.components for iteration -->
 		<Component:dynamic name={layout}
+			content={content}
 			{...content.fields}
 			bind:shadowContent={shadowContent} />
 			<!-- TEMPORARILY DISABLED: Magic variables causing JS syntax errors
```

---

## Prevention Guidelines

For future component development:

1. **When creating layout templates**:
   - Document required props in comments
   - Add fence section with `export let` if needed
   - Test with and without parent scope

2. **When using Component:dynamic**:
   - Always pass props explicitly
   - Include parent objects for nested access
   - Test runtime resolution code paths

3. **When debugging "null" errors**:
   - Check x-data scopes in rendered HTML
   - Trace prop passing chain from server → wrapper → layout → component
   - Verify runtime wrapper includes all needed variables

---

## Status: RESOLVED ✓

The fix has been implemented and verified. The `/pages` route now renders correctly without Alpine.js errors.
