package builder

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jimafisk/custom_go_template/ast"
	"github.com/jimafisk/custom_go_template/types"
)

// ComponentTemplate is an alias for types.ComponentTemplate for backward compatibility.
// New code should use types.ComponentTemplate directly.
type ComponentTemplate = types.ComponentTemplate

// RenderContext tracks rendering context to handle expressions differently in different contexts
// Pattern: Context Object Pattern [Load: 5]
type RenderContext struct {
	insideLiteral    bool            // Inside <style> or <script> tags where expressions should NOT be converted
	insideAlpineAttr bool            // Inside Alpine.js directive where expressions should be preserved
	loopVars         map[string]bool // Track loop variables (iterator, value) that should NOT get props. prefix
}

// GenerateComponentRegistry converts component ASTs to ES module with template literal factories.
// For Plenti compatibility, uses signatures as primary keys with short-name aliases.
//
// Output format:
//
//	const registry = {
//	  'layouts_components_Hero2436_html': (props) => `...`,
//	  'Hero2436': (props) => registry['layouts_components_Hero2436_html'](props),
//	};
//
// Cognitive Load: 10 (Main orchestration with signature handling)
// Pattern: Service Implementation Pattern
func GenerateComponentRegistry(components []ComponentTemplate) string {
	if len(components) == 0 {
		return "export default {};"
	}

	var sb strings.Builder

	// Header comment
	sb.WriteString("// Auto-generated Plenti-compatible component registry\n")
	sb.WriteString("// Signature format: layouts_{category}_{name}_{extension}\n")
	sb.WriteString("// Lookup: registry['layouts_components_' + name + '_html']\n\n")

	sb.WriteString("const registry = {\n")

	// Track which signatures we've added to avoid duplicates
	addedSignatures := make(map[string]bool)
	// Track short names for alias generation
	shortNameAliases := make(map[string]string) // shortName -> signature

	// First pass: Add components by their signature (primary key)
	for i, component := range components {
		// Determine the signature - use component.Signature if available,
		// otherwise generate from Name
		signature := component.Signature
		if signature == "" {
			signature = types.GenerateSignature(component.Name)
		}

		// Skip if we've already added this signature
		if addedSignatures[signature] {
			continue
		}
		addedSignatures[signature] = true

		// Generate component entry with signature as key
		sb.WriteString(fmt.Sprintf("  '%s': (props) => `", signature))

		// Convert AST to JavaScript template literal
		templateContent := convertToJSTemplate(component.AST)
		sb.WriteString(templateContent)

		sb.WriteString("`")

		// Track short name for alias
		shortName := component.Name
		if shortName == "" {
			shortName = types.SignatureToShortName(signature)
		}
		if shortName != "" && shortName != signature {
			shortNameAliases[shortName] = signature
		}

		// Add comma separator
		if i < len(components)-1 || len(shortNameAliases) > 0 {
			sb.WriteString(",\n")
		} else {
			sb.WriteString("\n")
		}
	}

	// Second pass: Add short-name aliases for backward compatibility
	aliasCount := 0
	totalAliases := len(shortNameAliases)
	for shortName, signature := range shortNameAliases {
		// Skip if short name is same as signature (no alias needed)
		if shortName == signature {
			continue
		}
		// Skip if short name was already added as a signature
		if addedSignatures[shortName] {
			continue
		}

		aliasCount++
		sb.WriteString(fmt.Sprintf("  '%s': (props) => registry['%s'](props)", shortName, signature))

		if aliasCount < totalAliases {
			sb.WriteString(",\n")
		} else {
			sb.WriteString("\n")
		}
	}

	sb.WriteString("};\n\nexport default registry;")
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
		loopVars:         make(map[string]bool),
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
		} else if expressionReferencesLoopVar(n.Expression, ctx.loopVars) {
			// LOOP VARIABLE FIX: Expression references loop variable → use Alpine x-text directive
			// Cannot use ${...} template literal because loop variable doesn't exist during template function execution
			sb.WriteString(`<span x-text="`)
			sb.WriteString(n.Expression)
			sb.WriteString(`"></span>`)
		} else {
			// Use identifier-level prefixing to handle complex expressions
			// ARROW FUNCTION FIX: Extract arrow function parameters before prefixing
			arrowParams := extractArrowFunctionParams(n.Expression)
			// LOOP VARIABLE FIX: Pass loop variables to skip list
			converted := prefixIdentifiersInExpression(n.Expression, arrowParams, ctx.loopVars)
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

	case *ast.StyleSection:
		// Render <style> tags with CSS content
		// This is critical for component-specific styling
		sb.WriteString("<style>")
		sb.WriteString(escapeTemplateLiteral(n.Content))
		sb.WriteString("</style>")

	case *ast.ScriptSection:
		// Render <script> tags with JavaScript content
		sb.WriteString("<script>")
		sb.WriteString(escapeTemplateLiteral(n.Content))
		sb.WriteString("</script>")

	// Ignore fence sections - they're metadata, not component markup
	case *ast.FenceSection:
		// Skip fence section (props, variables, etc.)

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
		insideAlpineAttr: false,        // Reset for element children
		loopVars:         ctx.loopVars, // CRITICAL: Preserve loop variables from parent context
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
func convertAttributeExpressions(attrValue string, loopVars map[string]bool) string {
	// Find all {expression} patterns
	return expressionPattern.ReplaceAllStringFunc(attrValue, func(match string) string {
		// Extract the expression without braces
		expr := match[1 : len(match)-1] // Remove { and }

		// Check if this looks like an Alpine object literal content (after braces are stripped)
		// An object literal contains : for key-value pairs
		if strings.Contains(expr, ":") && !strings.Contains(expr, "?") {
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
		converted := prefixIdentifiersInExpression(expr, arrowParams, loopVars)

		return "${" + converted + "}"
	})
}

// extractArrowFunctionParams extracts parameter names from arrow functions in an expression
// Returns a map of parameter names to skip during identifier prefixing
// Cognitive Load: 12 (Improved regex with state tracking)
// Pattern: Arrow function parameter detection
//
// CRITICAL FIX: This function now properly handles nested method calls
// Example: ".reduce((sum, p) => ..." now correctly extracts ["sum", "p"]
// Previous bug: Would capture "(" from ".reduce(" leading to invalid param names
func extractArrowFunctionParams(expr string) map[string]bool {
	params := make(map[string]bool)

	// Pattern: Find all "=>" occurrences and work backwards to find params
	// This approach is more reliable than trying to match the full pattern

	offset := 0
	for {
		// Find next =>
		arrowIndex := strings.Index(expr[offset:], "=>")
		if arrowIndex == -1 {
			break
		}
		arrowIndex += offset // Adjust to absolute position

		// Look backwards from => to find the parameter list
		// Skip whitespace before =>
		paramEnd := arrowIndex - 1
		for paramEnd >= 0 && (expr[paramEnd] == ' ' || expr[paramEnd] == '\t') {
			paramEnd--
		}

		if paramEnd < 0 {
			offset = arrowIndex + 2
			continue
		}

		// Check if params are in parens: (sum, p) =>
		if expr[paramEnd] == ')' {
			// Find matching opening paren by tracking depth
			parenDepth := 1
			paramStart := paramEnd - 1
			for paramStart >= 0 && parenDepth > 0 {
				if expr[paramStart] == ')' {
					parenDepth++
				} else if expr[paramStart] == '(' {
					parenDepth--
				}
				paramStart--
			}

			if parenDepth == 0 {
				// Extract params between parens
				paramStr := expr[paramStart+2 : paramEnd] // +2 to skip opening (
				extractParamNames(paramStr, params)
			}
		} else {
			// Single param without parens: item =>
			// Work backwards to find the identifier
			paramStart := paramEnd
			for paramStart >= 0 && isIdentifierChar(expr[paramStart]) {
				paramStart--
			}

			if paramStart < paramEnd {
				paramName := expr[paramStart+1 : paramEnd+1]
				paramName = strings.TrimSpace(paramName)
				if paramName != "" && isSimpleIdentifier(paramName) {
					params[paramName] = true
				}
			}
		}

		offset = arrowIndex + 2 // Move past =>
	}

	return params
}

// extractParamNames extracts individual parameter names from a parameter list
// Example: "sum, p" → ["sum", "p"]
// Cognitive Load: 5 (String splitting and cleaning)
func extractParamNames(paramStr string, params map[string]bool) {
	// Split by comma
	paramList := strings.Split(paramStr, ",")
	for _, param := range paramList {
		param = strings.TrimSpace(param)

		// Remove any destructuring syntax (for now, just skip)
		if strings.Contains(param, "{") || strings.Contains(param, "[") {
			continue
		}

		// Extract just the identifier
		if param != "" && isSimpleIdentifier(param) {
			params[param] = true
		}
	}
}

// isIdentifierChar checks if a character can be part of a JavaScript identifier
// Cognitive Load: 2 (Simple character check)
func isIdentifierChar(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') ||
		(ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') ||
		ch == '_' || ch == '$'
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
// Cognitive Load: 22 (Complex string traversal with string literal tracking)
// ARROW FUNCTION FIX: Now accepts arrowParams map to skip arrow function parameters
// STRING LITERAL FIX: Now tracks string literals to avoid processing content inside quotes
// METHOD CHAIN FIX: Now properly continues method chains after closing parens
// LOOP VARIABLE FIX: Now accepts loopVars map to skip loop variables
func prefixIdentifiersInExpression(expr string, arrowParams map[string]bool, loopVars map[string]bool) string {
	// Merge arrow params and loop vars with skip list
	combinedSkip := make(map[string]bool)
	for k, v := range skipIdentifiers {
		combinedSkip[k] = v
	}
	if arrowParams != nil {
		for k, v := range arrowParams {
			combinedSkip[k] = v
		}
	}
	if loopVars != nil {
		for k, v := range loopVars {
			combinedSkip[k] = v
		}
	}

	// Use a context-aware tokenizer that distinguishes method calls from grouping parens
	// AND tracks string literals to skip processing inside them
	var result strings.Builder
	var currentToken strings.Builder
	parenDepth := 0       // Track parentheses depth
	isMethodCall := false // Track if current parens are for a method call
	inString := false     // Track if we're inside a string literal
	stringChar := byte(0) // Track which quote character started the string (' or ")
	escaped := false      // Track if previous char was backslash

	for i := 0; i < len(expr); i++ {
		ch := expr[i]

		// STRING LITERAL HANDLING (NEW)
		// Check for escaped characters
		if escaped {
			// Previous char was \, so this char is escaped
			result.WriteByte(ch)
			escaped = false
			continue
		}

		if ch == '\\' {
			// Escape character - next char should be treated literally
			result.WriteByte(ch)
			escaped = true
			continue
		}

		// Check for string delimiters
		if ch == '\'' || ch == '"' {
			if !inString {
				// Entering string literal
				inString = true
				stringChar = ch

				// Process any accumulated token before entering string
				if currentToken.Len() > 0 {
					token := currentToken.String()
					result.WriteString(processToken(token, combinedSkip))
					currentToken.Reset()
				}

				result.WriteByte(ch)
				continue
			} else if ch == stringChar {
				// Exiting string literal (matching quote)
				inString = false
				stringChar = 0
				result.WriteByte(ch)
				continue
			} else {
				// Different quote character while in string - just part of string content
				result.WriteByte(ch)
				continue
			}
		}

		// If we're inside a string, just copy characters as-is
		if inString {
			result.WriteByte(ch)
			continue
		}

		// REST OF EXISTING LOGIC (for when NOT in string)
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
				// Method call - process the method name/chain first, then recursively process args
				methodName := currentToken.String()
				currentToken.Reset()

				// Process and append the method name
				result.WriteString(processToken(methodName, combinedSkip))
				result.WriteByte('(')

				// Find the matching closing paren and recursively process the arguments
				argStart := i + 1
				parenDepth = 1
				argEnd := argStart

				for argEnd < len(expr) && parenDepth > 0 {
					if expr[argEnd] == '(' {
						parenDepth++
					} else if expr[argEnd] == ')' {
						parenDepth--
					}
					if parenDepth > 0 {
						argEnd++
					}
				}

				// Recursively process arguments
				if argEnd > argStart {
					args := expr[argStart:argEnd]
					argsArrowParams := extractArrowFunctionParams(args)
					// Pass loopVars through recursive call
					processedArgs := prefixIdentifiersInExpression(args, argsArrowParams, loopVars)
					result.WriteString(processedArgs)
				}

				result.WriteByte(')')
				i = argEnd // Skip past the closing paren
				parenDepth = 0
				isMethodCall = false

				// BUG FIX #2: Check if method chain continues after closing paren
				// Example: .split('').reverse() - after split('') check for .reverse()
				// If next char is '.', add it to currentToken to continue the chain
				// This prevents the next method from being treated as a new token
				if i+1 < len(expr) && expr[i+1] == '.' {
					currentToken.WriteByte('.')
					i++ // Skip the '.' in the next iteration so it's not processed again
				}
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
			if parenDepth > 0 {
				// End of grouping parens - process token and keep paren as operator
				if currentToken.Len() > 0 {
					token := currentToken.String()
					result.WriteString(processToken(token, combinedSkip))
					currentToken.Reset()
				}
				result.WriteByte(ch)
				parenDepth--
			} else {
				// Stray closing paren - shouldn't happen, but handle gracefully
				result.WriteByte(ch)
			}
		} else if ch == '.' {
			// Check if this is the spread operator (...) vs property access (.)
			// Spread operator should be treated as an operator, not part of identifier
			if i+2 < len(expr) && expr[i+1] == '.' && expr[i+2] == '.' {
				// This is spread operator (...) - process accumulated token first
				if currentToken.Len() > 0 {
					token := currentToken.String()
					result.WriteString(processToken(token, combinedSkip))
					currentToken.Reset()
				}
				// Write spread operator as-is
				result.WriteString("...")
				i += 2   // Skip the next two dots
				continue // Skip to next iteration (don't process current char again)
			} else {
				// Property access - keep as part of token
				currentToken.WriteByte(ch)
			}
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
		ch == ',' || ch == '?' || ch == ':' || ch == '!' || ch == ';' ||
		ch == '>' || ch == '<' || ch == '=' || ch == '&' || ch == '|' ||
		ch == ' ' || ch == '\t' || ch == '\n'
}

// processToken processes a single token (identifier, property chain, or method call)
// Cognitive Load: 10 (Token classification with method call handling)
// ARROW FUNCTION FIX: Now accepts skipList parameter
// NOTE: Method call arguments are now processed recursively by prefixIdentifiersInExpression
func processToken(token string, skipList map[string]bool) string {
	token = strings.TrimSpace(token)
	if token == "" {
		return token
	}

	// BUG FIX #2: If token starts with '.', it's a continuation of a method chain
	// that was already processed. Just return it as-is without prefixing.
	// Example: after "animal.split()", if we see ".reverse", return it unchanged
	if strings.HasPrefix(token, ".") {
		return token
	}
	// Check if it's a property access chain (contains .)
	// Note: Method calls with () are now handled in prefixIdentifiersInExpression
	if strings.Contains(token, ".") {
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

// escapeQuotesInAttributeValue escapes double quotes that aren't inside {...} expressions
// BUG FIX #1: This prevents: x-if="value == "test"" from breaking the attribute
// Cognitive Load: 8 (State machine for expression tracking)
func escapeQuotesInAttributeValue(value string) string {
	var result strings.Builder
	inExpression := false

	for i := 0; i < len(value); i++ {
		ch := value[i]

		if ch == '{' {
			inExpression = true
			result.WriteByte(ch)
		} else if ch == '}' {
			inExpression = false
			result.WriteByte(ch)
		} else if ch == '"' && !inExpression {
			// Escape double quotes outside of expressions
			result.WriteString(`\"`)
		} else {
			result.WriteByte(ch)
		}
	}

	return result.String()
}

// isEventHandlerAttribute checks if an attribute name is an event handler
// Event handlers like onclick, @click, x-on:click should NOT have {expr} converted
// because they contain executable code, not template data
func isEventHandlerAttribute(name string) bool {
	// Standard HTML event handlers
	if strings.HasPrefix(name, "on") {
		return true
	}
	// Alpine.js @ shorthand (@click, @submit, etc.)
	if strings.HasPrefix(name, "@") {
		return true
	}
	// Alpine.js x-on: syntax (x-on:click, x-on:submit, etc.)
	if strings.HasPrefix(name, "x-on:") {
		return true
	}
	return false
}

// renderAttributeToJS renders an HTML attribute with Alpine-aware handling
// Cognitive Load: 12 (Attribute rendering with expression conversion)
//
// CRITICAL FIX: This function now properly converts {expression} patterns in attribute values
// to ${props.expression} for JavaScript template literals using identifier-level prefixing
// BUG FIX #1: Quote escaping happens BEFORE conversion to avoid escaping quotes in ${...}
// BUG FIX #2: Event handler attributes (onclick, @click) are NOT converted
// LOOP VARIABLE FIX: Attributes with loop variable expressions use Alpine : binding
func renderAttributeToJS(attr ast.Attribute, sb *strings.Builder, ctx *RenderContext, children []ast.Node) {
	// LOOP VARIABLE FIX: Check if attribute value contains loop variable expression
	// If so, convert to Alpine : binding syntax
	// Example: src="{card.icon.src}" → :src="card.icon.src"
	if attributeReferencesLoopVar(attr.Value, ctx.loopVars) {
		// Use Alpine binding syntax
		sb.WriteString(":")
		sb.WriteString(attr.Name)
		sb.WriteString("=\"")

		// Extract expression from braces and write directly
		// {card.icon.src} → card.icon.src
		expr := extractExpressionFromBraces(attr.Value)
		sb.WriteString(expr)
		sb.WriteString("\"")
		return
	}

	// Normal attribute handling (for non-loop-variable attributes)
	sb.WriteString(attr.Name)
	sb.WriteString("=\"")

	escaped := attr.Value

	// Step 1: Escape backslashes and backticks for template literal
	escaped = strings.ReplaceAll(escaped, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, "`", "\\`")

	// Step 2: Escape double quotes BEFORE conversion (for attribute context)
	// BUG FIX #1: Smart escaping - only escape quotes outside of {...} expressions
	// This prevents: x-if="animal == "cat"" from becoming x-if="animal == \"cat\""
	// Instead we get: x-if="animal == \"cat\"" BEFORE conversion, then after:
	// x-if="${props.animal == \"cat\"}" which is correct!
	escaped = escapeQuotesInAttributeValue(escaped)

	// Step 3: Convert {expression} patterns to ${props.expression}
	// This handles cases like: x-data="{ count: {count} }" → x-data="{ count: ${props.count} }"
	// But NOT for event handlers (onclick, @click, etc.) where the code should remain as-is
	if !isEventHandlerAttribute(attr.Name) {
		escaped = convertAttributeExpressions(escaped, ctx.loopVars)
	}

	sb.WriteString(escaped)
	sb.WriteString("\"")
}

// renderConditionalToJS renders conditional blocks as Alpine.js <template> elements
// Cognitive Load: 13 (Conditional structure with context)
func renderConditionalToJS(cond *ast.Conditional, sb *strings.Builder, ctx *RenderContext) {
	// Main if block
	sb.WriteString(`<template x-if="`)
	// BUG FIX: Escape quotes in condition (same issue as attributes)
	escapedCondition := escapeQuotesInAttributeValue(cond.IfCondition)
	sb.WriteString(escapedCondition)
	sb.WriteString(`">`)

	for _, node := range cond.IfContent {
		renderNodeToJS(node, sb, ctx)
	}

	sb.WriteString("</template>")

	// Else-if blocks
	for i, elseIfCond := range cond.ElseIfConditions {
		sb.WriteString(`<template x-else-if="`)
		// BUG FIX: Escape quotes in else-if condition (same issue as if condition)
		escapedElseIfCond := escapeQuotesInAttributeValue(elseIfCond)
		sb.WriteString(escapedElseIfCond)
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
// Cognitive Load: 12 (Loop structure with context and variable tracking)
func renderLoopToJS(loop *ast.Loop, sb *strings.Builder, ctx *RenderContext) {
	sb.WriteString(`<template x-for="`)

	// Build x-for expression
	// Note: In AST, Value = item variable, Iterator = optional index variable
	// For simple loops: "item in collection"
	// For loops with index: "(item, index) in collection"
	iteratorTrimmed := strings.TrimSpace(loop.Iterator)
	if iteratorTrimmed != "" {
		// Loop with index: (item, index) in collection
		sb.WriteString("(")
		sb.WriteString(loop.Value) // item
		sb.WriteString(", ")
		sb.WriteString(iteratorTrimmed) // index
		sb.WriteString(")")
	} else {
		// Simple loop: item in collection
		sb.WriteString(loop.Value) // item only
	}

	sb.WriteString(" in ")

	// CRITICAL FIX: Prefix collection with props. if it's not a loop variable or special identifier
	collection := loop.Collection
	if !skipIdentifiers[collection] && !ctx.loopVars[collection] {
		// Check if it's a property chain - only prefix if root is not in skip list
		if strings.Contains(collection, ".") {
			parts := strings.Split(collection, ".")
			if !skipIdentifiers[parts[0]] && !ctx.loopVars[parts[0]] {
				collection = "props." + collection
			}
		} else {
			collection = "props." + collection
		}
	}
	sb.WriteString(collection)
	sb.WriteString(`">`)

	// CRITICAL FIX: Create new context with loop variables tracked
	loopCtx := &RenderContext{
		insideLiteral:    ctx.insideLiteral,
		insideAlpineAttr: ctx.insideAlpineAttr,
		loopVars:         make(map[string]bool),
	}

	// Copy parent loop vars
	for k, v := range ctx.loopVars {
		loopCtx.loopVars[k] = v
	}

	// Add current loop variables
	if iteratorTrimmed != "" {
		loopCtx.loopVars[iteratorTrimmed] = true
	}
	if loop.Value != "" {
		loopCtx.loopVars[loop.Value] = true
	}

	// Render loop content with loop variables in context
	for _, node := range loop.Content {
		renderNodeToJS(node, sb, loopCtx)
	}

	sb.WriteString("</template>")
}

// escapeTemplateLiteral escapes special characters for JavaScript template literals
// Cognitive Load: 4 (Simple string replacement)
// Pattern: Basic string escaping
func escapeTemplateLiteral(s string) string {
	// Order matters: escape backslashes first to avoid double-escaping
	s = strings.ReplaceAll(s, `\`, `\\`)    // \ → \\
	s = strings.ReplaceAll(s, "`", "\\`")   // ` → \`
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

// expressionReferencesLoopVar checks if an expression references any loop variable
// Cognitive Load: 7 (Expression parsing with identifier extraction)
// Examples:
//   - "text" with loopVars{"text": true} → true
//   - "card.icon.src" with loopVars{"card": true} → true
//   - "props.title" with loopVars{"card": true} → false
func expressionReferencesLoopVar(expr string, loopVars map[string]bool) bool {
	if len(loopVars) == 0 {
		return false
	}

	// Extract root identifier from expression
	// For "card.icon.src" → "card"
	// For "text" → "text"
	// For "item.name + ' - ' + item.price" → check both "item" references

	// Split by operators and delimiters to extract all identifiers
	// Simple approach: check if any loop variable appears as a standalone identifier
	for loopVar := range loopVars {
		// Check if loop variable appears as complete identifier
		// Use word boundary regex to avoid matching "card" in "discard"
		pattern := regexp.MustCompile(`\b` + regexp.QuoteMeta(loopVar) + `\b`)
		if pattern.MatchString(expr) {
			return true
		}
	}

	return false
}

// attributeReferencesLoopVar checks if an attribute value contains expressions that reference loop variables
// Cognitive Load: 6 (Expression extraction with loop var check)
// Examples:
//   - "{card.icon.src}" with loopVars{"card": true} → true
//   - "{props.title}" with loopVars{"card": true} → false
//   - "static-value" → false
func attributeReferencesLoopVar(attrValue string, loopVars map[string]bool) bool {
	if len(loopVars) == 0 {
		return false
	}

	// Find all {expression} patterns in attribute value
	matches := expressionPattern.FindAllStringSubmatch(attrValue, -1)
	for _, match := range matches {
		if len(match) > 1 {
			expr := match[1] // Extract expression without braces
			if expressionReferencesLoopVar(expr, loopVars) {
				return true
			}
		}
	}

	return false
}

// extractExpressionFromBraces extracts the expression content from an attribute value
// Cognitive Load: 5 (Simple extraction)
// Examples:
//   - "{card.icon.src}" → "card.icon.src"
//   - "{text}" → "text"
//   - "prefix {expr} suffix" → "expr" (first expression)
func extractExpressionFromBraces(attrValue string) string {
	// Find first {expression} pattern
	match := expressionPattern.FindStringSubmatch(attrValue)
	if len(match) > 1 {
		return match[1] // Return expression without braces
	}
	// If no braces found, return as-is (shouldn't happen in normal flow)
	return attrValue
}
