# Spec Tasks

These are the tasks to be completed for the spec detailed in @.agent-os/specs/2025-10-19-build-time-loop-expansion/spec.md
**MANDATORY: Use go-backend agent for all Go implementation**
> Created: 2025-10-19
> Status: Ready for Implementation

## Tasks

- [ ] 1. Implement scope cloning utilities
  - [ ] 1.1 Write tests for scope cloning (test deep vs shallow copy behavior)
  - [ ] 1.2 Create `cloneScope` function in transformer/scope.go (or add to existing scope utilities)
  - [ ] 1.3 Create `resolveCollectionFromScope` function to look up arrays in dataScope
  - [ ] 1.4 Add error handling for missing or non-array collections
  - [ ] 1.5 Verify all scope utility tests pass
  - [ ] 1.6 **MANDATORY: Use go-backend agent for all Go implementation**

- [ ] 2. Modify loop transformer for build-time expansion
  - [ ] 2.1 Write tests for build-time loop expansion (simple array, component array, nested loops)
  - [ ] 2.2 Read current transformer/loops.go implementation to understand existing structure
  - [ ] 2.3 Refactor TransformFor function to iterate arrays in Go instead of creating x-for templates
  - [ ] 2.4 Add collection resolution from dataScope
  - [ ] 2.5 Implement iteration loop: clone scope, add loop variable, transform body
  - [ ] 2.6 Handle edge cases (empty arrays, missing collections, type mismatches)
  - [ ] 2.7 Update any helper functions that depend on x-for generation
  - [ ] 2.8 Verify all loop transformer tests pass
  - [ ] 2.9 **MANDATORY: Use go-backend agent for all Go implementation**

- [ ] 3. Integration with component resolution
  - [ ] 3.1 Write integration tests using component loops from JSON data
  - [ ] 3.2 Test that component.name resolves correctly with loop variables in scope
  - [ ] 3.3 Verify dynamic component resolution (transformer/dynamic_component_by_name.go) works without modifications
  - [ ] 3.4 Test nested property access (e.g., component.fields.title)
  - [ ] 3.5 Verify all integration tests pass
  - [ ] 3.6 **MANDATORY: Use go-backend agent for all Go implementation**

- [ ] 4. Output validation and comparison
  - [ ] 4.1 Create test cases comparing output to expected Svelte-style fully expanded HTML
  - [ ] 4.2 Verify no x-for templates appear in generated HTML
  - [ ] 4.3 Test with real component data from content/pages/*.json files
  - [ ] 4.4 Verify each component in array produces separate rendered HTML
  - [ ] 4.5 Check performance with reasonable array sizes (10-50 components)
  - [ ] 4.6 Verify all validation tests pass
  - [ ] 4.7 **MANDATORY: Use go-backend agent for all Go implementation**

- [ ] 5. Documentation and cleanup
  - [ ] 5.1 Update CLAUDE.md with build-time loop expansion behavior
  - [ ] 5.2 Add code comments explaining scope cloning and iteration logic
  - [ ] 5.3 Document any breaking changes (x-for removal from loop output)
  - [ ] 5.4 Add examples to documentation showing component loop patterns
  - [ ] 5.5 Run full test suite to ensure no regressions
  - [ ] 5.6 Update spec completion status
  - [ ] 5.7 **MANDATORY: Use go-backend agent for all Go implementation**