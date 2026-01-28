# Plenti Integration API Specification

**Date:** 2026-01-28
**Status:** Discovery
**Priority:** P0 - Core Integration
**Goal:** Replace Svelte compiler in Plenti with Go template engine

---

## Problem Statement

Plenti currently uses V8 to run the 60k-line Svelte compiler for SSR and DOM compilation. This causes:
- Memory crashes on large applications
- Slow build times
- Excessive client JS for mostly static content
- The `allLayouts` hack for dynamic components

**Solution:** Replace Svelte compilation with a native Go template engine that:
- Compiles directly in Go (no V8)
- Outputs minimal Alpine.js for interactivity
- Eliminates the `allLayouts` component signature hack

---

## Architecture Overview

```
BEFORE (Plenti + Svelte):
┌─────────────────────────────────────────────────────────────┐
│ Plenti Build                                                │
│   Content JSON → V8 Engine → Svelte Compiler → SSR + DOM JS │
│                     ↓                              ↓        │
│              Memory issues              All components load │
└─────────────────────────────────────────────────────────────┘

AFTER (Plenti + Go Templates):
┌─────────────────────────────────────────────────────────────┐
│ Plenti Build                                                │
│   Content JSON → Go Template Engine → SSR HTML + Alpine.js  │
│                     ↓                              ↓        │
│              Fast & stable              Minimal runtime JS  │
└─────────────────────────────────────────────────────────────┘
```

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

// GenerateLayoutsJS generates the component registry (replaces Svelte compilation)
func GenerateLayoutsJS(layoutsDir string) (string, error)

// GenerateContentJS generates content.js from content entries
func GenerateContentJS(entries []ContentEntry) (string, error)
```

### Integration Points

| Plenti Calls | Go Template Provides |
|--------------|---------------------|
| SSR compilation | `RenderPage()` |
| DOM compilation | Not needed (Alpine.js runtime) |
| Component registry | `GenerateLayoutsJS()` |
| Content array | `GenerateContentJS()` |

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
  <script type="module" src="{fingerprint}/core/main.js"></script>
  <link rel="stylesheet" href="global.css">
  <link rel="stylesheet" href="{fingerprint}/bundle.css">
</head>
<body>
  <!-- Pre-rendered content -->
  <main>
    {content}
  </main>
</body>
</html>
```

### Key Attributes

| Attribute | Purpose | Example |
|-----------|---------|---------|
| `data-content-filepath` | Hydration matching | `content/pages/about.json` |

### Generated Files

**layouts.js** (component registry):
```javascript
// Using signature format for backwards compatibility
export { default as layouts_components_hero_html } from "../layouts/components/hero.js";
export { default as layouts_content_pages_html } from "../layouts/content/pages.js";

// Component template functions
export const Hero = (props) => `<div class="hero">...</div>`;
```

**content.js** (Plenti format):
```javascript
const allContent = [
  {
    pager: null,
    type: "pages",
    path: "about",
    filepath: "content/pages/about.json",
    filename: "about.json",
    fields: { title: "About", ... }
  }
];
export default allContent;
```

---

## Implementation Plan

### Phase 1: Core API (3-4 hours)

- [ ] Create `plenti/` package
- [ ] Implement `PageContext` and `ContentEntry` structs
- [ ] Implement `RenderPage()` function
- [ ] Wire up existing renderer to new API
- [ ] Add `data-content-filepath` to HTML output

### Phase 2: Generated Files (2-3 hours)

- [ ] Implement `GenerateLayoutsJS()` with signature format
- [ ] Implement `GenerateContentJS()` matching Plenti format
- [ ] Implement `RenderComponent()` for incremental builds

### Phase 3: Output Compatibility (2-3 hours)

- [ ] Ensure HTML matches Plenti format exactly
- [ ] Test with actual Plenti project structure
- [ ] Document integration steps for Plenti

### Phase 4: Testing (2-3 hours)

- [ ] Integration tests with Plenti content
- [ ] Performance benchmarks vs Svelte
- [ ] Memory usage comparison
- [ ] Test dynamic components without allLayouts hack

---

## Success Criteria

1. **Drop-in replacement**: Plenti can call Go template engine instead of V8/Svelte
2. **Format compatibility**: Output matches Plenti's expected structure
3. **No allLayouts hack**: Dynamic components work without signatures
4. **Performance**: 10x faster builds, no memory crashes
5. **Stability**: No V8 timeout issues on large projects

---

## Estimated Effort

| Phase | Hours |
|-------|-------|
| Core API | 3-4 |
| Generated Files | 2-3 |
| Output Compatibility | 2-3 |
| Testing | 2-3 |
| **Total** | **9-13** |

---

## Dependencies

| Dependency | Status |
|------------|--------|
| Template parser | ✅ Complete |
| AST transformer | ✅ Complete |
| HTML renderer | ✅ Complete |
| Component system | ✅ Complete |
| Store system | ✅ Complete |
| Tree-shaking | ✅ Complete |

---

## Related Specs (Separate Implementation)

| Spec | Purpose | Priority |
|------|---------|----------|
| [Hydration Directives](../hydration-directives/spec.md) | `client:visible`, `client:idle` for islands architecture | P1 |
| [LoadAllContent](../load-all-content/spec.md) | Enhanced content loader with filtering/sorting | P1 |
