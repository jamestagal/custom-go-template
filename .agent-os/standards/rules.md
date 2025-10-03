# Rules for Implementing the Plenti AST Transformation Layer

## Crucial Elements for Success

1. **Data Scope Management is Critical**
   - Every variable referenced in templates must be properly initialized in Alpine.js x-data objects
   - Track all variables across template blocks (fence, expressions, conditionals, loops)
   - Create a solid scoping mechanism that handles nested blocks correctly
   - Always validate that variables used in expressions exist in the current scope

2. **Clean AST Transformation**
   - Create a well-defined system for transforming each AST node type
   - Ensure transformations preserve the semantic meaning of the original syntax
   - Design transformations to be composable and reusable
   - Use immutable patterns when transforming nodes to avoid side effects

3. **Component Handling**
   - Implement a robust system for component imports and references
   - Ensure props passing works correctly in all scenarios
   - Handle dynamic components (variables as component names)
   - Create a clean interface between components and parent templates

4. **Testing at Every Step**
   - Write unit tests for each transformation type before implementation
   - Create comprehensive test cases covering edge cases
   - Set up integration tests for the full transformation pipeline
   - Test with real-world templates that use all features

5. **Incremental Development**
   - Start with simple expression transformations and build up complexity
   - Add one feature at a time and fully test before moving on
   - Use feature flags to enable/disable more complex transformations during development
   - Create a minimal working system first, then enhance it

## Additional Implementation Rules

6. **Error Handling Must Be Exceptional**
   - Provide detailed, actionable error messages
   - Include context about what was being transformed
   - Add line/column information where possible
   - Suggest fixes for common errors
   - Fail gracefully and avoid cascading errors

7. **Performance Considerations**
   - Minimize unnecessary string operations
   - Use efficient data structures for scope tracking
   - Consider caching for repeated transformations of the same patterns
   - Benchmark transformation steps to identify bottlenecks

8. **Code Organization**
   - Keep transformation functions small and focused
   - Create separate files for different transformation types
   - Use consistent naming conventions
   - Document function purpose, parameters, and return values

9. **Maintain Compatibility**
   - Follow Alpine.js best practices
   - Test with different Alpine.js versions
   - Ensure output HTML is valid and semantically correct
   - Document any Alpine.js-specific features or limitations

10. **Documentation Driven Development**
    - Document the expected transformation for each syntax element
    - Create examples of input/output pairs for each feature
    - Document any assumptions or special cases
    - Update documentation as implementation progresses

11. **User Experience Focus**
    - Optimize for developer experience and ease of use
    - Ensure transformed templates behave as expected
    - Create helpful debugging tools and logs
    - Document common patterns and anti-patterns

12. **Extensibility**
    - Design for future syntax extensions
    - Create hooks for custom transformations
    - Use interfaces and abstraction where appropriate
    - Document extension points and processes

13. **Security Awareness**
    - Be aware of XSS risks in expression handling
    - Properly encode/escape content as needed
    - Document security considerations
    - Don't evaluate user-provided JavaScript unsafely

14. **Configuration Options**
    - Allow customization of transformation behavior
    - Support different output styles (e.g., minimized, pretty-printed)
    - Provide options for handling edge cases
    - Allow developer overrides for specific transformations

15. **Maintainability**
    - Write clear, self-documenting code
    - Add comments for complex algorithms
    - Create architectural documentation
    - Use consistent patterns throughout the codebase
