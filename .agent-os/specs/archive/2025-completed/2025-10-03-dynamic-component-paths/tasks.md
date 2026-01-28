# Tasks: Spec 4 - Dynamic Component Paths

**Status**: Not Started
**Created**: 2025-10-03
**Total Estimated Load**: 60

---

## Task 1: Add DynamicComponentNode to AST ✅

**Cognitive Load**: 5
**Estimated Time**: 30 minutes
**Status**: Not Started

### Subtasks

1. Add `DynamicComponentNode` struct to `ast/ast.go`
   - [ ] Define struct with PathExpression, Props, SelfClosing fields
   - [ ] Add String() method for debugging

2. Update transformer type switches
   - [ ] Add case in `transformer/transformer.go` transformNode()
   - [ ] Add case in any visitor patterns

3. Test AST node creation
   - [ ] Create test instances
   - [ ] Verify String() output

**Files to Modify**:
- `ast/ast.go`
- `transformer/transformer.go`

---

## Task 2: Implement DynamicComponentParser ✅

**Cognitive Load**: 15
**Estimated Time**: 1.5 hours
**Status**: Not Started

### Subtasks

1. Create `DynamicComponentParser()` in `parser/components.go`
   - [ ] Match `<=` prefix
   - [ ] Extract quoted path (single or double quotes)
   - [ ] Parse props after path
   - [ ] Check for self-closing `/>`
   - [ ] Return DynamicComponentNode

2. Integrate with TemplateParser
   - [ ] Add DynamicComponentParser BEFORE ComponentParser in `parser/parser.go`
   - [ ] Ensure <= components are parsed before < components

3. Write parser tests in `parser/components_test.go`
   - [ ] Test basic syntax: `<='./path.html' />`
   - [ ] Test with variables: `<='./views/{comp}.html' />`
   - [ ] Test with props: `<='path' prop={value} />`
   - [ ] Test double quotes: `<="path" />`
   - [ ] Test multiple variables: `<='./views/{dir}/{comp}.html' />`
   - [ ] Test error cases: missing quotes, unclosed quotes
   - [ ] Verify all tests pass

**Files to Modify**:
- `parser/components.go` (add DynamicComponentParser)
- `parser/parser.go` (integrate parser)
- `parser/components_test.go` (add tests)

---

## Task 3: Implement transformDynamicComponent ✅

**Cognitive Load**: 20
**Estimated Time**: 2 hours
**Status**: Not Started

### Subtasks

1. Create helper functions in `transformer/components.go`
   - [ ] Implement `extractVariablesFromPath()` - extract {variable} from path strings
   - [ ] Implement `resolveDynamicPath()` - resolve path with variable substitution
   - [ ] Implement `transformComponentWithTemplate()` - refactor existing component logic

2. Implement `transformDynamicComponent()` in `transformer/components.go`
   - [ ] Extract variables from path expression
   - [ ] Resolve path (compile-time if possible)
   - [ ] Look up component template from registry
   - [ ] Handle missing component gracefully
   - [ ] Transform like regular component using transformComponentWithTemplate()

3. Add error handling
   - [ ] Missing component: return error placeholder
   - [ ] Invalid path: log warning, return empty
   - [ ] Clear error messages

4. Write transformer tests in `transformer/components_test.go`
   - [ ] Test static path resolution
   - [ ] Test dynamic path with variable
   - [ ] Test variable not resolved (runtime)
   - [ ] Test missing component error handling
   - [ ] Test prop passing to dynamic components
   - [ ] Verify all tests pass

**Files to Modify**:
- `transformer/components.go` (add transformation logic)
- `transformer/transformer.go` (add case for DynamicComponentNode)
- `transformer/components_test.go` (add tests)

---

## Task 4: Integration and Testing ✅

**Cognitive Load**: 12
**Estimated Time**: 1.5 hours
**Status**: Not Started

### Subtasks

1. Create integration test file `tests/alpine/dynamic_components_test.go`
   - [ ] Test static dynamic component
   - [ ] Test dynamic with variable
   - [ ] Test dynamic in loop
   - [ ] Test nested dynamic components
   - [ ] Test error cases (missing component)

2. Register test components in test setup
   - [ ] Create Card.html test component
   - [ ] Create List.html test component
   - [ ] Register components before tests

3. Verify Alpine.js output
   - [ ] Check x-data scope contains path variables
   - [ ] Check component props passed correctly
   - [ ] Check transformed HTML structure

4. Run full test suite
   - [ ] Run `go test ./... -v`
   - [ ] Verify no regressions
   - [ ] Verify all dynamic component tests pass

**Files to Create**:
- `tests/alpine/dynamic_components_test.go`

**Files to Modify**:
- Test setup in integration tests

---

## Task 5: Documentation and Cleanup ✅

**Cognitive Load**: 8
**Estimated Time**: 1 hour
**Status**: Not Started

### Subtasks

1. Update CLAUDE.md
   - [ ] Add dynamic component syntax documentation
   - [ ] Add examples: `<='./views/{comp}.html' />`
   - [ ] Explain use cases (runtime component selection)
   - [ ] Add to "Template Syntax" section

2. Create example templates
   - [ ] Add `examples/dynamic_components.html` with examples
   - [ ] Show static dynamic components
   - [ ] Show variable interpolation
   - [ ] Show dynamic components in loops

3. Create completion summary
   - [ ] Document what was implemented
   - [ ] List test results
   - [ ] Note any limitations or future work
   - [ ] Save as `COMPLETION_SUMMARY.md` in spec folder

4. Final cleanup
   - [ ] Remove debug logging if excessive
   - [ ] Run `go fmt ./...`
   - [ ] Run `go vet ./...`
   - [ ] Final test run: `go test ./... -v`

**Files to Modify**:
- `CLAUDE.md`

**Files to Create**:
- `examples/dynamic_components.html`
- `.agent-os/specs/2025-10-03-dynamic-component-paths/COMPLETION_SUMMARY.md`

---

## Implementation Order

Execute tasks in sequence:

1. **Task 1** → AST foundation (30 min)
2. **Task 2** → Parser implementation (1.5 hrs)
3. **Task 3** → Transformer implementation (2 hrs)
4. **Task 4** → Integration testing (1.5 hrs)
5. **Task 5** → Documentation (1 hr)

**Total Time**: ~6.5 hours

---

## Success Criteria

- [ ] All parser tests pass
- [ ] All transformer tests pass
- [ ] All integration tests pass
- [ ] No regressions in existing tests
- [ ] Dynamic component syntax works: `<='path' />`
- [ ] Variable interpolation works: `<='./views/{comp}.html' />`
- [ ] Props pass correctly to dynamic components
- [ ] Clear error messages for missing components
- [ ] Documentation updated
- [ ] Examples created

---

## Notes

- Follow TDD: write tests before implementation
- Maintain cognitive load < 30 per function
- Use parser combinator patterns
- Follow existing code style and conventions
- Reference Spec 1 (components) for similar patterns
