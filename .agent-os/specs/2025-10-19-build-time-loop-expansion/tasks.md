# Spec Tasks

These are the tasks to be completed for the spec detailed in @.agent-os/specs/2025-10-19-build-time-loop-expansion/spec.md
**MANDATORY: Use go-backend agent for all Go implementation**
> Created: 2025-10-19
> Status: In Progress

## Tasks

- [x] 1. Implement scope cloning utilities
  - [x] 1.1 Write tests for scope cloning (test deep vs shallow copy behavior)
  - [x] 1.2 Create `cloneScope` function in transformer/scope.go (or add to existing scope utilities)
  - [x] 1.3 Create `resolveCollectionFromScope` function to look up arrays in dataScope
  - [x] 1.4 Add error handling for missing or non-array collections
  - [x] 1.5 Verify all scope utility tests pass
  - [x] 1.6 **MANDATORY: Use go-backend agent for all Go implementation**

- [x] 2. Modify loop transformer for build-time expansion
  - [x] 2.1 Write tests for build-time loop expansion (simple array, component array, nested loops)
  - [x] 2.2 Read current transformer/loops.go implementation to understand existing structure
  - [x] 2.3 Refactor TransformFor function to iterate arrays in Go instead of creating x-for templates
  - [x] 2.4 Add collection resolution from dataScope
  - [x] 2.5 Implement iteration loop: clone scope, add loop variable, transform body
  - [x] 2.6 Handle edge cases (empty arrays, missing collections, type mismatches)
  - [x] 2.7 Update any helper functions that depend on x-for generation
  - [x] 2.8 Verify all loop transformer tests pass
  - [x] 2.9 **MANDATORY: Use go-backend agent for all Go implementation**

- [x] 3. Integration with component resolution
  - [x] 3.1 Write integration tests using component loops from JSON data
  - [x] 3.2 Test that component.name resolves correctly with loop variables in scope
  - [x] 3.3 Verify dynamic component resolution (transformer/dynamic_component_by_name.go) works without modifications
  - [x] 3.4 Test nested property access (e.g., component.fields.title)
  - [x] 3.5 Verify all integration tests pass
  - [x] 3.6 **MANDATORY: Use go-backend agent for all Go implementation**

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
