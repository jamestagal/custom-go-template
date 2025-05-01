package testutils

import (
	"strings"

	"github.com/jimafisk/custom_go_template/ast"
)

// RenderNode renders an AST node to string for testing
func RenderNode(node ast.Node) string {
	var sb strings.Builder
	RenderNodeToBuilder(&sb, node)
	return sb.String()
}

// RenderNodeToBuilder renders an AST node to a string builder
func RenderNodeToBuilder(sb *strings.Builder, node ast.Node) {
	switch n := node.(type) {
	case *ast.Element:
		sb.WriteString("<")
		sb.WriteString(n.TagName)
		
		// Render attributes
		for _, attr := range n.Attributes {
			sb.WriteString(" ")
			sb.WriteString(attr.Name)
			if attr.Value != "" {
				sb.WriteString("=\"")
				// Escape double quotes within the value for HTML attribute
				escapedValue := strings.ReplaceAll(attr.Value, "\"", "&quot;")
				sb.WriteString(escapedValue)
				sb.WriteString("\"")
			}
		}
		
		if n.SelfClosing {
			sb.WriteString(" />")
			return
		}
		
		sb.WriteString(">")
		
		// Render children
		for _, child := range n.Children {
			RenderNodeToBuilder(sb, child)
		}
		
		sb.WriteString("</")
		sb.WriteString(n.TagName)
		sb.WriteString(">")
		
	case *ast.TextNode:
		// Render the content directly without trimming
		// Let NormalizeWhitespace handle differences during comparison
		sb.WriteString(n.Content)
		
	case *ast.ExpressionNode:
		sb.WriteString("<span x-text=\"")
		sb.WriteString(n.Expression)
		sb.WriteString("\"></span>")
		
	case *ast.CommentNode:
		sb.WriteString("<!--")
		sb.WriteString(n.Content)
		sb.WriteString("-->")
		
	case *ast.Loop:
		sb.WriteString("<template x-for=\"")
		sb.WriteString(n.Iterator)
		if n.Value != "" {
			sb.WriteString(", ")
			sb.WriteString(n.Value)
		}
		if n.IsOf {
			sb.WriteString(" of Object.entries(")
			sb.WriteString(n.Collection)
			sb.WriteString(")")
		} else {
			sb.WriteString(" in ")
			sb.WriteString(n.Collection)
		}
		sb.WriteString("\">")
		
		// Render loop content
		for _, child := range n.Content {
			RenderNodeToBuilder(sb, child)
		}
		
		sb.WriteString("</template>")
		
	case *ast.Conditional:
		// Check if we need to handle the special case for Alpine.js directives
		if hasAlpineDirectives(n) {
			renderAlpineConditional(sb, n)
		} else {
			// Legacy rendering for tests that expect x-if with negated conditions
			renderLegacyConditional(sb, n)
		}
		
	default:
		// Skip other node types like FenceSection
	}
}

// hasAlpineDirectives checks if the conditional uses native Alpine.js directives
func hasAlpineDirectives(cond *ast.Conditional) bool {
	// This is a heuristic - we check if any of the nodes in the conditional
	// are elements with x-else or x-else-if attributes
	for _, node := range cond.IfContent {
		if elem, ok := node.(*ast.Element); ok {
			for _, attr := range elem.Attributes {
				if attr.Name == "x-else" || attr.Name == "x-else-if" {
					return true
				}
			}
		}
	}
	
	return false
}

// renderAlpineConditional renders a conditional using native Alpine.js directives
func renderAlpineConditional(sb *strings.Builder, cond *ast.Conditional) {
	// Render if branch
	sb.WriteString("<template x-if=\"")
	sb.WriteString(cond.IfCondition)
	sb.WriteString("\">")
	
	for _, child := range cond.IfContent {
		RenderNodeToBuilder(sb, child)
	}
	
	sb.WriteString("</template>")
	
	// Render else-if branches
	for i, condition := range cond.ElseIfConditions {
		sb.WriteString("<template x-else-if=\"")
		sb.WriteString(condition)
		sb.WriteString("\">")
		
		if i < len(cond.ElseIfContent) {
			for _, child := range cond.ElseIfContent[i] {
				RenderNodeToBuilder(sb, child)
			}
		}
		
		sb.WriteString("</template>")
	}
	
	// Render else branch
	if len(cond.ElseContent) > 0 {
		sb.WriteString("<template x-else>")
		
		for _, child := range cond.ElseContent {
			RenderNodeToBuilder(sb, child)
		}
		
		sb.WriteString("</template>")
	}
}

// renderLegacyConditional renders a conditional using x-if with negated conditions
// This is kept for backward compatibility with existing tests
func renderLegacyConditional(sb *strings.Builder, cond *ast.Conditional) {
	// Render if branch
	sb.WriteString("<template x-if=\"")
	sb.WriteString(cond.IfCondition)
	sb.WriteString("\">")
	
	for _, child := range cond.IfContent {
		RenderNodeToBuilder(sb, child)
	}
	
	sb.WriteString("</template>")
	
	// Render else-if branches
	for i, condition := range cond.ElseIfConditions {
		sb.WriteString("<template x-if=\"!(")
		sb.WriteString(cond.IfCondition)
		sb.WriteString(") && ")
		sb.WriteString(condition)
		sb.WriteString("\">")
		
		if i < len(cond.ElseIfContent) {
			for _, child := range cond.ElseIfContent[i] {
				RenderNodeToBuilder(sb, child)
			}
		}
		
		sb.WriteString("</template>")
	}
	
	// Render else branch
	if len(cond.ElseContent) > 0 {
		sb.WriteString("<template x-if=\"!(")
		sb.WriteString(cond.IfCondition)
		sb.WriteString(")\">")
		
		for _, child := range cond.ElseContent {
			RenderNodeToBuilder(sb, child)
		}
		
		sb.WriteString("</template>")
	}
}

// NormalizeWhitespace removes extra whitespace from HTML for comparison
func NormalizeWhitespace(s string) string {
	// Replace multiple whitespace with a single space
	re := strings.NewReplacer(
		"\n", " ",
		"\t", " ",
	)
	s = re.Replace(s)
	
	// Replace multiple spaces with a single space
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	
	return strings.TrimSpace(s)
}
