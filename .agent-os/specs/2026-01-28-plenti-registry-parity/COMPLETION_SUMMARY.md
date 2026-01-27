# Plenti Registry Parity - Completion Summary

**Date Completed:** 2026-01-28
**Status:** ✅ Ready for Merge

---

## Overview

Implemented Plenti-compatible component signatures and registry structure to enable future merge of the custom Go template engine into Plenti.

---

## What Was Implemented

### Phase 1: Shared Type & Signature System ✅

**Files Created:**
- `types/component.go` - Canonical `ComponentTemplate` struct with auto-generated metadata

**Key Functions:**
| Function | Purpose |
|----------|---------|
| `NewComponentTemplate(filePath, ast, props)` | Constructor with auto-generated signature |
| `GenerateSignature(filePath)` | Converts path to Plenti signature |
| `ParseSignature(signature)` | Extracts category, name, extension |
| `ShortNameToSignature(name, ext)` | Builds lookup signature |
| `SignatureToShortName(signature)` | Extracts component name |

**Signature Format:**
```
layouts_{category}_{name}_{extension}

Examples:
  layouts/components/Hero2436.html → layouts_components_Hero2436_html
  layouts/content/_index.html      → layouts_content__index_html
  layouts/global/nav.html          → layouts_global_nav_html
```

### Phase 2: Registry & Runtime ✅

**Files Updated:**
- `transformer/components.go` - Dual registration (short name + signature)
- `builder/registry_generator.go` - Signatures as primary keys with aliases
- `cmd/server/main.go` - Uses `RegisterComponentTemplate()` with proper paths
- `core/runtime-components.js` - `resolveComponentSignature()` for Plenti lookup

**Lookup Resolution Order:**
1. Full Plenti signature: `layouts_components_Hero2436_html`
2. Short name: `Hero2436`
3. Case-insensitive: `hero2436`
4. File path: `layouts/components/Hero2436.html`

**Generated Registry Format:**
```javascript
// Auto-generated Plenti-compatible component registry
// Signature format: layouts_{category}_{name}_{extension}

const registry = {
  'layouts_components_Hero2436_html': (props) => `...`,
  'layouts_global_nav_html': (props) => `...`,

  // Short-name aliases for backward compatibility
  'Hero2436': (props) => registry['layouts_components_Hero2436_html'](props),
  'nav': (props) => registry['layouts_global_nav_html'](props),
};

export default registry;
```

### Phase 3: Fingerprinting ✅

**Files Created:**
- `builder/fingerprint.go` - Build fingerprint generation
- `renderer/plenti_html.go` - Plenti-compatible HTML rendering
- `core/main.js` - Runtime entry point

### Phase 4: Content Integration ✅

**Files Created:**
- `builder/content_generator.go` - Generates `content.js` from JSON files
- `builder/content_generator_test.go` - Comprehensive tests
- `tests/integration/plenti_content_test.go` - Integration tests

**Content.js Format:**
```javascript
const allContent = [
  {
    type: "pages",
    path: "about",
    filepath: "content/pages/about.json",
    filename: "about.json",
    fields: { ... }
  },
];
export default allContent;
```

### Directory Structure Alignment ✅

Restructured to match Plenti's ejectable core pattern:

```
Before:                          After:
static/js/                       core/
  component-registry.js            main.js
  runtime-components.js            runtime-components.js
                                 generated/
                                   layouts.js
                                   content.js
```

---

## Files Changed

### New Files (7)
| File | Purpose |
|------|---------|
| `types/component.go` | Canonical ComponentTemplate type |
| `builder/fingerprint.go` | Build fingerprint generation |
| `builder/content_generator.go` | Content.js generation |
| `builder/content_generator_test.go` | Content generator tests |
| `renderer/plenti_html.go` | Plenti HTML rendering |
| `core/main.js` | Runtime entry point |
| `core/runtime-components.js` | Alpine.js dynamic component magic |

### Updated Files (5)
| File | Changes |
|------|---------|
| `transformer/components.go` | Path normalization, dual registration |
| `builder/registry_generator.go` | Signature-based keys with aliases |
| `cmd/server/main.go` | Uses `RegisterComponentTemplate()`, new paths |
| `cmd/regenerate_registry/main.go` | Outputs to `generated/layouts.js` |
| `tests/integration/runtime_resolution_test.go` | Updated paths |

### Removed Files (3)
| File | Reason |
|------|--------|
| `static/js/component-registry.js` | Moved to `generated/layouts.js` |
| `static/js/runtime-components.js` | Moved to `core/` |
| `static/js/test-import.mjs` | Test file no longer needed |

---

## Test Results

```
ok  github.com/jimafisk/custom_go_template/analyzer
ok  github.com/jimafisk/custom_go_template/ast
ok  github.com/jimafisk/custom_go_template/builder
ok  github.com/jimafisk/custom_go_template/cmd/server
ok  github.com/jimafisk/custom_go_template/loader
ok  github.com/jimafisk/custom_go_template/parser
ok  github.com/jimafisk/custom_go_template/renderer
ok  github.com/jimafisk/custom_go_template/tests
ok  github.com/jimafisk/custom_go_template/tests/alpine
ok  github.com/jimafisk/custom_go_template/tests/build_time_loop_expansion
ok  github.com/jimafisk/custom_go_template/tests/components
ok  github.com/jimafisk/custom_go_template/tests/integration
ok  github.com/jimafisk/custom_go_template/transformer
ok  github.com/jimafisk/custom_go_template/types
```

**All 14 packages pass.**

---

## Backward Compatibility

| Feature | Status |
|---------|--------|
| Short name lookup (`Hero2436`) | ✅ Works |
| Path lookup (`layouts/components/Hero2436.html`) | ✅ Works |
| Signature lookup (`layouts_components_Hero2436_html`) | ✅ Works |
| Case-insensitive lookup (`hero2436`) | ✅ Works |
| Existing tests | ✅ All pass |

---

## Integration with Plenti

This implementation matches Plenti's patterns:

| Plenti Pattern | Our Implementation |
|----------------|-------------------|
| `allLayouts["layouts_components_" + name + "_svelte"]` | `registry["layouts_components_" + name + "_html"]` |
| `data-content-filepath` attribute | ✅ Supported |
| `content.type` for template selection | ✅ Supported |
| Fingerprinted paths | ✅ Supported |

---

## Next Steps (Future Work)

1. **Tree-Shaking** - Discovery complete, spec pending
   - Potential 84% bundle size reduction per page
   - See: `.agent-os/specs/2026-01-28-tree-shaking-discovery/DISCOVERY.md`

2. **Full Plenti Merge** - When ready to integrate into main Plenti repo

---

## Commit Message Suggestion

```
feat: Plenti-compatible component signatures and registry

- Add types/component.go with canonical ComponentTemplate
- Implement signature generation: layouts_{category}_{name}_{ext}
- Update registry to use signatures as primary keys with short-name aliases
- Restructure to Plenti's ejectable core pattern (core/, generated/)
- Add content.js generation for Plenti content format
- Support multiple lookup strategies (signature, short name, path)
- All 14 test packages pass

This enables future merge of the Go template engine into Plenti.
```
