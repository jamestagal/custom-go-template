# Tree-Shaking Implementation Tasks

**Spec:** spec.md
**Status:** Not Started
**Estimated Effort:** Medium (2-3 days)

---

## Phase 1: Core Tree-Shaking

### 1.1 Component Usage Analyzer
- [ ] Create `builder/component_usage.go`
- [ ] Implement `ExtractComponentsFromJSON()` - scan content/*.json for component references
- [ ] Implement `ExtractStaticComponents()` - scan template files for `<ComponentName />` patterns
- [ ] Implement `ExtractGlobalComponents()` - scan layouts/global/* for always-loaded components
- [ ] Create `ComponentUsageMap` type mapping pages → components
- [ ] Add categorization: build-time vs runtime resolution
- [ ] Write tests for each extraction method

### 1.2 Common Chunk Generation
- [ ] Create `builder/common_chunk.go`
- [ ] Implement `CalculateComponentPageCount()` - count pages per component
- [ ] Implement `ExtractCommonComponents()` - components used on 2+ pages
- [ ] Implement `GenerateCommonChunkJS()` - generate common.js bundle
- [ ] Add content hash generation for cache busting
- [ ] Write tests for threshold logic

### 1.3 Per-Page Bundle Generation
- [ ] Create `builder/page_bundle.go`
- [ ] Implement `GeneratePageBundle()` - generate bundle with unique components
- [ ] Implement `GeneratePageBundleJS()` - output format with common import
- [ ] Add content hash to bundle filenames
- [ ] Handle pages with no unique components (common-only)
- [ ] Write tests for bundle generation

### 1.4 Bundle Manifest Generation
- [ ] Create `builder/bundle_manifest.go`
- [ ] Define `BundleManifest` struct matching spec format
- [ ] Implement `GenerateManifest()` - create bundle-manifest.json
- [ ] Include version, paths, component lists, sizes
- [ ] Add stats calculation (average savings, orphan count)
- [ ] Write to `generated/bundle-manifest.json`
- [ ] Write tests for manifest format

### 1.5 Bundle Hash Generation
- [ ] Create `builder/bundle_hash.go`
- [ ] Implement `GenerateBundleHash()` - SHA256 truncated to 8 chars
- [ ] Ensure deterministic output (sorted component order)
- [ ] Write tests for hash consistency

### 1.6 Tree-Shaking Orchestration
- [ ] Create `builder/tree_shaking.go`
- [ ] Implement `GenerateTreeShakenBundles()` - main entry point
- [ ] Wire together: usage → common → page bundles → manifest
- [ ] Ensure fallback registry (layouts.js) still generated
- [ ] Add CLI flag: `--tree-shake` (default: true)
- [ ] Write integration tests

---

## Phase 2: Runtime Loading

### 2.1 Update core/main.js
- [ ] Add bundle loading from data attributes
- [ ] Implement common chunk loading (if present)
- [ ] Implement page bundle loading
- [ ] Add fallback registry loading on error
- [ ] Expose `window.$componentRegistry`
- [ ] Add logging for fallback usage
- [ ] Test with Alpine.js initialization

### 2.2 Update HTML Renderer
- [ ] Modify `renderer/script_tags.go` or equivalent
- [ ] Generate script tag with data-bundle attribute
- [ ] Generate data-common attribute when page uses common chunk
- [ ] Generate data-fallback attribute
- [ ] Integrate with Plenti fingerprint directory
- [ ] Write tests for HTML output

---

## Phase 3: Server Integration

### 3.1 Development Server
- [ ] Update `cmd/server/main.go` to call tree-shaking
- [ ] Serve bundles from generated/bundles/
- [ ] Regenerate on file change (watch mode)
- [ ] Add bundle size logging to startup

### 3.2 Static Build
- [ ] Integrate with static site generation
- [ ] Ensure bundles copied to public/{fingerprint}/
- [ ] Verify manifest paths are correct for production

---

## Phase 4: Testing & Validation

### 4.1 Unit Tests
- [ ] `builder/component_usage_test.go`
- [ ] `builder/common_chunk_test.go`
- [ ] `builder/page_bundle_test.go`
- [ ] `builder/bundle_manifest_test.go`
- [ ] `builder/tree_shaking_test.go`

### 4.2 Integration Tests
- [ ] `tests/integration/tree_shaking_test.go`
- [ ] Test: Bundle only contains expected components
- [ ] Test: Common chunk contains shared components
- [ ] Test: Fallback registry still works
- [ ] Test: Dynamic component resolution from bundle
- [ ] Test: Cache busting hashes change on content change

### 4.3 Manual Testing
- [ ] Verify homepage loads with tree-shaken bundle
- [ ] Verify about page loads with smaller bundle
- [ ] Verify dynamic components still work
- [ ] Verify fallback triggers for runtime components
- [ ] Check browser DevTools for correct bundle sizes

---

## Definition of Done

- [ ] All tests pass
- [ ] Bundle sizes match expected from discovery (~84% reduction)
- [ ] Fallback works for runtime component resolution
- [ ] No regressions in existing functionality
- [ ] Documentation updated in CLAUDE.md
- [ ] Spec marked complete

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
