# Comprehensive Testing Standard

**Purpose**: Prevent critical feature loss during refactoring through mandatory regression testing.

**Created**: 2025-10-08
**Last Updated**: 2025-10-08
**Status**: Active

## Critical Lesson: The Style Aggregation Bug

### What Happened
When implementing the Global Store System, `RenderWithStores()` was created to replace `Render()`. The new function:
- ✅ Added store initialization (new feature)
- ✅ Combined store scripts with component scripts
- ❌ **LOST component style aggregation** (critical bug)

**Result**: All component styles (Age, UserProfile, AdminPanel, Footer) disappeared from rendered pages.

### Root Cause Analysis
1. **Incomplete Feature Transfer**: New function didn't include `GetAggregatedStyles()` call
2. **Similar Function Confusion**: Used `generateStyle()` instead of `GetAggregatedStyles()`
   - `generateStyle()`: Extracts inline `<style>` tags only
   - `GetAggregatedStyles()`: Aggregates ALL component styles
3. **Missing Context**: New signature didn't include `templatePath` and `originalAST` needed for aggregation
4. **No Regression Test**: No test verified component styles were still aggregated

## Mandatory Test Categories

### 1. Feature Preservation Tests (CRITICAL)

**When**: Creating a new version of an existing function

**Purpose**: Verify ALL original features still work

**Example from Style Aggregation Bug**:
```go
// MANDATORY: Test all features of original Render() are preserved in RenderWithStores()
func TestRenderWithStores_PreservesAllFeatures(t *testing.T) {
    // Setup: Use a real page with multiple components
    originalAST := parseTemplate("examples/pages/home.html")
    transformedAST := transformer.TransformAST(originalAST, props)
    stores := map[string]string{"theme": "{ mode: 'light' }"}

    // Execute
    markup, script, styles := renderer.RenderWithStores(
        originalAST,           // Need original for imports
        transformedAST,        // Need transformed for output
        stores,                // New feature
        "examples/pages/home.html", // Need path for component name
    )

    // CRITICAL: Verify component style aggregation (the feature we lost!)
    t.Run("aggregates_component_styles", func(t *testing.T) {
        // Each component should have its styles included
        assert.Contains(t, styles, "Styles from: Age",
            "Age component styles missing")
        assert.Contains(t, styles, "Styles from: UserProfile",
            "UserProfile component styles missing")
        assert.Contains(t, styles, "Styles from: AdminPanel",
            "AdminPanel component styles missing")
        assert.Contains(t, styles, "Styles from: Footer",
            "Footer component styles missing")

        // Verify actual CSS rules are present
        assert.Contains(t, styles, ".age-badge")
        assert.Contains(t, styles, ".profile-card")
        assert.Contains(t, styles, ".admin-panel")
    })

    // Verify new features work
    t.Run("initializes_stores", func(t *testing.T) {
        assert.Contains(t, script, "Alpine.store('theme'",
            "Store initialization missing")
    })

    // Verify original features still work
    t.Run("generates_markup", func(t *testing.T) {
        assert.Contains(t, markup, "age-badge",
            "Component markup missing")
        assert.Contains(t, markup, "x-data",
            "Alpine.js attributes missing")
    })
}
```

### 2. Integration Tests with Real Data

**When**: Testing systems that aggregate or compose multiple sources

**Purpose**: Catch bugs that unit tests miss

**Example**:
```go
func TestComponentStyleAggregation_Integration(t *testing.T) {
    // Use REAL templates, not mocks
    tests := []struct {
        name               string
        templatePath       string
        expectedComponents []string
        expectedClasses    []string
    }{
        {
            name:         "home page with all components",
            templatePath: "examples/pages/home.html",
            expectedComponents: []string{
                "Age", "UserProfile", "AdminPanel",
                "Footer", "HeaderSimple", "Notification",
            },
            expectedClasses: []string{
                ".age-badge", ".profile-card", ".admin-panel",
                ".footer", ".header", ".notification",
            },
        },
        {
            name:         "page with nested components",
            templatePath: "examples/pages/comprehensive-simple.html",
            expectedComponents: []string{"Header", "ProductCard"},
            expectedClasses: []string{".header", ".product-card"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Parse real template
            content, _ := os.ReadFile(tt.templatePath)
            originalAST, _ := parser.ParseTemplate(string(content))

            // Get aggregated styles
            componentName := extractComponentName(tt.templatePath)
            styles := renderer.GetAggregatedStyles(originalAST, componentName)

            // Verify all components present
            for _, component := range tt.expectedComponents {
                assert.Contains(t, styles, fmt.Sprintf("Styles from: %s", component),
                    "Component %s styles missing from %s", component, tt.name)
            }

            // Verify CSS classes present
            for _, class := range tt.expectedClasses {
                assert.Contains(t, styles, class,
                    "CSS class %s missing from %s", class, tt.name)
            }
        })
    }
}
```

### 3. Regression Tests for Known Bugs

**When**: After fixing a bug

**Purpose**: Ensure the bug never comes back

**Example**:
```go
// Regression test for: Component styles not aggregated in RenderWithStores
// Bug ID: #2025-10-08-style-aggregation
// Fixed: 2025-10-08 commit 0f1187c
func TestRegressionStyleAggregation_Issue20251008(t *testing.T) {
    // This bug occurred when RenderWithStores() was created
    // It used generateStyle() instead of GetAggregatedStyles()

    t.Run("home_page_has_all_component_styles", func(t *testing.T) {
        content, _ := os.ReadFile("examples/pages/home.html")
        originalAST, _ := parser.ParseTemplate(string(content))
        transformed := transformer.TransformAST(originalAST, nil)

        _, _, styles := renderer.RenderWithStores(
            originalAST, transformed, nil, "examples/pages/home.html",
        )

        // The bug: these were all missing!
        missingComponents := []string{"Age", "UserProfile", "AdminPanel", "Footer"}
        for _, comp := range missingComponents {
            assert.Contains(t, styles, fmt.Sprintf("Styles from: %s", comp),
                "REGRESSION: Component %s styles missing (bug returned!)", comp)
        }
    })

    t.Run("preserves_original_ast_for_imports", func(t *testing.T) {
        // The bug was caused by not having access to original AST
        // which contains FenceSection.Imports needed for aggregation

        content, _ := os.ReadFile("examples/pages/home.html")
        originalAST, _ := parser.ParseTemplate(string(content))

        // Verify imports are in original AST
        var hasImports bool
        for _, node := range originalAST.RootNodes {
            if fence, ok := node.(*ast.FenceSection); ok {
                hasImports = len(fence.Imports) > 0
                break
            }
        }

        assert.True(t, hasImports,
            "Original AST must preserve imports for style aggregation")
    })
}
```

### 4. Side Effect Verification Tests

**When**: Testing functions with side effects (logging, caching, etc.)

**Purpose**: Ensure "invisible" features aren't lost

**Example**:
```go
func TestRenderWithStores_PreservesSideEffects(t *testing.T) {
    var logBuffer bytes.Buffer
    logger := setupTestLogger(&logBuffer)

    originalAST, _ := parseTemplate("test.html")
    transformed := transformer.TransformAST(originalAST, nil)

    t.Run("logs_aggregation_calls", func(t *testing.T) {
        logBuffer.Reset()

        renderer.RenderWithStores(originalAST, transformed, nil, "test.html")

        logs := logBuffer.String()

        // Verify logging still happens (side effect)
        assert.Contains(t, logs, "[RenderWithStores] Aggregating styles")
        assert.Contains(t, logs, "Aggregated")
        assert.Contains(t, logs, "bytes of styles")
    })

    t.Run("uses_style_cache", func(t *testing.T) {
        // First call - cache miss
        renderer.ClearStyleCache()
        renderer.RenderWithStores(originalAST, transformed, nil, "test.html")

        firstLog := logBuffer.String()
        assert.Contains(t, firstLog, "MISS", "First call should be cache miss")

        // Second call - cache hit
        logBuffer.Reset()
        renderer.RenderWithStores(originalAST, transformed, nil, "test.html")

        secondLog := logBuffer.String()
        assert.Contains(t, secondLog, "HIT", "Second call should be cache hit")
    })
}
```

## Test Categories Checklist

For any function replacement or major refactoring:

```markdown
### Feature Preservation
□ Test all original features still work
□ Test all original parameters are preserved or accessible
□ Test all return values match expected format
□ Test with real data (not just mocks)

### Integration Testing
□ Test with actual files/templates
□ Test with multiple components
□ Test with edge cases (empty, maximum, nested)
□ Test end-to-end workflows

### Regression Prevention
□ Create test for the specific bug fixed
□ Add test ID and commit reference
□ Document why the bug occurred
□ Test the exact scenario that failed

### Side Effects
□ Test logging occurs correctly
□ Test caching behavior
□ Test performance characteristics
□ Test resource cleanup

### Edge Cases
□ Test with nil/empty inputs
□ Test with maximum sizes
□ Test with invalid inputs
□ Test concurrent access (if applicable)
```

## When Tests Are MANDATORY

### 1. Function Replacement
**Trigger**: Creating `FunctionV2()` to replace `Function()`

**Required**:
- Feature preservation test
- Integration test with real data
- Side effect verification test

### 2. Major Refactoring
**Trigger**: Changing > 50% of function implementation

**Required**:
- All existing tests must still pass
- New tests for changed behavior
- Integration test for affected features

### 3. Bug Fixes
**Trigger**: Fixing a production bug

**Required**:
- Regression test for specific bug
- Test must fail on old code
- Test must pass on new code

### 4. New Features with Dependencies
**Trigger**: Adding feature that depends on existing functionality

**Required**:
- Test new feature works
- Test existing features still work
- Integration test showing both working together

## Test Quality Checklist

```markdown
### Test Structure
□ Clear test name describes what's being tested
□ Arranged in Given-When-Then or AAA pattern
□ Uses real data when possible (not just mocks)
□ Asserts on specific expected values
□ Includes helpful error messages

### Test Coverage
□ Happy path tested
□ Error cases tested
□ Edge cases tested
□ Integration scenarios tested

### Test Maintainability
□ Test is independent (no global state)
□ Test is repeatable (same result every time)
□ Test is fast (< 1 second for unit tests)
□ Test failures are easy to diagnose
```

## Common Testing Mistakes to Avoid

### ❌ Mistake 1: Only Testing New Features
```go
// BAD: Only tests store initialization (new feature)
func TestRenderWithStores(t *testing.T) {
    _, script, _ := RenderWithStores(...)
    assert.Contains(t, script, "Alpine.store")
    // Missing: Tests for style aggregation, markup generation, etc.
}
```

**Fix**: Test ALL features, both new and original

### ❌ Mistake 2: Using Mocks When Real Data Is Better
```go
// BAD: Mock doesn't catch real template parsing issues
func TestStyleAggregation(t *testing.T) {
    mockAST := &ast.Template{...} // Simplified mock
    styles := GetAggregatedStyles(mockAST, "test")
    assert.NotEmpty(t, styles)
}
```

**Fix**: Use real templates from `examples/pages/`

### ❌ Mistake 3: Not Testing Side Effects
```go
// BAD: Doesn't verify logging or caching
func TestRender(t *testing.T) {
    markup, _, _ := Render("test.html", nil)
    assert.NotEmpty(t, markup)
    // Missing: Verify logs, cache usage, etc.
}
```

**Fix**: Add side effect tests

### ❌ Mistake 4: Generic Assertions
```go
// BAD: Too generic - won't catch specific missing components
func TestStyles(t *testing.T) {
    _, _, styles := Render(...)
    assert.NotEmpty(t, styles) // Could pass even if components missing!
}
```

**Fix**: Assert specific expected values

## Template for New Tests

```go
// Test Name Format: Test<Function>_<Scenario>_<Expected>
func TestRenderWithStores_MultipleComponents_AggregatesAllStyles(t *testing.T) {
    // GIVEN: A page template with 3 components
    content, err := os.ReadFile("examples/pages/test-page.html")
    require.NoError(t, err, "Failed to read test template")

    originalAST, err := parser.ParseTemplate(string(content))
    require.NoError(t, err, "Failed to parse template")

    transformedAST := transformer.TransformAST(originalAST, nil)
    stores := map[string]string{"test": "{}"}

    // WHEN: Rendering with stores
    markup, script, styles := renderer.RenderWithStores(
        originalAST,
        transformedAST,
        stores,
        "examples/pages/test-page.html",
    )

    // THEN: All component styles are aggregated
    t.Run("includes_all_component_styles", func(t *testing.T) {
        expectedComponents := []string{"Header", "Footer", "Sidebar"}
        for _, comp := range expectedComponents {
            assert.Contains(t, styles, fmt.Sprintf("Styles from: %s", comp),
                "Component %s styles should be included", comp)
        }
    })

    // AND: Store initialization works
    t.Run("initializes_stores", func(t *testing.T) {
        assert.Contains(t, script, "Alpine.store('test')",
            "Store should be initialized")
    })

    // AND: Markup is generated
    t.Run("generates_markup", func(t *testing.T) {
        assert.NotEmpty(t, markup, "Markup should not be empty")
        assert.Contains(t, markup, "x-data", "Should contain Alpine.js attributes")
    })
}
```

## Success Metrics

A comprehensive test suite should:
- ✅ Catch 100% of feature regressions like the style aggregation bug
- ✅ Execute in < 5 seconds (unit tests) or < 30 seconds (integration tests)
- ✅ Provide clear failure messages that point to the exact problem
- ✅ Use real data to catch integration issues
- ✅ Document why each test exists (reference to feature or bug)

## References

- **Real Bug Example**: Style aggregation loss in RenderWithStores (2025-10-08)
- **Fix Commit**: 0f1187c - "fix: Restore component style aggregation in RenderWithStores"
- **Related Standard**: Function Replacement Safety Checklist (go-backend.md)
