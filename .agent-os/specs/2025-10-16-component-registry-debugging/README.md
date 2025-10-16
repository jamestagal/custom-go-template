# Component Registry Debugging - Documentation Index

**Date**: 2025-10-16
**Status**: 🔴 ACTIVE INVESTIGATION - Components not rendering due to registry syntax errors

This directory contains comprehensive documentation of all troubleshooting efforts for runtime component resolution and component registry generation issues.

---

## Quick Navigation

### For Immediate Help
👉 **Start here**: [CURRENT_STATUS.md](./CURRENT_STATUS.md) - Current state and decision points

### For Debugging
👉 **Error details**: [ERROR_REFERENCE.md](./ERROR_REFERENCE.md) - Error messages and debugging commands

### For Full Context
👉 **Complete history**: [TROUBLESHOOTING_HISTORY.md](./TROUBLESHOOTING_HISTORY.md) - All attempted fixes and lessons learned

### For Original Issue
👉 **Known issue doc**: [../2025-10-15-runtime-component-resolution/KNOWN_ISSUE_COMPONENT_REGISTRY_SYNTAX.md](../2025-10-15-runtime-component-resolution/KNOWN_ISSUE_COMPONENT_REGISTRY_SYNTAX.md)

---

## Current Situation

### The Problem
Dynamic components from JSON files (e.g., hero2436, services2437) are not rendering on the page because the component registry has JavaScript syntax errors.

### Current Error
```
SyntaxError: Unexpected token '(' (at component-registry.js:1793:83)
```

Caused by invalid JavaScript:
```javascript
${props.(start * 1) + index + 1}  // ❌ Cannot have ( after props.
```

### Impact
- **ALL** runtime components blocked from loading
- Components defined in `/content/pages/_index.json` not rendering
- Console shows "Failed to load component registry after 3 attempts"

---

## What's Working ✅

1. Parser correctly extracts x-data attributes as strings
2. Simple expressions convert: `{count}` → `${props.count}`
3. All 65 components registered successfully
4. Runtime resolution infrastructure in place
5. Scripts loading correctly (runtime-components.js)

---

## What's Broken 🔴

1. **Complex expressions** produce invalid syntax (BLOCKING)
2. **Loop variables** incorrectly get `props.` prefix
3. **'this.' prefix bug** in nested content objects (separate issue)

---

## Documentation Files

### [CURRENT_STATUS.md](./CURRENT_STATUS.md)
**Purpose**: Quick reference for current state
**Contents**:
- What's fixed vs what's broken
- Current blocking error
- Recommended next steps
- Decision matrix for fix approaches

**Read this if**: You need to understand the current state and make a decision on how to proceed.

---

### [ERROR_REFERENCE.md](./ERROR_REFERENCE.md)
**Purpose**: Error debugging guide
**Contents**:
- Current error message with location
- Why the syntax is invalid
- Previous errors (now fixed)
- Common error patterns
- Debugging commands

**Read this if**: You're seeing console errors and need to debug the component registry.

---

### [TROUBLESHOOTING_HISTORY.md](./TROUBLESHOOTING_HISTORY.md)
**Purpose**: Complete chronological record
**Contents**:
- All 8 attempted fixes with results
- Root cause analysis
- Architectural issues identified
- Lessons learned
- Recommended solutions (3 options)

**Read this if**: You need full context of what's been tried, want to understand the architectural issues, or need to pick up where we left off.

---

## Quick Start Guide

### If you're a developer trying to fix this:

1. Read [CURRENT_STATUS.md](./CURRENT_STATUS.md) (5 min)
2. Check [ERROR_REFERENCE.md](./ERROR_REFERENCE.md) for current error details (2 min)
3. Decide on fix approach: Quick (Option 1) or Proper (Option 2)
4. Read relevant section in [TROUBLESHOOTING_HISTORY.md](./TROUBLESHOOTING_HISTORY.md) (10 min)
5. Implement fix following the documented patterns

**Estimated time to context**: 15-20 minutes

---

### If you're a user experiencing this issue:

**Workarounds** (use these until fix is implemented):

1. **Avoid complex expressions in x-data**:
   ```html
   <!-- Instead of: -->
   <div x-data="{ count: {count}, value: {(start * 1) + index} }">

   <!-- Use: -->
   <div x-data="{ count: 0, value: 0 }" x-init="count = {count}; value = {start} + {index}">
   ```

2. **Define data in fence section**:
   ```html
   ---
   prop count = 0
   prop start = 1
   ---
   <div x-data="{ count, start }">  <!-- No template expressions needed -->
   ```

3. **Use separate variables**:
   ```html
   ---
   prop count = 0
   function getInitialValue() { return count + 1; }
   ---
   <div x-data="{ count, value: getInitialValue() }">
   ```

---

## Timeline Summary

| Date | Event | Outcome |
|------|-------|---------|
| 2025-10-15 | Initial report: Components not rendering | Investigation started |
| 2025-10-16 | Fixed parser quote handling | ✅ x-data extraction works |
| 2025-10-16 | Added builder expression conversion | ✅ Simple expressions work |
| 2025-10-16 | Fixed component registration | ✅ 65 components registered |
| 2025-10-16 | Fixed script loading paths | ✅ runtime-components.js loads |
| 2025-10-16 | Discovered complex expression issue | 🔴 Line 1793 syntax error |
| 2025-10-16 | Created comprehensive documentation | 📚 This directory |

---

## Key Files in Codebase

### Parser
- `parser/html.go` (lines 584-620) - ✅ Fixed: Quote handling
- `parser/expressions.go` - Expression parsing

### Builder
- `builder/registry_generator.go` - ⚠️ Partial fix: Simple expressions work
- `builder/registry_generator_test.go` - Tests passing for simple cases

### Output
- `static/js/component-registry.js` - 🔴 Has syntax error at line 1793

### Runtime
- `static/js/runtime-components.js` - Client-side component loading
- `layouts/content/pages.html` - Component iterator layout
- `content/pages/_index.json` - Component data source

---

## Recommended Fix Approaches

### Option 1: Improved Regex (Quick - 2-4 hours)
**Pros**: Unblocks component rendering today
**Cons**: Won't handle all edge cases, technical debt

**Implementation**:
- Match individual identifiers in expressions
- Add skip list for loop variables and Alpine built-ins
- Handle parentheses and operators correctly

**Best for**: Getting components working quickly while planning proper fix

---

### Option 2: AST-Level Processing (Proper - 1-2 days)
**Pros**: Solves all expression handling issues permanently
**Cons**: Significant refactoring, medium risk

**Implementation**:
- Refactor `Attribute.Value` to support `[]Node` instead of just `string`
- Parser creates `ExpressionNode` objects in attributes
- Transformer handles with full scope context
- Builder renders pre-transformed expressions

**Best for**: Long-term architectural soundness

---

### Option 3: Hybrid Approach (Recommended)
**Phase 1** (Today): Implement Option 1 to unblock
**Phase 2** (Next sprint): Implement Option 2 for proper fix

---

## Testing Commands

```bash
# Validate component registry syntax
node -c static/js/component-registry.js

# Find problematic patterns
grep -n '\${props\.(' static/js/component-registry.js

# Test homepage
curl -s http://localhost:3333/ | grep "Welcome to Artistitch"

# Regenerate registry
pkill -f "go run cmd/server/main.go"
cd /Users/benjaminwaller/Projects/Jim\ Fisk/custom_go_template
go run cmd/server/main.go
```

---

## Success Criteria

The issue is resolved when:

- [ ] `node -c static/js/component-registry.js` passes with no errors
- [ ] Components from JSON render on homepage (hero2436, services2437)
- [ ] No console errors about component registry
- [ ] Complex expressions work: `{(start * 1) + index}` → `${(props.start * 1) + index}`
- [ ] Loop variables NOT prefixed: `index` stays as `index`
- [ ] Alpine built-ins NOT prefixed: `$store` stays as `$store`

---

## Contact & Support

**For questions about this documentation**:
- Check the detailed docs in this directory
- Review the code references provided
- See commit history for context

**For implementing fixes**:
- Follow the patterns in TROUBLESHOOTING_HISTORY.md
- Use the test cases provided
- Reference the cognitive load estimates

---

## Document Maintenance

**When updating these docs**:
1. Update the relevant specific file (CURRENT_STATUS, ERROR_REFERENCE, or TROUBLESHOOTING_HISTORY)
2. Update timestamps and status flags
3. Add new entries to the timeline
4. Update this README if navigation changes

**Version History**:
- 2025-10-16 21:50 UTC - Initial creation with comprehensive troubleshooting docs
- 2025-10-16 21:55 UTC - Added cross-references to known issue doc

---

**Last Updated**: 2025-10-16 21:55 UTC
**Status**: 🔴 ACTIVE - Blocking issue, documented, ready for fix implementation
