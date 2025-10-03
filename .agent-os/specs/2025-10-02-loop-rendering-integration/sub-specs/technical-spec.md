# Technical Specification

This is the technical specification for the spec detailed in @.agent-os/specs/2025-10-02-loop-rendering-integration/spec.md

## Technical Requirements

### 1. Investigation Phase

**Location**: `transformer/loops.go`

**Current Function**: `transformLoop(node *ast.Loop, dataScope map[string]any) []ast.Node`

**Investigation Steps**:

1. **Review Current Implementation**:
   - Read the full `transformLoop()` function
   - Identify how it handles `node.Iterator` and `node.Collection`
   - Check how it formats the x-for expression
   - Verify how it transforms loop body content

2. **Analyze Failing Tests**:
   ```bash
   go test ./tests/alpine -v -run TestAlpineIntegration/loop_rendering 2>&1 | tee loop_test_output.txt
   go test ./tests/alpine -v -run TestAlpineIntegration/nested_conditionals_and_loops 2>&1 | tee nested_test_output.txt
   ```

3. **Compare Expected vs Actual Output**:
   - Document what the tests expect
   - Document what's actually being generated
   - Identify the specific discrepancy

4. **Check Test Files**:
   - `tests/alpine/alpine_integration_test.go`
   - `tests/alpine/loops_test.go`
   - Look at test expectations for loop syntax

### 2. Common Loop Issues to Check

**Issue 1: x-for Expression Format**

Alpine.js expects: `x-for="item in items"`

**Check**:
- Is the expression being formatted correctly?
- Are there extra spaces or quotes?
- Is the iterator name preserved correctly?

**Issue 2: Loop Body Transformation**

**Check**:
- Are expressions inside the loop body transformed with `x-text`?
- Are child elements being transformed recursively?
- Is the loop body wrapped correctly in the template element?

**Issue 3: Collection Variable in Scope**

**Check**:
- Is `node.Collection` being added to `dataScope`?
- Is it using `extractVariablesFromExpr()` correctly?
- Are nested property accesses handled (e.g., `user.posts`)?

**Issue 4: Iterator Variable Scope**

**Problem**: Iterator variables should be local to the loop, not added to parent scope

**Check**:
```go
// Current code might be doing this (WRONG):
extractVariablesFromExpr(node.Iterator, dataScope)

// Should NOT add iterator to parent scope - Alpine handles this
```

### 3. Expected Loop Transformation

**Input AST**:
```go
&ast.Loop{
    Iterator: "item",
    Collection: "items",
    Content: []ast.Node{
        &ast.Element{
            TagName: "div",
            Children: []ast.Node{
                &ast.ExpressionNode{Expression: "item.name"},
            },
        },
    },
}
```

**Expected Output**:
```go
[]ast.Node{
    &ast.Element{
        TagName: "template",
        Attributes: []ast.Attribute{
            {
                Name: "x-for",
                Value: "item in items",
                Dynamic: true,
                IsAlpine: true,
                AlpineType: "for",
            },
        },
        Children: []ast.Node{
            &ast.Element{
                TagName: "div",
                Children: []ast.Node{
                    &ast.Element{
                        TagName: "span",
                        Attributes: []ast.Attribute{
                            {Name: "x-text", Value: "item.name", ...},
                        },
                    },
                },
            },
        },
    },
}
```

### 4. Scope Handling for Loops

**Requirement**: Create child scope for loop body that includes iterator variable

**Implementation**:

```go
func transformLoop(node *ast.Loop, dataScope map[string]any) []ast.Node {
    // Add collection to parent scope
    extractVariablesFromExpr(node.Collection, dataScope)

    // Create child scope for loop body
    loopBodyScope := CreateChildScope(dataScope)

    // Add iterator to loop body scope (but NOT parent scope)
    // This makes the iterator available for expressions inside the loop
    loopBodyScope[node.Iterator] = nil // Placeholder value

    // Format x-for expression
    xForExpr := fmt.Sprintf("%s in %s", node.Iterator, node.Collection)

    // Transform loop body with child scope
    transformedContent := transformNodes(node.Content, loopBodyScope, false)

    // Create template element
    templateElement := &ast.Element{
        TagName: "template",
        Attributes: []ast.Attribute{
            {
                Name:       "x-for",
                Value:      xForExpr,
                Dynamic:    true,
                IsAlpine:   true,
                AlpineType: "for",
            },
        },
        Children: transformedContent,
    }

    return []ast.Node{templateElement}
}
```

### 5. Nested Loop Handling

**Input**:
```
{for category in categories}
  {for item in category.items}
    <div>{item.name}</div>
  {/for}
{/for}
```

**Requirements**:
- Each loop should have its own scope
- Inner loop's iterator (`item`) should not conflict with outer scope
- Inner loop's collection (`category.items`) should resolve in outer loop's scope (where `category` is available)

**Implementation Check**:
```go
// Outer loop creates scope with 'category'
outerScope := CreateChildScope(dataScope)
outerScope["category"] = nil

// Transform outer loop content (which includes inner loop)
// Inner loop will create its own scope from outerScope
innerScope := CreateChildScope(outerScope)
innerScope["item"] = nil  // Inner iterator
// 'category' is still accessible from outerScope
```

### 6. Loops in Conditionals

**Input**:
```
{if showItems}
  {for item in items}
    <div>{item.name}</div>
  {/for}
{/if}
```

**Expected Output**:
```html
<template x-if="showItems">
  <template x-for="item in items">
    <div><span x-text="item.name"></span></div>
  </template>
</template>
```

**Requirement**: The conditional should transform the loop normally, loop transformation should be independent

**Check**: `transformer/conditionals.go` - does it call `transformNodes()` recursively on conditional content? If yes, loops inside should work automatically.

### 7. Components in Loops

**Input**:
```
{for product in products}
  <ProductCard product={product} />
{/for}
```

**Expected Output** (after Spec 1 is implemented):
```html
<template x-for="product in products">
  <div x-data="{ product: product, name: product.name, price: product.price }">
    <!-- ProductCard content with bound props -->
  </div>
</template>
```

**Requirement**: Component transformation should work normally inside loop scope

**Note**: This might already work once Spec 1 (Recursive Component Transformation) is implemented. The component's `resolvePropValue()` should be able to look up `product` in the loop body scope.

### 8. Loop with Index

**Check if Supported**: Does the parser support index syntax?

**Syntax**: `{for item, index in items}`

**Alpine.js Output**: `x-for="(item, index) in items"`

**Implementation** (if supported):
```go
if node.Index != "" {
    xForExpr = fmt.Sprintf("(%s, %s) in %s", node.Iterator, node.Index, node.Collection)
    loopBodyScope[node.Index] = nil
}
```

### 9. Error Handling

**Add Checks**:
- Log warning if `node.Collection` is empty
- Log warning if `node.Iterator` is empty
- Handle empty `node.Content` (empty loop body)

**Example**:
```go
if node.Iterator == "" {
    log.Printf("Warning: Loop has empty iterator name")
    node.Iterator = "item" // Fallback
}

if node.Collection == "" {
    log.Printf("Warning: Loop has empty collection name")
    return []ast.Node{} // Skip this loop
}
```

### 10. Debugging and Logging

**Add Debug Logs**:
```go
log.Printf("Transforming loop: %s in %s", node.Iterator, node.Collection)
log.Printf("  Loop body has %d nodes", len(node.Content))
log.Printf("  Parent scope before loop: %v", dataScope)
log.Printf("  Loop body scope: %v", loopBodyScope)
log.Printf("  Transformed loop body to %d nodes", len(transformedContent))
```

### 11. Test-Driven Fixes

**Process**:

1. Run failing test and capture output
2. Identify exact difference between expected and actual
3. Add debug logging to relevant functions
4. Run test again to see where transformation diverges
5. Implement fix
6. Verify test passes
7. Run full test suite to ensure no regressions

**Test Commands**:
```bash
# Individual test
go test ./tests/alpine -v -run TestAlpineIntegration/loop_rendering

# With race detection
go test ./tests/alpine -v -race -run TestAlpineIntegration/loop_rendering

# All loop-related tests
go test ./tests/alpine -v -run Loop

# Full integration tests
go test ./tests/alpine -v -run TestAlpineIntegration
```

### 12. Potential Issues to Look For

**Issue**: Double transformation of loop content
- **Symptom**: Loop content appears twice or has extra wrappers
- **Cause**: `transformNodes()` being called multiple times on same content

**Issue**: Missing x-for attribute
- **Symptom**: Template element without x-for
- **Cause**: Attribute not being added correctly

**Issue**: Wrong x-for syntax
- **Symptom**: Alpine.js errors in browser console
- **Cause**: Expression format doesn't match Alpine.js expectations

**Issue**: Iterator variable not available in loop body
- **Symptom**: Expressions like `{item.name}` show "item is undefined"
- **Cause**: Iterator not added to loop body scope

**Issue**: Collection variable not found
- **Symptom**: "items is undefined" error
- **Cause**: Collection not added to parent scope

### 13. Integration with Other Specs

**Dependency on Spec 1**:
- If components in loops are failing, wait for Spec 1 to be implemented first
- Component prop resolution inside loops requires proper scope passing

**Dependency on Spec 2**:
- If loop tests use functions in loop body, may need Spec 2 fixes
- However, loop structure should work independently of function handling

**Priority**:
- Can work on loop fixes in parallel with Specs 1 and 2
- Or can implement after Spec 1 to see if component transformation fixes some loop issues

## External Dependencies

No new external dependencies required. Uses only Go standard library and internal packages.

## Success Criteria

1. All loop transformation tests pass
2. Loops render correctly in HTML output
3. No Alpine.js console errors related to loop rendering
4. Nested structures work correctly
5. Iterator variables are properly scoped
