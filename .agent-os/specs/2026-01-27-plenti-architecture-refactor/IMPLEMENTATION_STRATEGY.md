# Jim-Test Implementation Strategy

**Date:** 2026-10-29
**Status:** Active

## Analysis Summary

### Current Structure

**File:** `layouts/content/jim-test.html`
**Type:** Complete standalone HTML page with imports, fence section, full HTML wrapper

### Identified Sections

1. **Header Component** - Imported from `../global/header.html`
2. **Build Time Display** - Shows `{buildTime}` variable
3. **Greeting Section** - `{salutation} {name}!`
4. **Name Length Conditionals** - Nested if/else-if/else with age check
5. **Age Component Examples** - 3 instances with different props
6. **User Profile Examples** - 3 instances using dynamic component syntax
7. **Animals Loop** - Interactive add/remove with nested conditionals
8. **Advanced Loop Patterns** - Array spread demonstrations
9. **Todos Components** - 2 instances with different start/number props
10. **Notification Examples** - Interactive button grid with conditional display
11. **Footer Component** - Imported from `../components/footer-old.html`

### Data Requirements

**Props (from fence section):**
- `name = "Benjamin"`
- `age = 55`
- `animals = ["dog", "cat", "bird"]`
- `newAnimal = ""`

**Variables (from fence section):**
- `salutation = "Hello"`
- `path = "../components/userprofile.html"`
- `comp = "userprofile"`
- `user1`, `user2`, `user3` (user objects with name, email, role, avatar, joinDate)
- `notifications` (array of notification objects)
- `currentNotification = null`

**Magic Variables:**
- `buildTime` (provided by renderer)

### Component Analysis

**Existing Components Used:**
- `Head` - Global component
- `Header` - Global component
- `Footer-old` - Component
- `Age` - Component
- `Todos` - Component
- `Notification` - Component
- Dynamic UserProfile - Via path syntax

**Decision:** These are REAL components already registered. We will NOT create new components for sections.

## Implementation Strategy

### Strategy: Inline Sections with Virtual Components

**Rationale:**
1. Jim-test is a showcase/demo page, not a production pattern
2. Sections share state (e.g., `currentNotification`, `animals`)
3. Creating real components would require complex prop drilling
4. Virtual components in JSON preserve exact functionality

### JSON Structure Approach

**Pattern:** Create "virtual component" entries for each major section that can be rendered inline:

```json
{
  "path": "/jim-test",
  "title": "Jim Test Showcase",
  "type": "page",
  "components": [
    {
      "name": "greeting",
      "fields": {
        "salutation": "Hello",
        "name": "Benjamin"
      }
    },
    {
      "name": "conditionals",
      "fields": {
        "name": "Benjamin",
        "age": 55
      }
    },
    {
      "name": "age_examples",
      "fields": {
        "name": "Benjamin",
        "age": 55
      }
    },
    {
      "name": "user_profiles",
      "fields": {
        "user1": {...},
        "user2": {...},
        "user3": {...},
        "path": "../components/userprofile.html",
        "comp": "userprofile"
      }
    },
    {
      "name": "animals_loop",
      "fields": {
        "name": "Benjamin",
        "animals": ["dog", "cat", "bird"],
        "newAnimal": ""
      }
    },
    {
      "name": "advanced_loops",
      "fields": {
        "name": "Benjamin",
        "animals": ["dog", "cat", "bird"]
      }
    },
    {
      "name": "todos_examples",
      "fields": {}
    },
    {
      "name": "notifications",
      "fields": {
        "notifications": [...],
        "currentNotification": null
      }
    }
  ]
}
```

### Template Structure Approach

**Pattern:** Use component loop with if/else-if conditionals for each section:

```html
---
export let components
---

{for component in components}
  {if component.name === 'greeting'}
    <h1>{component.fields.salutation} {component.fields.name}!</h1>
  {else if component.name === 'conditionals'}
    <!-- Conditionals section HTML with component.fields.* -->
  {else if component.name === 'age_examples'}
    <div style="margin: 2rem 0;">
      <h2>Age Component Examples</h2>
      <Age name={component.fields.name} age={component.fields.age} />
      <Age name={"Bo"} age={component.fields.age + 50} />
      <Age name={"Baggins"} age={201} />
    </div>
  {else if component.name === 'animals_loop'}
    <!-- Animals loop section with component.fields.* -->
  <!-- ... other sections ... -->
  {/if}
{/for}
```

**Note:** Need to preserve Alpine.js directives like `x-model="newAnimal"` and `onclick` handlers.

## Safety Measures

1. **Backup Created:** `jim-test.html.backup` ✓
2. **Feature Branch:** Already on `plenti-architecture-refactor` ✓
3. **Testing Strategy:**
   - Create temporary `/jim-test-new` route first
   - Compare side-by-side with original
   - Only switch route after verification

## Potential Challenges

### Challenge 1: Build Time Variable
**Issue:** `{buildTime}` is a magic variable provided by renderer, not in JSON
**Solution:** Keep it in template, don't move to JSON. Global wrapper should provide it.

### Challenge 2: Alpine.js Reactivity
**Issue:** Interactive features use Alpine directives (`x-model`, `onclick`)
**Solution:** Preserve exact directive syntax in template, these don't come from JSON

### Challenge 3: Component Imports
**Issue:** Template imports components in fence section
**Solution:** Remove fence imports. Global wrapper handles component resolution.

### Challenge 4: Shared State
**Issue:** Some sections share variables (e.g., `animals` used in multiple sections)
**Solution:** Each component entry gets its own copy of data. May need to consolidate sections.

## Simplified Alternative Strategy

**Alternative:** Instead of many small sections, group by functional area:

```json
{
  "components": [
    {
      "name": "demo_header",
      "fields": {
        "salutation": "Hello",
        "name": "Benjamin",
        "age": 55
      }
    },
    {
      "name": "component_demos",
      "fields": {
        "name": "Benjamin",
        "age": 55,
        "user1": {...},
        "user2": {...},
        "user3": {...}
      }
    },
    {
      "name": "loop_demos",
      "fields": {
        "name": "Benjamin",
        "animals": ["dog", "cat", "bird"],
        "newAnimal": ""
      }
    },
    {
      "name": "interactive_demos",
      "fields": {
        "notifications": [...],
        "currentNotification": null
      }
    }
  ]
}
```

**Recommendation:** Use simplified approach with 4 major sections. Easier to maintain.

## Final Decision

**Strategy Selected:** Simplified 4-section approach
- Fewer conditionals in template
- Logical grouping of related features
- Easier to understand JSON structure
- Less cognitive load

## Next Steps

1. Create `content/pages/jim-test.json` with 4-section structure
2. Update `jim-test.html` template to use component loop
3. Test with temporary route
4. Verify functionality
5. Switch route registration
