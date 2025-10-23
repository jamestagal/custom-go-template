# Technical Specification

This is the technical specification for the spec detailed in @.agent-os/specs/2025-10-22-x-data-optimization/spec.md

## Technical Requirements

### Phase 1: Remove Root Div Wrapper

**File**: `transformer/transformer.go`
**Function**: `transformNodes()` (lines 197-199)

**Change Required**:
- Remove call to `applyAlpineDataWrapper()` at root level
- Body tag injection (in `cmd/server/main.go`) already provides top-level x-data scope
- No additional wrapping needed at transformer level

**Implementation**:
```go
// BEFORE:
if applyAlpineWrapper && hasDataScope {
    return applyAlpineDataWrapper(transformedNodes, dataScope)
}

// AFTER:
// Root-level x-data is handled by server body injection
// Only wrap at component level if needed
if applyAlpineWrapper && hasDataScope && !isRootLevel() {
    return applyAlpineDataWrapper(transformedNodes, dataScope)
}
```

**Testing Requirements**:
- Verify pages render without root wrapper
- Confirm Alpine.js scope available throughout page
- Check no "undefined variable" errors in console
- Validate all example pages still work

---

### Phase 2: Optimize Component Wrappers

**New File**: `transformer/scope.go`
**Purpose**: Scope analysis and diffing utilities

**Functions to Implement**:

1. **ScopeDiff** - Compare child vs parent scope
```go
func ScopeDiff(child, parent map[string]any) map[string]any {
    diff := make(map[string]any)
    for key, childValue := range child {
        parentValue, existsInParent := parent[key]
        if !existsInParent || !reflect.DeepEqual(childValue, parentValue) {
            diff[key] = childValue
        }
    }
    return diff
}
```

2. **HasNewVariables** - Check if component introduces new variables
```go
func HasNewVariables(componentScope, parentScope map[string]any) bool {
    return len(ScopeDiff(componentScope, parentScope)) > 0
}
```

**File to Modify**: `transformer/components.go`
**Function**: `transformComponent()`

**Change Required**:
- Add scope diffing before wrapping components
- Only wrap if component has NEW variables
- Otherwise, skip wrapper and let component inherit from parent

**Implementation**:
```go
// Transform component nodes
transformedNodes := transformComponentNodes(node, componentScope)

// OPTIMIZATION: Only wrap if component has NEW variables
scopeDiff := ScopeDiff(componentScope, parentScope)

if len(scopeDiff) > 0 {
    log.Printf("[X-Data Optimization] Component '%s' adds %d new variables",
        node.Name, len(scopeDiff))
    return wrapWithXData(transformedNodes, scopeDiff)  // Wrap with DIFF only
} else {
    log.Printf("[X-Data Optimization] Component '%s' inherits from parent",
        node.Name)
    return transformedNodes  // No wrapper - inherit from parent
}
```

**Testing Requirements**:
- Test static components (should inherit, no wrapper)
- Test stateful components (should get wrapper with new vars only)
- Test nested components
- Test components in loops
- Verify Alpine.js reactivity works
- Check store access works

---

### Phase 3: Optimize Runtime Wrappers

**File**: `transformer/dynamic_component_by_name.go`
**Function**: `emitRuntimeWrapper()` (lines 124-212)

**Change Required**:
- Build x-data with prop REFERENCES instead of full data serialization
- Pass variable names that point to parent scope

**Implementation**:
```go
// BEFORE:
xDataValue := fmt.Sprintf("{compName: %s, compProps: %s}",
    node.NameExpression,
    serializeProps(mergedProps))  // Full serialization

// AFTER:
xDataValue := fmt.Sprintf("{compName: %s, compProps: %s}",
    node.NameExpression,
    buildPropReferences(mergedProps))  // Just references
```

**File**: `static/js/runtime-components.js`
**Function**: `renderDynamicComponent()` (line 219)

**Change Required**:
- Read from parent Alpine.js scope instead of creating new scope
- Use Alpine's `$data` magic to resolve references

**Implementation**:
```javascript
// BEFORE:
const xDataJSON = JSON.stringify({ props: normalizedProps });
el.innerHTML = `<div x-data='${xDataJSON}'>${html}</div>`;

// AFTER:
// Just insert HTML - Alpine inherits from parent scope
el.innerHTML = html;

// Only add x-data if component has LOCAL state (not just inherited props)
if (componentHasLocalState) {
    el.innerHTML = `<div x-data='${localStateOnly}'>${html}</div>`;
}
```

**Testing Requirements**:
- Test runtime components with static props
- Test runtime components in loops
- Test nested runtime components
- Verify prop resolution works
- Check Alpine.js reactivity maintained

---

## Performance Requirements

### Success Metrics

**HTML Payload Reduction**:
- Phase 1: 25% reduction in x-data bloat (~200KB saved)
- Phase 2: 60-70% additional reduction (~400-500KB saved)
- Phase 3: 10-15% additional reduction (~80-120KB saved)
- **Total Target**: 90-95% reduction (800KB → 40-80KB)

**Performance Benchmarks**:
- Measure HTML size before/after each phase
- Measure Alpine.js initialization time
- Measure page load time (DOMContentLoaded)
- Measure Lighthouse performance score

**Test Page**: Homepage with 10 components

**Before Optimization**:
- HTML size: ~850KB
- x-data count: 12-15
- Total x-data size: ~800KB

**After All Phases**:
- HTML size: ~220KB (↓74%)
- x-data count: 1-2
- Total x-data size: ~180KB (↓77%)

---

## Alpine.js Integration Requirements

### Scope Inheritance Pattern

Alpine.js uses prototype-based scope inheritance:

```html
<body x-data="{user: 'John', age: 30}">
  <div>
    <!-- Can access user and age via inheritance -->
    <span x-text="user"></span>
  </div>

  <div x-data="{role: 'Admin'}">
    <!-- Can access BOTH inherited (user, age) AND local (role) -->
    <span x-text="user + ' - ' + role"></span>
  </div>
</body>
```

**Key Principle**: Only add x-data when:
1. Introducing NEW variables not in parent
2. Overriding parent variables
3. Isolating component state

---

## Error Handling Requirements

### Scope Resolution Errors

**Risk**: Variables not available after removing wrappers

**Mitigation**:
1. Add comprehensive logging during transformation
2. Log when wrapper skipped vs added
3. Log variables being inherited vs new
4. Implement fallback: if variable not in parent, keep wrapper

**Example Logging**:
```go
log.Printf("[X-Data Optimization] Component '%s' inheriting from parent: %v",
    componentName, parentVarNames)
log.Printf("[X-Data Optimization] Component '%s' adding new variables: %v",
    componentName, newVarNames)
log.Printf("[X-Data Optimization] WARNING: Variable '%s' not found in parent scope, adding wrapper",
    varName)
```

---

## Backward Compatibility Requirements

### No Breaking Changes

**Requirement**: All existing templates must work without modification

**Validation**:
1. Run full test suite after each phase
2. Test all example pages
3. Verify no console errors
4. Check Alpine.js directives work
5. Confirm store access works

### Feature Flag

**Implementation**: Add feature flag to enable/disable optimization

```go
// transformer/config.go
var OptimizeXData = true  // Can disable if issues arise
```

**Usage**:
```go
if OptimizeXData {
    // Use scope diffing
    scopeDiff := ScopeDiff(componentScope, parentScope)
    if len(scopeDiff) > 0 {
        return wrapWithXData(transformedNodes, scopeDiff)
    }
    return transformedNodes
} else {
    // Legacy behavior - always wrap
    return wrapWithXData(transformedNodes, componentScope)
}
```

---
## Fallback Strategy

### Conservative Mode

**Trigger**: If optimization introduces errors, fall back to safe mode

**Implementation**:
```go
// transformer/scope.go
type OptimizationMode int

const (
    ModeAggressive OptimizationMode = iota  // Full optimization
    ModeConservative                         // Only safe optimizations
    ModeLegacy                               // No optimization
)

var currentMode = ModeAggressive

func (t *Transformer) optimizeXData(component, parent map[string]any) {
    switch currentMode {
    case ModeAggressive:
        // Full scope diffing + runtime prop references
        return aggressiveOptimization(component, parent)
        
    case ModeConservative:
        // Only remove obvious duplicates (Phase 1 + partial Phase 2)
        if isExactDuplicate(component, parent) {
            return nil  // No wrapper
        }
        return component  // Keep wrapper
        
    case ModeLegacy:
        // Original behavior
        return component  // Always wrap
    }
}
```

**Trigger Conditions**:
- Alpine.js console errors increase > 10%
- Undefined variable errors detected
- Test suite failures
- User reports component not working

**Recovery**:
```bash
# Automatic rollback if errors detected
go build -tags "xdata_mode=conservative"
```


## Code Quality Requirements

### Cognitive Load

**Requirement**: All modified files must maintain cognitive load < 30

**Validation**:
- Use cognitive load checker from `.agent-os/standards/`
- Keep functions focused and single-purpose
- Extract complex logic to helper functions

### Testing Coverage

**Requirement**: Maintain >80% test coverage

**Test Suite Structure**:
```
tests/x-data-optimization/
├── phase1_root_wrapper_test.go      # Phase 1 tests
├── phase2_component_wrapper_test.go # Phase 2 tests
├── phase3_runtime_wrapper_test.go   # Phase 3 tests
└── integration_test.go              # Full pipeline tests
```

### Error Wrapping

**Requirement**: All errors must be wrapped with context

```go
if err != nil {
    return nil, fmt.Errorf("scope diff failed for component %s: %w",
        componentName, err)
}
```

---

## External Dependencies

**No new external dependencies required**

This optimization uses only:
- Go standard library (`reflect` for deep equality)
- Existing project packages (`transformer`, `ast`, `renderer`)
- Existing Alpine.js 3.x (no upgrade needed)

All required functionality can be implemented with current dependencies.
