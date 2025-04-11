package parser

import (
	"log"
	"strings"

	"github.com/jimafisk/custom_go_template/ast"
)

// ComponentParser parses component tags (<Component /> or <={expr} />)
func ComponentParser() Parser {
	return func(input string) Result {
		// Log the input for debugging
		log.Printf("[ComponentParser] Starting on: '%.30s...'", input)

		// Trim leading whitespace for better matching
		trimmedInput := strings.TrimLeft(input, " \t\n\r")

		// Check if it's a component tag
		isDynamic := strings.HasPrefix(trimmedInput, "<=")

		// Check for static component tag (starts with < followed by uppercase letter)
		isStatic := false
		if strings.HasPrefix(trimmedInput, "<") && len(trimmedInput) > 1 {
			secondChar := trimmedInput[1]
			if secondChar >= 'A' && secondChar <= 'Z' {
				isStatic = true
				log.Printf("[ComponentParser] Detected static component tag starting with %c", secondChar)
			}
		}

		if !isDynamic && !isStatic {
			log.Printf("[ComponentParser] Not a component tag")
			return Result{nil, input, false, "not a component tag", false}
		}

		// Calculate how much of the original input to skip
		skipChars := len(input) - len(trimmedInput)

		// Get the component content
		contentStart := 1
		if isDynamic {
			contentStart = 2
		}

		// Find the closing tag
		endTag := "/>"
		endTagPos := strings.Index(trimmedInput, endTag)
		if endTagPos == -1 {
			// Try for a non-self-closing component
			closeTagStart := strings.Index(trimmedInput, "</")
			if closeTagStart == -1 {
				log.Printf("[ComponentParser] No closing tag found for component")
				return Result{nil, input, false, "no closing tag for component", false}
			}

			// For now, we'll only support self-closing components
			log.Printf("[ComponentParser] Non-self-closing component found, not supported yet")
			return Result{nil, input, false, "only self-closing components are supported", false}
		}

		// Extract all content between opening < and closing />
		fullContent := trimmedInput[contentStart:endTagPos]
		fullContent = strings.TrimSpace(fullContent)
		log.Printf("[ComponentParser] Component content: '%s'", fullContent)

		// Extract component name and props
		nameEndPos := strings.IndexAny(fullContent, " \t\n\r")
		nameOrPath := fullContent
		props := []ast.ComponentProp{}

		if nameEndPos > 0 {
			nameOrPath = fullContent[:nameEndPos]
			log.Printf("[ComponentParser] Component name: '%s'", nameOrPath)

			// Extract and parse the props
			propContent := strings.TrimSpace(fullContent[nameEndPos:])
			props = parseComponentProps(propContent)
			log.Printf("[ComponentParser] Found %d props", len(props))
		} else {
			log.Printf("[ComponentParser] Component with no props: '%s'", nameOrPath)
		}

		compNode := &ast.ComponentNode{
			Name:    nameOrPath,
			Dynamic: isDynamic,
			Props:   props,
		}

		// Calculate how much of the original input to consume
		consumed := skipChars + endTagPos + len(endTag)

		return Result{compNode, input[consumed:], true, "", false}
	}
}

// parseComponentProps parses component properties from a string
func parseComponentProps(propString string) []ast.ComponentProp {
	props := []ast.ComponentProp{}

	log.Printf("[parseComponentProps] Parsing props: '%s'", propString)
	if strings.TrimSpace(propString) == "" {
		return props
	}

	// This parser handles common React-style prop patterns:
	// - name="value" (static string)
	// - name={expression} (dynamic expression)
	// - name='value' (static string with single quotes)
	// - name={...spread} (spread operator)
	// - {shorthand} (shorthand props)

	remainingProps := propString
	for len(strings.TrimSpace(remainingProps)) > 0 {
		remainingProps = strings.TrimSpace(remainingProps)

		// Check for shorthand prop {prop}
		if strings.HasPrefix(remainingProps, "{") && !strings.HasPrefix(remainingProps, "{") {
			closeBracePos := findMatchingCloseBrace(remainingProps, 0)
			if closeBracePos > 0 {
				// Check if this is a shorthand prop or the start of a prop assignment
				equalsAfterBrace := false
				if len(remainingProps) > closeBracePos+1 {
					restOfProps := strings.TrimSpace(remainingProps[closeBracePos+1:])
					equalsAfterBrace = strings.HasPrefix(restOfProps, "=")
				}

				if !equalsAfterBrace {
					// This is a shorthand prop {prop}
					propName := strings.TrimSpace(remainingProps[1:closeBracePos])
					log.Printf("[parseComponentProps] Shorthand prop: %s", propName)

					props = append(props, ast.ComponentProp{
						Name:        propName,
						Value:       propName,
						IsShorthand: true,
						IsDynamic:   true,
					})

					// Move past this prop
					if len(remainingProps) > closeBracePos+1 {
						remainingProps = remainingProps[closeBracePos+1:]
					} else {
						break
					}
					continue
				}
			}
		}

		// Find the prop name (everything up to = or end)
		propNameEnd := strings.IndexAny(remainingProps, "= \t\n\r")
		if propNameEnd <= 0 {
			// This might be a boolean prop with no value
			propName := remainingProps
			log.Printf("[parseComponentProps] Boolean prop: %s", propName)

			props = append(props, ast.ComponentProp{
				Name:        propName,
				Value:       "true", // Default value for boolean props
				IsShorthand: false,
				IsDynamic:   false,
			})
			break
		}

		propName := strings.TrimSpace(remainingProps[:propNameEnd])
		log.Printf("[parseComponentProps] Found prop name: %s", propName)

		// Move past the prop name
		remainingProps = strings.TrimSpace(remainingProps[propNameEnd:])

		// Check if there's an equals sign
		if !strings.HasPrefix(remainingProps, "=") {
			// This is a boolean prop with no value
			log.Printf("[parseComponentProps] Boolean prop: %s", propName)

			props = append(props, ast.ComponentProp{
				Name:        propName,
				Value:       "true", // Default value for boolean props
				IsShorthand: false,
				IsDynamic:   false,
			})

			continue
		}

		// Move past the equals sign
		remainingProps = strings.TrimSpace(remainingProps[1:])

		// Parse the prop value
		if strings.HasPrefix(remainingProps, "{") && !strings.HasPrefix(remainingProps, "{") {
			// Dynamic prop with expression
			closeBracePos := findMatchingCloseBrace(remainingProps, 0)
			if closeBracePos <= 0 {
				log.Printf("[parseComponentProps] Warning: No matching close brace for prop %s", propName)
				break // Invalid format
			}

			propValue := strings.TrimSpace(remainingProps[1:closeBracePos])
			log.Printf("[parseComponentProps] Dynamic prop value: %s", propValue)

			props = append(props, ast.ComponentProp{
				Name:        propName,
				Value:       propValue,
				IsShorthand: false,
				IsDynamic:   true,
			})

			// Move past this prop
			if len(remainingProps) > closeBracePos+1 {
				remainingProps = remainingProps[closeBracePos+1:]
			} else {
				break
			}
		} else if strings.HasPrefix(remainingProps, "\"") {
			// Static prop with double quotes
			closingQuotePos := findMatchingQuote(remainingProps, 0, '"')
			if closingQuotePos <= 0 {
				log.Printf("[parseComponentProps] Warning: No matching close quote for prop %s", propName)
				break // Invalid format
			}

			propValue := remainingProps[1:closingQuotePos] // Extract without quotes
			log.Printf("[parseComponentProps] Static prop value (double quotes): %s", propValue)

			props = append(props, ast.ComponentProp{
				Name:        propName,
				Value:       propValue,
				IsShorthand: false,
				IsDynamic:   false,
			})

			// Move past this prop
			if len(remainingProps) > closingQuotePos+1 {
				remainingProps = remainingProps[closingQuotePos+1:]
			} else {
				break
			}
		} else if strings.HasPrefix(remainingProps, "'") {
			// Static prop with single quotes
			closingQuotePos := findMatchingQuote(remainingProps, 0, '\'')
			if closingQuotePos <= 0 {
				log.Printf("[parseComponentProps] Warning: No matching close quote for prop %s", propName)
				break // Invalid format
			}

			propValue := remainingProps[1:closingQuotePos] // Extract without quotes
			log.Printf("[parseComponentProps] Static prop value (single quotes): %s", propValue)

			props = append(props, ast.ComponentProp{
				Name:        propName,
				Value:       propValue,
				IsShorthand: false,
				IsDynamic:   false,
			})

			// Move past this prop
			if len(remainingProps) > closingQuotePos+1 {
				remainingProps = remainingProps[closingQuotePos+1:]
			} else {
				break
			}
		} else {
			// Unquoted value (probably a boolean or number)
			spacePos := strings.IndexAny(remainingProps, " \t\n\r")
			if spacePos <= 0 {
				// This is the last prop
				propValue := remainingProps
				log.Printf("[parseComponentProps] Unquoted prop value: %s", propValue)

				props = append(props, ast.ComponentProp{
					Name:        propName,
					Value:       propValue,
					IsShorthand: false,
					IsDynamic:   false,
				})
				break
			}

			propValue := remainingProps[:spacePos]
			log.Printf("[parseComponentProps] Unquoted prop value: %s", propValue)

			props = append(props, ast.ComponentProp{
				Name:        propName,
				Value:       propValue,
				IsShorthand: false,
				IsDynamic:   false,
			})

			remainingProps = remainingProps[spacePos:]
		}
	}

	return props
}

// findMatchingCloseBrace finds the matching closing brace for an opening brace at the specified position
func findMatchingCloseBrace(s string, startPos int) int {
	if len(s) <= startPos || s[startPos] != '{' {
		return -1
	}

	braceLevel := 1
	for i := startPos + 1; i < len(s); i++ {
		if s[i] == '{' {
			braceLevel++
		} else if s[i] == '}' {
			braceLevel--
			if braceLevel == 0 {
				return i
			}
		}
	}

	return -1 // No matching close brace found
}

// findMatchingQuote finds the matching closing quote for an opening quote at the specified position
func findMatchingQuote(s string, startPos int, quoteChar rune) int {
	if len(s) <= startPos || rune(s[startPos]) != quoteChar {
		return -1
	}

	for i := startPos + 1; i < len(s); i++ {
		// Skip escaped quotes
		if s[i] == '\\' && i+1 < len(s) && rune(s[i+1]) == quoteChar {
			i++ // Skip the next character
			continue
		}

		if rune(s[i]) == quoteChar {
			return i
		}
	}

	return -1 // No matching close quote found
}
