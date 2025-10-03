# component-props-rendering-fix - Lite Summary

Fix component rendering to preserve multi-line JavaScript prop values (arrays/objects) when passing them to Alpine.js x-data. Currently truncates to opening bracket only, breaking components with complex props.

## Key Points
- Fence parser correctly extracts multi-line prop values (verified by tests)
- Rendering phase truncates multi-line JavaScript to first line/bracket
- Issue occurs in prop value marshaling for Alpine.js x-data attribute
- Root cause: JSON marshaling fails on JavaScript syntax (unquoted keys), causing string truncation in parseValue() function
- Fix required in transformer/components.go and cmd/server/main.go
