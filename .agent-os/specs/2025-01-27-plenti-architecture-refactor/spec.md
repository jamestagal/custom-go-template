# Specification: Plenti Architecture Refactor

**Date:** 2025-01-27
**Status:** Draft
**Priority:** High
**Complexity:** High

## Table of Contents
1. [Objective](#objective)
2. [Background](#background)
3. [Current State vs Target State](#current-state-vs-target-state)
4. [Technical Approach](#technical-approach)
5. [Detailed Task Breakdown](#detailed-task-breakdown)
6. [Field Spreading Investigation](#field-spreading-investigation)
7. [Testing Strategy](#testing-strategy)
8. [Success Criteria](#success-criteria)
9. [Risk Mitigation](#risk-mitigation)
10. [Rollback Plan](#rollback-plan)

---

## Objective

Refactor the template engine to match Plenti's architecture for seamless future integration. This includes:

1. **Global wrapper rendering** - All pages render through `layouts/global/html.html`
2. **Component-based page structure** - Convert standalone pages to component-based layouts
3. **Field spreading pattern** - Implement `{...content.fields}` spreading for single content types
4. **Unified routing** - Single route registration path using `renderWithWrapper()`

**Scope:** Convert jim-test page first as proof of concept (incremental, safer approach).

---

## Background

### Critical Lesson Learned

As documented in [CRITICAL_LESSON_ROUTE_REGISTRATION.md](../../../docs/CRITICAL_LESSON_ROUTE_REGISTRATION.md), we discovered a fundamental architectural mismatch:

**Our Current System:**
- **Direct rendering path**: `registerContentRoutes()` → `renderTemplate()` → Standalone HTML pages
- **Wrapper rendering path**: Specific route handlers → `renderWithWrapper()` → Component injection

**Plenti's Architecture:**
- **Single path**: ALL pages → `html.svelte` wrapper → `<svelte:component>` injection
- **Field spreading**: `{...content.fields}` spreads fields into content layout props
- **Content type mapping**: `content/{type}/*.json` → `layouts/content/{type}.html`

### Why Refactor NOW

1. **Future Integration** - Plenti integration will require matching architecture
2. **Consistency** - Single rendering path easier to maintain
3. **Feature Parity** - Unlock Plenti patterns (field spreading, magic variables)
4. **Lessons Learned** - We now understand what went wrong and how to do it right

### What Broke Before

Previous attempt to implement this architecture **broke jim-test page** because:

1. ❌ Changed route registration globally without understanding impact
2. ❌ Jim-test was a **complete HTML page** (with `<!DOCTYPE html>`, `<html>`, `<body>`)
3. ❌ Wrapper tried to inject full HTML page as component inside another HTML structure
4. ❌ Field spreading applied to ALL pages indiscriminately
5. ❌ No testing checkpoints or rollback plan

**Key Insight:** Jim-test needs to be converted to component format FIRST, THEN route registration can change.

---

## Current State vs Target State

### Current State

**Jim-Test Page Structure:**
```
File: layouts/content/jim-test.html
Type: Complete standalone HTML page

<!DOCTYPE html>
<html>
<head>...</head>
<body x-data="...">
  <header>...</header>
  <main>
    <!-- All content sections inline -->
  </main>
  <footer>...</footer>
</body>
</html>
```

**Route Registration:**
```go
// cmd/server/main.go - registerContentRoutes()
http.HandleFunc("/jim-test", func(w http.ResponseWriter, r *http.Request) {
    renderTemplate("layouts/content/jim-test.html", w, r)
})
```

**Content:** No JSON file - all content hardcoded in template

### Target State (Plenti Architecture)

**Jim-Test Page Structure:**
```
File: layouts/content/jim-test.html
Type: Layout component (no HTML wrapper)

---
export let components
---

{for component in components}
  <Component:dynamic name={component.name} {...component.fields} />
{/for}
```

**Route Registration:**
```go
// cmd/server/main.go - registerContentRoutes()
http.HandleFunc("/jim-test", func(w http.ResponseWriter, r *http.Request) {
    renderWithWrapper("jim-test", w, r)
})
```

**Content:**
```json
// content/pages/jim-test.json
{
  "path": "/jim-test",
  "title": "Jim Test Showcase",
  "type": "page",
  "components": [
    {"name": "hero", "fields": {...}},
    {"name": "todos", "fields": {...}},
    {"name": "admin_panel", "fields": {...}},
    // ... all sections as components
  ]
}
```

**Rendering Flow:**
```
Route: /jim-test
  ↓
renderWithWrapper("jim-test", ...)
  ↓
Load: layouts/global/html.html (wrapper)
  ↓
Inject: jim-test layout as <Component:dynamic>
  ↓
Jim-test loops through components array
  ↓
Each component rendered with its fields
  ↓
Output: Full HTML with proper structure
```

---

## Technical Approach

### Phased Migration Strategy

**Phase 1: Convert Jim-Test to Component-Based** (This Spec)
- Create JSON content file with components array
- Extract sections into individual components (if needed)
- Update jim-test.html to use component loop pattern
- Test thoroughly before proceeding

**Phase 2: Update Route Registration**
- Change registerContentRoutes() to use renderWithWrapper()
- Add safeguards and validation
- Test ALL registered routes

**Phase 3: Implement Field Spreading**
- Investigate Plenti's exact field spreading patterns
- Implement for single content types only
- Test news page and other single-type pages

**Phase 4: Documentation & Migration Guide**
- Update DEVELOPER_GUIDE.md
- Document migration path for other pages
- Create examples and best practices

### Key Principles

1. **Test After Each Change** - Never batch multiple changes
2. **Maintain Rollback Points** - Git commit after each working phase
3. **Incremental Validation** - Verify existing pages still work
4. **Document As We Go** - Update docs with findings

---

## Detailed Task Breakdown

### Phase 1: Convert Jim-Test to Component-Based

#### Task 1.1: Analyze Current Jim-Test Structure

**File:** `layouts/content/jim-test.html`

**Action:** Document all sections/features in jim-test page

**Sections to identify:**
- Hero section
- Todo list
- Admin panel
- Notifications
- Data tables
- Charts/graphs
- Forms
- Any other showcase features

**Output:** List of sections with their data requirements

**Cognitive Load:** Low (read and document)

---

#### Task 1.2: Create Component Stubs (If Needed)

**Decision Point:** Are the sections already components?

**Check:**
```bash
ls layouts/components/ | grep -E "(hero|todo|admin|notification)"
```

**If components exist:** Use them directly

**If components don't exist:** Two options:
- **Option A:** Create generic wrapper components for each section
- **Option B:** Keep sections inline but structure as "virtual components" in JSON

**Recommendation:** Start with Option B (virtual components in JSON) to minimize changes

**Output:** Strategy decision documented

---

#### Task 1.3: Create content/pages/jim-test.json

**File:** `content/pages/jim-test.json`

**Structure:**
```json
{
  "path": "/jim-test",
  "title": "Jim Test Showcase Page",
  "description": "Demonstration of multiple features and components",
  "type": "page",
  "components": [
    {
      "name": "section_name_1",
      "fields": {
        "title": "Section Title",
        "content": "...",
        // ... all data for this section
      }
    },
    {
      "name": "section_name_2",
      "fields": {
        // ... data for section 2
      }
    }
    // ... all other sections
  ]
}
```

**Key Considerations:**
- Extract ALL hardcoded data from jim-test.html into JSON
- Each section becomes a component entry
- Preserve exact functionality and data

**Testing Checkpoint:**
- JSON file validates (valid JSON syntax)
- Can be loaded via `loader.LoadContentJSON()`
- Fields match what jim-test.html currently uses

---

#### Task 1.4: Update layouts/content/jim-test.html

**Current:** Complete HTML page with inline content

**Target:** Layout component using component loop

**Changes:**

1. **Remove HTML wrapper:**
```html
<!-- REMOVE: -->
<!DOCTYPE html>
<html>
<head>...</head>
<body x-data="...">

<!-- REMOVE: -->
</body>
</html>
```

2. **Add fence section:**
```html
---
export let components
---
```

3. **Replace inline sections with component loop:**
```html
{for component in components}
  <Component:dynamic name={component.name} {...component.fields} />
{/for}
```

**OR** if keeping inline structure:
```html
---
export let components
---

{for component in components}
  {if component.name === 'hero'}
    <section class="hero">
      <h1>{component.fields.title}</h1>
      <!-- ... hero content using component.fields -->
    </section>
  {else if component.name === 'todos'}
    <section class="todos">
      <!-- ... todos content using component.fields -->
    </section>
  {/if}
  <!-- ... other sections -->
{/for}
```

**Recommendation:** Use conditional blocks for now (safer, less refactoring)

**Testing Checkpoint:**
- Template parses without errors
- Exported props recognized by parser
- No syntax errors

---

#### Task 1.5: Test Jim-Test with Direct Rendering (Validation)

**Goal:** Verify JSON and template work BEFORE changing route registration

**Test Steps:**

1. **Create temporary test route:**
```go
// cmd/server/main.go - in main() function
http.HandleFunc("/jim-test-new", func(w http.ResponseWriter, r *http.Request) {
    if err := renderWithWrapper("jim-test", w, r); err != nil {
        log.Printf("Error: %v", err)
        http.Error(w, "Failed to render", http.StatusInternalServerError)
    }
})
```

2. **Start server:**
```bash
go run cmd/server/main.go
```

3. **Test new route:**
```bash
curl -s http://localhost:3333/jim-test-new | head -100
```

4. **Verify in browser:**
- Visit http://localhost:3333/jim-test-new
- Check all sections render
- Check all data displays correctly
- Check Alpine.js functionality works

5. **Compare with original:**
- Visit http://localhost:3333/jim-test (old route, still using renderTemplate)
- Verify identical functionality

**Success Criteria:**
- `/jim-test-new` displays ALL sections
- Content matches jim-test.json data
- Alpine.js reactivity works
- No console errors
- Visual appearance matches original

**If tests pass:** Proceed to Task 1.6

**If tests fail:** Debug and fix before proceeding

---

#### Task 1.6: Switch Jim-Test Route to Wrapper Rendering

**File:** `cmd/server/main.go`

**Current (in registerContentRoutes):**
```go
if routeName == "jim-test" {
    http.HandleFunc("/jim-test", func(w http.ResponseWriter, r *http.Request) {
        renderTemplate(currentFilePath, w, r)
    })
    continue
}
```

**Change to:**
```go
if routeName == "jim-test" {
    http.HandleFunc("/jim-test", func(w http.ResponseWriter, r *http.Request) {
        if err := renderWithWrapper("jim-test", w, r); err != nil {
            log.Printf("Error rendering jim-test: %v", err)
            http.Error(w, "Failed to render page", http.StatusInternalServerError)
        }
    })
    continue
}
```

**Testing Checkpoint:**

1. **Restart server**
2. **Test original route:**
```bash
curl -s http://localhost:3333/jim-test | head -100
```

3. **Verify in browser:**
- Visit http://localhost:3333/jim-test
- ALL sections should display
- All functionality should work

4. **Remove temporary route** (/jim-test-new)

**Success Criteria:**
- `/jim-test` works with wrapper rendering
- No functionality lost
- No visual changes
- Alpine.js works correctly

**Git Checkpoint:**
```bash
git add .
git commit -m "feat: Convert jim-test to component-based Plenti architecture

- Created content/pages/jim-test.json with components array
- Updated layouts/content/jim-test.html to use component loop
- Changed route to use renderWithWrapper instead of renderTemplate
- All sections and functionality preserved
- Tested and verified working

This is Phase 1 of Plenti architecture refactor."
```

---

### Phase 2: Update Route Registration (Future)

**Status:** NOT in this spec - will be separate spec after Phase 1 success

**Overview:**
- Change `registerContentRoutes()` to use `renderWithWrapper()` for ALL routes
- Add detection for content type (collection vs single)
- Handle field spreading appropriately
- Test each registered route

**Dependencies:**
- Phase 1 complete and tested
- Field spreading investigation complete (Phase 3 insights)

---

### Phase 3: Field Spreading Investigation & Implementation (Future)

**Status:** NOT in this spec - requires investigation first

**Investigation Tasks:**

1. **Study Plenti's exact pattern:**
   - Read `plenti/cmd/build/data_source.go` - how createProps works
   - Read capitaltigers `layouts/global/html.svelte` - spreading mechanism
   - Document exact behavior

2. **Understand when spreading applies:**
   - Only single content types (with `fields: {}`)
   - NOT collection types (with `components: []`)
   - How to detect and distinguish

3. **Implement spreading in renderWithWrapper:**
   - Extract fields from `contentData["fields"]`
   - Spread into dataScope for exported props
   - Don't break collection types

4. **Test with news page:**
   - News page should render with actual content
   - Fields should be available as individual props
   - allContent sidebar should work

---

## Field Spreading Investigation

### Questions to Answer

Before implementing field spreading, we need to understand:

1. **When does Plenti spread fields?**
   - Only for single content types?
   - Does it check content structure first?
   - How does it distinguish single vs collection types?

2. **What gets spread?**
   - All fields from `content.fields`?
   - Are nested objects preserved?
   - Are arrays handled differently?

3. **How does spreading interact with export let?**
   - Does layout request specific props via `export let`?
   - Does spreading happen regardless of export let?
   - What if prop isn't in fields?

4. **What about magic variables?**
   - Are `content`, `allContent`, etc. still passed separately?
   - Do they override spread fields?
   - What's the precedence order?

### Investigation Method

**Step 1: Read Plenti Source**
- File: `plenti/cmd/build/data_source.go`
- Function: `createProps` (around line 536)
- Understand exact logic

**Step 2: Test in Real Plenti Project**
- Project: capitaltigers
- Create test page with both patterns (single + collection)
- Inspect generated HTML
- Document actual behavior

**Step 3: Document Findings**
- Create `FIELD_SPREADING_INVESTIGATION.md` in this spec directory
- Document patterns discovered
- Provide examples of input → output

**Step 4: Implement Based on Findings**
- Update spec with exact implementation approach
- Create tests matching Plenti behavior
- Implement in renderWithWrapper

---

## Testing Strategy

### Manual Testing After Each Task

**Principle:** Test immediately after making change, before proceeding

### Test Checklist for Phase 1

After each task in Phase 1, verify:

- [ ] **Server starts without errors**
  ```bash
  go run cmd/server/main.go
  ```

- [ ] **Jim-test page loads**
  ```bash
  curl -s http://localhost:3333/jim-test | wc -l
  # Should return significant line count (>100 lines)
  ```

- [ ] **All sections visible in browser**
  - Visit http://localhost:3333/jim-test
  - Scroll through entire page
  - Verify every section from original is present

- [ ] **Data displays correctly**
  - Check section titles match JSON
  - Check content matches JSON
  - Check no placeholder text like "undefined"

- [ ] **Alpine.js functionality works**
  - Open browser console (F12)
  - Check for errors (should be none)
  - Test interactive features (if any)
  - Verify x-data is present in body tag

- [ ] **Other routes still work**
  ```bash
  curl -s http://localhost:3333/ | head -50
  curl -s http://localhost:3333/store-demo | head -50
  curl -s http://localhost:3333/pages | head -50
  ```

### Regression Testing

After Phase 1 complete, verify these pages still work:

- [ ] Home page: http://localhost:3333/
- [ ] Store demo: http://localhost:3333/store-demo
- [ ] Pages: http://localhost:3333/pages
- [ ] News: http://localhost:3333/news

### Comparison Testing

**Before making changes:**
1. Visit http://localhost:3333/jim-test
2. Take screenshot
3. Note all sections and features

**After making changes:**
1. Visit http://localhost:3333/jim-test
2. Take screenshot
3. Compare with before screenshot
4. Verify identical appearance and functionality

---

## Success Criteria

### Phase 1 Success Criteria

**Must Have:**
- ✅ Jim-test page loads without errors
- ✅ All sections from original page are present
- ✅ Content displays correctly from jim-test.json
- ✅ Alpine.js functionality works
- ✅ Other routes (home, store-demo, pages) still work
- ✅ No console errors in browser
- ✅ Visual appearance matches original

**Nice to Have:**
- ✅ Code is cleaner and more maintainable
- ✅ JSON content is well-structured
- ✅ Documentation updated

### Overall Refactor Success (All Phases)

**Must Have:**
- ✅ ALL pages use renderWithWrapper (single rendering path)
- ✅ Field spreading works for single content types
- ✅ Collection types (like jim-test) work correctly
- ✅ News page displays content properly
- ✅ System matches Plenti architecture patterns
- ✅ No existing functionality broken

**Nice to Have:**
- ✅ Migration guide for converting other pages
- ✅ Best practices documented
- ✅ Examples of both patterns (collection + single)

---

## Risk Mitigation

### Lessons from Previous Attempt

Based on [CRITICAL_LESSON_ROUTE_REGISTRATION.md](../../../docs/CRITICAL_LESSON_ROUTE_REGISTRATION.md):

**Risk 1: Breaking existing pages**

**Mitigation:**
- ✅ Convert pages to component format FIRST
- ✅ Test each page individually BEFORE changing route registration
- ✅ Use temporary routes for testing new rendering
- ✅ Never change route registration until page is ready

**Risk 2: Scope creep**

**Mitigation:**
- ✅ Start with single page (jim-test only)
- ✅ Create separate specs for each phase
- ✅ Don't attempt multiple changes simultaneously
- ✅ Document scope boundaries clearly

**Risk 3: Inadequate testing**

**Mitigation:**
- ✅ Test after EVERY single task
- ✅ Manual browser testing required
- ✅ Compare before/after screenshots
- ✅ Test all routes, not just changed route

**Risk 4: No rollback plan**

**Mitigation:**
- ✅ Git commit after each successful task
- ✅ Document exact reversion steps
- ✅ Keep original jim-test.html as jim-test.html.backup
- ✅ Tag commits for easy rollback

### Specific Risks for Phase 1

**Risk: JSON structure doesn't match template needs**

**Mitigation:**
- Create JSON incrementally, testing as we go
- Validate JSON loads correctly before updating template
- Use existing component patterns where possible

**Risk: Component loop doesn't handle all section types**

**Mitigation:**
- Start with conditional blocks (if/else if) instead of actual components
- Convert to real components later (Phase 2 or 3)
- Test each section type individually

**Risk: Field extraction breaks**

**Mitigation:**
- Don't implement field spreading in Phase 1
- Use collection type (components array) for jim-test
- Field spreading only in Phase 3 after investigation

---

## Rollback Plan

### If Phase 1 Fails

**Immediate Rollback:**

1. **Revert route registration:**
```go
// cmd/server/main.go - registerContentRoutes()
http.HandleFunc("/jim-test", func(w http.ResponseWriter, r *http.Request) {
    renderTemplate(currentFilePath, w, r)  // Back to direct rendering
})
```

2. **Restore original jim-test.html:**
```bash
cp layouts/content/jim-test.html.backup layouts/content/jim-test.html
```

3. **Remove JSON file:**
```bash
rm content/pages/jim-test.json
```

4. **Restart server and verify:**
```bash
go run cmd/server/main.go
# Visit http://localhost:3333/jim-test
# Should work exactly as before
```

### Git Rollback

**If committed but need to undo:**

```bash
# View recent commits
git log --oneline -5

# Revert specific commit
git revert <commit-hash>

# Or reset to previous commit (if not pushed)
git reset --hard <previous-commit-hash>
```

### Prevention Through Backup

**Before starting Phase 1:**

```bash
# Create backup of original jim-test.html
cp layouts/content/jim-test.html layouts/content/jim-test.html.backup

# Create feature branch
git checkout -b plenti-refactor-jim-test

# Commit clean state
git add .
git commit -m "checkpoint: Before Phase 1 - jim-test conversion"
```

---

## Timeline and Dependencies

### Phase 1 Timeline (This Spec)

**Estimated Time:** 4-6 hours

- Task 1.1: Analyze structure (30 min)
- Task 1.2: Component strategy (15 min)
- Task 1.3: Create JSON file (60-90 min)
- Task 1.4: Update template (45-60 min)
- Task 1.5: Test with temp route (30-45 min)
- Task 1.6: Switch route (15 min)
- Testing and verification (60 min)
- Documentation (30 min)

### Dependencies

**Phase 1 has no dependencies** - can start immediately

**Phase 2 depends on:**
- Phase 1 complete and tested
- Jim-test working with wrapper rendering

**Phase 3 depends on:**
- Field spreading investigation complete
- Clear understanding of Plenti patterns
- Phase 1 success (proven wrapper rendering works)

---

## Documentation Updates

### Files to Update After Phase 1

**1. CLAUDE.md**
- Add section on Plenti architecture patterns
- Document component-based page structure
- Update examples to show JSON + template pattern

**2. DEVELOPER_GUIDE.md** (create if doesn't exist)
- Document how to create component-based pages
- Show JSON structure for collection types
- Explain content/pages/*.json → layouts/content/*.html mapping

**3. README.md**
- Add example of component-based page
- Show jim-test as reference implementation

### New Documentation to Create

**MIGRATION_GUIDE.md**
- How to convert standalone pages to component-based
- Step-by-step process
- Examples and common patterns

**FIELD_SPREADING_INVESTIGATION.md** (for Phase 3)
- Plenti's exact behavior
- Test cases and results
- Implementation recommendations

---

## Validation

### Pre-Implementation Checklist

Before starting Phase 1, verify:

- [ ] CRITICAL_LESSON_ROUTE_REGISTRATION.md has been read and understood
- [ ] Current jim-test page works correctly (baseline established)
- [ ] Git branch created for this work
- [ ] Backup of jim-test.html created
- [ ] Context budget sufficient (>40% remaining recommended)

### Post-Implementation Checklist

After Phase 1 complete, verify:

- [ ] All success criteria met
- [ ] All tests passing
- [ ] Documentation updated
- [ ] Git commits made with clear messages
- [ ] Screenshots/evidence of working page
- [ ] Rollback plan tested (optional but recommended)

---

## Questions and Decisions

### Resolved

✅ **Scope:** Convert jim-test only (not all pages)
✅ **Structure:** Use collection type with components array
✅ **Migration:** Phased approach with testing checkpoints
✅ **Field Spreading:** Only for single content types, investigate Plenti patterns precisely
✅ **Testing:** Manual testing after each phase

### Open Questions

❓ **Component Strategy:** Should sections be real components or conditional blocks?
- **Recommendation:** Start with conditional blocks (less refactoring)
- **Decision needed before Task 1.4**

❓ **JSON Organization:** How granular should component entries be?
- **Recommendation:** One entry per major section
- **Can be refined later**

❓ **Backward Compatibility:** Should we support both rendering paths temporarily?
- **Recommendation:** No, full conversion is cleaner
- **But keep rollback capability**

---

## References

### Key Documents

- [CRITICAL_LESSON_ROUTE_REGISTRATION.md](../../../docs/CRITICAL_LESSON_ROUTE_REGISTRATION.md)
- [NEWS_EXAMPLE_SUCCESS.md](../../../docs/NEWS_EXAMPLE_SUCCESS.md)
- [CLAUDE.md](../../../CLAUDE.md)

### Plenti Source Files

- `plenti/cmd/build/data_source.go` - Content injection logic
- `capitaltigers/layouts/global/html.svelte` - Field spreading example
- `capitaltigers/layouts/content/news.svelte` - Single content type example

### Our Implementation Files

- `cmd/server/main.go` - Route registration and rendering
- `loader/loader.go` - Content JSON loading
- `renderer/content_injection.go` - Export let prop injection
- `transformer/loops.go` - Build-time loop expansion

---

## Approval and Sign-Off

**Spec Status:** Draft - Ready for Review

**Required Reviews:**
- [ ] Technical approach validated
- [ ] Risk mitigation sufficient
- [ ] Testing strategy adequate
- [ ] Timeline realistic

**Approval to Proceed:**
- [ ] User approves scope and approach
- [ ] Context budget sufficient
- [ ] Baseline established (current jim-test works)

---

## Implementation Notes

*This section will be updated during implementation with findings, issues, and solutions.*

---

**End of Specification**
