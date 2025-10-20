# Build-Time Loop Expansion - Examples

## Example 1: Component Loop from JSON

**Template** (`layouts/content/pages.html`):
```html
---
export let components
---

{for component in components}
  <Component:dynamic name={component.name} {...component.fields} />
{/for}
```

**Data** (`content/pages/index.json`):
```json
{
  "components": [
    {
      "name": "Hero2436",
      "fields": {"title": "Welcome", "subtitle": "To our site"}
    },
    {
      "name": "Services2437",
      "fields": {"title": "Our Services"}
    }
  ]
}
```

**Output** (fully expanded at build time):
```html
<div class="hero" x-data='{"title":"Welcome","subtitle":"To our site"}'>
  <h1 x-text="title">Welcome</h1>
  <p x-text="subtitle">To our site</p>
</div>

<div class="services" x-data='{"title":"Our Services"}'>
  <h2 x-text="title">Our Services</h2>
</div>
```

## Example 2: Nested Loops

**Template:**
```html
---
export let categories
---

{for category in categories}
  <div class="category">
    <h2>{category.name}</h2>
    {for item in category.items}
      <div class="item">{item.title}</div>
    {/for}
  </div>
{/for}
```

**Data:**
```json
{
  "categories": [
    {
      "name": "Electronics",
      "items": [
        {"title": "Laptop"},
        {"title": "Phone"}
      ]
    },
    {
      "name": "Books",
      "items": [
        {"title": "Fiction"},
        {"title": "Non-Fiction"}
      ]
    }
  ]
}
```

**Output:** Both loops expand at build time, producing fully rendered HTML for all categories and all items.

```html
<div class="category">
  <h2><span x-text="category.name">Electronics</span></h2>
  <div class="item"><span x-text="item.title">Laptop</span></div>
  <div class="item"><span x-text="item.title">Phone</span></div>
</div>

<div class="category">
  <h2><span x-text="category.name">Books</span></h2>
  <div class="item"><span x-text="item.title">Fiction</span></div>
  <div class="item"><span x-text="item.title">Non-Fiction</span></div>
</div>
```

## Example 3: Runtime Fallback

**Template:**
```html
---
import store from './stores/cart.js'
---

{for item in $store.cart.items}
  <div class="cart-item">{item.name} - ${item.price}</div>
{/for}
```

**Output:** Alpine x-for template (runtime evaluation):
```html
<template x-for="item in $store.cart.items">
  <div class="cart-item">
    <span x-text="item.name"></span> - $<span x-text="item.price"></span>
  </div>
</template>
```

**Why runtime?** Collection is from an Alpine store, which is only available at runtime.

## Example 4: Loop with Index

**Template:**
```html
---
export let items
---

{for item, index in items}
  <div>{index + 1}. {item}</div>
{/for}
```

**Data:**
```json
{
  "items": ["First", "Second", "Third"]
}
```

**Output** (with 3 items):
```html
<div><span x-text="index + 1">1</span>. <span x-text="item">First</span></div>
<div><span x-text="index + 1">2</span>. <span x-text="item">Second</span></div>
<div><span x-text="index + 1">3</span>. <span x-text="item">Third</span></div>
```

## Example 5: Simple Array Loop

**Template:**
```html
---
export let users
---

<ul>
{for user in users}
  <li>{user.name} ({user.email})</li>
{/for}
</ul>
```

**Data:**
```json
{
  "users": [
    {"name": "Alice", "email": "alice@example.com"},
    {"name": "Bob", "email": "bob@example.com"}
  ]
}
```

**Output:**
```html
<ul>
  <li><span x-text="user.name">Alice</span> (<span x-text="user.email">alice@example.com</span>)</li>
  <li><span x-text="user.name">Bob</span> (<span x-text="user.email">bob@example.com</span>)</li>
</ul>
```

## Example 6: Empty Array

**Template:**
```html
---
export let items
---

{for item in items}
  <div>{item}</div>
{/for}
<p>Done</p>
```

**Data:**
```json
{
  "items": []
}
```

**Output:**
```html
<p>Done</p>
```

**Behavior:** Empty arrays produce no output for the loop body, only content after the loop is rendered.

## Example 7: Mixed Build-Time and Runtime

**Template:**
```html
---
export let staticItems
import store from './stores/cart.js'
---

<h2>Static Content (Build-Time)</h2>
{for item in staticItems}
  <div>{item}</div>
{/for}

<h2>Dynamic Cart (Runtime)</h2>
{for item in $store.cart.items}
  <div>{item.name}</div>
{/for}
```

**Output:**
```html
<h2>Static Content (Build-Time)</h2>
<div><span x-text="item">Item 1</span></div>
<div><span x-text="item">Item 2</span></div>

<h2>Dynamic Cart (Runtime)</h2>
<template x-for="item in $store.cart.items">
  <div><span x-text="item.name"></span></div>
</template>
```

**Behavior:** The system intelligently uses build-time expansion for resolvable collections and runtime x-for for store-based collections.
