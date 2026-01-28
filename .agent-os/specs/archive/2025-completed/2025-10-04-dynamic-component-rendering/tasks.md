# Tasks: Dynamic Component Rendering Fix

**Status**: Not Started
**Created**: 2025-10-04
**Total Estimated Load**: 48

---

## Problem Statement

Dynamic components using `<=` syntax are currently creating placeholder divs instead of rendering actual component content. The root cause is a component registry mismatch - components are registered by name ("UserProfile") but looked up by path ("./components/UserProfile.html").

**Current Behavior**:
```html
<!-- What we get now (WRONG) -->
<div x-component="./components/UserProfile.html" data-prop-name="Jim"></div>
```

**Expected Behavior**:
```html
<!-- What we should get (CORRECT) -->
<div x-data="{...}">
  <!-- Actual UserProfile component content inline -->
</div>
```

---

## Task 1: Component Registry Normalization

**Cognitive Load**: 12
**Estimated Time**: 1 hour
**Status**: Not Started

### Subtasks

- [ ] 1.1 Write tests for component registration with multiple path formats
- [ ] 1.2 Create `normalizeComponentPath()` helper function to generate lookup key variants
- [ ] 1.3 Update component registration in `cmd/server/main.go` to register with multiple keys
- [ ] 1.4 Add tests for component lookup by name, full path, and relative path
- [ ] 1.5 Verify all component registry tests pass

**Files to Modify**:
- `cmd/server/main.go` (lines 208-244)
- `transformer/components.go` (add helper function)

**Success Criteria**:
- Component registered as "UserProfile" can be found by "./components/UserProfile.html"
- Component registered as "./components/UserProfile.html" can be found by "UserProfile"
- Backward compatibility maintained

---

## Task 2: Path Variable Resolution

**Cognitive Load**: 15
**Estimated Time**: 1.5 hours
**Status**: Not Started

### Subtasks

- [ ] 2.1 Write tests for `resolvePathVariables()` with various interpolation patterns
- [ ] 2.2 Implement `resolvePathVariables()` function in `transformer/components.go`
- [ ] 2.3 Add test cases for missing variables (error handling)
- [ ] 2.4 Add test cases for nested variable references
- [ ] 2.5 Verify all path resolution tests pass

**Files to Modify**:
- `transformer/components.go` (enhance `resolveDynamicPath()`)
- Create `transformer/path_resolution_test.go`

**Success Criteria**:
- `{comp}` resolves to "UserProfile" when `comp = "UserProfile"`
- `{path}` resolves to "./components/UserProfile.html" when `path = "./components/UserProfile.html"`
- Missing variables return helpful error messages

---

## Task 3: Component Inlining Implementation

**Cognitive Load**: 18
**Estimated Time**: 2 hours
**Status**: Not Started

### Subtasks

- [ ] 3.1 Write integration tests for dynamic component rendering in `tests/components/dynamic_components_test.go`
- [ ] 3.2 Implement AST cloning helper (`cloneTemplate()`) if not already available
- [ ] 3.3 Implement scope merging helper (`mergeMaps()`)
- [ ] 3.4 Rewrite `transformDynamicComponent()` to inline components instead of creating placeholders
- [ ] 3.5 Add proper error handling for missing components
- [ ] 3.6 Implement x-data wrapper logic for components with props
- [ ] 3.7 Test with home.html examples (static paths, variable interpolation, backticks)
- [ ] 3.8 Verify all component inlining tests pass

**Files to Modify**:
- `transformer/components.go` (lines 533-586, rewrite `transformDynamicComponent()`)
- Create `tests/components/dynamic_components_test.go`

**Success Criteria**:
- `<='./components/UserProfile.html' />` renders actual UserProfile HTML
- Props are properly passed to inlined components
- No placeholder divs with `x-component` attribute
- Component fence variables are properly scoped

---

## Task 4: End-to-End Validation

**Cognitive Load**: 8
**Estimated Time**: 1 hour
**Status**: Not Started

### Subtasks

- [ ] 4.1 Run full test suite (`go test ./... -v`)
- [ ] 4.2 Build and start development server
- [ ] 4.3 Verify "Dynamic Component Examples" section renders in browser at http://localhost:3333
- [ ] 4.4 Verify all three dynamic component examples show actual content (not placeholder divs)
- [ ] 4.5 Test component props are correctly passed and displayed
- [ ] 4.6 Verify no regression in existing static component rendering
- [ ] 4.7 Create completion summary document
- [ ] 4.8 Update roadmap and mark spec as complete

**Files to Create**:
- `.agent-os/specs/2025-10-04-dynamic-component-rendering/COMPLETION_SUMMARY.md`

**Success Criteria**:
- All tests pass
- Browser shows actual UserProfile component content
- Props are visible in rendered components
- No console errors

---

## Implementation Order

Execute tasks in sequence:

1. **Task 1** → Component Registry Normalization (1 hr)
2. **Task 2** → Path Variable Resolution (1.5 hrs)
3. **Task 3** → Component Inlining Implementation (2 hrs)
4. **Task 4** → End-to-End Validation (1 hr)

**Total Time**: ~5.5 hours

---

## Success Criteria

All three dynamic component examples in `examples/pages/home.html` (lines 52-58) must render actual component content:

1. ✓ Static path: `<={``./components/UserProfile.html``} />`
2. ✓ Variable interpolation: `<='{path}' />`
3. ✓ Variable in path: `<={``./components/{comp}.html``} />`

**Expected Browser Output**:
- See actual UserProfile card with user information
- Props (name, age) displayed correctly
- No placeholder divs
- Component styles applied

---

## Notes

- Follow TDD: write tests before implementation
- Maintain cognitive load < 30 per function
- Preserve existing component transformation logic
- Use Go 1.21+ error wrapping with `fmt.Errorf`
- Add detailed logging for debugging
