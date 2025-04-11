# Debugging Guide for Template Engine

Based on your files and the blank screen issue, we need to debug the entire pipeline from parsing through rendering to find the root cause. This guide provides a systematic approach to diagnose and fix issues with your template engine.

## 1. Verify Parser with Alpine.js Support

I've created a specialized test tool (`alpine-parser-test.go`) that focuses specifically on testing Alpine.js attribute parsing. This will help confirm if your parser is correctly handling Alpine.js directives.

To use it:

```bash
# Compile and run the Alpine.js parser test
go run cmd/test_alpine/alpine-parser-test.go
```

This tool will:
- Test various Alpine.js attribute patterns
- Validate that complex Alpine.js syntax like `x-data="{ message: 'Hello Alpine!', count: 0 }"` parses correctly
- Check if the Alpine.js attributes are properly identified in the AST
- Report any issues with detailed error messages

## 2. Debug the Entire Rendering Pipeline

The blank screen suggests that while parsing might succeed, there could be issues in the renderer. I've created a debugging tool (`debugging-renderer.go`) for the entire pipeline.

To use it:

```bash
# Compile and run the renderer debugging tool
go run cmd/debug_renderer/debugging-renderer.go
```

This tool:
- Tests the entire process from parsing to rendering for all template files
- Writes the output markup, scripts, and styles to a debug directory
- Provides detailed logging about what is happening at each step
- Helps identify where the pipeline is breaking down

## 3. Common Issues to Check

### Parsing Phase
- **Alpine.js attributes**: Ensure complex attributes like `x-data` with nested quotes are parsed correctly
- **Component imports**: Check if `import` statements in fence sections are correctly processed
- **Nested structures**: Verify that elements with many children are properly parsed

### Rendering Phase
- **Fence evaluation**: Check if JavaScript in fence sections is properly evaluated
- **Expression replacement**: Ensure expressions like `{name}` are replaced with their values
- **Component rendering**: Verify components are correctly recursively rendered
- **Alpine directive handling**: Check if Alpine directives are preserved in the output HTML

## 4. Main File Issues

Based on your view files, here are specific things to check:

### test_basic.html
- Verify Alpine.js `x-data` attribute on the `<html>` tag is correctly parsed
- Check that Alpine shortcuts like `@click` and `:disabled` are properly handled

### home.html
- Ensure component imports (`Age`, `Head`, etc.) are correctly processed
- Verify that conditional logic (`{if}`) is properly handled
- Check component integration with different prop passing styles

### age.html
- Verify nested conditionals work correctly
- Check component inclusion with prop passing

## 5. Checking the Generated Output

After running the debugging renderer, examine the generated files in the debug_output directory:

1. Compare the original template with the generated HTML to spot missing elements
2. Check if Alpine.js attributes are preserved correctly in the output
3. Verify that expressions like `{name}` are properly replaced with values
4. Ensure all styles and scripts are included correctly

## 6. Integration Testing

If both parsing and rendering seem to work independently, test the complete server:

```bash
# Build and run the server with verbose logging
go build -o server ./cmd/server
LOG_LEVEL=debug ./server
```

Then check the network requests in your browser's DevTools to see:
- If any requests are failing
- If the HTML is being served correctly
- If scripts or styles are missing

## Next Steps

If you identify specific issues using these tools, you can make targeted fixes to the parser or renderer. The most common problems with Alpine.js integration are:

1. Improper handling of complex attribute values
2. Issues with preserving Alpine.js directives in the output HTML
3. Problems with the evaluation of fence JavaScript that processes Alpine-related logic

Let me know what issues you discover, and I can provide specific solutions to address them.
