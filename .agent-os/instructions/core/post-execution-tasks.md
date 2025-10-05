---
description: Rules to finish off and deliver to user set of tasks that have been completed using Agent OS
globs:
alwaysApply: false
version: 1.0
encoding: UTF-8
---

# Task Execution Rules

## Overview

Follow these steps to mark your progress updates, create a recap, and deliver the final report to the user.

<pre_flight_check>
  EXECUTE: @.agent-os/instructions/meta/pre-flight.md
</pre_flight_check>

<process_flow>

<step number="1" subagent="test-runner" name="test_suite_verification">

### Step 1: Run All Tests

Use the test-runner subagent to run the ALL tests in the application's test suite to ensure no regressions and fix any failures until all tests pass.

<instructions>
  ACTION: Use test-runner subagent
  REQUEST: "Run the full test suite"
  WAIT: For test-runner analysis
  PROCESS: Fix any reported failures
  REPEAT: Until all tests pass
</instructions>

<test_execution>
  <order>
    1. Run entire test suite
    2. Fix any failures
  </order>
  <requirement>100% pass rate</requirement>
</test_execution>

<failure_handling>
  <action>troubleshoot and fix</action>
  <priority>before proceeding</priority>
</failure_handling>

</step>

<step number="2" subagent="git-workflow" name="git_workflow">

### Step 2: Git Workflow

Use the git-workflow subagent to create git commit, push to GitHub, and create pull request for the implemented features.

<instructions>
  ACTION: Use git-workflow subagent
  REQUEST: "Complete git workflow for [SPEC_NAME] feature:
            - Spec: [SPEC_FOLDER_PATH]
            - Changes: All modified files
            - Target: main branch
            - Description: [SUMMARY_OF_IMPLEMENTED_FEATURES]"
  WAIT: For workflow completion
  PROCESS: Save PR URL for summary
</instructions>

<commit_process>
  <commit>
    <message>descriptive summary of changes</message>
    <format>conventional commits if applicable</format>
  </commit>
  <push>
    <target>spec branch</target>
    <remote>origin</remote>
  </push>
  <pull_request>
    <title>descriptive PR title</title>
    <description>functionality recap</description>
  </pull_request>
</commit_process>

</step>

<step number="3" subagent="project-manager" name="tasks_list_check">

### Step 3: Tasks Completion Verification

Use the project-manager subagent to read the current spec's tasks.md file and verify that all tasks have been properly marked as complete with [x] or documented with blockers.

<instructions>
  ACTION: Use project-manager subagent
  REQUEST: "Verify that all tasks have been marked with their outcome:
            - Read [SPEC_FOLDER_PATH]/tasks.md
            - Check all tasks are marked complete with [x] or (in rare cases) a documented blocking issue."
  WAIT: For task verification analysis
  PROCESS: Update task status as needed
</instructions>

</step>

<step number="4" subagent="project-manager" name="roadmap_progress_check">

### Step 4: Roadmap Progress Update (conditional)

Use the project-manager subagent to read @.agent-os/product/roadmap.md and mark roadmap items as complete with [x] ONLY IF the executed tasks have completed any roadmap item(s) and the spec completes that item.

<conditional_execution>
  <preliminary_check>
    EVALUATE: Did executed tasks complete any roadmap item(s)?
    IF NO:
      SKIP this entire step
      PROCEED to step 5
    IF YES:
      CONTINUE with roadmap check
  </preliminary_check>
</conditional_execution>

<roadmap_criteria>
  <update_when>
    - spec fully implements roadmap feature
    - all related tasks completed
    - tests passing
  </update_when>
</roadmap_criteria>

<instructions>
  ACTION: First evaluate if roadmap check is needed
      SKIP: If tasks clearly don't complete roadmap items
  EVALUATE: If current spec completes roadmap goals
  UPDATE: Mark roadmap items complete with [x] if applicable
</instructions>

</step>

<step number="5" name="final_cognitive_load_validation">

### Step 5: Final Cognitive Load Validation

Run comprehensive cognitive load validation across ALL files modified during the entire spec implementation to ensure code quality standards.

<validation_scope>
  <scan_all_modified_files>
    IDENTIFY: All files changed in commits for this spec
    SCAN: Each file for cognitive load violations
    AGGREGATE: Violations across all tasks
    CALCULATE: Total cognitive load score per file
  </scan_all_modified_files>
</validation_scope>

<aggregate_report_format>
  ## 📊 Final Cognitive Load Validation Report

  **Spec**: [spec name from spec-lite.md]
  **Date**: [completion date]
  **Files Modified**: [total count]
  **Total Lines Changed**: [+additions -deletions]

  ### Summary by Pattern Category

  **Error Handling**:
  - GO-ERROR-CONTEXT: X violations (Y auto-fixed, Z manual)
  - Overall: ✅ PASS | ⚠️ WARNING | ❌ FAIL

  **Resource Management**:
  - GO-DEFER-LOOP: X violations (Y auto-fixed, Z manual)
  - GO-SLICE-PREALLOC: X violations (Y auto-fixed, Z manual)
  - GO-MAP-CONCURRENT: X violations (Y auto-fixed, Z manual)
  - Overall: ✅ PASS | ⚠️ WARNING | ❌ FAIL

  **GoFast Philosophy**:
  - GOFAST-SIMPLE-DI: X violations (Y auto-fixed, Z manual)
  - GOFAST-STRATEGY-PATTERN: X violations (Y auto-fixed, Z manual)
  - GOFAST-ERROR-VALUES: X violations (Y auto-fixed, Z manual)
  - GOFAST-MINIMAL-DEPS: X violations (Y auto-fixed, Z manual)
  - GOFAST-EXPLICIT-CONFIG: X violations (Y auto-fixed, Z manual)
  - Overall: ✅ PASS | ⚠️ WARNING | ❌ FAIL

  **Svelte/SvelteKit** (if applicable):
  - SVELTE-STORE-LOOP: X violations (Y auto-fixed, Z manual)
  - SVELTEKIT-WATERFALL: X violations (Y auto-fixed, Z manual)
  - Overall: ✅ PASS | ⚠️ WARNING | ❌ FAIL

  ### Cognitive Load Score by File

  | File | Score | Status | Issues |
  |------|-------|--------|--------|
  | parser/html.go | 8/30 | ✅ PASS | None |
  | transformer/loops.go | 12/30 | ✅ PASS | 2 fixed |
  | [other files] | X/30 | status | details |

  **Average Score**: X/30 across Y files
  **Max Score**: X/30 (file: [filename])

  ### Auto-Fixes Applied

  **Total Auto-Fixes**: [count]

  [IF auto-fixes were applied:]
  - Added error context wrapping: X locations
  - Preallocated slices: X locations
  - Fixed store references: X locations
  - Extracted defer from loops: X locations
  - [Other fixes]

  **Files Modified by Auto-Fix**:
  - `file1.go`: 3 fixes
  - `file2.go`: 2 fixes

  [IF no auto-fixes needed:]
  ✅ No auto-fixes required - code already follows all patterns

  ### Manual Review Required

  [IF manual reviews needed:]
  ⚠️ The following issues require developer attention:

  **High Priority** (blocking patterns):
  - GO-MAP-CONCURRENT in `cache.go:78-92` - Add mutex protection
  - GO-DEFER-LOOP in `processor.go:145-160` - Extract to separate function

  **Medium Priority** (warnings):
  - GOFAST-EXPLICIT-CONFIG in `config.go:23` - Extract magic number to constant
  - GO-NIL-CHECK in `validator.go:56` - Use len() instead of nil check

  [IF no manual review needed:]
  ✅ No manual review required

  ### Quality Gate Status

  [IF all files score ≤ 15:]
  ✅ **PASSED** - All files meet cognitive load standards
  - Ready for code review
  - No refactoring required

  [IF any file scores 16-30:]
  ⚠️ **WARNING** - Some files could benefit from refactoring
  - Recommend addressing warnings before merge
  - Not blocking but impacts maintainability

  [IF any file scores > 30:]
  ❌ **BLOCKED** - Refactoring required before proceeding
  - Cannot proceed until violations are resolved
  - See "Manual Review Required" section above

  ### Recommendations

  [IF score > 15:]
  Consider these improvements for future iterations:
  1. [Specific recommendation based on patterns found]
  2. [Specific recommendation based on patterns found]

  [IF score ≤ 15:]
  ✅ Code quality excellent - no recommendations

  ### Metrics

  - **Total Patterns Checked**: [count]
  - **Violations Per File Average**: X
  - **Auto-Fix Success Rate**: Y%
  - **Time Saved by Auto-Fixes**: ~[estimate] minutes
</aggregate_report_format>

<instructions>
  ACTION: Scan all files modified in this spec's commits
  IDENTIFY: Use git diff or git log to find changed files
  CHECK: Each file against patterns in `.agent-os/standards/cognitive-load/foundational-patterns.md`
  REFERENCE: Thresholds from `.agent-os/standards/cognitive-load/config.yml`
  CALCULATE: Aggregate statistics across entire spec
  AUTO-FIX: Apply any remaining fixable patterns
  REPORT: Comprehensive validation summary
  BLOCK: If any file scores > 30 (quality gate failure)
  EDUCATE: Provide actionable recommendations
</instructions>

<validation_workflow>
  <step_1>Git diff to identify all modified files</step_1>
  <step_2>Read each modified file</step_2>
  <step_3>Check against Go/Svelte patterns as applicable</step_3>
  <step_4>Calculate score per file</step_4>
  <step_5>Apply auto-fixes where enabled</step_5>
  <step_6>Aggregate statistics</step_6>
  <step_7>Generate comprehensive report</step_7>
</validation_workflow>

<quality_gate_criteria>
  <pass>
    - All files score ≤ 15/30
    - No critical violations (GO-DEFER-LOOP, GO-MAP-CONCURRENT, SVELTE-STORE-LOOP)
    - All auto-fixable patterns resolved
  </pass>
  <warning>
    - Some files score 16-30/30
    - Non-critical violations present
    - Recommend addressing before merge
  </warning>
  <fail>
    - Any file scores > 30/30
    - Critical violations present
    - Must refactor before proceeding
  </fail>
</quality_gate_criteria>

<validation_reference>
  <patterns_doc>`.agent-os/standards/cognitive-load/foundational-patterns.md`</patterns_doc>
  <config_file>`.agent-os/standards/cognitive-load/config.yml`</config_file>
  <quick_ref>`.agent-os/standards/cognitive-load/quick-reference.md`</quick_ref>
</validation_reference>

</step>

<step number="6" subagent="project-manager" name="document_recap">

### Step 6: Create Recap Document

Use the project-manager subagent to create a recap document in .agent-os/recaps/ folder that summarizes what was built for this spec.

<instructions>
  ACTION: Use project-manager subagent
  REQUEST: "Create recap document for current spec:
            - Create file: .agent-os/recaps/[SPEC_FOLDER_NAME].md
            - Use template format with completed features summary
            - Include context from spec-lite.md
            - Document: [SPEC_FOLDER_PATH]"
  WAIT: For recap document creation
  PROCESS: Verify file is created with proper content
</instructions>

<recap_template>
  # [yyyy-mm-dd] Recap: Feature Name

  This recaps what was built for the spec documented at .agent-os/specs/[spec-folder-name]/spec.md.

  ## Recap

  [1 paragraph summary plus short bullet list of what was completed]

  ## Context

  [Copy the summary found in spec-lite.md to provide concise context of what the initial goal for this spec was]
</recap_template>

<file_creation>
  <location>.agent-os/recaps/</location>
  <naming>[SPEC_FOLDER_NAME].md</naming>
  <format>markdown with yaml frontmatter if needed</format>
</file_creation>

<content_requirements>
  <summary>1 paragraph plus bullet points</summary>
  <context>from spec-lite.md summary</context>
  <reference>link to original spec</reference>
</content_requirements>

</step>

<step number="7" subagent="project-manager" name="completion_summary">

### Step 7: Completion Summary

Use the project-manager subagent to create a structured summary message with emojis showing what was done, any issues, testing instructions, and PR link.

<summary_template>
  ## ✅ What's been done

  1. **[FEATURE_1]** - [ONE_SENTENCE_DESCRIPTION]
  2. **[FEATURE_2]** - [ONE_SENTENCE_DESCRIPTION]

  ## ⚠️ Issues encountered

  [ONLY_IF_APPLICABLE]
  - **[ISSUE_1]** - [DESCRIPTION_AND_REASON]

  ## 👀 Ready to test in browser

  [ONLY_IF_APPLICABLE]
  1. [STEP_1_TO_TEST]
  2. [STEP_2_TO_TEST]

  ## 📦 Pull Request

  View PR: [GITHUB_PR_URL]
</summary_template>

<summary_sections>
  <required>
    - functionality recap
    - pull request info
  </required>
  <conditional>
    - issues encountered (if any)
    - testing instructions (if testable in browser)
  </conditional>
</summary_sections>

<instructions>
  ACTION: Create comprehensive summary
  INCLUDE: All required sections
  ADD: Conditional sections if applicable
  FORMAT: Use emoji headers for scannability
</instructions>

</step>

<step number="8" subagent="project-manager" name="completion_notification">

### Step 8: Task Completion Notification

Use the project-manager subagent to play a system sound to alert the user that tasks are complete.

<notification_command>
  afplay /System/Library/Sounds/Glass.aiff
</notification_command>

<instructions>
  ACTION: Play completion sound
  PURPOSE: Alert user that task is complete
</instructions>

</step>

</process_flow>

<post_flight_check>
  EXECUTE: @.agent-os/instructions/meta/post-flight.md
</post_flight_check>
