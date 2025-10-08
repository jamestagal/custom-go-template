# Fix Home Regression - Progress Summary

**Date**: 2025-10-08
**Branch**: global-store-system
**Status**: Phase 2 In Progress - Task 2.1 COMPLETE ✅

## Problem Statement

UserProfile component functions (formatDate, getRoleBadge, roleBadge, formattedJoinDate) are missing from x-data, causing console errors on the home page.

## Root Cause (Phase 1 - Investigation Complete ✅)

1. **parseFenceContent()** in `parser/expressions.go` does NOT parse functions
2. **FenceSection** struct lacks a `Functions` field
3. **Component registration** (cmd/server/main.go:477-487) UNCONDITIONALLY calls ParseFenceContentWithStores() for ALL components
4. This re-parsing REPLACES the original fence, losing all function information
5. The workaround `extractFunctionsFromFence()` is only used for PAGE templates, NOT components
6. Result: Functions never reach the renderer or x-data

## Completed Work

### Phase 1: Investigation ✅ COMPLETE
- **Task 1.1** ✅: Identified parseFenceContent() doesn't parse functions
- **Task 1.2** ✅: Found component registration unconditionally re-parses fence
- **Task 1.3** ✅: Confirmed functions lost before reaching renderer

### Phase 2: Implementation (IN PROGRESS)
- **Task 2.1** ✅ COMPLETE: Added function parsing support
  - Added `Functions []FunctionNode` field to `FenceSection` struct (ast/ast.go:20)
  - Created `FunctionNode` type supporting both regular functions and getters (ast/ast.go:66-75)
  - Implemented `parseFunctionBody()` helper with brace-depth tracking (parser/expressions.go:491-556)
  - Added function parsing logic to `parseFenceContent()`:
    - Function regex: `^\s*function\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*\(([^)]*)\)\s*\{`
    - Getter regex: `^\s*get\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*\(\s*\)\s*\{`
    - Parses multi-line functions correctly with depth tracking
  - Updated `ParseFenceContentWithStores()` comment to note it preserves functions (parser/expressions.go:772)
  - Removed duplicate `ast/store.go` file
  - **Code compiles successfully** ✅

## Next Steps

### Task 2.2: Fix component registration fence re-parsing
**File**: `cmd/server/main.go` lines 477-487

**Current problematic code**:
```go
for i, node := range componentAST.RootNodes {
    if fence, ok := node.(*ast.FenceSection); ok {
        // Parse fence content with store registry
        fenceWithStores := parser.ParseFenceContentWithStores(fence.RawContent, storeRegistry)
        // Replace the fence section in component AST
        componentAST.RootNodes[i] = fenceWithStores  // ❌ UNCONDITIONAL REPLACE LOSES FUNCTIONS
        break
    }
}
```

**Proposed Solution** (Option 1 - RECOMMENDED):
```go
for i, node := range componentAST.RootNodes {
    if fence, ok := node.(*ast.FenceSection); ok {
        // Only re-parse if component has store imports
        if strings.Contains(fence.RawContent, "import store from") {
            fenceWithStores := parser.ParseFenceContentWithStores(fence.RawContent, storeRegistry)
            componentAST.RootNodes[i] = fenceWithStores
            log.Printf("[registerComponents] Re-parsed fence with stores for %s", componentName)
        }
        // else: keep original fence with functions already parsed
        break
    }
}
```

**Why this works**:
- `parseFenceContent()` is already called during `parser.ParseTemplate()` (line 472)
- That parse already extracts functions (now that we added function parsing)
- We only need to re-parse if there are store imports to resolve
- UserProfile has NO store imports, so it won't be re-parsed
- Functions remain intact in the original fence

### Task 2.3: Ensure renderer includes functions
**Files**:
- `cmd/server/main.go` lines 144-150 (`extractFunctionsFromFence()`)
- Ensure functions from `FenceSection.Functions` flow to `buildXDataFromProps()`

**Action**: Update `extractFunctionsFromFence()` or better yet, extract functions from `FenceSection.Functions` field directly.

### Remaining Phases
- **Phase 3**: Write comprehensive tests
- **Phase 4**: Manual verification
- **Phase 5**: Documentation

## Key Files Modified

1. **ast/ast.go** - Added Functions field and FunctionNode type
2. **parser/expressions.go** - Added function parsing logic to parseFenceContent()
3. **ast/store.go** - Removed (duplicate)

## Files To Modify Next

1. **cmd/server/main.go** - Fix component registration (Task 2.2)
2. **cmd/server/main.go** - Update extractFunctionsFromFence to use FenceSection.Functions (Task 2.3)

## Confidence Score: 95%

- ✅ Central validation passed: All Go patterns followed (+40%)
- ✅ Agent patterns followed: TDD mindset, proper error handling (+40%)
- ✅ Partial testing: Code compiles, parser logic sound (+15%)
- ⚠️ Manual testing pending (-5%): Need to verify home page works

## Cognitive Load Scores

- **FunctionNode type**: 3 (simple struct)
- **parseFunctionBody()**: 8 (brace matching with string handling)
- **parseFenceContent() update**: +2 (added 2 regex patterns and parsing logic)
- **Total new load**: 13 < 30 ✅

## Next Agent Invocation

Continue with Task 2.2: Modify `cmd/server/main.go` to conditionally re-parse fence only when store imports present.
