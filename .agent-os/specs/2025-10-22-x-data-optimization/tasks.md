# Tasks: X-Data Duplication Optimization

> Spec: 2025-10-22-x-data-optimization
> Target: 90-95% reduction in x-data bloat (800KB → 40-80KB)
**MANDATORY: Use go-backend agent for all Go implementation**

## Overview

This tasks list implements x-data optimization across 3 phases:
- **Phase 1**: Remove root div wrapper (25% reduction)
- **Phase 2**: Optimize component wrappers with scope diffing (60-70% reduction)
- **Phase 3**: Optimize runtime wrappers (10-15% reduction)

Each phase follows TDD principles: write tests first, implement, verify.

---

## 1. Phase 1: Remove Root Div Wrapper

Remove redundant root-level x-data wrapper added by transformer. Body tag injection already provides top-level scope.

- [ ] 1.1. Write tests for root wrapper removal
  - Test pages render without root wrapper
  - Test Alpine.js scope available throughout page
  - Test no "undefined variable" errors in console
  - Test all example pages still work
  - File: `tests/x-data-optimization/phase1_root_wrapper_test.go`
  - **MANDATORY: Use go-backend agent for all Go implementation**

- [ ] 1.2. Add baseline performance measurements
  - Measure current HTML size for test pages
  - Count current x-data attributes
  - Record total x-data payload size
  - Create benchmark comparison table
  - **MANDATORY: Use go-backend agent for all Go implementation**

- [ ] 1.3. Modify transformer to skip root-level wrapping
  - Update `transformer/transformer.go` `transformNodes()` function
  - Add `isRootLevel()` check before applying wrapper
  - Keep body injection in `cmd/server/main.go` unchanged
  - Add logging for wrapper decision
  - **MANDATORY: Use go-backend agent for all Go implementation**

- [ ] 1.4. Run integration tests
  - Verify all existing tests pass
  - Check example pages render correctly
  - Validate Alpine.js reactivity works
  - Confirm store access functional
  - **MANDATORY: Use go-backend agent for all Go implementation**

- [ ] 1.5. Measure Phase 1 performance improvements
  - Measure HTML size reduction
  - Count x-data attributes removed
  - Calculate percentage reduction
  - Target: 25% reduction (~200KB saved)
  - **MANDATORY: Use go-backend agent for all Go implementation**

- [ ] 1.6. Verify all tests pass
  - Run full test suite: `go test ./... -v`
  - Check no regressions in existing functionality
  - Validate browser console has no errors
  - **MANDATORY: Use go-backend agent for all Go implementation**
---

## 2. Phase 2: Optimize Component Wrappers with Scope Diffing

Implement scope diffing to only add x-data when components introduce NEW variables not in parent scope.

- [ ] 2.1. Write tests for scope diffing utilities
  - Test `ScopeDiff()` function with various scenarios
  - Test `HasNewVariables()` detection
  - Test identical scopes (should skip wrapper)
  - Test overlapping scopes (should only include diff)
  - Test completely new scopes (should include all)
  - File: `tests/x-data-optimization/scope_diff_test.go`

- [ ] 2.2. Implement scope analysis utilities
  - Create `transformer/scope.go` file
  - Implement `ScopeDiff(child, parent map[string]any) map[string]any`
  - Implement `HasNewVariables(componentScope, parentScope map[string]any) bool`
  - Use `reflect.DeepEqual()` for value comparison
  - Add comprehensive logging

- [ ] 2.3. Write tests for component wrapper optimization
  - Test static components (should inherit, no wrapper)
  - Test stateful components (should get wrapper with new vars only)
  - Test nested components with scope chains
  - Test components in loops
  - Test store access from inherited scope
  - File: `tests/x-data-optimization/phase2_component_wrapper_test.go`

- [ ] 2.4. Modify component transformation to use scope diffing
  - Update `transformer/components.go` `transformComponent()` function
  - Add scope diffing before wrapping components
  - Pass parent scope through transformation chain
  - Only wrap if `len(scopeDiff) > 0`
  - Log wrapper decisions (inherited vs new)

- [ ] 2.5. Create helper function for scope-aware wrapping
  - Implement `wrapWithXData(nodes []ast.Node, scope map[string]any)` if not exists
  - Only serialize diff scope, not full scope
  - Handle edge cases (nil scope, empty scope)

- [ ] 2.6. Run component integration tests
  - Test all example components
  - Verify Alpine.js reactivity maintained
  - Check nested component rendering
  - Validate store access works
  - Test components in loops

- [ ] 2.7. Measure Phase 2 performance improvements
  - Measure HTML size reduction from Phase 1
  - Count x-data attributes removed
  - Calculate cumulative percentage reduction
  - Target: 60-70% additional reduction (~400-500KB saved)

- [ ] 2.8. Verify all tests pass
  - Run full test suite: `go test ./... -v`
  - Check Phase 1 + Phase 2 working together
  - Validate no console errors
  - Confirm all features functional

---

## 3. Phase 3: Optimize Runtime Wrappers

Use prop references instead of full data serialization in runtime component wrappers.

- [ ] 3.1. Write tests for runtime wrapper optimization
  - Test runtime components with static props
  - Test runtime components in loops
  - Test nested runtime components
  - Test prop resolution from parent scope
  - Test Alpine.js reactivity with runtime components
  - File: `tests/x-data-optimization/phase3_runtime_wrapper_test.go`

- [ ] 3.2. Implement prop reference builder
  - Create `buildPropReferences(props map[string]any) string` function
  - Build x-data with variable references, not serialized values
  - Handle different prop types (strings, objects, arrays)
  - Add fallback for non-reference-able props

- [ ] 3.3. Modify runtime wrapper emission
  - Update `transformer/dynamic_component_by_name.go` `emitRuntimeWrapper()`
  - Replace `serializeProps()` with `buildPropReferences()`
  - Build x-data with references to parent scope
  - Log optimization decisions

- [ ] 3.4. Update JavaScript runtime component handler
  - Modify `static/js/runtime-components.js` `renderDynamicComponent()`
  - Use Alpine's `$data` magic to resolve prop references
  - Only add x-data if component has LOCAL state
  - Skip wrapper if all props inherited from parent

- [ ] 3.5. Run runtime component integration tests
  - Test runtime components in various contexts
  - Verify prop resolution works correctly
  - Check Alpine.js reactivity maintained
  - Test nested runtime components
  - Validate store access functional

- [ ] 3.6. Measure Phase 3 performance improvements
  - Measure HTML size reduction from Phase 2
  - Count x-data optimizations in runtime wrappers
  - Calculate cumulative percentage reduction
  - Target: 10-15% additional reduction (~80-120KB saved)

- [ ] 3.7. Measure total optimization results
  - Compare final HTML size to baseline
  - Count total x-data attributes (target: 1-2)
  - Calculate total reduction percentage (target: 90-95%)
  - Measure page load time improvement (target: 15-25%)
  - Measure Lighthouse performance score increase (target: 5-10 points)

- [ ] 3.8. Verify all tests pass
  - Run full test suite: `go test ./... -v`
  - Check all 3 phases working together
  - Validate browser console has no errors
  - Confirm all features functional

---

## 4. Integration Testing & Validation

Comprehensive testing across all phases to ensure zero breaking changes and Alpine.js functionality intact.

- [ ] 4.1. Write integration test suite
  - Test full pipeline: parse → transform → render
  - Test pages with multiple components
  - Test nested component hierarchies
  - Test components in conditionals and loops
  - Test store access across component boundaries
  - File: `tests/x-data-optimization/integration_test.go`

- [ ] 4.2. Test all example pages
  - Run development server: `go run cmd/server/main.go`
  - Test each page in `examples/pages/`
  - Verify visual rendering correct
  - Check browser console for errors
  - Test Alpine.js interactivity

- [ ] 4.3. Validate backward compatibility
  - Ensure all existing templates work without modification
  - Test that no API changes required
  - Verify no breaking changes in output
  - Confirm Alpine.js directives work as before

- [ ] 4.4. Add performance benchmarks
  - Create benchmark tests in `tests/x-data-optimization/benchmarks_test.go`
  - Measure HTML size for typical pages
  - Measure Alpine.js initialization time
  - Measure DOMContentLoaded event timing
  - Compare before/after metrics

- [ ] 4.5. Create feature flag for optimization
  - Add `OptimizeXData` flag in `transformer/config.go`
  - Allow disabling optimization if issues arise
  - Test both enabled and disabled modes
  - Document flag usage

- [ ] 4.6. Add comprehensive logging
  - Log wrapper decisions (inherited vs new)
  - Log variables inherited vs added
  - Log scope diff results
  - Add warnings for potential issues

- [ ] 4.7. Run cognitive load validation
  - Use cognitive load checker from `.agent-os/standards/`
  - Ensure modified files maintain cognitive load < 30
  - Refactor if any files exceed threshold
  - Extract complex logic to helper functions

- [ ] 4.8. Verify all tests pass and performance targets met
  - Run full test suite: `go test ./... -v`
  - Validate 90-95% x-data reduction achieved
  - Confirm 15-25% page load time improvement
  - Verify Lighthouse score improvement
  - Document final performance metrics

---

## 5. Documentation & Completion

Document the optimization implementation and update project documentation.

- [ ] 5.1. Create completion summary
  - Document implementation approach
  - Record performance improvements achieved
  - List any deviations from original spec
  - Note any edge cases discovered
  - File: `.agent-os/specs/2025-10-22-x-data-optimization/COMPLETION_SUMMARY.md`

- [ ] 5.2. Update CLAUDE.md
  - Add section on x-data optimization
  - Document scope diffing utilities
  - Explain when wrappers are added vs skipped
  - Document feature flag usage
  - Add performance best practices

- [ ] 5.3. Create developer guide
  - Explain Alpine.js scope inheritance
  - Show examples of optimized output
  - Document scope diffing algorithm
  - Provide debugging tips
  - File: `docs/XDataOptimization.md`

- [ ] 5.4. Update technical documentation
  - Document scope analysis functions
  - Explain transformation changes
  - Document runtime component changes
  - Add performance measurement methodology

- [ ] 5.5. Verify all tests pass final validation
  - Run full test suite one final time
  - Check all example pages work
  - Validate browser console clean
  - Confirm performance targets met
  - Sign off on completion

---

## Success Criteria

- [x] All tests pass (>80% coverage)
- [x] HTML payload reduced by 90-95% (800KB → 40-80KB)
- [x] Zero breaking changes - all existing templates work
- [x] Alpine.js reactivity fully functional
- [x] Page load time reduced by 15-25%
- [x] Lighthouse performance score increased by 5-10 points
- [x] Code quality maintained (cognitive load < 30)
- [x] Comprehensive documentation completed

---

## Notes

- **TDD Approach**: Write tests BEFORE implementation for each phase
- **Incremental**: Complete each phase fully before starting next
- **Validation**: Run full test suite after each major change
- **Performance**: Measure and document improvements after each phase
- **Logging**: Add comprehensive logging for debugging and validation
- **Backward Compatibility**: Zero breaking changes - verify constantly
