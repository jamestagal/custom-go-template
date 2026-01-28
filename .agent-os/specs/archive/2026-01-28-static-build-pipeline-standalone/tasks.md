# Static Build Pipeline Tasks

**Spec:** spec.md
**Analysis:** plenti-build-analysis.md
**Status:** Not Started
**Estimated:** 12-18 hours

---

## Phase 1: Core Build Pipeline

### 1.1 CLI Entry Point
- [ ] Create `cmd/build/main.go`
- [ ] Define `BuildConfig` struct matching spec
- [ ] Parse CLI flags: `--output`, `--content`, `--layouts`, `--clean`, `--verbose`
- [ ] Parse feature flags: `--tree-shake`, `--bundle-css`, `--fingerprint`, `--baseurl`
- [ ] Implement `--help` with usage examples
- [ ] Validate config before build starts
- [ ] Call `Build(config)` orchestrator

### 1.2 Build Orchestrator
- [ ] Create `builder/build.go`
- [ ] Implement `Build(config *BuildConfig) error` main function
- [ ] Step logging with timing (1. Initialize, 2. Components, 3. Content, etc.)
- [ ] Implement `--clean` to remove output directory
- [ ] Create fingerprint directory structure
- [ ] Create build summary output

### 1.3 Fingerprint Generation
- [ ] Create `builder/fingerprint.go`
- [ ] Implement `GenerateFingerprint()` → 10-char random string (Plenti-compatible)
- [ ] Support `--fingerprint=auto|none|<value>` modes
- [ ] Return fingerprint for use in paths

---

## Phase 2: Content Scanner

### 2.1 Content Discovery
- [ ] Create `builder/content_scanner.go`
- [ ] Define `ContentEntry` struct (pager, type, path, filepath, filename, fields)
- [ ] Implement `ScanContent(contentDir string) ([]ContentEntry, error)`
- [ ] Walk `content/` recursively
- [ ] Parse each JSON file into ContentEntry
- [ ] Skip `_defaults.json`, `_schema.json`

### 2.2 Content Entry Fields
- [ ] Set `pager` field (null for regular pages, number for paginated)
- [ ] Set `type` from directory (pages, news, blog, _index)
- [ ] Set `path` as route path (e.g., "about", "news/article")
- [ ] Set `filepath` as source path (e.g., "content/pages/about.json")
- [ ] Set `filename` as just the filename
- [ ] Set `fields` from JSON content

### 2.3 Path Mapping
- [ ] Implement `ContentEntryToOutputPath(entry, outputDir) string`
- [ ] Handle `_index.json` → `/index.html`
- [ ] Handle `pages/about.json` → `/about/index.html`
- [ ] Handle `news/article.json` → `/news/article/index.html`

---

## Phase 3: Generated Files

### 3.1 content.js Generation
- [ ] Create `builder/content_generator.go` (or update existing)
- [ ] Generate Plenti-compatible format:
  ```javascript
  const allContent = [ { pager, type, path, filepath, filename, fields } ];
  export default allContent;
  ```
- [ ] Write to `{outputDir}/{fingerprint}/generated/content.js`

### 3.2 layouts.js Generation
- [ ] Update `builder/registry_generator.go` for signature format
- [ ] Generate named exports: `export { default as layouts_components_hero2436_html } from "..."`
- [ ] Use signature format: `layouts_{category}_{name}_html`
- [ ] Write to `{outputDir}/{fingerprint}/generated/layouts.js`

### 3.3 env.js Generation
- [ ] Create `builder/env_generator.go`
- [ ] Generate environment config:
  ```javascript
  export let env = { local, baseurl, routes, types, singleTypes, fingerprint, entrypointHTML, entrypointJS };
  ```
- [ ] Detect content types from content directory
- [ ] Write to `{outputDir}/{fingerprint}/generated/env.js`

---

## Phase 4: Page Renderer

### 4.1 Extract Rendering Logic
- [ ] Create `builder/page_renderer.go`
- [ ] Extract core logic from `cmd/server/main.go:renderWithWrapper()`
- [ ] Convert from `http.ResponseWriter` to `[]byte` output
- [ ] Remove HTTP-specific code (headers, error responses)
- [ ] Accept ContentEntry instead of http.Request

### 4.2 Render Function
- [ ] Implement `RenderPage(entry ContentEntry, config *BuildConfig) ([]byte, error)`
- [ ] Determine layout from `entry.Type` (pages → pages.html, news → news.html)
- [ ] Build props map: `content`, `allContent`, `allLayouts`, `env`
- [ ] Transform and render template
- [ ] Inject styles, scripts, Alpine.js CDN
- [ ] Add `data-content-filepath` attribute to `<html>` tag

### 4.3 HTML Post-Processing
- [ ] Update script src to `{fingerprint}/core/main.js`
- [ ] Update CSS links to `{fingerprint}/bundle.css`
- [ ] Add tree-shaking bundle data attributes (if enabled)
- [ ] Ensure proper DOCTYPE and base href

### 4.4 Batch Rendering
- [ ] Implement `RenderAllPages(entries []ContentEntry, config *BuildConfig) error`
- [ ] Render pages sequentially (avoid race conditions)
- [ ] Create output directories as needed
- [ ] Write HTML to correct output path
- [ ] Track success/failure counts
- [ ] Continue on render error (with warning)

---

## Phase 5: Asset Pipeline

### 5.1 CSS Bundling
- [ ] Create `builder/asset_pipeline.go`
- [ ] Implement `BundleCSS(stylesDirs []string, outputPath string) error`
- [ ] Concatenate all `*.css` files
- [ ] Write to `{outputDir}/{fingerprint}/bundle.css`
- [ ] Copy `global.css` to root (not fingerprinted)

### 5.2 JavaScript Assets
- [ ] Copy `core/*.js` to `{outputDir}/{fingerprint}/core/`
- [ ] Copy `generated/*.js` to `{outputDir}/{fingerprint}/generated/`
- [ ] Copy tree-shaken bundles to `{outputDir}/{fingerprint}/bundles/`

### 5.3 Static Assets
- [ ] Copy `static/images/*` to `{outputDir}/images/`
- [ ] Copy `static/media/*` to `{outputDir}/media/`
- [ ] Copy other static files (favicon, robots.txt)

### 5.4 Asset Manifest (Optional)
- [ ] Generate manifest mapping original → output paths
- [ ] Write to `{outputDir}/asset-manifest.json`

---

## Phase 6: Integration & Testing

### 6.1 Component Registration
- [ ] Extract `registerComponents()` to shared location
- [ ] Make it work for both server and build
- [ ] Include store registration

### 6.2 Unit Tests
- [ ] Create `builder/static_build_test.go`
- [ ] Test `ScanContent()` path mapping
- [ ] Test `GenerateFingerprint()` format
- [ ] Test `ContentEntryToOutputPath()`
- [ ] Test CSS bundling

### 6.3 Integration Tests
- [ ] Create `tests/build/static_build_test.go`
- [ ] Test full build to temp directory
- [ ] Verify all pages created
- [ ] Verify `data-content-filepath` attribute present
- [ ] Verify content.js format
- [ ] Verify layouts.js signature format
- [ ] Verify asset paths correct

### 6.4 Manual Testing
- [ ] Build static site
- [ ] Serve with `python -m http.server`
- [ ] Test all pages load
- [ ] Test Alpine.js components work
- [ ] Test navigation between pages
- [ ] Test on different browsers

---

## Phase 7: Polish

### 7.1 Build Summary
- [ ] Print summary at end of build
- [ ] Show pages rendered count
- [ ] Show assets copied count
- [ ] Show total size
- [ ] Show build time

### 7.2 Error Handling
- [ ] Clear error messages with file locations
- [ ] Exit codes: 0=success, 1=error
- [ ] Suggestions for common issues

### 7.3 Documentation
- [ ] Add build command to CLAUDE.md
- [ ] Document CLI options
- [ ] Add deployment examples (Netlify, Vercel, Cloudflare Pages)

---

## Definition of Done

- [ ] `go run cmd/build/main.go` works from project root
- [ ] Output structure matches Plenti exactly
- [ ] HTML has `data-content-filepath` attribute
- [ ] content.js in Plenti format (pager, type, path, filepath, filename, fields)
- [ ] layouts.js uses signature format (layouts_{category}_{name}_html)
- [ ] env.js generated with correct config
- [ ] CSS bundled into `{fingerprint}/bundle.css`
- [ ] Static site works with any HTTP server
- [ ] All tests pass
- [ ] Build time <5s for current content

---

## Dependencies

- ✅ Template parser (complete)
- ✅ AST transformer (complete)
- ✅ HTML renderer (complete)
- ✅ Component registry (complete)
- ✅ Tree-shaking (complete)
- ✅ Content loader (complete)

---

## Reference: Plenti Build Output

From analysis of `/Users/benjaminwaller/Projects/My Plenti Sites WIP/Plenti`:

```
public/
├── index.html                 # data-content-filepath=content/_index.json
├── about/index.html           # data-content-filepath=content/pages/about.json
├── blog/components/index.html # data-content-filepath=content/blog/components.json
├── aQwupMmCDl/               # Fingerprint directory
│   ├── bundle.css
│   ├── core/
│   │   ├── main.js
│   │   ├── router.js
│   │   └── cms/
│   ├── generated/
│   │   ├── content.js
│   │   ├── layouts.js
│   │   └── env.js
│   └── layouts/               # Compiled Svelte → JS
│       ├── components/
│       ├── content/
│       └── global/
├── global.css                 # Not fingerprinted
└── media/
```

HTML pattern:
```html
<!doctype html>
<html data-content-filepath=content/pages/about.json lang=en>
<script type=module src=aQwupMmCDl/core/main.js></script>
<link rel=stylesheet href=global.css>
<link rel=stylesheet href=aQwupMmCDl/bundle.css>
<!-- Pre-rendered SSR content -->
</html>
```
