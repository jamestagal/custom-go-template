# Spec Requirements Document

> Spec: X-Data Duplication Optimization
> Created: 2025-10-22
**MANDATORY: Use go-backend agent for all Go implementation**

## Overview

Optimize HTML output by eliminating redundant x-data attribute duplication across nested levels, leveraging Alpine.js scope inheritance to reduce HTML payload by 90-95%. This performance optimization addresses the current issue where x-data attributes are duplicated across 4 nested levels (root div, body, components, runtime wrappers), causing 800KB+ of unnecessary data in typical pages.

## User Stories

### Performance-Focused Developer

As a **Plenti site builder**, I want **fast page loads with minimal HTML bloat**, so that **my site performs well and provides a good user experience**.

**Workflow:**
1. Developer builds a Plenti site with multiple components
2. During development, notices large HTML payload sizes (800KB+)
3. After optimization implementation, sees HTML reduced to <80KB
4. Page load time improves by 15-25%
5. Lighthouse performance score increases by 5-10 points

**Problem Solved:** Eliminates performance bottleneck caused by massive data duplication in generated HTML

### Template Engine Maintainer

As the **template engine maintainer**, I want **clean, optimized HTML output that follows Alpine.js best practices**, so that **the engine produces production-ready code by default**.

**Workflow:**
1. Reviews generated HTML structure
2. Identifies 4 layers of duplicate x-data wrappers
3. Implements scope inheritance optimization
4. Validates that Alpine.js reactivity still works correctly
5. Confirms all components inherit scopes properly

**Problem Solved:** Aligns output with Alpine.js best practices (scope inheritance) and eliminates defensive over-wrapping

### End User (Site Visitor)

As a **website visitor**, I want **fast-loading pages**, so that **I can access content quickly without waiting**.

**Workflow:**
1. User visits a Plenti site built with the template engine
2. Browser downloads smaller HTML payload (<80KB instead of 800KB+)
3. Page renders 15-25% faster
4. Alpine.js initializes more quickly (less data to parse)
5. User has better overall experience

**Problem Solved:** Improved site performance translates to better user experience and higher engagement

## Spec Scope

1. **Phase 1: Remove Root Div Wrapper** - Eliminate redundant root-level x-data wrapper added by transformer (25% reduction)

2. **Phase 2: Optimize Component Wrappers** - Implement scope diffing to only add x-data when components introduce NEW variables (60-70% reduction)

3. **Phase 3: Optimize Runtime Wrappers** - Use prop references instead of full data serialization in runtime components (10-15% additional reduction)

4. **Scope Analysis Utilities** - Create scope diffing and comparison utilities to determine when x-data wrappers are needed

5. **Testing & Validation** - Comprehensive test suite to ensure Alpine.js functionality remains intact across all optimization phases

## Out of Scope

- Changing Alpine.js framework or version (staying on 3.x)
- Modifying template syntax or developer-facing API
- Optimizing other aspects of HTML output (CSS, JavaScript bundles)
- Performance improvements unrelated to x-data duplication
- Changes to parser or AST structure (optimization happens in transformer/renderer)

## Expected Deliverable

1. **HTML payload reduced by 90-95%** for x-data attributes - Typical pages go from 800KB to <80KB of x-data bloat

2. **Zero breaking changes** - All existing templates continue to work without modification; Alpine.js reactivity fully functional

3. **Performance metrics improved** - Page load time reduced by 15-25%, Lighthouse performance score increased by 5-10 points
