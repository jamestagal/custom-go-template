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
