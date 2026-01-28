# Component Style Aggregation

**Date:** 2025-10-07
**Status:** 📝 Specification Complete - Ready for Implementation
**Priority:** 🔴 High - Fixes HeaderSimple flashing bug

## Quick Summary

Automatically extract and aggregate `<style>` blocks from components so their styles are included in the parent page output. This fixes the HeaderSimple flashing issue and eliminates manual style copying.

## Problem

Components with `<style>` blocks (like HeaderSimple.html) don't have their styles included when imported, causing:
- Visual bugs (components flash and disappear)
- Manual workarounds (copying styles to parent pages)
- Maintenance burden (style changes need multiple file updates)

## Solution

Implement automatic style aggregation similar to Svelte/Plenti:

1. **Extract**: Parse `<style>` blocks into AST
2. **Traverse**: Walk component dependency tree
3. **Aggregate**: Collect styles (dependencies first)
4. **Deduplicate**: Remove identical style blocks
5. **Inject**: Add aggregated styles to page output

## Key Features

- ✅ **Zero Config**: No template syntax changes
- ✅ **Automatic**: Styles just work when components are imported
- ✅ **Smart Ordering**: Dependencies' styles come first
- ✅ **Deduplication**: No duplicate styles
- ✅ **Debuggable**: Source comments show which component contributed styles
- ✅ **Performant**: Cached aggregation, <10ms overhead

## Implementation Plan

### Phase 1: Parser Enhancement
Ensure `<style>` blocks are extracted into `StyleSection` AST nodes.

### Phase 2: Style Aggregation Logic
Create `renderer/styles.go` with tree traversal and aggregation.

### Phase 3: Renderer Integration
Modify `renderer/render.go` to inject aggregated styles.

### Phase 4: Caching
Add per-component style cache for performance.

### Phase 5: Testing
Unit tests for aggregation logic, integration tests for real components.

## Files Changed

**New Files:**
- `renderer/styles.go` - Style aggregation logic
- `renderer/styles_test.go` - Unit tests

**Modified Files:**
- `parser/parser.go` - Ensure style extraction
- `renderer/render.go` - Inject aggregated styles

**Unchanged (Already Ready):**
- `ast/ast.go` - StyleSection already exists ✅
- `transformer/components.go` - Already stores templates ✅

## Example Output

```html
<style>
/* Styles from: HeaderSimple */
.header { background-color: #f8f9fa; }
.brand svg { height: 32px; }

/* Styles from: Footer */
.footer { background-color: #333; }
</style>
```

## Testing Checklist

- [ ] Single component with styles
- [ ] Nested components (parent imports child)
- [ ] Multiple components on one page
- [ ] Deduplication of identical styles
- [ ] Circular dependencies (no infinite loop)
- [ ] Empty/missing style blocks
- [ ] HeaderSimple manual test (no flashing)

## Success Criteria

1. HeaderSimple displays correctly without flashing
2. No manual style copying needed
3. All tests pass
4. <10ms performance overhead
5. Documentation updated

## Next Steps

1. Review and approve this spec
2. Implement parser enhancement (Phase 1)
3. Implement style aggregation (Phase 2)
4. Integrate with renderer (Phase 3)
5. Add caching (Phase 4)
6. Write tests (Phase 5)
7. Test with HeaderSimple

## Related Documents

- **Full Spec**: `SPEC.md` - Detailed implementation guide
- **Future Work**: `/docs/FutureDevelopment.md` - Optional CSS scoping
- **AST Reference**: `/ast/ast.go` - StyleSection definition
- **Component System**: `/transformer/components.go` - Component registry

## Questions?

See `SPEC.md` for comprehensive details including:
- Detailed algorithm pseudocode
- Edge case handling
- Performance analysis
- Migration path
- Future enhancement plans
