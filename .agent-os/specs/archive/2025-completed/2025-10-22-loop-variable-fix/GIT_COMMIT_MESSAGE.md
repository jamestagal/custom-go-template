# Git Commit Message

```
fix: Component registry generator handles loop variables correctly

Previously, the component registry was generating JavaScript template
functions that tried to evaluate loop variables (like `text`, `card`,
`item`) as JavaScript variables during template function execution.
This caused runtime errors because these variables only exist as
Alpine.js loop variables in the DOM.

Changes:
- Detect expressions that reference loop variables using regex
- Convert loop variable text content to Alpine x-text directives
- Convert loop variable attributes to Alpine : binding syntax
- Add comprehensive test coverage for loop variable handling

Examples:
  Before: <p>${text}</p>              // ReferenceError: text is not defined
  After:  <p><span x-text="text"></span></p>  // Alpine evaluates at DOM level

  Before: <img src="${card.icon.src}" />  // ReferenceError: card is not defined
  After:  <img :src="card.icon.src" />    // Alpine binding syntax

New functions:
- expressionReferencesLoopVar() - Detects loop variable references
- attributeReferencesLoopVar() - Checks attributes for loop vars
- extractExpressionFromBraces() - Extracts expression from braces

Test coverage:
- builder/loop_var_test.go - 15+ test cases covering all scenarios
- All builder tests passing

Fixes: whyChoose2425 component and 65+ other components with loops

See: .agent-os/specs/2025-10-22-loop-variable-fix/ for full details
```
