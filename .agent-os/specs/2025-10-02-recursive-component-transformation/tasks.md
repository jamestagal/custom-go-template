# Spec Tasks

These are the tasks to be completed for the spec detailed in @.agent-os/specs/2025-10-02-recursive-component-transformation/spec.md

> Created: 2025-10-02
> Status: Ready for Implementation

## Tasks

### 1. Implement Helper Functions for Component Transformation

**Purpose**: Build foundational utilities for fence data processing, prop resolution, and value parsing.

**1.1** Write tests for `parseValue()` helper function
- Test boolean parsing (`"true"`, `"false"`)
- Test null parsing (`"null"`)
- Test number parsing (integers and floats)
- Test string parsing (single and double quoted)
- Test array literals (should return as string for Alpine)
- Test object literals (should return as string for Alpine)
- Test expression/variable references (should return as string)
- Location: `transformer/components_test.go` or new `transformer/helpers_test.go`

**1.2** Implement `parseValue()` helper function
- Extract from `cmd/server/main.go` if it exists there
- Handle all test cases from 1.1
- Place in `transformer/components.go` or `transformer/utils.go`
- Add debug logging for parsed values

**1.3** Write tests for `filterOutFence()` helper function
- Test with nodes containing one FenceSection
- Test with nodes containing multiple FenceSections
- Test with nodes containing no FenceSection
- Test preserving order of non-fence nodes
- Location: `transformer/components_test.go`

**1.4** Implement `filterOutFence()` helper function
- Simple filter that removes `*ast.FenceSection` nodes
- Preserve order of remaining nodes
- Return new slice without modifying input
- Place in `transformer/components.go`

**1.5** Write tests for `collectComponentFenceData()` helper function
- Test extracting variables from fence.Variables
- Test extracting prop defaults from fence.Props
- Test extracting function declarations from fence.RawContent
- Test extracting function expressions (const/let/var)
- Test extracting arrow functions
- Test extracting method shorthand syntax
- Test with empty fence section
- Test scope map is correctly populated
- Location: `transformer/components_test.go`

**1.6** Implement `collectComponentFenceData()` helper function
- Process `fence.Variables` using `parseValue()`
- Process `fence.Props` to add defaults to scope
- Extract functions using regex patterns (as specified in technical spec)
- Store function definitions as raw strings in scope
- Add debug logging for each collected item
- Place in `transformer/components.go` or `transformer/scope.go`

**1.7** Write tests for `resolvePropValue()` helper function
- Test dynamic prop resolution (IsDynamic=true) from parent scope
- Test shorthand prop resolution (IsShorthand=true) from parent scope
- Test static prop parsing with `parseValue()`
- Test fallback when prop not found in parent scope
- Test all prop value types (string, number, boolean, expression)
- Location: `transformer/components_test.go`

**1.8** Implement `resolvePropValue()` helper function
- Handle dynamic props: look up expression in parentScope
- Handle shorthand props: look up prop.Name in parentScope
- Handle static props: parse using `parseValue()`
- Return resolved value or expression string if not in parent
- Add debug logging showing prop name, type, and resolved value
- Place in `transformer/components.go`

### 2. Implement x-data Wrapper and Core Component Transformation

**Purpose**: Create the wrapper logic and refactor the main transformComponent function to use recursive transformation.

**2.1** Write tests for `addComponentDataWrapper()` helper function
- Test single root element: x-data added as attribute
- Test multiple root nodes: wrapper div created with x-data
- Test empty dataScope (should still wrap)
- Test dataScope formatting with variables, props, and functions
- Test preserving existing attributes on root element
- Location: `transformer/components_test.go`

**2.2** Implement `addComponentDataWrapper()` helper function
- Format dataScope using `alpineDataFormatter()`
- Single root element: add x-data attribute directly
- Multiple roots: create `<div x-data="...">` wrapper with children
- Handle empty dataScope gracefully
- Return properly wrapped nodes
- Place in `transformer/components.go` or `transformer/alpine.go`

**2.3** Write comprehensive tests for refactored `transformComponent()` function
- Test component lookup from registry (success and failure)
- Test isolated scope creation for component instance
- Test fence data collection from component template
- Test prop resolution from parent to component scope
- Test body transformation with component's isolated scope
- Test final x-data wrapper application
- Test nested components (component using another component)
- Test components in conditionals
- Test components in loops
- Test dynamic component references (`<{componentVar} />`)
- Location: `transformer/components_test.go` and `tests/alpine/components_test.go`

**2.4** Refactor `transformComponent()` function - Phase 1: Component Lookup
- Implement component template lookup using `GetComponentTemplate()`
- Add error handling for component not found
- Return comment placeholder on error
- Add debug logging for component transformation start
- Keep existing placeholder code temporarily (will remove in 2.5)

**2.5** Refactor `transformComponent()` function - Phase 2: Remove Test-Specific Code
- Remove lines 172-272 in `transformer/components.go` (all `isAlpineIntegrationTest()` special cases)
- Remove `resetRenderedComponents()` calls if test-specific
- Remove any hardcoded rendering logic for specific test cases
- Remove test-specific code in `transformer/alpine.go` (lines 24-44)

**2.6** Refactor `transformComponent()` function - Phase 3: Implement Recursive Logic
- Create isolated componentDataScope map
- Call `collectComponentFenceData()` to process component's fence
- Resolve passed props using `resolvePropValue()` in loop
- Filter out fence nodes using `filterOutFence()`
- Transform component body with `transformNodes(componentBodyNodes, componentDataScope, false)`
- Call `addComponentDataWrapper()` to wrap with x-data
- Add comprehensive debug logging at each step

**2.7** Handle dynamic component references
- Detect component names starting with `{` and ending with `}`
- Extract variable name from expression
- Add variable to scope using `extractVariablesFromExpr()`
- Use appropriate x-component directive or transformation strategy
- Add tests for this edge case

**2.8** Verify all tests pass
- Run `go test ./transformer -v`
- Run `go test ./tests/alpine -v`
- Run `go test ./tests/components -v` if exists
- Verify specific tests: `TestComponentPropsTransformation`, `TestStaticComponentTransformation`
- Verify `TestAlpineIntegration/component_integration`
- Verify `TestAlpineIntegration/nested_conditionals_and_loops`
- Check browser output at http://localhost:3000 for visual verification

### 3. Component Registry Verification and Enhancement

**Purpose**: Ensure component registration and lookup infrastructure supports the new transformation approach.

**3.1** Write tests for component registry functionality
- Test `RegisterComponent()` with valid component templates
- Test `GetComponentTemplate()` retrieval success
- Test `GetComponentTemplate()` retrieval failure
- Test registry contains parsed AST and props list
- Test multiple component registrations
- Location: `transformer/components_test.go`

**3.2** Verify and document `ComponentTemplate` struct
- Ensure struct includes `*ast.Template` field
- Ensure struct includes component name field
- Ensure struct includes props list field
- Add comments documenting the structure
- Location: `transformer/components.go`

**3.3** Verify `RegisterComponent()` function
- Ensure it stores parsed `*ast.Template`
- Ensure it extracts and stores props list
- Ensure it handles component name correctly
- Add error handling for nil templates
- Add debug logging for registration events
- Location: `transformer/components.go`

**3.4** Verify `GetComponentTemplate()` function
- Ensure it returns full ComponentTemplate struct
- Ensure it returns exists boolean
- Ensure thread-safe access if needed (document if concurrent access is possible)
- Add debug logging for lookup events
- Location: `transformer/components.go`

**3.5** Review component registration in `cmd/server/main.go`
- Verify components from `examples/components/` are being registered
- Verify fence sections are being parsed correctly
- Verify props are being extracted via `extractComponentProps()`
- Ensure registration happens before server starts handling requests
- Document the registration flow

**3.6** Verify all component registry tests pass
- Run `go test ./transformer -v -run Registry`
- Run `go test ./transformer -v -run Component`
- Verify no registration errors in server startup logs
- Test with multiple component files

### 4. Integration Testing and Validation

**Purpose**: Ensure the complete recursive transformation system works end-to-end with real-world scenarios.

**4.1** Create integration test for simple component with props
- Component with fence section containing variables and functions
- Parent passes props to component
- Verify transformed output has x-data wrapper
- Verify x-data contains component's scope (variables, functions, resolved props)
- Verify component body is properly transformed
- Location: `tests/alpine/components_test.go`

**4.2** Create integration test for nested components
- Parent component uses child component
- Both have fence sections with different variables
- Verify each component has isolated scope
- Verify parent props are resolved in child
- Verify proper x-data nesting
- Location: `tests/alpine/components_test.go`

**4.3** Create integration test for components in conditionals
- Test component inside `{if}` block
- Test component inside `{else if}` block
- Test component inside `{else}` block
- Verify component transformation happens before conditional wrapping
- Location: `tests/alpine/conditionals_test.go` or `components_test.go`

**4.4** Create integration test for components in loops
- Test component inside `{for}` loop
- Verify component receives loop variables as props
- Verify each iteration creates proper component instance
- Test with `ProductCard` example from spec
- Location: `tests/alpine/loops_test.go` or `components_test.go`

**4.5** Test browser rendering of components
- Start dev server: `go run cmd/server/main.go`
- Navigate to example pages with components
- Open browser console and check for errors
- Verify no "undefined" errors in Alpine.js expressions
- Verify component functions are callable
- Verify component reactivity works
- Document visual verification checklist

**4.6** Run full test suite and verify output
- Run `go test ./... -v`
- Verify all transformer tests pass
- Verify all alpine integration tests pass
- Verify all component tests pass
- Check for any new warnings or errors in logs
- Confirm no regression in existing tests

**4.7** Performance and edge case testing
- Test with deeply nested components (5+ levels)
- Test with component receiving 10+ props
- Test with large fence sections (many variables/functions)
- Test with component not found in registry (error handling)
- Test with component receiving undefined prop from parent
- Test with empty component (no fence, no content)
- Test with component that has only fence section

**4.8** Final validation and documentation
- Verify all acceptance criteria from spec.md are met
- Verify all technical requirements from technical-spec.md are implemented
- Run complete test suite one final time
- Test all examples in browser
- Verify debug logging is helpful and not excessive
- Document any known limitations or future enhancements needed
