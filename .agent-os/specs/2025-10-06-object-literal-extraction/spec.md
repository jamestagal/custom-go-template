# Spec Requirements Document

> Spec: Object Literal Extraction for Component Props
> Created: 2025-10-06
> Status: Planning

## Overview

Enable JavaScript object literal extraction from fence sections to support Plenti's JSON-based content model. Currently, the fence parser extracts object values using JSON marshaling, which produces JSON-formatted strings. This prevents complex data structures (nested objects, arrays) from passing correctly as Alpine.js component props.

This spec implements native JavaScript object literal formatting, allowing developers to pass complex Plenti content objects as component props while maintaining proper data structure and type preservation.

## User Stories

**As a** developer building a Plenti site with reusable components,
**I want to** pass complex data structures (objects with nested properties and arrays) from page-level fence sections to component props,
**so that** I can display rich content data (like user profiles with nested address objects and arrays of interests) without flattening or stringifying the data.

**Acceptance Criteria**:
- Object props maintain their structure (nested objects remain nested)
- Array props preserve array syntax and types
- String values are properly escaped
- Boolean and numeric values maintain their types
- Multiple component instances on the same page can receive different object data

## Spec Scope

### In Scope
- JavaScript object literal formatting functions
  - `formatValueForXData(value)` - Top-level dispatcher for any value type
  - `formatObjectLiteral(obj)` - Format objects with proper key:value syntax
  - `formatArrayLiteral(arr)` - Format arrays with proper element formatting
  - `escapeString(str)` - Handle special characters in strings
- Replace JSON marshaling in x-data generation
  - Page-level x-data building (renderer/render.go)
  - Component x-data building (renderer/component.go)
- Support for nested data structures
  - Objects within objects
  - Arrays within objects
  - Mixed type arrays
  - Objects within arrays
- Type preservation
  - Strings (with quotes)
  - Numbers (no quotes)
  - Booleans (no quotes)
  - Null values
  - Arrays and objects (proper bracket/brace syntax)
- Special character handling
  - Quote escaping in strings
  - Newline and tab characters
  - Unicode characters

### Success Criteria
- UserProfile component displays different data for 3 instances on home.html
- Nested objects (e.g., user.address.city) render correctly
- Arrays (e.g., user.interests) render correctly
- No JSON artifacts (escaped quotes, stringified objects) in rendered HTML
- All existing tests pass
- New tests validate object literal formatting

## Out of Scope

- Modifying parser syntax for fence sections (continues to use JavaScript syntax)
- Changing the Goja JavaScript engine integration
- Altering behavior of simple string/number props (backward compatibility maintained)
- Adding new template syntax features
- Performance optimization beyond basic literal formatting
- Support for JavaScript functions, classes, or symbols in prop values
- Deep cloning or reactive data observation

## Expected Deliverable

### Primary Deliverable
Working implementation where `examples/pages/home.html` displays 3 UserProfile component instances with different data:

```html
<!-- In home.html fence section -->
---
import UserProfile from './components/UserProfile.html'

let user1 = {
  name: "Alice Johnson",
  role: "Software Engineer",
  address: {
    city: "San Francisco",
    state: "CA"
  },
  interests: ["Alpine.js", "Go", "Hiking"]
}

let user2 = {
  name: "Bob Smith",
  role: "Product Manager",
  address: {
    city: "Austin",
    state: "TX"
  },
  interests: ["Leadership", "Strategy"]
}

let user3 = {
  name: "Carol White",
  role: "Designer",
  address: {
    city: "New York",
    state: "NY"
  },
  interests: ["UI/UX", "Typography", "Art"]
}
---

<UserProfile user={user1} />
<UserProfile user={user2} />
<UserProfile user={user3} />
```

Each UserProfile component correctly displays:
- User name and role
- Nested address (city, state)
- Array of interests

### Technical Deliverables
1. Object literal formatting functions in new file `renderer/js_literals.go`
2. Updated x-data generation in `renderer/render.go`
3. Updated x-data generation in `renderer/component.go`
4. Test file `renderer/js_literals_test.go` with comprehensive coverage
5. Integration test in `tests/components/object_props_test.go`
6. Updated `examples/pages/home.html` demonstrating multiple UserProfile instances
7. All existing tests passing

## Spec Documentation

- Tasks: @.agent-os/specs/2025-10-06-object-literal-extraction/tasks.md
- Technical Specification: @.agent-os/specs/2025-10-06-object-literal-extraction/sub-specs/technical-spec.md
