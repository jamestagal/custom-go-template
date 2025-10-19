package builder

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jimafisk/custom_go_template/ast"
)

// ComponentTemplate represents a component with its name and parsed AST
// This structure is used by the registry generator to convert components to JavaScript
type ComponentTemplate struct {
	Name string        // Component name (e.g., "Hero2436", "Services2437")
	AST  *ast.Template // Parsed and transformed component AST
}

// RenderContext tracks rendering context to handle expressions differently in different contexts
// Pattern: Context Object Pattern [Load: 5]
type RenderContext struct {
	insideLiteral    bool // Inside <style> or <script> tags where expressions should NOT be converted
	insideAlpineAttr bool // Inside Alpine.js directive where expressions should be preserved
}

// GenerateComponentRegistry converts component ASTs to ES module with template literal factories
// Cognitive Load: 8 (Main orchestration with clear structure)
// Pattern: Service Implementation Pattern
func GenerateComponentRegistry(components []ComponentTemplate) string {
	if len(components) == 0 {
		return "export default {};"
	}

	var sb strings.Builder
	sb.WriteString("export default {\n")

	// Preallocate for performance (COGNITIVE LOAD RULE)
	for i, component := range components {
		// Generate component entry: 'ComponentName': (props) => `...`
		sb.WriteString(fmt.Sprintf("  '%s': (props) => `", component.Name))

		// Convert AST to JavaScript template literal
		templateContent := convertToJSTemplate(component.AST)
		sb.WriteString(templateContent)

		sb.WriteString("`")

		// Add comma separator except for last component
		if i < len(components)-1 {
			sb.WriteString(",\n")
		} else {
			sb.WriteString("\n")
		}
	}

	sb.WriteString("};")
	return sb.String()
}

// convertToJSTemplate recursively converts AST nodes to JavaScript template literal content
// Cognitive Load: 12 (Recursive traversal with clear cases)
func convertToJSTemplate(template *ast.Template) string {
	if template == nil {
		return ""
	}

	var sb strings.Builder
	ctx := &RenderContext{
		insideLiteral:    false,
		insideAlpineAttr: false,
	}

	for _, node := range template.RootNodes {
		renderNodeToJS(node, &sb, ctx)
	}

	return sb.String()
}

// isLiteralContentElement checks if a tag name represents literal content
// where expressions should NOT be converted to template literals
// Cognitive Load: 3 (Simple tag check)
// Pattern: Basic helper function
func isLiteralContentElement(tagName string) bool {
	return tagName == "style" || tagName == "script"
}

// renderNodeToJS renders a single AST node to JavaScript template literal syntax
// Cognitive Load: 15 (Type switch with context tracking)
func renderNodeToJS(node ast.Node, sb *strings.Builder, ctx *RenderContext) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *ast.Element:
		renderElementToJS(n, sb, ctx)

	case *ast.TextNode:
		// Escape special template literal characters (COGNITIVE LOAD RULE)
		escaped := escapeTemplateLiteral(n.Content)
		sb.WriteString(escaped)

	case *ast.ExpressionNode:
		// CRITICAL FIX: Context-aware expression conversion
		if ctx.insideLiteral {
			// Inside style/script tags - preserve as-is
			escaped := escapeTemplateLiteral(n.Expression)
			sb.WriteString(escaped)
		} else if ctx.insideAlpineAttr {
			// Inside Alpine directive attribute - preserve {expression} syntax for Alpine
			sb.WriteString("{")
			sb.WriteString(n.Expression)
			sb.WriteString("}")
		} else {
			// Use identifier-level prefixing to handle complex expressions
			// ARROW FUNCTION FIX: Extract arrow function parameters before prefixing
			arrowParams := extractArrowFunctionParams(n.Expression)
			converted := prefixIdentifiersInExpression(n.Expression, arrowParams)
			// Normal content - transform {variable} to ${props.variable}
			sb.WriteString("${")
			sb.WriteString(converted)
			sb.WriteString("}")
		}

	case *ast.Conditional:
		// Conditionals are rendered as <template> elements with x-if
		// Alpine.js handles them at runtime
		renderConditionalToJS(n, sb, ctx)

	case *ast.Loop:
		// Loops are rendered as <template> elements with x-for
		renderLoopToJS(n, sb, ctx)

	case *ast.CommentNode:
		// Preserve HTML comments
		sb.WriteString("<!--")
		sb.WriteString(escapeTemplateLiteral(n.Content))
		sb.WriteString("-->")

	// Ignore fence/script/style sections - they're not part of component markup
	case *ast.FenceSection, *ast.ScriptSection, *ast.StyleSection:
		// Skip metadata sections

	default:
		// Unknown node types are silently skipped
		// This allows for graceful handling of future node types
	}
}

// renderElementToJS renders an HTML element with attributes and children
// Cognitive Load: 16 (Element rendering with context propagation and Alpine detection)
func renderElementToJS(elem *ast.Element, sb *strings.Builder, ctx *RenderContext) {
	// Opening tag
	sb.WriteString("<")
	sb.WriteString(elem.TagName)

	// Render attributes with Alpine directive detection
	if len(elem.Attributes) > 0 {
		for _, attr := range elem.Attributes {
			sb.WriteString(" ")
			renderAttributeToJS(attr, sb, ctx, elem.Children)
		}
	}

	// Self-closing elements
	if elem.SelfClosing {
		sb.WriteString(">")
		return
	}

	sb.WriteString(">")

	// Determine if children are in literal context
	childCtx := &RenderContext{
		insideLiteral:    ctx.insideLiteral || isLiteralContentElement(elem.TagName),
		insideAlpineAttr: false, // Reset for element children
	}

	// Render children with updated context
	for _, child := range elem.Children {
		renderNodeToJS(child, sb, childCtx)
	}

	// Closing tag
	sb.WriteString("</")
	sb.WriteString(elem.TagName)
	sb.WriteString(">")
}

// Skip list for identifiers that should NOT get props. prefix
// Cognitive Load: 5 (Static configuration)
var skipIdentifiers = map[string]bool{
	// Loop variables (common in x-for)
	"index":     true,
	"item":      true,
	"todo":      true,
	"component": true,

	// Alpine.js built-ins (magic properties)
	"$store":    true,
	"$el":       true,
	"$refs":     true,
	"$watch":    true,
	"$dispatch": true,
	"$nextTick": true,
	"$data":     true,
	"$root":     true,

	// JavaScript built-ins
	"window":   true,
	"document": true,
	"console":  true,
	"Math":     true,
	"Date":     true,
	"JSON":     true,
	"Array":    true,
	"Object":   true,
	"String":   true,
	"Number":   true,
	"Boolean":  true,
}

// expressionPattern matches {expression} patterns in attribute values
// Cognitive Load: 3 (Simple regex pattern)
// Updated to match any content within braces (not just identifiers)
var expressionPattern = regexp.MustCompile(`\{([^{}]+)\}`)

// identifierPattern matches JavaScript identifiers at word boundaries
// Cognitive Load: 3 (Identifier matching)
// Important: This uses negative lookbehind/lookahead to avoid matching property access
var identifierPattern = regexp.MustCompile(`(?:^|[^a-zA-Z0-9_$.\]])([a-zA-Z_$][\w]*)(?:$|[^a-zA-Z0-9_$.])`)

// arrowFunctionPattern matches arrow function parameter patterns
// Pattern: (param1, param2) => or param =>
// Cognitive Load: 4 (Arrow function parameter detection)
var arrowFunctionPattern = regexp.MustCompile(`\(([^)]+)\)\s*=>|([a-zA-Z_$][\w]*)\s*=>`)

// convertAttributeExpressions converts {expression} patterns in attribute values to ${props.expression}
// Cognitive Load: 14 (Regex replacement with skip list, property chain handling, and object literal handling)
// Pattern: String transformation with identifier-level prefixing
//
// CRITICAL FIX: This function now prefixes individual identifiers instead of whole expressions
// For example: {(start * 1) + index + 1} → ${(props.start * 1) + index + 1}
// Note: 'index' is a loop variable, so it stays as 'index' (no props. prefix)
//
// The function handles:
// - Simple identifiers: {count} → ${props.count}
// - Complex expressions: {(start * 1) + index} → ${(props.start * 1) + index}
// - Property access: {item.name} → ${item.name} (entire chain preserved)
// - Method calls: {$store.theme.getCurrentColors().background} → ${$store.theme.getCurrentColors().background}
// - Alpine object literals: { count: {count} } → { count: ${props.count} }
// - Arrow functions: {products.reduce((sum, p) => sum + p)} → ${props.products.reduce((sum, p) => sum + p)}
// - Skip list: Loop variables, Alpine built-ins, JS built-ins, arrow function params are NOT prefixed
func convertAttributeExpressions(attrValue string) string {
	// Find all {expression} patterns
	return expressionPattern.ReplaceAllStringFunc(attrValue, func(match string) string {
		// Extract the expression without braces
		expr := match[1 : len(match)-1] // Remove { and }

		// Check if this looks like an Alpine object literal content (after braces are stripped)
		// An object literal contains : for key-value pairs
		if strings.Contains(expr, ":") {
			// Check if it actually contains any template expressions (nested {})
			if !strings.Contains(expr, "{") {
				// Pure object literal like " count: 0, message: 'hello' " - don't convert
				return match
			}
			// Process object literal with nested expressions
			return "{" + convertObjectLiteralExpressions(expr) + "}"
		}

		// Detect arrow function parameters and add to skip list
		arrowParams := extractArrowFunctionParams(expr)

		// For other expressions, prefix each identifier with props.
		// We need to be careful about property access chains and method calls
		converted := prefixIdentifiersInExpression(expr, arrowParams)

		return "${" + converted + "}"
	})
}

// extractArrowFunctionParams extracts parameter names from arrow functions in an expression
// Returns a map of parameter names to skip during identifier prefixing
// Cognitive Load: 10 (Regex matching with parameter extraction and cleanup)
// Pattern: Arrow function parameter detection
func extractArrowFunctionParams(expr string) map[string]bool {
	params := make(map[string]bool)

	// Find all arrow function patterns
	matches := arrowFunctionPattern.FindAllStringSubmatch(expr, -1)
	for _, match := range matches {
		if len(match) > 1 {
			// match[1] is for (param1, param2) => pattern
			// match[2] is for param => pattern
			if match[1] != "" {
				// Multiple parameters: (param1, param2) =>
				// The captured group may contain extra parens from nested calls like .reduce((sum, p)
				// We need to clean these up
				paramStr := match[1]

				// Remove leading parentheses that might be from method calls
				paramStr = strings.TrimPrefix(paramStr, "(")

				// Split by comma
				paramList := strings.Split(paramStr, ",")
				for _, param := range paramList {
					param = strings.TrimSpace(param)

					// Remove any remaining parens or whitespace
					param = strings.Trim(param, "() \t")

					// Extract just the identifier (in case of destructuring like {x, y})
					// For now, we only support simple identifiers
					if param != "" && isSimpleIdentifier(param) {
						params[param] = true
					}
				}
			} else if match[2] != "" {
				// Single parameter: param =>
				param := strings.TrimSpace(match[2])
				if param != "" {
					params[param] = true
				}
			}
		}
	}

	return params
}

// isSimpleIdentifier checks if a string is a valid JavaScript identifier
// Cognitive Load: 3 (Simple validation)
func isSimpleIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}
	// First character must be letter, underscore, or dollar sign
	first := s[0]
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || first == '_' || first == '$') {
		return false
	}
	// Remaining characters can also be digits
	for i := 1; i < len(s); i++ {
		ch := s[i]
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' || ch == '$') {
			return false
		}
	}
	return true
}

// prefixIdentifiersInExpression prefixes standalone identifiers with props. while preserving property chains and method calls
// Cognitive Load: 18 (Complex string traversal with parentheses tracking for method calls vs grouping)
// ARROW FUNCTION FIX: Now accepts arrowParams map to skip arrow function parameters
func prefixIdentifiersInExpression(expr string, arrowParams map[string]bool) string {
	// Merge arrow params with skip list
	combinedSkip := make(map[string]bool)
	for k, v := range skipIdentifiers {
		combinedSkip[k] = v
	}
	if arrowParams != nil {
		for k, v := range arrowParams {
			combinedSkip[k] = v
		}
	}

	// Use a context-aware tokenizer that distinguishes method calls from grouping parens
	var result strings.Builder
	var currentToken strings.Builder
	parenDepth := 0 // Track parentheses depth
	isMethodCall := false // Track if current parens are for a method call

	for i := 0; i < len(expr); i++ {
		ch := expr[i]

		if ch == '(' {
			// Check if this is a method call (preceded by identifier/property chain)
			// vs grouping parentheses (preceded by operator/whitespace)
			if currentToken.Len() > 0 {
				lastChar := currentToken.String()[currentToken.Len()-1]
				// If preceded by identifier character, it's a method call
				isMethodCall = (lastChar >= 'a' && lastChar <= 'z') ||
					(lastChar >= 'A' && lastChar <= 'Z') ||
					(lastChar >= '0' && lastChar <= '9') ||
					lastChar == '_' || lastChar == '$' || lastChar == ')'
			} else {
				// Preceded by nothing or operator - grouping parens
				isMethodCall = false
			}

			if isMethodCall {
				// Method call - include in current token
				currentToken.WriteByte(ch)
				parenDepth++
			} else {
				// Grouping parens - process accumulated token and keep paren as operator
				if currentToken.Len() > 0 {
					token := currentToken.String()
					result.WriteString(processToken(token, combinedSkip))
					currentToken.Reset()
				}
				result.WriteByte(ch)
				parenDepth++
			}
		} else if ch == ')' {
			if parenDepth > 0 && isMethodCall {
				// End of method call - include in current token
				currentToken.WriteByte(ch)
				parenDepth--
				if parenDepth == 0 {
					isMethodCall = false
				}
			} else {
				// End of grouping parens - process token and keep paren as operator
				if currentToken.Len() > 0 {
					token := currentToken.String()
					result.WriteString(processToken(token, combinedSkip))
					currentToken.Reset()
				}
				result.WriteByte(ch)
				parenDepth--
			}
		} else if parenDepth > 0 && isMethodCall {
			// Inside method call parentheses - include in current token
			currentToken.WriteByte(ch)
		} else if ch == '.' {
			// Property access - keep as part of token
			currentToken.WriteByte(ch)
		} else if isOperatorOrDelimiter(ch) {
			// Process accumulated token
			if currentToken.Len() > 0 {
				token := currentToken.String()
				result.WriteString(processToken(token, combinedSkip))
				currentToken.Reset()
			}
			// Add the operator/delimiter
			result.WriteByte(ch)
		} else {
			// Accumulate identifier/property chain/method call
			currentToken.WriteByte(ch)
		}
	}

	// Process final token
	if currentToken.Len() > 0 {
		token := currentToken.String()
		result.WriteString(processToken(token, combinedSkip))
	}

	return result.String()
}

// isOperatorOrDelimiter checks if a character is an operator or delimiter
// Cognitive Load: 3 (Simple character check)
// Note: Parentheses and dot are NOT treated as delimiters here - they're handled specially
func isOperatorOrDelimiter(ch byte) bool {
	return ch == '+' || ch == '-' || ch == '*' || ch == '/' || ch == '%' ||
		ch == '[' || ch == ']' ||
		ch == ',' || ch == '?' || ch == ':' || ch == '!' ||
		ch == '>' || ch == '<' || ch == '=' || ch == '&' || ch == '|' ||
		ch == ' ' || ch == '\t' || ch == '\n'
}

// processToken processes a single token (identifier, property chain, or method call)
// Cognitive Load: 10 (Token classification with method call handling)
// ARROW FUNCTION FIX: Now accepts skipList parameter
func processToken(token string, skipList map[string]bool) string {
	token = strings.TrimSpace(token)
	if token == "" {
		return token
	}

	// Check if it's a property access chain or method call (contains . or ())
	if strings.Contains(token, ".") || strings.Contains(token, "(") {
		// Split by dot to find the root identifier
		parts := strings.Split(token, ".")
		firstPart := parts[0]

		// Only prefix if the first part is not in skip list and not already "props"
		if !skipList[firstPart] && firstPart != "props" {
			return "props." + token
		}
		return token
	}

	// Check if it's in skip list
	if skipList[token] {
		return token
	}

	// Check if already prefixed
	if token == "props" {
		return token
	}

	// Check if it's a number or string literal
	if isNumeric(token) || isStringLiteral(token) {
		return token
	}

	// Prefix with props.
	return "props." + token
}

// isNumeric checks if a string is a numeric literal
// Cognitive Load: 3 (Simple check)
func isNumeric(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, ch := range s {
		if !((ch >= '0' && ch <= '9') || ch == '.') {
			return false
		}
	}
	return true
}

// isStringLiteral checks if a string is a string literal
// Cognitive Load: 3 (Simple check)
func isStringLiteral(s string) bool {
	return (strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'")) ||
		(strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\""))
}

// isAlpineObjectLiteral checks if expression looks like an Alpine object literal
// Examples: "{ count: 0, items: [] }" or "{ count: {count}, message: '{message}' }"
// Cognitive Load: 5 (Simple pattern detection)
func isAlpineObjectLiteral(expr string) bool {
	trimmed := strings.TrimSpace(expr)
	// Must start with { and end with }
	if !strings.HasPrefix(trimmed, "{") || !strings.HasSuffix(trimmed, "}") {
		return false
	}
	// Must contain : (key: value syntax)
	return strings.Contains(trimmed, ":")
}

// convertObjectLiteralExpressions processes object literals to convert nested expressions
// Example: " count: {count}, message: '{message}' " → " count: ${props.count}, message: '${props.message}' "
// Cognitive Load: 8 (Nested expression conversion)
func convertObjectLiteralExpressions(objLiteral string) string {
	// Match {identifier} patterns but NOT at the start/end (those are object literal braces)
	nestedExprPattern := regexp.MustCompile(`\{([a-zA-Z_$][\w]*)\}`)

	converted := nestedExprPattern.ReplaceAllStringFunc(objLiteral, func(match string) string {
		// Extract identifier
		id := match[1 : len(match)-1] // Remove { and }

		// Skip if in skip list
		if skipIdentifiers[id] {
			return match
		}

		// Convert to ${props.identifier}
		return "${props." + id + "}"
	})

	// Return without the wrapping ${}
	return converted
}

// renderAttributeToJS renders an HTML attribute with Alpine-aware handling
// Cognitive Load: 12 (Attribute rendering with expression conversion)
//
// CRITICAL FIX: This function now properly converts {expression} patterns in attribute values
// to ${props.expression} for JavaScript template literals using identifier-level prefixing
func renderAttributeToJS(attr ast.Attribute, sb *strings.Builder, ctx *RenderContext, children []ast.Node) {
	sb.WriteString(attr.Name)
	sb.WriteString("=\"")

	// Escape backslashes and backticks for template literal safety
	// Then escape quotes for attribute syntax
	// Then convert {expression} patterns to ${props.expression}
	escaped := attr.Value

	// Step 1: Escape backslashes and backticks for template literal
	escaped = strings.ReplaceAll(escaped, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, "`", "\\`")

	// Step 2: Convert {expression} patterns to ${props.expression}
	// This handles cases like: x-data="{ count: {count} }" → x-data="{ count: ${props.count} }"
	escaped = convertAttributeExpressions(escaped)

	// Step 3: Escape double quotes for attribute value syntax
	// Important: Do this AFTER expression conversion to avoid escaping quotes in ${props.x}
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)

	sb.WriteString(escaped)
	sb.WriteString("\"")
}

// renderConditionalToJS renders conditional blocks as Alpine.js <template> elements
// Cognitive Load: 13 (Conditional structure with context)
func renderConditionalToJS(cond *ast.Conditional, sb *strings.Builder, ctx *RenderContext) {
	// Main if block
	sb.WriteString(`<template x-if="`)
	sb.WriteString(cond.IfCondition)
	sb.WriteString(`">`)

	for _, node := range cond.IfContent {
		renderNodeToJS(node, sb, ctx)
	}

	sb.WriteString("</template>")

	// Else-if blocks
	for i, elseIfCond := range cond.ElseIfConditions {
		sb.WriteString(`<template x-else-if="`)
		sb.WriteString(elseIfCond)
		sb.WriteString(`">`)

		for _, node := range cond.ElseIfContent[i] {
			renderNodeToJS(node, sb, ctx)
		}

		sb.WriteString("</template>")
	}

	// Else block
	if len(cond.ElseContent) > 0 {
		sb.WriteString(`<template x-else>`)

		for _, node := range cond.ElseContent {
			renderNodeToJS(node, sb, ctx)
		}

		sb.WriteString("</template>")
	}
}

// renderLoopToJS renders loop blocks as Alpine.js <template> elements
// Cognitive Load: 9 (Loop structure with context)
func renderLoopToJS(loop *ast.Loop, sb *strings.Builder, ctx *RenderContext) {
	sb.WriteString(`<template x-for="`)

	// Build x-for expression
	if loop.Value != "" {
		// For (key, value) in collection syntax
		sb.WriteString(loop.Value)
		sb.WriteString(", ")
	}

	sb.WriteString(loop.Iterator)
	sb.WriteString(" in ")
	sb.WriteString(loop.Collection)
	sb.WriteString(`">`)

	// Render loop content
	for _, node := range loop.Content {
		renderNodeToJS(node, sb, ctx)
	}

	sb.WriteString("</template>")
}

// escapeTemplateLiteral escapes special characters for JavaScript template literals
// Cognitive Load: 4 (Simple string replacement)
// Pattern: Basic string escaping
func escapeTemplateLiteral(s string) string {
	// Order matters: escape backslashes first to avoid double-escaping
	s = strings.ReplaceAll(s, `\`, `\\`)   // \ → \\
	s = strings.ReplaceAll(s, "`", "\\`")  // ` → \`
	s = strings.ReplaceAll(s, "${", "\\${") // ${ → \${
	return s
}

// isAlpineDirective checks if an attribute name is an Alpine.js directive
// Cognitive Load: 3 (Simple prefix check)
func isAlpineDirective(attrName string) bool {
	return strings.HasPrefix(attrName, "x-") ||
		strings.HasPrefix(attrName, "@") ||
		strings.HasPrefix(attrName, ":")
}
