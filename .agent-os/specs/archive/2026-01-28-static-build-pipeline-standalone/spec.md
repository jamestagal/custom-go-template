# Static Build Pipeline Specification

**Date:** 2026-01-28
**Status:** Ready for Implementation
**Priority:** P0 - Production Blocker
**Reference:** [plenti-build-analysis.md](plenti-build-analysis.md)

---

## Overview

Implement a static site build command (`cmd/build/main.go`) that generates a Plenti-compatible static site. The build pipeline will:

1. Generate a fingerprint for cache-busting
2. Register all components and generate `layouts.js`
3. Scan content and generate `content.js`
4. Generate `env.js` with environment config
5. Render each page with SSR (pre-rendered HTML)
6. Copy and bundle assets into fingerprinted directory
7. Output Plenti-compatible directory structure

---

## Plenti-Compatible Output Structure

Based on analysis of actual Plenti build output:

```
public/
├── index.html                      # Root page (data-content-filepath attribute)
├── about/
│   └── index.html
├── news/
│   └── team-expansion/
│       └── index.html
├── {fingerprint}/                  # e.g., aQwupMmCDl/ (10-char)
│   ├── bundle.css                  # All CSS concatenated
│   ├── core/
│   │   ├── main.js                 # Entry point (hydration)
│   │   └── runtime-components.js   # Alpine.js runtime
│   ├── generated/
│   │   ├── content.js              # allContent array
│   │   ├── layouts.js              # Component registry
│   │   └── env.js                  # Environment config
│   └── bundles/                    # Tree-shaken bundles
│       ├── common.a1b2c3d4.js
│       └── pages/
│           └── _index.e5f6g7h8.js
├── global.css                      # Not fingerprinted (stable path)
├── images/                         # Static images
└── media/                          # Media files
```

---

## HTML Output Format

Each page follows Plenti's SSR pattern:

```html
<!doctype html>
<html data-content-filepath=content/pages/about.json lang=en>
<meta charset=utf-8>
<meta name=viewport content="width=device-width,initial-scale=1">
<title>About</title>
<base href=/>
<script type=module src={fingerprint}/core/main.js></script>
<link rel=stylesheet href=global.css>
<link rel=stylesheet href={fingerprint}/bundle.css>
<!-- Pre-rendered content (SSR) -->
<body x-data="...">
  <main>
    <nav>...</nav>
    <div class="container">
      <!-- Content rendered at build-time -->
    </div>
    <footer>...</footer>
  </main>
</body>
</html>
```

**Key attributes:**
- `data-content-filepath` on `<html>` - Source JSON path for hydration matching
- Single module script to `{fingerprint}/core/main.js`
- CSS: `global.css` (stable) + `{fingerprint}/bundle.css` (versioned)

---

## Generated Files

### content.js

Array format matching Plenti exactly:

```javascript
const allContent = [
  {
    pager: null,
    type: "pages",
    path: "about",
    filepath: "content/pages/about.json",
    filename: "about.json",
    fields: {
      title: "About Us",
      description: ["..."],
      // ... all content fields
    }
  },
  // ... more entries
];
export default allContent;
```

### layouts.js

Named exports with signature format:

```javascript
export { default as layouts_components_hero2436_html } from "../layouts/components/hero2436.js";
export { default as layouts_components_services2437_html } from "../layouts/components/services2437.js";
export { default as layouts_content_pages_html } from "../layouts/content/pages.js";
export { default as layouts_global_html_html } from "../layouts/global/html.js";
```

**Signature format:** `layouts_{category}_{name}_html`

### env.js

```javascript
export let env = {
  local: false,
  baseurl: "/",
  routes: {
    pages: ":filename",
    _index: ":paginate(totalPages)"
  },
  types: ["news", "pages"],
  singleTypes: ["_index"],
  fingerprint: "aQwupMmCDl",
  entrypointHTML: "global/html.html",
  entrypointJS: "aQwupMmCDl"
};
```

---

## Build Config

```go
type BuildConfig struct {
    // Source directories
    ContentDir   string   // "content"
    LayoutsDir   string   // "layouts"
    StaticDirs   []string // ["static", "styles", "core"]

    // Output
    OutputDir    string   // "public"

    // Fingerprinting
    Fingerprint  string   // "auto" (random), "none", or explicit value

    // Features
    TreeShake    bool     // Enable tree-shaking (default: true)
    BundleCSS    bool     // Concatenate CSS into bundle.css (default: true)

    // Build options
    CleanBuild   bool     // Remove OutputDir before building
    Verbose      bool     // Verbose logging
    BaseURL      string   // Base URL for links (default: "/")
}
```

---

## CLI Interface

```bash
# Basic build
go run cmd/build/main.go

# Full options
go run cmd/build/main.go \
    --output=public \
    --content=content \
    --layouts=layouts \
    --tree-shake=true \
    --bundle-css=true \
    --fingerprint=auto \
    --baseurl=/ \
    --clean \
    --verbose

# Production build
go build -o plenti-build cmd/build/main.go
./plenti-build --clean
```

---

## Build Pipeline Steps

### Step 1: Initialize

```go
func Build(config *BuildConfig) error {
    // 1. Validate config
    // 2. Generate fingerprint (if auto)
    // 3. Clean output directory (if --clean)
    // 4. Create directory structure
}
```

### Step 2: Register Components

```go
func registerComponentsForBuild(layoutsDir string) (map[string]*ComponentTemplate, error) {
    // Reuse logic from cmd/server/main.go:registerComponents()
    // Return map of component name -> template
}
```

### Step 3: Generate Generated Files

```go
func generateBuildFiles(config *BuildConfig, components map[string]*ComponentTemplate) error {
    // 1. Generate content.js from content/*.json
    // 2. Generate layouts.js with signature exports
    // 3. Generate env.js with config
    // 4. Generate tree-shaken bundles (if enabled)
}
```

### Step 4: Scan Content

```go
type ContentEntry struct {
    Pager    interface{}            // null or page number
    Type     string                 // "pages", "news", etc.
    Path     string                 // Route path
    Filepath string                 // Source JSON path
    Filename string                 // JSON filename
    Fields   map[string]interface{} // Content data
}

func scanContent(contentDir string) ([]ContentEntry, error)
```

### Step 5: Render Pages

```go
func renderPage(entry ContentEntry, config *BuildConfig) ([]byte, error) {
    // 1. Determine layout from entry.Type
    // 2. Build props (content, allContent, allLayouts, env)
    // 3. Render with wrapper (html.html)
    // 4. Add data-content-filepath attribute
    // 5. Update asset paths with fingerprint
    // 6. Return HTML bytes
}
```

### Step 6: Copy Assets

```go
func copyAssets(config *BuildConfig) error {
    // 1. Concatenate CSS into bundle.css (if enabled)
    // 2. Copy core/*.js to {fingerprint}/core/
    // 3. Copy generated/*.js to {fingerprint}/generated/
    // 4. Copy bundles to {fingerprint}/bundles/
    // 5. Copy global.css to root (not fingerprinted)
    // 6. Copy static files (images, media)
}
```

---

## Integration with Existing Code

### Reuse from Dev Server

| Function | Location | Reuse Strategy |
|----------|----------|----------------|
| `registerComponents()` | `cmd/server/main.go` | Extract to `builder/` |
| `renderWithWrapper()` | `cmd/server/main.go` | Extract, modify for []byte output |
| `getAllContent()` | `cmd/server/main.go` | Use as-is for allContent |
| `generateComponentRegistry()` | `cmd/server/main.go` | Already in `builder/` |
| `generateTreeShakenBundles()` | `cmd/server/main.go` | Already in `builder/` |

### Reuse from Builder Package

| Existing | Purpose |
|----------|---------|
| `builder/registry_generator.go` | Generate layouts.js |
| `builder/tree_shaking.go` | Generate tree-shaken bundles |
| `builder/content_generator.go` | Generate content.js |

---

## New Files to Create

| File | Purpose |
|------|---------|
| `cmd/build/main.go` | CLI entry point |
| `builder/build.go` | Build orchestrator |
| `builder/content_scanner.go` | Scan content directory |
| `builder/page_renderer.go` | Render pages to HTML |
| `builder/asset_pipeline.go` | Copy and bundle assets |
| `builder/env_generator.go` | Generate env.js |
| `builder/fingerprint.go` | Generate fingerprint |

---

## Error Handling

| Error | Behavior |
|-------|----------|
| Missing content directory | Fatal error |
| Missing layouts directory | Fatal error |
| Invalid JSON content | Skip file, log warning |
| Template parse error | Fatal error with location |
| Missing component | Log warning, continue |
| Asset copy failure | Fatal error |
| Output write failure | Fatal error |

---

## Testing Strategy

### Unit Tests

```go
// builder/static_build_test.go
func TestContentScanner(t *testing.T)
func TestFingerprintGeneration(t *testing.T)
func TestPageRenderer(t *testing.T)
func TestCSSBundling(t *testing.T)
```

### Integration Tests

```go
// tests/build/static_build_test.go
func TestFullBuild(t *testing.T) {
    // 1. Build to temp directory
    // 2. Verify all pages exist
    // 3. Verify asset paths correct
    // 4. Verify content.js format
    // 5. Verify HTML has data-content-filepath
}
```

### Manual Verification

```bash
# Build
go run cmd/build/main.go --clean

# Serve with any static server
cd public && python -m http.server 8000

# Test in browser
open http://localhost:8000
```

---

## Success Criteria

1. ✅ `go run cmd/build/main.go` generates complete static site
2. ✅ Output structure matches Plenti conventions exactly
3. ✅ HTML includes `data-content-filepath` attribute
4. ✅ Assets fingerprinted correctly
5. ✅ content.js matches Plenti format
6. ✅ layouts.js uses signature format
7. ✅ Tree-shaken bundles work correctly
8. ✅ Site works when served with any static file server
9. ✅ Build completes in <5 seconds for 20-page site
10. ✅ CI/CD ready (exit codes, logging)

---

## Implementation Order

| Phase | Description | Hours |
|-------|-------------|-------|
| 1 | Core Build Pipeline (CLI, orchestrator) | 3-4 |
| 2 | Content Scanner (discovery, path mapping) | 2-3 |
| 3 | Page Renderer (extract from server, add attributes) | 2-3 |
| 4 | Asset Pipeline (fingerprinting, CSS bundling) | 2-3 |
| 5 | Generated Files (content.js, env.js, layouts.js) | 1-2 |
| 6 | Testing & Verification | 2-3 |
| **Total** | | **12-18** |
