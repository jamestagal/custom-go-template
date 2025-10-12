# Spec Requirements Document

> Spec: Export Let Content Injection System
> Created: 2025-10-11

## Overview

Implement Svelte-compatible `export let` syntax in fence sections to enable JSON content injection into templates, providing a foundation for Plenti integration. This feature allows templates to declare props that should be populated from external content JSON files, matching the familiar `export let` pattern used in Svelte components.

## User Stories

### Template Author - Content-Driven Pages

As a template author, I want to use `export let` to declare props that come from JSON content files, so that I can separate content from presentation and follow familiar Svelte patterns.

**Workflow**: Author creates a template with `export let title, description` in the fence section. When the route handler renders this template, it loads the corresponding JSON file (e.g., `content/pages/about.json`) and injects the `title` and `description` fields as props before rendering. The template uses these props like any other variable: `<h1>{title}</h1>`.

### Developer - Plenti Compatibility

As a developer integrating this engine into Plenti, I want the `export let` syntax to work exactly like Svelte, so that existing Plenti templates can be migrated with minimal changes.

**Workflow**: Developer converts a Plenti Svelte component that uses `export let teachers, courses, students` to the Go template engine syntax. The component continues to receive data from `content/pages/about.json` in the same structure, with the `components` array mapping component names to their field data. The migration is straightforward because the syntax and data flow match Svelte's patterns.

### Content Editor - JSON-Driven Content

As a content editor, I want to edit page content in simple JSON files without touching template code, so that I can update website content safely and easily.

**Workflow**: Editor updates `content/pages/store-demo.json` with new title and description text. The changes appear immediately on the rendered page without any template modifications. The JSON structure is intuitive, with clear field names that match what appears on the page.

## Spec Scope

1. **Export Let Parser** - Parse `export let` declarations in fence sections, extracting prop names and distinguishing them from regular `prop` declarations
2. **Content JSON Loader** - Load JSON files from `content/` directory based on route path, with support for both simple and Plenti-style component structures
3. **Prop Injection System** - Inject JSON data as props into fence section before template rendering, merging with default prop values
4. **Route Handler Integration** - Modify `renderTemplate()` to accept content data and inject it before parsing/transforming
5. **Plenti Structure Support** - Handle Plenti's `components` array format where each component's data is nested under its name

## Out of Scope

- Full Plenti build system integration (data_source.go compilation)
- Client-side content loading or hot-reloading
- Content validation or schema enforcement
- Multi-language/i18n content support
- Content pagination or filtering

## Expected Deliverable

1. Templates using `export let title, description` successfully receive data from matching JSON files (e.g., `content/pages/about.json`)
2. Store-demo page loads title and description from `content/pages/store-demo.json` instead of hardcoded props
3. Both simple flat JSON and Plenti's nested `components` array structures are supported
