# Tree-Shaking Discovery Appraisal

**Date:** 2026-01-28  
**Reviewer:** Claude (Technical Architecture Review)  
**Document Reviewed:** DISCOVERY.md  
**Status:** ✅ Approved for Phase 1 Implementation

---

## Executive Summary

The discovery document successfully identifies a **high-impact optimization opportunity** with solid data backing. The 84% average waste figure is compelling and the methodology is sound. The competitive analysis against Plenti's architecture is particularly strong, clearly articulating why our build-time resolution approach enables optimizations that Plenti cannot achieve.

**Recommendation:** Proceed with Phase 1 implementation. The opportunity is too significant to ignore, and all prerequisites are complete.

---

## Overall Assessment

| Aspect | Rating | Notes |
|--------|--------|-------|
| **Data Quality** | ★★★★★ | Concrete, verifiable metrics |
| **Problem Identification** | ★★★★★ | Clear articulation of waste |
| **Competitive Analysis** | ★★★★★ | Excellent Plenti comparison |
| **Solution Strategy** | ★★★★☆ | Good phases, needs manifest detail |
| **Integration with Existing Work** | ★★★☆☆ | Should connect to v4.0 blueprint |
| **Actionability** | ★★★★★ | Clear next steps |

---

## Strengths

### 1. Data-Driven Analysis

The document provides concrete metrics rather than estimates:

- **181 KB** total layouts.js → **12.8 KB** average actual usage
- Per-page breakdowns show exactly what's wasted
- Component-by-component size inventory enables prioritization

This level of specificity makes the business case undeniable.

### 2. Excellent Competitive Analysis

The Plenti architecture comparison is the document's strongest section. It clearly articulates the core architectural advantage:

```
Plenti (Runtime):     import * as allLayouts → Runtime lookup → All loaded
Our Engine (Build):   Content scan → Static analysis → Only used bundled
```

The fact that we expand loops at build time (not Alpine x-for) means we can **know exactly which components each page needs** before generating JavaScript. This is a fundamental capability difference, not just an optimization.

### 3. Practical Phased Approach

The three-phase strategy is well-structured with appropriate effort/impact ratios:

| Phase | Effort | Impact | Risk |
|-------|--------|--------|------|
| 1. Static tree-shaking | Medium | **84% reduction** | Low |
| 2. Category separation | Low | Cleaner architecture | Minimal |
| 3. Dynamic imports | High | Edge cases only | Medium |

**Smart prioritization**: Phase 1 delivers most value with reasonable effort. Phases 2-3 are incremental improvements, not blockers.

### 4. Orphan Component Detection

The ~36 KB of orphan components (45% of registry) is a critical finding:

| Component | Size | Likely Status |
|-----------|------|---------------|
| `header` | 7.9 KB | Used by global template? |
| `headerlogo` | 7.1 KB | Orphan |
| `userdashboard` | 6.2 KB | Test component |
| `adminpanel` | 4.1 KB | Test component |
| `userprofile` | 3.9 KB | Test component |
| `footer-old` | 2.9 KB | Dead code |

This suggests **immediate wins without any new infrastructure** - just remove confirmed dead code.

---

## Gaps & Considerations

### 1. Global Template Analysis Missing

The document notes orphan components might be used by global templates but doesn't confirm:

```
⚠️ "header (7.9 KB) - Orphan or global?"
```

**Impact:** Could incorrectly classify actively-used components as orphans.

**Recommendation:** Before claiming orphan status, scan `layouts/global/*` for component references. The header is almost certainly used by the global navigation or layout template.

**Suggested Addition to Discovery:**
```
Global Template Scan Required:
- layouts/global/html.html → Check for component references
- layouts/global/nav.html → Check for header/headerlogo usage
- layouts/global/footer.html → Check for footer-old reference
```

### 2. Dynamic Component Edge Cases Understated

The document assumes most components are statically resolvable, but the v4.0 blueprint describes runtime resolution scenarios:

```html
{for component in components}
  <Component:dynamic name={component.name} {...component.fields} />
{/for}
```

**The critical distinction:**
- If `component.name` comes from **content JSON** (known at build) → Tree-shakeable ✅
- If `component.name` comes from **Alpine store** (runtime) → Needs fallback registry ⚠️

**Impact:** Without categorization, we might tree-shake components that are actually needed at runtime.

**Recommendation:** Add a categorization pass during build:

```go
type ComponentReference struct {
    Name       string
    Source     string  // "literal", "content", "runtime"
    ResolvesAt string  // "build" or "runtime"
}

// Build-time resolvable (tree-shakeable)
"Header"                    → Source: literal,  ResolvesAt: build
content.hero.component      → Source: content,  ResolvesAt: build
components[0].name          → Source: content,  ResolvesAt: build

// Runtime-only (needs fallback)
$store.selectedComponent    → Source: runtime,  ResolvesAt: runtime
component.name (in x-for)   → Source: runtime,  ResolvesAt: runtime
```

### 3. Bundle Naming & Caching Strategy Not Specified

The document shows output structure:
```
bundles/
  pages/_index.js       # 26 KB
  pages/about.js        # 1.5 KB
```

But doesn't address several important concerns:

| Concern | Question | Impact |
|---------|----------|--------|
| **Cache busting** | How do we invalidate cached bundles on content changes? | Stale bundles served |
| **Shared chunks** | What if `hero2436` is used on 5 pages? | Duplicate code across bundles |
| **Loading strategy** | How does the page know which bundle to load? | Runtime confusion |

**Recommendation:** Specify manifest generation strategy:

```javascript
// generated/bundle-manifest.json
{
  "version": "a1b2c3d4",  // Build hash for cache busting
  "bundles": {
    "pages/_index": {
      "path": "/bundles/pages/_index.a1b2c3.js",
      "components": ["hero2436", "services2437", "whyChoose2425"],
      "size": 26100
    },
    "pages/about": {
      "path": "/bundles/pages/about.d4e5f6.js",
      "components": ["hero", "team"],
      "size": 1500
    }
  },
  "shared": {
    "hero2436": ["pages/_index", "pages/test-dynamic"]
  }
}
```

**Loading Strategy Options:**

| Strategy | Pros | Cons |
|----------|------|------|
| **Inline manifest** | No extra request | Increases HTML size |
| **Script data attribute** | Clean separation | Requires parsing |
| **Global variable** | Simple access | Global namespace pollution |

Recommended: Inline manifest in a small `<script>` tag on each page.

### 4. Integration with v4.0 Blueprint Not Addressed

The SSG + Hydration v4.0 blueprint already specifies a registry manifest:

```javascript
// From v4.0 Blueprint
window.$registryManifest = {
  "version": "abc123",
  "main": "/js/registry.main.abcd.js",
  "components": {
    "Hero2436": "/js/registry.hero2436.ef01.js"
  }
};
```

This discovery document should explicitly connect to that architecture.

**Alignment Opportunities:**

| v4.0 Blueprint | Tree-Shaking Discovery | Integration |
|----------------|------------------------|-------------|
| Per-page registry chunks | Per-page bundles | Same output structure |
| Registry manifest | Bundle manifest | Merge into single manifest |
| Lazy loading via manifest | Dynamic import fallback | Same loading mechanism |
| Signature-based lookup | Signature-based tree-shaking | Leverage existing signatures |

**Recommendation:** The tree-shaking implementation should output bundles that conform to the v4.0 registry manifest specification. This ensures consistency and avoids parallel systems.

### 5. Incremental Build Impact Not Analyzed

The document focuses on full builds but doesn't address incremental scenarios:

| Scenario | Question |
|----------|----------|
| Single content file changes | Do we regenerate all bundles or just affected ones? |
| Component template changes | Which bundles need regeneration? |
| New component added | How does this affect existing bundles? |

**Impact:** Without incremental build support, development iteration could become slow as the site grows.

**Recommendation:** Track dependency graph:

```
Component Dependencies:
  hero2436 → [pages/_index, pages/test-dynamic]
  services2437 → [pages/_index, pages/test-dynamic]
  
Content Dependencies:
  pages/_index.json → bundles/pages/_index.js
  pages/about.json → bundles/pages/about.js
```

When a component changes, only regenerate bundles that include it.

---

## Validation of Key Claims

### Claim: "84% average potential savings"

**Validated** ✅

The math checks out:
- 79.5 KB total components
- 12.8 KB average used per page
- (79.5 - 12.8) / 79.5 = **83.9%**

### Claim: "Plenti can't tree-shake because of runtime lookup"

**Validated** ✅

The pattern identified is indeed un-analyzable by bundlers:

```javascript
import * as allLayouts from "../generated/layouts.js";
allLayouts["layouts_components_" + name + "_svelte"]  // Runtime string concatenation
```

No bundler (esbuild, Rollup, Webpack) can statically determine which exports are used when the export name is computed at runtime.

### Claim: "Our build-time resolution enables tree-shaking"

**Validated with Caveat** ✅/⚠️

True for build-resolved components (the majority), but we still need a fallback strategy for runtime-resolved cases. The document acknowledges this in Phase 3 but should be more explicit about the boundary.

---

## Recommendations

### Immediate Actions (No New Infrastructure)

| Action | Effort | Impact | Notes |
|--------|--------|--------|-------|
| Remove `footer-old` | 5 min | 2.9 KB saved | Confirmed dead code |
| Review test components | 30 min | ~18 KB saved | userdashboard, adminpanel, userprofile |
| Scan global templates | 1 hour | Validates orphan list | Required before removing header/headerlogo |

### Phase 1 Implementation Checklist

```markdown
## Phase 1: Build-Time Static Tree-Shaking

### Component Usage Analyzer
- [ ] Scan content JSON for component references
- [ ] Scan templates for static component usage
- [ ] Categorize references as build-time vs runtime
- [ ] Output: `ComponentUsageMap` per page

### Per-Page Bundle Generation
- [ ] Generate bundle containing only used components
- [ ] Include component dependencies (if any)
- [ ] Apply content-hash to filename for cache busting

### Bundle Manifest Generation
- [ ] Generate manifest mapping pages → bundle paths
- [ ] Include component list per bundle (for debugging)
- [ ] Include size metrics (for monitoring)
- [ ] Conform to v4.0 registry manifest specification

### Fallback Registry
- [ ] Keep full registry available for dynamic cases
- [ ] Load fallback only when component not in page bundle
- [ ] Log fallback usage for optimization feedback
```

### Architecture Integration

| Task | Priority | Notes |
|------|----------|-------|
| Align manifest format with v4.0 blueprint | High | Single source of truth |
| Connect to signature system | Medium | Leverage existing infrastructure |
| Add to build pipeline | High | Integrate with existing Go build |
| Document loading strategy | Medium | Developer documentation |

---

## Success Metrics

Upon Phase 1 completion, measure:

| Metric | Current | Target | Method |
|--------|---------|--------|--------|
| Average JS per page | 181 KB | < 30 KB | Build output analysis |
| Bundle generation time | N/A | < 500ms | Build timing |
| Fallback registry usage | N/A | < 5% of loads | Runtime logging |
| Cache hit rate | Unknown | > 90% | CDN analytics |

---

## Conclusion

The discovery document provides a solid foundation for implementing tree-shaking. The 84% reduction opportunity is real and achievable with the proposed Phase 1 approach.

**Key Strengths:**
- Excellent data quality and competitive analysis
- Practical phased implementation strategy
- Clear prerequisites already completed

**Address Before Implementation:**
- Scan global templates to validate orphan classification
- Specify manifest format aligned with v4.0 blueprint
- Document caching and shared chunk strategies
- Add categorization for build-time vs runtime component references

**Final Verdict:** ✅ **Approved for Phase 1 Implementation**

The prerequisites are complete, the opportunity is validated, and the risks are manageable. Proceed with implementation while addressing the gaps identified in this appraisal.

---

## Appendix: Related Documents

| Document | Location | Relevance |
|----------|----------|-----------|
| SSG + Hydration v4.0 Blueprint | Project docs | Registry manifest specification |
| Component Signatures Spec | .agent-os/specs | Signature format for tree-shaking |
| Focused Solution: Runtime Resolution | Project docs | Build vs runtime categorization |

---

*Appraisal completed: 2026-01-28*  
*Next review: After Phase 1 implementation*
