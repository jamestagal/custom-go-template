# Store Expression Context Examples

**Last Updated:** 2026-01-29

This guide provides clear examples of how store expressions (`{$storeName.property}`) are transformed differently based on their **context** in the template.

---

## The Dual Syntax System

**What you write** (template syntax):
```html
{$theme.mode}
```

**What gets generated** (Alpine.js syntax):
```html
$store.theme.mode
```

**The magic**: The transformation changes based on WHERE the expression appears!

---

## Context 1: Text Content

### What You Write
```html
<p>Current theme: {$theme.mode}</p>
<h1>Welcome, {$auth.user.name}!</h1>
<span>Total: ${$cart.formattedTotal}</span>
```

### What Gets Generated
```html
<p>Current theme: <span x-text="$store.theme.mode"></span></p>
<h1>Welcome, <span x-text="$store.auth.user.name"></span>!</h1>
<span>Total: $<span x-text="$store.cart.formattedTotal"></span></span>
```

### Why This Transformation?
- Text content needs a **wrapper element** for Alpine.js to bind to
- `x-text` directive tells Alpine to reactively update the text
- The `<span>` is invisible but allows Alpine to manage the content

### Live Example
```html
---
import store from './stores/theme.js'
---

<div>
  <p>Mode: {$theme.mode}</p>
  <button @click="$store.theme.toggle()">Toggle</button>
</div>
```

**Output HTML:**
```html
<div>
  <p>Mode: <span x-text="$store.theme.mode"></span></p>
  <button @click="$store.theme.toggle()">Toggle</button>
</div>
```

**Rendered in browser:**
```
Mode: light
[Toggle] ← button
```

After clicking Toggle:
```
Mode: dark
[Toggle] ← button (Alpine reactively updates the span!)
```

---

## Context 2: HTML Attributes (Regular)

### What You Write
```html
<div class={$theme.mode}>Content</div>
<img src={$user.avatar}>
<a href={$settings.homeUrl}>Home</a>
```

### What Gets Generated
```html
<div :class="$store.theme.mode">Content</div>
<img :src="$store.user.avatar">
<a :href="$store.settings.homeUrl">Home</a>
```

### Why This Transformation?
- Regular HTML attributes need `:` prefix for Alpine.js **reactive binding**
- `:class` = "bind this class attribute to reactive data"
- `:src` = "bind this src attribute to reactive data"
- Alpine watches the store and updates the attribute when it changes

### Live Example
```html
---
import store from './stores/theme.js'
---

<body class={$theme.mode}>
  <div class="container">
    <p>This page has theme class on body!</p>
    <button @click="$store.theme.toggle()">Toggle Theme</button>
  </div>
</body>
```

**Output HTML:**
```html
<body :class="$store.theme.mode">
  <div class="container">
    <p>This page has theme class on body!</p>
    <button @click="$store.theme.toggle()">Toggle Theme</button>
  </div>
</body>
```

**Rendered in browser (initially):**
```html
<body class="light">
  <div class="container">
    <p>This page has theme class on body!</p>
    <button>Toggle Theme</button>
  </div>
</body>
```

**After clicking Toggle:**
```html
<body class="dark">  <!-- Alpine updated this! -->
  <div class="container">
    <p>This page has theme class on body!</p>
    <button>Toggle Theme</button>
  </div>
</body>
```

---

## Context 3: Alpine Directives (x-*, @*)

### What You Write
```html
<div x-show="$store.auth.isLoggedIn">Welcome!</div>
<div x-if="$store.cart.isEmpty">Cart is empty</div>
<button @click="$store.cart.addItem(item)">Add</button>
```

### What Gets Generated
```html
<div x-show="$store.auth.isLoggedIn">Welcome!</div>
<div x-if="$store.cart.isEmpty">Cart is empty</div>
<button @click="$store.cart.addItem(item)">Add</button>
```

### Why This Transformation?
- **NO TRANSFORMATION** - already in Alpine syntax!
- Alpine directives (`x-*`, `@*`) expect raw JavaScript expressions
- You write `$store.*` directly in Alpine directives
- No curly braces `{}` needed

### Live Example
```html
---
import store from './stores/auth.js'
---

<div>
  <div x-show="$store.auth.isLoggedIn">
    <p>Welcome back, <span x-text="$store.auth.user.name"></span>!</p>
    <button @click="$store.auth.logout()">Logout</button>
  </div>

  <div x-show="!$store.auth.isLoggedIn">
    <p>Please log in</p>
    <button @click="$store.auth.login()">Login</button>
  </div>
</div>
```

**Output HTML (unchanged):**
```html
<div>
  <div x-show="$store.auth.isLoggedIn">
    <p>Welcome back, <span x-text="$store.auth.user.name"></span>!</p>
    <button @click="$store.auth.logout()">Logout</button>
  </div>

  <div x-show="!$store.auth.isLoggedIn">
    <p>Please log in</p>
    <button @click="$store.auth.login()">Login</button>
  </div>
</div>
```

---

## Context 4: Template Curly Braces (`{$...}`)

### What You Write
```html
{if $auth.isLoggedIn}
  <p>Logged in</p>
{/if}

{for item in $cart.items}
  <div>{item.name}</div>
{/for}
```

### What Gets Generated
```html
<template x-if="$store.auth.isLoggedIn">
  <p>Logged in</p>
</template>

<template x-for="item in $store.cart.items">
  <div x-text="item.name"></div>
</template>
```

### Why This Transformation?
- Template directives (`{if}`, `{for}`) transform to Alpine `<template>` tags
- Store expressions inside get transformed: `$auth` → `$store.auth`
- Maintains reactivity through Alpine.js

### Live Example
```html
---
import store from './stores/cart.js'
---

<div>
  {if $cart.isEmpty}
    <p>Your cart is empty</p>
  {else}
    <p>You have {$cart.itemCount} items</p>

    {for item in $cart.items}
      <div class="cart-item">
        {item.name} - ${item.price}
      </div>
    {/for}

    <p>Total: ${$cart.formattedTotal}</p>
  {/if}
</div>
```

**Output HTML:**
```html
<div>
  <template x-if="$store.cart.isEmpty">
    <p>Your cart is empty</p>
  </template>

  <template x-else>
    <p>You have <span x-text="$store.cart.itemCount"></span> items</p>

    <template x-for="item in $store.cart.items">
      <div class="cart-item">
        <span x-text="item.name"></span> - $<span x-text="item.price"></span>
      </div>
    </template>

    <p>Total: $<span x-text="$store.cart.formattedTotal"></span></p>
  </template>
</div>
```

---

## Context 5: Dynamic Styling (`:style`)

### What You Write
```html
<body :style="`background: ${$store.theme.colors.background}; color: ${$store.theme.colors.text}`">
```

### What Gets Generated
```html
<body :style="`background: ${$store.theme.colors.background}; color: ${$store.theme.colors.text}`">
```

### Why This Transformation?
- **NO TRANSFORMATION** - already Alpine syntax!
- `:style` expects a JavaScript expression (template literal)
- You write `$store.*` directly inside the template literal

### Live Example
```html
---
import store from './stores/theme.js'
---

<body :style="`background: ${$store.theme.colors.background}; color: ${$store.theme.colors.text}`">
  <div class="container">
    <h1>Themed Page</h1>
    <button @click="$store.theme.toggle()">Toggle Theme</button>
  </div>
</body>
```

**Rendered in browser (light mode):**
```html
<body style="background: #ffffff; color: #1a1a1a">
  <div class="container">
    <h1>Themed Page</h1>
    <button>Toggle Theme</button>
  </div>
</body>
```

**After toggle (dark mode):**
```html
<body style="background: #1a1a1a; color: #e0e0e0">
  <div class="container">
    <h1>Themed Page</h1>
    <button>Toggle Theme</button>
  </div>
</body>
```

---

## Summary: Context Detection

| Context | Input Syntax | Output Transformation | Reason |
|---------|--------------|----------------------|--------|
| **Text content** | `{$theme.mode}` | `<span x-text="$store.theme.mode"></span>` | Needs wrapper for Alpine binding |
| **Regular attribute** | `class={$theme.mode}` | `:class="$store.theme.mode"` | Needs `:` prefix for reactive binding |
| **Alpine directive** | `x-show="$store.auth.isLoggedIn"` | *(unchanged)* | Already correct Alpine syntax |
| **Event handler** | `@click="$store.cart.add()"` | *(unchanged)* | Already correct Alpine syntax |
| **Template directive** | `{if $auth.isLoggedIn}` | `<template x-if="$store.auth.isLoggedIn">` | Template → Alpine template tag |
| **Loop** | `{for item in $cart.items}` | `<template x-for="item in $store.cart.items">` | Template → Alpine x-for |

---

## Key Rules

### ✅ Use `{$store.property}` When:
1. **In text content** - needs transformation to `<span x-text>`
2. **In regular attributes** - needs transformation to `:attribute`
3. **In template directives** - `{if}`, `{for}` conditions/collections

### ✅ Use `$store.store.property` When:
1. **In Alpine directives** - `x-show`, `x-if`, `x-text`, etc.
2. **In event handlers** - `@click`, `@submit`, etc.
3. **In `:style` or `:class`** - already Alpine binding context

### ❌ Never Mix:
```html
<!-- ❌ WRONG -->
<div class="$store.theme.mode">  <!-- Missing binding -->
{x-show="$auth.isLoggedIn"}      <!-- Invalid syntax -->
<span x-text="{$cart.total}">    <!-- x-text doesn't need {} -->

<!-- ✅ CORRECT -->
<div :class="$store.theme.mode"> <!-- Reactive binding -->
<div x-show="$store.auth.isLoggedIn">  <!-- Alpine directive -->
<span x-text="$store.cart.total">      <!-- Alpine directive -->
```

---

## Complete Working Example

```html
---
import store from './stores/theme.js'
import store from './stores/auth.js'
import store from './stores/cart.js'
---

<!DOCTYPE html>
<html>
<head>
  <title>Store Context Examples</title>
  <script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
</head>

<!-- Context 2: Regular attribute -->
<body class={$theme.mode}>

  <!-- Context 1: Text content -->
  <h1>Theme: {$theme.mode}</h1>

  <!-- Context 3: Alpine directive -->
  <div x-show="$store.auth.isLoggedIn">
    <!-- Context 1: Text content -->
    <p>Welcome, {$auth.user.name}!</p>

    <!-- Context 3: Event handler -->
    <button @click="$store.auth.logout()">Logout</button>
  </div>

  <!-- Context 4: Template directive -->
  {if $cart.isEmpty}
    <p>Cart is empty</p>
  {else}
    <!-- Context 1: Text content -->
    <p>Items: {$cart.itemCount}</p>

    <!-- Context 4: Loop -->
    {for item in $cart.items}
      <div>
        <!-- Context 1: Text content -->
        {item.name} - ${item.price}
      </div>
    {/for}
  {/if}

  <!-- Context 5: Dynamic styling -->
  <div :style="`background: ${$store.theme.colors.background}`">
    Themed background
  </div>

</body>
</html>
```

---

## Pro Tip: When in Doubt

**Inside curly braces `{...}`?** → Use `{$storeName.property}`
```html
{$theme.mode}
{if $auth.isLoggedIn}
{for item in $cart.items}
```

**Inside Alpine attributes?** → Use `$store.storeName.property`
```html
x-show="$store.auth.isLoggedIn"
@click="$store.cart.addItem()"
:style="`color: ${$store.theme.colors.text}`"
```

**The system knows the context and transforms accordingly!**

---

## See Also

- [STORE_SYNTAX_DESIGN.md](./STORE_SYNTAX_DESIGN.md) - Design rationale and analysis
- [STORE_DEVELOPER_GUIDE.md](./STORE_DEVELOPER_GUIDE.md) - Best practices and patterns
- [DEVELOPER_GUIDE.md](./DEVELOPER_GUIDE.md) - Complete developer documentation
