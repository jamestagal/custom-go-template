# Registry & allLayouts Simplification Specification

**Date:** 2026-01-28
**Status:** Discovery
**Priority:** P1 - Code Cleanup
**Based On:** Comprehensive audit of registry-related code

---

## Problem Statement

The current codebase contains significant infrastructure for component registry and tree-shaking that was cargo-culted from Svelte patterns. With the Go template engine's **build-time component resolution**, most of this infrastructure is unnecessary.

### The Svelte Problem (Why allLayouts Existed)

```
Svelte SSR → Can't resolve dynamic components at SSR time →
Must pass ALL component constructors to client → allLayouts hack
```

### The Go Solution (Why allLayouts Is Largely Unnecessary)

```
Go Template → Build-time loop expansion → Component names resolved during transformation →
Components INLINED into HTML → No client-side registry needed*
```

*Except for truly runtime cases (store-based names, user-generated names)

---

## Audit Findings

### Files That Can Be REMOVED

| File | Lines | Purpose | Why Removable |
|------|-------|---------|---------------|
| `builder/tree_shaking.go` | 283 | Bundle orchestration | SSR makes this unnecessary |
| `builder/tree_shaking_test.go` | 439 | Tests | Associated tests |
| `builder/page_bundle.go` | 214 | Per-page bundles | SSR makes this unnecessary |
| `builder/common_chunk.go` | 158 | Common chunk | SSR makes this unnecessary |
| `builder/bundle_hash.go` | 63 | Hash generation | Only used by tree-shaking |
| `builder/bundle_manifest.go` | 260 | Manifest generation | Only used by tree-shaking |
| `builder/component_usage.go` | 418 | Usage analysis | Only used by tree-shaking |
| **Total** | **~1,835** | | |

### Files That Can Be SIMPLIFIED

| File | Current | Simplified | Purpose |
|------|---------|------------|---------|
| `builder/registry_generator.go` | Always generates | Only on-demand | Generate registry only when runtime components detected |
| `core/main.js` | Loads full registry | Conditional | Only import allLayouts when needed |
| `core/runtime-components.js` | Always included | Conditional | Only inject when runtime wrappers present |

### Files To KEEP AS-IS

| File | Purpose | Why Essential |
|------|---------|---------------|
| `transformer/dynamic_component_by_name.go` | Build/runtime routing | Core decision logic - works correctly |
| `analyzer/scope.go` | Runtime expression detection | Determines build vs runtime path |

---

## What Changes

### Before (Current Architecture)

```
Build Process:
1. Parse templates
2. Register ALL components
3. Generate full layouts.js (65 components)
4. Generate tree-shaken bundles
5. Generate per-page bundles
6. Generate common chunk
7. Generate bundle manifest
8. Render pages

Runtime:
- Load layouts.js (all components)
- Load tree-shaken bundle
- $renderDynamicComponent for runtime cases
```

### After (Simplified Architecture)

```
Build Process:
1. Parse templates
2. Register ALL components (for build-time resolution)
3. Render pages (components INLINED)
4. IF runtime components detected:
   - Generate minimal layouts.js (only runtime-needed)
   - Include runtime-components.js

Runtime (only when needed):
- Load minimal layouts.js
- $renderDynamicComponent for edge cases
```

---

## Implementation Plan

### Phase 1: Remove Tree-Shaking (2-3 hours)

**Files to delete:**
```
builder/tree_shaking.go
builder/tree_shaking_test.go
builder/page_bundle.go
builder/common_chunk.go
builder/bundle_hash.go
builder/bundle_manifest.go
builder/component_usage.go
```

**Files to update:**
- `cmd/server/main.go` - Remove tree-shaking calls
- Remove `generated/bundles/` directory creation

**Risk:** Low - This code path was added recently and is not entangled

### Phase 2: Conditional Registry Generation (2-3 hours)

**Add detection:**
```go
// transformer/runtime_tracker.go

var hasRuntimeComponents bool

func MarkRuntimeComponentUsed() {
    hasRuntimeComponents = true
}

func HasRuntimeComponents() bool {
    return hasRuntimeComponents
}
```

**Update registry generation:**
```go
// builder/registry_generator.go

func GenerateComponentRegistry(components map[string]*types.ComponentTemplate) (string, error) {
    // Only generate if runtime components were detected
    if !transformer.HasRuntimeComponents() {
        return "", nil // No registry needed
    }

    // ... existing code
}
```

### Phase 3: Conditional Script Injection (1-2 hours)

**Update HTML output:**
```go
// renderer/plenti_html.go

func renderScripts(hasRuntimeComponents bool) string {
    scripts := []string{
        `<script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>`,
    }

    if hasRuntimeComponents {
        scripts = append(scripts,
            `<script type="module" src="/core/runtime-components.js"></script>`,
            `<script type="module" src="/generated/layouts.js"></script>`,
        )
    }

    return strings.Join(scripts, "\n")
}
```

### Phase 4: Clean Up Server (1 hour)

**Remove from cmd/server/main.go:**
- `generateTreeShakenBundles()` call
- Tree-shaking summary logging
- Bundle serving routes (if no longer needed)

**Update CLAUDE.md:**
- Remove tree-shaking documentation
- Update architecture diagrams

---

## Code Reduction Summary

| Category | Lines Removed | Lines Added | Net Change |
|----------|---------------|-------------|------------|
| Tree-shaking files | -1,835 | 0 | -1,835 |
| Conditional logic | 0 | +50 | +50 |
| Test cleanup | -439 | 0 | -439 |
| **Total** | | | **-2,224 lines** |

---

## Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| Breaking runtime component resolution | Keep `dynamic_component_by_name.go` and `runtime-components.js` intact |
| Removing needed functionality | Run all tests before/after |
| Future need for tree-shaking | Archive files instead of delete |

---

## Testing Plan

### Before Removal

```bash
# Run all existing tests
go test ./... -v

# Verify runtime component resolution still works
go test ./transformer -run TestDynamicComponent -v
go test ./tests/integration -run TestRuntimeComponentResolution -v
```

### After Removal

```bash
# Same tests should still pass
go test ./... -v

# Verify pages render correctly
go run cmd/server/main.go
# Visit http://localhost:3333 and check jim-test page
```

---

## Success Criteria

1. **All existing tests pass** (except tree-shaking tests which are removed)
2. **Pages render correctly** - No visual regression
3. **Runtime components still work** - `<Component:dynamic>` with runtime names
4. **Build-time components work** - `<Component:dynamic>` with build-time names
5. **~2,200 lines removed** - Significant code reduction
6. **No allLayouts on static pages** - Registry only loaded when needed

---

## Estimated Effort

| Phase | Hours |
|-------|-------|
| Remove Tree-Shaking | 2-3 |
| Conditional Registry | 2-3 |
| Conditional Scripts | 1-2 |
| Server Cleanup | 1 |
| Testing | 1-2 |
| Documentation | 1 |
| **Total** | **8-12** |

---

## Decision: Archive vs Delete

**Recommendation: ARCHIVE**

Move tree-shaking files to `builder/archive/` instead of deleting:
- Preserves git history at readable location
- Easy to restore if needed
- Documents architectural decision

```bash
mkdir -p builder/archive/tree-shaking-2026-01-28
mv builder/tree_shaking*.go builder/archive/tree-shaking-2026-01-28/
mv builder/page_bundle.go builder/archive/tree-shaking-2026-01-28/
mv builder/common_chunk.go builder/archive/tree-shaking-2026-01-28/
mv builder/bundle_*.go builder/archive/tree-shaking-2026-01-28/
mv builder/component_usage.go builder/archive/tree-shaking-2026-01-28/
```

---

## Related Specs

| Spec | Relationship |
|------|-------------|
| Plenti Integration API | Simplification enables cleaner API |
| Hydration Directives | Can be implemented without registry overhead |
| LoadAllContent | Unaffected by this change |
