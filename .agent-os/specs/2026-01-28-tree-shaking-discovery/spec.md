# Tree-Shaking Implementation Specification

**Date:** 2026-01-28
**Status:** Draft
**Prerequisites:** Component Signatures (Completed 2026-01-28)
**Related:** DISCOVERY.md, APPRAISAL.md

---

## Executive Summary

Implement build-time tree-shaking to reduce per-page JavaScript payload by **84%** on average. The system analyzes component usage at build time and generates per-page bundles containing only the components each page actually uses.

**Key Metrics:**
- Current: 181 KB loaded on every page
- Target: 12.8 KB average per page (84% reduction)
- Shared chunk threshold: Components used on 2+ pages

---

## Architecture Overview

```
Build Pipeline:
┌─────────────────┐    ┌─────────────────┐
│ Content JSON    │───▶│ Component Usage │
│ pages/*.json    │    │ Analyzer        │
└─────────────────┘    └────────┬────────┘
                                │
┌─────────────────┐             │
│ Template Files  │─────────────┤
│ layouts/*/*.html│             │
└─────────────────┘             ▼
                       ┌────────────────────┐
                       │ Bundle Generator   │
                       └────────┬───────────┘
                                │
              ┌─────────────────┼─────────────────┐
              ▼                 ▼                 ▼
        ┌──────────┐     ┌──────────┐     ┌──────────┐
        │ common   │     │ _index   │     │ about    │
        │ .e7f8.js │     │ .a1b2.js │     │ .d4e5.js │
        └──────────┘     └──────────┘     └──────────┘
```

---

## Design Decisions

### 1. Cache Busting Strategy

**Decision:** Content-hash filenames within Plenti fingerprint directory

**Format:**
```
public/{fingerprint}/
  bundles/
    pages/_index.a1b2c3d4.js    ← 8-char content hash
    pages/about.e5f6g7h8.js
    common.i9j0k1l2.js
  generated/
    layouts.js                   ← Fallback (full registry)
    bundle-manifest.json         ← Maps pages to bundles
```

**Rationale:**
- Plenti compatibility (directory fingerprint regenerates each build)
- Granular caching (unchanged bundles keep same hash)
- CDN-friendly (immutable file names)

**Hash Generation:**
```go
// builder/bundle_hash.go
func GenerateBundleHash(content []byte) string {
    hash := sha256.Sum256(content)
    return hex.EncodeToString(hash[:])[:8]  // 8 chars = 4 bytes = ~4 billion combinations
}
```

### 2. Shared Chunks Strategy

**Decision:** Threshold-based extraction for components used on 2+ pages

**Algorithm:**
```go
type ComponentUsage struct {
    Name     string
    Pages    []string  // Which pages use this component
    Size     int       // Component template size in bytes
}

func ShouldExtractToCommon(usage ComponentUsage) bool {
    return len(usage.Pages) >= 2
}
```

**Output Structure:**
```javascript
// common.i9j0k1l2.js
const common = {
  'hero2436': (props) => `...`,      // Used on: _index, test-dynamic
  'services2437': (props) => `...`,  // Used on: _index, test-dynamic
};
export default common;

// pages/_index.a1b2c3d4.js
import common from '../common.i9j0k1l2.js';
const page = {
  'whyChoose2425': (props) => `...`,  // Only used on this page
};
export default { ...common, ...page };
```

**Benefits:**
- `hero2436` (5.3 KB) loaded once, cached, reused across pages
- Page-specific bundles only contain unique components
- Reduced total download across multi-page sessions

### 3. Loading Strategy

**Decision:** Data attribute on script tag

**HTML Output:**
```html
<script type="module"
        src="/{fingerprint}/core/main.js"
        data-bundle="/{fingerprint}/bundles/pages/about.d4e5f6g7.js"
        data-common="/{fingerprint}/bundles/common.i9j0k1l2.js"
        data-fallback="/{fingerprint}/generated/layouts.js">
</script>
```

**Runtime Loading (core/main.js):**
```javascript
const script = document.currentScript;
const commonPath = script.dataset.common;
const bundlePath = script.dataset.bundle;
const fallbackPath = script.dataset.fallback;

let registry = {};

try {
  // Load common chunk first (shared components)
  if (commonPath) {
    const common = await import(commonPath);
    registry = { ...registry, ...common.default };
  }

  // Load page-specific bundle
  const bundle = await import(bundlePath);
  registry = { ...registry, ...bundle.default };

} catch (e) {
  console.warn('[Tree-Shaking] Bundle load failed, using fallback registry');
  const fallback = await import(fallbackPath);
  registry = fallback.default;
}

// Make registry available for dynamic components
window.$componentRegistry = registry;
```

**Benefits:**
- No extra HTTP requests for manifest
- Parallel loading of common + page bundle possible
- Graceful fallback to full registry
- SSG and SSR compatible

---

## Bundle Manifest Format

**File:** `generated/bundle-manifest.json`

```json
{
  "version": "a1b2c3d4e5f6",
  "generated": "2026-01-28T10:30:00Z",
  "fingerprint": "aQwupMmCDl",
  "common": {
    "path": "/aQwupMmCDl/bundles/common.i9j0k1l2.js",
    "components": ["hero2436", "services2437"],
    "size": 15555
  },
  "bundles": {
    "pages/_index": {
      "path": "/aQwupMmCDl/bundles/pages/_index.a1b2c3d4.js",
      "components": ["whyChoose2425"],
      "includes_common": true,
      "total_size": 26100,
      "unique_size": 10545
    },
    "pages/about": {
      "path": "/aQwupMmCDl/bundles/pages/about.e5f6g7h8.js",
      "components": ["hero", "team"],
      "includes_common": false,
      "total_size": 1550,
      "unique_size": 1550
    }
  },
  "fallback": {
    "path": "/aQwupMmCDl/generated/layouts.js",
    "size": 181000
  },
  "stats": {
    "total_components": 25,
    "common_components": 2,
    "orphan_components": 8,
    "average_bundle_size": 12800,
    "average_savings_percent": 84
  }
}
```

**Use Cases:**
- Build tooling: Generate bundles based on manifest
- Debugging: Verify correct bundle for each page
- Monitoring: Track bundle sizes over time
- CI: Fail if bundle size exceeds threshold

---

## Component Usage Analyzer

### Input Sources

1. **Content JSON** - Component references in `components` array
2. **Template Files** - Static `<ComponentName />` references
3. **Global Templates** - Components used in `layouts/global/*`

### Analysis Algorithm

```go
// analyzer/component_usage.go

type PageUsage struct {
    Page       string            // e.g., "pages/_index"
    Components []string          // Components used on this page
    Source     map[string]string // Component → how it was discovered
}

type UsageAnalyzer struct {
    contentDir  string
    layoutsDir  string
}

func (a *UsageAnalyzer) Analyze() (map[string]PageUsage, error) {
    usage := make(map[string]PageUsage)

    // 1. Scan content JSON files
    contentFiles, _ := filepath.Glob(a.contentDir + "/**/*.json")
    for _, file := range contentFiles {
        page := a.extractPagePath(file)
        components := a.extractComponentsFromJSON(file)
        usage[page] = PageUsage{
            Page:       page,
            Components: components,
            Source:     map[string]string{},
        }
        for _, c := range components {
            usage[page].Source[c] = "content"
        }
    }

    // 2. Scan template files for static component usage
    for page, pu := range usage {
        templatePath := a.layoutsDir + "/content/" + pu.Page + ".html"
        staticComponents := a.extractStaticComponents(templatePath)
        for _, c := range staticComponents {
            if _, exists := pu.Source[c]; !exists {
                pu.Components = append(pu.Components, c)
                pu.Source[c] = "template"
            }
        }
    }

    // 3. Add global components to every page
    globalComponents := a.extractGlobalComponents()
    for page := range usage {
        for _, c := range globalComponents {
            if _, exists := usage[page].Source[c]; !exists {
                usage[page].Components = append(usage[page].Components, c)
                usage[page].Source[c] = "global"
            }
        }
    }

    return usage, nil
}
```

### Categorization: Build-Time vs Runtime

```go
type ComponentReference struct {
    Name       string
    ResolvesAt string  // "build" or "runtime"
    Source     string  // "literal", "content", "store", "loop"
}

// Build-time resolvable (tree-shakeable)
"Header"                    → ResolvesAt: build, Source: literal
content.hero.component      → ResolvesAt: build, Source: content
components[0].name          → ResolvesAt: build, Source: content

// Runtime-only (needs fallback registry)
$store.selectedComponent    → ResolvesAt: runtime, Source: store
component.name (in x-for)   → ResolvesAt: runtime, Source: loop
```

**Rule:** If ANY page has runtime component resolution, the fallback registry MUST be available.

---

## Bundle Generator

### Phase 1: Generate Common Chunk

```go
// builder/common_chunk.go

func GenerateCommonChunk(usage map[string]PageUsage, registry map[string]ComponentTemplate) ([]byte, []string) {
    componentPageCount := make(map[string]int)

    // Count pages per component
    for _, pu := range usage {
        for _, comp := range pu.Components {
            componentPageCount[comp]++
        }
    }

    // Extract components used on 2+ pages
    var commonComponents []string
    for comp, count := range componentPageCount {
        if count >= 2 {
            commonComponents = append(commonComponents, comp)
        }
    }

    // Generate common chunk
    return generateChunkJS(commonComponents, registry), commonComponents
}
```

### Phase 2: Generate Per-Page Bundles

```go
// builder/page_bundle.go

func GeneratePageBundle(page string, usage PageUsage, commonComponents []string, registry map[string]ComponentTemplate) []byte {
    // Filter out common components
    var uniqueComponents []string
    for _, comp := range usage.Components {
        if !contains(commonComponents, comp) {
            uniqueComponents = append(uniqueComponents, comp)
        }
    }

    // Generate bundle that imports common + adds unique
    return generatePageBundleJS(uniqueComponents, commonComponents, registry)
}
```

### Output Format

**common.js:**
```javascript
// Auto-generated common component chunk
// Components used on 2+ pages

const common = {
  'hero2436': (props) => `
    <div class="hero" x-data='${JSON.stringify(props)}'>
      <h1 x-text="title">${props.title || ''}</h1>
    </div>
  `,
  'services2437': (props) => `...`,
};

export default common;
```

**pages/_index.js:**
```javascript
// Auto-generated bundle for pages/_index
// Unique components: whyChoose2425

import common from '../common.i9j0k1l2.js';

const unique = {
  'whyChoose2425': (props) => `...`,
};

export default { ...common, ...unique };
```

---

## Integration with Go Renderer

### HTML Generation

```go
// renderer/script_tags.go

func (r *Renderer) GenerateScriptTag(pagePath string, manifest BundleManifest) string {
    bundle := manifest.Bundles[pagePath]

    attrs := []string{
        `type="module"`,
        fmt.Sprintf(`src="/%s/core/main.js"`, manifest.Fingerprint),
        fmt.Sprintf(`data-bundle="%s"`, bundle.Path),
    }

    if manifest.Common.Path != "" && bundle.IncludesCommon {
        attrs = append(attrs, fmt.Sprintf(`data-common="%s"`, manifest.Common.Path))
    }

    attrs = append(attrs, fmt.Sprintf(`data-fallback="%s"`, manifest.Fallback.Path))

    return fmt.Sprintf("<script %s></script>", strings.Join(attrs, " "))
}
```

**Example Output:**
```html
<script type="module"
        src="/aQwupMmCDl/core/main.js"
        data-bundle="/aQwupMmCDl/bundles/pages/_index.a1b2c3d4.js"
        data-common="/aQwupMmCDl/bundles/common.i9j0k1l2.js"
        data-fallback="/aQwupMmCDl/generated/layouts.js">
</script>
```

---

## Incremental Build Support

### Dependency Graph

```go
// builder/dependency_graph.go

type DependencyGraph struct {
    ComponentToPages map[string][]string  // component → pages using it
    PageToComponents map[string][]string  // page → components it uses
    ComponentHashes  map[string]string    // component → content hash
    BundleHashes     map[string]string    // bundle → content hash
}

func (g *DependencyGraph) InvalidatedBundles(changedComponents []string) []string {
    invalidated := make(map[string]bool)

    for _, comp := range changedComponents {
        for _, page := range g.ComponentToPages[comp] {
            invalidated[page] = true
        }
    }

    // Also invalidate common chunk if any shared component changed
    for _, comp := range changedComponents {
        if len(g.ComponentToPages[comp]) >= 2 {
            invalidated["common"] = true
            break
        }
    }

    var result []string
    for bundle := range invalidated {
        result = append(result, bundle)
    }
    return result
}
```

### Incremental Build Flow

1. Compare current component hashes with previous build
2. Identify changed components
3. Lookup affected bundles via dependency graph
4. Regenerate only affected bundles
5. Update manifest with new hashes

---

## Fallback Strategy

### When Fallback Is Needed

1. **Runtime component resolution** - Component name from Alpine store or x-for loop
2. **Bundle load failure** - Network error, 404, parse error
3. **Missing component** - Component added after bundle generation

### Fallback Loading

```javascript
// core/main.js - Fallback handling

async function loadComponent(name) {
    // Try page bundle first
    if (window.$componentRegistry && window.$componentRegistry[name]) {
        return window.$componentRegistry[name];
    }

    // Load fallback if not already loaded
    if (!window.$fallbackRegistry) {
        console.log('[Tree-Shaking] Loading fallback registry for:', name);
        const script = document.currentScript || document.querySelector('script[data-fallback]');
        const fallbackPath = script.dataset.fallback;
        const fallback = await import(fallbackPath);
        window.$fallbackRegistry = fallback.default;
    }

    return window.$fallbackRegistry[name];
}
```

### Fallback Usage Logging

```javascript
// Track fallback usage for optimization feedback
if (window.$fallbackRegistry && !window.$componentRegistry[name]) {
    console.warn(`[Tree-Shaking] Component '${name}' loaded from fallback. Consider adding to bundle.`);

    // Optional: Send to analytics
    if (window.$treeshakingAnalytics) {
        window.$treeshakingAnalytics.recordFallback(name, location.pathname);
    }
}
```

---

## File Structure

```
builder/
  tree_shaking.go           # Main tree-shaking orchestration
  component_usage.go        # Usage analyzer
  bundle_generator.go       # Bundle generation
  common_chunk.go           # Common chunk extraction
  page_bundle.go            # Per-page bundle generation
  bundle_hash.go            # Content hash generation
  dependency_graph.go       # Incremental build support
  bundle_manifest.go        # Manifest generation

core/
  main.js                   # Updated with bundle loading

generated/
  layouts.js                # Full registry (fallback)
  bundle-manifest.json      # Bundle metadata
  bundles/
    common.{hash}.js        # Shared components
    pages/
      _index.{hash}.js      # Per-page bundles
      about.{hash}.js
      ...
```

---

## Success Criteria

| Metric | Current | Target | Validation |
|--------|---------|--------|------------|
| Average JS per page | 181 KB | < 30 KB | Build output analysis |
| Bundle generation time | N/A | < 2s full build | CI timing |
| Incremental build time | N/A | < 200ms | CI timing |
| Fallback usage | N/A | < 5% of loads | Runtime logging |
| Common chunk hit rate | N/A | > 60% | CDN analytics |

---

## Phases

### Phase 1: Core Tree-Shaking (This Spec)
- Component usage analyzer
- Per-page bundle generation
- Common chunk extraction (2+ pages threshold)
- Bundle manifest generation
- Data attribute loading strategy
- Fallback registry support

### Phase 2: Optimization (Future)
- Incremental builds via dependency graph
- Bundle size budgets and CI checks
- Fallback usage analytics
- Lazy loading for large components

### Phase 3: Advanced (Future)
- Dynamic import for runtime components
- Route-based prefetching
- Service worker caching

---

## Appendix: Expected Output Sizes

Based on discovery analysis:

| Bundle | Components | Size | vs Full Registry |
|--------|------------|------|------------------|
| common | hero2436, services2437 | 15.5 KB | N/A |
| pages/_index | whyChoose2425 + common | 26.0 KB | **86% smaller** |
| pages/about | hero, team | 1.5 KB | **99% smaller** |
| pages/jim-test | 7 test components | 8.7 KB | **95% smaller** |

**Total saved across 4 pages:** 614.8 KB (vs 724 KB loading full registry 4 times)
