package analyzer

import (
	"strings"
)

// ScopeAnalyzer distinguishes build-time resolvable expressions from runtime-only expressions
//
// Pattern: Service Implementation Pattern [Load: 10]
// Cognitive Load: 10 (map operations: 3, expression parsing: 4, decision logic: 3)
//
// This analyzer is the foundation of the runtime component resolution system.
// It tracks which variables are available at build time vs runtime only.
//
// Decision Logic:
//   - Loop variables (component, item) → RUNTIME-ONLY
//   - Content props (title, description from export let) → BUILD-TIME
//   - Exported props (export let variables) → BUILD-TIME
//   - Alpine stores ($store.*, $auth.*) → RUNTIME-ONLY
//   - String literals ("ComponentName") → BUILD-TIME
//   - Expressions with operators (+, -, *, etc.) → RUNTIME-ONLY (for safety)
//
// Example Usage:
//   analyzer := NewScopeAnalyzer(dataScope)
//   analyzer.TrackLoopVariable("component")
//   analyzer.TrackContentProp("title")
//
//   analyzer.IsRuntimeExpression("component.name") // → true (loop variable)
//   analyzer.IsRuntimeExpression("title")          // → false (content prop)
//   analyzer.IsRuntimeExpression("$auth.user")     // → true (Alpine store)
//   analyzer.IsRuntimeExpression(`"Hero2436"`)     // → false (string literal)
type ScopeAnalyzer struct {
	buildVars   map[string]bool // Variables resolvable at build time
	runtimeVars map[string]bool // Variables only available at runtime
	dataScope   map[string]any  // Reference to data scope for context
}

// NewScopeAnalyzer creates a new scope analyzer
//
// Pattern: Constructor Pattern [Load: 3]
// Cognitive Load: 3 (map initialization: 3)
//
// The dataScope parameter is optional and used for enhanced analysis.
func NewScopeAnalyzer(dataScope map[string]any) *ScopeAnalyzer {
	return &ScopeAnalyzer{
		buildVars:   make(map[string]bool),
		runtimeVars: make(map[string]bool),
		dataScope:   dataScope,
	}
}

// TrackLoopVariable marks a variable as runtime-only (e.g., loop iterator)
//
// Pattern: Helper Function [Load: 2]
// Cognitive Load: 2 (map write: 2)
//
// Example:
//   analyzer.TrackLoopVariable("component")
//   analyzer.IsRuntimeExpression("component.name") // → true
func (s *ScopeAnalyzer) TrackLoopVariable(name string) {
	s.runtimeVars[name] = true
}

// TrackContentProp marks a variable as build-time resolvable (from content JSON)
//
// Pattern: Helper Function [Load: 2]
// Cognitive Load: 2 (map write: 2)
//
// Example:
//   analyzer.TrackContentProp("title")
//   analyzer.IsRuntimeExpression("title") // → false
func (s *ScopeAnalyzer) TrackContentProp(name string) {
	s.buildVars[name] = true
}

// TrackExportedProp marks a variable as build-time resolvable (export let)
//
// Pattern: Helper Function [Load: 2]
// Cognitive Load: 2 (map write: 2)
//
// Example:
//   analyzer.TrackExportedProp("title")
//   analyzer.IsRuntimeExpression("title") // → false
func (s *ScopeAnalyzer) TrackExportedProp(name string) {
	s.buildVars[name] = true
}

// IsRuntimeExpression determines if an expression requires runtime resolution
//
// Pattern: Decision Function [Load: 15]
// Cognitive Load: 15 (string parsing: 5, store detection: 3, variable lookup: 4, operator check: 3)
//
// Returns TRUE if expression contains:
//   - Loop variables (component, item, index)
//   - Alpine store references ($store.*, $auth.*)
//   - Operators (+, -, *, /, etc.) - for safety
//
// Returns FALSE if expression is:
//   - String literal ("ComponentName", 'Hero2436')
//   - Content prop only (title, description)
//   - Exported prop only (from export let)
//   - Simple identifier in buildVars
//
// Example:
//   analyzer.IsRuntimeExpression("component.name")    // → true (loop var)
//   analyzer.IsRuntimeExpression(`"Hero2436"`)        // → false (string literal)
//   analyzer.IsRuntimeExpression("$store.auth.user")  // → true (Alpine store)
//   analyzer.IsRuntimeExpression("title")             // → false (content prop)
//   analyzer.IsRuntimeExpression("count + 1")         // → true (has operator)
func (s *ScopeAnalyzer) IsRuntimeExpression(expr string) bool {
	expr = strings.TrimSpace(expr)

	// 1. String literals are build-time (COGNITIVE LOAD: 2)
	if isStringLiteral(expr) {
		return false
	}

	// 2. Alpine store references are runtime-only (COGNITIVE LOAD: 3)
	if isAlpineStoreReference(expr) {
		return true
	}

	// 3. Expressions with operators are runtime (for safety) (COGNITIVE LOAD: 3)
	if hasOperators(expr) {
		return true
	}

	// 4. Extract root variable from expression (COGNITIVE LOAD: 4)
	// Example: "component.name" → "component"
	variables := extractVariablesFromExpression(expr)

	// 5. Check if any variable is runtime-only (COGNITIVE LOAD: 3)
	for _, variable := range variables {
		if s.runtimeVars[variable] {
			return true
		}
	}

	// 6. If all variables are in buildVars or unknown, it's build-time
	// Unknown variables default to build-time for backwards compatibility
	return false
}

// extractVariablesFromExpression extracts all variable names from an expression
//
// Pattern: Helper Function [Load: 8]
// Cognitive Load: 8 (string parsing: 5, array operations: 3)
//
// Handles:
//   - Simple identifiers: "component" → ["component"]
//   - Property access: "component.name" → ["component"]
//   - Nested properties: "data.component.name" → ["data"]
//
// Example:
//   extractVariablesFromExpression("component.name") // → ["component"]
//   extractVariablesFromExpression("item.value")     // → ["item"]
func extractVariablesFromExpression(expr string) []string {
	// For property access, get the root variable
	// Example: "component.name" → "component"
	parts := strings.Split(expr, ".")
	if len(parts) == 0 {
		return []string{}
	}

	rootVar := parts[0]

	// Clean up any brackets or parentheses
	rootVar = strings.Split(rootVar, "[")[0]
	rootVar = strings.Split(rootVar, "(")[0]
	rootVar = strings.TrimSpace(rootVar)

	if rootVar == "" {
		return []string{}
	}

	return []string{rootVar}
}

// isStringLiteral checks if expression is a quoted string literal
//
// Pattern: Helper Function [Load: 3]
// Cognitive Load: 3 (string prefix/suffix checks: 3)
//
// Example:
//   isStringLiteral(`"Hero2436"`)  // → true
//   isStringLiteral(`'Hero2436'`)  // → true
//   isStringLiteral(`Hero2436`)    // → false
func isStringLiteral(expr string) bool {
	expr = strings.TrimSpace(expr)

	// Double quoted
	if strings.HasPrefix(expr, `"`) && strings.HasSuffix(expr, `"`) {
		return true
	}

	// Single quoted
	if strings.HasPrefix(expr, `'`) && strings.HasSuffix(expr, `'`) {
		return true
	}

	return false
}

// isAlpineStoreReference checks if expression references an Alpine.js store
//
// Pattern: Helper Function [Load: 3]
// Cognitive Load: 3 (string prefix check: 3)
//
// Detects:
//   - $store.auth.user
//   - $auth.user (shorthand)
//   - $cart.items
//
// Example:
//   isAlpineStoreReference("$store.auth")  // → true
//   isAlpineStoreReference("$auth.user")   // → true
//   isAlpineStoreReference("auth.user")    // → false
func isAlpineStoreReference(expr string) bool {
	expr = strings.TrimSpace(expr)
	return strings.HasPrefix(expr, "$")
}

// hasOperators checks if expression contains operators
//
// Pattern: Helper Function [Load: 4]
// Cognitive Load: 4 (substring checks: 4)
//
// Detects common operators:
//   - Arithmetic: +, -, *, /
//   - Comparison: ==, !=, <, >, <=, >=
//   - Logical: &&, ||
//
// Example:
//   hasOperators("count + 1")     // → true
//   hasOperators("value > 10")    // → true
//   hasOperators("component.name") // → false
func hasOperators(expr string) bool {
	operators := []string{
		"+", "-", "*", "/",
		"==", "!=", "<", ">", "<=", ">=",
		"&&", "||",
		"?", ":", // Ternary
	}

	for _, op := range operators {
		if strings.Contains(expr, op) {
			return true
		}
	}

	return false
}
