# Spec Requirements Document

> Spec: Fix Home.html Regression - UserProfile Component
> Created: 2025-10-08

## Overview

Fix the home.html page regression on the global-store-system branch where UserProfile component functions are being stripped during fence section parsing with store support, causing the page to display empty UserProfile cards with console errors.

## User Stories

### Restore Working Home Page

As a developer working on the global-store-system branch, I want the home.html page to render correctly without console errors, so that I can verify the store system doesn't break existing functionality and can safely merge the branch to main.

**Workflow**: Developer runs the server on global-store-system branch and navigates to http://localhost:3333/. The page should display three UserProfile components with user avatars (initials), names, roles, and role badges - all without console errors.

### Preserve Component Functions During Store Parsing

As the template engine, I need to preserve all fence section functions (formatDate, getRoleBadge) when parsing components with store imports, so that components maintain their full functionality regardless of whether they use stores or not.

**Workflow**: When a component's fence section is parsed with `ParseFenceContentWithStores()`, all function definitions must be preserved in the rendered x-data object, not stripped out during the parsing process.

## Spec Scope

1. **Root Cause Investigation** - Use go-backend agent to identify why `ParseFenceContentWithStores()` strips function definitions from UserProfile component fence section
2. **Fence Parser Fix** - Modify fence parsing logic to preserve function definitions when processing store imports
3. **Component Rendering Fix** - Ensure component registration and rendering preserves all fence functions regardless of store usage
4. **Regression Tests** - Add tests to prevent future regressions where functions are stripped during fence parsing
5. **Verification Testing** - Test all pages (home.html, store-components-demo.html, etc.) to ensure no regressions

## Out of Scope

- Refactoring the entire fence parsing system
- Adding new features to the store system
- Fixing unrelated test failures (transformer/renderer test expectations)
- Performance optimization of fence parsing

## Expected Deliverable

1. Home page (http://localhost:3333/) displays correctly with NO console errors
2. All three UserProfile components show user initials, names, roles, and colored role badges
3. Store demo page (http://localhost:3333/store-components-demo) continues working perfectly
4. All existing pages render without errors
5. Regression tests added to prevent function stripping in future
