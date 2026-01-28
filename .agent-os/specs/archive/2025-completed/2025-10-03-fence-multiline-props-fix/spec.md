# Spec Requirements Document

> Spec: fence-multiline-props-fix
> Created: 2025-10-03
> Status: Planning

## Overview

The fence parser currently fails to correctly parse multi-line prop values in component fence sections. When props contain multi-line arrays or objects, only the first line is captured, resulting in incomplete or broken component data. This bug prevents components from receiving properly structured data for arrays, objects, and complex expressions.

This spec addresses fixing the fence parser to handle multi-line prop values, including arrays, objects, and function expressions that span multiple lines.

## User Stories

**As a** template developer
**I want to** define multi-line arrays and objects in component props
**So that** I can pass complex structured data to components without worrying about formatting

**As a** component author
**I want to** use multi-line prop definitions with proper indentation
**So that** my component prop definitions are readable and maintainable

**As a** developer using the template engine
**I want** prop values to be parsed completely regardless of line breaks
**So that** my components receive the full data structure they expect

## Spec Scope

**In Scope:**
- Fix parsing of multi-line array literals in fence section props
- Fix parsing of multi-line object literals in fence section props
- Handle nested brackets and braces correctly
- Preserve function expressions as strings for Alpine.js runtime evaluation
- Handle proper bracket/brace matching across lines
- Support standard JavaScript literal syntax for arrays and objects
- Update `extractComponentProps()` in `cmd/server/main.go`
- Update fence parsing logic to accumulate multi-line values
- Add comprehensive tests for multi-line prop parsing

**Affected Components:**
- Footer.html (links array, socialLinks array)
- Any component using multi-line prop values
- Future components that need complex prop structures

## Out of Scope

- Parsing of non-standard JavaScript syntax
- Template literal strings (backticks) - can be added in future iteration
- Comment handling within multi-line props (can be added separately)
- Validation of JavaScript syntax within prop values
- Transpilation or transformation of prop values
- Props using spread operators or destructuring

## Expected Deliverable

A fully functional fence parser that:

1. **Correctly parses multi-line arrays:**
   ```
   prop links = [
     { label: "Home", url: "/" },
     { label: "About", url: "/about" }
   ]
   ```

2. **Correctly parses multi-line objects:**
   ```
   prop config = {
     theme: "dark",
     options: {
       nested: true
     }
   }
   ```

3. **Handles function expressions:**
   ```
   prop year = new Date().getFullYear()
   ```

4. **Maintains backward compatibility** with single-line props

5. **Provides accurate error messages** when brackets/braces are mismatched

6. **Includes comprehensive test coverage** for all multi-line scenarios

## Spec Documentation

- Tasks: @.agent-os/specs/2025-10-03-fence-multiline-props-fix/tasks.md
- Technical Specification: @.agent-os/specs/2025-10-03-fence-multiline-props-fix/sub-specs/technical-spec.md
