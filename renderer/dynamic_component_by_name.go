package renderer

import (
	"fmt"
	"strings"

	"github.com/jimafisk/custom_go_template/ast"
)

// RenderDynamicComponentByName renders a fallback HTML comment for DynamicComponentByNameNode.
// This function should RARELY be called since the transformer converts these nodes to actual components.
// It exists for debugging scenarios when transformation fails or is incomplete.
//
// Output format:
//
//	<!-- Component:dynamic (not resolved): name={expr}, spreads=[...], props=[...] -->
//
// Cognitive Load: 8
// - Build name part: 2
// - Build spreads part: 2
// - Build props part: 2
// - Combine into comment: 2
func RenderDynamicComponentByName(node *ast.DynamicComponentByNameNode) string {
	// Build diagnostic parts (COGNITIVE LOAD: preallocate for known parts)
	parts := make([]string, 0, 3) // name, spreads, props

	// Part 1: Name expression (COGNITIVE LOAD: wrapped error context)
	if node.NameExpression != "" {
		// Check if name looks like an expression (contains . [ ( etc)
		if strings.ContainsAny(node.NameExpression, ".[]()") {
			parts = append(parts, fmt.Sprintf("name={%s}", node.NameExpression))
		} else {
			parts = append(parts, fmt.Sprintf("name=%s", node.NameExpression))
		}
	}

	// Part 2: Spread props
	if len(node.SpreadProps) > 0 {
		spreads := formatSpreadProps(node.SpreadProps)
		parts = append(parts, fmt.Sprintf("spreads=[%s]", spreads))
	}

	// Part 3: Regular props
	if len(node.Props) > 0 {
		props := formatComponentProps(node.Props)
		parts = append(parts, fmt.Sprintf("props=[%s]", props))
	}

	// Combine parts into diagnostic comment
	diagnostic := "Component:dynamic (not resolved)"
	if len(parts) > 0 {
		diagnostic = fmt.Sprintf("%s: %s", diagnostic, strings.Join(parts, ", "))
	}

	return fmt.Sprintf("<!-- %s -->", diagnostic)
}

// formatSpreadProps formats spread prop expressions for diagnostic output
// Example: ["component.fields", "baseProps"] → "{...component.fields}, {...baseProps}"
//
// Cognitive Load: 4
// - Preallocate result slice: 1
// - Loop and format: 2
// - Join results: 1
func formatSpreadProps(spreads []string) string {
	// Preallocate for spreads (COGNITIVE LOAD RULE)
	formatted := make([]string, 0, len(spreads))

	for _, spread := range spreads {
		formatted = append(formatted, fmt.Sprintf("{...%s}", spread))
	}

	return strings.Join(formatted, ", ")
}

// formatComponentProps formats component props for diagnostic output
// Example: [theme="dark", active={isActive}] → `theme="dark", active={isActive}`
//
// Cognitive Load: 5
// - Preallocate result slice: 1
// - Loop through props: 2
// - Format dynamic/static: 1
// - Join results: 1
func formatComponentProps(props []ast.ComponentProp) string {
	// Preallocate for props (COGNITIVE LOAD RULE)
	formatted := make([]string, 0, len(props))

	for _, prop := range props {
		if prop.IsDynamic {
			// Dynamic prop: name={value}
			formatted = append(formatted, fmt.Sprintf("%s={%s}", prop.Name, prop.Value))
		} else {
			// Static prop: name="value"
			formatted = append(formatted, fmt.Sprintf(`%s="%s"`, prop.Name, prop.Value))
		}
	}

	return strings.Join(formatted, ", ")
}
