# Comprehensive Template Pattern Test Checklist

**Date**: 2026-01-28 (Last Updated)
**Primary Test Files**:
- `content/pages/jim-test.json` - http://localhost:3333/jim-test (syntax showcase with 7 components)
- `layouts/content/store-demo.html` - http://localhost:3333/store-demo (global stores demo)
- `layouts/content/news_page.html` - http://localhost:3333/news_page (aggregate listing)
- `layouts/content/committee_page.html` - http://localhost:3333/committee_page (aggregate listing)

**Test Result**: ✅ **PASSED** - Zero console errors, all core patterns rendering correctly

This document tracks testing of all template patterns and provides a **Plenti compatibility comparison** to ensure feature parity with the original Plenti SSG.

---

## Plenti Compatibility Comparison

### Feature Parity Matrix

| Feature | Plenti | Go Template Engine | Status |
|---------|--------|-------------------|--------|
| **Template Syntax** | Svelte-style `{#if}`, `{#each}` | Svelte-style `{if}`, `{for}` | ✅ Implemented |
| **Fence Sections** | `---` with props/exports | `---` with props/exports | ✅ Implemented |
| **Component System** | `.svelte` files | `.html` files | ✅ Implemented |
| **Magic Variables** | `allContent`, `allLayouts`, `content` | `allContent`, `content`, `env`, `buildTime` | ✅ Implemented (allLayouts not needed) |
| **Content Types** | Folder-based (`content/pages/`, `content/news/`) | Folder-based | ✅ Implemented |
| **Single Content Types** | `--single=true` (JSON in content root) | Automatic detection | ✅ Implemented |
| **Aggregate Pages** | Type filtering in layouts | `{for}` + `{if type === ""}` | ✅ Implemented |
| **Global Stores** | Not native (use Svelte stores) | Alpine.js stores | ✅ Implemented |
| **Dynamic Components** | `<svelte:component>` | `<Component:dynamic>` | ✅ Implemented |
| **Build-Time Loop Expansion** | Yes | Yes (hybrid approach) | ✅ Implemented |
| **CSS Scoping** | Svelte scoped styles | Component `<style>` tags | ✅ Implemented |
| **JSON Content Loading** | Automatic from `content/` | `export let` + loader | ✅ Implemented |
| **HTML Wrapper** | `layouts/global/html.svelte` | `layouts/global/html.html` | ✅ Implemented |

### Content Type Patterns

#### 1. Component-Based Pages (Plenti Pattern)

**JSON Location**: `content/pages/*.json`
**Template**: `layouts/content/pages.html`

```json
{
  "components": [
    {"name": "hero2436", "fields": {"title": "...", "description": "..."}},
    {"name": "services2437", "fields": {}}
  ]
}
```

```html
---
export let components, allContent, content
---

{for component in components}
  <Component:dynamic name={component.name} {...component.fields} />
{/for}
```

#### 2. Custom Template Pages (Flat Structure)

**JSON Location**: `content/portfolio/*.json`
**Template**: `layouts/content/portfolio.html`

```json
{
  "title": "Project Name",
  "subtitle": "Description",
  "features": [...]
}
```

```html
---
export let title, subtitle, features
---

<h1>{title}</h1>
<p>{subtitle}</p>
```

#### 3. Single Content Type (--single=true Pattern)

**JSON Location**: `content/news_page.json` (root level, no folder)
**Template**: `layouts/content/news_page.html`

```json
{}
```

```html
---
export let allContent
---

{for post in allContent}
  {if post.type === "news"}
    <article>{post.fields.title}</article>
  {/if}
{/for}
```

### Magic Variable Usage

| Variable | Description | How to Access |
|----------|-------------|---------------|
| `content` | Current page JSON data | Always available |
| `allContent` | All site content (for navigation, listings) | `export let allContent` |
| `components` | Page's component array | `export let components` |

> **Note**: `allLayouts` (Plenti's component registry hack) is not needed - our build-time component resolution handles `<Component:dynamic>` at compile time.

### Aggregate Page Pattern

For creating listing pages (news, blog, committee):

```html
---
export let allContent
---

<section id="blog-listing">
  {for post in allContent}
    {if post.type === "news"}
      <article class="cs-item">
        <a href={post.path}>
          <h3>{post.fields.title}</h3>
          <p>{post.fields.description}</p>
        </a>
      </article>
    {/if}
  {/for}
</section>
```

**Key Points**:
- Uses `allContent` to access all site content
- Filters by `post.type` to show only specific content type
- Access fields via `post.fields.*`
- Access path via `post.path` (auto-generated from filename)

---

## Test Status Summary

### Core Features (Tested in jim-test.html)
- **Console Errors**: ✅ Zero errors (jim-test, store-demo, aggregate pages)
- **Page Styling**: ✅ Working (page + component style aggregation)
- **Basic Expressions**: ✅ Working (jim_test_greeting: `{salutation} {name}`)
- **Conditionals**: ✅ Working (if/else-if/else, nested) (jim_test_greeting)
- **Loops (arrays)**: ✅ Working (`{for animal of animals}`) (jim_test_animals_loop)
- **Dynamic Components**: ✅ Working (`<='{path}'` and `<="../components/{comp}.html"`) (jim_test_user_profiles)
- **Component Props**: ✅ Working (static, dynamic, expression props) (jim_test_age_examples)
- **Reactive State**: ✅ Working (add/remove animals, notifications) (jim_test_animals_loop)

### Advanced Features (Tested in jim-test.html)
- **Array Spread in Loops**: ✅ Working `{for animal of ["🦄 unicorn", ...animals]}` (jim_test_advanced_loops)
- **Variable Component Paths**: ✅ Working `<='{path}'` and `<="../components/{comp}.html"` (jim_test_user_profiles)
- **Notification Components**: ✅ Working (loop with types: success, info, warning, error) (jim_test_notifications)
- **Interactive State**: ✅ Working (onclick handlers updating reactive vars) (jim_test_animals_loop)
- **String Manipulation**: ✅ Working (`.split('').reverse().join('')`) (jim_test_animals_loop)
- **Array Methods**: ✅ Working (`.filter()`, inline arrays) (jim_test_animals_loop)
- **x-model Input Binding**: ✅ Working (newAnimal input field) (jim_test_animals_loop)
- **Inline Array Iteration**: ✅ Working `{for word of ["Waller", "loves Plenti"]}` (jim_test_advanced_loops)

### Store Features (Tested in store-demo.html)
- **External Store Imports**: ✅ Working (`import store from './stores/auth.js'`)
- **Store Expressions**: ✅ Working (`{$auth.isLoggedIn}`, `{$cart.items.length}`)
- **Store Methods**: ✅ Working (`@click="$store.auth.login()"`)
- **Store Computed Properties**: ✅ Working (`$cart.formattedTotal`, `$theme.getCurrentColors()`)
- **Store Conditionals**: ✅ Working (`{if $auth.isLoggedIn}`)
- **Store Loops**: ✅ Working (`{for item in $cart.items}`)
- **Dynamic Styling with Stores**: ✅ Working (`:style` bindings with store values)

### Aggregate Page Features (Tested in news_page.html, committee_page.html)
- **allContent Magic Variable**: ✅ Working (`export let allContent`)
- **Content Type Filtering**: ✅ Working (`{if post.type === "news"}`)
- **Path Access**: ✅ Working (`{post.path}`, `{post.fields.title}`)
- **Single Content Type Routes**: ✅ Working (JSON files in content root)

## Fixed Issues (Previously Bugs)

### ✅ FIXED #1: Server x-data Building
- **Was**: Server manually extracted fence data and built x-data as JSON
- **Fix**: Server now uses `renderer.Render()` and `buildXDataFromProps()` with proper function extraction
- **Status**: Functions work correctly in fence section (see store-demo.html login/logout demo)
- **Spec**: Completed in fix-server-xdata-building spec (2025-10-07)

### Still TODO: Component Props in Loops
- **Issue**: Props like `inStock={product.inStock}` in loops may not evaluate correctly
- **Status**: Not fully tested with complex scenarios
- **Workaround**: Use static prop values or test more thoroughly

### Still TODO: Attribute Expressions
- **Issue**: Expressions in attributes like `class="{dynamicClass}"` may need Alpine binding support
- **Status**: Not tested in current files
- **Future**: May need transformer enhancement for attribute-position expressions

---

## Feature Coverage by File

### jim-test.html - Syntax Features Showcase
**URL**: http://localhost:3333/jim-test
**JSON**: `content/pages/jim-test.json`

This page showcases template syntax features through 7 specialized components:

#### Component 1: jim_test_greeting
**Features**: Basic expressions, nested conditionals
```html
<h1>{salutation} {name}!</h1>

{if name.length > 3}
  <div>{name} is a long name</div>
  {if age > 1}
    <div>Has been born</div>
  {/if}
{else if name.length == 2}
  <div>{name} is medium</div>
{else}
  <div>{name} is a short name</div>
{/if}
```

#### Component 2: jim_test_age_examples
**Features**: Component imports, dynamic props, expression props
```html
import Age from "./age.html";

<Age name={name} age={age} />
<Age name={"Bo"} age={age + 50} />
<Age name={"Baggins"} age={201} />
```

#### Component 3: jim_test_user_profiles
**Features**: Dynamic components with variable paths
```html
let path = "../components/userprofile.html"
let comp = "userprofile"

<="../components/userprofile.html" user={user1} showRole={true} />
<='{path}' user={user2} showRole={true} />
<="../components/{comp}.html" user={user3} showRole={true} />
```

#### Component 4: jim_test_animals_loop
**Features**: Loops with `of` syntax, string manipulation, array methods, x-model
```html
{for animal of animals}
  <div>{name} likes: {animal}s</div>
  <div>Backwards: {animal.split('').reverse().join('')}</div>
  <button onclick="{animals = animals.filter(a => a !== animal)}">Remove</button>
{/for}

<input type="text" x-model="newAnimal">
<button onclick="{animals = [newAnimal, ...animals]}">Add Animal</button>
```

#### Component 5: jim_test_advanced_loops
**Features**: Array spread in loops, inline array iteration
```html
{for animal of ["🦄 unicorn", ...animals]}
  <div>{animal}</div>
{/for}

{for word of ["Waller", "loves Plenti", "uses AI", "is Australian"]}
  <div>{name} {word}</div>
{/for}
```

#### Component 6: jim_test_todos
**Features**: Component with numeric props
```html
<Todos number={5} />
<Todos start={5} number={5} />
```

#### Component 7: jim_test_notifications
**Features**: Interactive notifications, conditional component rendering
```html
{for notif of notifications}
  <button onclick="{currentNotification = notif}">Show {notif.type}</button>
{/for}

{if currentNotification}
  <Notification type={currentNotification.type} message={currentNotification.message} />
{/if}
```

---

### store-demo.html - Global Store System Demo
**URL**: http://localhost:3333/store-demo

**Store Imports**:
```html
import store from './stores/auth.js'
import store from './stores/cart.js'
import store from './stores/theme.js'
```

**Features Tested**:

1. **Store Expressions in Templates**:
   - `{$auth.isLoggedIn}` - Boolean state
   - `{$auth.user.name}` - Nested object access
   - `{$cart.items.length}` - Array length
   - `{$cart.formattedTotal}` - Computed property (getter)

2. **Store Conditionals**:
   ```html
   {if $auth.isLoggedIn}
     <p>Welcome, {$auth.user.name}!</p>
   {else}
     <p>Please log in</p>
   {/if}
   ```

3. **Store Loops**:
   ```html
   {for item in $cart.items}
     <li>{item.name} - ${item.price}</li>
   {/for}
   ```

4. **Store Methods in Event Handlers**:
   ```html
   <button @click="$store.auth.login()">Login</button>
   <button @click="$store.cart.addItem({ name: 'Widget', price: 9.99 })">Add</button>
   <button @click="$store.theme.toggle()">Toggle Theme</button>
   ```

5. **Dynamic Styling with Stores**:
   ```html
   <body :style="`background: ${$store.theme.getCurrentColors().background}`">
   ```

6. **Reusable Store Components**:
   - `<LoginStatus />` - Displays auth state
   - `<CartBadge />` - Shows cart count/total
   - `<ThemeToggle />` - Theme switcher buttons

---

## Test Instructions

1. Start server: `go run cmd/server/main.go`
2. Test syntax features: http://localhost:3333/jim-test
3. Test global stores: http://localhost:3333/store-demo
4. Test aggregate pages: http://localhost:3333/news_page and http://localhost:3333/committee_page
5. Open browser console to check for errors (should be zero)
6. Verify each section visually and functionally
7. Check that page styling is applied (component + page styles aggregated)

---

## Section 1: Basic Expressions (Lines 139-151)

**Purpose**: Test simple variable interpolation, object property access, and function calls

### Test Cases

- [ ] **1.1 Simple variable**: `{title}` displays "Custom Template Showcase"
- [ ] **1.2 Object property**: `{user.name}` displays "John Doe"
- [ ] **1.3 Object nested property**: `{user.role}` displays "admin"
- [ ] **1.4 Function call**: `{getGreeting()}` displays time-appropriate greeting
- [ ] **1.5 Function with parameter**: `{user.name}` used in greeting context
- [ ] **1.6 Array length**: `{products.length}` displays "4"
- [ ] **1.7 Array length (categories)**: `{categories.length}` displays count
- [ ] **1.8 Object property (settings)**: `{settings.theme}` displays "light"
- [ ] **1.9 Object property (settings)**: `{settings.currency}` displays "USD"

**Expected Alpine.js Output**:
```html
<span x-text="title"></span>
<span x-text="user.name"></span>
<span x-text="getGreeting()"></span>
```

**Browser Verification**:
- No `{variable}` literal text visible
- All values render correctly
- Console: No Alpine.js errors

---

## Section 2: Conditionals (Lines 153-196)

**Purpose**: Test if/else, if/else-if/else, and nested conditionals

### Test Cases

- [ ] **2.1 Simple if/else**: Login status displays correctly based on `isLoggedIn`
  - When `isLoggedIn = true`: Shows "You are currently logged in as John Doe"
  - When `isLoggedIn = false`: Shows "You are not logged in"

- [ ] **2.2 Multiple else-if branches**: User role check
  - Admin role: "Welcome, Administrator!"
  - Manager role: "Welcome, Manager!"
  - Editor role: "Welcome, Editor!"
  - Default: "Welcome, User!"

- [ ] **2.3 Nested conditionals**: Product availability
  - Outer condition: `filteredProducts.length > 0`
  - Inner condition: `settings.filters.inStockOnly`
  - Both branches display correctly

**Expected Alpine.js Output**:
```html
<template x-if="isLoggedIn">
  <p>You are currently logged in...</p>
</template>
<template x-if="!(isLoggedIn)">
  <p>You are not logged in...</p>
</template>
```

**Browser Verification**:
- No `{if}`, `{else}`, `{/if}` literal text
- Only one branch renders (correct conditional logic)
- Console: No template errors

---

## Section 3: Loops (Lines 198-279)

**Purpose**: Test array loops, indexed loops, object loops, and nested loops

### Test Cases

- [ ] **3.1 Simple array loop**: `{for product in filteredProducts}`
  - All filtered products display
  - Out-of-stock indicator shows correctly

- [ ] **3.2 Loop with index**: `{for product, index in filteredProducts}`
  - List numbers start at 1 (index + 1)
  - Featured star (★) displays for featured products

- [ ] **3.3 Object loop**: `{for key, value of settings}`
  - All settings keys display (theme, currency, showFeatured, filters)
  - Nested object (filters) handled correctly

- [ ] **3.4 Nested object loop**: Within settings loop
  - `{for subKey, subValue of value}` displays filter properties
  - Shows: minPrice, maxPrice, inStockOnly

- [ ] **3.5 Triple-nested loops**: Categories → Items → Tags
  - Category loop: `{for category in categories}`
  - Item loop: `{for item in category.items}`
  - Tag loop: `{for tag in item.tags}`
  - Tag classes apply correctly (bg-blue-100, bg-green-100, etc.)

**Expected Alpine.js Output**:
```html
<template x-for="product in filteredProducts">
  <li><span x-text="product.name"></span></li>
</template>

<template x-for="(product, index) in filteredProducts">
  <li :value="index + 1">...</li>
</template>

<template x-for="[key, value] in Object.entries(settings)">
  <dt x-text="key"></dt>
</template>
```

**Browser Verification**:
- No `{for}`, `{/for}` literal text
- All items render (count matches data)
- Nested loops render correctly
- Console: No iteration errors

---

## Section 4: Components (Lines 281-327)

**Purpose**: Test static components, components in loops, conditional components, dynamic components

### Test Cases

- [ ] **4.1 Static component with props**: `<Header title={title} user={user} />`
  - Header renders at top of page
  - Props passed correctly
  - Header styles applied

- [ ] **4.2 Component with boolean prop**: `<UserProfile user={user} showRole={true} />`
  - User profile displays
  - Role is visible (showRole=true)

- [ ] **4.3 Component in loop**: Product cards
  - `{for product in filteredProducts.filter(p => p.featured)}`
  - Only featured products show (2 cards expected)
  - Each ProductCard receives correct props
  - `formatPrice` function passed and works

- [ ] **4.4 Conditional component loop**: Notifications
  - 3 notifications render
  - Each has correct type (info, success, warning)
  - Type-specific styling applied

- [ ] **4.5 Dynamic component**: `<="./components/AdminPanel.html" user={user} />`
  - When `user.role === "admin"`: AdminPanel renders
  - When `user.role !== "admin"`: UserDashboard renders
  - Correct component loaded based on condition

- [ ] **4.6 Footer component**: `<Footer />`
  - Footer renders at bottom
  - Footer styles applied
  - No empty x-data wrapper (semantic HTML fix)

**Expected Alpine.js Output**:
```html
<!-- Static component -->
<header x-data="{ title: '...', user: {...} }">...</header>

<!-- Component in loop -->
<template x-for="product in filteredProducts.filter(p => p.featured)">
  <div x-data="{ name: product.name, price: product.price }">...</div>
</template>

<!-- Dynamic component -->
<template x-if="user.role === 'admin'">
  <div x-data="{ user: {...} }"><!-- AdminPanel --></div>
</template>
```

**Browser Verification**:
- All components render
- Component styles appear in `<head>` (style aggregation)
- No duplicate styles (SHA256 deduplication)
- Console: No component loading errors

---

## Section 5: Advanced Features (Lines 329-385)

**Purpose**: Test computed values, complex expressions, and combined features

### Test Cases

- [ ] **5.1 Computed values in fence**: `filteredProducts` (line 42-46)
  - Pre-computed array filters correctly
  - Used throughout template without recalculation

- [ ] **5.2 Complex inline expressions**:
  - Average price: `products.reduce((sum, p) => sum + p.price, 0) / products.length`
  - Featured count: `products.filter(p => p.featured).length`
  - In-stock count: `products.filter(p => p.inStock).length`
  - Total inventory value: `products.filter(p => p.inStock).reduce(...)`

- [ ] **5.3 Combined features**: Top 3 products section
  - Conditional: `{if filteredProducts.length > 0}`
  - Math function: `Math.min(3, filteredProducts.length)`
  - Loop: `{for product, index in filteredProducts.slice(0, 3)}`
  - Nested conditional in loop: `{if product.inStock}`
  - Nested loop in loop: `{for tag in product.tags}`

**Expected Alpine.js Output**:
```html
<!-- Complex expression -->
<span x-text="formatPrice(products.reduce((sum, p) => sum + p.price, 0) / products.length)"></span>

<!-- Combined features -->
<template x-if="filteredProducts.length > 0">
  <div>
    <span x-text="Math.min(3, filteredProducts.length)"></span>
    <template x-for="(product, index) in filteredProducts.slice(0, 3)">
      <!-- Nested structures -->
    </template>
  </div>
</template>
```

**Browser Verification**:
- All calculations correct
- No JavaScript errors from complex expressions
- Nested structures render properly
- Console: No Alpine.js evaluation errors

---

## Style Aggregation Verification

**Purpose**: Verify Spec 7 (Component Style Aggregation) works correctly

### Test Cases

- [ ] **Styles in `<head>`**: Check page source for aggregated styles
  - Page-level styles from jim-test.html (via pages.html layout)
  - jim_test_greeting component styles
  - jim_test_animals_loop component styles
  - UserProfile component styles
  - Notification component styles
  - Age component styles

- [ ] **No duplicate styles**: Each component's styles appear only once
  - SHA256 deduplication working
  - No multiple identical `<style>` blocks

- [ ] **Dependency order**: Child components before parent
  - Component dependencies loaded in correct order
  - Styles available when components render

- [ ] **Cache performance** (dev tools Network tab):
  - First load: All styles aggregated
  - Subsequent loads: Fast (cache hit)

**Browser Verification**:
- View page source → `<head>` section
- Count `<style>` blocks (should match component count)
- No duplicate CSS rules
- Console: No missing style warnings

---

## Browser Console Checklist

Open browser console and verify:

- [ ] **No Alpine.js errors**: No "Cannot read property" errors
- [ ] **No template errors**: No "Unexpected token" in expressions
- [ ] **No component errors**: No "Component not found" warnings
- [ ] **No style errors**: No CSS syntax errors
- [ ] **Alpine initialized**: "Alpine.js initialized" or similar message
- [ ] **x-data scope**: Use Alpine DevTools to inspect x-data values

---

## Visual Regression Checklist

Compare with expected layout:

- [ ] **Header**: Displays at top with title and user info
- [ ] **5 sections**: All sections visible with borders and padding
- [ ] **Product grid**: Grid layout (not list) for product cards
- [ ] **Tags**: Colored tags with correct background colors
- [ ] **Notifications**: Color-coded borders (blue, green, yellow)
- [ ] **Footer**: Displays at bottom with company info

---

## Debugging Steps (If Issues Found)

1. **Check server logs**: Look for parser/transformer errors
2. **View page source**: Verify Alpine.js directives generated correctly
3. **Browser console**: Check for JavaScript/Alpine errors
4. **Alpine DevTools**: Inspect x-data scope and reactive state
5. **Network tab**: Verify all component files loaded
6. **Elements inspector**: Check rendered DOM structure

---

## Test Results Summary

**Date Tested**: __________
**Tester**: __________
**Browser**: __________
**Server Version**: __________

### Overall Status

- [ ] ✅ All basic expressions working
- [ ] ✅ All conditionals working
- [ ] ✅ All loops working
- [ ] ✅ All components working
- [ ] ✅ All advanced features working
- [ ] ✅ Style aggregation working
- [ ] ✅ No console errors

### Issues Found

| Section | Issue | Severity | Status |
|---------|-------|----------|--------|
|         |       |          |        |

### Missing Patterns

List any patterns that should be tested but aren't currently covered:

1. Object loops (`{for key, value of object}`)
2. Complex array methods (`.reduce()`, chained `.filter().map()`)
3. Math functions (`Math.min()`, `Math.max()`)

---

## Future Enhancements

### Features to Add to Test Pages

**Object Loops** (not currently in jim-test):
- `{for key, value of settings}` example
- Nested object loops: `{for subKey, subValue of value}`
- Type checking: `{if typeof value === 'object'}`

**Complex Array Methods**:
- Array reduce: `products.reduce((sum, p) => sum + p.price, 0)`
- Array slicing in loops: `.slice(0, 3)`
- Chained methods: `.filter().map()`

**Math Functions**:
- `Math.min()`, `Math.max()`, `Math.floor()`

### Current Test Coverage Summary

| Feature Category | Test Page | Status |
|-----------------|-----------|--------|
| Basic Expressions | jim-test (greeting) | ✅ Complete |
| Conditionals | jim-test (greeting) | ✅ Complete |
| Loops with `of` | jim-test (animals, advanced) | ✅ Complete |
| Dynamic Components | jim-test (user_profiles) | ✅ Complete |
| Component Props | jim-test (age_examples) | ✅ Complete |
| String Methods | jim-test (animals) | ✅ Complete |
| Array Methods | jim-test (animals) | ✅ Complete |
| x-model Binding | jim-test (animals) | ✅ Complete |
| Global Stores | store-demo | ✅ Complete |
| Store Expressions | store-demo | ✅ Complete |
| Store Methods | store-demo | ✅ Complete |
| Store Computed Props | store-demo | ✅ Complete |
| Aggregate Pages | news_page, committee_page | ✅ Complete |
| Content Type Filtering | news_page, committee_page | ✅ Complete |
| Magic Variables | news_page, committee_page | ✅ Complete |

---

---

## Section 6: Aggregate Pages & Content Filtering

**Purpose**: Test Plenti-compatible aggregate/listing pages with content type filtering

### Test Cases

- [ ] **6.1 News aggregate page**: http://localhost:3333/news_page
  - Page renders with HTML wrapper (html.html)
  - All news posts display (3 expected)
  - Content filtered by `post.type === "news"`
  - Post titles, descriptions, dates visible
  - Links navigate to individual posts

- [ ] **6.2 Committee aggregate page**: http://localhost:3333/committee_page
  - Page renders with HTML wrapper
  - All committee posts display (3 expected)
  - Content filtered by `post.type === "committee"`
  - Meeting details visible

- [ ] **6.3 allContent injection**: Magic variable
  - `export let allContent` loads all site content
  - Content accessible via `{for post in allContent}`
  - Each post has `type`, `path`, `fields` properties

- [ ] **6.4 Type filtering in loops**:
  - `{if post.type === "news"}` correctly filters
  - Only matching content type renders
  - Non-matching content types excluded

- [ ] **6.5 Navigation links**:
  - News dropdown shows all news posts
  - Committee dropdown shows all committee meetings
  - Links work: `/news/product-launch`, `/committee/october-2025`, etc.

**Expected Pattern**:
```html
{for post in allContent}
  {if post.type === "news"}
    <article>
      <h3>{post.fields.title}</h3>
      <p>{post.fields.description}</p>
      <a href={post.path}>Read More</a>
    </article>
  {/if}
{/for}
```

---

## Section 7: Single Content Type Routes

**Purpose**: Test Plenti's `--single=true` pattern where JSON files in content root don't need _index.json

### Test Cases

- [ ] **7.1 Single content type registration**:
  - `content/news_page.json` → route `/news_page`
  - `content/committee_page.json` → route `/committee_page`
  - No folder required (vs `content/news/_index.json`)

- [ ] **7.2 Template matching**:
  - `news_page.json` uses `layouts/content/news_page.html`
  - Template name matches JSON filename (without extension)

- [ ] **7.3 HTML wrapper applied**:
  - `layouts/global/html.html` wraps content
  - `<head>`, `<body>`, navigation included
  - Alpine.js scripts injected

- [ ] **7.4 No route conflicts**:
  - Single content routes don't conflict with folder routes
  - Both `/news_page` and `/news/product-launch` work

**Key Files**:
- `cmd/server/main.go`: `registerSingleContentTypeRoutes()`
- `cmd/server/main.go`: `renderSingleContentTypePage()`

---

## Section 8: Global Store System

**Purpose**: Test Alpine.js global store integration

### Test Cases

- [ ] **8.1 Store demo page**: http://localhost:3333/store-demo
  - Auth store: login/logout toggles state
  - Cart store: add items, update total
  - Theme store: toggle light/dark mode

- [ ] **8.2 Inline store definition**:
  ```html
  store cart = {
    items: [],
    total: 0,
    addItem(item) { this.items.push(item); this.total += item.price; }
  }
  ```
  - Store initializes correctly
  - Methods callable via `$store.cart.addItem()`

- [ ] **8.3 Store expressions**:
  - `{$auth.isLoggedIn}` → `$store.auth.isLoggedIn`
  - `{if $auth.isLoggedIn}` → `<template x-if="$store.auth.isLoggedIn">`
  - `@click="$store.auth.login()"` preserved

- [ ] **8.4 Computed properties (getters)**:
  - `{$cart.formattedTotal}` displays formatted currency
  - Getters recalculate when dependencies change

- [ ] **8.5 External store files**:
  - `import store from './stores/auth.js'`
  - Store loaded from external file
  - Multiple stores can be imported

**Key Files**:
- `stores/auth.js`, `stores/cart.js`, `stores/theme.js`
- `transformer/stores.go`: Store expression transformation
- `renderer/stores.go`: Store initialization rendering

---

## Next Steps

After completing this checklist:

1. ✅ **Update checklist**: Document current feature coverage (DONE 2026-01-28)
2. ✅ **Plenti compatibility**: Feature parity matrix added
3. ✅ **Aggregate pages**: News and committee listings working
4. ✅ **Single content types**: Route registration implemented
5. ✅ **Global stores**: Full store system operational
6. **Future**: Additional content types (portfolio, blog, etc.)
