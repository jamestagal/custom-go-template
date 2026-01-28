# Object Literal Extraction for Component Props - Lite Summary

Enable fence parser to extract JavaScript objects as structured data instead of JSON strings, allowing complex Plenti content objects to pass as component props while preserving nested structures, arrays, and proper type formatting for Alpine.js x-data initialization.

## Key Points
- Replace JSON marshaling with JavaScript literal formatting in x-data generation
- Support nested objects, arrays, and mixed types with proper escaping
- Maintain backward compatibility with simple string/number props
