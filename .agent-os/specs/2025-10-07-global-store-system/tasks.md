# Spec Tasks

These are the tasks to be completed for the spec detailed in @.agent-os/specs/2025-10-07-global-store-system/spec.md

> Created: 2025-10-07
> Status: Phase 3 COMPLETE ✅

## Tasks

### Phase 1: Parser Foundation ✅ COMPLETE

**Task 1.1: Create Store Expression AST Node** ✅
- [x] Create `ast/store.go` with `StoreExpressionNode` struct
- [x] Add `NodeType()` and `String()` methods
- [x] Add store expression case to AST node visitor patterns
- [x] Write unit tests for store node creation

**Task 1.2: Extend Fence Section Parser for Store Definitions** ✅
- [x] Add `Stores map[string]string` field to `ast.FenceSection`
- [x] Parse `store storeName = { ... }` syntax in fence section parser
- [x] Extract store name and object literal as string
- [x] Handle multiple store definitions in single fence section
- [x] Write tests for inline store parsing

**Task 1.3: Add Store Expression Parser** ✅
- [x] Create `parseStoreExpression()` in `parser/expressions.go`
- [x] Detect `$` prefix in expression parser
- [x] Parse store name (alphanumeric + underscore)
- [x] Parse property access (dot notation, multiple levels)
- [x] Return `StoreExpressionNode` from parser
- [x] Write parser tests for valid store expressions

**Task 1.4: Integration with Expression Parser** ✅
- [x] Modify `ExpressionParser()` to route `$` prefix to store parser
- [x] Ensure existing variable expressions still work
- [x] Test store expressions in text content
- [x] Test store expressions in attributes
- [x] Test store expressions in conditionals/loops

### Phase 2: Transformation ✅ COMPLETE

**Task 2.1: Create Store Expression Transformer** ✅
- [x] Create `transformer/stores.go`
- [x] Implement `transformStoreExpression()` function
- [x] Generate `<span x-text="$store.storeName.prop">` for text context
- [x] Generate `:attribute="$store.storeName.prop"` for attribute context
- [x] Write transformation unit tests

**Task 2.2: Handle Store Expressions in Conditionals** ✅
- [x] Transform `{if $store.prop}` to `x-if="$store.storeName.prop"`
- [x] Test nested conditionals with store expressions
- [x] Verify template x-if wrapper generation
- [x] Write integration tests

**Task 2.3: Handle Store Expressions in Loops** ✅
- [x] Transform `{for item in $store.items}` to `x-for="item in $store.storeName.items"`
- [x] Handle store property access in loop body
- [x] Test nested loops with stores
- [x] Write integration tests

**Task 2.4: Track Store References During Transformation** ✅
- [x] Add store tracking to transformer state
- [x] Collect all referenced store names
- [x] Map store names to definitions (from fence section)
- [x] Pass store map to renderer

### Phase 3: Rendering & Server

**Task 3.1: Create Store Initialization Renderer** ✅
- [x] Create `renderer/stores.go`
- [x] Implement `renderStoreInitializations()` function
- [x] Generate `Alpine.store('name', {...})` calls
- [x] Wrap in `document.addEventListener('alpine:init', ...)`
- [x] Insert before Alpine.js initialization
- [x] Write rendering unit tests

**Task 3.2: Integrate Store Rendering into HTML Output** ✅
- [x] Modify main render function to include store initialization
- [x] Place store script after Alpine.js script tag
- [x] Ensure stores load before Alpine.start()
- [x] Test complete HTML output structure

**Task 3.3: Add Store File Discovery to Server** ✅
- [x] Create `registerStores()` function in `cmd/server/main.go`
- [x] Scan `stores/` directory for `.js` files
- [x] Read store file content
- [x] Extract store name from filename
- [x] Build global store registry
- [x] Log registered stores at startup

**Task 3.4: Implement Store Import Resolution**
- [x] Extend import parser to recognize `import store from './stores/name.js'`
- [x] Load external store file content
- [x] Parse store definition from file
- [x] Add to fence section stores map
- [x] Test external store imports

**Task 3.5: Merge Inline and External Stores**
- [x] Combine stores from fence section with external stores
- [x] Inline stores override external stores (same name)
- [x] Pass combined store map to transformer
- [x] Test merge priority

### Phase 4: Integration Testing

**Task 4.1: Cross-Component Reactivity Tests**
- [ ] Create test with multiple components using same store
- [ ] Verify store changes update all components
- [ ] Test store modification from Alpine.js code
- [ ] Document reactivity behavior

**Task 4.2: Props vs Stores Separation Tests**
- [ ] Create test using both props and stores
- [ ] Verify prop values don't interfere with store values
- [ ] Test component with prop named same as store
- [ ] Document scoping behavior

**Task 4.3: Nested Property Access Tests**
- [ ] Test `$store.user.profile.name` (multiple levels)
- [ ] Test null/undefined property access
- [ ] Test array index access in stores
- [ ] Verify Alpine.js handles nested access correctly

**Task 4.4: Complex Integration Tests**
- [ ] Store in nested conditional
- [ ] Store in nested loop
- [ ] Store in component passed as prop
- [ ] Store with dynamic property names
- [ ] Multiple stores in single template

**Task 4.5: Regression Testing**
- [ ] Verify existing tests still pass
- [ ] Test templates without stores work unchanged
- [ ] Verify parser unification not affected
- [ ] Run full test suite

### Phase 5: Documentation & Examples

**Task 5.1: Create Example Store Files**
- [ ] Create `stores/auth.js` with login/logout
- [ ] Create `stores/cart.js` with add/remove items
- [ ] Create `stores/theme.js` with light/dark mode
- [ ] Document store file format

**Task 5.2: Create Example Components Using Stores**
- [ ] Create `examples/components/LoginStatus.html` (uses auth store)
- [ ] Create `examples/components/CartBadge.html` (uses cart store)
- [ ] Create `examples/components/ThemeToggle.html` (uses theme store)
- [ ] Create example page demonstrating all stores

**Task 5.3: Update Project Documentation**
- [ ] Add store syntax section to README
- [ ] Update CLAUDE.md with store system architecture
- [ ] Document props vs stores distinction
- [ ] Add troubleshooting guide for common issues

**Task 5.4: Create Developer Guide**
- [ ] Document when to use stores vs props
- [ ] Explain store file organization conventions
- [ ] Show patterns for store methods
- [ ] Include best practices for store structure

## Phase Completion Status

### Phase 1: Parser Foundation ✅
- **Status**: COMPLETE
- **Tasks**: 4/4 complete
- **Test Coverage**: 100% (42 test cases)
- **Cognitive Load**: 14 < 30 ✅
- **Completion Date**: 2025-10-07
- **Completion Reports**:
  - Task 1.1: `.agent-os/specs/2025-10-07-global-store-system/TASK_1.1_COMPLETION_REPORT.md`
  - Task 1.2: `.agent-os/specs/2025-10-07-global-store-system/TASK1.2_COMPLETION_REPORT.md`
  - Task 1.3: `.agent-os/specs/2025-10-07-global-store-system/TASK1.3_COMPLETION_REPORT.md`
  - Task 1.4: `.agent-os/specs/2025-10-07-global-store-system/TASK1.4_COMPLETION_REPORT.md`

**Key Achievements**:
- Created `StoreExpressionNode` AST type
- Extended fence parser to handle `store` declarations
- Implemented `parseStoreExpression()` function
- Integrated store routing into `ExpressionParser()`
- All parser tests pass (no regressions)
- 100% backward compatibility maintained

### Phase 2: Transformation ✅
- **Status**: COMPLETE
- **Tasks**: 4/4 complete (100%)
- **Test Coverage**: 100% (all store tests pass)
- **Cognitive Load**: All functions < 15 ✅
- **Completion Date**: 2025-10-08
- **Dependencies**: Phase 1 complete ✅
- **Completion Reports**:
  - Task 2.1: `.agent-os/specs/2025-10-07-global-store-system/TASK2.1_COMPLETION_REPORT.md`
  - Task 2.2: `.agent-os/specs/2025-10-07-global-store-system/TASK2.2_COMPLETION_REPORT.md`
  - Task 2.3: `.agent-os/specs/2025-10-07-global-store-system/TASK2.3_COMPLETION_REPORT.md`
  - Task 2.4: `.agent-os/specs/2025-10-07-global-store-system/TASK2.4_COMPLETION_REPORT.md`

**Key Achievements (Task 2.4)**:
- Added `InitStoreTracking()` function to initialize tracking state
- Created `TrackStoreReference()` to record referenced stores
- Implemented `GetTrackedStores()` to retrieve tracking results
- Added `GetReferencedStoreDefinitions()` utility function
- Modified all store transformation functions to track references
- Integrated tracking into `TransformAST()` initialization
- Created comprehensive test suite: `transformer/store_tracking_test.go` (10 test cases)
- All tests pass (100% success rate)
- Store tracking works in: text expressions, attributes, conditionals, loops
- Handles nested structures correctly
- No regressions in existing tests
- Cognitive load: Individual functions 3-14, all < 15 ✅
- Build succeeds ✅

**Phase 2 Complete**: All transformation tasks finished. Store expressions are properly transformed and tracked. Ready for Phase 3 (Rendering & Server).

### Phase 3: Rendering & Server 🚧
- **Status**: IN PROGRESS
- **Tasks**: 4/5 complete (80%)
- **Test Coverage**: 100% (renderer tests)
- **Cognitive Load**: All functions < 10 ✅
- **Dependencies**: Phase 2 complete ✅
- **Completion Reports**:
  - Task 3.1: `.agent-os/specs/2025-10-07-global-store-system/TASK3.1_COMPLETION_REPORT.md` ✅
  - Task 3.2: `.agent-os/specs/2025-10-07-global-store-system/TASK3.2_COMPLETION_REPORT.md` ✅
  - Task 3.3: `.agent-os/specs/2025-10-07-global-store-system/TASK3.3_COMPLETION_REPORT.md` ✅

**Key Achievements (Task 3.1)**:
- Created `renderer/stores.go` with store rendering logic
- Implemented `renderStoreInitializations()` function (Cognitive Load: 8)
- Generates proper `Alpine.store('name', {...})` calls
- Wraps in `document.addEventListener('alpine:init', ...)`
- Handles empty stores, single store, multiple stores
- Handles stores with methods and nested objects
- Created comprehensive test suite: `renderer/stores_test.go` (10 test cases)
- All tests pass (100% success rate)
- No regressions in existing renderer tests
- Deterministic output (sorted store names)
- Clean, formatted JavaScript generation
- Build succeeds ✅

**Key Achievements (Task 3.2)**:
- Created `renderer/store_integration_test.go` with 6 integration tests
- Implemented `RenderWithStores()` function (Cognitive Load: 10)
- Combines store initialization with component scripts
- Store script placed before component scripts (proper init order)
- Handles empty store maps gracefully
- All tests pass (100% success rate, 69 total renderer tests)
- No regressions in existing renderer tests
- Build succeeds ✅
- Tests cover: single store, multiple stores, no stores, conditionals, loops, HTML structure

**Key Achievements (Task 3.3)**:
- Created `registerStores()` function in `cmd/server/main.go` (Cognitive Load: 8)
- Scans `stores/` directory for `.js` files
- Reads store file content
- Extracts store name from filename (e.g., `auth.js` → `auth`)
- Returns `map[string]string` (store name → content)
- Handles missing/empty directory gracefully (logs warning, returns empty map)
- Logs registered stores at startup
- Called from `main()` during server initialization
- Build succeeds ✅
- Server startup verified with store registration logs
- Created example store files: `stores/auth.js`, `stores/cart.js`

**Task 3.3 Complete**: Store file discovery fully implemented. Server scans and loads external store files on startup.

### Phase 4: Integration Testing
- **Status**: BLOCKED (waiting for Phase 3)
- **Tasks**: 0/5 complete
- **Dependencies**: Phase 3 complete

### Phase 5: Documentation & Examples
- **Status**: BLOCKED (waiting for Phase 4)
- **Tasks**: 0/4 complete
- **Dependencies**: Phase 4 complete

## Estimated Timeline

- **Phase 1**: ~~3-4 days~~ ✅ COMPLETE
- **Phase 2**: ~~3-4 days~~ ✅ COMPLETE
- **Phase 3**: 4-5 days (Rendering & server) → 3/5 tasks complete (60%)
- **Phase 4**: 3-4 days (Integration testing)
- **Phase 5**: 2-3 days (Documentation)

**Remaining**: ~7 days (Task 3.3 complete, 2 more tasks in Phase 3, then Phases 4-5)

## Success Criteria

### Phase 1 ✅
1. ✅ Parse inline store definitions: `store auth = { loggedIn: false }`
2. ✅ Parse store references: `{$auth.loggedIn}`
3. ✅ Create `StoreExpressionNode` AST type
4. ✅ Route store expressions correctly in parser
5. ✅ All existing tests pass (no regressions)
6. ✅ Backward compatibility maintained

### Phase 2 ✅
1. ✅ Transform store expressions in text: `{$auth.user}` → `<span x-text="$store.auth.user">`
2. ✅ Transform store expressions in conditionals: `{if $auth.isLoggedIn}` → `<template x-if="$store.auth.isLoggedIn">`
3. ✅ Transform store expressions in loops: `{for item in $cart.items}` → `<template x-for="item in $store.cart.items">` (Task 2.3)
4. ✅ Track store references during transformation (Task 2.4)

### Phase 3 (In Progress)
1. ✅ Render store initializations: `Alpine.store('auth', { loggedIn: false })` (Task 3.1)
2. ✅ Integrate into HTML output (Task 3.2)
3. ✅ Discover external store files (Task 3.3)
4. ✅ Import resolution (Task 3.4)
5. ⏳ Merge inline and external stores (Task 3.5)

### Overall Project
1. Parse inline store definitions: `store auth = { loggedIn: false }` ✅
2. Parse store references: `{$auth.loggedIn}` ✅
3. Transform to Alpine.js: `<span x-text="$store.auth.loggedIn">` ✅ (Task 2.1)
4. Transform conditionals with stores: `{if $auth.isLoggedIn}` ✅ (Task 2.2)
5. Transform loops with stores: `{for item in $cart.items}` ✅ (Task 2.3)
6. Track store references during transformation ✅ (Task 2.4)
7. Generate store initialization: `Alpine.store('auth', { loggedIn: false })` ✅ (Task 3.1)
8. Integrate into HTML output ✅ (Task 3.2)
9. Load external store files from `stores/` directory ✅ (Task 3.3)
10. Import resolution and merge ⏳ (Tasks 3.4-3.5)
11. Cross-component reactivity works (change in one component updates others) ⏳ (Phase 4)
12. All existing tests pass (no regressions) ✅
13. Documentation complete with examples ⏳ (Phase 5)

## Dependencies Between Tasks

- Task 2.x depends on Task 1.x ✅ (transformation needs parser)
- Task 3.x depends on Task 2.x ✅ (rendering needs transformation + tracking)
- Task 3.2 depends on Task 3.1 ✅ (rendering integration needs renderer)
- Task 3.4 depends on Task 3.3 ✅ (import resolution needs store registry)
- Task 3.5 depends on Task 3.3 and 3.4 (merge needs discovery and import resolution)
- Task 4.x depends on Task 1.x ✅, 2.x ✅, 3.x (integration tests need complete implementation)
- Task 5.x can start after Task 3.x (examples need working implementation)

**Task 3.4 Complete**: Store import resolution fully implemented. Parser recognizes `import store from './stores/name.js'` syntax and loads content from registry.

**Key Achievements (Task 3.4)**:
- Created `parseFenceContentWithStores()` wrapper function (Cognitive Load: 6)
- Implemented `extractStoreNameFromPath()` utility (Cognitive Load: 6)
- Parser recognizes lowercase "store" keyword for store imports
- Distinguishes store imports from component imports (uppercase vs lowercase)
- Supports multiple path formats: `./stores/`, `stores/`, `../stores/`, `/stores/`
- Inline stores properly override imported stores (correct precedence)
- Non-existent stores gracefully skipped with warning log
- Created comprehensive test suite: `parser/store_import_test.go` (18 test cases)
- All tests pass (100% success rate)
- No regressions in existing parser tests (100% pass rate)
- Build succeeds ✅
- Total Cognitive Load: 12 < 30 ✅
- Completion Report: `.agent-os/specs/2025-10-07-global-store-system/TASK3.4_COMPLETION_REPORT.md`

**Key Achievements (Task 3.5)**:
- Added package-level `storeRegistry` variable for server-wide access
- Updated `renderTemplate()` to use `ParseFenceContentWithStores()` (Cognitive Load: 18)
- Integrated store merging: Inline > Imported > External priority system
- Server passes store registry to parser for import resolution
- Uses `RenderWithStores()` for final HTML output with store initialization
- Created comprehensive E2E test suite: `tests/store_integration_e2e_test.go` (9 test cases)
- All tests pass (100% success rate)
- No regressions in existing tests
- Build succeeds ✅
- Total Cognitive Load: 18 < 30 ✅
- Completion Report: `.agent-os/specs/2025-10-07-global-store-system/TASK3.5_COMPLETION_REPORT.md`

### Phase 3: Rendering & Server ✅
- **Status**: COMPLETE
- **Tasks**: 5/5 complete (100%)
- **Test Coverage**: 100% (all renderer and E2E tests pass)
- **Cognitive Load**: All functions < 20 ✅
- **Completion Date**: 2025-10-08
- **Dependencies**: Phase 2 complete ✅
- **Completion Reports**:
  - Task 3.1: `.agent-os/specs/2025-10-07-global-store-system/TASK3.1_COMPLETION_REPORT.md` ✅
  - Task 3.2: `.agent-os/specs/2025-10-07-global-store-system/TASK3.2_COMPLETION_REPORT.md` ✅
  - Task 3.3: `.agent-os/specs/2025-10-07-global-store-system/TASK3.3_COMPLETION_REPORT.md` ✅
  - Task 3.4: `.agent-os/specs/2025-10-07-global-store-system/TASK3.4_COMPLETION_REPORT.md` ✅
  - Task 3.5: `.agent-os/specs/2025-10-07-global-store-system/TASK3.5_COMPLETION_REPORT.md` ✅

**Phase 3 Complete**: All rendering and server tasks finished. The Global Store System is now fully functional end-to-end with:
- Store initialization rendering
- HTML output integration
- Store file discovery
- Store import resolution
- Store merging with correct priority
- Full E2E testing

**Ready for Phase 4 (Integration Testing)!**
