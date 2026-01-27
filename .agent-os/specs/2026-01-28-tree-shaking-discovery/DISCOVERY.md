# Tree-Shaking Discovery Document

**Date:** 2026-01-28
**Status:** Discovery Complete
**Prerequisite:** Component Signatures (Completed 2026-01-28)

---

## Executive Summary

Analysis of the current template engine reveals **significant tree-shaking opportunities**. The average page only uses **16% of available components**, meaning **84% of the component registry is unused per page**. Implementing tree-shaking could reduce per-page JavaScript payload by **54-78 KB**.

---

## Current State Analysis

### Bundle Sizes

| File | Size | Description |
|------|------|-------------|
| `generated/layouts.js` | **181 KB** | Full component registry |
| `generated/content.js` | 19 KB | Content data |
| Component templates only | 79.5 KB | Without aliases/overhead |

### Component Inventory

| Category | Count | Loading Strategy (Current) |
|----------|-------|---------------------------|
| `components/*` | 25 | All loaded on every page |
| `content/*` | 13 | All loaded on every page |
| `global/*` | 5 | All loaded on every page |

### Top Components by Size

| Component | Size | Usage |
|-----------|------|-------|
| `whyChoose2425` | 10.5 KB | 1 page (home) |
| `services2437` | 10.2 KB | 2 pages |
| `header` | 7.9 KB | 0 pages (orphan?) |
| `headerlogo` | 7.1 KB | 0 pages (orphan?) |
| `userdashboard` | 6.2 KB | 0 pages (orphan?) |
| `hero2436` | 5.3 KB | 2 pages |
| `adminpanel` | 4.1 KB | 0 pages (orphan?) |
| `userprofile` | 3.9 KB | 0 pages (orphan?) |

---

## Per-Page Analysis

### Homepage (`/_index`)

```
Components: hero2436, services2437, whyChoose2425
Used:       26.1 KB (3 components)
Unused:     54.0 KB (22 components)
Savings:    67.9%
```

### About Page (`/about`)

```
Components: hero, team
Used:       1.5 KB (2 components)
Unused:     78.0 KB (23 components)
Savings:    98.1%
```

### Jim Test Page (`/jim-test`)

```
Components: 7 test components
Used:       8.7 KB (7 components)
Unused:     70.8 KB (18 components)
Savings:    89.0%
```

### Test Dynamic (`/test-dynamic`)

```
Components: hero2436, services2437
Used:       15.2 KB (2 components)
Unused:     64.3 KB (23 components)
Savings:    80.9%
```

---

## Aggregate Metrics

| Metric | Value |
|--------|-------|
| Average components per page | 3.5 |
| Average bytes used per page | 12.8 KB |
| Average bytes wasted per page | 66.7 KB |
| **Average potential savings** | **84.0%** |

---

## Tree-Shaking Strategy Analysis

### Strategy 1: Build-Time Static Analysis

**How it works:**
1. During build, scan each page's content JSON for component references
2. Generate per-page bundles containing only used components
3. Use signature categories to determine loading requirements

**Pros:**
- Zero runtime overhead
- Deterministic bundles
- Works with SSG (Static Site Generation)

**Cons:**
- Requires rebuild when content changes
- Doesn't handle truly dynamic component names

**Implementation complexity:** Medium

### Strategy 2: Route-Based Code Splitting

**How it works:**
1. Group components by which content types use them
2. Generate route-based bundles (e.g., `pages.bundle.js`, `news.bundle.js`)
3. Load bundle based on current route

**Pros:**
- Fewer bundles to manage
- Natural grouping by content type
- Works with existing routing

**Cons:**
- Less granular than per-page
- May still include unused components within a type

**Implementation complexity:** Low

### Strategy 3: Dynamic Import with Component Registry

**How it works:**
1. Keep minimal core bundle (global + content templates)
2. Components lazy-loaded on first reference
3. Use `import()` with component signatures

**Pros:**
- Maximum flexibility
- Only loads what's actually rendered
- Works with runtime dynamic components

**Cons:**
- Waterfall loading possible
- More network requests
- Cache management needed

**Implementation complexity:** High

---

## Recommended Approach

### Phase 1: Build-Time Static Tree-Shaking (Recommended First)

Given our semantic signatures and build-time loop expansion, **Strategy 1** is the natural fit:

```
Build Pipeline:
┌─────────────────┐
│ Content JSON    │───┐
└─────────────────┘   │
                      ▼
┌─────────────────┐  ┌──────────────────────┐
│ Template Files  │──│ Component Usage      │
└─────────────────┘  │ Analyzer             │
                     └──────────────────────┘
                              │
              ┌───────────────┼───────────────┐
              ▼               ▼               ▼
        ┌──────────┐   ┌──────────┐   ┌──────────┐
        │ home.js  │   │ about.js │   │ news.js  │
        │ 26 KB    │   │ 1.5 KB   │   │ 0 KB     │
        └──────────┘   └──────────┘   └──────────┘
```

**Output structure:**
```
generated/
  layouts.js              # Full registry (fallback)
  bundles/
    pages/_index.js       # 26 KB (hero, services, whyChoose)
    pages/about.js        # 1.5 KB (hero, team)
    pages/jim-test.js     # 8.7 KB (test components)
    news/*.js             # Per-article bundles
```

### Phase 2: Global/Content Template Optimization

Separate concerns:
```
core/
  global.js     # nav, header, footer (always loaded)
  content/
    pages.js    # pages.html template
    news.js     # news.html template
```

### Phase 3: Dynamic Import for Runtime Components

For components resolved at runtime (from loops with unknown names):
```javascript
// Only when component name isn't known at build time
const component = await import(`/bundles/components/${name}.js`);
```

---

## Implementation Effort Estimate

| Phase | Effort | Impact |
|-------|--------|--------|
| Phase 1: Static tree-shaking | Medium | **High (84% reduction)** |
| Phase 2: Category separation | Low | Medium (cleaner architecture) |
| Phase 3: Dynamic imports | High | Low (edge cases only) |

---

## Prerequisites (All Completed ✓)

- [x] Semantic signatures (`layouts_components_Hero2436_html`)
- [x] Category extraction from signatures
- [x] Build-time loop expansion
- [x] Component registry generation

---

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Missing component at runtime | Fallback to full registry |
| Build complexity increase | Incremental adoption per content type |
| Cache invalidation | Use content-hash in bundle names |
| Dynamic component names | Runtime import fallback |

---

## Orphan Component Analysis

Components in registry but not referenced in any content:

| Component | Size | Status |
|-----------|------|--------|
| `header` | 7.9 KB | Orphan or global? |
| `headerlogo` | 7.1 KB | Orphan |
| `userdashboard` | 6.2 KB | Orphan (test?) |
| `adminpanel` | 4.1 KB | Orphan (test?) |
| `userprofile` | 3.9 KB | Orphan (test?) |
| `footer-old` | 2.9 KB | Deprecated? |
| `productcard` | 2.5 KB | Orphan |
| `featured_posts_sidebar` | 2.0 KB | Orphan |

**Total orphan size: ~36 KB (45% of component registry)**

These could be:
1. Used by global templates (need to analyze)
2. Test components (should be excluded from prod)
3. Dead code (should be removed)

---

## Current Plenti (Svelte) Architecture Analysis

Analyzed from: `/Users/benjaminwaller/Projects/My Plenti Sites WIP/Plenti/`

### Plenti's Current Approach

**File Structure:**
```
public/{fingerprint}/
  generated/
    layouts.js         # 1.5 KB - RE-EXPORTS ONLY
    content.js         # Content data
  layouts/
    components/
      ball.js          # 748 bytes (individual file)
      pager.js         # 8.3 KB (individual file)
      ...
    content/
      pages.js         # Content template
      blog.js
    global/
      nav.js
      footer.js
```

**Key Insight: layouts.js is NOT a bundle!**

```javascript
// Plenti's generated/layouts.js (1.5 KB)
export {default as layouts_components_ball_svelte} from "../layouts/components/ball.js";
export {default as layouts_components_pager_svelte} from "../layouts/components/pager.js";
// ... just re-exports, no actual code
```

**Plenti's main.js loading pattern:**
```javascript
// Loads ALL layouts (no tree-shaking)
import * as allLayouts from "../generated/layouts.js";

// BUT dynamically imports content template (partial tree-shaking)
import("../layouts/content/" + content.type + ".js")
```

### What Plenti Does Well

| Feature | Status | Impact |
|---------|--------|--------|
| Separate component files | ✅ Yes | Enables potential tree-shaking |
| Signature-based exports | ✅ Yes | Consistent lookup pattern |
| Dynamic content imports | ✅ Yes | Route-based splitting for templates |

### What Plenti Doesn't Do (Our Opportunity)

| Feature | Status | Potential Impact |
|---------|--------|------------------|
| Component tree-shaking | ❌ No | **84% bundle reduction** |
| Per-page bundles | ❌ No | Eliminates unused components |
| Lazy component loading | ❌ No | Faster initial load |

### Why Plenti Can't Tree-Shake Components

```javascript
// This pattern defeats tree-shaking:
import * as allLayouts from "../generated/layouts.js";  // Imports EVERYTHING

// Runtime lookup requires all exports present:
allLayouts["layouts_components_" + name + "_svelte"]
```

The `import *` pattern plus runtime string concatenation means bundlers (esbuild, Rollup) cannot statically analyze which exports are used.

### Our Advantage: Build-Time Resolution

Because our Go template engine resolves components at **build time** (not runtime), we can:

1. **Know exactly which components each page uses** before generating bundles
2. **Generate per-page bundles** with only required components
3. **Keep runtime lookup as fallback** for truly dynamic cases

```
Plenti (Runtime):     Page → Runtime lookup → All components loaded
Our Engine (Build):   Page → Build analysis → Only used components bundled
```

### Plenti Component Sizes (Reference)

```
8,330 bytes  pager.js
4,948 bytes  source.js
2,254 bytes  grid.js
  757 bytes  incrementer.js
  757 bytes  decrementer.js
  748 bytes  ball.js
  564 bytes  block.js
─────────────────────
18,358 bytes  TOTAL (17.9 KB) - 7 components
```

Plenti's example site is small (7 components, 18 KB). Our test site has 25 components (80 KB), making tree-shaking more valuable.

---

## Next Steps

1. **Decision:** Proceed with tree-shaking spec? (Y/N)
2. **Scope:** Which phases to include?
3. **Priority:** Where does this fit in roadmap?

---

## Appendix: Raw Data

### Full Component Size List

```
 10,545 bytes  whyChoose2425
 10,249 bytes  services2437
  7,922 bytes  header
  7,061 bytes  headerlogo
  6,246 bytes  userdashboard
  5,343 bytes  hero2436
  4,061 bytes  adminpanel
  3,915 bytes  userprofile
  2,865 bytes  footer-old
  2,538 bytes  productcard
  2,421 bytes  jim_test_animals_loop
  2,008 bytes  featured_posts_sidebar
  1,524 bytes  notification
  1,486 bytes  age
  1,466 bytes  ThemeToggle
  1,386 bytes  CartBadge
  1,350 bytes  jim_test_todos
  1,280 bytes  jim_test_notifications
  1,200 bytes  jim_test_advanced_loops
  1,180 bytes  LoginStatus
  1,150 bytes  todos
    950 bytes  jim_test_greeting
    900 bytes  jim_test_user_profiles
    850 bytes  jim_test_age_examples
    800 bytes  hero
    750 bytes  team
─────────────────────────────
 79,623 bytes  TOTAL (77.8 KB)
```

### Content Type Distribution

```
pages/     - 10 files (component-based pages)
news/      - 3 files (flat JSON articles)
```
