# Spec Requirements Document

> Spec: Function Expression Handling in Alpine Data Formatter
> Created: 2025-10-02

## Overview

Fix the `alpineDataFormatter()` function to correctly output JavaScript function definitions without quotes in x-data attributes, enabling Alpine.js to properly recognize and execute component methods and event handlers.

## User Stories

### Template Developer Using Functions

As a template developer, I want to define functions in my component's fence section, so that I can create interactive Alpine.js components with methods and event handlers.

When I write:
```
---
let count = 0
function increment() {
  return count++
}
---
<button @click="increment">Click me</button>
```

The transformer should generate:
```html
<div x-data="{ count: 0, increment: function() { return count++ } }">
  <button @click="increment">Click me</button>
</div>
```

NOT:
```html
<div x-data="{ count: 0, increment: &quot;function() { return count++ }&quot; }">
```

This solves the problem where functions are stringified and Alpine.js cannot execute them.

### Component with Multiple Function Types

As a developer, I want all function syntax variations to work correctly, so that I can use modern JavaScript patterns in my templates.

The formatter should correctly handle:
- Function declarations: `function name() {}`
- Function expressions: `const name = function() {}`
- Arrow functions: `const name = () => {}`
- Method shorthand: `name() {}`
- Async functions: `async function name() {}`
- Getters/setters: `get name() {}`

All should be output without quotes in the x-data object.

## Spec Scope

1. **Improve isFunctionExpression()** - Enhance the function detection logic to correctly identify all JavaScript function syntax patterns.

2. **Fix formatGoValueToJS()** - Ensure this helper outputs function strings without wrapping quotes.

3. **Update alpineDataFormatter()** - Use the fixed `formatGoValueToJS()` helper instead of `json.Marshal()` for generating x-data object strings.

4. **Preserve Complex JS Objects** - Ensure objects with methods, getters, and Alpine.js magic properties are preserved correctly.

## Out of Scope

- Component transformation logic (covered in separate spec)
- Loop rendering (covered in separate spec)
- Parser changes for fence section parsing
- New function syntax support beyond standard JavaScript
- Optimization of function detection performance

## Expected Deliverable

1. Test `TestAlpineDataWrapper/Function_Expressions` passes successfully.

2. x-data attributes in rendered HTML contain executable JavaScript functions without quote escaping.

3. All existing Alpine.js integration tests continue to pass.
