# Plenti Build System Analysis

**Source:** `/Users/benjaminwaller/Projects/My Plenti Sites WIP/Plenti` (ejected core)
**Date:** 2026-01-28

---

## Key Discoveries

### 1. Fingerprint Directory Structure

Plenti uses a random 10-character fingerprint (e.g., `aQwupMmCDl`) for cache-busting:

```
public/
├── index.html                      # Root page
├── about/index.html                # /about
├── blog/
│   ├── components/index.html       # /blog/components
│   ├── perry/index.html            # /blog/perry
│   └── stores/index.html           # /blog/stores
├── {fingerprint}/                  # e.g., aQwupMmCDl/
│   ├── bundle.css                  # All CSS combined
│   ├── core/
│   │   ├── main.js                 # Entry point (hydration)
│   │   ├── router.js               # Client-side router
│   │   ├── live-reload.js          # Dev only
│   │   └── cms/                    # CMS components
│   ├── generated/
│   │   ├── content.js              # allContent array
│   │   ├── layouts.js              # Component registry (exports)
│   │   ├── env.js                  # Environment config
│   │   ├── defaults.js             # Content type defaults
│   │   └── schemas.js              # Content schemas
│   ├── layouts/
│   │   ├── components/             # Compiled component JS
│   │   ├── content/                # Compiled content layouts
│   │   ├── global/                 # Compiled global layouts
│   │   └── scripts/                # Compiled scripts
│   └── web_modules/                # Dependencies (navaid, etc.)
├── global.css                      # Not fingerprinted
├── logo.svg                        # Static assets
├── media/                          # Media files
└── robots.txt
```

### 2. HTML Output Pattern

Each page has this structure:

```html
<!doctype html>
<html data-content-filepath=content/pages/about.json lang=en>
<meta charset=utf-8>
<meta name=viewport content="width=device-width,initial-scale=1">
<title>About</title>
<base href=/>
<script type=module src=aQwupMmCDl/core/main.js></script>
<link href="https://fonts.googleapis.com/css2?family=Rubik..." rel=stylesheet>
<link rel=icon type=image/svg+xml href=logo.svg>
<link rel=stylesheet href=global.css>
<link rel=stylesheet href=aQwupMmCDl/bundle.css>
<!-- Pre-rendered content (SSR) -->
<main>...</main>
```

**Key attributes:**
- `data-content-filepath` on `<html>` - Used by main.js for hydration
- Single module script to `{fingerprint}/core/main.js`
- CSS: `global.css` (stable) + `{fingerprint}/bundle.css` (versioned)

### 3. Hydration Flow (main.js)

```javascript
// Minified main.js behavior:
import Router from "./router.js";
import allContent from "../generated/content.js";
import * as allLayouts from "../generated/layouts.js";
import { env } from "../generated/env.js";

// 1. Find content by filepath from HTML attribute
let content = allContent.find(e =>
  e.filepath === document.documentElement.dataset.contentFilepath
);

// 2. Get current path/params
let path = location.pathname;
let params = new URLSearchParams(location.search);

// 3. Dynamic import of layout based on content type
import("../layouts/content/" + content.type + ".js")
  .then(component => {
    let layout = component.default;
    // 4. Hydrate SSR content
    new Router({
      target: document,
      hydrate: true,
      props: { content, layout, allContent, allLayouts, path, params, env }
    });
  });
```

### 4. Content Structure (content.js)

Each entry in `allContent` array:

```javascript
{
  pager: null,              // Pagination info (or page number)
  type: "pages",            // Content type (determines layout)
  path: "about",            // Route path
  filepath: "content/pages/about.json",  // Source file (for hydration match)
  filename: "about.json",   // Just the filename
  fields: {                 // Actual content data
    title: "About Plenti",
    description: ["..."],
    image: "media/perry.webp",
    source: { layout: true, content: true }
  }
}
```

### 5. Layout Registry (layouts.js)

ES module with named exports using signature format:

```javascript
export { default as layouts_components_ball_svelte } from "../layouts/components/ball.js";
export { default as layouts_components_block_svelte } from "../layouts/components/block.js";
export { default as layouts_content_pages_svelte } from "../layouts/content/pages.js";
export { default as layouts_global_html_svelte } from "../layouts/global/html.js";
// ... etc
```

**Signature format:** `layouts_{category}_{name}_svelte`
- `layouts/components/ball.svelte` → `layouts_components_ball_svelte`
- `layouts/content/pages.svelte` → `layouts_content_pages_svelte`

### 6. Environment Config (env.js)

```javascript
export let env = {
  local: true,                          // Dev mode flag
  baseurl: "/",                         // Base URL
  routes: {
    pages: ":filename",
    _index: ":paginate(totalPages)"
  },
  types: ["blog", "pages"],             // Content types
  singleTypes: ["_index"],              // Single-file types
  fingerprint: "UgZcIhSCMY",            // Current fingerprint
  entrypointHTML: "global/html.svelte", // Wrapper template
  entrypointJS: "UgZcIhSCMY",           // JS entry fingerprint
  cms: { ... }                          // CMS config
};
```

### 7. Router Pattern (router.svelte - ejected core)

```svelte
<Html
  {path}
  {params}
  {content}
  {layout}
  {allContent}
  {allLayouts}
  {env}
  {user}
  {shadowContent}
/>

<script>
  import Html from '../layouts/global/html.svelte';
  import Navaid from 'navaid';

  export let content, layout, path, params, allContent, allLayouts, env;
  let shadowContent = {};

  // Set up routes for each content entry
  const router = Navaid('/', handle404);
  allContent.forEach(currentContent => {
    router.on(env.baseurl + currentContent.path, () => {
      import('../layouts/content/' + currentContent.type + '.js')
        .then(component => {
          content = currentContent;
          layout = component.default;
        });
    });
  });
  router.listen();
</script>
```

### 8. Wrapper Template (html.svelte)

```svelte
<script>
  export let content, layout, allContent, allLayouts, env, user, shadowContent;
</script>

<html lang="en">
<Head title={makeTitle(content.filename)} {env} />
<body>
  <main>
    <Nav />
    <div class="container">
      <!-- Dynamic component rendering -->
      <svelte:component this={layout} {...content.fields} {content} {allContent} {allLayouts} {user} />
    </div>
    <Footer {allContent} />
  </main>
</body>
</html>
```

**Key pattern:** `{...content.fields}` spreads all content fields as props to the layout.

### 9. Dynamic Component Resolution (blog.svelte)

```svelte
<script>
  export let components, allLayouts;
</script>

{#if components}
  {#each components as { name }}
    <svelte:component this="{allLayouts["layouts_components_" + name + "_svelte"]}" />
  {/each}
{/if}
```

Uses signature format to look up components from `allLayouts` object.

### 10. Content JSON Patterns

**Pattern 1: Flat fields (pages)**
```json
{
  "title": "About Plenti",
  "description": ["..."],
  "image": "media/perry.webp"
}
```

**Pattern 2: With components array (blog)**
```json
{
  "title": "Dynamic components example",
  "body": "...",
  "components": [
    { "name": "ball" },
    { "name": "block" }
  ]
}
```

---

## Implications for Go Build Pipeline

### Match Plenti Patterns

1. **Fingerprint Generation**
   - Generate 10-char random string (or use timestamp)
   - Store in `env.fingerprint` for consistency

2. **Directory Structure**
   - Match exact structure: `{fingerprint}/core/`, `{fingerprint}/generated/`, etc.
   - Keep `global.css` at root (not fingerprinted)

3. **HTML Output**
   - Add `data-content-filepath` attribute to `<html>` tag
   - Single script tag to `{fingerprint}/core/main.js`
   - CSS links: `global.css` + `{fingerprint}/bundle.css`

4. **Generated Files**
   - `content.js` - allContent array with exact field structure
   - `layouts.js` - Named exports with signature format
   - `env.js` - Environment config with fingerprint

5. **Hydration Support**
   - Pre-render all content (SSR)
   - Client hydrates by matching `data-content-filepath`
   - Layout loaded dynamically based on `content.type`

### Key Differences from Current Go Implementation

| Aspect | Plenti (Svelte) | Current Go | Needed Change |
|--------|-----------------|------------|---------------|
| Fingerprint | 10-char random | 8-char hash | Change to random or keep hash |
| CSS | Single bundle.css | Per-page styles | Bundle all CSS |
| Hydration | `data-content-filepath` | None | Add attribute |
| Router | navaid (client-side) | None (static) | Optional SPA mode |
| Layouts | Compiled to JS modules | Registry functions | Already compatible |

### Build Pipeline Steps

1. **Generate fingerprint** (10-char random or timestamp)
2. **Register components** → Generate `layouts.js`
3. **Scan content** → Generate `content.js`
4. **Generate env.js** with config
5. **For each content entry:**
   - Determine layout from `content.type`
   - Render with wrapper (html.html)
   - Add `data-content-filepath` to output
   - Write to `public/{path}/index.html`
6. **Copy assets:**
   - `styles/*.css` → `public/{fingerprint}/bundle.css` (concatenated)
   - `core/*.js` → `public/{fingerprint}/core/`
   - `generated/*.js` → `public/{fingerprint}/generated/`
   - Static files → `public/`

---

## Files to Reference

| File | Purpose |
|------|---------|
| `core/router.svelte` | Ejected router with hydration pattern |
| `layouts/global/html.svelte` | Wrapper template with dynamic component |
| `generated/content.js` | allContent array structure |
| `generated/layouts.js` | Component registry with signatures |
| `generated/env.js` | Environment config |
| `public/index.html` | Example output HTML |
| `public/aQwupMmCDl/core/main.js` | Hydration entry point |
