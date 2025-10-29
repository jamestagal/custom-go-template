# Spec Tasks

These are the tasks to be completed for the spec detailed in @.agent-os/specs/2025-01-27-plenti-architecture-refactor/spec.md

> Created: 2025-10-27
> Status: Ready for Implementation
**MANDATORY: Use go-backend agent for all Go implementation**

## Tasks

- [ ] 1. Analyze Jim-Test Structure and Create Implementation Strategy
  - [ ] 1.1 Read and document all sections in layouts/content/jim-test.html
  - [ ] 1.2 Identify data requirements for each section
  - [ ] 1.3 Check if sections already exist as components in layouts/components/
  - [ ] 1.4 Decide on component strategy (inline conditionals vs real components)
  - [ ] 1.5 Create backup of original jim-test.html
  - [ ] 1.6 Create feature branch for this work
  - [ ] 1.7 Document implementation strategy
  - [ ] 1.8 **MANDATORY: Use go-backend agent for all Go implementation**

- [ ] 2. Create JSON Content File for Jim-Test
  - [ ] 2.1 Create content/pages/jim-test.json with proper structure
  - [ ] 2.2 Extract all hardcoded data from jim-test.html template
  - [ ] 2.3 Structure data as components array with name/fields pattern
  - [ ] 2.4 Validate JSON syntax is correct
  - [ ] 2.5 Test JSON loads with loader.LoadContentJSON()
  - [ ] 2.6 Verify field names match what template currently uses
  - [ ] 2.7 **MANDATORY: Use go-backend agent for all Go implementation**

- [ ] 3. Update Jim-Test Template to Use Component Loop
  - [ ] 3.1 Remove HTML wrapper (<!DOCTYPE>, <html>, <body> tags)
  - [ ] 3.2 Add fence section with "export let components"
  - [ ] 3.3 Replace inline sections with component loop using conditionals
  - [ ] 3.4 Update all data references to use component.fields.*
  - [ ] 3.5 Preserve Alpine.js directives and functionality
  - [ ] 3.6 Verify template parses without syntax errors
  - [ ] 3.7 **MANDATORY: Use go-backend agent for all Go implementation**

- [ ] 4. Test With Temporary Route Before Switching
  - [ ] 4.1 Add temporary /jim-test-new route using renderWithWrapper
  - [ ] 4.2 Start server and test new route with curl
  - [ ] 4.3 Verify all sections render in browser
  - [ ] 4.4 Check all data displays correctly from JSON
  - [ ] 4.5 Test Alpine.js functionality and check console for errors
  - [ ] 4.6 Compare side-by-side with original /jim-test route
  - [ ] 4.7 Take screenshots for comparison
  - [ ] 4.8 Debug and fix any issues found
  - [ ] 4.9 **MANDATORY: Use go-backend agent for all Go implementation**

- [ ] 5. Switch Jim-Test Route to Wrapper Rendering
  - [ ] 5.1 Update route registration in cmd/server/main.go
  - [ ] 5.2 Change from renderTemplate to renderWithWrapper
  - [ ] 5.3 Restart server and test /jim-test route
  - [ ] 5.4 Verify all sections and functionality preserved
  - [ ] 5.5 Remove temporary /jim-test-new route
  - [ ] 5.6 Test regression: verify home, store-demo, pages still work
  - [ ] 5.7 Git commit with descriptive message
  - [ ] 5.8 Update documentation (CLAUDE.md, create MIGRATION_GUIDE.md)
  - [ ] 5.9 **MANDATORY: Use go-backend agent for all Go implementation**