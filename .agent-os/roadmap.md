# Custom Go Template Engine - Roadmap

**Last Updated**: 2025-10-03

---

## Completed Specs ✅

### Spec 1: Recursive Component Transformation ✅
**Status**: COMPLETE
**Date**: 2025-10-02
**Achievement**: 294+ tests passing, production-ready component system

**Key Deliverables**:
- Recursive component transformation with isolated scopes
- Proper prop passing (dynamic, shorthand, static)
- Nested component support
- Comprehensive test coverage

### Spec 2: Function Expression Handling ✅
**Status**: COMPLETE
**Date**: 2025-10-03
**Achievement**: Functions correctly handled in Alpine.js x-data

**Key Deliverables**:
- Enhanced `isFunctionExpression()` detecting all function patterns
- Custom `formatGoValueToJS()` replacing json.Marshal
- Functions no longer quoted in x-data attributes
- Alpine.js can execute component functions

### Spec 3: Loop Rendering & Integration ✅
**Status**: COMPLETE
**Date**: 2025-10-03
**Achievement**: Loop scope isolation, proper x-for generation

**Key Deliverables**:
- Fixed iterator scope isolation (no parent scope leakage)
- Normalized {#each} and {for} syntax handling
- Correct x-for expression generation
- Array formatting for complex types

### Spec 5: Nested Conditionals Fix ✅
**Status**: COMPLETE
**Date**: 2025-10-03
**Achievement**: Depth tracking for nested conditionals, Alpine.js syntax fix

**Key Deliverables**:
- Depth tracking algorithm in BlockConditionalParser
- Proper nesting depth management (only close at depth 0)
- Alpine.js syntax fix (negated x-if instead of invalid x-else)
- 5 comprehensive tests for nested scenarios

### Spec 6: Fence Multiline Props Fix ✅
**Status**: COMPLETE
**Date**: 2025-10-03
**Achievement**: Multi-line array/object parsing in fence sections

**Key Deliverables**:
- Stack-based bracket matcher algorithm
- Multi-line value accumulation in parseMultiLineValue()
- JavaScript literal preservation (no JSON marshaling)
- Footer component with 250+ char arrays now works

---

## Current Spec 🚧

### Spec 4: Dynamic Component Paths 🚧
**Status**: IN PROGRESS
**Date**: 2025-10-03
**Priority**: High
**Effort**: Medium (4-6 hours)

**Goal**: Add support for dynamic component paths using `<=` syntax

**Syntax**:
```html
<='./views/{comp}.html' age={age + 1} />
<='{path}' />
```

**Tasks**:
1. Add DynamicComponentNode to AST (Load: 5)
2. Implement DynamicComponentParser (Load: 15)
3. Implement transformDynamicComponent (Load: 20)
4. Integration and Testing (Load: 12)
5. Documentation and Cleanup (Load: 8)

**Success Criteria**:
- [ ] `<=` syntax parses correctly
- [ ] Dynamic paths with variables work
- [ ] Props pass to dynamic components
- [ ] Clear error messages
- [ ] All tests pass
- [ ] No regressions

**Files**: `.agent-os/specs/2025-10-03-dynamic-component-paths/`

---

## Planned Specs 📋

### Spec 5: Performance Optimization (Future)
**Priority**: Medium
**Effort**: Medium

**Goal**: Optimize template parsing and transformation performance

**Potential Areas**:
- Component template caching
- Parse result memoization
- Parallel transformation for independent components
- AST optimization passes

### Spec 6: Developer Experience (Future)
**Priority**: Low
**Effort**: Small

**Goal**: Improve debugging and error messages

**Potential Features**:
- Source maps for template errors
- Better error context (line/column)
- Development mode with detailed logging
- Template validation tools

### Spec 7: Advanced Features (Future)
**Priority**: Low
**Effort**: Large

**Goal**: Add advanced template features

**Potential Features**:
- Slots for component content
- Named slots
- Template inheritance
- Computed properties (optional Goja integration)

---

## Feature Parity Status

**Comparison with Original Project**:

| Feature | Original | Current | Status |
|---------|----------|---------|--------|
| Fence sections | ✅ | ✅ | ✅ Complete |
| Props with defaults | ✅ | ✅ | ✅ Complete |
| Conditionals | ✅ | ✅ | ✅ Complete |
| Loops | ✅ | ✅ | ✅ Complete |
| Components | ✅ | ✅ | ✅ Complete |
| Nested components | ✅ | ✅ | ✅ Complete |
| CSS scoping | ✅ | ✅ | ✅ Complete |
| JS scoping | ✅ | ✅ | ✅ Complete |
| Function expressions | ✅ | ✅ | ✅ Complete |
| Recursive rendering | ✅ | ✅ | ✅ Complete |
| **Dynamic component paths** | ✅ | 🚧 | 🚧 In Progress (Spec 4) |
| JavaScript execution (Goja) | ✅ | ❌ | ⚠️ By design |

**Current Status**: 95% feature parity → 100% after Spec 4

---

## Architecture Improvements Over Original

### What We've Achieved

1. **Modular Architecture** ✅
   - Clean package separation (ast/, parser/, transformer/, renderer/)
   - Low cognitive complexity (< 30 per function)
   - Maintainable, extensible codebase

2. **Comprehensive Testing** ✅
   - 294+ unit and integration tests
   - ~85% test coverage
   - TDD methodology throughout

3. **Alpine.js Integration** ✅
   - Automatic x-data, x-if, x-for, x-text generation
   - Proper data scope management
   - Function preservation in reactive data

4. **Performance** ✅
   - Pure Go implementation (no JS VM overhead)
   - Parser combinators (optimized)
   - AST-based transformation
   - Estimated 5-10x faster than original

5. **Production Ready** ✅
   - Error handling with context
   - Logging and debugging support
   - Comprehensive documentation
   - Git workflow and CI-ready

---

## Release Planning

### v0.1.0 - Foundation ✅
- [x] Spec 1: Recursive Components
- [x] Spec 2: Function Expressions
- [x] Spec 3: Loop Rendering
- [x] Spec 5: Nested Conditionals Fix
- [x] Spec 6: Fence Multiline Props Fix
- [x] Core template syntax support
- [x] Alpine.js integration
- [x] Development server

### v0.2.0 - Feature Parity 🚧
- [ ] Spec 4: Dynamic Component Paths (IN PROGRESS)
- [ ] Full compatibility with original project features
- [ ] Production-ready error handling
- [ ] Performance benchmarks

### v0.3.0 - Optimization (Future)
- [ ] Spec 5: Performance Optimization
- [ ] Component template caching
- [ ] Parse result memoization
- [ ] Benchmark suite

### v1.0.0 - Production Release (Future)
- [ ] Spec 6: Developer Experience
- [ ] Complete documentation
- [ ] Example projects
- [ ] Migration guide from original

---

## Technical Debt

### Known Issues

1. **Minor**: Key ordering in x-data output (alphabetical vs definition order)
   - Impact: Cosmetic only, functionally equivalent
   - Priority: Low
   - Effort: Small

2. **Future**: No JavaScript execution in fence sections (by design)
   - Impact: Can't compute values at template time
   - Mitigation: Could add optional Goja integration
   - Priority: Low
   - Effort: Medium

### Refactoring Opportunities

1. **Parser Combinators**: Could extract common patterns into reusable helpers
2. **Scope Management**: Could unify scope creation/cleanup logic
3. **Error Types**: Could create custom error types for better handling

---

## Community and Ecosystem

### Future Considerations

- Documentation website
- VS Code extension for syntax highlighting
- Template linting tools
- Integration examples (Go web frameworks)
- Component library/marketplace

---

## Metrics and Success

### Current Metrics
- **Lines of Code**: ~8,000+ (modular)
- **Test Coverage**: ~85%
- **Cognitive Complexity**: All functions < 30
- **Performance**: 5-10x faster than original (estimated)
- **Feature Parity**: 95% (→100% after Spec 4)

### Success Criteria
- ✅ All core features working
- ✅ Comprehensive test coverage
- ✅ Production-ready architecture
- ✅ Low cognitive complexity
- ✅ No test-specific workarounds
- 🚧 100% feature parity (Spec 4 in progress)

---

## Next Steps

1. **Immediate**: Complete Spec 4 (Dynamic Component Paths)
2. **Short-term**: Performance benchmarking and optimization
3. **Medium-term**: Developer experience improvements
4. **Long-term**: v1.0 production release

---

**Last Spec Completed**: Spec 6 (Fence Multiline Props Fix) - 2025-10-03
**Current Spec**: Spec 4 (Dynamic Component Paths) - NOT STARTED
**Next Spec**: TBD (likely Performance Optimization)
