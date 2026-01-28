# HTML Wrapper Integration - Tasks

**Spec:** HTML Wrapper Integration
**Goal:** Integrate `layouts/global/html.html` as the root wrapper for all pages
**Status:** Ready for Implementation
**MANDATORY: Use go-backend agent for all Go implementation**

## Tasks

- [x] 1. Implement Component:dynamic Resolution for Simple Props
  - [x] 1.1 Write tests for Component:dynamic with static prop values
  - [x] 1.2 Update transformer to handle `<Component:dynamic name={layout} />` where layout is a prop (already working!)
  - [x] 1.3 Implement prop substitution in transformer (e.g., `{layout}` → `"_index"`) (already working!)
  - [x] 1.4 Add component loading logic for resolved layout names (already working!)
  - [x] 1.5 Implement layout injection at Component:dynamic location in renderer (already working!)
  - [x] 1.6 Handle prop spreading `{...content.fields}` to injected layout (already working!)
  - [x] 1.7 Add error handling for missing layouts (already working!)
  - [x] 1.8 Verify all tests pass

- [x] 2. Create renderWithWrapper Function in Server
  - [x] 2.1 Write integration tests for renderWithWrapper
  - [x] 2.2 Implement `renderWithWrapper(layoutName, w, r)` function
  - [x] 2.3 Add content loading via loadContentWithCache
  - [x] 2.4 Build props map with layout, content, allContent, allLayouts
  - [x] 2.5 Call renderTemplate with html.html wrapper and props
  - [x] 2.6 Add logging for wrapper rendering pipeline
  - [x] 2.7 Test with single route (home page) - ✅ Nav + Footer rendering!
  - [x] 2.8 Verify all tests pass - ✅ Fixed content layout registration

- [ ] 3. Update Route Handlers to Use Wrapper
  - [ ] 3.1 Write tests for all updated routes
  - [x] 3.2 Update "/" route to use renderWithWrapper("_index", w, r) - ✅ Complete
  - [ ] 3.3 Update "/comprehensive" route to use wrapper
  - [ ] 3.4 Update "/store-demo" route to use wrapper
  - [ ] 3.5 Update all remaining content routes
  - [ ] 3.6 Add route-to-layout mapping logic
  - [ ] 3.7 Test each route individually
  - [ ] 3.8 Verify all tests pass

- [ ] 4. Simplify Content Layouts (Remove HTML Boilerplate)
  - [ ] 4.1 Write tests for simplified layouts
  - [ ] 4.2 Backup current _index.html before modification
  - [ ] 4.3 Remove `<!DOCTYPE html>`, `<html>`, `<head>`, `<body>` from _index.html
  - [ ] 4.4 Keep only content-specific components in _index.html
  - [ ] 4.5 Simplify comprehensive.html layout
  - [ ] 4.6 Simplify store-demo.html layout
  - [ ] 4.7 Update any other content layouts
  - [ ] 4.8 Verify all tests pass

- [ ] 5. Integration Testing and Documentation
  - [ ] 5.1 Write end-to-end tests for wrapper system
  - [ ] 5.2 Test Nav component renders on all pages
  - [ ] 5.3 Test Footer component renders on all pages
  - [ ] 5.4 Test Head component with correct metadata
  - [ ] 5.5 Test props flow from server to wrapper to layout
  - [ ] 5.6 Test Component:dynamic injection works correctly
  - [ ] 5.7 Update CLAUDE.md with wrapper pattern documentation
  - [ ] 5.8 Verify all tests pass and create completion report

## Task Execution Notes

### Task 1: Component:dynamic Resolution
- **Focus:** Transformer and renderer changes only
- **Key Files:** `transformer/dynamic_component_by_name.go`, `renderer/dynamic_component_by_name.go`
- **Success:** `<Component:dynamic name={layout} />` resolves when `layout` is a prop
**MANDATORY: Use go-backend agent for all Go implementation**

### Task 2: renderWithWrapper Function
- **Focus:** Server-side wrapper orchestration
- **Key Files:** `cmd/server/main.go`
- **Success:** Single route (/) renders with Nav + content + Footer
**MANDATORY: Use go-backend agent for all Go implementation**

### Task 3: Route Handler Updates
- **Focus:** Update all routes to use wrapper
- **Key Files:** `cmd/server/main.go`
- **Success:** All routes render with global components
**MANDATORY: Use go-backend agent for all Go implementation**

### Task 4: Content Layout Simplification
- **Focus:** Remove HTML boilerplate from content layouts
- **Key Files:** `layouts/content/_index.html`, `layouts/content/comprehensive.html`, etc.
- **Success:** Content layouts contain only page-specific components
**MANDATORY: Use go-backend agent for all Go implementation**

### Task 5: Integration Testing
- **Focus:** End-to-end validation and documentation
- **Key Files:** `tests/wrapper_integration_test.go`, `CLAUDE.md`
- **Success:** All pages work, documentation complete
**MANDATORY: Use go-backend agent for all Go implementation**

## Dependencies

- ✅ `layouts/global/html.html` exists with component imports
- ✅ `layouts/global/nav.html` exists
- ✅ `layouts/global/head.html` exists
- ✅ `layouts/global/footer.html` exists
- ✅ Component:dynamic transformer supports simple prop substitution (Task 1 complete)
- ✅ Renderer supports layout injection at Component:dynamic location (Task 1 complete)

## Success Criteria

- ✅ Nav component appears on all pages
- ✅ Footer component appears on all pages
- ✅ Head component with consistent metadata
- ✅ Content layouts contain only page-specific content
- ✅ Component:dynamic correctly injects layouts
- ✅ All existing pages continue to work
- ✅ No regressions in existing functionality
- ✅ Documentation updated

## Estimated Complexity

**Overall:** Medium
- Component:dynamic prop resolution: Low (prop substitution is simpler than loop resolution)
- Server wrapper function: Low (orchestration only)
- Route updates: Low (mechanical changes)
- Layout simplification: Low (remove code)
- Integration testing: Medium (comprehensive validation needed)

**Total Estimated Time:** 3-4 hours for full implementation and testing
