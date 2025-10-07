# Spec-Lite: Fix Server x-data Building Architecture

**Date**: 2025-10-07 | **Priority**: High | **Complexity**: Low

## Problem
Server bypasses transformer and manually builds x-data using regex, causing functions to be extracted as truncated JSON strings instead of proper Alpine.js methods.

## Solution
Remove manual x-data building from `cmd/server/main.go` and use `renderer.Render()` to leverage the transformer's `alpineDataFormatter`.

## Changes
1. **cmd/server/main.go**: Remove regex extraction, call `renderer.Render()` (~100 lines removed, ~20 added)
2. **examples/pages/comprehensive-simple.html**: Restore functions to test file (~30 lines added)

## Success Criteria
- ✅ Functions render as Alpine.js method shorthand (not JSON strings)
- ✅ Zero console errors when using functions
- ✅ All existing features continue working
- ✅ All tests pass

## Timeline
2-4 hours total

## Impact
**High** - Unblocks function support (formatters, computed values, event handlers)
