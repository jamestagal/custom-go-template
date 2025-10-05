---
description: Rules to execute a task and its sub-tasks using Agent OS
globs:
alwaysApply: false
version: 1.0
encoding: UTF-8
---

# Task Execution Rules

## Overview

Execute a specific task along with its sub-tasks systematically following a TDD development workflow.

<pre_flight_check>
  EXECUTE: @.agent-os/instructions/meta/pre-flight.md
</pre_flight_check>


<process_flow>

<step number="1" name="task_understanding">

### Step 1: Task Understanding

Read and analyze the given parent task and all its sub-tasks from tasks.md to gain complete understanding of what needs to be built.

<task_analysis>
  <read_from_tasks_md>
    - Parent task description
    - All sub-task descriptions
    - Task dependencies
    - Expected outcomes
  </read_from_tasks_md>
</task_analysis>

<instructions>
  ACTION: Read the specific parent task and all its sub-tasks
  ANALYZE: Full scope of implementation required
  UNDERSTAND: Dependencies and expected deliverables
  NOTE: Test requirements for each sub-task
</instructions>

</step>

<step number="2" name="technical_spec_review">

### Step 2: Technical Specification Review

Search and extract relevant sections from technical-spec.md to understand the technical implementation approach for this task.

<selective_reading>
  <search_technical_spec>
    FIND sections in technical-spec.md related to:
    - Current task functionality
    - Implementation approach for this feature
    - Integration requirements
    - Performance criteria
  </search_technical_spec>
</selective_reading>

<instructions>
  ACTION: Search technical-spec.md for task-relevant sections
  EXTRACT: Only implementation details for current task
  SKIP: Unrelated technical specifications
  FOCUS: Technical approach for this specific feature
</instructions>

</step>

<step number="3" subagent="context-fetcher" name="best_practices_review">

### Step 3: Best Practices Review

Use the context-fetcher subagent to retrieve relevant sections from @.agent-os/standards/best-practices.md that apply to the current task's technology stack and feature type.

<selective_reading>
  <search_best_practices>
    FIND sections relevant to:
    - Task's technology stack
    - Feature type being implemented
    - Testing approaches needed
    - Code organization patterns
  </search_best_practices>
</selective_reading>

<instructions>
  ACTION: Use context-fetcher subagent
  REQUEST: "Find best practices sections relevant to:
            - Task's technology stack: [CURRENT_TECH]
            - Feature type: [CURRENT_FEATURE_TYPE]
            - Testing approaches needed
            - Code organization patterns"
  PROCESS: Returned best practices
  APPLY: Relevant patterns to implementation
</instructions>

</step>

<step number="4" subagent="context-fetcher" name="code_style_review">

### Step 4: Code Style Review

Use the context-fetcher subagent to retrieve relevant code style rules from @.agent-os/standards/code-style.md for the languages and file types being used in this task.

<selective_reading>
  <search_code_style>
    FIND style rules for:
    - Languages used in this task
    - File types being modified
    - Component patterns being implemented
    - Testing style guidelines
  </search_code_style>
</selective_reading>

<instructions>
  ACTION: Use context-fetcher subagent
  REQUEST: "Find code style rules for:
            - Languages: [LANGUAGES_IN_TASK]
            - File types: [FILE_TYPES_BEING_MODIFIED]
            - Component patterns: [PATTERNS_BEING_IMPLEMENTED]
            - Testing style guidelines"
  PROCESS: Returned style rules
  APPLY: Relevant formatting and patterns
</instructions>

</step>

<step number="5" name="task_execution">

### Step 5: Task and Sub-task Execution

Execute the parent task and all sub-tasks in order using test-driven development (TDD) approach.

<typical_task_structure>
  <first_subtask>Write tests for [feature]</first_subtask>
  <middle_subtasks>Implementation steps</middle_subtasks>
  <final_subtask>Verify all tests pass</final_subtask>
</typical_task_structure>

<execution_order>
  <subtask_1_tests>
    IF sub-task 1 is "Write tests for [feature]":
      - Write all tests for the parent feature
      - Include unit tests, integration tests, edge cases
      - Run tests to ensure they fail appropriately
      - Mark sub-task 1 complete
  </subtask_1_tests>

  <middle_subtasks_implementation>
    FOR each implementation sub-task (2 through n-1):
      - Implement the specific functionality
      - Make relevant tests pass
      - Update any adjacent/related tests if needed
      - Refactor while keeping tests green
      - Mark sub-task complete
  </middle_subtasks_implementation>

  <final_subtask_verification>
    IF final sub-task is "Verify all tests pass":
      - Run entire test suite
      - Fix any remaining failures
      - Ensure no regressions
      - Mark final sub-task complete
  </final_subtask_verification>
</execution_order>

<test_management>
  <new_tests>
    - Written in first sub-task
    - Cover all aspects of parent feature
    - Include edge cases and error handling
  </new_tests>
  <test_updates>
    - Made during implementation sub-tasks
    - Update expectations for changed behavior
    - Maintain backward compatibility
  </test_updates>
</test_management>

<instructions>
  ACTION: Execute sub-tasks in their defined order
  RECOGNIZE: First sub-task typically writes all tests
  IMPLEMENT: Middle sub-tasks build functionality
  VERIFY: Final sub-task ensures all tests pass
  UPDATE: Mark each sub-task complete as finished
</instructions>

</step>

<step number="6" subagent="test-runner" name="task_test_verification">

### Step 6: Task-Specific Test Verification

Use the test-runner subagent to run and verify only the tests specific to this parent task (not the full test suite) to ensure the feature is working correctly.

<focused_test_execution>
  <run_only>
    - All new tests written for this parent task
    - All tests updated during this task
    - Tests directly related to this feature
  </run_only>
  <skip>
    - Full test suite (done later in execute-tasks.md)
    - Unrelated test files
  </skip>
</focused_test_execution>

<final_verification>
  IF any test failures:
    - Debug and fix the specific issue
    - Re-run only the failed tests
  ELSE:
    - Confirm all task tests passing
    - Ready to proceed
</final_verification>

<instructions>
  ACTION: Use test-runner subagent
  REQUEST: "Run tests for [this parent task's test files]"
  WAIT: For test-runner analysis
  PROCESS: Returned failure information
  VERIFY: 100% pass rate for task-specific tests
  CONFIRM: This feature's tests are complete
</instructions>

</step>

<step number="7" name="task_status_updates">

### Step 7: Mark this task and sub-tasks complete

IMPORTANT: In the tasks.md file, mark this task and its sub-tasks complete by updating each task checkbox to [x].

<update_format>
  <completed>- [x] Task description</completed>
  <incomplete>- [ ] Task description</incomplete>
  <blocked>
    - [ ] Task description
    ⚠️ Blocking issue: [DESCRIPTION]
  </blocked>
</update_format>

<blocking_criteria>
  <attempts>maximum 3 different approaches</attempts>
  <action>document blocking issue</action>
  <emoji>⚠️</emoji>
</blocking_criteria>

<instructions>
  ACTION: Update tasks.md after each task completion
  MARK: [x] for completed items immediately
  DOCUMENT: Blocking issues with ⚠️ emoji
  LIMIT: 3 attempts before marking as blocked
</instructions>

</step>

<step number="8" name="cognitive_load_validation">

### Step 8: Cognitive Load Validation Report

After completing task implementation, validate code quality against cognitive load patterns and report violations found in all modified files.

<validation_requirements>
  <check_patterns>
    FOR Go code (.go files):
      - GO-ERROR-CONTEXT (naked error returns without context)
      - GO-DEFER-LOOP (defer statements inside loops)
      - GO-SLICE-PREALLOC (slices without preallocation)
      - GO-MAP-CONCURRENT (concurrent map access without sync)
      - GO-NIL-CHECK (nil checks instead of len() for slices)
      - GOFAST-SIMPLE-DI (complex DI containers)
      - GOFAST-STRATEGY-PATTERN (over-abstracted patterns)
      - GOFAST-ERROR-VALUES (panic/recover instead of errors)
      - GOFAST-MINIMAL-DEPS (external deps for stdlib tasks)
      - GOFAST-EXPLICIT-CONFIG (hardcoded magic numbers)
      - GOFAST-DTO-PATTERN (sql.Null* in API responses)

    FOR Svelte code (.svelte files):
      - SVELTE-STORE-LOOP (direct store assignments in reactive contexts)
      - SVELTE-INIT-GUARD (missing initialization guards in onMount)

    FOR SvelteKit code (+page.ts, +page.server.ts):
      - SVELTEKIT-SERVER-CLIENT (browser APIs in server files)
      - SVELTEKIT-WATERFALL (data fetching in onMount instead of load)
  </check_patterns>
</validation_requirements>

<scoring_criteria>
  <thresholds>
    - BLOCK: score > 30 (stop and refactor)
    - WARN: score 16-30 (suggest improvements)
    - PASS: score ≤ 15 (good to proceed)
  </thresholds>
  <pattern_scores>
    - GO-MAP-CONCURRENT: 9 points (critical)
    - GO-DEFER-LOOP: 7 points (high)
    - SVELTE-STORE-LOOP: 7 points (high)
    - GO-SLICE-PREALLOC: 4 points (medium)
    - GO-ERROR-CONTEXT: 3 points (low)
  </pattern_scores>
</scoring_criteria>

<report_format>
  ## 📊 Cognitive Load Validation Report

  **Task**: [Task name and number]
  **Files Modified**: [Count and list]
  **Patterns Checked**: [Count of patterns relevant to tech stack]

  ### Violations Found

  [IF violations found:]
  - **GO-ERROR-CONTEXT**: X violations (auto-fixed: Y)
    - Files: [list affected files with line numbers]
  - **GO-DEFER-LOOP**: X violations (blocked/fixed)
    - Files: [list affected files with line numbers]
  - **[Other patterns]**: X violations

  [IF no violations:]
  ✅ No cognitive load violations detected

  ### Cognitive Load Score

  **Total Score**: X/30 (threshold: 30)
  **Status**: [✅ PASS | ⚠️ WARNING | ❌ BLOCKED]

  **Per-File Breakdown**:
  - `file1.go`: X/30
  - `file2.go`: X/30

  ### Auto-Fixes Applied

  [IF auto-fixes applied:]
  - Added error context wrapping with fmt.Errorf (3 locations)
  - Preallocated slices with known capacity (2 locations)
  - [Other fixes]

  [IF no auto-fixes needed:]
  ✅ No auto-fixes required

  ### Manual Review Required

  [IF manual review needed:]
  ⚠️ The following patterns require manual refactoring:
  - GO-DEFER-LOOP in `parser.go:145` - Extract to separate function
  - GO-MAP-CONCURRENT in `cache.go:78` - Add mutex protection

  [IF no manual review needed:]
  ✅ No manual review required

  ### Quality Gate

  [IF score ≤ 15:]
  ✅ **PASSED** - Code meets cognitive load standards

  [IF score 16-30:]
  ⚠️ **WARNING** - Consider refactoring for better maintainability

  [IF score > 30:]
  ❌ **BLOCKED** - Refactoring required before proceeding
</report_format>

<instructions>
  ACTION: Scan all files modified during this task
  CHECK: Against patterns in `.agent-os/standards/cognitive-load/foundational-patterns.md`
  REFERENCE: Thresholds from `.agent-os/standards/cognitive-load/config.yml`
  CALCULATE: Total cognitive load score per file
  AUTO-FIX: Apply fixes for patterns with autofix: true
  REPORT: Detailed violation summary with locations
  BLOCK: If any file scores > 30
  EDUCATE: Provide context for violations found
</instructions>

<pattern_detection_examples>
  <go_error_context>
    DETECT: `return err` without fmt.Errorf wrapper
    AUTO-FIX: Wrap with context: `return fmt.Errorf("functionName: operation failed: %w", err)`
  </go_error_context>

  <go_defer_loop>
    DETECT: defer statement inside for/range loop
    MANUAL: Suggest extracting loop body to separate function
  </go_defer_loop>

  <go_slice_prealloc>
    DETECT: `var slice []Type` followed by append in loop
    AUTO-FIX: Add preallocation: `slice := make([]Type, 0, knownCapacity)`
  </go_slice_prealloc>

  <svelte_store_loop>
    DETECT: Direct store assignment in reactive context
    AUTO-FIX: Add spread operator: `data = { ...store.value }`
  </svelte_store_loop>
</pattern_detection_examples>

<validation_reference>
  <patterns_doc>`.agent-os/standards/cognitive-load/foundational-patterns.md`</patterns_doc>
  <config_file>`.agent-os/standards/cognitive-load/config.yml`</config_file>
  <quick_ref>`.agent-os/standards/cognitive-load/quick-reference.md`</quick_ref>
</validation_reference>

</step>

</process_flow>

<post_flight_check>
  EXECUTE: @.agent-os/instructions/meta/post-flight.md
</post_flight_check>
