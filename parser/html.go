package parser

import (
	"fmt"
	"log"
	"strings"

	"github.com/jimafisk/custom_go_template/ast"
)

// Global counter to track element nesting depth
var elementParserDepth int

const maxElementDepth = 100 // Adjust as needed

// ElementParser parses HTML elements and their children
func ElementParser() Parser {
	return func(input string) Result {
		// Track element depth to prevent infinite recursion
		elementParserDepth++
		log.Printf("[ElementParser] DEPTH++ = %d, starting parse of: '%.30s...'", elementParserDepth, input)

		// Ensure we decrement the counter on all exit paths
		defer func() {
			elementParserDepth--
			log.Printf("[ElementParser] DEPTH-- = %d after parse", elementParserDepth)
		}()

		if elementParserDepth > maxElementDepth {
			return Result{nil, input, false, fmt.Sprintf("maximum element nesting depth (%d) exceeded", maxElementDepth), false}
		}

		// --- Opening Tag ---
		startRes := String("<")(input)
		if !startRes.Successful {
			log.Printf("[ElementParser] Not an element tag (doesn't start with <)")
			return Result{nil, input, false, "not an element tag", false}
		}
		remaining := startRes.Remaining

		// Check if it's something else (closing tag, doctype, or comment)
		// Note: We're no longer checking for uppercase letters (components) here
		// to allow the ComponentParser to handle them
		if strings.HasPrefix(remaining, "/") ||
			strings.HasPrefix(remaining, "=") ||
			strings.HasPrefix(strings.ToUpper(remaining), "DOCTYPE") ||
			strings.HasPrefix(remaining, "!--") {
			log.Printf("[ElementParser] Not a standard element: %.10s...", remaining)
			return Result{nil, input, false, "not a standard HTML element opening tag", false}
		}

		tagNameRes := Identifier()(remaining)
		if !tagNameRes.Successful {
			log.Printf("[ElementParser] Failed to parse tag name")
			return Result{nil, input, false, "invalid tag name", false}
		}

		tagName := tagNameRes.Value.(string)
		remaining = tagNameRes.Remaining
		log.Printf("[ElementParser] Identified tag: <%s>", tagName)

		// --- Attributes ---
		attributes := []ast.Attribute{}
		for {
			wsRes := Whitespace()(remaining)
			remaining = wsRes.Remaining

			// Check if we're at the end of attributes
			if len(remaining) == 0 || strings.HasPrefix(remaining, ">") || strings.HasPrefix(remaining, "/>") {
				break
			}

			// Parse attribute using the enhanced attribute parser
			attrRes := EnhancedAttributeParser()(remaining)
			if !attrRes.Successful {
				log.Printf("[ElementParser] Failed to parse attribute: %s", attrRes.Error)
				return Result{nil, input, false, fmt.Sprintf("failed to parse attribute for <%s>: %s", tagName, attrRes.Error), false}
			}

			// Add the attribute to our list
			if attr, ok := attrRes.Value.(ast.Attribute); ok {
				attributes = append(attributes, attr)
				log.Printf("[ElementParser] Parsed attribute: %s (Value: %s)", attr.Name, attr.Value)
			} else {
				log.Printf("[ElementParser] Warning: Attribute parser returned non-attribute value: %T", attrRes.Value)
			}

			// Ensure we're making progress
			if attrRes.Remaining == remaining {
				log.Printf("[ElementParser] Warning: Attribute parser made no progress - breaking attribute loop")
				break
			}

			remaining = attrRes.Remaining
		}

		log.Printf("[ElementParser] Finished parsing %d attributes for <%s>", len(attributes), tagName)

		// --- Tag Closing ---
		wsRes := Whitespace()(remaining)
		remaining = wsRes.Remaining
		selfClosing := false

		if len(remaining) == 0 {
			log.Printf("[ElementParser] Unexpected end of input after attributes")
			return Result{nil, input, false, "unexpected end of input after attributes", false}
		}

		if strings.HasPrefix(remaining, "/>") {
			selfClosing = true
			remaining = remaining[2:]
			log.Printf("[ElementParser] Found self-closing tag <%s/>", tagName)
		} else if strings.HasPrefix(remaining, ">") {
			remaining = remaining[1:]
			log.Printf("[ElementParser] Found opening tag <%s>", tagName)
		} else {
			log.Printf("[ElementParser] Expected > or /> to close opening tag, got: %.10s...", remaining)
			return Result{nil, input, false, "invalid tag closing", false}
		}

		// --- Children ---
		children := []ast.Node{}
		if !selfClosing && !isVoidElement(tagName) {
			log.Printf("[ElementParser] <%s>: Starting to parse children", tagName)
			childrenRes := parseChildren(remaining, tagName)
			if !childrenRes.Successful {
				log.Printf("[ElementParser] <%s>: Failed to parse children: %s", tagName, childrenRes.Error)
				return Result{nil, input, false, fmt.Sprintf("failed to parse children for <%s>: %s", tagName, childrenRes.Error), false}
			}

			if childNodes, ok := childrenRes.Value.([]ast.Node); ok {
				children = childNodes
				log.Printf("[ElementParser] <%s>: Finished parsing with %d children", tagName, len(children))
			} else {
				log.Printf("[ElementParser] <%s>: Warning: Children parser returned non-node value: %T", tagName, childrenRes.Value)
			}

			remaining = childrenRes.Remaining
		}

		// Create the element node
		element := &ast.Element{
			TagName:     tagName,
			Attributes:  attributes,
			Children:    children,
			SelfClosing: selfClosing,
		}

		log.Printf("[ElementParser] Successfully parsed <%s>. Remaining: '%.30s...'", tagName, remaining)
		return Result{element, remaining, true, "", false}
	}
}

// Helper struct to store tag information
type tagInfo struct {
	name        string
	attributes  []ast.Attribute
	selfClosing bool
}

// parseOpeningTag handles parsing the opening tag including tag name and attributes
func parseOpeningTag(input string) Result {
	startRes := String("<")(input)
	if !startRes.Successful {
		return Result{nil, input, false, "not an element tag", false}
	}
	remaining := startRes.Remaining

	tagNameRes := Identifier()(remaining)
	if !tagNameRes.Successful {
		return Result{nil, input, false, "invalid tag name", false}
	}

	tagName := tagNameRes.Value.(string)
	remaining = tagNameRes.Remaining

	attributes, remaining, selfClosing, err := parseAttributes(remaining, tagName)
	if err != "" {
		return Result{nil, input, false, err, false}
	}

	info := tagInfo{
		name:        tagName,
		attributes:  attributes,
		selfClosing: selfClosing,
	}

	return Result{info, remaining, true, "", false}
}

// parseAttributes parses all attributes in an opening tag
func parseAttributes(input string, tagName string) ([]ast.Attribute, string, bool, string) {
	attributes := []ast.Attribute{}
	remaining := input
	selfClosing := false

	for {
		wsRes := Whitespace()(remaining)
		remaining = wsRes.Remaining

		if len(remaining) == 0 {
			return nil, input, false, "unexpected end of input after attributes"
		}

		if strings.HasPrefix(remaining, "/>") {
			selfClosing = true
			remaining = remaining[2:]
			break
		}

		if strings.HasPrefix(remaining, ">") {
			remaining = remaining[1:]
			break
		}

		attrRes := EnhancedAttributeParser()(remaining)
		if !attrRes.Successful {
			return nil, input, false, fmt.Sprintf("failed to parse attribute for <%s>: %s", tagName, attrRes.Error)
		}

		if attr, ok := attrRes.Value.(ast.Attribute); ok {
			attributes = append(attributes, attr)
		}

		if attrRes.Remaining == remaining {
			break
		}

		remaining = attrRes.Remaining
	}

	return attributes, remaining, selfClosing, ""
}

// parseChildren parses all children of an element until its closing tag
func parseChildren(input string, parentTag string) Result {
	children := []ast.Node{}
	remaining := input

	for {
		if len(remaining) == 0 {
			return Result{nil, input, false, fmt.Sprintf("unexpected end of input while parsing children of <%s>", parentTag), false}
		}

		// Check for closing tag
		if strings.HasPrefix(remaining, "</") {
			closeTagStart := remaining
			remaining = remaining[2:] // Skip </

			// Extract tag name
			wsRes := Whitespace()(remaining)
			remaining = wsRes.Remaining

			tagNameRes := Identifier()(remaining)
			if !tagNameRes.Successful {
				return Result{nil, input, false, "invalid closing tag name", false}
			}

			closingTagName := tagNameRes.Value.(string)
			remaining = tagNameRes.Remaining

			wsRes = Whitespace()(remaining)
			remaining = wsRes.Remaining

			if !strings.HasPrefix(remaining, ">") {
				return Result{nil, input, false, "invalid closing tag format", false}
			}
			remaining = remaining[1:] // Skip >

			if closingTagName != parentTag {
				log.Printf("[parseChildren] Warning: Mismatched closing tag. Expected </%s>, got </%s>", parentTag, closingTagName)
				// Add the closing tag as text and continue
				textNode := &ast.TextNode{Content: closeTagStart[:len(closeTagStart)-len(remaining)]}
				children = append(children, textNode)
				continue
			}

			// Found the matching closing tag
			break
		}

		// Parse a child node
		childRes := parseChildNode(remaining)
		if childRes.Successful {
			if childNode, ok := childRes.Value.(ast.Node); ok {
				children = append(children, childNode)
			} else if childNodes, ok := childRes.Value.([]ast.Node); ok {
				children = append(children, childNodes...)
			}
			remaining = childRes.Remaining
		} else {
			// If we can't parse a node, treat the next character as text
			if len(remaining) > 0 {
				appendCharToChildren(string(remaining[0]), &children)
				remaining = remaining[1:]
			} else {
				break
			}
		}
	}

	return Result{children, remaining, true, "", false}
}

// parseChildNode attempts to parse a single child node
func parseChildNode(input string) Result {
	// Try to parse as element
	elemRes := ElementParser()(input)
	if elemRes.Successful {
		return elemRes
	}

	// Try to parse as expression
	exprRes := ExpressionParser()(input)
	if exprRes.Successful {
		return exprRes
	}

	// Try to parse as comment
	commentRes := CommentParser()(input)
	if commentRes.Successful {
		return commentRes
	}

	// Try to parse as conditional
	condRes := ConditionalParser()(input)
	if condRes.Successful {
		return condRes
	}

	// Try to parse as loop
	loopRes := LoopParser()(input)
	if loopRes.Successful {
		return loopRes
	}

	// Try to parse as component
	compRes := ComponentParser()(input)
	if compRes.Successful {
		return compRes
	}

	// Try to parse as fence
	fenceRes := FenceParser()(input)
	if fenceRes.Successful {
		return fenceRes
	}

	// Try to parse as text node (any character)
	if len(input) > 0 {
		char := string(input[0])

		// Skip < as it might be the start of a tag
		if char == "<" {
			return Result{nil, input, false, "not a valid child node", false}
		}

		textNode := &ast.TextNode{Content: char}
		return Result{textNode, input[1:], true, "", false}
	}

	return Result{nil, input, false, "empty input", false}
}

// appendCharToChildren adds a character to the last text node or creates a new one
func appendCharToChildren(char string, children *[]ast.Node) {
	if len(*children) > 0 {
		if textNode, ok := (*children)[len(*children)-1].(*ast.TextNode); ok {
			textNode.Content += char
			return
		}
	}

	// Create a new text node
	*children = append(*children, &ast.TextNode{Content: char})
}

// isVoidElement checks if a tag is a void element that can't have children
func isVoidElement(tagName string) bool {
	voidElements := map[string]bool{
		"area": true, "base": true, "br": true, "col": true, "embed": true,
		"hr": true, "img": true, "input": true, "link": true, "meta": true,
		"param": true, "source": true, "track": true, "wbr": true,
	}

	return voidElements[tagName]
}

// hasAlpineAttributes checks if an element has any Alpine.js directives
func hasAlpineAttributes(attributes []ast.Attribute) bool {
	for _, attr := range attributes {
		if attr.IsAlpine {
			return true
		}
	}
	return false
}

// EnhancedAttributeParser handles all types of HTML attributes, with special handling for Alpine.js
func EnhancedAttributeParser() Parser {
	return func(input string) Result {
		// Parse the attribute name
		nameRes := AttributeNameParser()(input)
		if !nameRes.Successful {
			return Result{nil, input, false, "invalid attribute name", false}
		}

		name := nameRes.Value.(string)
		remaining := nameRes.Remaining

		// Check if this is an Alpine directive
		alpineInfo := parseAlpineDirective(name)

		// Skip whitespace after name
		wsRes := Whitespace()(remaining)
		remaining = wsRes.Remaining

		// Check if there's a value
		hasValue := false
		var valueResult Result
		var value interface{}
		dynamic := false

		if strings.HasPrefix(remaining, "=") {
			hasValue = true
			remaining = remaining[1:] // Skip =

			// Skip whitespace after =
			wsRes = Whitespace()(remaining)
			remaining = wsRes.Remaining

			// Special handling for x-data attribute which can contain complex objects
			if alpineInfo.isAlpine && alpineInfo.directiveType == "data" {
				dataRes := parseAlpineDataAttribute(remaining)
				if !dataRes.Successful {
					return Result{nil, input, false, fmt.Sprintf("invalid x-data value: %s", dataRes.Error), false}
				}
				value = dataRes.Value
				remaining = dataRes.Remaining
				dynamic = dataRes.Dynamic
			} else {
				// Regular attribute value
				valueResult = parseAttributeValue(remaining)
				if !valueResult.Successful {
					return Result{nil, input, false, fmt.Sprintf("invalid attribute value: %s", valueResult.Error), false}
				}
				value = valueResult.Value
				remaining = valueResult.Remaining

				// Check if it's a dynamic value (expression)
				if exprNode, ok := value.(*ast.ExpressionNode); ok {
					value = exprNode.Expression
					dynamic = true
				}
			}
		}

		// Create the attribute
		attr := ast.Attribute{
			Name:       name,
			Value:      "",
			Dynamic:    dynamic,
			IsAlpine:   alpineInfo.isAlpine,
			AlpineType: alpineInfo.directiveType,
			AlpineKey:  alpineInfo.key,
		}

		if hasValue && value != nil {
			if strValue, ok := value.(string); ok {
				attr.Value = strValue
			} else {
				// Convert other types to string
				attr.Value = fmt.Sprintf("%v", value)
			}
		}

		return Result{attr, remaining, true, "", false}
	}
}

// Special parser for x-data attribute values which can contain complex JavaScript object literals
func parseAlpineDataAttribute(input string) ValueResult {
	// Check for opening quote
	if len(input) == 0 {
		return ValueResult{nil, input, false, "empty input", false}
	}

	quoteChar := input[0]
	if quoteChar != '"' && quoteChar != '\'' {
		// Try to parse as an expression
		exprRes := ExpressionParser()(input)
		if exprRes.Successful {
			if exprNode, ok := exprRes.Value.(*ast.ExpressionNode); ok {
				return ValueResult{exprNode.Expression, exprRes.Remaining, true, "", true}
			}
		}
		return ValueResult{nil, input, false, "x-data value must be quoted or an expression", false}
	}

	// Parse the complex value
	return parseComplexAlpineValue(input)
}

// parseAttributeValue handles regular attribute values
func parseAttributeValue(input string) Result {
	if len(input) == 0 {
		return Result{nil, input, false, "empty input", false}
	}

	// Check for expression
	if strings.HasPrefix(input, "{") && !strings.HasPrefix(input, "{") {
		exprRes := ExpressionParser()(input)
		if exprRes.Successful {
			return exprRes
		}
	}

	// Check for quoted string
	quoteChar := input[0]
	if quoteChar == '"' {
		return DoubleQuotedString()(input)
	} else if quoteChar == '\'' {
		return SingleQuotedString()(input)
	}

	// Unquoted value (up to whitespace or >)
	var builder strings.Builder
	i := 0

	for i < len(input) {
		char := input[i]

		if char == ' ' || char == '\t' || char == '\n' || char == '\r' || char == '>' || char == '/' {
			break
		}

		builder.WriteByte(char)
		i++
	}

	if builder.Len() == 0 {
		return Result{nil, input, false, "empty attribute value", false}
	}

	return Result{builder.String(), input[i:], true, "", false}
}

// parseComplexAlpineValue handles Alpine.js values with complex structure
func parseComplexAlpineValue(input string) ValueResult {
	if len(input) == 0 {
		return ValueResult{nil, input, false, "empty input", false}
	}

	quoteChar := input[0]
	if quoteChar != '"' && quoteChar != '\'' {
		return ValueResult{nil, input, false, "value must start with a quote", false}
	}

	var valueBuilder strings.Builder
	remaining := input[1:] // Skip opening quote

	braceCount := 0
	inSingleQuote := false
	inDoubleQuote := false
	escaped := false
	i := 0

	for i < len(remaining) {
		char := remaining[i]

		if escaped {
			valueBuilder.WriteByte(char)
			escaped = false
		} else if char == '\\' {
			escaped = true
		} else if char == '\'' && !inDoubleQuote {
			inSingleQuote = !inSingleQuote
			valueBuilder.WriteByte(char)
		} else if char == '"' && !inSingleQuote {
			inDoubleQuote = !inDoubleQuote
			valueBuilder.WriteByte(char)
		} else if char == '{' && !inSingleQuote && !inDoubleQuote {
			braceCount++
			valueBuilder.WriteByte(char)
		} else if char == '}' && !inSingleQuote && !inDoubleQuote {
			braceCount--
			valueBuilder.WriteByte(char)
		} else if char == quoteChar && !inSingleQuote && !inDoubleQuote && !escaped {
			// Found closing quote
			break
		} else {
			valueBuilder.WriteByte(char)
		}

		i++
	}

	if i >= len(remaining) {
		return ValueResult{nil, input, false, "unclosed complex Alpine value", false}
	}

	return ValueResult{valueBuilder.String(), remaining[i+1:], true, "", true}
}

// DoubleQuotedString parses a string enclosed in double quotes
func DoubleQuotedString() Parser {
	return func(input string) Result {
		if !strings.HasPrefix(input, `"`) {
			return Result{nil, input, false, "not a double-quoted string", false}
		}

		var builder strings.Builder
		i := 1 // Skip opening quote
		escaped := false

		for i < len(input) {
			char := input[i]

			if escaped {
				builder.WriteByte(char)
				escaped = false
			} else if char == '\\' {
				escaped = true
			} else if char == '"' {
				// Found our closing quote
				return Result{builder.String(), input[i+1:], true, "", false}
			} else {
				builder.WriteByte(char)
			}

			i++
		}

		// If we got here, we never found the closing quote
		return Result{nil, input, false, "unclosed double-quoted string", false}
	}
}

// SingleQuotedString parses a string enclosed in single quotes
func SingleQuotedString() Parser {
	return func(input string) Result {
		if !strings.HasPrefix(input, `'`) {
			return Result{nil, input, false, "not a single-quoted string", false}
		}

		var builder strings.Builder
		i := 1 // Skip opening quote
		escaped := false

		for i < len(input) {
			char := input[i]

			if escaped {
				builder.WriteByte(char)
				escaped = false
			} else if char == '\\' {
				escaped = true
			} else if char == '\'' {
				// Found our closing quote
				return Result{builder.String(), input[i+1:], true, "", false}
			} else {
				builder.WriteByte(char)
			}

			i++
		}

		// If we got here, we never found the closing quote
		return Result{nil, input, false, "unclosed single-quoted string", false}
	}
}

// CommentParser parses HTML comments
func CommentParser() Parser {
	return Map(
		Between(String("<!--"), String("-->"), TakeUntil(String("-->"))),
		func(value interface{}) (interface{}, error) {
			content := value.(string)
			log.Printf("[CommentParser] Parsed comment: %.30s...", content)
			return &ast.CommentNode{Content: content}, nil
		},
	)
}

// Helper struct for Alpine directive parsing
type alpineDirectiveInfo struct {
	isAlpine      bool
	directiveType string
	key           string
}

// parseAlpineDirective analyzes an attribute name and extracts Alpine.js directive info
func parseAlpineDirective(name string) alpineDirectiveInfo {
	if strings.HasPrefix(name, "x-") {
		parts := strings.SplitN(name[2:], ":", 2)
		directiveType := parts[0]
		key := ""
		if len(parts) > 1 {
			key = parts[1]
		}
		return alpineDirectiveInfo{true, directiveType, key}
	} else if name == "@" || strings.HasPrefix(name, "@") {
		key := ""
		if name != "@" {
			key = name[1:] // Extract event name
		}
		return alpineDirectiveInfo{true, "on", key}
	} else if name == ":" || strings.HasPrefix(name, ":") {
		key := ""
		if name != ":" {
			key = name[1:] // Extract binding key
		}
		return alpineDirectiveInfo{true, "bind", key}
	}

	return alpineDirectiveInfo{false, "", ""}
}

// Extended Result type that includes Dynamic field
type ValueResult struct {
	Value      interface{}
	Remaining  string
	Successful bool
	Error      string
	Dynamic    bool
}

// AttributeNameParser parses HTML attribute names
func AttributeNameParser() Parser {
	return func(input string) Result {
		if len(input) == 0 {
			return Result{nil, input, false, "empty input", false}
		}

		// Special case for @ shorthand
		if input[0] == '@' {
			// Check if it's just @ or @something
			if len(input) == 1 {
				return Result{"@", input[1:], true, "", false}
			}

			// Parse the rest as an identifier
			i := 1
			for i < len(input) {
				char := input[i]
				if !isValidAttributeNameChar(char) {
					break
				}
				i++
			}

			if i == 1 {
				return Result{"@", input[1:], true, "", false}
			}

			return Result{input[:i], input[i:], true, "", false}
		}

		// Special case for : shorthand
		if input[0] == ':' {
			// Check if it's just : or :something
			if len(input) == 1 {
				return Result{":", input[1:], true, "", false}
			}

			// Parse the rest as an identifier
			i := 1
			for i < len(input) {
				char := input[i]
				if !isValidAttributeNameChar(char) {
					break
				}
				i++
			}

			if i == 1 {
				return Result{":", input[1:], true, "", false}
			}

			return Result{input[:i], input[i:], true, "", false}
		}

		// Regular attribute name
		i := 0
		for i < len(input) {
			char := input[i]
			if !isValidAttributeNameChar(char) {
				break
			}
			i++
		}

		if i == 0 {
			return Result{nil, input, false, "invalid attribute name", false}
		}

		return Result{input[:i], input[i:], true, "", false}
	}
}

// isValidAttributeNameChar checks if a character is valid in an attribute name
func isValidAttributeNameChar(char byte) bool {
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9') ||
		char == '-' || char == '_' || char == '.' || char == ':'
}
