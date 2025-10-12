# Known Issues

This document tracks known issues and regressions in the codebase that are being worked on or need to be addressed in future releases.

## Active Issues

### 1. Home.html Regression - UserProfile Component Broken (2025-10-08)

**Branch:** `global-store-system`
**Status:** Active - To be fixed separately
**Severity:** High
**Affected Files:** `examples/pages/home.html`, `examples/components/UserProfile.html`

#### Description

The home.html page is broken on the `global-store-system` branch with multiple console errors related to the UserProfile component. The page worked perfectly on the main branch before the global store system implementation.

#### Errors Observed

```
Cannot read properties of undefined (reading 'charAt')
  at home.html:1:1

Cannot read properties of undefined (reading 'toLowerCase')
  at getRoleBadge (home.html:1:1)

Cannot set properties of null (setting '_x_dataStack')
  at cdn.min.js:5:11927
```

These errors occur 3 times each (once for each UserProfile component instance on the page), totaling 6 console errors.

#### Root Cause

The fence section parsing with store support (`ParseFenceContentWithStores()`) or component rendering is stripping function definitions from the UserProfile component's fence section. Specifically:

- `formatDate()` function is missing from rendered x-data
- `getRoleBadge()` function is missing from rendered x-data
- This causes undefined references when the template tries to call these functions

#### Impact

- Home page displays empty UserProfile cards (no content)
- All UserProfile component functionality is broken
- User avatar initials not displayed (calls `.charAt(0)` on undefined)
- Role badges not rendered (calls `.toLowerCase()` on undefined role)

#### Workaround

None currently. The main branch version works correctly.

#### Notes

- The `/store-components-demo` page works perfectly - this issue is specific to home.html
- Store transformation appears to interfere with component function preservation
- Attempted fixes with conditional fence re-parsing did not resolve the issue
- Defensive null checks in UserProfile.html did not resolve the issue

#### Next Steps

1. Investigate why `ParseFenceContentWithStores()` strips functions
2. Ensure fence parsing preserves all function declarations
3. Add regression tests for component functions in fence sections
4. Fix the fence parser to handle both store imports AND function preservation

---

## Resolved Issues

None yet.

---

## Issue Reporting

When reporting new issues, please include:

1. Branch name where issue occurs
2. Steps to reproduce
3. Expected vs actual behavior
4. Console errors (if applicable)
5. Affected files
6. Whether issue is a regression (worked before) or new
