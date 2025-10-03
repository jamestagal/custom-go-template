# Spec Tasks

These are the tasks to be completed for the spec detailed in @.agent-os/specs/2025-10-02-function-expression-handling/spec.md

> Created: 2025-10-02
> Status: Ready for Implementation

## Tasks

- [ ] 1. Implement Enhanced isFunctionExpression() Detection
  - [ ] 1.1 Write tests for isFunctionExpression() with all JavaScript function patterns
    - [ ] 1.1.1 Test function declarations: `function name() {}`
    - [ ] 1.1.2 Test anonymous functions: `function() {}`
    - [ ] 1.1.3 Test arrow functions: `() => {}`, `(x) => {}`, `x => {}`
    - [ ] 1.1.4 Test async functions: `async function name() {}`
    - [ ] 1.1.5 Test getters/setters: `get name() {}`, `set name(v) {}`
    - [ ] 1.1.6 Test method shorthand: `name() {}`
    - [ ] 1.1.7 Test negative cases: strings, numbers, booleans
  - [ ] 1.2 Implement updated isFunctionExpression() in transformer/alpine.go
    - [ ] 1.2.1 Add function declaration detection (starts with "function")
    - [ ] 1.2.2 Add arrow function detection (contains "=>")
    - [ ] 1.2.3 Add async function detection (starts with "async")
    - [ ] 1.2.4 Add getter/setter detection (starts with "get " or "set ")
    - [ ] 1.2.5 Add method shorthand detection using regex pattern
  - [ ] 1.3 Verify all isFunctionExpression() tests pass

- [ ] 2. Implement isValidIdentifier() Helper Function
  - [ ] 2.1 Write tests for isValidIdentifier()
    - [ ] 2.1.1 Test valid identifiers: `count`, `_private`, `$jquery`, `camelCase`
    - [ ] 2.1.2 Test invalid identifiers: keywords (`function`, `const`, `let`, `true`, `false`)
    - [ ] 2.1.3 Test invalid identifiers: numbers as first char, special chars, empty string
  - [ ] 2.2 Implement isValidIdentifier() in transformer/alpine.go
    - [ ] 2.2.1 Define JavaScript keyword map
    - [ ] 2.2.2 Check first character is valid (letter, underscore, dollar sign)
    - [ ] 2.2.3 Check subsequent characters (letters, digits, underscore, dollar sign)
    - [ ] 2.2.4 Reject JavaScript keywords
  - [ ] 2.3 Verify all isValidIdentifier() tests pass

- [ ] 3. Refactor formatGoValueToJS() Function
  - [ ] 3.1 Write comprehensive tests for formatGoValueToJS()
    - [ ] 3.1.1 Test function expressions are returned without quotes
    - [ ] 3.1.2 Test complex JS objects/arrays are preserved as-is
    - [ ] 3.1.3 Test valid identifiers are returned without quotes
    - [ ] 3.1.4 Test regular strings are quoted and escaped
    - [ ] 3.1.5 Test primitives: booleans, integers, floats
    - [ ] 3.1.6 Test nil values return "null"
    - [ ] 3.1.7 Test arrays format elements recursively
    - [ ] 3.1.8 Test maps format key-value pairs
  - [ ] 3.2 Implement updated formatGoValueToJS() in transformer/alpine.go
    - [ ] 3.2.1 Add nil check returning "null"
    - [ ] 3.2.2 Handle string values with function expression check
    - [ ] 3.2.3 Handle complex JS object/array literal detection
    - [ ] 3.2.4 Handle valid identifier detection
    - [ ] 3.2.5 Handle regular strings with proper escaping
    - [ ] 3.2.6 Handle boolean, integer, float types
    - [ ] 3.2.7 Handle array recursion
    - [ ] 3.2.8 Handle map/object formatting
  - [ ] 3.3 Verify all formatGoValueToJS() tests pass

- [ ] 4. Update alpineDataFormatter() Implementation
  - [ ] 4.1 Write tests for alpineDataFormatter() with function values
    - [ ] 4.1.1 Test single function in data scope
    - [ ] 4.1.2 Test multiple functions with variables
    - [ ] 4.1.3 Test complex objects with methods
    - [ ] 4.1.4 Test nested structures
    - [ ] 4.1.5 Test Alpine.js magic properties ($refs, $el, etc.)
  - [ ] 4.2 Refactor alpineDataFormatter() in transformer/alpine.go
    - [ ] 4.2.1 Remove test-specific hardcoded values (lines 24-44)
    - [ ] 4.2.2 Keep ensureCriticalVariables() logic
    - [ ] 4.2.3 Sort keys for consistent output
    - [ ] 4.2.4 Replace json.Marshal() with formatGoValueToJS() for each value
    - [ ] 4.2.5 Build object literal string with proper key-value formatting
    - [ ] 4.2.6 Skip internal Alpine.js $ prefixed variables appropriately
    - [ ] 4.2.7 Add debug logging for generated x-data
  - [ ] 4.3 Verify alpineDataFormatter() tests pass

- [ ] 5. Integration Testing and Validation
  - [ ] 5.1 Run TestAlpineDataWrapper/Function_Expressions test
    - [ ] 5.1.1 Verify function is not quoted in x-data attribute
    - [ ] 5.1.2 Verify HTML entity escaping only affects quotes, not function body
    - [ ] 5.1.3 Confirm expected output matches: `{&quot;count&quot;:0,&quot;increment&quot;:function() { return count++ }}`
  - [ ] 5.2 Run all TestAlpineDataWrapper subtests
    - [ ] 5.2.1 Verify no regressions in existing tests
    - [ ] 5.2.2 Check basic Alpine data wrapper still works
    - [ ] 5.2.3 Check component props handling still works
    - [ ] 5.2.4 Check nested components still work
  - [ ] 5.3 Run TestIsComplexJSObject tests
    - [ ] 5.3.1 Verify complex object detection still works
    - [ ] 5.3.2 Ensure objects with methods are handled correctly
  - [ ] 5.4 Run full Alpine.js integration test suite
    - [ ] 5.4.1 Run: `go test ./tests/alpine -v`
    - [ ] 5.4.2 Verify all tests pass
    - [ ] 5.4.3 Check for any unexpected failures or warnings
  - [ ] 5.5 Manual browser testing (if development server available)
    - [ ] 5.5.1 Create test component with function in fence section
    - [ ] 5.5.2 Verify function executes in Alpine.js context
    - [ ] 5.5.3 Test click handlers and other event bindings
    - [ ] 5.5.4 Inspect x-data attribute in browser DevTools
  - [ ] 5.6 Final validation
    - [ ] 5.6.1 Run full test suite: `go test ./... -v`
    - [ ] 5.6.2 Verify no regressions across entire codebase
    - [ ] 5.6.3 Document any edge cases discovered
    - [ ] 5.6.4 Update technical spec if implementation deviates from plan
