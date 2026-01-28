# Registry Simplification Tasks

**Spec:** [spec.md](spec.md)
**Branch:** `feature/registry-simplification`
**Status:** All Phases Complete ✅

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

## Phase 4: Server Cleanup ✅

**Goal:** Remove obsolete code paths and update documentation
**Completed:** 2026-01-28

### 4.1 Clean Up cmd/server/main.go

- [x] Tree-shaking code already removed in Phase 1
- [x] No bundle-related routes remaining
- [x] No `generated/bundles/` directory creation

### 4.2 Clean Up Generated Directories

- [x] `generated/bundles/` already removed in Phase 1
- [x] `generated/layouts.js` verified working

### 4.3 Update CLAUDE.md

- [x] No tree-shaking section to remove (was never added)
- [x] Updated architecture description for `builder/` and `cmd/server/`
- [x] Added "Conditional Script Injection" section documenting the new system
- [x] Updated runtime component resolution section with correct file paths

### 4.4 Update Other Documentation

- [x] Removed unused bundle fields from `renderer/plenti_html.go`
- [x] Removed `addBundleDataAttributes()` function (dead code)
- [x] Updated integration tests to reflect conditional script injection

---

## Post-Implementation ✅

**Completed:** 2026-01-28

### Verification

- [x] Full test suite passes: `go test ./... -v`
- [x] Server starts and serves pages correctly
- [x] jim-test page renders correctly (visit http://localhost:3333/jim-test)
- [x] Runtime component resolution still works
- [x] Build-time component resolution still works

### Code Quality

- [x] Run `go fmt ./...`
- [x] Review for any orphaned imports (build succeeds with no warnings)
- [ ] Run `golangci-lint run` (if configured) - skipped, not configured

### Metrics

- [x] Lines removed: **1,835 lines** (tree-shaking infrastructure archived)
- [x] Lines added: **~84 lines** (runtime tracking system)
- [x] Net reduction: **~1,751 lines**
- [x] Behavior change: Static pages no longer include runtime-components.js or layouts.js

---

## Success Criteria ✅

1. [x] All existing tests pass (except removed tree-shaking tests)
2. [x] Pages render correctly with no visual regression
3. [x] Runtime components (`<Component:dynamic>` with runtime names) work
4. [x] Build-time components (`<Component:dynamic>` with build-time names) work
5. [x] ~1,751 lines net reduction from active codebase (1,835 removed - 84 added)
6. [x] No runtime-components.js/layouts.js on static pages (conditional injection)

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
