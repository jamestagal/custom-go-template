# Hydration Directives - Implementation Tasks

## Overview

**Estimated Total:** 16-24 hours
**Priority:** High (major performance impact)

---

## Phase 1: AST & Parser Foundation

### Task 1.1: Add HydrationDirective to AST
**File:** `ast/ast.go`
**Effort:** 1 hour

- [ ] Add `HydrationDirective` struct
- [ ] Add `Hydration *HydrationDirective` field to `ComponentNode`
- [ ] Add `Hydration *HydrationDirective` field to `DynamicComponentByNameNode`
- [ ] Add helper method `IsDeferred() bool`

```go
type HydrationDirective struct {
    Type  string // "load", "visible", "idle", "media", "none"
    Value string // Optional (media query, rootMargin)
}
```

### Task 1.2: Parse client:* Attributes
**File:** `parser/components.go`
**Effort:** 2 hours

- [ ] Create `parseHydrationDirective()` function
- [ ] Integrate into `parseComponentNode()`
- [ ] Integrate into `parseDynamicComponentByName()`
- [ ] Handle boolean form: `client:visible`
- [ ] Handle value form: `client:media="(max-width: 768px)"`

### Task 1.3: Parser Unit Tests
**File:** `parser/hydration_directive_test.go`
**Effort:** 1 hour

- [ ] Test parsing `client:load`
- [ ] Test parsing `client:visible`
- [ ] Test parsing `client:visible="300px"`
- [ ] Test parsing `client:idle`
- [ ] Test parsing `client:media="(...)"`
- [ ] Test parsing `client:none`
- [ ] Test no directive (default)

---

## Phase 2: Transformer Changes

### Task 2.1: Build-Time Component Hydration Wrapper
**File:** `transformer/components.go`
**Effort:** 3 hours

- [ ] Create `wrapWithHydrationContainer()` function
- [ ] Extract x-data from transformed content
- [ ] Move x-data to `data-x-data` attribute
- [ ] Add `data-hydrate` attribute
- [ ] Add `data-hydrate-value` attribute
- [ ] Handle `client:none` (strip Alpine directives)
- [ ] Preserve static HTML content

### Task 2.2: Runtime Component Hydration Support
**File:** `transformer/dynamic_component_by_name.go`
**Effort:** 2 hours

- [ ] Update `emitRuntimeWrapper()` to include hydration attributes
- [ ] Conditionally include `x-init` (only for immediate)
- [ ] Pass hydration directive to runtime wrapper

### Task 2.3: Transformer Unit Tests
**File:** `transformer/hydration_test.go`
**Effort:** 2 hours

- [ ] Test `client:visible` produces correct HTML
- [ ] Test `client:idle` produces correct HTML
- [ ] Test `client:media` produces correct HTML
- [ ] Test `client:none` strips x-data
- [ ] Test default (no directive) unchanged
- [ ] Test runtime component with hydration

---

## Phase 3: Runtime JavaScript

### Task 3.1: Create Hydration Manager
**File:** `static/js/hydration-manager.js`
**Effort:** 3 hours

- [ ] Create `HydrationManager` class
- [ ] Implement `init()` method
- [ ] Implement `hydrateImmediate()`
- [ ] Implement `hydrateOnVisible()` with IntersectionObserver
- [ ] Implement `hydrateOnIdle()` with requestIdleCallback
- [ ] Implement `hydrateOnMedia()` with matchMedia
- [ ] Implement `hydrate()` core method
- [ ] Add Safari fallback for requestIdleCallback
- [ ] Implement `destroy()` for cleanup

### Task 3.2: Integrate with Alpine.js
**File:** `static/js/runtime-components.js`
**Effort:** 1 hour

- [ ] Export `$renderDynamicComponent` globally
- [ ] Ensure proper initialization order
- [ ] Test integration with HydrationManager

### Task 3.3: Runtime Unit Tests
**File:** `static/js/hydration-manager.test.js` (or browser tests)
**Effort:** 2 hours

- [ ] Test IntersectionObserver callback
- [ ] Test requestIdleCallback fallback
- [ ] Test matchMedia listener
- [ ] Test hydration state tracking
- [ ] Test cleanup on destroy

---

## Phase 4: Integration

### Task 4.1: Server Integration
**File:** `cmd/server/main.go`
**Effort:** 1 hour

- [ ] Serve `hydration-manager.js`
- [ ] Ensure correct script loading order in HTML
- [ ] Add to global HTML wrapper

### Task 4.2: End-to-End Tests
**File:** `tests/integration/hydration_e2e_test.go`
**Effort:** 2 hours

- [ ] Test page with mixed hydration strategies
- [ ] Test `client:visible` with scroll simulation
- [ ] Test `client:idle` with timing
- [ ] Test `client:none` produces no JS

### Task 4.3: Performance Benchmarks
**Effort:** 1 hour

- [ ] Measure TTI before/after
- [ ] Measure TBT before/after
- [ ] Document performance gains

---

## Phase 5: Documentation

### Task 5.1: Update CLAUDE.md
**Effort:** 30 min

- [ ] Add hydration directive syntax
- [ ] Add examples
- [ ] Add best practices

### Task 5.2: Developer Guide
**File:** `docs/hydration-directives.md`
**Effort:** 1 hour

- [ ] Full syntax reference
- [ ] When to use each directive
- [ ] Performance guidelines
- [ ] Migration guide

---

## Dependency Graph

```
Phase 1 (AST/Parser)
    │
    ▼
Phase 2 (Transformer) ──► Phase 3 (Runtime JS)
    │                          │
    ▼                          ▼
Phase 4 (Integration) ◄────────┘
    │
    ▼
Phase 5 (Documentation)
```

---

## Quick Wins (Ship Early)

If you want to ship incrementally:

1. **MVP (8-10 hours):** Just `client:visible`
   - Most common use case
   - Biggest performance impact
   - Simplest implementation

2. **V2 (4-6 hours):** Add `client:idle` and `client:none`

3. **V3 (4-6 hours):** Add `client:media`

---

## Checklist Summary

### Phase 1: Foundation
- [ ] AST changes
- [ ] Parser changes
- [ ] Parser tests

### Phase 2: Transformer
- [ ] Build-time wrapper
- [ ] Runtime wrapper
- [ ] Transformer tests

### Phase 3: Runtime
- [ ] HydrationManager
- [ ] Alpine integration
- [ ] Runtime tests

### Phase 4: Integration
- [ ] Server setup
- [ ] E2E tests
- [ ] Benchmarks

### Phase 5: Docs
- [ ] CLAUDE.md
- [ ] Developer guide
