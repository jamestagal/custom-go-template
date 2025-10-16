package analyzer

import (
	"testing"
)

// TestScopeAnalyzer_TrackLoopVariable tests that loop variables are correctly tracked as runtime-only
// Pattern: Table-Driven Test [Load: 6]
func TestScopeAnalyzer_TrackLoopVariable(t *testing.T) {
	tests := []struct {
		name         string
		loopVars     []string
		testExpr     string
		wantRuntime  bool
		description  string
	}{
		{
			name:        "simple loop variable",
			loopVars:    []string{"component"},
			testExpr:    "component",
			wantRuntime: true,
			description: "direct reference to loop variable should be runtime",
		},
		{
			name:        "loop variable with property",
			loopVars:    []string{"component"},
			testExpr:    "component.name",
			wantRuntime: true,
			description: "property access on loop variable should be runtime",
		},
		{
			name:        "nested loop variables",
			loopVars:    []string{"component", "item"},
			testExpr:    "item.value",
			wantRuntime: true,
			description: "nested loop variables should be runtime",
		},
		{
			name:        "non-loop variable",
			loopVars:    []string{"component"},
			testExpr:    "title",
			wantRuntime: false,
			description: "non-loop variable should not be runtime",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewScopeAnalyzer(nil)

			// Track loop variables
			for _, loopVar := range tt.loopVars {
				analyzer.TrackLoopVariable(loopVar)
			}

			// Test expression
			got := analyzer.IsRuntimeExpression(tt.testExpr)
			if got != tt.wantRuntime {
				t.Errorf("IsRuntimeExpression(%q) = %v, want %v - %s",
					tt.testExpr, got, tt.wantRuntime, tt.description)
			}
		})
	}
}

// TestScopeAnalyzer_TrackContentProp tests content prop tracking
// Pattern: Table-Driven Test [Load: 5]
func TestScopeAnalyzer_TrackContentProp(t *testing.T) {
	tests := []struct {
		name        string
		contentVars []string
		testExpr    string
		wantRuntime bool
	}{
		{
			name:        "content prop is build-time",
			contentVars: []string{"title", "description"},
			testExpr:    "title",
			wantRuntime: false,
		},
		{
			name:        "content prop with property access",
			contentVars: []string{"content"},
			testExpr:    "content.title",
			wantRuntime: false,
		},
		{
			name:        "unknown variable",
			contentVars: []string{"title"},
			testExpr:    "unknown",
			wantRuntime: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewScopeAnalyzer(nil)

			for _, contentVar := range tt.contentVars {
				analyzer.TrackContentProp(contentVar)
			}

			got := analyzer.IsRuntimeExpression(tt.testExpr)
			if got != tt.wantRuntime {
				t.Errorf("IsRuntimeExpression(%q) = %v, want %v",
					tt.testExpr, got, tt.wantRuntime)
			}
		})
	}
}

// TestScopeAnalyzer_AlpineStores tests Alpine store detection
// Pattern: Table-Driven Test [Load: 6]
func TestScopeAnalyzer_AlpineStores(t *testing.T) {
	tests := []struct {
		name        string
		expr        string
		wantRuntime bool
	}{
		{
			name:        "Alpine store reference",
			expr:        "$store.auth.isLoggedIn",
			wantRuntime: true,
		},
		{
			name:        "Alpine store root",
			expr:        "$store.auth",
			wantRuntime: true,
		},
		{
			name:        "Alpine store shorthand",
			expr:        "$auth.isLoggedIn",
			wantRuntime: true,
		},
		{
			name:        "not a store",
			expr:        "auth.isLoggedIn",
			wantRuntime: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewScopeAnalyzer(nil)
			got := analyzer.IsRuntimeExpression(tt.expr)
			if got != tt.wantRuntime {
				t.Errorf("IsRuntimeExpression(%q) = %v, want %v",
					tt.expr, got, tt.wantRuntime)
			}
		})
	}
}

// TestScopeAnalyzer_StringLiterals tests string literal detection
// Pattern: Table-Driven Test [Load: 5]
func TestScopeAnalyzer_StringLiterals(t *testing.T) {
	tests := []struct {
		name        string
		expr        string
		wantRuntime bool
	}{
		{
			name:        "double quoted string",
			expr:        `"ComponentName"`,
			wantRuntime: false,
		},
		{
			name:        "single quoted string",
			expr:        `'ComponentName'`,
			wantRuntime: false,
		},
		{
			name:        "unquoted literal",
			expr:        "ComponentName",
			wantRuntime: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewScopeAnalyzer(nil)
			got := analyzer.IsRuntimeExpression(tt.expr)
			if got != tt.wantRuntime {
				t.Errorf("IsRuntimeExpression(%q) = %v, want %v",
					tt.expr, got, tt.wantRuntime)
			}
		})
	}
}

// TestScopeAnalyzer_MixedExpressions tests complex expressions
// Pattern: Table-Driven Test [Load: 8]
func TestScopeAnalyzer_MixedExpressions(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(*ScopeAnalyzer)
		expr         string
		wantRuntime  bool
		description  string
	}{
		{
			name: "runtime var with build-time prop",
			setup: func(a *ScopeAnalyzer) {
				a.TrackLoopVariable("component")
				a.TrackContentProp("title")
			},
			expr:        "component.name",
			wantRuntime: true,
			description: "expression with runtime variable should be runtime",
		},
		{
			name: "only build-time vars",
			setup: func(a *ScopeAnalyzer) {
				a.TrackContentProp("title")
				a.TrackContentProp("link")
			},
			expr:        "title",
			wantRuntime: false,
			description: "expression with only build-time vars should be build-time",
		},
		{
			name: "store mixed with content prop",
			setup: func(a *ScopeAnalyzer) {
				a.TrackContentProp("title")
			},
			expr:        "$auth.user",
			wantRuntime: true,
			description: "Alpine store always makes expression runtime",
		},
		{
			name: "operator expression",
			setup: func(a *ScopeAnalyzer) {
				a.TrackContentProp("count")
			},
			expr:        "count + 1",
			wantRuntime: true,
			description: "expressions with operators should be runtime for safety",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewScopeAnalyzer(nil)
			tt.setup(analyzer)

			got := analyzer.IsRuntimeExpression(tt.expr)
			if got != tt.wantRuntime {
				t.Errorf("IsRuntimeExpression(%q) = %v, want %v - %s",
					tt.expr, got, tt.wantRuntime, tt.description)
			}
		})
	}
}

// TestScopeAnalyzer_NestedLoops tests nested loop variable tracking
// Pattern: Table-Driven Test [Load: 6]
func TestScopeAnalyzer_NestedLoops(t *testing.T) {
	tests := []struct {
		name        string
		loopVars    []string
		testExpr    string
		wantRuntime bool
	}{
		{
			name:        "outer loop variable",
			loopVars:    []string{"component", "item"},
			testExpr:    "component.name",
			wantRuntime: true,
		},
		{
			name:        "inner loop variable",
			loopVars:    []string{"component", "item"},
			testExpr:    "item.value",
			wantRuntime: true,
		},
		{
			name:        "both loop variables",
			loopVars:    []string{"outer", "inner"},
			testExpr:    "outer.inner.value",
			wantRuntime: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewScopeAnalyzer(nil)

			for _, loopVar := range tt.loopVars {
				analyzer.TrackLoopVariable(loopVar)
			}

			got := analyzer.IsRuntimeExpression(tt.testExpr)
			if got != tt.wantRuntime {
				t.Errorf("IsRuntimeExpression(%q) = %v, want %v",
					tt.testExpr, got, tt.wantRuntime)
			}
		})
	}
}

// TestScopeAnalyzer_ExportedProps tests exported prop tracking (export let)
// Pattern: Table-Driven Test [Load: 5]
func TestScopeAnalyzer_ExportedProps(t *testing.T) {
	tests := []struct {
		name         string
		exportedVars []string
		testExpr     string
		wantRuntime  bool
	}{
		{
			name:         "exported prop is build-time",
			exportedVars: []string{"title", "description"},
			testExpr:     "title",
			wantRuntime:  false,
		},
		{
			name:         "exported prop with property",
			exportedVars: []string{"data"},
			testExpr:     "data.title",
			wantRuntime:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewScopeAnalyzer(nil)

			for _, exportedVar := range tt.exportedVars {
				analyzer.TrackExportedProp(exportedVar)
			}

			got := analyzer.IsRuntimeExpression(tt.testExpr)
			if got != tt.wantRuntime {
				t.Errorf("IsRuntimeExpression(%q) = %v, want %v",
					tt.testExpr, got, tt.wantRuntime)
			}
		})
	}
}

// TestScopeAnalyzer_DataScopeIntegration tests integration with existing dataScope map
// Pattern: Integration Test [Load: 8]
func TestScopeAnalyzer_DataScopeIntegration(t *testing.T) {
	dataScope := map[string]any{
		"title":       "Hello",
		"description": "World",
		"component": map[string]any{
			"name": "Hero2436",
		},
	}

	analyzer := NewScopeAnalyzer(dataScope)

	// Variables in dataScope but not tracked as loop vars should be build-time
	if analyzer.IsRuntimeExpression("title") {
		t.Error("title should be build-time (in dataScope, not tracked as loop var)")
	}

	// Track component as loop variable
	analyzer.TrackLoopVariable("component")

	// Now component should be runtime
	if !analyzer.IsRuntimeExpression("component.name") {
		t.Error("component.name should be runtime after tracking component as loop var")
	}
}

// TestIsRuntimeExpression_LoopVariablesInDataScope tests the critical bug fix
// for detecting loop variables stored in dataScope with nil values.
//
// Pattern: Regression Test [Load: 6]
// Cognitive Load: 6 (setup: 3, assertions: 3)
//
// This test ensures that loop variables stored in dataScope with nil values
// are correctly identified as runtime expressions, even without explicit
// TrackLoopVariable() calls.
//
// Context:
//   - transformLoop in transformer/loops.go adds loop vars with nil values
//   - ScopeAnalyzer must detect these nil-valued entries as runtime markers
//   - Bug: IsRuntimeExpression("component.name") returned false
//   - Fix: Check dataScope for nil values in step 6 of IsRuntimeExpression
func TestIsRuntimeExpression_LoopVariablesInDataScope(t *testing.T) {
	// Setup: dataScope with loop variable (nil value)
	dataScope := map[string]any{
		"component": nil,  // Loop variable marker (set in transformer/loops.go:47)
		"item":      nil,  // Another loop variable
		"title":     "Static Value", // Build-time value (non-nil)
	}

	analyzer := NewScopeAnalyzer(dataScope)

	tests := []struct {
		expr     string
		expected bool
		reason   string
	}{
		{
			expr:     "component.name",
			expected: true,
			reason:   "Loop variable property access should be runtime",
		},
		{
			expr:     "item.field",
			expected: true,
			reason:   "Loop variable property access should be runtime",
		},
		{
			expr:     "component",
			expected: true,
			reason:   "Loop variable itself should be runtime",
		},
		{
			expr:     "title",
			expected: false,
			reason:   "Build-time variable (non-nil value) should be build-time",
		},
		{
			expr:     "unknown",
			expected: false,
			reason:   "Unknown variables default to build-time for backwards compat",
		},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result := analyzer.IsRuntimeExpression(tt.expr)
			if result != tt.expected {
				t.Errorf("IsRuntimeExpression(%q) = %v, want %v - %s",
					tt.expr, result, tt.expected, tt.reason)
			}
		})
	}
}

// TestIsRuntimeExpression_NilValueMarkerPriority tests that nil-valued loop variables
// are detected even when other tracking methods are used.
//
// Pattern: Integration Test [Load: 8]
func TestIsRuntimeExpression_NilValueMarkerPriority(t *testing.T) {
	// Setup: Mix of nil-valued (loop vars) and build-time props
	dataScope := map[string]any{
		"component": nil,         // Loop variable (nil marker)
		"title":     "Welcome",   // Build-time content prop
		"count":     42,          // Build-time variable
	}

	analyzer := NewScopeAnalyzer(dataScope)

	// Also track title as content prop explicitly
	analyzer.TrackContentProp("title")

	tests := []struct {
		expr     string
		expected bool
	}{
		{"component.name", true},  // Nil marker → runtime
		{"title", false},          // Non-nil value + tracked → build-time
		{"count", false},          // Non-nil value, not tracked → build-time (default)
		{"component", true},       // Nil marker → runtime
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result := analyzer.IsRuntimeExpression(tt.expr)
			if result != tt.expected {
				t.Errorf("IsRuntimeExpression(%q) = %v, want %v",
					tt.expr, result, tt.expected)
			}
		})
	}
}

// TestIsRuntimeExpression_NoDataScope tests that analyzer works without dataScope
//
// Pattern: Edge Case Test [Load: 4]
func TestIsRuntimeExpression_NoDataScope(t *testing.T) {
	// Nil dataScope should not cause panics
	analyzer := NewScopeAnalyzer(nil)

	// Without explicit tracking, variables default to build-time
	if analyzer.IsRuntimeExpression("component.name") {
		t.Error("Without dataScope or tracking, expressions should default to build-time")
	}

	// But Alpine stores are always runtime
	if !analyzer.IsRuntimeExpression("$store.auth") {
		t.Error("Alpine stores should always be runtime")
	}

	// Track variable explicitly - should work without dataScope
	analyzer.TrackLoopVariable("component")
	if !analyzer.IsRuntimeExpression("component.name") {
		t.Error("Explicitly tracked loop variables should be runtime")
	}
}
