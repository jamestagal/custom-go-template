# Spec Tasks

## Component Style Aggregation Implementation

**Spec Location:** `.agent-os/specs/2025-10-07-component-style-aggregation/SPEC.md`

**Goal:** Automatically extract and aggregate `<style>` blocks from components so their styles are included in parent page output, fixing the HeaderSimple flashing issue.

---

## Tasks

- [ ] 1. Parser Enhancement: Ensure Style Extraction
  - [ ] 1.1 Write tests for style section parsing
  - [ ] 1.2 Verify `<style>` blocks are extracted into `StyleSection` AST nodes
  - [ ] 1.3 Ensure `StyleSection` nodes are added to `Template.RootNodes`
  - [ ] 1.4 Handle multiple `<style>` blocks in single component
  - [ ] 1.5 Handle empty `<style>` blocks gracefully
  - [ ] 1.6 Verify all parser tests pass

- [ ] 2. Style Aggregation Core Logic
  - [ ] 2.1 Write tests for `AggregateComponentStyles` function
  - [ ] 2.2 Create `renderer/styles.go` with `StyleBlock` struct
  - [ ] 2.3 Implement component dependency tree traversal with cycle detection
  - [ ] 2.4 Implement style collection (dependencies first, then parent)
  - [ ] 2.5 Implement deduplication using SHA256 content hashing
  - [ ] 2.6 Implement style ordering with source comments
  - [ ] 2.7 Handle edge cases (empty styles, circular deps, no styles)
  - [ ] 2.8 Verify all aggregation tests pass

- [ ] 3. Renderer Integration
  - [ ] 3.1 Write integration tests for renderer with style injection
  - [ ] 3.2 Modify `renderer/render.go` to call `AggregateComponentStyles`
  - [ ] 3.3 Inject aggregated styles into appropriate location (head or style section)
  - [ ] 3.4 Ensure styles are injected only once per page
  - [ ] 3.5 Verify integration tests pass

- [ ] 4. Performance Optimization: Caching
  - [ ] 4.1 Write tests for style cache functionality
  - [ ] 4.2 Implement per-component style cache with mutex protection
  - [ ] 4.3 Implement `GetAggregatedStyles` with cache lookup
  - [ ] 4.4 Implement `ClearStyleCache` for dev mode reloads
  - [ ] 4.5 Add cache hit/miss logging for debugging
  - [ ] 4.6 Verify cache tests pass and performance is <10ms overhead

- [ ] 5. Real-World Testing and Validation
  - [ ] 5.1 Test HeaderSimple component (primary use case)
  - [ ] 5.2 Verify HeaderSimple no longer flashes on page load
  - [ ] 5.3 Test page with multiple components (Header, Footer, etc.)
  - [ ] 5.4 Test nested component imports (3+ levels deep)
  - [ ] 5.5 Inspect rendered HTML to verify style comments and content
  - [ ] 5.6 Remove manual style workarounds from home.html
  - [ ] 5.7 Performance test: Verify <10ms overhead on typical page
  - [ ] 5.8 Verify all tests pass (unit + integration)

---

## Task Execution Notes

### Dependencies
- Task 1 must complete before Task 2 (need StyleSection nodes)
- Task 2 must complete before Task 3 (need aggregation logic)
- Task 3 must complete before Task 4 (need basic integration)
- Task 5 can begin after Task 3 (parallel to Task 4)

### Testing Strategy
- Follow TDD: Write tests first for each major component
- Each task includes verification step as final subtask
- Integration tests validate end-to-end functionality
- Real-world testing ensures HeaderSimple issue is resolved

### Success Criteria
- All unit tests pass ✅
- All integration tests pass ✅
- HeaderSimple displays without flashing ✅
- No manual style copying needed ✅
- Performance overhead <10ms ✅
- Code is documented and maintainable ✅

### Files to Create/Modify

**New Files:**
- `renderer/styles.go` - Style aggregation logic
- `renderer/styles_test.go` - Unit tests
- `tests/components/style_aggregation_test.go` - Integration tests

**Modified Files:**
- `parser/parser.go` - Ensure style extraction (may already work)
- `renderer/render.go` - Inject aggregated styles

**Unchanged (Already Ready):**
- `ast/ast.go` - StyleSection already exists ✅
- `transformer/components.go` - Already stores templates ✅
