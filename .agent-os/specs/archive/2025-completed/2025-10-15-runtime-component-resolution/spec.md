# Spec Requirements Document

> Spec: Runtime Component Resolution for Loop Variables
> Created: 2025-10-15
> Status: Planning

## Overview

Implement runtime component resolution to enable dynamic component iteration in loops where component names are only known at runtime (e.g., `{for component in components} <Component:dynamic name={component.name} /> {/for}`). Currently, the system fails because it attempts to resolve component names at build time when they're loop variables that only exist in Alpine.js runtime scope.

## User Stories

### Content Author Creating Dynamic Pages

As a content author, I want to define a list of components in my JSON content files, so that I can compose pages without modifying template code.

**Workflow:** Author creates `content/pages/homepage.json` with a `components` array listing component names and their props. The template uses `{for component in components} <Component:dynamic name={component.name} {...component.fields} /> {/for}` to render all components automatically. The system detects that `component.name` is a runtime expression, emits an Alpine.js wrapper, and the components render correctly in the browser without build-time errors.

### Developer Building Plenti-Compatible Templates

As a template developer, I want to use Svelte-style component iteration syntax, so that my templates work identically to Plenti's current Svelte implementation.

**Workflow:** Developer writes `layouts/content/pages.html` with `export let components` and uses the loop pattern. The Go engine analyzes the scope, determines that `component` is a loop variable, and generates runtime wrappers instead of attempting build-time resolution. The Alpine.js runtime loads component templates from the JavaScript registry and renders them dynamically.

### Site Builder Running Static Builds

As a site builder, I want component iteration to work during static site generation, so that pre-rendered HTML contains all components with proper SEO and fast first paint.

**Workflow:** Builder runs `plenti build`, which processes all content JSON files. For pages with `components` arrays, the Go engine performs server-side expansion: it iterates the array at build time, looks up each component template, renders the HTML, and outputs static files with all components pre-rendered. No client-side rendering needed for initial page load.

## Spec Scope

1. **Scope Tracking** - Add analyzer to distinguish build-time variables (content props) from runtime-only variables (loop iterators, Alpine stores).
2. **Runtime Wrapper Emission** - Modify transformer to emit Alpine.js-compatible runtime wrappers for components with runtime-only name expressions.
3. **Client-Side Runtime** - Create Alpine.js magic (`$renderDynamicComponent`) and component registry loader for browser-side component rendering.
4. **Registry Code Generation** - Build system to convert Go component templates into JavaScript template literal functions and generate registry manifest.
5. **Server-Side Loop Expansion** - For static builds, pre-expand component loops at build time to render full HTML without client-side rendering.

## Out of Scope

- Deterministic signatures and hydration validation (v4.0 blueprint feature - not needed for basic loop iteration)
- Advanced error recovery and circuit breakers (production hardening - can be added later)
- Component lazy loading and code splitting (optimization - ship all components in single registry for MVP)
- Development overlay and component inspector UI (nice-to-have tooling)
- Props serialization strategies for large data (inline all props for MVP)
- Merge plans for complex spread operator precedence (implement basic left-to-right precedence)

## Expected Deliverable

1. **Browser Test**: Navigate to `http://localhost:3333/` and see both hero2436 and services2437 components rendered from the `components` array in `_index.json` with no console errors.
2. **Build Test**: Run static build and verify all pages with `components` arrays generate complete HTML files with all components pre-rendered.
3. **Scope Test**: Template with `<Component:dynamic name="StaticName" />` still uses build-time resolution (no regression), while `<Component:dynamic name={component.name} />` in a loop uses runtime resolution.

## Spec Documentation

- Tasks: @.agent-os/specs/2025-10-15-runtime-component-resolution/tasks.md
- Technical Specification: @.agent-os/specs/2025-10-15-runtime-component-resolution/sub-specs/technical-spec.md
