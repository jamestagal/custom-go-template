# Tree-Shaking Implementation Tasks

**Spec:** spec.md
**Status:** ✅ Complete
**Completed:** 2026-01-28

---

## Phase 1: Core Tree-Shaking

### 1.1 Component Usage Analyzer
- [x] Create `builder/component_usage.go`
- [x] Implement `ExtractComponentsFromJSON()` - scan content/*.json for component references
- [x] Implement `ExtractStaticComponents()` - scan template files for `<ComponentName />` patterns
- [x] Implement `ExtractGlobalComponents()` - scan layouts/global/* for always-loaded components
- [x] Create `ComponentUsageMap` type mapping pages → components
- [x] Add categorization: build-time vs runtime resolution
- [x] Write tests for each extraction method

### 1.2 Common Chunk Generation
- [x] Create `builder/common_chunk.go`
- [x] Implement `CalculateComponentPageCount()` - count pages per component
- [x] Implement `ExtractCommonComponents()` - components used on 2+ pages
- [x] Implement `GenerateCommonChunkJS()` - generate common.js bundle
- [x] Add content hash generation for cache busting
- [x] Write tests for threshold logic

### 1.3 Per-Page Bundle Generation
- [x] Create `builder/page_bundle.go`
- [x] Implement `GeneratePageBundle()` - generate bundle with unique components
- [x] Implement `GeneratePageBundleJS()` - output format with common import
- [x] Add content hash to bundle filenames
- [x] Handle pages with no unique components (common-only)
- [x] Write tests for bundle generation

### 1.4 Bundle Manifest Generation
- [x] Create `builder/bundle_manifest.go`
- [x] Define `BundleManifest` struct matching spec format
- [x] Implement `GenerateManifest()` - create bundle-manifest.json
- [x] Include version, paths, component lists, sizes
- [x] Add stats calculation (average savings, orphan count)
- [x] Write to `generated/bundle-manifest.json`
- [x] Write tests for manifest format

### 1.5 Bundle Hash Generation
- [x] Create `builder/bundle_hash.go`
- [x] Implement `GenerateBundleHash()` - SHA256 truncated to 8 chars
- [x] Ensure deterministic output (sorted component order)
- [x] Write tests for hash consistency

### 1.6 Tree-Shaking Orchestration
- [x] Create `builder/tree_shaking.go`
- [x] Implement `GenerateTreeShakenBundles()` - main entry point
- [x] Wire together: usage → common → page bundles → manifest
- [x] Ensure fallback registry (layouts.js) still generated
- [x] Add CLI flag: `--tree-shake` (default: true)
- [x] Write integration tests

---

## Phase 2: Runtime Loading

### 2.1 Update core/main.js
- [x] Add bundle loading from data attributes
- [x] Implement common chunk loading (if present)
- [x] Implement page bundle loading
- [x] Add fallback registry loading on error
- [x] Expose `window.$componentRegistry`
- [x] Add logging for fallback usage
- [x] Test with Alpine.js initialization

### 2.2 Update HTML Renderer
- [x] Modify `renderer/plenti_html.go` (script tag generation)
- [x] Generate script tag with data-bundle attribute
- [x] Generate data-common attribute when page uses common chunk
- [x] Generate data-fallback attribute
- [x] Integrate with Plenti fingerprint directory
- [x] Write tests for HTML output

---

## Phase 3: Server Integration

### 3.1 Development Server
- [x] Update `cmd/server/main.go` to call tree-shaking
- [x] Serve bundles from generated/bundles/
- [x] Regenerate on file change (watch mode)
- [x] Add bundle size logging to startup

### 3.2 Static Build
- [x] Integrate with static site generation
- [x] Ensure bundles copied to public/{fingerprint}/
- [x] Verify manifest paths are correct for production

---

## Phase 4: Testing & Validation

### 4.1 Unit Tests
- [x] `builder/tree_shaking_test.go` (comprehensive tests for all modules)

### 4.2 Integration Tests
- [x] Test: Bundle only contains expected components
- [x] Test: Common chunk contains shared components
- [x] Test: Fallback registry still works
- [x] Test: Dynamic component resolution from bundle
- [x] Test: Cache busting hashes change on content change

### 4.3 Manual Testing
- [x] Verify homepage loads with tree-shaken bundle
- [x] Verify about page loads with smaller bundle
- [x] Verify dynamic components still work
- [x] Verify fallback triggers for runtime components
- [x] Check browser DevTools for correct bundle sizes

---

## Definition of Done

- [x] All tests pass (except pre-existing integration test issue)
- [x] Bundle sizes match expected from discovery (~84% reduction)
- [x] Fallback works for runtime component resolution
- [x] No regressions in existing functionality
- [x] Documentation updated in CLAUDE.md
- [x] Spec marked complete

---

## Notes

### Dependencies
- Requires: Component Signatures (Completed)
- Requires: Build-time loop expansion (Completed)
- Requires: Component registry generation (Completed)

### Risk Mitigation
- Keep full registry as fallback (layouts.js)
- Log fallback usage for monitoring
- Graceful degradation on bundle load failure

### Future Improvements (Not in Scope)
- Incremental builds via dependency graph
- Lazy loading for large components
- Service worker caching
- Bundle size budgets in CI
