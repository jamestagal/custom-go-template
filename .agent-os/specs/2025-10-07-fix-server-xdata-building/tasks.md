# Spec Tasks

These are the tasks to be completed for the spec detailed in @.agent-os/specs/2025-10-07-fix-server-xdata-building/spec.md

> Created: 2025-10-07
> Status: Ready for Implementation

## Overview

This tasks list implements the fix for Bug #1 (Server Manually Builds x-data). The goal is to remove manual x-data building from the server and use the proper rendering pipeline (renderer.Render with transformer's alpineDataFormatter).

**Estimated Duration**: 2-4 hours
**Priority**: High (Blocks function support)

## Tasks

- [ ] 1. Refactor Server Route Handlers
  - [ ] 1.1 Write unit tests for the new renderTemplate function (cmd/server/main_test.go)
  - [ ] 1.2 Create unified renderTemplate function that uses renderer.Render (cmd/server/main.go:~200)
  - [ ] 1.3 Update rootHandler to use new renderTemplate function (cmd/server/main.go:~85)
  - [ ] 1.4 Update componentHandler to use new renderTemplate function (cmd/server/main.go:~115)
  - [ ] 1.5 Update testHandler to use new renderTemplate function (cmd/server/main.go:~145)
  - [ ] 1.6 Remove obsolete buildXData helper function (cmd/server/main.go:~250)
  - [ ] 1.7 Remove obsolete buildComponentXData helper function (cmd/server/main.go:~280)
  - [ ] 1.8 Verify all unit tests pass

- [ ] 2. Verify Transformer Integration
  - [ ] 2.1 Check that renderer.Render calls transformer.Transform (renderer/render.go:~50)
  - [ ] 2.2 Verify alpineDataFormatter is called during transformation (transformer/alpine.go:~150)
  - [ ] 2.3 Verify alpineDataFormatter correctly handles props, variables, and functions (transformer/alpine.go:~160-200)
  - [ ] 2.4 Add debug logging to track x-data building (optional, can remove after verification)
  - [ ] 2.5 Test with simple function examples (create minimal test template with 1-2 functions)
  - [ ] 2.6 Verify functions appear in x-data object with correct syntax
  - [ ] 2.7 Verify all transformer tests pass

- [ ] 3. Restore Functions to Test File
  - [ ] 3.1 Add getGreeting function to comprehensive-simple.html fence section (examples/pages/comprehensive-simple.html:~8)
  - [ ] 3.2 Add formatPrice function to comprehensive-simple.html fence section (examples/pages/comprehensive-simple.html:~12)
  - [ ] 3.3 Use getGreeting function in template body Section 1 (examples/pages/comprehensive-simple.html:~35)
  - [ ] 3.4 Use formatPrice function in template body Section 3 (examples/pages/comprehensive-simple.html:~65)
  - [ ] 3.5 Add Section 6: Functions Tests to template body (examples/pages/comprehensive-simple.html:~120)
  - [ ] 3.6 Verify fence section syntax is correct (no syntax errors)
  - [ ] 3.7 Verify file parses without errors

- [ ] 4. Integration Testing and Verification
  - [ ] 4.1 Build the server: `go build cmd/server/main.go`
  - [ ] 4.2 Run the development server: `go run cmd/server/main.go`
  - [ ] 4.3 Test comprehensive-simple page at http://localhost:3000/test/comprehensive-simple
  - [ ] 4.4 Check browser console for JavaScript errors
  - [ ] 4.5 Verify x-data syntax in page source (view source, check x-data attribute)
  - [ ] 4.6 Verify functions are included in x-data object
  - [ ] 4.7 Verify function calls work correctly in browser (check rendered output)
  - [ ] 4.8 Run full test suite: `go test ./... -v`
  - [ ] 4.9 Performance verification: Check no significant performance regression
  - [ ] 4.10 Update CLAUDE.md if architecture changes require documentation updates

## Success Criteria

- [ ] All route handlers use renderer.Render instead of manual x-data building
- [ ] Functions appear correctly in x-data object
- [ ] comprehensive-simple.html displays function results correctly
- [ ] All existing tests pass
- [ ] No browser console errors
- [ ] x-data object in page source has correct JavaScript syntax
- [ ] Server code is cleaner and follows proper architecture

## Notes

- Follow TDD principles: Write tests before implementation
- Each task should end with verification step
- Keep manual buildXData functions until new renderTemplate is proven working
- Test incrementally - don't break working functionality
- Reference spec at @.agent-os/specs/2025-10-07-fix-server-xdata-building/spec.md for detailed requirements
