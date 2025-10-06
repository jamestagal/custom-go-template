# Future Development Tasks

This document tracks enhancement tasks for future development of the Custom Go Template engine.

## Component Prop Scoping

### Current Limitations

The current implementation uses a global x-data scope workaround which has several limitations:

1. **No component isolation**: All props must exist in the global scope
2. **Naming conflicts**: Component prop names can clash with page variables
3. **No prop validation**: Can't enforce required props or default values properly
4. **Performance**: Can't optimize component re-rendering based on prop changes

### Proposed Enhancement

A proper component prop scoping solution would involve:

- Each component instance gets wrapped with its own `<div x-data="{ ...props }">`
- Prop expressions are evaluated in parent scope and serialized to component scope
- Component default values merge with passed props
- Scoped styles and scripts per component instance

### Tasks

1. **Create spec document for proper component prop scoping with isolation**
   - Document the architectural changes needed
   - Define the transformation pipeline modifications
   - Specify the scope inheritance model
   - Define prop serialization strategy

2. **Implement component prop scoping: Each component instance gets wrapped with own x-data scope**
   - Modify transformer to wrap component instances with isolated x-data
   - Ensure proper Alpine.js initialization order
   - Handle nested component scoping

3. **Implement prop expression evaluation in parent scope with serialization to component scope**
   - Evaluate prop expressions in parent context
   - Serialize evaluated values to component's x-data
   - Handle dynamic prop updates
   - Support both literal values and expressions

4. **Implement component default value merging with passed props**
   - Extract default prop values from component fence section
   - Merge passed props with defaults
   - Handle undefined/null prop values
   - Preserve type information during merge

5. **Add prop validation: enforce required props and default values**
   - Define prop validation syntax in fence section
   - Implement required prop checking at transformation time
   - Validate prop types when specified
   - Provide helpful error messages for missing/invalid props

6. **Optimize component re-rendering based on prop changes**
   - Track which props are used in component template
   - Implement change detection for props
   - Optimize Alpine.js reactivity for component updates
   - Consider memoization for expensive prop transformations

## Component Style Extraction

### Current Limitation

Component scoped styles defined in `<style>` tags within component files are not automatically extracted and included in the parent page output. Currently, styles must be manually copied to the parent page's style section.

### How Svelte/Plenti Handles Component Styles

In Svelte (and by extension, Plenti which uses Svelte), component styles work as follows:

1. **Scoped CSS**: Each component's `<style>` block is automatically scoped to that component
2. **Hash-based Class Names**: Svelte adds unique hash-based class names (e.g., `svelte-1m9hqrl`) to elements
3. **CSS Extraction**: All component styles are extracted and bundled into a single CSS file
4. **Automatic Inclusion**: The bundled CSS is automatically included in the page output
5. **Cascade Behavior**: Component styles from child components are accessible to parent components

**Example from Plenti:**
- Component `progress_circle.svelte` has `<style>` with `#progress { ... }`
- Svelte compiles this to `.svelte-1m9hqrl` classes
- All styles from `progress_circle`, `WebLensPopup`, and `footer` components are bundled
- The footer component (parent) can use styles from its child components without manual importing

### Proposed Enhancement

Implement automatic extraction and inclusion of component scoped styles into parent page output, similar to how Svelte handles it.

### Task

7. **Implement automatic extraction and inclusion of component scoped styles into parent page output**
   - Extract `<style>` blocks from component templates during parsing
   - Generate unique scope identifiers for each component (e.g., hash-based class names)
   - Add scope identifiers to component elements during transformation
   - Transform CSS selectors to include scope identifiers
   - Track which components are used on a page
   - Aggregate all component styles (component + nested component styles)
   - Deduplicate identical style blocks
   - Inject aggregated scoped styles into page `<style>` section or separate CSS file
   - Ensure child component styles are accessible to parent components
   - Handle style precedence and cascade order
   - Consider performance: Cache transformed styles per component

## Priority and Dependencies

### High Priority
- Task 7: Component style extraction (immediate pain point)
- Task 1: Create spec document (foundation for other tasks)

### Medium Priority
- Task 2: Component prop scoping implementation
- Task 4: Default value merging

### Low Priority (depends on Tasks 1-4)
- Task 3: Prop expression evaluation
- Task 5: Prop validation
- Task 6: Re-rendering optimization

## Notes

These enhancements would significantly improve:
- Developer experience (less manual workarounds)
- Component reusability and isolation
- Performance and optimization opportunities
- Type safety and error prevention

The current global scope workaround is functional but limits the scalability and maintainability of complex component hierarchies.
