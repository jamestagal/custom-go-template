# Static Build Pipeline Discovery

**Date:** 2026-01-28
**Status:** Discovery
**Priority:** P0 - Production Blocker

---

## Problem Statement

The Go template engine currently operates only as a development server (`cmd/server/main.go`). There is no mechanism to generate static HTML files for production deployment. This blocks:

1. **Production deployment** to CDNs (Netlify, Vercel, Cloudflare Pages)
2. **Static hosting** on traditional servers
3. **Pre-rendering** for SEO and performance
4. **CI/CD integration** for automated builds

## Current State Analysis

### What Exists

| Component | Status | Location |
|-----------|--------|----------|
| Template parser | ✅ Complete | `parser/` |
| AST transformer | ✅ Complete | `transformer/` |
| HTML renderer | ✅ Complete | `renderer/` |
| Component registry | ✅ Complete | `builder/registry_generator.go` |
| Tree-shaking | ✅ Complete | `builder/tree_shaking.go` |
| Content loader | ✅ Complete | `loader/loader.go` |
| Dev server | ✅ Complete | `cmd/server/main.go` |
| Static builder | ❌ Missing | `cmd/build/` (does not exist) |

### What the Dev Server Does

From [cmd/server/main.go](cmd/server/main.go):

1. Registers components from `layouts/components/`
2. Generates component registry (`generated/layouts.js`)
3. Generates tree-shaken bundles (`generated/bundles/`)
4. Serves pages via `renderWithWrapper()` for each request
5. Serves static assets from `static/`, `core/`, `generated/`

### What a Build Pipeline Needs

1. **Enumerate all pages** - Scan `content/` for all JSON files
2. **Render each page** - Call `renderWithWrapper()` and save to file
3. **Copy static assets** - `static/`, `styles/`, `core/`
4. **Generate component registry** - `generated/layouts.js`
5. **Generate bundles** - Tree-shaken per-page bundles
6. **Output structure** - Match Plenti's `public/` directory structure
7. **Fingerprinting** - Content-hash filenames for cache busting

---

## Plenti Build Output Structure

Reference from existing Plenti projects:

```
public/
├── index.html                    # Homepage
├── about/
│   └── index.html                # /about page
├── portfolio/
│   ├── index.html                # /portfolio listing
│   ├── project-1/
│   │   └── index.html            # /portfolio/project-1
│   └── project-2/
│       └── index.html            # /portfolio/project-2
├── {fingerprint}/                # Versioned assets (e.g., 1706123456789/)
│   ├── css/
│   │   └── style.css
│   ├── js/
│   │   ├── main.js
│   │   └── layouts.js            # Component registry
│   ├── bundles/                  # Tree-shaken bundles
│   │   ├── common.a1b2c3d4.js
│   │   └── pages/
│   │       ├── _index.e5f6g7h8.js
│   │       └── about.i9j0k1l2.js
│   └── images/
│       └── ...
├── core/                         # Runtime JS (not fingerprinted)
│   └── runtime-components.js
└── images/                       # Static images (may be fingerprinted)
```

### Key Patterns

1. **Pretty URLs** - `/about/index.html` serves as `/about`
2. **Fingerprinted assets** - Long cache TTL for versioned files
3. **Non-fingerprinted core** - Runtime JS stays at fixed paths
4. **Nested content** - Collections generate nested directories

---

## Implementation Approach

### Phase 1: Core Build Pipeline

```go
// cmd/build/main.go

type BuildConfig struct {
    ContentDir   string // "content"
    LayoutsDir   string // "layouts"
    StaticDir    string // "static"
    OutputDir    string // "public"
    Fingerprint  string // Unix timestamp or content hash
    TreeShake    bool   // Enable tree-shaking
}

func Build(config *BuildConfig) error {
    // 1. Initialize renderer with all components
    // 2. Generate component registry
    // 3. Generate tree-shaken bundles
    // 4. Enumerate all content files
    // 5. Render each page to output
    // 6. Copy static assets (with fingerprinting)
    // 7. Generate asset manifest
}
```

### Phase 2: Content Enumeration

```go
// builder/content_scanner.go

type ContentFile struct {
    Path        string // "content/pages/about.json"
    Type        string // "pages" (determines template)
    Slug        string // "about"
    OutputPath  string // "public/about/index.html"
}

func ScanContent(contentDir string) ([]ContentFile, error)
```

### Phase 3: Page Rendering

```go
// builder/page_renderer.go

func RenderPage(file ContentFile, components map[string]*ComponentTemplate) ([]byte, error) {
    // 1. Load JSON content
    // 2. Determine template (layouts/content/{type}.html)
    // 3. Call renderWithWrapper()
    // 4. Return HTML bytes
}
```

### Phase 4: Asset Pipeline

```go
// builder/asset_pipeline.go

type AssetManifest struct {
    Fingerprint string
    CSS         map[string]string // original -> fingerprinted path
    JS          map[string]string
    Images      map[string]string
}

func CopyAssets(staticDir, outputDir, fingerprint string) (*AssetManifest, error)
```

---

## CLI Interface

```bash
# Basic build
go run cmd/build/main.go

# With options
go run cmd/build/main.go \
    --output=public \
    --content=content \
    --layouts=layouts \
    --tree-shake=true \
    --fingerprint=auto  # "auto" = unix timestamp, or provide explicit value

# Clean build
go run cmd/build/main.go --clean
```

---

## Integration Points

### With Tree-Shaking (Already Exists)

```go
// Tree-shaking already generates:
// - generated/bundles/common.{hash}.js
// - generated/bundles/pages/{page}.{hash}.js
// - generated/bundle-manifest.json

// Build pipeline will:
// 1. Call GenerateTreeShakenBundles()
// 2. Copy bundles to public/{fingerprint}/bundles/
// 3. Update HTML script tags with correct paths
```

### With Component Registry (Already Exists)

```go
// Registry generation already exists:
// - builder/registry_generator.go
// - Generates generated/layouts.js

// Build pipeline will:
// 1. Call GenerateComponentRegistry()
// 2. Copy to public/{fingerprint}/js/layouts.js
```

### With Renderer (Already Exists)

```go
// renderer/render.go provides:
// - renderWithWrapper() for Plenti-pattern pages
// - renderTemplate() for standalone pages

// Build pipeline will call these with file-based output
```

---

## Output Verification

### Expected Build Stats

```
=== Build Complete ===
Pages rendered: 12
  - content/index.json → public/index.html
  - content/pages/about.json → public/about/index.html
  - content/pages/contact.json → public/contact/index.html
  - content/portfolio/*.json → public/portfolio/*/index.html (5 pages)
  - ...

Assets copied:
  - CSS: 3 files (style.css, fonts.css, code.css)
  - JS: 4 files (main.js, layouts.js, runtime-components.js, Alpine.js)
  - Images: 24 files

Bundles generated:
  - common.a1b2c3d4.js (15.5 KB)
  - pages/_index.e5f6g7h8.js (8.7 KB)
  - ... (12 page bundles)

Total build time: 1.2s
Output directory: public/
```

### Verification Checklist

- [ ] All content files have corresponding HTML output
- [ ] Pretty URLs work (`/about` → `about/index.html`)
- [ ] Static assets copied with correct paths
- [ ] Fingerprinted assets have correct hashes
- [ ] Script tags reference correct bundle paths
- [ ] Alpine.js initializes correctly
- [ ] Component interactivity works
- [ ] No broken internal links

---

## Risk Assessment

| Risk | Mitigation |
|------|------------|
| Path resolution differences (dev vs build) | Use relative paths, test both modes |
| Missing content files | Validate content before build |
| Large builds timeout | Add progress reporting, consider parallelization |
| Asset path mismatches | Generate manifest, verify in HTML |
| Tree-shaking edge cases | Fallback to full registry |

---

## Success Criteria

1. **`go run cmd/build/main.go` produces deployable static site**
2. **Output matches Plenti directory structure**
3. **All pages render correctly when served statically**
4. **Tree-shaken bundles work correctly**
5. **Build completes in <5 seconds for 20-page site**
6. **CI/CD integration works (exit codes, logging)**

---

## Dependencies

- ✅ Template parser (complete)
- ✅ AST transformer (complete)
- ✅ HTML renderer (complete)
- ✅ Component registry (complete)
- ✅ Tree-shaking (complete)
- ✅ Content loader (complete)

**No blockers** - all dependencies are satisfied.

---

## Estimated Effort

| Phase | Hours |
|-------|-------|
| Core build pipeline (`cmd/build/main.go`) | 3-4 |
| Content enumeration | 2-3 |
| Page rendering integration | 2-3 |
| Asset pipeline with fingerprinting | 2-3 |
| Testing and verification | 2-3 |
| **Total** | **11-16** |

---

## Next Steps

1. Create `spec.md` with detailed implementation plan
2. Create `tasks.md` with task breakdown
3. Implement Phase 1: Core build pipeline
4. Implement Phase 2-4 incrementally
5. Test with real Plenti project structure
