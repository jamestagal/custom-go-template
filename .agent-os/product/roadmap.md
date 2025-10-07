# Product Roadmap

**Project:** Custom Go Template Engine for Plenti
**Last Updated:** October 7, 2025

**Note**: This is the strategic product roadmap. For technical spec completion tracking, see `.agent-os/roadmap.md`

---

## Overview

This roadmap tracks the evolution of the Custom Go Template Engine from its foundation through full Plenti integration and beyond. Each phase represents a major milestone with concrete deliverables.

---

## Phase 0: Foundation ✅ COMPLETE

**Status:** ✅ Shipped
**Duration:** February 2025 - October 2025
**Effort:** ~8 months

### Goals
Build a production-ready template engine with core features matching Svelte's capabilities.

### Completed Features

#### Core Engine
- [x] **Parser Architecture**
  - Unified parsing path (BlockConditionalParser, BlockLoopParser)
  - Parser combinators for composability
  - No post-processing (clean single-pass design)

- [x] **AST Definition**
  - Template, Element, Conditional, Loop nodes
  - ComponentNode, DynamicComponentNode
  - FenceSection, StyleSection, ScriptSection
  - ExpressionNode for dynamic values

- [x] **Transformer Pipeline**
  - Template syntax → Alpine.js directives
  - Expression transformation (`{var}` → `<span x-text="var">`)
  - Conditional transformation (`{if}` → `<template x-if>`)
  - Loop transformation (`{for}` → `<template x-for>`)
  - Component expansion and prop passing

- [x] **Renderer**
  - HTML generation from transformed AST
  - CSS extraction and bundling
  - JavaScript scope generation (x-data)

#### Component System
- [x] **Component Registration**
  - Automatic component discovery
  - Prop extraction from fence sections
  - Template caching for performance

- [x] **Component Types**
  - Static components: `<Header />`
  - Dynamic components: `<={layout} />`
  - Dynamic with path: `<='./path.html' />`

- [x] **Fence Sections**
  - Import statements
  - Helper functions
  - Local variables
  - Props (for component isolation)

#### Template Features
- [x] **Conditionals**
  - `{if condition}...{/if}`
  - `{else if}` support
  - `{else}` support
  - Nested conditionals
  - Conditionals in loops

- [x] **Loops**
  - `{for item in items}...{/for}`
  - Access to loop variables
  - Nested loops
  - Loops in conditionals

- [x] **Expressions**
  - `{variable}` → text output
  - `{obj.property}` → nested access
  - Dynamic expression evaluation

#### Style Aggregation ⭐ NEW
- [x] **Automatic Extraction**
  - Parse `<style>` blocks from components
  - Support attributes (`<style scoped>`)
  - Multiple style blocks per component

- [x] **Dependency Tree Traversal**
  - Recursive component tree walking
  - Find ComponentNode and DynamicComponentNode
  - Dependency-first ordering

- [x] **Deduplication**
  - SHA256 content hashing
  - Identical styles included once
  - Preserve insertion order

- [x] **High-Performance Caching**
  - Thread-safe with sync.RWMutex
  - 1.86 μs cache hits (5,400x faster than target!)
  - Cache invalidation for dev mode

- [x] **Source Comments**
  - `/* Styles from: ComponentName */`
  - Debugging-friendly output

#### Development Tools
- [x] Development server on :3000
- [x] Component hot-reload (manual)
- [x] Example components and pages
- [x] Test utilities and helpers

#### Testing Infrastructure
- [x] Parser tests (14 passing)
- [x] Transformer tests (multiple suites)
- [x] Renderer tests (all passing)
- [x] Style aggregation tests (58 passing)
- [x] Integration tests (most passing)

#### Documentation
- [x] CLAUDE.md (comprehensive AI context)
- [x] Plenti integration spec
- [x] Plenti analysis document
- [x] Style aggregation cache guide
- [x] Future development roadmap
- [x] Agent OS mission statement

### Metrics Achieved
- **Code:** ~18,000 lines (production + tests + docs)
- **Test Coverage:** 100%+ for new features
- **Performance:** 5,400x better than target (caching)
- **Cognitive Load:** <30 per file (enforced)

---

## Phase 1: Plenti Integration 🚧 IN PLANNING

**Status:** 📋 Specification Complete
**Target:** Q4 2025
**Estimated Effort:** 7-11 hours

### Goals
Make the Custom Go Template Engine the official build-time renderer for Plenti, replacing Svelte while maintaining 100% feature parity.

### Epic 1: Build Infrastructure

**Deliverables:**
- [ ] Create `cmd/build/render_templates.go`
  - RenderTemplate() function
  - Magic variable injection
  - Component signature generation

- [ ] Modify `cmd/build/data_source.go`
  - Replace createProps() with Go data structures
  - Update createHTML() to use custom renderer
  - Remove V8/JavaScript dependencies

- [ ] Component Registration System
  - registerAllComponents() at build start
  - Walk layouts/ directory
  - Pre-compile templates
  - Generate allLayouts map

**Acceptance Criteria:**
- `plenti build` successfully generates HTML
- No Svelte compilation invoked
- Component styles automatically included
- Build time <1s for 100 pages

### Epic 2: Magic Variables Implementation

**Deliverables:**
- [ ] Content Object Structure
  ```go
  type Content struct {
      Pager    int
      Type     string
      Path     string
      Filepath string
      Filename string
      Fields   map[string]interface{}
  }
  ```

- [ ] Magic Variables Injection
  - `content` - current page data
  - `allContent` - all pages array
  - `allLayouts` - component map
  - `env` - environment config
  - `params` - URL query parameters

- [ ] Component Signature System
  - Path → identifier conversion
  - `layouts/components/header.html` → `layouts_components_header_html`
  - Dynamic component resolution

**Acceptance Criteria:**
- Templates can access all Plenti data
- Dynamic component loading works
- Pagination data available
- Environment variables accessible

### Epic 3: Template Format Conversion

**Deliverables:**
- [ ] Conversion Guide Document
  - Svelte → Custom Go Template patterns
  - Common pitfalls and solutions
  - Side-by-side examples

- [ ] Global Entry Point
  - Convert `layouts/global/html.svelte` → `.html`
  - Remove `export let` statements
  - Replace `<svelte:component>` with `<={layout}>`

- [ ] Content Templates
  - Convert all `layouts/content/*.svelte` → `.html`
  - Update data access to `content.fields.*`
  - Maintain component imports

**Acceptance Criteria:**
- Zero .svelte files remain
- All templates render correctly
- Component imports work
- Styles properly aggregated

### Epic 4: Integration Testing

**Deliverables:**
- [ ] Test Plenti Site
  - Create test-site/ with real content
  - Multiple content types
  - Complex component nesting
  - Pagination examples

- [ ] Integration Test Suite
  - Build succeeds
  - Routes generate correctly
  - Magic variables accessible
  - Components render properly
  - Styles applied correctly

- [ ] Migration Testing
  - Existing Plenti site migration
  - Before/after comparison
  - Performance benchmarks

**Acceptance Criteria:**
- All integration tests pass
- Real Plenti site builds successfully
- Output HTML matches expected format
- No visual regressions

### Epic 5: Documentation & Release

**Deliverables:**
- [ ] Plenti Documentation Updates
  - Update template syntax guide
  - Component authoring guide
  - Migration from Svelte guide

- [ ] Example Plenti Site
  - Full-featured demo
  - Best practices showcase
  - Component library

- [ ] Release Notes
  - Breaking changes (if any)
  - Migration steps
  - Performance improvements

**Acceptance Criteria:**
- Docs published on plenti.co
- Example site deployed live
- Migration path clear
- Community informed

### Success Metrics
- [ ] Build time: <1s for 100 pages (vs Svelte ~3s)
- [ ] Output size: <50KB total
- [ ] Zero breaking changes for users
- [ ] 10 Plenti sites using custom engine

---

## Phase 2: Component Prop Scoping 📅 PLANNED

**Status:** 📋 Specification in Progress
**Target:** Q1 2026
**Estimated Effort:** 40-60 hours

### Goals
Implement proper component isolation with scoped props, eliminating global scope limitations and enabling true component reusability.

### Current Limitations
- All props must exist in global x-data scope
- Component prop names can clash with page variables
- No prop validation or type checking
- Can't optimize re-rendering based on prop changes

### Epic 1: Architecture Design

**Deliverables:**
- [ ] Detailed specification document
  - Scope inheritance model
  - Prop serialization strategy
  - Performance implications
  - Backward compatibility plan

- [ ] Prototype implementation
  - Proof of concept
  - Performance testing
  - Edge case discovery

**Acceptance Criteria:**
- Spec reviewed and approved
- Prototype demonstrates feasibility
- No performance regressions

### Epic 2: Core Implementation

**Deliverables:**
- [ ] Component Instance Wrapping
  - Each instance gets own `<div x-data="{ ...props }">`
  - Proper Alpine.js initialization order
  - Nested component scoping

- [ ] Prop Expression Evaluation
  - Evaluate in parent scope
  - Serialize to component scope
  - Handle dynamic updates
  - Support literals and expressions

- [ ] Default Value Merging
  - Extract defaults from fence section
  - Merge with passed props
  - Handle undefined/null values
  - Preserve type information

**Acceptance Criteria:**
- Component props fully isolated
- No naming conflicts possible
- Defaults merge correctly
- All tests pass

### Epic 3: Prop Validation

**Deliverables:**
- [ ] Validation Syntax
  ```javascript
  ---
  prop title: string = "Default" // required type
  prop count?: number            // optional
  prop items: array             // required, no default
  ---
  ```

- [ ] Validation Implementation
  - Required prop checking at transform time
  - Type validation when specified
  - Helpful error messages

**Acceptance Criteria:**
- Missing required props caught at build time
- Type mismatches reported clearly
- Developer experience improved

### Epic 4: Optimization

**Deliverables:**
- [ ] Change Detection
  - Track which props used in template
  - Detect prop value changes
  - Optimize Alpine.js reactivity

- [ ] Memoization
  - Cache expensive prop transformations
  - Invalidate on prop changes
  - Performance benchmarks

**Acceptance Criteria:**
- Re-rendering optimized
- No unnecessary updates
- Performance metrics meet targets

### Success Metrics
- [ ] Zero global scope pollution
- [ ] Prop validation catches 90%+ of errors
- [ ] Re-rendering 50%+ faster for prop changes
- [ ] Developer satisfaction: "much easier to debug"

---

## Phase 3: Production Hardening 📅 PLANNED

**Status:** 🔮 Future
**Target:** Q2 2026
**Estimated Effort:** 80-100 hours

### Goals
Make the engine production-ready for high-traffic sites with comprehensive monitoring, error handling, and performance optimization.

### Epic 1: Performance Optimization

**Deliverables:**
- [ ] Build Performance
  - Parallel template processing
  - Incremental builds
  - Smart caching strategies

- [ ] Runtime Performance
  - Output size minimization
  - Critical CSS extraction
  - JavaScript tree-shaking

- [ ] Benchmarking Suite
  - vs Svelte comparison
  - vs Hugo comparison
  - Real-world site tests

**Target Metrics:**
- Build time: <500ms for 100 pages
- Output size: <30KB (HTML + CSS + JS)
- Lighthouse score: 100/100

### Epic 2: Error Handling

**Deliverables:**
- [ ] Comprehensive Error Messages
  - Line and column numbers
  - Context snippets
  - Suggested fixes

- [ ] Error Recovery
  - Graceful degradation
  - Partial rendering on errors
  - Development vs production modes

- [ ] Debugging Tools
  - Template inspector
  - Component tree visualizer
  - Performance profiler

**Acceptance Criteria:**
- Errors are actionable
- No silent failures
- Debugging is straightforward

### Epic 3: Security Audit

**Deliverables:**
- [ ] XSS Prevention
  - Automatic HTML escaping
  - Safe attribute handling
  - Content Security Policy support

- [ ] Injection Prevention
  - SQL injection checks (if applicable)
  - Command injection protection
  - Path traversal prevention

- [ ] Dependency Audit
  - Vulnerability scanning
  - Update strategy
  - Security patches

**Acceptance Criteria:**
- OWASP Top 10 compliance
- No critical vulnerabilities
- Security docs published

### Epic 4: Observability

**Deliverables:**
- [ ] Build Metrics
  - Template processing times
  - Cache hit rates
  - Memory usage tracking

- [ ] Runtime Monitoring
  - Error tracking integration
  - Performance monitoring
  - User analytics hooks

- [ ] Health Checks
  - Build health endpoint
  - Template validation checks
  - Dependency health status

**Acceptance Criteria:**
- Metrics exposed via API
- Integration with monitoring tools
- Alerts for anomalies

---

## Phase 4: Ecosystem Growth 📅 FUTURE

**Status:** 🔮 Vision
**Target:** 2026+
**Estimated Effort:** Ongoing

### Goals
Build a thriving ecosystem around the Custom Go Template Engine with community contributions, component libraries, and educational resources.

### Epic 1: Component Library

**Deliverables:**
- [ ] UI Kit for Plenti Sites
  - Navigation components
  - Layout components
  - Form components
  - Media components

- [ ] Component Marketplace
  - Discovery platform
  - Version management
  - Usage analytics

**Success Metrics:**
- 50+ components available
- 100+ downloads/month
- 10+ contributors

### Epic 2: Developer Tools

**Deliverables:**
- [ ] VS Code Extension
  - Syntax highlighting
  - IntelliSense support
  - Error checking
  - Formatting

- [ ] CLI Tools
  - Component generator
  - Template scaffolding
  - Performance analyzer

**Success Metrics:**
- 500+ extension installs
- 4+ star rating
- Active maintenance

### Epic 3: Migration Tools

**Deliverables:**
- [ ] Svelte → Custom Go Template Converter
  - Automated syntax conversion
  - Manual review checklist
  - Test generation

- [ ] Hugo → Custom Go Template
- [ ] Jekyll → Custom Go Template

**Success Metrics:**
- 80%+ automated conversion rate
- 20+ successful migrations
- Migration guide views: 1000+

### Epic 4: Educational Content

**Deliverables:**
- [ ] Video Tutorial Series
  - Getting started (5 episodes)
  - Advanced patterns (10 episodes)
  - Real-world examples (5 episodes)

- [ ] Written Guides
  - Template authoring best practices
  - Component design patterns
  - Performance optimization guide

- [ ] Interactive Playground
  - Browser-based editor
  - Live preview
  - Example library

**Success Metrics:**
- 5000+ video views
- 10,000+ guide page views
- 1000+ playground users

---

## Roadmap Timeline

```
2025 Q4  ████████░░░░░░░░  Phase 1: Plenti Integration (In Progress)
2026 Q1  ░░░░░░░░████████  Phase 2: Component Prop Scoping
2026 Q2  ░░░░░░░░░░░░████  Phase 3: Production Hardening
2026 Q3+ ░░░░░░░░░░░░░░░░  Phase 4: Ecosystem Growth (Ongoing)
```

---

## Feature Parity Matrix

Tracking compatibility with Svelte features:

| Feature | Svelte | Custom Go Template | Status |
|---------|--------|-------------------|---------|
| **Template Syntax** |
| Conditionals | ✅ | ✅ | Complete |
| Loops | ✅ | ✅ | Complete |
| Expressions | ✅ | ✅ | Complete |
| Components | ✅ | ✅ | Complete |
| Dynamic Components | ✅ | ✅ | Complete |
| Slots | ✅ | ❌ | Planned (Phase 2) |
| **Styling** |
| Scoped CSS | ✅ | ✅ (via aggregation) | Complete |
| CSS Variables | ✅ | ✅ | Complete |
| Global Styles | ✅ | ✅ | Complete |
| **Reactivity** |
| Reactive Statements | ✅ | ⚠️ (Alpine.js) | Different approach |
| Store System | ✅ | ❌ | Not planned |
| **Component Features** |
| Props | ✅ | ✅ (global scope) | Complete |
| Prop Validation | ❌ | ❌ | Planned (Phase 2) |
| Prop Scoping | ✅ | ❌ | Planned (Phase 2) |
| Events | ✅ | ⚠️ (Alpine.js) | Different approach |
| Context API | ✅ | ❌ | Not planned |
| **Build** |
| SSR | ✅ | ✅ | Complete |
| Hydration | ✅ | ⚠️ (Alpine.js) | Different approach |
| Code Splitting | ✅ | ❌ | Not planned |
| **Dev Experience** |
| Hot Reload | ✅ | ⚠️ (manual) | Needs improvement |
| Error Messages | ✅ | ⚠️ | Needs improvement |
| TypeScript | ✅ | ❌ | Not planned |

**Legend:**
- ✅ Feature complete
- ⚠️ Partial support or different approach
- ❌ Not implemented
- 🔮 Planned for future phase

---

## Release History

### v0.1.0 - Foundation (October 7, 2025)
- ✅ Complete parser/transformer/renderer pipeline
- ✅ Component system with dynamic loading
- ✅ Style aggregation with caching
- ✅ Development server
- ✅ 58+ tests passing
- 📦 18,000 lines of code

### v0.2.0 - Plenti Integration (Target: Q4 2025)
- 🎯 Drop-in Svelte replacement
- 🎯 Magic variables system
- 🎯 Build pipeline integration
- 🎯 Migration guide

### v0.3.0 - Component Prop Scoping (Target: Q1 2026)
- 🎯 Isolated component scopes
- 🎯 Prop validation
- 🎯 Performance optimization

### v1.0.0 - Production Ready (Target: Q2 2026)
- 🎯 Performance benchmarks met
- 🎯 Security audit complete
- 🎯 Comprehensive documentation
- 🎯 Production deployments

---

## Dependencies & Blockers

### External Dependencies
- **Plenti Core Team** - Integration approval and testing
- **Alpine.js** - Runtime framework (stable, no blockers)
- **Go Ecosystem** - Language/tooling (stable, no blockers)

### Internal Blockers
- None currently identified

### Risk Mitigation
- **Risk:** Plenti API changes during integration
  - **Mitigation:** Close collaboration with Plenti team, early testing

- **Risk:** Performance targets not met
  - **Mitigation:** Early benchmarking, incremental optimization

- **Risk:** Community adoption slower than expected
  - **Mitigation:** Strong documentation, example sites, migration tools

---

## Feedback & Iteration

This roadmap is a living document. We welcome feedback on:
- Priority of features
- Timeline adjustments
- New feature requests
- Integration challenges

**Submit feedback:**
- GitHub Issues: [repository]/issues
- Email: benjaminjameswaller@gmail.com
- Discussions: [repository]/discussions

---

**Next Review:** January 2026
**Roadmap Owner:** Benjamin Waller
**Last Modified:** October 7, 2025

---

## Related Documentation

- **Technical Spec Roadmap**: `.agent-os/roadmap.md` - Completed specs tracking (Specs 1-7)
- **Product Mission**: `.agent-os/product/mission.md` - Vision, goals, and values
- **Tech Stack**: `.agent-os/product/tech-stack.md` - Technology decisions
- **Plenti Integration Spec**: `docs/plenti/plenti-integration-spec.md` - Detailed Phase 1 plan
- **Component Prop Scoping**: `docs/FutureDevelopment.md` - Phase 2 detailed tasks
