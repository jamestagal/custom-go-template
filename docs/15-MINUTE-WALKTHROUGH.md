# Custom Go Template Engine: 15-Minute Technical Walkthrough

**For:** Lead Developer Onboarding
**Duration:** 15 minutes
**Last Updated:** 2026-01-29

---

## 1. What Is This? (2 minutes)

### The Elevator Pitch

A **Go-based template engine** that transforms **Svelte-inspired syntax** into **Alpine.js-compatible HTML** with a unique twist: **build-time loop expansion**.

```
Svelte Developer Experience + Alpine.js Reactivity - Runtime Overhead = This Engine
```

### Core Innovation: Build-Time Loop Expansion

**The Problem:** Most template engines ship all component templates to the browser and resolve dynamic components at runtime with JavaScript.

**Our Solution:** Resolve loops and components during template transformation on the server, shipping only fully-expanded static HTML.

**Impact:**
- ✅ Zero JavaScript for static content
- ✅ Perfect SEO (full HTML in source)
- ✅ Instant page loads (no JS execution needed)
- ✅ Smaller payloads (only what's needed)

---

## 2. Architecture Overview (3 minutes)

### The Pipeline

```
Template Source → Parser → AST → Transformer → Renderer → HTML
                                      ↓
                                 JSON Content
```

### Build Time vs Runtime

| Phase | When | Who | What |
|-------|------|-----|------|
| **Build Time** | Server processes request | Go engine | Parse, expand loops, resolve components, inject content |
| **Runtime** | User views page | Alpine.js | Reactive binding, events, dynamic updates |

**Philosophy:** Do 80% of work at build time, 20% at runtime.

### Key Packages

1. **`ast/`** - Abstract Syntax Tree node definitions
2. **`parser/`** - Template → AST conversion (unified parser architecture)
3. **`transformer/`** - AST → Alpine.js transformation
4. **`renderer/`** - AST → HTML generation
5. **`loader/`** - JSON content loading from `content/` directory
6. **`analyzer/`** - Build-time vs runtime expression detection
7. **`builder/`** - Component registry generation for runtime fallback

---

## 3. Template Syntax (3 minutes)

### Expressions

```html
{variable}                    → <span x-text="variable"></span>
{post.title || 'Untitled'}    → Resolved at build-time when possible
{author?.name}                → Optional chaining supported
```

### Conditionals

```html
{if published}
  <article>{content}</article>
{else}
  <p>Draft</p>
{/if}
```

Transforms to `<template x-if>` when runtime evaluation needed, or expands at build-time when condition is known.

### Loops (The Headline Feature!)

```html
{for component in components}
  <Component:dynamic name={component.name} {...component.fields} />
{/for}
```

**Build-time expansion** (when `components` array is available):
```html
<!-- Output: Fully expanded HTML -->
<section class="hero">
  <h1>Welcome</h1>
</section>

<section class="services">
  <div class="grid">...</div>
</section>
```

**Runtime fallback** (when collection is only known at runtime):
```html
<!-- Output: Alpine x-for template -->
<template x-for="item in $store.cart.items">
  <div x-text="item.name"></div>
</template>
```

### Components

```html
<ComponentName prop1="value" prop2={dynamicValue} />
```

Components are inlined at build-time with proper prop passing and CSS/JS scoping.

### Global Stores (Alpine.js)

```html
{$auth.isLoggedIn}           → <span x-text="$store.auth.isLoggedIn"></span>
{if $theme.mode === 'dark'}  → <template x-if="$store.theme.mode === 'dark'">
```

Store definitions can be inline or imported from `stores/` directory.

---

## 4. Content Patterns (3 minutes)

### Two Distinct Patterns

#### Pattern 1: Component-Based Pages (Plenti Collection Type)

**Template:** `layouts/content/pages.html`
```html
---
export let components
---

{for component in components}
  <Component:dynamic name={component.name} {...component.fields} />
{/for}
```

**JSON:** `content/pages/about.json`
```json
{
  "components": [
    {"name": "hero2436", "fields": {"title": "About Us"}},
    {"name": "services2437", "fields": {}}
  ]
}
```

**Key:** Generic template that loops through components array. Each component is inlined at build-time.

#### Pattern 2: Custom Template Pages (Plenti Single Type)

**Template:** `layouts/content/news.html`
```html
---
export let title, author, publish, textItems
---

<article>
  <h1>{title}</h1>
  <span>{author?.name}</span>

  {for item in textItems}
    <section>
      <h4>{item.title}</h4>
      <p>{item.paragraph}</p>
    </section>
  {/for}
</article>
```

**JSON:** `content/news/post1.json`
```json
{
  "title": "Breaking News",
  "author": {"name": "John Doe"},
  "textItems": [
    {"title": "Intro", "paragraph": "..."}
  ]
}
```

**Key:** Custom layout per content type with flat JSON structure and `export let` for prop injection.

### Export Let System

Similar to Svelte's prop system - declare what content fields the template expects:

```html
---
export let title, description, images
---
```

Content is automatically injected from JSON files based on route path.

---

## 5. The Magic: How Build-Time Loop Expansion Works (2 minutes)

### Decision Tree

```
{for item in items}
       ↓
Is "items" available in dataScope (from JSON)?
       ↓
   YES ━━━━━━━━━━━━━━━━ NO
    ↓                    ↓
BUILD-TIME          RUNTIME
EXPANSION          FALLBACK
    ↓                    ↓
Loop through       Generate
actual array       x-for template
    ↓                    ↓
Expand each        Let Alpine
iteration          handle at runtime
    ↓                    ↓
Static HTML        Dynamic template
```

### Example Transformation

**Input:**
```html
{for post in allContent}
  {if post.type === "news"}
    <div>{post.fields.title}</div>
  {/if}
{/for}
```

**Build-Time Output** (when allContent available):
```html
<div>Breaking News Story</div>
<div>Product Launch</div>
<div>Team Update</div>
```

**Runtime Output** (when collection is $store.items):
```html
<template x-for="item in $store.items">
  <div x-text="item.title"></div>
</template>
```

### x-data Optimization

The `RuntimeVarTracker` filters x-data to only include variables needed at runtime, dramatically reducing page weight:

**Before optimization:**
```html
<div x-data="{ allContent: [...50KB], photos: [], title: '...' }">
```

**After optimization:**
```html
<div x-data="{ photos: [] }">
```

Only `photos` is tracked because it's used in a runtime `x-for`. Build-time-only vars are excluded.

---

## 6. Key Technical Decisions (1 minute)

### Parser Architecture

**Unified Single-Path Parsing** (fixed Oct 2025):
- `BlockConditionalParser` and `BlockLoopParser` with depth tracking
- Deprecated: Old marker-based post-processing approach
- Eliminates bugs where content after `{/if}` was incorrectly consumed

### Transformer Pipeline

1. **Parse** template to AST
2. **Extract** fence data (props, variables, stores)
3. **Transform** AST nodes to Alpine.js equivalents
4. **Build** x-data scope from discovered variables
5. **Filter** scope to runtime-only variables
6. **Render** final HTML with Alpine directives

### Alpine.js Integration Fixes

Recent fix (Jan 2026): Preserve Alpine object literal syntax:
```html
:class="{'active': isActive}"  ← Must NOT be transformed as template expression
```

Added `isAlpineObjectLiteral()` detector to prevent breaking Alpine directives.

---

## 7. File Organization (1 minute)

```
custom_go_template/
├── ast/              # AST node definitions
├── parser/           # Template → AST
├── transformer/      # AST → Alpine.js
├── renderer/         # AST → HTML
├── loader/           # JSON content loading
├── analyzer/         # Runtime expression detection
├── builder/          # Component registry generation
├── scoping/          # CSS/JS scoping utilities
├── cmd/server/       # Development server (port 3333)
├── layouts/
│   ├── global/       # HTML wrapper, nav, footer, head
│   ├── content/      # Content type templates
│   └── components/   # Reusable components
├── content/
│   ├── pages/        # Page content JSON
│   ├── news/         # News posts
│   └── committee/    # Committee posts
├── scripts/          # Client-side JS (cms.js, script.js)
└── styles/           # CSS files
```

---

## 8. Common Patterns & Best Practices (1 minute)

### Build-Time Friendly

✅ **DO:**
- Use simple conditionals: `{if photos}` (resolved at build-time)
- Access known arrays directly: `{for post in allContent}`
- Use optional chaining: `{author?.name}`
- Keep expressions simple in loops

❌ **DON'T:**
- Use complex conditionals: `{if photos && photos.length > 0}` (may fall back to runtime)
- Use bracket notation: `{items[0].title}` (use loops instead)
- Rely on JavaScript operators that can't be evaluated at build-time

### Content Organization

**Listing Pages:**
- Use `allContent` prop for listing all items
- Template example: `news_page.html`, `committee_page.html`

**Detail Pages:**
- Use content type templates: `news.html`, `committee.html`
- Load specific post from `content/news/post-slug.json`

### Store Best Practices

**Use stores for:**
- Global app state (auth, theme, cart)
- Cross-component communication
- Runtime-only reactive data

**Use props for:**
- Component-specific data
- One-directional parent → child flow
- Build-time-resolved values

---

## 9. Development Workflow (1 minute)

### Running the Server

```bash
go run cmd/server/main.go
# Visit http://localhost:3333
```

### Testing

```bash
# All tests
go test ./... -v

# Specific package
go test ./transformer -v

# Integration tests
go test ./tests/alpine -v
```

### Component Registration

Components in `layouts/components/*.html` are auto-registered on server startup with extracted props and scoped CSS/JS.

### Debugging

Enable expression debugging:
```bash
DEBUG_EXPRESSIONS=true go run cmd/server/main.go
```

### Kill Server

```bash
lsof -ti:3333 | xargs kill -9
```

---

## 10. Recent Fixes & Gotchas (1 minute)

### Console Error Fixes (Jan 2026)

1. **Alpine :class object literals** - Added `isAlpineObjectLiteral()` to preserve `{:class="{'active': bool}"}` syntax
2. **Photos undefined** - Removed empty arrays from JSON (omit optional fields entirely)
3. **Fallback operators** - Removed `||` from loop bodies that fall back to runtime
4. **CMS 404s** - Skip JSON fetch for listing pages that use `allContent` prop

### Parser Gotchas

- Always use `AnyNodeParser` for child node parsing (unified path)
- Depth tracking prevents content after `{/if}` from being consumed
- Template expressions use single braces: `{var}` not `{{var}}`

### Content Gotchas

- Route conflicts: Don't have both `content/page.json` AND `content/pages/page.json`
- Plenti pattern: Arrays use loops (`{for}`), not bracket notation (`[0]`)
- Listing pages should have simple excerpt fields, not complex nested arrays

---

## Quick Reference Card

### Template Syntax Cheat Sheet

```html
<!-- Variables -->
{title}                          → <span x-text="title"></span>
{post?.author?.name}             → Optional chaining supported

<!-- Conditionals -->
{if condition}...{/if}           → Build-time or <template x-if>
{if a}{else if b}{else}{/if}     → Full if/else-if/else chain

<!-- Loops -->
{for item in items}...{/for}     → Build-time expansion or x-for

<!-- Components -->
<Hero title="Welcome" />         → Inlined at build-time
<Component:dynamic name={var} /> → Dynamic by name resolution

<!-- Stores -->
{$auth.user}                     → <span x-text="$store.auth.user"></span>
{if $theme.dark}                 → <template x-if="$store.theme.dark">

<!-- Attributes -->
src="{image.url}"                → Build-time or :src="image.url"
:class="{'active': isActive}"    → Alpine binding preserved
@click="handleClick()"           → Alpine event handler
```

### Content Type Decision Tree

```
Do you need multiple reusable components?
    ↓ YES                          ↓ NO
Pattern 1:                     Pattern 2:
Component-Based               Custom Template
    ↓                              ↓
Use pages.html               Create custom layout
+ components array           + flat JSON structure
+ export let components      + export let fields
```

---

## Resources

- **Full Architecture:** `docs/GO-BUILD-TIME-ENGINE-GUIDES/ARCHITECTURE.md`
- **Build-Time Loops:** `docs/GO-BUILD-TIME-ENGINE-GUIDES/BUILD_TIME_LOOP_EXPANSION.md`
- **Developer Guide:** `docs/GO-BUILD-TIME-ENGINE-GUIDES/DEVELOPER_GUIDE.md`
- **Store System:** `docs/GO-BUILD-TIME-ENGINE-GUIDES/STORE_DEVELOPER_GUIDE.md`
- **Migration Guide:** `docs/GO-BUILD-TIME-ENGINE-GUIDES/MIGRATION_GUIDE.md`

---

## Key Takeaways

1. **Build-time loop expansion** is the headline feature - loops expand on the server when data is available
2. **Hybrid approach** - build-time when possible, runtime fallback when needed
3. **Two content patterns** - component-based (generic) vs custom templates (specific)
4. **Export let system** - Svelte-style prop injection from JSON content
5. **Alpine.js integration** - Reactive runtime layer for dynamic features
6. **x-data optimization** - Only runtime-needed variables included
7. **Parser uses unified path** - BlockConditionalParser + BlockLoopParser with depth tracking

**The Big Idea:** Combine the best of Svelte (compile-time optimization) and Alpine.js (runtime reactivity) while minimizing JavaScript payload and maximizing page performance.
