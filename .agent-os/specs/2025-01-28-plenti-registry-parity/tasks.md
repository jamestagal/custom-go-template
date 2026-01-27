# Plenti Registry Parity - Implementation Tasks

## Overview

**Priority:** CRITICAL (required for Plenti compatibility)
**Estimated Total:** 16-20 hours
**Dependencies:** None (foundational change)

---

## Phase 1: Signature System

### Task 1.1: Create Signature Generator
**File:** `builder/signature.go` (NEW)
**Effort:** 2 hours

- [ ] Create `GenerateSignature(filePath string) string`
- [ ] Handle path normalization (forward slashes)
- [ ] Replace `/` and `.` with `_`
- [ ] Preserve `layouts_` prefix
- [ ] Unit tests for all layout categories

**Test cases:**
```go
assert(GenerateSignature("layouts/components/Hero2436.html") == "layouts_components_Hero2436_html")
assert(GenerateSignature("layouts/content/_index.html") == "layouts_content__index_html")
assert(GenerateSignature("layouts/global/nav.html") == "layouts_global_nav_html")
```

### Task 1.2: Create Signature Parser
**File:** `builder/signature.go`
**Effort:** 1 hour

- [ ] Create `ParseSignature(signature string) SignatureInfo`
- [ ] Extract category (components, content, global, scripts)
- [ ] Extract name
- [ ] Extract extension
- [ ] Add helper methods: `IsComponent()`, `IsGlobal()`, `IsContentTemplate()`

### Task 1.3: Signature Unit Tests
**File:** `builder/signature_test.go` (NEW)
**Effort:** 1 hour

- [ ] Test generation for each category
- [ ] Test parsing for each category
- [ ] Test round-trip (generate → parse → matches original)
- [ ] Test edge cases (underscores in names, nested paths)

---

## Phase 2: Registry Generator Updates

### Task 2.1: Update ComponentTemplate Struct
**File:** `builder/registry_generator.go`
**Effort:** 1 hour

- [ ] Add `Signature string` field
- [ ] Add `FilePath string` field
- [ ] Add `Category string` field
- [ ] Update all code that creates ComponentTemplate

### Task 2.2: Update Component Registration
**File:** `cmd/server/main.go`
**Effort:** 2 hours

- [ ] Extract file path during component scanning
- [ ] Generate signature for each component
- [ ] Determine category from path
- [ ] Pass full ComponentTemplate to registry generator

### Task 2.3: Update Registry Output Format
**File:** `builder/registry_generator.go`
**Effort:** 1 hour

- [ ] Use signature as key (not component name)
- [ ] Add header comment with format documentation
- [ ] Ensure Plenti-compatible output structure

**Expected output:**
```javascript
export default {
  'layouts_components_Hero2436_html': (props) => `...`,
  'layouts_content_pages_html': (props) => `...`,
};
```

### Task 2.4: Registry Generator Tests
**File:** `builder/registry_generator_test.go`
**Effort:** 1 hour

- [ ] Test output uses signatures as keys
- [ ] Test all categories represented
- [ ] Test output is valid JavaScript
- [ ] Test compatibility with Plenti lookup pattern

---

## Phase 3: Runtime Updates

### Task 3.1: Update Runtime Component Resolution
**File:** `static/js/runtime-components.js`
**Effort:** 2 hours

- [ ] Add `resolveComponentSignature(name)` function
- [ ] Convert short name to full signature
- [ ] Support both short names and full signatures
- [ ] Case-insensitive matching for short names

```javascript
function resolveComponentSignature(name) {
  // If already a signature, return as-is
  if (name.startsWith('layouts_')) {
    return name;
  }
  // Convert short name to signature
  return `layouts_components_${name}_html`;
}
```

### Task 3.2: Create Plenti-Compatible Main.js
**File:** `static/js/core/main.js` (NEW)
**Effort:** 2 hours

- [ ] Import allContent and allLayouts
- [ ] Read `data-content-filepath` from HTML
- [ ] Find content for current page
- [ ] Initialize Alpine.js store with Plenti data
- [ ] Register `$component` magic helper
- [ ] Export globals (`window.allLayouts`, etc.)

### Task 3.3: Runtime Integration Tests
**File:** `tests/integration/plenti_runtime_test.go`
**Effort:** 1 hour

- [ ] Test signature resolution
- [ ] Test short name → signature conversion
- [ ] Test case-insensitive lookup
- [ ] Test `allLayouts[signature]` access pattern

---

## Phase 4: Fingerprinting

### Task 4.1: Create Fingerprint Generator
**File:** `builder/fingerprint.go` (NEW)
**Effort:** 2 hours

- [ ] Hash source files (SHA-256)
- [ ] Return first 10 characters (Plenti pattern)
- [ ] Include only relevant files (.html, .js, .css)
- [ ] Handle incremental builds (cache hash)

### Task 4.2: Create Fingerprinted Output Structure
**File:** `builder/fingerprint.go`
**Effort:** 1 hour

- [ ] Create fingerprinted directory
- [ ] Create subdirectory structure (core/, generated/, layouts/)
- [ ] Copy/generate files to correct locations

### Task 4.3: Update HTML Renderer
**File:** `renderer/render.go`
**Effort:** 1 hour

- [ ] Accept fingerprint parameter
- [ ] Add `data-content-filepath` attribute
- [ ] Use fingerprinted paths for scripts/styles
- [ ] Ensure base href is set

### Task 4.4: Fingerprinting Tests
**File:** `builder/fingerprint_test.go` (NEW)
**Effort:** 1 hour

- [ ] Test hash consistency (same input = same hash)
- [ ] Test hash changes when files change
- [ ] Test directory structure creation
- [ ] Test output file placement

---

## Phase 5: Content Integration

### Task 5.1: Generate content.js
**File:** `builder/content_generator.go` (NEW)
**Effort:** 2 hours

- [ ] Scan content directory
- [ ] Build allContent array
- [ ] Match Plenti's content format exactly
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

### Task 5.2: Content Format Compatibility
**Effort:** 1 hour

- [ ] Verify `pager` field handling
- [ ] Verify `type` field derivation
- [ ] Verify `path` field format
- [ ] Verify `fields` extraction

### Task 5.3: End-to-End Content Test
**File:** `tests/integration/plenti_content_test.go`
**Effort:** 1 hour

- [ ] Load real Plenti content JSON
- [ ] Generate content.js
- [ ] Verify output matches Plenti format
- [ ] Test with multiple content types

---

## Dependency Graph

```
Phase 1 (Signatures)
    │
    ▼
Phase 2 (Registry) ──────► Phase 3 (Runtime)
    │                           │
    ▼                           ▼
Phase 4 (Fingerprinting) ◄──────┘
    │
    ▼
Phase 5 (Content Integration)
```

---

## Verification Checklist

### After Phase 1
- [ ] `GenerateSignature()` produces Plenti-compatible output
- [ ] `ParseSignature()` extracts correct components
- [ ] All tests pass

### After Phase 2
- [ ] `component-registry.js` uses signatures as keys
- [ ] Format matches Plenti's `generated/layouts.js`
- [ ] All tests pass

### After Phase 3
- [ ] `allLayouts["layouts_components_" + name + "_html"]` works
- [ ] Short name resolution works
- [ ] `core/main.js` initializes correctly
- [ ] All tests pass

### After Phase 4
- [ ] Fingerprinted directory created
- [ ] Hash changes when source changes
- [ ] HTML uses fingerprinted paths
- [ ] All tests pass

### After Phase 5
- [ ] `content.js` format matches Plenti
- [ ] `data-content-filepath` enables page lookup
- [ ] Real Plenti content works without modification
- [ ] All tests pass

---

## Quick Reference: Plenti Patterns

### Signature Format
```
{folder}_{subfolder}_{name}_{extension}
layouts_components_Hero2436_html
```

### Registry Lookup
```javascript
allLayouts["layouts_components_" + name + "_html"]
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
