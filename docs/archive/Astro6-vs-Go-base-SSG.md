# Astro 6 vs Go-based SSG: A technical architecture comparison

Astro 6's islands architecture and your Go templating approach share the same core philosophy—**ship minimal JavaScript by defaulting to static HTML**—but diverge significantly in implementation. Astro achieves this through compiler-level static analysis and custom elements, while your approach offers more control through explicit component registries and deterministic signatures. The key insight: Astro's patterns provide proven solutions, but a Go-based system with Alpine.js can achieve equivalent or better performance with a simpler runtime footprint.

## How islands architecture compares to your SSG + Alpine.js approach

Astro wraps every hydrated component in an `<astro-island>` custom element that carries all metadata needed for hydration—component URL, renderer path, serialized props, and hydration directive. This **structural embedding** approach eliminates the need for signature matching because the metadata travels with the DOM.

```html
<astro-island uid="1iat0c" component-url="/assets/Counter.js" 
  client="visible" props='{"count":0}'>
  <button>Counter: 0</button>
</astro-island>
```

Your deterministic signature approach is architecturally different. By computing signatures from component name + props hash, you're building a **lookup-based system** that requires synchronization between server and client registries. This is more flexible—you can resolve components by name at runtime—but demands careful signature management.

**The trade-offs are clear**: Astro's approach is zero-configuration but couples you to its build system. Your signature-based approach requires more implementation work but enables patterns Astro cannot support, like truly dynamic component resolution from CMS data where the component name itself is a variable.

For hydration directives, Astro provides five built-in options (`client:load`, `client:idle`, `client:visible`, `client:media`, `client:only`) plus an API for custom directives. Implementing equivalent functionality with Alpine.js is straightforward using IntersectionObserver for visibility-based hydration and requestIdleCallback for idle hydration.

## Astro's dynamic component rendering versus your `<Component:dynamic>` pattern

Astro **cannot** resolve components from string names dynamically due to static analysis requirements for bundling. The standard pattern requires explicit imports with conditional mapping:

```astro
---
import Hero from './blocks/Hero.astro';
import Gallery from './blocks/Gallery.astro';
const componentMap = { 'hero': Hero, 'gallery': Gallery };
---
{blocks.map(block => {
  const Component = componentMap[block.type];
  return Component ? <Component {...block.fields} /> : null;
})}
```

Your `<Component:dynamic name={component.name} {...component.fields} />` pattern represents a significant architectural advantage. Because you control the build system, you can implement true runtime component resolution that Astro's static analysis constraints prevent. However, this requires your build step to analyze which components *might* be rendered on each page to enable tree-shaking.

**Critical limitation in Astro**: Hydration directives (`client:*`) don't work with dynamic tags. If a component reference comes from a variable, Astro cannot statically determine what JavaScript to bundle. Your registry-based approach doesn't have this limitation—you can hydrate any registered component regardless of how it was selected.

## Build-time versus runtime resolution in both systems

Astro separates build-time and runtime through three output modes (`static`, `server`, `hybrid`) and per-route `prerender` exports. For static mode, `getStaticPaths()` must enumerate every dynamic route at build time. For SSR mode, parameters resolve per-request via `Astro.params`.

Your two-mode resolution system (content-known data vs loop variables/stores) maps conceptually but offers finer granularity. In Astro, the division happens at the page level—either a page prerenders or it doesn't. Your approach allows **expression-level resolution**: a single template can mix build-time variables (from content files) with runtime variables (from Alpine stores) in the same component.

Implementation pattern for your Go templates:
```html
<!-- Build-time resolution (Go template) -->
<div x-data="{ items: {{ .Items | jsonify }}, filter: '' }">
  
  <!-- Runtime resolution (Alpine) -->
  <template x-for="item in items">
    <span x-show="!filter || item.category === filter" x-text="item.name"></span>
  </template>
</div>
```

Astro 6 introduces **Server Islands** (`server:defer`) for mixing static shells with dynamic content within a single page. This addresses the same problem your runtime resolution mode solves—showing personalized or frequently-updated content within otherwise static pages.

## Lessons from Content Collections and Live Collections

Astro's Content Collections API provides **schema validation, type safety, and consistent querying** across content sources. Collections are defined with Zod schemas that validate at build time and generate TypeScript types automatically:

```typescript
const blog = defineCollection({
  loader: glob({ pattern: "**/*.md", base: "./src/data/blog" }),
  schema: z.object({
    title: z.string(),
    pubDate: z.coerce.date(),
    draft: z.boolean().optional(),
  })
});
```

For Plenti, this pattern suggests defining a schema layer between your JSON content files and templates. Even without TypeScript, validating content against schemas during build catches errors early and documents the expected data shape.

**Live Collections** (new in Astro 6) fetch data per-request rather than at build time. The API includes explicit error handling (`LiveEntryNotFoundError`, `LiveCollectionValidationError`) and cache hints. This pattern is valuable for inventory levels, user-specific data, or CMS preview functionality. However, Live Collections cannot render MDX or optimize images—those require build-time processing.

For your architecture, Live Collections' key insight is **separating collection configuration** from fetch timing. The same schema and loader interface can support both build-time and request-time fetching, with the resolution mode as a configuration option rather than an API difference.

## CSP implementation and applicability to Go templating

Astro 6 chose **hash-based CSP over nonces** specifically to support static sites. Nonces require rewriting HTML per-request with random values, which conflicts with purely static deployment. Astro's implementation uses `<meta http-equiv="content-security-policy">` elements instead of response headers.

```javascript
export default defineConfig({
  csp: {
    algorithm: 'SHA-256',
    directives: ["default-src 'self'"],
    scriptDirective: {
      resources: ["'self'"],
      strictDynamic: true // for dynamic script injection
    }
  }
});
```

The build process automatically generates hashes for all scripts and styles, including dynamically loaded ones, and deduplicates resources. Runtime APIs (`Astro.csp.insertScriptHash()`) allow per-page customization.

**For your Go SSG**: Implement hash-based CSP by computing SHA-256 hashes of inline scripts during build and generating the CSP meta tag per page. Since Alpine.js uses inline directives, you'd hash the Alpine initialization scripts. With strict-dynamic enabled, dynamically injected scripts from Alpine components are automatically trusted if they originate from a trusted script.

## Trade-offs: established framework versus custom Go solution

| Factor | Astro 6 | Custom Go + Alpine.js |
|--------|---------|----------------------|
| **Build complexity** | High (Vite, esbuild, Rollup) | Lower (Go templates, simple bundler) |
| **Runtime overhead** | Custom elements + framework runtimes | Alpine.js only (~15KB gzipped) |
| **Dynamic components** | Requires static import mapping | True runtime resolution possible |
| **Hydration control** | 5 directives + custom API | Full implementation control |
| **Multi-framework** | React, Vue, Svelte, Solid | Alpine.js only (simpler) |
| **Ecosystem** | Mature, extensive integrations | DIY, but Plenti-specific |
| **Dev experience** | Hot reload, error overlays via Vite | Must build or adapt tooling |

The fundamental trade-off is **ecosystem leverage versus architectural control**. Astro provides battle-tested solutions but constrains you to its patterns. Your Go approach requires more implementation but enables optimizations Astro cannot—like a unified Go/JavaScript type system or tighter CMS integration.

Performance parity is achievable. Alpine.js (~15KB) is significantly smaller than React (44KB) or Vue (34KB) runtimes. Your total client bundle can be smaller than equivalent Astro functionality while maintaining full interactivity.

## Vite Environment API: relevance for dev server parity

Astro 6's headline feature addresses **dev/prod environment divergence** using Vite's Environment API. Previously, code that worked locally could behave differently after deployment because dev and prod ran in different JavaScript runtimes with different globals.

The Environment API enables Astro to run applications inside the same runtime deployed to production—including Cloudflare's workerd runtime locally with full access to KV, Durable Objects, and R2. This eliminates the need for simulation APIs.

**For Plenti's dev server**, the principle matters more than the specific implementation. If your production environment has specific constraints (edge runtime, specific Go version), consider running those same constraints during development. For pure static sites deployed to CDNs, this is less critical since there's no runtime variation.

One applicable pattern: Vite's Environment API manages separate module graphs for client, server, and prerender environments in a single process. If Plenti needs to coordinate build-time template rendering with client-side Alpine component bundling, a unified dev server managing both could improve the development experience.

## Architectural patterns worth adopting

**1. Component hydration markers with all metadata embedded**—Astro's `<astro-island>` pattern demonstrates that carrying component-url, props, and hydration directive together simplifies the runtime. Your equivalent:

```html
<div data-hydrate="visible" 
     data-component="counter" 
     data-props='{"count":0}'
     data-sig="a1b2c3">
  {{ template "counter" . }}
</div>
```

**2. Explicit loader interface for content sources**—Content Collections' loader abstraction (`loadCollection`, `loadEntry`) provides a clean API for multiple data sources (files, CMS, API) with consistent querying. Consider:

```go
type ContentLoader interface {
    LoadCollection(name string, filter func(Entry) bool) ([]Entry, error)
    LoadEntry(collection, id string) (*Entry, error)
}
```

**3. Hash-based CSP for static deployments**—Generating script hashes at build time enables security features without server-side processing.

**4. Per-page component manifests**—Build-time analysis should produce a minimal set of components per page, similar to how Astro's code splitting generates page-specific bundles.

**5. Schema validation at build time**—Validating content against explicit schemas catches errors early and documents data contracts between content and templates.

## Conclusion

Astro 6 represents the state of the art in JavaScript-based SSG architecture, with sophisticated solutions for partial hydration, content management, and environment parity. However, its reliance on static analysis creates constraints your Go-based approach can sidestep.

Your architecture's key advantages are **true runtime component resolution** (vs Astro's map-based workarounds), **simpler runtime** (Alpine.js vs framework adapters), and **unified Go tooling**. The trade-off is building infrastructure that Astro provides out-of-box.

Worth borrowing from Astro 6: the content loader abstraction, hash-based CSP approach, hydration directive patterns, and the principle of embedding all hydration metadata directly in the DOM. Worth avoiding: Astro's constraint that dynamic component tags can't use hydration directives—your registry system should explicitly support this.

The ideal outcome combines Astro's proven patterns with your architecture's flexibility: schema-validated content collections, deterministic hydration signatures, per-page tree-shaking, and true dynamic component resolution—all running on Go's fast build times with Alpine's minimal runtime.