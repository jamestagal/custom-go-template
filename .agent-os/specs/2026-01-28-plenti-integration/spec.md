# Plenti Integration API Specification

**Date:** 2026-01-28
**Updated:** 2026-01-28 (with findings from Plenti source exploration)
**Status:** In Progress
**Priority:** P0 - Core Integration
**Goal:** Replace Svelte compiler in Plenti with Go template engine

---

## Problem Statement

Plenti currently uses V8 to run the 60k-line Svelte compiler for SSR and DOM compilation. This causes:
- Memory crashes on large applications
- Slow build times
- Excessive client JS for mostly static content

**Solution:** Replace Svelte compilation with a native Go template engine that:
- Compiles directly in Go (no V8)
- Outputs minimal Alpine.js for interactivity
- Expands loops at build-time (more efficient than Svelte)

---

## What We've Already Built (Status)

Based on current implementation in `custom_go_template`:

| Feature | Status | Notes |
|---------|--------|-------|
| Template parser | ✅ Complete | Fence section, expressions, conditionals, loops |
| AST transformer | ✅ Complete | Alpine.js output (x-data, x-for, x-if, x-text) |
| Component system | ✅ Complete | Registration, props, dynamic resolution |
| `export let` syntax | ✅ Complete | Svelte-compatible prop declaration |
| Content injection | ✅ Complete | JSON content → component props |
| Build-time loop expansion | ✅ Complete | More efficient than Svelte runtime loops |
| `allContent` magic variable | ✅ Complete | Passed to templates that declare it |
| `allLayouts` registry | ✅ Complete | Component signature → template mapping |
| Conditional script injection | ✅ Complete | Only load runtime JS when needed |
| CMS integration | ✅ Complete | JSON content discovery via `data-content-filepath` |

---

## Architecture Overview

```
BEFORE (Plenti + Svelte):
┌─────────────────────────────────────────────────────────────┐
│ Plenti Build                                                │
│   Content JSON → V8 Engine → Svelte Compiler → SSR + DOM JS │
│                     ↓                              ↓        │
│              Memory issues              All components load │
│              60k line compiler          Full Svelte runtime │
└─────────────────────────────────────────────────────────────┘

AFTER (Plenti + Go Templates):
┌─────────────────────────────────────────────────────────────┐
│ Plenti Build                                                │
│   Content JSON → Go Template Engine → SSR HTML + Alpine.js  │
│                     ↓                              ↓        │
│              Fast & stable              Minimal runtime JS  │
│              Native Go                  Only when needed    │
└─────────────────────────────────────────────────────────────┘
```

---

## Plenti Compatibility Analysis

### Magic Variables (Plenti Pattern)

Plenti injects these magic variables to templates:

| Variable | Description | Our Implementation |
|----------|-------------|-------------------|
| `content` | Current page data `{type, path, fields}` | ✅ Passed as props |
| `allContent` | Array of all content entries | ✅ Opt-in via `export let allContent` |
| `allLayouts` | Component signature → template map | ✅ Generated registry |
| `env` | Environment config | ⏳ To implement |
| `user` | CMS auth state | ⏳ To implement |

### Template Syntax Compatibility

**Plenti/Svelte:**
```svelte
<script>
  export let title, author, allContent;

  let blogPosts = allContent.filter(c => c.type === "blog");
</script>

<h1>{title}</h1>
{#each blogPosts as post}
  <a href="{post.path}">{post.fields.title}</a>
{/each}
```

**Our Go Template Engine:**
```html
---
export let title, author, allContent
---

<h1>{title}</h1>
{for post in allContent}
  {if post.type === "blog"}
    <a href="{post.path}">{post.fields.title}</a>
  {/if}
{/for}
```

**Key Differences:**
| Svelte | Go Template | Notes |
|--------|-------------|-------|
| `{#each}` | `{for}` | Different syntax, same result |
| `{#if}` | `{if}` | Slightly different syntax |
| `.filter()` | Build-time conditionals | We expand at build-time (more efficient!) |
| `<script>` | `---` fence | Same purpose, different syntax |
| `export let` | `export let` | ✅ Identical |

### Content Structure (Plenti Format)

```javascript
// allContent array entry
{
  pager: null,           // Pagination number
  type: "blog",          // Content type (folder name)
  path: "/blog/post1",   // URL path
  filepath: "content/blog/post1.json",
  filename: "post1.json",
  fields: {              // User-defined JSON content
    title: "My Post",
    author: "Jane",
    // ... any fields
  }
}
```

**Our implementation matches this exactly.**

### Component Signatures (allLayouts)

Plenti uses component signatures for dynamic loading:

```
layouts/components/hero.svelte → layouts_components_hero_svelte
layouts/content/blog.svelte   → layouts_content_blog_svelte
```

**We use the same pattern with `.html` extension:**
```
layouts/components/hero.html → layouts_components_hero_html
layouts/content/blog.html   → layouts_content_blog_html
```

---

## Key Architectural Decision: Build-Time Expansion

**Plenti/Svelte approach:**
1. Pass entire `allContent` array to client
2. Run `.filter()` and `{#each}` at runtime
3. Results in larger page weight

**Our approach:**
1. Expand loops at build-time
2. Filter content during transformation
3. Only include what's needed in HTML

**Example:**
```html
<!-- Template -->
{for post in allContent}
  {if post.type === "blog"}
    <div>{post.fields.title}</div>
  {/if}
{/for}

<!-- Plenti output: runtime loop -->
<template x-for="post in allContent">
  <template x-if="post.type === 'blog'">
    <div x-text="post.fields.title"></div>
  </template>
</template>

<!-- Our output: build-time expanded -->
<div>First Blog Post</div>
<div>Second Blog Post</div>
<div>Third Blog Post</div>
```

**This is an IMPROVEMENT over Plenti, not a difference to fix.**

---

## Core Integration API

### Interface for Plenti

```go
package plenti

// PageContext contains all data Plenti passes to the renderer
type PageContext struct {
    // Content entry for this page
    Content ContentEntry

    // All content entries (for allContent prop)
    AllContent []ContentEntry

    // Path to layouts directory
    LayoutsDir string

    // Fingerprint for asset paths
    Fingerprint string

    // Base URL
    BaseURL string

    // Environment (dev/prod)
    IsDev bool

    // CMS configuration
    CMS *CMSConfig
}

// ContentEntry matches Plenti's content structure
type ContentEntry struct {
    Pager    interface{}            `json:"pager"`
    Type     string                 `json:"type"`
    Path     string                 `json:"path"`
    Filepath string                 `json:"filepath"`
    Filename string                 `json:"filename"`
    Fields   map[string]interface{} `json:"fields"`
}

// RenderResult contains the rendered output
type RenderResult struct {
    HTML     string            // Full HTML document
    CSS      string            // Extracted component styles
    JS       string            // Interactive JS (minimal)
    Metadata map[string]string // Title, description, etc.
}

// RenderPage renders a single page with the given context
func RenderPage(ctx PageContext) (*RenderResult, error)

// RenderComponent renders a single component (for incremental builds)
func RenderComponent(name string, props map[string]interface{}, layoutsDir string) (string, error)

// GenerateLayoutsJS generates the component registry
func GenerateLayoutsJS(layoutsDir string) (string, error)
```

### Integration Points

| Plenti Calls | Go Template Provides |
|--------------|---------------------|
| SSR compilation | `RenderPage()` |
| DOM compilation | Not needed (Alpine.js handles reactivity) |
| Component registry | `GenerateLayoutsJS()` |
| Content discovery | Keep Plenti's existing system |

---

## Output Format Compatibility

### HTML Structure

Must match Plenti's expected format:

```html
<!doctype html>
<html data-content-filepath="content/pages/about.json" lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width,initial-scale=1">
  <title>{title}</title>
  <base href="/">
  <!-- Alpine.js for reactivity -->
  <script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
  <link rel="stylesheet" href="/{fingerprint}/bundle.css">
</head>
<body x-data="{...}">
  <!-- Pre-rendered content -->
  <main>
    {content}
  </main>
</body>
</html>
```

### Key Attributes

| Attribute | Purpose | Status |
|-----------|---------|--------|
| `data-content-filepath` | CMS content discovery | ✅ Implemented |
| `x-data` | Alpine.js reactive state | ✅ Implemented |
| `<base href="/">` | Asset path resolution | ✅ Implemented |

---

## Implementation Plan

### Phase 1: Package Structure (2-3 hours)
- [ ] Create `plenti/` package with public API
- [ ] Implement `PageContext` and `ContentEntry` structs
- [ ] Wrapper function `RenderPage()` calling existing renderer
- [ ] Export component registry generation

### Phase 2: Output Compatibility (2-3 hours)
- [ ] Ensure HTML matches Plenti format exactly
- [ ] Add `env` magic variable support
- [ ] Add `user` magic variable for CMS auth
- [ ] Test fingerprinted asset paths

### Phase 3: Plenti Build Integration (3-4 hours)
- [ ] Create integration layer for Plenti's `cmd/build/`
- [ ] Replace `compile.go` calls with Go template rendering
- [ ] Test with actual Plenti project structure
- [ ] Document integration steps

### Phase 4: Testing & Performance (2-3 hours)
- [ ] Integration tests with Plenti content
- [ ] Performance benchmarks vs Svelte compilation
- [ ] Memory usage comparison
- [ ] Verify no V8 timeout issues

---

## Success Criteria

1. ✅ **Svelte-compatible syntax**: `export let` props work identically
2. ✅ **Build-time expansion**: Loops expand during build (improvement!)
3. ✅ **Component signatures**: `allLayouts` pattern preserved
4. ✅ **CMS integration**: `data-content-filepath` for content discovery
5. ⏳ **Drop-in replacement**: Plenti can call Go template instead of V8/Svelte
6. ⏳ **Format compatibility**: Output matches Plenti's expected structure
7. ⏳ **Performance**: 10x faster builds, no memory crashes

---

## What We DON'T Need

Based on Plenti source analysis, these features are NOT needed:

| Feature | Reason |
|---------|--------|
| Custom `loadContent()` API | Plenti uses `.filter()` on `allContent` |
| Client-side hydration directives | Static pages don't need hydration |
| Complex routing | Plenti handles routing |
| Content generation | Plenti handles content discovery |

---

## Estimated Effort

| Phase | Hours |
|-------|-------|
| Package Structure | 2-3 |
| Output Compatibility | 2-3 |
| Plenti Build Integration | 3-4 |
| Testing | 2-3 |
| **Total** | **9-13** |

---

## Documentation References

- [Plenti allContent docs](https://plenti.co/docs/allcontent)
- [Plenti analysis](../../../docs/plenti/plenti-analysis.md)
- [Earlier integration spec](../../../docs/plenti/plenti-integration-spec.md)
