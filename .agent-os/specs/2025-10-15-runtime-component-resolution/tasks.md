# Spec Tasks

> Spec: Runtime Component Resolution for Loop Variables
> Created: 2025-10-15
> Status: Phase 1 Complete
**Spec:** Runtime Component Resolution for Loop Variables
**Goal:** Implement runtime component resolution to enable dynamic component iteration in loops where component names are only known at runtime (e.g., `{for component in components} <Component:dynamic name={component.name} /> {/for}`).
**Status:** Phase 1 Complete - Ready for Phase 2
**MANDATORY: Use go-backend agent for all Go implementation**

## Tasks

- [x] 1. Phase 1: Scope Tracking (4-6h, Medium Cognitive Load) - **COMPLETED 2025-10-15**
  - [x] 1.1 Write tests for ScopeAnalyzer in `analyzer/scope_test.go`
    - Test build-time variable detection (content props, exported props)
    - Test runtime variable detection (loop iterators, Alpine stores)
    - Test expression analysis with mixed variables
    - Test nested loop variable tracking
    - Files: Created `analyzer/scope_test.go` ✓
  - [x] 1.2 Implement ScopeAnalyzer struct in `analyzer/scope.go`
    - Create `ScopeAnalyzer` with `buildVars` and `runtimeVars` maps
    - Implement `NewScopeAnalyzer(dataScope map[string]any)` constructor
    - Files: Created `analyzer/scope.go` ✓
  - [x] 1.3 Implement IsRuntimeExpression method
    - Return true for loop variables, Alpine stores, operators
    - Return false for string literals, content props, exported props
    - Handle nested property access (component.name, item.field)
    - Files: `analyzer/scope.go` ✓
  - [x] 1.4 Implement variable tracking methods
    - `TrackLoopVariable(name string)` - marks variable as runtime-only
    - `TrackContentProp(name string)` - marks variable as build-resolvable
    - `TrackExportedProp(name string)` - marks variable as build-resolvable
    - Files: `analyzer/scope.go` ✓
  - [x] 1.5 Add expression traversal utilities
    - `extractVariablesFromExpression(expr string)` - get all variable names (string-based)
    - Implemented helper functions: isStringLiteral, isAlpineStoreReference, hasOperators
    - Files: `analyzer/scope.go` ✓
  - [x] 1.6 Verify all tests pass for Phase 1
    - Run `go test ./analyzer -v` ✓ - ALL PASS
    - Verify edge cases: nested loops, mixed expressions ✓
  - [x] 1.7 **MANDATORY: Use go-backend agent for all Go implementation** ✓

- [ ] 2. Phase 2: Runtime Wrapper Emission (6-8h, High Cognitive Load)
  - [ ] 2.1 Write tests for runtime wrapper emission in `transformer/dynamic_component_by_name_test.go`
    - Test runtime path: component name is loop variable
    - Test build-time path: component name is string literal (regression)
    - Test mixed props: static and dynamic values
    - Test wrapper structure: x-data, x-init, class attributes
    - Files: `transformer/dynamic_component_by_name_test.go`
  - [ ] 2.2 Modify TransformDynamicComponentByName to integrate ScopeAnalyzer
    - Import and initialize ScopeAnalyzer
    - Add conditional logic: check IsRuntimeExpression(node.NameExpression)
    - Route to runtime path or build-time path
    - Files: `transformer/dynamic_component_by_name.go`
  - [ ] 2.3 Implement emitRuntimeWrapper function
    - Create wrapper Element node with class="dyn-comp-runtime"
    - Generate x-data JSON with compName and compProps
    - Generate x-init with $renderDynamicComponent call
    - Files: `transformer/dynamic_component_by_name.go`
  - [ ] 2.4 Implement props serialization for runtime wrappers
    - Convert props map to JSON for x-data attribute
    - Handle nested objects, arrays, escaped strings
    - Preserve Alpine expressions (don't escape x-text, x-bind)
    - Files: `transformer/dynamic_component_by_name.go`
  - [ ] 2.5 Add scope awareness to transformer pipeline
    - Track loop variables during AST traversal in `transformer/transform.go`
    - Pass ScopeAnalyzer to relevant transform functions
    - Update LoopNode transformation to register loop variable
    - Files: `transformer/transform.go`
  - [ ] 2.6 Verify all tests pass for Phase 2
    - Run `go test ./transformer -v -run Dynamic`
    - Check regression: static component names still resolve at build-time
    - Validate wrapper HTML structure matches spec
  - [ ] 2.8 **MANDATORY: Use go-backend agent for all Go implementation**

- [ ] 3. Phase 3: Client-Side Runtime (6-8h, Medium Cognitive Load)
  - [ ] 3.1 Create test HTML page for client-side runtime in `examples/pages/runtime-component-test.html`
    - Create page with runtime wrapper HTML (manual)
    - Create page with loop using {for component in components}
    - Include both valid and invalid component names
    - Files: Create `examples/pages/runtime-component-test.html`
  - [ ] 3.2 Create runtime-components.js with Alpine magic
    - Implement Alpine.magic('renderDynamicComponent')
    - Core function: (el, componentName, props) => void
    - Files: Create `static/js/runtime-components.js`
  - [ ] 3.3 Implement component registry loading
    - `loadComponentRegistry()` - fetch and import registry module
    - Cache registry in window.$componentRegistry
    - Handle network errors with retry logic (3 attempts, exponential backoff)
    - Files: `static/js/runtime-components.js`
  - [ ] 3.4 Implement component rendering logic
    - Get template function from registry by name
    - Call template function with props
    - Set el.innerHTML with rendered HTML
    - Files: `static/js/runtime-components.js`
  - [ ] 3.5 Add error handling and warnings
    - Missing component: insert HTML comment warning (silent)
    - Network error: log and retry
    - Template error: console.error, show dev mode message
    - Files: `static/js/runtime-components.js`
  - [ ] 3.6 Integrate runtime script into server
    - Add script tag to base layout or component rendering
    - Ensure Alpine.js loads before runtime-components.js
    - Files: `cmd/server/main.go` or `renderer/render.go`
  - [ ] 3.7 Verify client-side runtime works in browser
    - Start dev server: `go run cmd/server/main.go`
    - Navigate to test page
    - Check browser console for errors
    - Verify components render correctly
  - [ ] 3.8 **MANDATORY: Use go-backend agent for all Go implementation**

- [ ] 4. Phase 4: Registry Generation (6-8h, High Cognitive Load)
  - [ ] 4.1 Write tests for registry generator in `builder/registry_generator_test.go`
    - Test single component conversion to JS template
    - Test multiple components in registry
    - Test template conversion: {var} -> ${props.var}
    - Test Alpine directive preservation
    - Test HTML escaping and template literal escaping
    - Files: Create `builder/registry_generator_test.go`
  - [ ] 4.2 Create registry generator in `builder/registry_generator.go`
    - Implement `GenerateComponentRegistry(components []ComponentTemplate) string`
    - Create ES module export structure
    - Files: Create `builder/registry_generator.go`
  - [ ] 4.3 Implement AST to JS template conversion
    - `convertToJSTemplate(node ast.Node) string` - recursive converter
    - Handle TextNode: escape backticks, preserve content
    - Handle ExpressionNode: convert {var} to ${props.var}
    - Handle Element nodes: preserve HTML structure
    - Files: `builder/registry_generator.go`
  - [ ] 4.4 Handle Alpine.js directive preservation
    - Preserve x-text, x-bind, x-if, x-for as-is
    - Don't convert expressions inside Alpine directives
    - Special handling for x-text: add ${props.var} as fallback content
    - Files: `builder/registry_generator.go`
  - [ ] 4.5 Add registry output to build process
    - Hook into server startup to generate registry
    - Write registry to `static/js/component-registry.js`
    - Serve registry file at `/js/component-registry.js`
    - Files: `cmd/server/main.go`
  - [ ] 4.6 Handle component template collection
    - Get all registered components from component registry
    - Parse and transform each component to AST
    - Pass component list to registry generator
    - Files: `cmd/server/main.go`
  - [ ] 4.7 Verify all tests pass for Phase 4
    - Run `go test ./builder -v`
    - Check generated registry file structure
    - Validate ES module syntax (no syntax errors)
    - Test in browser: import registry and call template function
  - [ ] 4.8 **MANDATORY: Use go-backend agent for all Go implementation**

- [ ] 5. Phase 5: Integration & Testing (2-4h, Low Cognitive Load)
  - [ ] 5.1 Create integration test content in `content/pages/_index.json`
    - Add components array with hero2436 and services2437
    - Include props for each component (title, description, etc.)
    - Files: Update `content/pages/_index.json`
  - [ ] 5.2 Create or update homepage template
    - Add `export let components` in fence section
    - Add loop: `{for component in components} <Component:dynamic name={component.name} {...component.fields} /> {/for}`
    - Files: `layouts/content/pages.html` or similar
  - [ ] 5.3 Run end-to-end browser test
    - Start server: `go run cmd/server/main.go`
    - Navigate to http://localhost:3000/
    - Verify both components render with correct props
    - Check browser console for errors (should be none)
    - Verify registry loads and caches correctly
  - [ ] 5.4 Test build-time vs runtime resolution
    - Create test with static component name: `<Component:dynamic name="Hero2436" />`
    - Verify it still uses build-time resolution (no wrapper emitted)
    - Create test with runtime name in loop
    - Verify runtime wrapper is emitted
    - Files: Add test cases to `transformer/dynamic_component_by_name_test.go`
  - [ ] 5.5 Update documentation
    - Document runtime component resolution in CLAUDE.md
    - Add examples of {for component in components} pattern
    - Document ScopeAnalyzer decision logic
    - Explain when build-time vs runtime resolution is used
    - Files: `CLAUDE.md`
  - [ ] 5.6 Final regression testing
    - Run all tests: `go test ./... -v`
    - Test existing components still work (no breaking changes)
    - Test nested loops with components
    - Test component props injection with runtime resolution
  - [ ] 5.7 Mark spec as complete
    - Update status in spec.md to "Complete"
    - Document any deviations from original plan
    - Files: `.agent-os/specs/2025-10-15-runtime-component-resolution/spec.md`
  - [ ] 5.8 **MANDATORY: Use go-backend agent for all Go implementation**

## Notes

### Phase 1 Completion Summary (2025-10-15)

**Files Created:**
- `analyzer/scope.go` - ScopeAnalyzer implementation with all required methods
- `analyzer/scope_test.go` - Comprehensive test suite with 8 test functions

**Cognitive Load Validation:**
- `analyzer/scope.go`: Total Load = 10 (Pattern: Service Implementation) ✓
- Individual methods all < 10 cognitive load
- Test file: Load = 6-8 per test function ✓

**Test Results:**
- All 8 test functions PASS (38 sub-tests total)
- No regressions in existing packages (ast, parser)
- Pre-existing transformer test failures unrelated to this work

**Implementation Notes:**
- Used string-based expression parsing (not AST-based) for simplicity
- Helper functions: isStringLiteral, isAlpineStoreReference, hasOperators
- Defaults unknown variables to build-time for backwards compatibility
- Properly handles nested property access (component.name, item.field)

**Decision Logic Implemented:**
1. String literals ("ComponentName") → build-time ✓
2. Alpine stores ($store.*, $auth.*) → runtime ✓
3. Loop variables (component, item) → runtime ✓
4. Content props (from export let) → build-time ✓
5. Exported props → build-time ✓
6. Operators (+, -, *, etc.) → runtime (safety) ✓

**Ready for Phase 2:** Runtime Wrapper Emission

### Key Implementation Details

1. **Scope Analysis Decision Logic**:
   - Loop variables (component, item) are ALWAYS runtime-only
   - Content props (title, description from export let) are build-resolvable
   - Alpine stores ($store.*) are runtime-only
   - String literals ("ComponentName") are build-resolvable

2. **Runtime Wrapper Structure**:
   ```html
   <div class="dyn-comp-runtime"
        x-data="{compName: component.name, compProps: component.fields}"
        x-init="$renderDynamicComponent($el, compName, compProps)">
   </div>
   ```

3. **Registry Template Conversion**:
   - `{variable}` -> `${props.variable}`
   - Alpine directives preserved as-is
   - HTML structure maintained
   - Template literals escaped (backticks, ${})

4. **Error Handling Priorities**:
   - Missing component: Silent (HTML comment only)
   - Network errors: Retry with backoff
   - Template errors: Console log in dev mode

### Test Coverage Requirements

- Unit tests for each major component (ScopeAnalyzer, registry generator)
- Integration tests for transformer pipeline with scope tracking
- Browser tests for client-side runtime
- Regression tests for build-time resolution (ensure no breaking changes)

### Performance Targets

- Registry generation: < 500ms for 10-20 components
- Component rendering: < 50ms per component
- Registry size: < 100KB for typical site
- Scope analyzer overhead: < 1MB per page

### Dependencies

- No new external dependencies required
- Uses existing Alpine.js 3.x
- Uses existing AST and parser infrastructure
- Uses Go standard library only

### Out of Scope (Deferred)

- Server-side loop expansion (Phase 5 from original spec - deferred to future iteration)
- Deterministic signatures and hydration validation
- Component lazy loading and code splitting
- Development overlay and inspector UI
- Advanced error recovery and circuit breakers
