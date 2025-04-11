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
				sb.WriteString(attr.Value)
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
		sb.WriteString("<template x-if=\"")
		sb.WriteString(n.IfCondition)
		sb.WriteString("\">")
		
		// Render if content
		for _, child := range n.IfContent {
			RenderNodeToBuilder(sb, child)
		}
		
		sb.WriteString("</template>")
		
		// Render else-if branches
		for i, condition := range n.ElseIfConditions {
			sb.WriteString("<template x-if=\"!(")
			sb.WriteString(n.IfCondition)
			sb.WriteString(") && ")
			sb.WriteString(condition)
			sb.WriteString("\">")
			
			if i < len(n.ElseIfContent) {
				for _, child := range n.ElseIfContent[i] {
					RenderNodeToBuilder(sb, child)
				}
			}
			
			sb.WriteString("</template>")
		}
		
		// Render else branch
		if len(n.ElseContent) > 0 {
			sb.WriteString("<template x-if=\"!(")
			sb.WriteString(n.IfCondition)
			sb.WriteString(")\">")
			
			for _, child := range n.ElseContent {
				RenderNodeToBuilder(sb, child)
			}
			
			sb.WriteString("</template>")
		}
		
	default:
		// Skip other node types like FenceSection
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
