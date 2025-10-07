# Spec Tasks

These are the tasks to be completed for the spec detailed in @.agent-os/specs/2025-10-07-global-store-system/spec.md

> Created: 2025-10-07
> Status: Ready for Implementation

## Tasks

### Phase 1: Parser Foundation

**Task 1.1: Create Store Expression AST Node**
- [ ] Create `ast/store.go` with `StoreExpressionNode` struct
- [ ] Add `NodeType()` and `String()` methods
- [ ] Add store expression case to AST node visitor patterns
- [ ] Write unit tests for store node creation

**Task 1.2: Extend Fence Section Parser for Store Definitions**
- [ ] Add `Stores map[string]string` field to `ast.FenceSection`
- [ ] Parse `store storeName = { ... }` syntax in fence section parser
- [ ] Extract store name and object literal as string
- [ ] Handle multiple store definitions in single fence section
- [ ] Write tests for inline store parsing

**Task 1.3: Add Store Expression Parser**
- [ ] Create `parseStoreExpression()` in `parser/expressions.go`
- [ ] Detect `$` prefix in expression parser
- [ ] Parse store name (alphanumeric + underscore)
- [ ] Parse property access (dot notation, multiple levels)
- [ ] Return `StoreExpressionNode` from parser
- [ ] Write parser tests for valid store expressions

**Task 1.4: Integration with Expression Parser**
- [ ] Modify `parseExpression()` to route `$` prefix to store parser
- [ ] Ensure existing variable expressions still work
- [ ] Test store expressions in text content
- [ ] Test store expressions in attributes
- [ ] Test store expressions in conditionals/loops

### Phase 2: Transformation

**Task 2.1: Create Store Expression Transformer**
- [ ] Create `transformer/stores.go`
- [ ] Implement `transformStoreExpression()` function
- [ ] Generate `<span x-text="$store.storeName.prop">` for text context
- [ ] Generate `:attribute="$store.storeName.prop"` for attribute context
- [ ] Write transformation unit tests

**Task 2.2: Handle Store Expressions in Conditionals**
- [ ] Transform `{if $store.prop}` to `x-if="$store.storeName.prop"`
- [ ] Test nested conditionals with store expressions
- [ ] Verify template x-if wrapper generation
- [ ] Write integration tests

**Task 2.3: Handle Store Expressions in Loops**
- [ ] Transform `{for item in $store.items}` to `x-for="item in $store.storeName.items"`
- [ ] Handle store property access in loop body
- [ ] Test nested loops with stores
- [ ] Write integration tests

**Task 2.4: Track Store References During Transformation**
- [ ] Add store tracking to transformer state
- [ ] Collect all referenced store names
- [ ] Map store names to definitions (from fence section)
- [ ] Pass store map to renderer

### Phase 3: Rendering & Server

**Task 3.1: Create Store Initialization Renderer**
- [ ] Create `renderer/stores.go`
- [ ] Implement `renderStoreInitializations()` function
- [ ] Generate `Alpine.store('name', {...})` calls
- [ ] Wrap in `document.addEventListener('alpine:init', ...)`
- [ ] Insert before Alpine.js initialization
- [ ] Write rendering unit tests

**Task 3.2: Integrate Store Rendering into HTML Output**
- [ ] Modify main render function to include store initialization
- [ ] Place store script after Alpine.js script tag
- [ ] Ensure stores load before Alpine.start()
- [ ] Test complete HTML output structure

**Task 3.3: Add Store File Discovery to Server**
- [ ] Create `registerStores()` function in `cmd/server/main.go`
- [ ] Scan `stores/` directory for `.js` files
- [ ] Parse store file content
- [ ] Extract store name from filename
- [ ] Build global store registry
- [ ] Log registered stores at startup

**Task 3.4: Implement Store Import Resolution**
- [ ] Extend import parser to recognize `import store from './stores/name.js'`
- [ ] Load external store file content
- [ ] Parse store definition from file
- [ ] Add to fence section stores map
- [ ] Test external store imports

**Task 3.5: Merge Inline and External Stores**
- [ ] Combine stores from fence section with external stores
- [ ] Inline stores override external stores (same name)
- [ ] Pass combined store map to transformer
- [ ] Test merge priority

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

## Estimated Timeline

- **Phase 1**: 3-4 days (Parser foundation)
- **Phase 2**: 3-4 days (Transformation)
- **Phase 3**: 4-5 days (Rendering & server)
- **Phase 4**: 3-4 days (Integration testing)
- **Phase 5**: 2-3 days (Documentation)

**Total**: 15-20 days (3-4 weeks)

## Success Criteria

1. Parse inline store definitions: `store auth = { loggedIn: false }`
2. Parse store references: `{$auth.loggedIn}`
3. Transform to Alpine.js: `<span x-text="$store.auth.loggedIn">`
4. Generate store initialization: `Alpine.store('auth', { loggedIn: false })`
5. Load external store files from `stores/` directory
6. Cross-component reactivity works (change in one component updates others)
7. All existing tests pass (no regressions)
8. Documentation complete with examples

## Dependencies Between Tasks

- Task 2.x depends on Task 1.x (transformation needs parser)
- Task 3.2 depends on Task 3.1 (rendering integration needs renderer)
- Task 3.5 depends on Task 3.3 and 3.4 (merge needs discovery and import resolution)
- Task 4.x depends on Task 1.x, 2.x, 3.x (integration tests need complete implementation)
- Task 5.x can start after Task 3.x (examples need working implementation)
