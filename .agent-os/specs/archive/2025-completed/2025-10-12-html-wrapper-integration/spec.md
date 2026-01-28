# HTML Wrapper Integration Spec

**Date:** 2025-10-12
**Status:** Planning
**Goal:** Integrate `layouts/global/html.html` as the root wrapper for all pages
**MANDATORY: Use go-backend agent for all Go implementation**
## Problem Statement

Currently, the server renders content layouts directly (e.g., `layouts/content/_index.html`), which means:
- ❌ No global Nav component appears on pages
- ❌ No global Footer component appears on pages
- ❌ No global Head component with consistent metadata
- ❌ Each content layout has to duplicate `<!DOCTYPE html>`, `<html>`, `<body>` tags
- ❌ Not following Plenti's wrapper pattern

## Desired Behavior

Following Plenti's `html.svelte` pattern, we want:
- ✅ **Single root wrapper** (`html.html`) for all pages
- ✅ **Global components** (Nav, Head, Footer) render on every page
- ✅ **Dynamic content injection** via `Component:dynamic`
- ✅ **Content layouts** focus only on page-specific content

## Architecture

### Current Flow (Broken)
```
Request → Server → renderTemplate("layouts/content/_index.html")
                              ↓
                    Parse _index.html (has full HTML doc)
                              ↓
                    Transform → Render → Response
```

**Result:** _index.html is a complete HTML document, no wrapper

### New Flow (Correct)
```
Request → Server → renderTemplateWithWrapper("layouts/global/html.html", "_index")
                              ↓
                    Parse html.html (imports Nav, Head, Footer)
                              ↓
                    Resolve Component:dynamic → inject _index.html
                              ↓
                    Transform → Render → Response
```

**Result:** html.html wraps everything, Nav/Footer render, _index.html is injected as content

## Component Structure

### html.html (Wrapper)
```html
---
import Head from './head.html'
import Nav from './nav.html'
import Footer from './footer.html'
export let content, layout, allContent, allLayouts
---

<!DOCTYPE html>
<html>
  <Head title={content.title} />
  <body>
    <main>
      <Nav />

      <!-- This resolves to layouts/content/_index.html -->
      <Component:dynamic name={layout} {...content.fields} />

      <Footer />
    </main>
  </body>
</html>
```

### _index.html (Content Layout - Simplified)
```html
---
import Hero2436 from '../components/hero2436.html'
export let topper, title, description, buttonText, buttonLink
---

<!-- No more <!DOCTYPE html>, <html>, <body> - just content! -->
<Hero2436
  topper={topper}
  title={title}
  description={description}
  buttonText={buttonText}
  buttonLink={buttonLink} />
```

## Implementation Steps

### Step 1: Update Server Routes

**Before:**
```go
http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    renderTemplate("layouts/content/_index.html", w, r)
})
```

**After:**
```go
http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    renderWithWrapper("_index", w, r)
})
```

### Step 2: Create `renderWithWrapper` Function

```go
func renderWithWrapper(layoutName string, w http.ResponseWriter, r *http.Request) {
    // 1. Load content for this route
    contentData := loadContentWithCache(r.URL.Path)

    // 2. Create props with layout name
    props := map[string]interface{}{
        "layout": layoutName,
        "content": contentData,
        "allContent": getAllContent(),
        "allLayouts": getComponentRegistry(),
    }

    // 3. Render html.html wrapper (which will inject the layout)
    renderTemplateWithProps("layouts/global/html.html", props, w, r)
}
```

### Step 3: Implement Component:dynamic Resolution

The renderer needs to:
1. Detect `<Component:dynamic name={layout} />` syntax
2. Resolve `{layout}` from props (e.g., `layout = "_index"`)
3. Load `layouts/content/_index.html`
4. Parse and transform it
5. Inject it at the Component:dynamic location

### Step 4: Simplify Content Layouts

Remove HTML boilerplate from content layouts:

**Before (_index.html):**
```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Home</title>
</head>
<body>
    <Hero2436 ... />
</body>
</html>
```

**After (_index.html):**
```html
---
import Hero2436 from '../components/hero2436.html'
export let topper, title, description
---

<Hero2436 topper={topper} title={title} description={description} />
```

## Benefits

1. **DRY Principle** - Nav/Footer defined once, used everywhere
2. **Easier Maintenance** - Change Nav once, affects all pages
3. **Consistent Structure** - Every page has same wrapper
4. **Plenti Compatible** - Matches Svelte/Plenti patterns
5. **Cleaner Content Layouts** - Focus on content, not boilerplate

## Testing Plan

1. **Test home page** - Verify Nav, Footer render with _index content
2. **Test multiple pages** - Ensure wrapper works for all routes
3. **Test component imports** - Verify Head, Nav, Footer load correctly
4. **Test props passing** - Ensure content data flows to layouts
5. **Test Component:dynamic** - Verify layout injection works

## Success Criteria

- ✅ Nav component appears on all pages
- ✅ Footer component appears on all pages
- ✅ Head component with consistent metadata
- ✅ Content layouts contain only page-specific content
- ✅ Component:dynamic correctly injects layouts
- ✅ All existing pages continue to work

## Next Steps

1. Implement `renderWithWrapper` function in server
2. Update all route handlers to use wrapper
3. Implement Component:dynamic resolution in renderer
4. Simplify content layouts (remove HTML boilerplate)
5. Test all routes
6. Document new pattern
