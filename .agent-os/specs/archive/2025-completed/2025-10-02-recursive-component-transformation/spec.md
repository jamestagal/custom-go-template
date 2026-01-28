# Spec Requirements Document

> Spec: Recursive Component Transformation
> Created: 2025-10-02

## Overview

Implement true recursive component transformation in the AST transformer to replace the current placeholder-based approach, enabling components to be fully transformed with their own fence data, prop resolution, and x-data scoping during the AST transformation phase rather than at render time.

## User Stories

### Developer Using Template Components

As a template developer, I want to use components with props and local state, so that I can build reusable, self-contained UI elements that work correctly with Alpine.js reactive data.

When I create a component like `<ProductCard product={item} formatPrice={formatPrice} />`, the transformer should:
1. Look up the registered ProductCard template
2. Extract the component's own fence section (variables, functions, props)
3. Resolve the passed props (`product`, `formatPrice`) from the parent scope
4. Transform the component's body with its own isolated data scope
5. Wrap the result with an x-data attribute containing the component's scope
6. Return fully transformed nodes that can be rendered directly to HTML

This solves the current problem where components are just placeholders (`<div x-component="ProductCard">`) and component-specific functions like `formatPrice` are undefined at runtime.

### Template Engine Developer

As a developer maintaining the template engine, I want the transformation logic to be recursive and composable, so that nested components and complex template structures work correctly without special-case handling.

The transformation should handle:
- Components within conditionals
- Components within loops
- Nested components (components using other components)
- Dynamic component references (`<{componentVar} />`)
- All prop types (static, dynamic, shorthand)

This eliminates the brittle test-specific code currently in `transformComponent()` and makes the system work for real-world use cases.

### End User of Templates

As an end user viewing a rendered template, I want all interactive features and data bindings to work correctly, so that the application behaves as expected.

When a page with components loads:
- All Alpine.js expressions should evaluate without "undefined" errors
- Component-specific functions should be available in their scope
- Props should pass data correctly from parent to child
- Nested components should each have their own isolated reactive state

## Spec Scope

1. **Component Template Registry** - Refactor component registration to ensure parsed ASTs are accessible during transformation with their fence data and props list.

2. **Recursive transformComponent Function** - Implement the full recursive transformation logic described in `docs/RootCauseAnalysis.md`, including component lookup, scope creation, fence data collection, prop resolution, body transformation, and x-data wrapping.

3. **Helper Functions** - Create supporting functions: `collectComponentFenceData()`, `resolvePropValue()`, `addComponentDataWrapper()`, `filterOutFence()`, and improved `parseValue()`.

4. **Scope Management** - Ensure each component instance gets an isolated data scope that includes its own fence variables/functions plus resolved props from the parent.

5. **Test Cleanup** - Remove all test-specific special-case code from `transformComponent()` in `transformer/components.go` (lines 172-272) and `transformer/alpine.go`, ensuring tests pass with the proper implementation.

## Out of Scope

- Function expression handling in `alpineDataFormatter()` (covered in separate spec)
- Loop rendering fixes (covered in separate spec)
- Parser changes or AST node modifications
- Renderer changes (except removing old placeholder rendering code if still present)
- New template syntax features
- Performance optimizations (can be addressed later if needed)

## Expected Deliverable

1. All component-related tests passing: `TestComponentPropsTransformation`, `TestStaticComponentTransformation`, `TestAlpineIntegration/component_integration`, `TestAlpineIntegration/nested_conditionals_and_loops`.

2. Components render with actual transformed content instead of placeholder elements, visible in test output and browser.

3. Alpine.js data scopes properly nested for components, with each component having its own x-data attribute containing its isolated scope.
