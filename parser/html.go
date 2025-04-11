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

		// Check if it's something else (closing tag, component, etc.)
		if strings.HasPrefix(remaining, "/") ||
			strings.HasPrefix(remaining, "=") ||
			(len(remaining) > 0 && remaining[0] >= 'A' && remaining[0] <= 'Z') ||
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
			return Result{nil, input, false, "expected > or /> to close opening tag", false}
		}

		// Check for void elements
		if isVoidElement(tagName) {
			selfClosing = true
			log.Printf("[ElementParser] <%s> is a void element, treating as self-closing", tagName)
		}

		// Parse children of non-self-closing elements
		children := []ast.Node{}
		if !selfClosing {
			log.Printf("[ElementParser] <%s>: Starting to parse children", tagName)

			// Define the closing tag we're looking for
			closingTag := "</" + tagName

			// Keep parsing children until we find our closing tag
			for {
				// Check if we've reached our closing tag
				if strings.HasPrefix(remaining, closingTag) {
					// Try to match the full closing tag: </tagName>
					closingTagParser := Sequence(
						String(closingTag),
						Whitespace(),
						String(">"),
					)

					closeTagResult := closingTagParser(remaining)
					if closeTagResult.Successful {
						log.Printf("[ElementParser] <%s>: Found and consumed closing tag", tagName)
						remaining = closeTagResult.Remaining
						break
					}
					// If it looks like a closing tag but didn't fully match, continue parsing
					log.Printf("[ElementParser] <%s>: Found potential closing tag but it didn't match fully", tagName)
				}

				// If end of input before closing tag, it's an error
				if len(remaining) == 0 {
					log.Printf("[ElementParser] <%s>: Reached end of input before finding closing tag", tagName)

					// For Alpine.js documents, force successful parsing with what we have
					if hasAlpineAttributes(attributes) {
						log.Printf("[ElementParser] <%s>: Has Alpine attributes, treating as successful despite missing closing tag", tagName)
						break
					}

					return Result{nil, input, false, fmt.Sprintf("unclosed tag: <%s>", tagName), false}
				}

				// Try to parse a child node
				childResult := parseChildNode(remaining)

				// Handle the child parsing result
				if !childResult.Successful {
					// If parsing fails, consume a character and try again
					if len(remaining) > 0 {
						char := string(remaining[0])
						log.Printf("[ElementParser] <%s>: Failed to parse child, consuming character: %s", tagName, char)

						// Add character to the last text node or create a new one
						if len(children) > 0 {
							if textNode, ok := children[len(children)-1].(*ast.TextNode); ok {
								textNode.Content += char
							} else {
								children = append(children, &ast.TextNode{Content: char})
							}
						} else {
							children = append(children, &ast.TextNode{Content: char})
						}

						remaining = remaining[1:]
						continue
					} else {
						log.Printf("[ElementParser] <%s>: Unexpected end of input while parsing children", tagName)
						return Result{nil, input, false, "unexpected end of input while parsing children", false}
					}
				}

				// Handle a successful child node parsing
				if node, ok := childResult.Value.(ast.Node); ok && node != nil {
					if textNode, isText := node.(*ast.TextNode); isText && len(children) > 0 {
						if lastNode, isLastText := children[len(children)-1].(*ast.TextNode); isLastText {
							// Merge consecutive text nodes
							lastNode.Content += textNode.Content
							log.Printf("[ElementParser] <%s>: Merged text node: %s", tagName, lastNode.Content)
						} else {
							children = append(children, node)
							log.Printf("[ElementParser] <%s>: Added text node: %s", tagName, textNode.Content)
						}
					} else {
						children = append(children, node)
						log.Printf("[ElementParser] <%s>: Added child node: %T", tagName, node)
					}
				} else if nodes, ok := childResult.Value.([]ast.Node); ok {
					for _, n := range nodes {
						if n != nil {
							children = append(children, n)
							log.Printf("[ElementParser] <%s>: Added child node from slice: %T", tagName, n)
						}
					}
				}

				// Ensure we're making progress
				if childResult.Remaining == remaining {
					log.Printf("[ElementParser] <%s>: Warning - No progress parsing children. Consuming next char.", tagName)
					if len(remaining) > 0 {
						char := string(remaining[0])

						// Add character to the last text node or create a new one
						if len(children) > 0 {
							if textNode, ok := children[len(children)-1].(*ast.TextNode); ok {
								textNode.Content += char
							} else {
								children = append(children, &ast.TextNode{Content: char})
							}
						} else {
							children = append(children, &ast.TextNode{Content: char})
						}

						remaining = remaining[1:] // Force progress
					} else {
						break // End of input
					}
				} else {
					remaining = childResult.Remaining
				}
			}
		}

		log.Printf("[ElementParser] <%s>: Finished parsing with %d children", tagName, len(children))

		// Create the element node
		elemNode := &ast.Element{
			TagName:     tagName,
			Attributes:  attributes,
			Children:    children,
			SelfClosing: selfClosing,
		}

		log.Printf("[ElementParser] Successfully parsed <%s>. Remaining: '%.30s...'", tagName, remaining)
		return Result{elemNode, remaining, true, "", false}
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
	// Check for opening angle bracket
	startRes := String("<")(input)
	if !startRes.Successful {
		return Result{nil, input, false, "not an element tag", false} // Added Dynamic
	}
	remaining := startRes.Remaining

	// Check if it's something else (closing tag, component, etc.)
	if strings.HasPrefix(remaining, "/") ||
		strings.HasPrefix(remaining, "=") ||
		(len(remaining) > 0 && remaining[0] >= 'A' && remaining[0] <= 'Z') ||
		strings.HasPrefix(strings.ToUpper(remaining), "DOCTYPE") ||
		strings.HasPrefix(remaining, "!--") {
		return Result{nil, input, false, "not a standard HTML element opening tag", false} // Added Dynamic
	}

	// Parse tag name
	tagNameRes := Identifier()(remaining)
	if !tagNameRes.Successful {
		return Result{nil, input, false, "invalid tag name", false} // Added Dynamic
	}
	tagName := tagNameRes.Value.(string)
	remaining = tagNameRes.Remaining
	log.Printf("[ElementParser] Identified tag: <%s>", tagName)

	// Parse attributes
	attributes, attrRemaining, attrSuccess, attrError := parseAttributes(remaining, tagName)
	if !attrSuccess {
		return Result{nil, input, false, attrError, false} // Added Dynamic
	}
	remaining = attrRemaining

	// Parse tag closing
	wsRes := Whitespace()(remaining)
	remaining = wsRes.Remaining
	selfClosing := false

	if strings.HasPrefix(remaining, "/>") {
		selfClosing = true
		remaining = remaining[2:]
	} else if strings.HasPrefix(remaining, ">") {
		remaining = remaining[1:]
	} else {
		return Result{nil, input, false, "expected > or /> to close opening tag", false} // Added Dynamic
	}

	// Check for void elements
	if isVoidElement(tagName) {
		selfClosing = true
	}

	return Result{
		tagInfo{
			name:        tagName,
			attributes:  attributes,
			selfClosing: selfClosing,
		},
		remaining,
		true,
		"",
		false, // Added Dynamic
	}
}

// parseAttributes parses all attributes in an opening tag
func parseAttributes(input string, tagName string) ([]ast.Attribute, string, bool, string) {
	attributes := []ast.Attribute{}
	remaining := input

	for {
		wsRes := Whitespace()(remaining)
		remaining = wsRes.Remaining

		// Check if we've reached the end of the attributes
		if strings.HasPrefix(remaining, ">") || strings.HasPrefix(remaining, "/>") {
			break
		}

		// Parse attribute
		attrRes := EnhancedAttributeParser()(remaining) // Use EnhancedAttributeParser
		if !attrRes.Successful {
			return nil, input, false, fmt.Sprintf("failed to parse attribute for <%s> near: %s",
				tagName, remaining[:min(20, len(remaining))])
		}

		if attr, ok := attrRes.Value.(ast.Attribute); ok {
			attributes = append(attributes, attr)
			// log.Printf("[ElementParser] Parsed attribute: %s (Value: %s)", attr.Name, attr.Value) // Logging done in EnhancedAttributeParser
		}

		// Make sure we're making progress
		if attrRes.Remaining == remaining {
			return nil, input, false, "parser not advancing while parsing attributes"
		}

		remaining = attrRes.Remaining
	}

	log.Printf("[ElementParser] Finished parsing attributes for <%s>", tagName)
	return attributes, remaining, true, ""
}

// parseChildren parses all children of an element until its closing tag
func parseChildren(input string, parentTag string) Result {
	children := []ast.Node{}
	remaining := input

	log.Printf("[ElementParser] <%s>: Parsing children starting from: '%s'", parentTag, remaining[:min(30, len(remaining))])

	if isVoidElement(parentTag) {
		// Void elements can't have children
		return Result{children, remaining, true, "", false} // Added Dynamic
	}

	// Define the closing tag pattern
	closingTag := "</" + parentTag

	// Keep parsing children until we find our closing tag
	for {
		// Check if we've reached our closing tag
		if strings.HasPrefix(remaining, closingTag) {
			// Try to match the full closing tag: </tagName>
			closingTagParser := Sequence(
				String(closingTag),
				Whitespace(), // Allow whitespace before >
				String(">"),
			)

			closeTagResult := closingTagParser(remaining)
			if closeTagResult.Successful {
				log.Printf("[ElementParser] <%s>: Found and consumed closing tag.", parentTag)
				remaining = closeTagResult.Remaining
				break
			}
			// If it looks like a closing tag but didn't fully match, continue parsing
		}

		// If end of input before closing tag, it's an error
		if len(remaining) == 0 {
			log.Printf("[ElementParser] <%s>: Error - Reached end of input before finding closing tag.", parentTag)
			return Result{nil, input, false, fmt.Sprintf("unclosed tag: <%s>", parentTag), false} // Added Dynamic
		}

		// Try to parse a child node
		childResult := parseChildNode(remaining)

		// Handle the child parsing result
		if !childResult.Successful {
			// If parsing fails, consume a character and try again
			if len(remaining) > 0 {
				char := string(remaining[0])
				// Try to append to the last text node if possible
				appendCharToChildren(char, &children)
				remaining = remaining[1:]
				continue
			} else {
				return Result{nil, input, false, "unexpected end of input while parsing children", false} // Added Dynamic
			}
		}

		// Handle a successful child node parsing
		if node, ok := childResult.Value.(ast.Node); ok && node != nil {
			if textNode, isText := node.(*ast.TextNode); isText && len(children) > 0 {
				if lastNode, isLastText := children[len(children)-1].(*ast.TextNode); isLastText {
					// Merge consecutive text nodes
					lastNode.Content += textNode.Content
				} else {
					children = append(children, node)
				}
			} else {
				children = append(children, node)
			}
		} else if nodes, ok := childResult.Value.([]ast.Node); ok {
			for _, n := range nodes {
				if n != nil {
					children = append(children, n)
				}
			}
		}

		// Ensure we're making progress
		if childResult.Remaining == remaining {
			log.Printf("[ElementParser] <%s>: Warning - No progress parsing children. Consuming next char.", parentTag)
			if len(remaining) > 0 {
				char := string(remaining[0])
				appendCharToChildren(char, &children)
				remaining = remaining[1:] // Force progress
			} else {
				break // End of input
			}
		} else {
			remaining = childResult.Remaining
		}
	}

	log.Printf("[ElementParser] <%s>: Parsed %d children.", parentTag, len(children))
	return Result{children, remaining, true, "", false} // Added Dynamic
}

// parseChildNode attempts to parse a single child node
func parseChildNode(input string) Result {
	// For debugging
	origInput := input

	// Try to parse a comment first
	commentResult := CommentParser()(input)
	if commentResult.Successful {
		return commentResult
	}

	// Try to parse an expression {name} next
	exprResult := ExpressionParser()(input)
	if exprResult.Successful {
		return exprResult
	}

	// Try to parse an element <tag> next
	elemResult := ElementParser()(input)
	if elemResult.Successful {
		return elemResult
	}

	// Try to parse text until the next special character
	delimiters := []Parser{String("<"), String("{")}
	textResult := TextParser(delimiters...)(input)
	if textResult.Successful && textResult.Value != nil {
		textNode, ok := textResult.Value.(*ast.TextNode)
		if ok && len(textNode.Content) > 0 {
			return textResult
		}
	}

	// Nothing matched
	log.Printf("[parseChildNode] Failed to parse any child node type from: %.30s...", origInput)
	return Result{nil, input, false, "failed to parse child node", false}
}

// appendCharToChildren adds a character to the last text node or creates a new one
func appendCharToChildren(char string, children *[]ast.Node) {
	if len(*children) > 0 {
		if textNode, ok := (*children)[len(*children)-1].(*ast.TextNode); ok {
			textNode.Content += char
			return
		}
	}
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
		log.Printf("[EnhancedAttributeParser] Starting on: '%.30s...'", input)

		// Try parsing a name first
		nameRes := Identifier()(input)
		if !nameRes.Successful {
			log.Printf("[EnhancedAttributeParser] Failed to parse identifier")
			return Result{nil, input, false, "not a valid attribute name", false}
		}

		name := nameRes.Value.(string)
		remaining := nameRes.Remaining
		log.Printf("[EnhancedAttributeParser] Parsed name: %s", name)

		// Parse Alpine.js directives
		isAlpine := false
		alpineType := ""
		alpineKey := ""

		if strings.HasPrefix(name, "x-") {
			isAlpine = true
			parts := strings.SplitN(name[2:], ":", 2)
			alpineType = parts[0]
			if len(parts) > 1 {
				alpineKey = parts[1]
			}
			log.Printf("[EnhancedAttributeParser] Detected Alpine directive: type=%s, key=%s", alpineType, alpineKey)
		} else if name == "@" || strings.HasPrefix(name, "@") {
			isAlpine = true
			alpineType = "on"
			if name != "@" {
				alpineKey = name[1:] // Extract event name
			}
			log.Printf("[EnhancedAttributeParser] Detected Alpine @ shorthand: %s", name)
		} else if name == ":" || strings.HasPrefix(name, ":") {
			isAlpine = true
			alpineType = "bind"
			if name != ":" {
				alpineKey = name[1:] // Extract binding key
			}
			log.Printf("[EnhancedAttributeParser] Detected Alpine : shorthand: %s", name)
		}

		// Skip whitespace after name
		wsRes := Whitespace()(remaining)
		remaining = wsRes.Remaining

		// Variable for attribute value
		valueStr := ""
		isDynamic := false

		// Parse attribute value if it has one
		if len(remaining) > 0 && remaining[0] == '=' {
			// Special handling for x-data attribute with complex content
			if isAlpine && alpineType == "data" {
				log.Printf("[EnhancedAttributeParser] Special handling for x-data attribute")
				valueRes := parseAlpineDataAttribute(remaining[1:])
				if !valueRes.Successful {
					log.Printf("[EnhancedAttributeParser] Failed to parse x-data attribute: %s", valueRes.Error)
					return Result{nil, input, false, fmt.Sprintf("failed to parse x-data attribute: %s", valueRes.Error), false}
				}

				valueStr = valueRes.Value.(string)
				isDynamic = true
				remaining = valueRes.Remaining
				log.Printf("[EnhancedAttributeParser] Successfully parsed x-data value: %s", valueStr)
			} else {
				// Handle regular attribute value
				log.Printf("[EnhancedAttributeParser] Parsing regular attribute value")
				valueRes := parseAttributeValue(remaining[1:])
				if !valueRes.Successful {
					log.Printf("[EnhancedAttributeParser] Failed to parse attribute value: %s", valueRes.Error)
					return Result{nil, input, false, fmt.Sprintf("failed to parse attribute value: %s", valueRes.Error), false}
				}

				valueStr = valueRes.Value.(string)
				isDynamic = valueRes.Dynamic
				remaining = valueRes.Remaining
				log.Printf("[EnhancedAttributeParser] Successfully parsed attribute value: %s", valueStr)
			}
		} else {
			// Boolean attribute (no value)
			log.Printf("[EnhancedAttributeParser] Boolean attribute (no value)")
			valueStr = ""
		}

		// Create and return the attribute
		attr := ast.Attribute{
			Name:       name,
			Value:      valueStr,
			Dynamic:    isDynamic,
			IsAlpine:   isAlpine,
			AlpineType: alpineType,
			AlpineKey:  alpineKey,
		}

		log.Printf("[EnhancedAttributeParser] Successfully parsed attribute: %s=%s (Alpine=%v)", name, valueStr, isAlpine)
		return Result{attr, remaining, true, "", isDynamic}
	}
}

// Special parser for x-data attribute values which can contain complex JavaScript object literals
func parseAlpineDataAttribute(input string) Result {
	log.Printf("[parseAlpineDataAttribute] Starting on: '%.30s...'", input)

	// Skip whitespace
	wsRes := Whitespace()(input)
	remaining := wsRes.Remaining

	if len(remaining) == 0 {
		return Result{nil, input, false, "empty x-data value", false}
	}

	// Check for opening quote
	if remaining[0] != '"' && remaining[0] != '\'' {
		log.Printf("[parseAlpineDataAttribute] No opening quote found, trying to parse as unquoted value")
		// Try to parse as unquoted value
		i := 0
		for i < len(remaining) && !strings.ContainsRune(" \t\n\r>/", rune(remaining[i])) {
			i++
		}

		if i == 0 {
			return Result{nil, input, false, "empty x-data value", false}
		}

		log.Printf("[parseAlpineDataAttribute] Parsed unquoted value: %s", remaining[:i])
		return Result{remaining[:i], remaining[i:], true, "", true}
	}

	quoteChar := remaining[0]
	remaining = remaining[1:] // Skip opening quote

	var builder strings.Builder
	braceCount := 0
	inSingleQuote := false
	inDoubleQuote := false
	escaped := false

	// Enhanced logging for debugging
	log.Printf("[parseAlpineDataAttribute] Starting to parse quoted value with quote char: %c", quoteChar)

	for i := 0; i < len(remaining); i++ {
		char := remaining[i]

		if escaped {
			// Any escaped character is added literally
			builder.WriteByte(char)
			escaped = false
			continue
		}

		if char == '\\' {
			escaped = true
			builder.WriteByte(char)
			continue
		}

		// Track quotes status
		if char == '\'' && !inDoubleQuote {
			inSingleQuote = !inSingleQuote
			builder.WriteByte(char)
			continue
		}

		if char == '"' && !inSingleQuote {
			inDoubleQuote = !inDoubleQuote
			builder.WriteByte(char)
			continue
		}

		// Track braces only when not in quotes
		if !inSingleQuote && !inDoubleQuote {
			if char == '{' {
				braceCount++
				log.Printf("[parseAlpineDataAttribute] Found opening brace, count: %d", braceCount)
			} else if char == '}' {
				braceCount--
				log.Printf("[parseAlpineDataAttribute] Found closing brace, count: %d", braceCount)
			}
		}

		// Check for closing quote
		if char == quoteChar && !inSingleQuote && !inDoubleQuote && !escaped {
			// Found our closing quote
			log.Printf("[parseAlpineDataAttribute] Found closing quote at position %d", i)
			return Result{builder.String(), remaining[i+1:], true, "", true}
		}

		// Add character to our value
		builder.WriteByte(char)
	}

	// If we got here, we never found the closing quote
	// For Alpine.js, we'll be more lenient and return what we have so far
	log.Printf("[parseAlpineDataAttribute] Warning: Unclosed quote in x-data value, but returning partial content for Alpine.js compatibility")
	return Result{builder.String(), "", true, "", true}
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

// parseAttributeValue handles regular attribute values
func parseAttributeValue(input string) Result {
	log.Printf("[parseAttributeValue] Starting on: '%.30s...'", input)

	// Skip whitespace
	wsRes := Whitespace()(input)
	remaining := wsRes.Remaining

	if len(remaining) == 0 {
		return Result{nil, input, false, "empty attribute value", false}
	}

	// Handle different value formats
	if remaining[0] == '{' {
		// Dynamic value like prop={expr}
		exprRes := Between(String("{"), String("}"), TakeUntil(String("}")))(remaining)
		if !exprRes.Successful {
			return Result{nil, input, false, "unclosed expression in attribute value", true}
		}

		log.Printf("[parseAttributeValue] Parsed expression: %s", exprRes.Value.(string))
		return Result{exprRes.Value, exprRes.Remaining, true, "", true}
	} else if remaining[0] == '"' || remaining[0] == '\'' {
		// Quoted value
		quote := remaining[0]
		var valueRes Result

		if quote == '"' {
			valueRes = DoubleQuotedString()(remaining)
		} else {
			valueRes = SingleQuotedString()(remaining)
		}

		if !valueRes.Successful {
			return Result{nil, input, false, "unclosed quoted value", false}
		}

		// Check if value contains expressions (makes it dynamic)
		isDynamic := strings.Contains(valueRes.Value.(string), "{") && strings.Contains(valueRes.Value.(string), "}")

		log.Printf("[parseAttributeValue] Parsed quoted value: %s (dynamic=%v)", valueRes.Value.(string), isDynamic)
		return Result{valueRes.Value, valueRes.Remaining, true, "", isDynamic}
	} else {
		// Unquoted value
		i := 0
		for i < len(remaining) && !strings.ContainsRune(" \t\n\r>/", rune(remaining[i])) {
			i++
		}

		if i == 0 {
			return Result{nil, input, false, "empty unquoted attribute value", false}
		}

		log.Printf("[parseAttributeValue] Parsed unquoted value: %s", remaining[:i])
		return Result{remaining[:i], remaining[i:], true, "", false}
	}
}

// parseComplexAlpineValue handles Alpine.js values with complex structure
func parseComplexAlpineValue(input string) ValueResult {
	quoteChar := input[0]
	remaining := input[1:] // Skip opening quote

	var valueBuilder strings.Builder
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
