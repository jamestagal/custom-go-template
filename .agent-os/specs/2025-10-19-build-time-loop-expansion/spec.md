# Spec Requirements Document

> Spec: Build-Time Loop Expansion
> Created: 2025-10-19

## Overview

Implement build-time loop expansion to fix component name resolution failures in dynamic component templates. Currently, when templates use `{for component in components}` with `<Component:dynamic name={component.name}>`, the loop transformer creates Alpine x-for templates (runtime execution), but the Go transformer tries to resolve `component.name` at build time when the loop variable doesn't exist in dataScope yet, causing resolution to fail. This spec implements Svelte-style build-time loop expansion, where loops are expanded in Go during transformation, adding loop variables to dataScope for each iteration so component names can be successfully resolved.

## User Stories

### Template Author Using Data-Driven Components

As a template author, I want to use `{for component in components}` to iterate over component data from JSON files and render each component using `<Component:dynamic name={component.name} {...component.fields} />`, so that I can create data-driven pages without hardcoding component names.

**Current Problem:** When I write this template, the build fails because `component` doesn't exist at build time - the transformer creates an Alpine x-for template (runtime), but tries to resolve `component.name` before the loop executes.

**Expected Behavior:** The loop should expand at build time (like Svelte does), creating fully rendered HTML for each component in the array, with all component names resolved during the build process.

### Developer Maintaining Template Engine

As a developer maintaining the template engine, I want loop expansion to happen at build time (in Go) instead of runtime (in Alpine.js), so that:
- Component name resolution succeeds (loop variables exist in dataScope)
- Output HTML is fully expanded (no x-for templates)
- Behavior matches Svelte's proven approach
- No complex runtime component registry is needed

**Workflow:** When transforming a `{for}` loop node, the transformer should:
1. Resolve the collection array from dataScope (e.g., `components` from JSON)
2. For each item in the array, create a new iteration scope by cloning dataScope
3. Add the loop variable to the iteration scope (e.g., `component = currentItem`)
4. Transform the loop body nodes with the iteration scope
5. Append transformed nodes to result (fully expanded)

## Spec Scope

1. **Build-Time Loop Expansion** - Modify loop transformer (`transformer/loops.go`) to iterate arrays in Go and expand loop body for each item instead of creating Alpine x-for templates
2. **Loop Variable Scope Management** - Add loop variables to dataScope for each iteration, ensuring component name resolution and property access works correctly
3. **Component Name Resolution** - Enable dynamic component names (like `{component.name}`) to be resolved from loop variables during build-time transformation
4. **Scope Cloning Utility** - Implement safe dataScope cloning to ensure each loop iteration has isolated scope
5. **Testing & Validation** - Verify output matches Svelte behavior (fully expanded HTML) and component resolution succeeds

## Out of Scope

- Runtime loop generation using Alpine x-for templates (this is what we're removing)
- Client-side component registry with JavaScript template literals (not needed with build-time expansion)
- Runtime component resolution (all resolution happens at build time)
- Nested loop optimization (future enhancement - current implementation handles nested loops but may not optimize)
- Loop performance for very large arrays (>100 items) (future enhancement if needed)

## Expected Deliverable

1. **Fully Expanded HTML Output** - Templates using `{for component in components}` should produce fully expanded HTML at build time with no Alpine x-for templates in the output
2. **Component Name Resolution Works** - Component names like `{component.name}` should successfully resolve from loop variables during build-time transformation
3. **Matches Svelte Behavior** - Output should match how Svelte handles `{#each components as component}` - fully expanded, build-time resolution
4. **Passes Integration Tests** - Test cases with component loops from JSON data should render correctly
