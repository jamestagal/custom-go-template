# fence-multiline-props-fix - Lite Summary

Fix the fence parser to correctly handle multi-line prop values (arrays, objects, and functions) instead of truncating them at the first line break.

## Key Points
- Multi-line array and object props currently break - only first line is captured
- Need bracket/brace matching logic to detect when prop value spans multiple lines
- Affects Footer.html and any component using complex prop structures
- Must preserve function expressions as strings for Alpine.js runtime evaluation
- Backward compatible with existing single-line props
