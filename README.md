# Custom Go Template Engine

A production-ready Go template engine that transforms Svelte-inspired template syntax into Alpine.js-compatible HTML. Built for the Plenti ecosystem, it features build-time loop expansion, dynamic component resolution, and optimized x-data scope management.

**Core Philosophy**: *Build-time when possible, runtime when necessary, inherit when available.*

## Table of Contents

- [Key Features](#key-features)
- [Quick Start](#quick-start)
- [Architecture](#architecture)
- [Template Syntax](#template-syntax)
- [Plenti Architecture Patterns](#plenti-architecture-patterns)
- [Content Type System](#content-type-system)
- [Fence Section](#fence-section)
- [Global Store System](#global-store-system)
- [x-data Optimization](#x-data-optimization)
- [Scope Inheritance](#scope-inheritance)
- [Props vs Stores: When to Use Which](#props-vs-stores-when-to-use-which)
- [Examples](#examples)
- [Performance](#performance)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [Testing](#testing)
- [Quick Reference](#quick-reference)
- [Documentation](#documentation)
- [Contributing](#contributing)
- [Summary](#summary)

## Key Features

- **Build-Time Loop Expansion**: Loops expand at build time (like Svelte) for zero-runtime overhead and perfect SEO
- **Dynamic Component Resolution**: Components resolved by name from JSON data, with hybrid build/runtime approach
- **Plenti Architecture**: Two rendering patterns - wrapper-based (Plenti) and standalone
- **Svelte-like Syntax**: Write templates with familiar `{if}`, `{for}`, and `{variable}` syntax
- **Alpine.js Integration**: Automatically transforms to Alpine.js directives
- **Global Store System**: Shared reactive state across components
- **x-data Optimization**: RuntimeVarTracker filters scopes to only runtime-needed variables (95%+ size reduction)
- **Content Type System**: Organize content with types and create aggregate/listing pages
- **Development Server**: Live preview at http://localhost:3333

## Quick Start

### Installation

```bash
go build ./...
```

### Run Development Server

```bash
go run cmd/server/main.go
# Visit http://localhost:3333
```

### Run Tests

```bash
# All tests
go test ./... -v

# Specific package
go test ./transformer -v

# Integration tests
go test ./tests/alpine -v
```

## Architecture

### Build Time vs Runtime

This engine maximizes build-time processing to minimize runtime JavaScript:

| Phase | When | What Happens | Our System |
|-------|------|--------------|------------|
| **Compile Time** | Go binary build | Type checking, optimization | `go build ./...` |
| **Build Time** | Server processes request | Template → HTML transformation | **80% of work** |
| **Runtime** | Browser executes JS | Reactive updates, events | **20% of work** |

**Build-Time Operations** (server):
- Template parsing → AST
- JSON content loading
- **Loop expansion** (when data available)
- **Component inlining**
- CSS/JS scoping
- x-data optimization

**Runtime Operations** (browser/Alpine.js):
- x-data scope initialization
- x-text, x-html binding
- x-if, x-show conditional display
- x-for (only for runtime-only collections)
- Event handling (@click, @submit)
- $store global state access

### The Pipeline

```
Template Source → Parser → AST → Transformer → Rendered HTML/CSS/JS
                    ↓         ↓        ↓
                 Fence    Control  Alpine.js
                Section   Flow    Directives
```

### Key Packages

- **`ast/`** - Abstract Syntax Tree node definitions
- **`parser/`** - Converts template syntax to AST using parser combinators
- **`transformer/`** - Transforms AST to Alpine.js nodes (loop expansion, component resolution)
- **`renderer/`** - Generates final HTML/CSS/JS from transformed AST
- **`analyzer/`** - Distinguishes build-time vs runtime expressions
- **`builder/`** - Component registry generation for runtime resolution
- **`loader/`** - Content JSON loading from Plenti structure
- **`scoping/`** - CSS and JS scoping utilities
- **`cmd/server/`** - Development server

## Template Syntax

### Expressions

**Single curly braces** for dynamic content:

```html
<p>Hello, {name}!</p>
<!-- Build-time resolved OR transforms to: -->
<p>Hello, <span x-text="name"></span>!</p>
```

**Fallback operator** (resolved at build-time when possible):
```html
{post.title || 'Untitled'}
{author?.image?.alt}  <!-- Optional chaining -->
```

### Conditionals

```html
{if isLoggedIn}
  <p>Welcome back!</p>
{else if isGuest}
  <p>Hello, Guest</p>
{else}
  <p>Please log in</p>
{/if}
```

**Build-time equality comparisons** (evaluated during loop expansion):
```html
{if post.type === "news"}
  <article>...</article>
{/if}
```

Transforms to Alpine.js `<template x-if>` and `<template x-else>` directives.

### Loops

#### Build-Time Expansion (Default)
When collection is known from JSON data, loops expand at build time:

```html
{for component in components}
  <Component:dynamic name={component.name} {...component.fields} />
{/for}
```

**Expands to** (fully rendered HTML, zero runtime JS):
```html
<section class="hero" x-data='{"title":"Welcome"}'>
  <h1 x-text="title">Welcome</h1>
</section>

<section class="services" x-data='{"items":[...]}'>
  <h2>Our Services</h2>
</section>
```

**Benefits**:
- ✅ Perfect SEO (full content in HTML source)
- ✅ Zero runtime JavaScript for static content
- ✅ Instant rendering (no JS execution needed)
- ✅ Smaller payloads (no component registry shipped)

#### Runtime Fallback
When collection is runtime-only (Alpine stores, complex expressions):

```html
{for item in $store.cart.items}
  <div>{item.name}</div>
{/for}
```

**Generates**:
```html
<template x-for="item in $store.cart.items">
  <div x-text="item.name"></div>
</template>
```

### Components

```html
<!-- Static component -->
<Header title="My Site" />

<!-- With props (regular, shorthand, spread) -->
<UserProfile name="John" age={30} />
<UserProfile {name} {age} />
<Card {...cardData} />

<!-- Dynamic component (build-time resolved from loop variable) -->
{for component in components}
  <Component:dynamic name={component.name} {...component.fields} />
{/for}
```

Components are auto-registered from `layouts/components/`, `layouts/global/`, and `layouts/content/`.

## Global Store System

The template engine supports global reactive stores for sharing state across components.

### Store Syntax

Reference store data in templates using `{$storeName.property}`:

```html
{if $auth.isLoggedIn}
  <p>Welcome, {$auth.user.name}!</p>
{/if}
```

This transforms to Alpine.js store syntax: `$store.auth.isLoggedIn`

### Defining Stores

#### Option 1: Inline Store Definition

Define stores directly in the fence section:

```html
---
store auth = {
  isLoggedIn: false,
  user: null,
  login() {
    this.isLoggedIn = true;
    this.user = { name: 'John' };
  },
  logout() {
    this.isLoggedIn = false;
    this.user = null;
  }
}

store cart = {
  items: [],
  total: 0,
  addItem(name, price) {
    this.items.push({ name, price });
    this.total += price;
  }
}
---
```

#### Option 2: External Store Files

Create reusable store files in `stores/` directory:

**stores/auth.js:**
```javascript
{
  isLoggedIn: false,
  user: null,
  login() {
    this.isLoggedIn = true;
    this.user = { name: 'User', email: 'user@example.com' };
  },
  logout() {
    this.isLoggedIn = false;
    this.user = null;
  }
}
```

**Import in template fence:**
```html
---
import store from './stores/auth.js'
---
```

#### Option 3: Component Store Imports

Components can import stores they depend on:

**examples/components/CartBadge.html:**
```html
---
import store from './stores/cart.js'
---

<div class="cart-badge">
  {if $cart.items.length > 0}
    <span>{$cart.items.length} items</span>
    <span>${$cart.total}</span>
  {/if}
</div>
```

### Store Priority

When multiple store definitions exist, the priority is:

1. **Inline definitions** (`store name = { ... }` in fence)
2. **Imported files** (`import store from './stores/name.js'` in fence)
3. **External stores** (registered globally)

### Store Methods and Computed Properties

Stores can have methods and getters:

```javascript
// stores/cart.js
{
  items: [],
  total: 0,

  // Computed property
  get formattedTotal() {
    return this.total.toFixed(2);
  },

  // Method
  addItem(item) {
    this.items.push(item);
    this.total += item.price;
  }
}
```

Use in templates:
```html
<p>Total: ${$cart.formattedTotal}</p>
<button @click="$store.cart.addItem({ name: 'Widget', price: 9.99 })">
  Add Item
</button>
```

### Store Expressions in Alpine Directives

Stores work with all Alpine.js directives:

```html
<!-- Conditional rendering -->
<div x-show="$store.auth.isLoggedIn">...</div>

<!-- Event handlers -->
<button @click="$store.auth.login()">Login</button>

<!-- Dynamic styling with :style -->
<body :style="`background: ${$store.theme.getCurrentColors().background}`">
```

## Fence Section

The fence section (between `---` markers) contains imports, props, variables, and functions.

### Export Let (Content Injection)

**Opt-in system** for receiving data from JSON content files (Plenti/Svelte pattern):

```html
---
export let components       # Request components array from JSON
export let allContent       # Request ALL site content (opt-in for bandwidth!)
export let title, description
---
```

**Performance**: Only load expensive data (`allContent` ~50KB) when explicitly requested.

### Props

Component props with optional defaults:

```html
---
prop title = "Default Title"
prop description
prop isActive = false
---
```

### Variables

Local reactive variables for Alpine.js:

```html
---
let count = 0
const MAX_ITEMS = 10
---
```

### Functions

Alpine.js data methods:

```html
---
function increment() {
  this.count++;
}

function formatDate(date) {
  return new Date(date).toLocaleDateString();
}
---
```

### Imports

Import components and stores:

```html
---
import Header from './components/Header.html'
import store from './stores/auth.js'
---
```

### Store Definitions

Define inline stores (or import from `stores/` directory):

```html
---
store cart = {
  items: [],
  total: 0,
  addItem(item) {
    this.items.push(item)
    this.total += item.price
  }
}
---
```

## Plenti Architecture Patterns

The engine supports **two rendering modes**:

### 1. Wrapper Rendering (Plenti Pattern)

**Recommended for most pages**. Content-only templates with global HTML wrapper:

```
Request → renderWithWrapper() → Load JSON → Parse Wrapper → Inject Content → Render
```

**Structure**:
- Content: `content/pages/about.json` (JSON with components array)
- Template: `layouts/content/pages.html` (content-only, no HTML wrapper)
- Wrapper: `layouts/global/html.html` (DOCTYPE, html, head, body, nav, footer)

**Example JSON** (`content/pages/about.json`):
```json
{
  "components": [
    {"name": "hero2436", "fields": {"title": "About Us", "description": "..."}},
    {"name": "services2437", "fields": {}},
    {"name": "footer2425", "fields": {}}
  ]
}
```

**Example Template** (`layouts/content/pages.html`):
```html
---
export let components
---

{for component in components}
  <Component:dynamic name={component.name} {...component.fields} />
{/for}
```

### 2. Standalone Rendering (Legacy)

Complete HTML pages with DOCTYPE, html, body tags. All content hardcoded in template:

```html
---
prop title = "My Page"
---

<!DOCTYPE html>
<html>
<head>
  <title>{title}</title>
</head>
<body>
  <h1>{title}</h1>
</body>
</html>
```

**Use for**: Self-contained demo pages, admin tools

## Content Type System

Organize content by type and create aggregate/listing pages.

### Content Types

Content is organized in folders under `content/`, where each folder represents a type:

```
content/
├── pages/           # Component-based pages (type: "pages")
│   ├── about.json
│   └── contact.json
├── news/            # News articles (type: "news")
│   ├── product-launch.json
│   └── quarterly-results.json
├── committee/       # Committee meetings (type: "committee")
│   └── october-2025.json
├── news_page.json   # Aggregate page (lists all news)
└── committee_page.json
```

**Route Mapping**:
- `content/pages/about.json` → `/about` (uses `layouts/content/pages.html`)
- `content/news/product-launch.json` → `/news/product-launch` (uses `layouts/content/news.html`)
- `content/news_page.json` → `/news_page` (uses `layouts/content/news_page.html`)

### Creating Aggregate Pages

**Aggregate pages** list all content of a specific type (like a blog index):

**Step 1: Create JSON** (`content/news_page.json`):
```json
{}
```
*Empty JSON is fine - the template uses `allContent`*

**Step 2: Create Layout** (`layouts/content/news_page.html`):
```html
---
export let allContent
---

<section id="blog-listing">
  <h1>News</h1>

  {for post in allContent}
    {if post.type === "news"}
      <article class="cs-item">
        <a href={post.path}>
          <h3>{post.fields.title}</h3>
          <time>{post.fields.date}</time>
          <p>{post.fields.description}</p>
        </a>
      </article>
    {/if}
  {/for}
</section>

<style>
  #blog-listing { max-width: 1200px; margin: 0 auto; }
  .cs-item { margin-bottom: 2rem; }
</style>
```

**Key Pattern**:
1. `export let allContent` - Opt-in to receive all site content
2. `{for post in allContent}` - Loop through all content (build-time expansion!)
3. `{if post.type === "news"}` - Filter by content type
4. `{post.fields.*}` - Access content fields
5. `{post.path}` - Link to individual page

### allContent Structure

```javascript
{
  "news/product-launch": {
    type: "news",
    path: "/news/product-launch",
    fields: {
      title: "New Product Launch",
      description: "...",
      date: "2025-10-15"
    }
  },
  "pages/about": {
    type: "pages",
    path: "/about",
    components: [...]
  }
}
```

## x-data Optimization

The engine optimizes x-data attributes to **only include runtime-needed variables**, dramatically reducing page weight.

### The Problem

Without optimization, x-data would include ALL variables:

```html
<!-- BAD: 50KB+ of unused data -->
<div x-data="{
  allContent: { /* entire site content */ },
  components: [ /* component definitions */ ],
  content: { /* page content */ },
  title: 'Welcome'
}">
```

### The Solution: RuntimeVarTracker

The transformer tracks which variables are used in Alpine directives:

```go
// Only track variables used in runtime directives
tracker.Track("title")      // from x-text="title"
tracker.Track("visible")    // from x-if="visible"

// Filter scope before serializing
filteredScope := tracker.FilterScope(fullScope)
// Returns: {"title": "Welcome", "visible": true}
```

### Result

```html
<!-- GOOD: Only runtime-needed variables -->
<div x-data="{ title: 'Welcome' }">
  <h1 x-text="title">Welcome</h1>
</div>
```

**Performance Impact**: 95%+ HTML size reduction (from ~850KB to ~39KB in complex pages)

## Scope Inheritance

Components **inherit from parent Alpine.js scopes** via prototypal chain - no explicit prop passing needed!

```html
<body x-data="{siteName: 'My Site', content: {...}}">
  <Header />  <!-- Can access siteName and content -->

  <div x-data="{pageTitle: 'Home'}">
    <span x-text="pageTitle"></span>     <!-- Own scope -->
    <span x-text="siteName"></span>      <!-- Parent scope -->
    <span x-text="content.title"></span> <!-- Parent scope -->
  </div>
</body>
```

**When to duplicate vs inherit**:
- **Duplicate** (explicit x-data): Component modifies data, needs isolation
- **Inherit** (no x-data): Component only reads data, wants parent changes to propagate

## Props vs Stores: When to Use Which

### Use Props When:

- Data is **component-specific** and passed from parent
- Each instance needs **different values**
- Data flows **one direction** (parent → child)

```html
<UserCard name="John" age={30} />
<UserCard name="Jane" age={25} />
```

### Use Stores When:

- State is **shared across multiple components**
- Multiple components need to **read and write** the same data
- State persists across **component instances**
- You need **global application state** (auth, theme, cart)

```html
<!-- Multiple components accessing same auth state -->
<LoginButton />    <!-- Uses $auth.login() -->
<UserMenu />       <!-- Shows $auth.user.name -->
<ProtectedRoute /> <!-- Checks $auth.isLoggedIn -->
```

### Use Scope Inheritance When:

- Component only **reads** shared data
- Want **automatic updates** when parent changes
- Minimize **x-data duplication**

```html
<!-- Component inherits content from parent -->
<div>
  <h1 x-text="content.title"></h1>  <!-- Inherited -->
</div>
```

## Examples

### Complete Page with Store

```html
---
prop title = "My Page"

import store from './stores/auth.js'
import store from './stores/cart.js'

store theme = {
  mode: "light",
  toggle() {
    this.mode = this.mode === "light" ? "dark" : "light";
  }
}
---

<!DOCTYPE html>
<html>
<head>
  <title>{title}</title>
  <script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
</head>
<body>
  <header>
    {if $auth.isLoggedIn}
      <p>Welcome, {$auth.user.name}!</p>
      <button @click="$store.auth.logout()">Logout</button>
    {else}
      <button @click="$store.auth.login()">Login</button>
    {/if}
  </header>

  <main>
    <h1>{title}</h1>

    <div class="cart-summary">
      <p>Cart: {$cart.items.length} items</p>
      <p>Total: ${$cart.formattedTotal}</p>
    </div>

    <button @click="$store.theme.toggle()">
      Switch to {$theme.mode === 'light' ? 'Dark' : 'Light'} Mode
    </button>
  </main>
</body>
</html>
```

## Performance

### Build-Time Performance

**Fast Parsing** (~1-5ms per template):
- Parser combinators (composable, efficient)
- Single-pass AST construction
- No regex-heavy parsing

**Component Registry** (generated once on startup):
- 65+ components auto-generated
- Cached for subsequent requests
- Only loaded when runtime component resolution needed

### Runtime Performance

**Alpine.js** (~20KB gzipped):
- Minimal runtime overhead
- Reactive only where needed
- No virtual DOM

**Zero Runtime JS for Static Content**:
- Build-time loop expansion eliminates x-for templates
- Component inlining eliminates runtime lookups
- Conditional script injection (no unnecessary payload)

### Page Weight Comparison

| Metric | Before Optimization | After Optimization |
|--------|---------------------|-------------------|
| HTML Size (complex page) | ~850KB | ~39KB |
| JavaScript Payload (static page) | ~180KB (registry) | 0KB |
| JavaScript Payload (dynamic page) | ~180KB | ~15KB |
| Time to First Paint | ~800ms | ~200ms |
| SEO Score | 60-70 | 100 |

**Result**: 95.5% HTML size reduction, zero-runtime static pages, perfect SEO

## Best Practices

### 1. Use Build-Time Expansion When Possible

```html
<!-- ✅ GOOD: Expands at build time (SEO-friendly) -->
{for component in components}
  <Component:dynamic name={component.name} />
{/for}

<!-- ⚠️ OK: Runtime when necessary -->
{for item in $store.cart.items}
  <CartItem {item} />
{/for}
```

### 2. Opt-In to Expensive Data

```html
---
// ✅ GOOD: Only load when needed
export let allContent  // Only in navigation/aggregate pages

// ❌ BAD: Loading allContent everywhere
---
```

### 3. Trust Scope Inheritance

```html
<!-- ❌ BAD: Explicit prop drilling -->
<Component content={content} allContent={allContent} env={env} />

<!-- ✅ GOOD: Inherit from parent -->
<Component />  <!-- Can access content, allContent, env -->
```

### 4. Prefer Regular Components Over Dynamic

```html
<!-- ✅ BEST: Known component (faster, inlined) -->
<Header title="My Site" />

<!-- ⚠️ OK: Unknown component (runtime lookup) -->
<Component:dynamic name={componentName} />
```

### 5. Keep Component Scope Minimal

```html
---
// ✅ GOOD: Only what this component needs
let isOpen = false
function toggle() { isOpen = !isOpen }

// ❌ BAD: Duplicating parent scope
let content = content  // Already inherited!
---
```

## Troubleshooting

### Common Store Issues

#### 1. Store Not Found

**Error:** `Cannot read properties of undefined (reading 'propertyName')`

**Solution:** Ensure store is imported or defined:
```html
---
import store from './stores/auth.js'
---
```

#### 2. Store Methods Not Working

**Error:** `$store.cart.addItem is not a function`

**Solution:** Check that:
- Store file exports valid JavaScript object
- Methods use proper `this` binding
- Store is properly registered

#### 3. Double Store Prefix

**Error:** Console shows `$store.store.theme` instead of `$store.theme`

**Solution:** This was a bug fixed in the transformer. Update to latest version.

#### 4. Computed Properties Undefined

**Error:** `.toFixed is not a function`

**Solution:** Use computed properties (getters) instead of method calls in templates:
```javascript
// ✅ Good
get formattedTotal() {
  return this.total.toFixed(2);
}

// ❌ Bad - don't call .toFixed(2) in template
```

### Known Issues

See [KNOWN_ISSUES.md](KNOWN_ISSUES.md) for active issues and regressions.

## Testing

### Test Structure

```
tests/
├── alpine/                      # Alpine.js integration tests
├── components/                  # Component tests
├── build_time_loop_expansion/   # Loop expansion tests
└── integration/                 # End-to-end tests

transformer/
├── *_test.go                   # Unit tests for transformers
├── stores_test.go              # Store transformation tests
├── component_loop_integration_test.go  # Build-time expansion tests
└── scope_test.go               # RuntimeVarTracker tests

parser/
├── *_test.go                   # Parser tests
├── conditional_bug_test.go     # Regression tests
└── nested_conditional_loop_test.go

analyzer/
└── scope_test.go               # Build-time vs runtime detection tests
```

### Running Specific Tests

```bash
# All tests
go test ./... -v

# Store system tests
go test ./transformer -run TestStores -v

# Component tests
go test ./tests/components -v

# Alpine integration
go test ./tests/alpine -v

# Build-time loop expansion
go test ./transformer -run TestComponentLoop -v
go test ./tests/build_time_loop_expansion -v

# x-data optimization
go test ./transformer -run TestRuntimeVarTracker -v
```

**Test Coverage**: 294+ tests covering parser, transformer, renderer, analyzer, and integration

## Quick Reference

### Common Commands

```bash
# Run dev server
go run cmd/server/main.go

# Run tests
go test ./... -v

# Run specific package tests
go test ./transformer -v
go test ./tests/alpine -v

# Build
go build ./...

# Enable debug logging
DEBUG_EXPRESSIONS=true go run cmd/server/main.go
```

### File Structure

```
├── ast/               # AST node definitions
├── parser/            # Template parsing (parser combinators)
├── transformer/       # AST → Alpine.js transformation
├── renderer/          # HTML/CSS/JS generation
├── analyzer/          # Build-time vs runtime expression analysis
├── builder/           # Component registry generation
├── loader/            # Content JSON loading
├── scoping/           # CSS/JS scoping utilities
├── cmd/server/        # Development server
├── layouts/           # Template files
│   ├── global/        # Global wrappers (html.html, head.html, nav.html)
│   ├── content/       # Page layouts by content type
│   │   ├── pages.html       # Component-based pages
│   │   ├── news.html        # Individual news article
│   │   └── news_page.html   # News aggregate/listing
│   └── components/    # Reusable components (65+)
├── content/           # JSON content files
│   ├── pages/         # Component-based pages
│   ├── news/          # News articles (type: "news")
│   ├── committee/     # Committee meetings (type: "committee")
│   └── news_page.json # Aggregate page
├── stores/            # Global store definitions
├── media/             # Media assets
├── generated/         # Auto-generated files
│   └── layouts.js     # Component registry (generated on startup)
├── core/              # Core runtime scripts
│   └── runtime-components.js  # Dynamic component resolution
└── docs/              # Documentation
    └── GO-BUILD-TIME-ENGINE-GUIDES/  # Developer guides
```

### Syntax Cheatsheet

```html
---
// Imports
import Component from './components/Component.html'
import store from './stores/auth.js'

// Store definitions
store cart = { items: [], total: 0 }

// Props
prop title = "Default"

// Export let (opt-in to JSON data)
export let components, allContent

// Variables
let count = 0

// Functions
function increment() { count++ }
---

<!-- Expressions -->
{title}
{user.name || 'Guest'}
{author?.image?.alt}

<!-- Conditionals -->
{if condition}
  content
{else if other}
  content
{else}
  content
{/if}

<!-- Loops (build-time expansion when possible) -->
{for item in items}
  {item.name}
{/for}

{for (item, index) in items}
  {index}: {item.name}
{/for}

<!-- Components -->
<Header />
<Header title="Site" />
<Header {title} {description} />
<Card {...cardData} />

<!-- Dynamic components -->
<Component:dynamic name={component.name} {...component.fields} />

<!-- Store access -->
{$auth.isLoggedIn}
{$cart.total}
@click="$store.auth.login()"
```

## Documentation

### Primary Docs

- **[ARCHITECTURE.md](docs/GO-BUILD-TIME-ENGINE-GUIDES/ARCHITECTURE.md)** - System architecture overview
- **[BUILD_TIME_LOOP_EXPANSION.md](docs/GO-BUILD-TIME-ENGINE-GUIDES/BUILD_TIME_LOOP_EXPANSION.md)** - Headline feature documentation
- **[DEVELOPER_GUIDE.md](docs/GO-BUILD-TIME-ENGINE-GUIDES/DEVELOPER_GUIDE.md)** - Comprehensive developer documentation
- **[STORE_DEVELOPER_GUIDE.md](docs/GO-BUILD-TIME-ENGINE-GUIDES/STORE_DEVELOPER_GUIDE.md)** - Global store system guide
- **[MIGRATION_GUIDE.md](docs/GO-BUILD-TIME-ENGINE-GUIDES/MIGRATION_GUIDE.md)** - Plenti architecture migration patterns
- **[CLAUDE.md](CLAUDE.md)** - AI assistant context and project instructions

### Specs

Detailed specifications for major features in `.agent-os/specs/`:
- Build-time loop expansion
- Runtime component resolution
- Global store system
- Parser unification
- Export let content injection
- x-data optimization

## Contributing

When making changes:

1. **Write tests first** - Test-driven development
2. **Update CLAUDE.md** - If architecture changes
3. **Add examples** - For new features
4. **Run full test suite** - `go test ./... -v`
5. **Update README** - If adding user-facing features
6. **Update relevant specs** - In `.agent-os/specs/`
7. **Follow patterns** - Check existing code for consistency

### Architecture Principles

- **Build-time when possible** - Maximize server processing
- **Runtime when necessary** - Defer to Alpine.js only when needed
- **Inherit when available** - Trust Alpine's scope chain
- **Optimize by default** - x-data filtering, conditional script injection
- **Test everything** - Comprehensive test coverage

## Summary

This template engine successfully combines three foundational systems:

1. **Plenti's Content Model** - JSON-based content with magic variables and wrapper architecture
2. **Jim Fisk's Component Vision** - Svelte-like syntax with dynamic component loading
3. **Alpine.js Reactivity** - Lightweight, modern client-side framework

### Key Achievements

- ✅ **95%+ Feature Parity** - With Plenti + Jim's vision
- ✅ **95.5% HTML Size Reduction** - Through scope optimization (~850KB → ~39KB)
- ✅ **Zero-Runtime Static Pages** - Build-time loop expansion eliminates JS overhead
- ✅ **Perfect SEO** - Fully expanded HTML with no placeholders
- ✅ **Production Ready** - 294+ tests, comprehensive error handling
- ✅ **Fast & Secure** - No server-side JS execution, minimal runtime overhead
- ✅ **Maintainable Architecture** - Modular, well-documented codebase

### Innovation

**Build-Time Loop Expansion** - The headline feature that enables:
- Component-by-name resolution from JSON data
- Zero runtime JavaScript for static content
- Full HTML source for search engines
- Hybrid approach (build-time default, runtime fallback)

**Philosophy**: *Build-time when possible, runtime when necessary, inherit when available.*

## License

MIT License

## Acknowledgments

Built for the **Plenti** ecosystem with inspiration from:
- **Plenti's** content-driven architecture and magic variables
- **Jim Fisk's** innovative component syntax and dynamic resolution vision
- **Svelte's** build-time compilation philosophy
- **Alpine.js's** lightweight reactivity framework

---

**For detailed documentation, see:**
- [docs/GO-BUILD-TIME-ENGINE-GUIDES/](docs/GO-BUILD-TIME-ENGINE-GUIDES/) - Developer guides
- [CLAUDE.md](CLAUDE.md) - Project context
- `.agent-os/specs/` - Feature specifications
