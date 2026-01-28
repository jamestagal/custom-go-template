# Known Issue: Component Registry Syntax Error in Alpine Directives

**Status**: 🔴 PARSER/BUILDER BUG - Complex expressions still failing
**Priority**: HIGH (blocking component rendering)
**Date**: 2025-10-16 (Updated: 2025-10-16 21:50 UTC)
**Root Cause**: Multiple architectural issues in expression conversion pipeline

**📚 Comprehensive Documentation**: See [TROUBLESHOOTING_HISTORY.md](../.agent-os/specs/2025-10-16-component-registry-debugging/TROUBLESHOOTING_HISTORY.md) for full troubleshooting timeline and [ERROR_REFERENCE.md](../.agent-os/specs/2025-10-16-component-registry-debugging/ERROR_REFERENCE.md) for current error details.

---

## UPDATE 2025-10-16: Partial Fixes Implemented, Complex Expressions Still Broken

### What's Been Fixed ✅

1. **Parser Quote Handling** (2025-10-16):
   - Fixed `parseComplexAlpineValue()` to check closing quote FIRST
   - x-data attributes now correctly extracted as complete strings
   - File: `parser/html.go` lines 584-620

2. **Builder Expression Conversion** (2025-10-16):
   - Added `convertAttributeExpressions()` function with regex pattern
   - Simple expressions now convert correctly: `{count}` → `${props.count}`
   - File: `builder/registry_generator.go`

3. **Component Registration** (2025-10-16):
   - Fixed case sensitivity: `pages` → `Pages`
   - All 65 components successfully registered
   - Runtime wrapper elements present in HTML

### What's Still Broken 🔴

**Current Error** (BLOCKING):
```
runtime-components.js:97 Failed to load component registry after 3 attempts
SyntaxError: Unexpected token '(' (at component-registry.js:1793:83)
```

**Problematic Code** (Line 1793):
```javascript
${props.(start * 1) + index + 1}  // ❌ INVALID - cannot have ( after props.
```

**Root Cause**: Regex-based conversion adds `props.` prefix to entire expression including parentheses, producing invalid JavaScript syntax.

**Expected Output**:
```javascript
${(props.start * 1) + index + 1}  // ✅ VALID
```

---

## Problem Statement

When components contain x-data attributes with object literal expressions like `{ count: {count} }`, the **parser and transformer** incorrectly handle the expressions, producing corrupted HTML that cannot be rendered by the component registry.

### Example

**Component Template**:
```html
---
prop count = 0
prop message = "Hello"
---
<div x-data="{ count: {count}, message: '{message}' }">
  <p>Count: <span x-text="count"></span></p>
</div>
```

**Current Output (BROKEN)**:
```javascript
'ComponentName': (props) => `
  div x-data="<span x-text="count: {count}, message: '{message}'"></span>">
    <p>Count: <span x-text="count"></span></p>
  </div>
`
```

**Issues**:
1. `<div` opening bracket is missing
2. Attribute value contains `<span x-text="">` elements (invalid HTML)
3. Results in: `SyntaxError: Missing } in template expression`

---

## Root Cause Analysis

### The Bug is NOT in the Registry Generator

Investigation revealed the bug occurs **BEFORE** the registry generator runs:

1. **Parser Phase** (`parser/` package):
   - When parsing `<div x-data="{ count: {count} }">`, the parser encounters `{count}` inside the attribute value
   - Instead of keeping it as part of the attribute value string, it creates an **ExpressionNode** as a child of the `<div>` element
   - This is architecturally incorrect - attribute values should be strings, not node arrays

2. **Transformer Phase** (`transformer/transformer.go` lines 137-163):
   - The transformer has a blanket rule: ALL ExpressionNodes → `<span x-text="">`
   - It doesn't know (and can't easily know) that these ExpressionNodes came from attribute values
   - So it converts `{count}` to `<span x-text="count"></span>` as an element child
   - This corrupts the HTML structure

3. **Registry Generator Phase** (`builder/registry_generator.go`):
   - Receives already-corrupted AST
   - Renders the corrupted structure faithfully
   - Output is syntactically broken

### Code Evidence

**Transformer Code** (transformer/transformer.go:137-163):
```go
case *ast.ExpressionNode:
	// Transform expression nodes
	log.Printf("transformNodes: Transforming Expression node")
	// Clean the expression by removing any extra curly braces
	cleanedExpr := n.Expression
	cleanedExpr = strings.TrimPrefix(cleanedExpr, "{")
	cleanedExpr = strings.TrimSuffix(cleanedExpr, "}")
	cleanedExpr = strings.TrimSpace(cleanedExpr)

	// Add variables from the expression to the data scope
	extractVariablesFromExpr(cleanedExpr, dataScope)

	// Create an Alpine.js x-text element
	// THIS IS THE PROBLEM: Converts ALL expressions to elements
	transformedNodes = append(transformedNodes, &ast.Element{
		TagName: "span",
		Attributes: []ast.Attribute{
			{
				Name:       "x-text",
				Value:      cleanedExpr,
				Dynamic:    true,
				IsAlpine:   true,
				AlpineType: "text",
			},
		},
		Children:    []ast.Node{},
		SelfClosing: false,
	})
```

This code assumes ALL ExpressionNodes should become `<span>` elements, which is only true for expressions in text content, NOT for expressions in attribute values.

---

## Impact Assessment

### ✅ What Still Works:
- Components without expressions in Alpine attributes
- Components with simple attribute bindings (no `{expression}` syntax in values)
- Components with x-text, x-show, x-if on separate elements
- **Most components (60+) work fine**

### ❌ What's Broken:
- Components with x-data containing object literals with `{expression}` syntax
- Any Alpine directive attribute that contains `{expression}` patterns
- Affects **1-2 test components** (script tag preservation tests)
- Component registry fails to load: `Failed to load component registry after 3 attempts`

### 🎯 Workarounds:
Users can rewrite components to avoid the problematic pattern:

**Option 1: Use x-init**:
```html
<!-- Instead of this -->
<div x-data="{ count: {count} }">

<!-- Use this -->
<div x-data="{ count: 0 }" x-init="count = {count}">
```

**Option 2: Define data in fence section**:
```html
---
prop count = 0
---
<div x-data="{ count }">  <!-- No template expression needed -->
```

**Option 3: Avoid object literals**:
```html
<!-- Instead of inline objects -->
<div x-data="{ count: {count}, message: '{message}' }">

<!-- Use function that returns object -->
---
prop count = 0
prop message = "Hello"

function getData() {
  return {
    count: count,
    message: message
  };
}
---
<div x-data="getData()">
```

---

## Proper Fix Required

The fix requires **architectural changes** to the parser and/or transformer:

### Option 1: Parser Fix (RECOMMENDED)
Modify the parser to NOT create ExpressionNodes for expressions inside attribute values:

```go
// In parser/html.go or parser/attributes.go
func parseAttributeValue(...) {
	// When parsing attribute value, DON'T call expression parser
	// Just consume the string including {expression} patterns
	// Store the raw string in Attribute.Value

	// The registry generator or transformer can then detect {expr}
	// patterns in Alpine attributes and handle them appropriately
}
```

**Cognitive Load**: 15 (parser modification: 10, integration: 5)

### Option 2: Transformer Context Tracking
Add context tracking to know when expressions are part of attributes:

```go
type TransformContext struct {
	inAttributeValue bool
	currentAttribute string
}

// In transformer/transformer.go
case *ast.ExpressionNode:
	if ctx.inAttributeValue {
		// DON'T convert to <span>
		// Leave as part of attribute value
	} else {
		// Convert to <span x-text="">
	}
```

**Issue**: Attributes don't currently contain ExpressionNodes - they're element children. This would require refactoring attribute values to be `[]ast.Node` instead of `string`.

**Cognitive Load**: 20 (AST refactor: 12, transformer update: 8)

### Option 3: Attribute Value Node Arrays (LONG TERM)
Refactor the AST so attribute values can contain node arrays:

```go
type Attribute struct {
	Name       string
	Value      string    // For simple values
	ValueNodes []ast.Node // For values with expressions
	// ... other fields
}
```

This would properly represent the structure and make context-aware transformation possible.

**Cognitive Load**: 25 (AST changes: 10, parser updates: 10, transformer updates: 5)

---

## What Was Fixed (Partial)

### Registry Generator Improvements
File: `builder/registry_generator.go`

Added context tracking infrastructure (ready for upstream fix):

```go
// RenderContext tracks rendering context to handle expressions differently in different contexts
type RenderContext struct {
	insideLiteral    bool // Inside <style> or <script> tags
	insideAlpineAttr bool // Inside Alpine.js directive (READY FOR FUTURE USE)
}

case *ast.ExpressionNode:
	// CRITICAL FIX: Context-aware expression conversion
	if ctx.insideLiteral {
		// Inside style/script tags - preserve as-is ✓ WORKING
		escaped := escapeTemplateLiteral(n.Expression)
		sb.WriteString(escaped)
	} else if ctx.insideAlpineAttr {
		// Inside Alpine directive attribute - preserve {expression} syntax
		// NOTE: This code path is ready but won't be reached until parser/transformer fixed
		sb.WriteString("{")
		sb.WriteString(n.Expression)
		sb.WriteString("}")
	} else {
		// Normal content - transform {variable} to ${props.variable}
		sb.WriteString("${props.")
		sb.WriteString(n.Expression)
		sb.WriteString("}")
	}
```

**Status**: The `insideAlpineAttr` logic is implemented but won't be used until the parser/transformer is fixed to not create ExpressionNode children for attribute values.

---

## Recommended Action Plan

### Immediate (For Users):
1. Document the workaround in user-facing docs
2. Add validation warnings when detecting `{expression}` patterns in x-data attributes
3. Provide migration guide for affected components

### Short Term (For go-backend Agent):
1. **Option 1 (Parser Fix)** is recommended as it's the cleanest solution
2. Modify parser to keep attribute values as strings, don't create ExpressionNodes
3. Update transformer or registry generator to handle `{expr}` patterns in Alpine attribute strings
4. Cognitive Load: ~15 (manageable)

### Long Term:
1. Consider refactoring attribute values to support node arrays (Option 3)
2. This enables proper mixed content in attributes (text + expressions)
3. More flexible but higher complexity

---

## Test Cases Required

When implementing the fix, these tests MUST pass:

```go
func TestParserAttributeExpressions(t *testing.T) {
	input := `<div x-data="{ count: {count} }">`
	ast := parser.ParseTemplate(input)

	// Attribute value should be a STRING, not contain ExpressionNodes
	div := ast.RootNodes[0].(*ast.Element)
	assert.Equal(t, "{ count: {count} }", div.Attributes[0].Value)

	// Element should have ZERO children (no ExpressionNode children)
	assert.Len(t, div.Children, 0)
}

func TestTransformerAttributeExpressions(t *testing.T) {
	ast := createASTWithAttributeExpression() // Helper to create proper AST
	transformed := transformer.TransformAST(ast, props)

	// After transformation, attribute value should still be intact
	// (Possibly with {count} converted to something else, but NOT to child elements)
	div := transformed.RootNodes[0].(*ast.Element)
	assert.Contains(t, div.Attributes[0].Value, "count")

	// Should NOT have <span> children from attribute expressions
	assert.Len(t, div.Children, 0)
}

func TestRegistryGeneratorAlpineAttributes(t *testing.T) {
	// After parser/transformer fix, this should generate valid JS
	component := ComponentTemplate{
		Name: "TestComp",
		AST: /* properly transformed AST */,
	}

	result := GenerateComponentRegistry([]ComponentTemplate{component})

	// Should have valid HTML
	assert.Contains(t, result, `<div x-data="`)

	// Should NOT have <span> inside attribute
	assert.NotContains(t, result, `x-data="<span`)

	// Should be valid JavaScript (would need JS parser to fully validate)
	assert.NotContains(t, result, `${props. count:`) // The bug signature
}
```

---

## Error Messages

```
runtime-components.js:97 Failed to load component registry after 3 attempts
SyntaxError: Missing } in template expression (at component-registry.js:61:17)
```

```
runtime-components.js:138 Failed to load component registry:
Failed to load component registry after 3 attempts: Missing } in template expression
```

---

## Files Involved

- `parser/` - **Root cause**: Creates ExpressionNodes for attribute values
- `transformer/transformer.go` (lines 137-163) - **Root cause**: Converts ALL ExpressionNodes to elements
- `builder/registry_generator.go` - Partially fixed (context tracking infrastructure ready)
- `builder/registry_generator_test.go` - Tests passing for simple cases
- `static/js/component-registry.js` - Generated output (has syntax error)
- `layouts/content/script-test.html` - Example problematic component

---

## Why This is Medium Priority

1. **Main functionality works** - Static rendering, page loading, Alpine.js all work
2. **Runtime component resolution works** - Dynamic components render correctly
3. **Affects edge cases** - Only components with specific x-data patterns break
4. **Easy workaround** - Components can be rewritten to avoid the pattern
5. **Proper fix is complex** - Requires parser/transformer architectural changes

---

## Conclusion

This is a **parser and transformer architectural issue**, not a simple registry generator bug. The registry generator has been partially fixed (context tracking infrastructure added), but the full solution requires upstream changes to how the parser handles attribute values with expressions.

**For now**: Users should avoid using `{expression}` syntax inside Alpine directive attributes. Use the workarounds documented above.

**For later**: Implement Option 1 (Parser Fix) to properly handle attribute expressions without creating ExpressionNode children.

---

## Cognitive Load for Full Fix

**Investigation**: 8 (understand rendering pipeline) ✓ COMPLETE
**Registry Generator Fix**: 12 (add context tracking) ✓ COMPLETE
**Parser Fix (Option 1)**: 15 (modify attribute parsing)
**Testing**: 8 (test various Alpine directives and expressions)
**Total Remaining**: 23 (manageable for go-backend agent)

---

## Troubleshooting History (2025-10-16)

**Full documentation**: See [TROUBLESHOOTING_HISTORY.md](../../2025-10-16-component-registry-debugging/TROUBLESHOOTING_HISTORY.md)

### Attempts Made:

1. ✅ **Fixed parser quote handling** - parseComplexAlpineValue() now works correctly
2. ✅ **Added builder expression conversion** - Simple expressions convert properly
3. ❌ **Complex expressions still fail** - Regex cannot parse expression trees
4. ✅ **Component registration fixed** - All 65 components registered
5. ✅ **Script loading path fixed** - runtime-components.js loads correctly
6. ⚠️ **'this.' prefix bug persists** - Separate issue, still investigating

### Current Blocker:

Line 1793 in component-registry.js:
```javascript
${props.(start * 1) + index + 1}  // ❌ Invalid syntax
```

**Problem**: The regex pattern `\{([a-zA-Z_$][\w.$]*(?:\[[^\]]+\])?(?:\([^)]*\))?)\}` matches the entire expression `(start * 1)` including parentheses, then adds `props.` prefix before it, resulting in invalid `props.(...)` syntax.

**Solutions Documented**:
- **Option 1 (Quick)**: Improve regex to match individual identifiers and add skip list for loop variables
- **Option 2 (Proper)**: Refactor attributes to support AST node arrays instead of strings
- **Option 3 (Hybrid)**: Add Alpine directive flags + improved regex

See [CURRENT_STATUS.md](../../2025-10-16-component-registry-debugging/CURRENT_STATUS.md) for decision matrix.

---

## Next Steps

### Immediate (Today)
1. ✅ **Document workaround** in user-facing documentation
2. ✅ **Document all troubleshooting attempts** (TROUBLESHOOTING_HISTORY.md created)
3. ⏭️ **Implement Option 1 (improved regex)** as quick fix
4. ⏭️ **Test with hero2436 and services2437 components**

### Short Term (This Week)
1. **Decide on fix approach** - Option 1 (quick) vs Option 2 (proper)
2. **Implement parser/builder changes** based on decision
3. **Add comprehensive tests** for all expression patterns
4. **Fix 'this.' prefix bug** (separate investigation needed)

### Long Term (Next Sprint)
1. **Refactor attribute values** to support node arrays (Option 2)
2. **Add scope tracking** for Alpine directives and loop variables
3. **Performance optimization** of registry generation
4. **Add regression tests** to prevent future issues

---

## Related Documentation

- **[TROUBLESHOOTING_HISTORY.md](../../2025-10-16-component-registry-debugging/TROUBLESHOOTING_HISTORY.md)** - Complete troubleshooting timeline with all attempted fixes
- **[ERROR_REFERENCE.md](../../2025-10-16-component-registry-debugging/ERROR_REFERENCE.md)** - Quick reference for current errors and patterns
- **[CURRENT_STATUS.md](../../2025-10-16-component-registry-debugging/CURRENT_STATUS.md)** - Summary of what's working/broken and decision points

---

**Last Updated**: 2025-10-16 21:50 UTC
**Status**: 🔴 BLOCKING - Complex expressions prevent component registry from loading
**Next Action**: Implement improved regex pattern (Option 1) to unblock component rendering
