# Spec Requirements Document

> Spec: component-props-rendering-fix
> Created: 2025-10-03
> Status: Planning

## Overview

Component props with multi-line JavaScript values (arrays, objects) are being truncated during the rendering phase. The fence parser correctly extracts full multi-line values (verified by tests), but when components are rendered and their props are passed to Alpine.js x-data, the values get truncated to just the opening bracket.

Example of the issue:
- **Parsed correctly**: `links = [{ label: "Home", url: "/" }, ...]` (250 chars)
- **Rendered incorrectly**: `"links":"["` (truncated to first line)

This spec addresses fixing the component rendering pipeline to preserve complete multi-line JavaScript prop values.

## User Stories

**As a developer** using components with complex props (arrays of objects, nested structures), **I want** those props to render correctly in the final HTML **so that** Alpine.js receives the complete data structure and components function as expected.

**As a developer** viewing the rendered HTML, **I want** to see the full prop values in the x-data attribute **so that** I can verify the component received the correct data and debug any issues.

**As a user** of a site built with this template engine, **I want** components with complex data (like navigation links) to render correctly **so that** all functionality works as designed.

## Spec Scope

1. **Fix prop value marshaling** to preserve multi-line JavaScript syntax when building Alpine.js x-data
2. **Update parseValue() function** in `cmd/server/main.go` to correctly handle JavaScript object/array syntax with unquoted keys
3. **Ensure Alpine.js x-data receives complete prop values** by tracing the data flow from component AST → transformer → renderer
4. **Add comprehensive tests** to verify multi-line prop values render correctly for arrays, objects, and nested structures
5. **Document the fix** with code comments explaining how JavaScript syntax is preserved

## Out of Scope

1. Changing the fence parser implementation (already fixed and working correctly)
2. Modifying the template syntax for defining props
3. Adding new prop types beyond current JavaScript primitives, arrays, and objects
4. Performance optimization of prop parsing/rendering
5. Changing Alpine.js integration beyond x-data prop passing

## Expected Deliverable

1. Footer component with complex `links` array prop renders with full array visible in browser HTML
2. Inspecting the rendered HTML shows complete prop values in x-data attribute, not truncated
3. All existing tests continue to pass
4. New tests verify multi-line prop rendering for arrays and objects
5. Code changes properly documented with comments explaining the fix

## Spec Documentation

- Tasks: @.agent-os/specs/2025-10-03-component-props-rendering-fix/tasks.md
- Technical Specification: @.agent-os/specs/2025-10-03-component-props-rendering-fix/sub-specs/technical-spec.md
