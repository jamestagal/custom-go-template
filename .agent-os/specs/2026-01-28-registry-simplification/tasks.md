# Registry Simplification Tasks

**Spec:** [spec.md](spec.md)
**Branch:** `feature/registry-simplification`
**Status:** Phase 1 Complete ✅

---

## Pre-Implementation

- [x] Run full test suite to establish baseline
- [x] Verify tests pass before making changes

---

## Phase 1: Archive Tree-Shaking Files ✅

**Goal:** Remove ~1,835 lines of tree-shaking infrastructure
**Completed:** 2026-01-28

### 1.1 Create Archive Directory

- [x] Create archive directory structure (`.agent-os/archive/tree-shaking-2026-01-28/`)

### 1.2 Move Tree-Shaking Files

- [x] Archive `builder/tree_shaking.go` (283 lines)
- [x] Archive `builder/tree_shaking_test.go` (439 lines)
- [x] Archive `builder/page_bundle.go` (214 lines)
- [x] Archive `builder/common_chunk.go` (158 lines)
- [x] Archive `builder/bundle_hash.go` (63 lines)
- [x] Archive `builder/bundle_manifest.go` (260 lines)
- [x] Archive `builder/component_usage.go` (418 lines)

### 1.3 Update Imports

- [x] Remove tree-shaking calls from `cmd/server/main.go`
- [x] Remove `generateTreeShakenBundles()` function
- [x] Remove `getBundlePathsForRoute()` function
- [x] Remove `routeToPageKey()` function
- [x] Remove bundle manifest global variables
- [x] Remove bundle data attributes injection code

### 1.4 Clean Up Generated Files

- [x] Remove `generated/bundles/` directory
- [x] Remove `generated/bundle-manifest.json`

### 1.5 Verify Phase 1

- [x] Code compiles without errors: `go build ./...`
- [x] Tests pass (excluding archived tests): `go test ./...`
- [x] Server starts successfully
- [x] HTML output no longer has `data-bundle`, `data-common`, `data-fallback` attributes

---

## Phase 2: Conditional Registry Generation ✅

**Goal:** Track runtime component usage for conditional script injection
**Completed:** 2026-01-28

**Note:** Registry is still generated at startup (needed if ANY page uses runtime components).
The real optimization happens in Phase 3 where scripts are conditionally injected per-page.

### 2.1 Add Runtime Component Tracking

- [x] Create `transformer/runtime_tracker.go`:
  - [x] Thread-safe tracking with mutex
  - [x] `MarkRuntimeComponentUsed()` - marks runtime component usage
  - [x] `HasRuntimeComponents()` - checks if any runtime components used
  - [x] `ResetRuntimeComponentTracking()` - resets state between builds

### 2.2 Integrate Tracking in Dynamic Component Resolution

- [x] Update `transformer/dynamic_component_by_name.go`:
  - [x] Call `MarkRuntimeComponentUsed()` when emitting runtime wrapper

### 2.3 Update Server Integration

- [x] Modify `cmd/server/main.go`:
  - [x] Reset tracking at start of registry generation
  - [x] Add documentation about conditional approach

### 2.4 Add Tests for Runtime Tracking

- [x] Create `transformer/runtime_tracker_test.go`:
  - [x] Test: Initial state after reset
  - [x] Test: Mark runtime component
  - [x] Test: Multiple marks
  - [x] Test: Reset after mark

### 2.5 Verify Phase 2

- [x] All tests pass: `go test ./...`
- [x] Server starts and serves pages correctly
- [x] Runtime tracking mechanism works

---

## Phase 3: Conditional Script Injection ✅

**Goal:** Only include runtime scripts when needed
**Completed:** 2026-01-28

**Implementation Notes:**
- Removed hardcoded `runtime-components.js` from `layouts/global/head.html`
- Added per-page tracking reset in both `renderTemplateWithProps` and `renderTemplate`
- Created `injectRuntimeScripts()` helper function in `cmd/server/main.go`
- Scripts are injected AFTER transformation completes, based on `HasRuntimeComponents()`

### 3.1 Update HTML Renderer

- [x] Modify `cmd/server/main.go` (actual location of script injection):
  - [x] Add `injectRuntimeScripts()` helper function
  - [x] Conditionally include `runtime-components.js`
  - [x] `layouts.js` loaded by runtime-components.js (no separate injection needed)

### 3.2 Update Script Output Logic

- [x] Create helper function `injectRuntimeScripts(html string) string`
- [x] Alpine.js CDN always included
- [x] Runtime scripts only when `HasRuntimeComponents() == true`

### 3.3 Integration with Build Process

- [x] Check `HasRuntimeComponents()` after transformation
- [x] Reset tracking at start of each page render (per-page isolation)

### 3.4 Add Tests for Conditional Scripts

- [x] Test: `TestInjectRuntimeScripts` - table-driven test for both cases
- [x] Test: `TestStaticPageNoRuntimeScripts` - verifies static pages don't get scripts

### 3.5 Verify Phase 3

- [x] Static pages have no `runtime-components.js`
- [x] Pages with runtime components would get scripts (when `MarkRuntimeComponentUsed()` is called)
- [x] All tests pass: `go test ./...`

---

## Phase 4: Server Cleanup

**Goal:** Remove obsolete code paths and update documentation

### 4.1 Clean Up cmd/server/main.go

- [ ] Remove `generateTreeShakenBundles()` call (if exists)
- [ ] Remove tree-shaking summary logging
- [ ] Remove bundle-related routes (if unused)
- [ ] Remove `generated/bundles/` directory creation

### 4.2 Clean Up Generated Directories

- [ ] Remove `generated/bundles/` if it exists
- [ ] Verify `generated/layouts.js` still works when needed

### 4.3 Update CLAUDE.md

- [ ] Remove tree-shaking documentation section
- [ ] Update architecture description
- [ ] Document conditional registry generation
- [ ] Add note about runtime vs build-time component resolution

### 4.4 Update Other Documentation

- [ ] Review and update any references to tree-shaking
- [ ] Update architecture diagrams if present

---

## Post-Implementation

### Verification

- [ ] Full test suite passes: `go test ./... -v`
- [ ] Server starts and serves pages correctly
- [ ] jim-test page renders correctly (visit http://localhost:3333/jim-test)
- [ ] Runtime component resolution still works
- [ ] Build-time component resolution still works

### Code Quality

- [ ] Run `go fmt ./...`
- [ ] Run `golangci-lint run` (if configured)
- [ ] Review for any orphaned imports

### Metrics

- [ ] Count lines removed vs added
- [ ] Document final code reduction
- [ ] Note any behavior changes

---

## Success Criteria

1. [ ] All existing tests pass (except removed tree-shaking tests)
2. [ ] Pages render correctly with no visual regression
3. [ ] Runtime components (`<Component:dynamic>` with runtime names) work
4. [ ] Build-time components (`<Component:dynamic>` with build-time names) work
5. [ ] ~2,200 lines removed from active codebase
6. [ ] No allLayouts/registry on static pages

---

## Rollback Plan

If issues arise:
1. Restore archived files from `builder/archive/tree-shaking-2026-01-28/`
2. Revert changes to `cmd/server/main.go`
3. Revert changes to registry generator
4. Remove runtime tracking code

---

## Estimated Time

| Phase | Estimate |
|-------|----------|
| Pre-Implementation | 15 min |
| Phase 1: Archive Files | 1-2 hours |
| Phase 2: Conditional Registry | 2-3 hours |
| Phase 3: Conditional Scripts | 1-2 hours |
| Phase 4: Server Cleanup | 1 hour |
| Post-Implementation | 30 min |
| **Total** | **6-9 hours** |
