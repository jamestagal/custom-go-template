# Plenti Registry Parity - Implementation Tasks

## Overview

**Priority:** CRITICAL (required for Plenti compatibility)
**Estimated Total:** 18-22 hours
**Dependencies:** None (foundational change)

---

## Review-Driven Changes

This task list incorporates findings from the go-reviewer analysis:

| Issue | Resolution | Impact |
|-------|------------|--------|
| Dual Registry Problem | Create shared `types.ComponentTemplate` | Phase 1 |
| Phase 2+3 Coupling | Merged into single atomic phase | Phase 2 |
| Breaking Changes | Backward compatibility layer | Phase 2 |
| Missing Edge Cases | Added test requirements | All phases |
| Missing Regression Tests | Added to each phase | All phases |

---

## Phase 1: Shared Type & Signature System

### Task 1.1: Create Shared ComponentTemplate Type
**File:** `types/component.go` (NEW)
**Effort:** 2 hours

- [ ] Create `types/` package directory
- [ ] Define unified `ComponentTemplate` struct with all fields:
  - `Name` (short name)
  - `Signature` (Plenti signature)
  - `FilePath` (original path)
  - `Category` (components/content/global/scripts)
  - `AST` (parsed template)
  - `Props` (prop names)
- [ ] Implement `NewComponentTemplate(filePath, ast, props)` constructor
- [ ] Implement `ExtractNameFromPath(filePath)` helper
- [ ] Implement `CategoryFromPath(filePath)` helper
- [ ] Implement `GenerateSignature(filePath)` function

**Test cases:**
```go
// types/component_test.go
func TestNewComponentTemplate(t *testing.T) {
    ct := NewComponentTemplate("layouts/components/Hero2436.html", ast, []string{"title"})
    assert.Equal(t, "Hero2436", ct.Name)
    assert.Equal(t, "layouts_components_Hero2436_html", ct.Signature)
    assert.Equal(t, "components", ct.Category)
}
```

### Task 1.2: Update Signature Utilities
**File:** `builder/signature.go` (UPDATE)
**Effort:** 1 hour

- [ ] Import and re-export from `types` package
- [ ] Create `SignatureInfo` struct (or re-export from types)
- [ ] Implement `ParseSignature(signature)` function
- [ ] Implement `ShortNameToSignature(name, ext)` helper
- [ ] Implement `SignatureToShortName(signature)` helper
- [ ] Add `IsComponent()`, `IsGlobal()`, `IsContentTemplate()` methods

### Task 1.3: Signature Edge Case Tests
**File:** `builder/signature_test.go` (UPDATE)
**Effort:** 1.5 hours

- [ ] Test standard paths for all categories
- [ ] Test nested paths: `layouts/components/forms/Input.html`
- [ ] Test underscores in names: `layouts/components/Hero_2436.html`
- [ ] Test leading underscores: `layouts/content/_index.html`
- [ ] Test case sensitivity handling
- [ ] Test round-trip: generate → parse → reconstruct path
- [ ] Test invalid signatures return `Valid: false`

**Edge case test matrix:**

| Input Path | Expected Signature | Notes |
|-----------|-------------------|-------|
| `layouts/components/Hero2436.html` | `layouts_components_Hero2436_html` | Standard |
| `layouts/content/_index.html` | `layouts_content__index_html` | Leading underscore |
| `layouts/components/Hero_2436.html` | `layouts_components_Hero_2436_html` | Underscore in name |
| `layouts/components/forms/Input.html` | `layouts_components_forms_Input_html` | Nested path |
| `./layouts/global/nav.html` | `layouts_global_nav_html` | Leading `./` |

### Task 1.4: Phase 1 Regression Tests
**Effort:** 0.5 hours

- [ ] Run full test suite: `go test ./...`
- [ ] Verify no existing tests broken
- [ ] Document any test updates needed

---

## Phase 2: Registry & Runtime (Combined)

**CRITICAL:** Tasks 2.1-2.6 must be completed and deployed together.

### Task 2.1: Update Transformer Registry
**File:** `transformer/components.go` (UPDATE)
**Effort:** 2 hours

- [ ] Import `types` package
- [ ] Change `componentTemplateRegistry` to use `*types.ComponentTemplate`
- [ ] Update `RegisterComponent()` to:
  - Accept `*types.ComponentTemplate` parameter
  - Register by BOTH short name and signature
- [ ] Update `GetComponentTemplate()` to:
  - Accept name OR signature
  - Try exact match first
  - Fall back to case-insensitive match
- [ ] Update all callers of `RegisterComponent()`
- [ ] Update all callers of `GetComponentTemplate()`

**Backward compatibility requirement:**
```go
// BOTH must work:
GetComponentTemplate("Hero2436")                           // Short name
GetComponentTemplate("layouts_components_Hero2436_html")   // Full signature
```

### Task 2.2: Update Builder Registry Generator
**File:** `builder/registry_generator.go` (UPDATE)
**Effort:** 1.5 hours

- [ ] Import `types` package
- [ ] Update `GenerateComponentRegistry()` to accept `[]*types.ComponentTemplate`
- [ ] Use `Signature` as primary key in output
- [ ] Generate short-name aliases for backward compatibility
- [ ] Add header comments documenting format
- [ ] Remove old `ComponentTemplate` struct (use shared type)

**Expected output format:**
```javascript
// Auto-generated Plenti-compatible component registry
// Signature format: layouts_{category}_{name}_{extension}
// Lookup: allLayouts['layouts_components_' + name + '_html']

const registry = {
  'layouts_components_Hero2436_html': (props) => `...`,
  'layouts_content_pages_html': (props) => `...`,
};

// Short-name aliases for backward compatibility
registry['Hero2436'] = registry['layouts_components_Hero2436_html'];

export default registry;
```

### Task 2.3: Update Server Component Registration
**File:** `cmd/server/main.go` (UPDATE)
**Effort:** 1.5 hours

- [ ] Import `types` package
- [ ] Update component scanning to capture file paths
- [ ] Use `types.NewComponentTemplate()` to create components
- [ ] Pass `*types.ComponentTemplate` to `RegisterComponent()`
- [ ] Pass `[]*types.ComponentTemplate` to `GenerateComponentRegistry()`

### Task 2.4: Update Runtime Component Resolution
**File:** `static/js/runtime-components.js` (UPDATE)
**Effort:** 1 hour

- [ ] Add `resolveComponentSignature(name)` function
- [ ] Update `renderDynamicComponent()` to:
  - Try full signature first
  - Fall back to short name
  - Fall back to case-insensitive search
- [ ] Add debug logging for lookup failures
- [ ] Export new functions

### Task 2.5: Phase 2 Integration Tests
**File:** `tests/integration/plenti_registry_test.go` (NEW)
**Effort:** 1 hour

- [ ] Test Plenti signature lookup: `registry["layouts_components_Hero2436_html"]`
- [ ] Test short name lookup: `registry["Hero2436"]`
- [ ] Test case-insensitive lookup: `registry["hero2436"]`
- [ ] Test transformer lookup with both formats
- [ ] Test runtime resolution function

### Task 2.6: Phase 2 Regression Tests
**Effort:** 1 hour

- [ ] Run full test suite: `go test ./...`
- [ ] Test existing dynamic component examples
- [ ] Test existing static rendering
- [ ] Verify `component-registry.js` generates correctly
- [ ] Verify runtime components render correctly
- [ ] Document any breaking changes

---

## Phase 3: Fingerprinting

### Task 3.1: Create Fingerprint Generator
**File:** `builder/fingerprint.go` (NEW)
**Effort:** 2 hours

- [ ] Implement `GenerateBuildFingerprint(sourceDir)` function
- [ ] Hash all `.html`, `.js`, `.css` files
- [ ] Return first 10 characters of SHA-256 hash
- [ ] Handle errors gracefully (return empty string on failure)
- [ ] Add caching support for incremental builds (optional)

### Task 3.2: Create Fingerprinted Output Structure
**File:** `builder/fingerprint.go`
**Effort:** 1 hour

- [ ] Implement `CreateFingerprintedOutput(publicDir, fingerprint)` function
- [ ] Create directory structure matching Plenti:
  - `{fingerprint}/core/`
  - `{fingerprint}/generated/`
  - `{fingerprint}/layouts/components/`
  - `{fingerprint}/layouts/content/`
  - `{fingerprint}/layouts/global/`
- [ ] Return `BuildOutput` struct with paths

### Task 3.3: Update HTML Renderer
**File:** `renderer/render.go` (UPDATE)
**Effort:** 1 hour

- [ ] Add `fingerprint` parameter to render functions
- [ ] Add `data-content-filepath` attribute to `<html>` tag
- [ ] Use fingerprinted paths for `<script>` tags
- [ ] Use fingerprinted paths for `<link>` stylesheets
- [ ] Ensure `<base href="/">` is set

### Task 3.4: Create Runtime Entry Point
**File:** `static/js/core/main.js` (NEW)
**Effort:** 1 hour

- [ ] Import `allContent` from `../generated/content.js`
- [ ] Import `allLayouts` from `../generated/layouts.js`
- [ ] Read `data-content-filepath` from HTML
- [ ] Find content for current page
- [ ] Register Alpine.js store with Plenti data
- [ ] Register `$component` magic helper
- [ ] Export globals to `window`

### Task 3.5: Fingerprinting Tests
**File:** `builder/fingerprint_test.go` (NEW)
**Effort:** 1 hour

- [ ] Test hash consistency (same input = same hash)
- [ ] Test hash changes when files change
- [ ] Test directory structure creation
- [ ] Test output file placement
- [ ] Test error handling for missing directories

---

## Phase 4: Content Integration

### Task 4.1: Generate content.js
**File:** `builder/content_generator.go` (NEW)
**Effort:** 2 hours

- [ ] Implement `GenerateContentJS(contentDir)` function
- [ ] Scan content directory recursively
- [ ] Build `allContent` array with Plenti-compatible format
- [ ] Include all required fields:
  - `type` (from folder name)
  - `path` (URL path)
  - `filepath` (content file path)
  - `filename` (just the filename)
  - `fields` (parsed JSON content)
- [ ] Output as ES module

**Expected output:**
```javascript
const allContent = [
  {
    type: "pages",
    path: "about",
    filepath: "content/pages/about.json",
    filename: "about.json",
    fields: { title: "About", ... }
  },
  // ...
];
export default allContent;
```

### Task 4.2: Content Format Compatibility
**Effort:** 1 hour

- [ ] Verify `pager` field handling
- [ ] Verify `type` field derivation from folder
- [ ] Verify `path` field format (no leading slash, no extension)
- [ ] Verify `fields` extraction preserves all JSON content
- [ ] Test with nested content directories

### Task 4.3: End-to-End Content Test
**File:** `tests/integration/plenti_content_test.go` (NEW)
**Effort:** 1.5 hours

- [ ] Load real Plenti content JSON from test fixtures
- [ ] Generate `content.js`
- [ ] Verify output matches Plenti format exactly
- [ ] Test with multiple content types (pages, blog, etc.)
- [ ] Test `data-content-filepath` page lookup

### Task 4.4: Phase 4 Regression Tests
**Effort:** 0.5 hours

- [ ] Run full test suite: `go test ./...`
- [ ] Test full build pipeline with fingerprinting
- [ ] Verify HTML output includes all Plenti attributes
- [ ] Test with real Plenti site content

---

## Dependency Graph (Revised)

```
Phase 1 (Shared Type & Signatures)
    │
    ▼
Phase 2 (Registry & Runtime - ATOMIC)
    │
    ├─── transformer/components.go
    ├─── builder/registry_generator.go
    ├─── cmd/server/main.go
    └─── static/js/runtime-components.js
    │
    ▼
Phase 3 (Fingerprinting)
    │
    ▼
Phase 4 (Content Integration)
```

---

## Verification Checklist

### After Phase 1
- [ ] `types.ComponentTemplate` compiles and is importable
- [ ] `GenerateSignature()` produces Plenti-compatible output
- [ ] `ParseSignature()` handles all edge cases
- [ ] All existing tests pass: `go test ./...`

### After Phase 2
- [ ] `component-registry.js` uses signatures as primary keys
- [ ] Short-name aliases are generated
- [ ] `allLayouts["layouts_components_" + name + "_html"]` works in JS
- [ ] `registry["Hero2436"]` still works (backward compat)
- [ ] `GetComponentTemplate("Hero2436")` works in Go
- [ ] `GetComponentTemplate("layouts_components_Hero2436_html")` works in Go
- [ ] All existing tests pass: `go test ./...`

### After Phase 3
- [ ] Fingerprinted directory created (e.g., `public/aQwupMmCDl/`)
- [ ] Hash changes when source files change
- [ ] HTML uses fingerprinted paths for assets
- [ ] `core/main.js` exists and initializes correctly
- [ ] All tests pass

### After Phase 4
- [ ] `content.js` format matches Plenti exactly
- [ ] `data-content-filepath` attribute present on `<html>`
- [ ] `window.content` contains current page data
- [ ] Real Plenti content works without modification
- [ ] All tests pass

---

## Quick Reference: Plenti Patterns

### Signature Format
```
{folder}_{subfolder}_{name}_{extension}
layouts_components_Hero2436_html
```

### Registry Lookup (JavaScript)
```javascript
// Plenti pattern
allLayouts["layouts_components_" + name + "_svelte"]

// Go/Alpine.js pattern
allLayouts["layouts_components_" + name + "_html"]
```

### Go Registry Lookup
```go
// Both work after Phase 2:
GetComponentTemplate("Hero2436")
GetComponentTemplate("layouts_components_Hero2436_html")
```

### HTML Attribute
```html
<html data-content-filepath="content/pages/about.json">
```

### Fingerprint Path
```html
<script src="aQwupMmCDl/core/main.js">
```

### Content Format
```javascript
{
  type: "pages",
  path: "about",
  filepath: "content/pages/about.json",
  fields: { ... }
}
```

---

## Risk Mitigation

### Breaking Change Prevention

1. **Dual Registration**: Components registered under BOTH name and signature
2. **Fallback Lookups**: Try signature → short name → case-insensitive
3. **Alias Generation**: JS registry includes `registry['Hero2436'] = registry['layouts_...']`
4. **Regression Tests**: Run full test suite after each phase

### Rollback Plan

If issues discovered after deployment:
1. Revert to simple-name-only registration
2. Remove signature generation
3. Keep `types.ComponentTemplate` for future use
