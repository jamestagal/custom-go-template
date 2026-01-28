package parser

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/jimafisk/custom_go_template/ast"
)

// DynamicComponentByNameParser parses <Component:dynamic name={expr} {...spread} prop={val} />
// This parser handles Plenti-style dynamic component iteration.
//
// Pattern: Service Implementation Pattern [Load: 18]
// Cognitive Load: 18 (tag matching: 3, name parsing: 5, spread parsing: 4, prop parsing: 4, validation: 2)
//
// Syntax:
//
//	<Component:dynamic name={component.name} {...component.fields} allContent={allContent} />
//
// Key features:
//   - MUST detect colon in tag name: Component:dynamic
//   - MUST require name={} attribute
//   - Handles {...expr} spread props
//   - Handles regular props (static and dynamic)
//   - Handles shorthand props: {varName} → varName={varName}
//   - Skips bind: directives (Svelte-specific, not supported in Alpine)
//   - Preserves prop order for override logic
func DynamicComponentByNameParser() Parser {
	return func(input string) Result {
		log.Printf("[DynamicComponentByNameParser] Starting on: '%.50s...'", input)

		trimmed := strings.TrimSpace(input)

		// STEP 1: Check for <Component:dynamic prefix (COGNITIVE LOAD: 3)
		// The colon is MANDATORY - distinguishes from regular <Component>
		if !strings.HasPrefix(trimmed, "<Component:dynamic") {
			return Result{nil, input, false, "not a Component:dynamic tag", false}
		}

		// Calculate skip offset for whitespace
		skipChars := len(input) - len(trimmed)

		// Move past opening tag to attributes
		afterTag := strings.TrimLeft(trimmed[len("<Component:dynamic"):], " \t\n\r")
		if len(afterTag) == 0 {
			return Result{nil, input, false, "Component:dynamic: no attributes found", false}
		}

		// STEP 2: Find closing /> (we only support self-closing for v1) (COGNITIVE LOAD: 2)
		closeTagPos := strings.Index(afterTag, "/>")
		if closeTagPos == -1 {
			// Check if user tried non-self-closing
			if strings.Contains(afterTag, ">") {
				return Result{nil, input, false, "Component:dynamic must be self-closing (use />)", false}
			}
			return Result{nil, input, false, "Component:dynamic: no closing /> found", false}
		}

		// Extract attribute content between tag name and />
		attrContent := strings.TrimSpace(afterTag[:closeTagPos])
		if len(attrContent) == 0 {
			return Result{nil, input, false, "Component:dynamic: missing name attribute", false}
		}

		log.Printf("[DynamicComponentByNameParser] Attribute content: '%s'", attrContent)

		// STEP 3: Parse name attribute (REQUIRED) (COGNITIVE LOAD: 5)
		nameExpr, remainingAttrs, err := parseNameAttribute(attrContent)
		if err != nil {
			return Result{nil, input, false, fmt.Sprintf("Component:dynamic: %v", err), false}
		}

		log.Printf("[DynamicComponentByNameParser] Parsed name: %s", nameExpr)

		// STEP 4: Parse remaining attributes (spread and regular props) (COGNITIVE LOAD: 8)
		spreadProps, regularProps, err := parseComponentAttributes(remainingAttrs)
		if err != nil {
			return Result{nil, input, false, fmt.Sprintf("Component:dynamic: %v", err), false}
		}

		log.Printf("[DynamicComponentByNameParser] Parsed %d spread props, %d regular props",
			len(spreadProps), len(regularProps))

		// STEP 5: Create DynamicComponentByNameNode
		node := &ast.DynamicComponentByNameNode{
			NameExpression: nameExpr,
			Props:          regularProps,
			SpreadProps:    spreadProps,
			SelfClosing:    true,
		}

		// Calculate consumed length
		consumed := skipChars +
			len("<Component:dynamic") +
			(len(trimmed[len("<Component:dynamic"):]) - len(afterTag)) + // whitespace after tag
			closeTagPos + // attribute content
			2 // />

		log.Printf("[DynamicComponentByNameParser] Success! Consumed %d chars", consumed)
		return Result{node, input[consumed:], true, "", false}
	}
}

// parseNameAttribute parses the required name attribute from attribute string
// Returns: nameExpression, remainingAttributes, error
//
// Supports:
//   - name={expression} - dynamic name
//   - name="literal" - static name
//   - name='literal' - static name with single quotes
func parseNameAttribute(attrContent string) (string, string, error) {
	trimmed := strings.TrimSpace(attrContent)

	// Look for "name=" or "name =" at the start
	namePrefix := "name"
	nameIdx := strings.Index(trimmed, namePrefix)
	if nameIdx != 0 {
		return "", "", fmt.Errorf("name attribute required (must be first attribute)")
	}

	// Move past "name"
	afterName := strings.TrimLeft(trimmed[len(namePrefix):], " \t\n\r")
	if !strings.HasPrefix(afterName, "=") {
		return "", "", fmt.Errorf("name attribute requires value (name=...)")
	}

	// Move past "="
	afterEquals := strings.TrimLeft(afterName[1:], " \t\n\r")
	if len(afterEquals) == 0 {
		return "", "", fmt.Errorf("name attribute value is empty")
	}

	// Parse the value
	var nameExpr string
	var remaining string

	if strings.HasPrefix(afterEquals, "{") {
		// Dynamic expression: name={expr}
		closeBracePos := findMatchingCloseBrace(afterEquals, 0)
		if closeBracePos <= 0 {
			return "", "", fmt.Errorf("name attribute: no matching closing brace")
		}

		nameExpr = strings.TrimSpace(afterEquals[1:closeBracePos])
		remaining = strings.TrimLeft(afterEquals[closeBracePos+1:], " \t\n\r")

	} else if strings.HasPrefix(afterEquals, "\"") {
		// Static string: name="value"
		closeQuotePos := findMatchingQuote(afterEquals, 0, '"')
		if closeQuotePos <= 0 {
			return "", "", fmt.Errorf("name attribute: no matching closing quote")
		}

		nameExpr = afterEquals[1:closeQuotePos]
		remaining = strings.TrimLeft(afterEquals[closeQuotePos+1:], " \t\n\r")

	} else if strings.HasPrefix(afterEquals, "'") {
		// Static string: name='value'
		closeQuotePos := findMatchingQuote(afterEquals, 0, '\'')
		if closeQuotePos <= 0 {
			return "", "", fmt.Errorf("name attribute: no matching closing quote")
		}

		nameExpr = afterEquals[1:closeQuotePos]
		remaining = strings.TrimLeft(afterEquals[closeQuotePos+1:], " \t\n\r")

	} else {
		return "", "", fmt.Errorf("name attribute: value must be quoted or an expression")
	}

	if strings.TrimSpace(nameExpr) == "" {
		return "", "", fmt.Errorf("name attribute: expression is empty")
	}

	return nameExpr, remaining, nil
}

// parseComponentAttributes parses spread props and regular props from attribute string
// Returns: spreadProps, regularProps, error
//
// Handles mixed attributes like:
//
//	{...component.fields} title="Hello" {...overrides} debug={true}
//	{allContent} {allLayouts} - shorthand props (Svelte-style)
//	bind:propName={value} - skipped (Svelte-specific, not supported)
//
// Order is preserved (important for override logic in transformer)
func parseComponentAttributes(attrContent string) ([]string, []ast.ComponentProp, error) {
	spreadProps := []string{}
	regularProps := []ast.ComponentProp{}

	remaining := strings.TrimSpace(attrContent)

	// Compile regex for shorthand prop detection (COGNITIVE LOAD: 3)
	// Pattern: {variableName} where variableName is a valid identifier
	shorthandRegex := regexp.MustCompile(`^\{([a-zA-Z_$][a-zA-Z0-9_$]*)\}`)

	for len(remaining) > 0 {
		// STEP 1: Skip bind: directives (Svelte-specific, not supported) (COGNITIVE LOAD: 2)
		if strings.HasPrefix(remaining, "bind:") {
			log.Printf("[parseComponentAttributes] Skipping bind: directive (Svelte-specific, not supported in Alpine)")

			// Find the end of this attribute (next space or end of string)
			// bind:propName={value} or bind:propName="value"
			endPos := findAttributeEnd(remaining)
			if endPos <= 0 {
				endPos = len(remaining)
			}

			skipped := remaining[:endPos]
			log.Printf("[parseComponentAttributes] Skipped: %s", skipped)

			remaining = strings.TrimLeft(remaining[endPos:], " \t\n\r")
			continue
		}

		// STEP 2: Check for spread prop: {...expr} (COGNITIVE LOAD: 4)
		if strings.HasPrefix(remaining, "{...") {
			spreadExpr, rest, err := parseSpreadProp(remaining)
			if err != nil {
				return nil, nil, fmt.Errorf("spread prop: %w", err)
			}

			spreadProps = append(spreadProps, spreadExpr)
			remaining = strings.TrimLeft(rest, " \t\n\r")

			log.Printf("[parseComponentAttributes] Parsed spread: {...%s}", spreadExpr)
			continue
		}

		// STEP 3: Check for shorthand prop: {varName} (COGNITIVE LOAD: 5)
		if matches := shorthandRegex.FindStringSubmatch(remaining); len(matches) > 1 {
			varName := matches[1]

			// Expand shorthand: {varName} → varName={varName}
			prop := ast.ComponentProp{
				Name:      varName,
				Value:     varName,
				IsDynamic: true, // Shorthand props are always dynamic
			}

			regularProps = append(regularProps, prop)
			remaining = strings.TrimLeft(remaining[len(matches[0]):], " \t\n\r")

			log.Printf("[parseComponentAttributes] Parsed shorthand prop: {%s} → %s=%s", varName, varName, varName)
			continue
		}

		// STEP 4: Parse as regular prop (COGNITIVE LOAD: 8)
		prop, rest, err := parseRegularProp(remaining)
		if err != nil {
			return nil, nil, fmt.Errorf("regular prop: %w", err)
		}

		regularProps = append(regularProps, prop)
		remaining = strings.TrimLeft(rest, " \t\n\r")

		log.Printf("[parseComponentAttributes] Parsed prop: %s=%s", prop.Name, prop.Value)
	}

	return spreadProps, regularProps, nil
}

// findAttributeEnd finds the end of an attribute (for skipping bind: directives)
// Returns the position after the attribute value
func findAttributeEnd(input string) int {
	// Skip "bind:propName"
	equalsPos := strings.Index(input, "=")
	if equalsPos <= 0 {
		// No value, just "bind:propName"
		spacePos := strings.IndexAny(input, " \t\n\r")
		if spacePos < 0 {
			return len(input)
		}
		return spacePos
	}

	// Has value: bind:propName={value} or bind:propName="value"
	afterEquals := strings.TrimLeft(input[equalsPos+1:], " \t\n\r")

	if strings.HasPrefix(afterEquals, "{") {
		// Expression value
		closeBracePos := findMatchingCloseBrace(afterEquals, 0)
		if closeBracePos > 0 {
			return equalsPos + 1 + (len(input[equalsPos+1:]) - len(afterEquals)) + closeBracePos + 1
		}
	} else if strings.HasPrefix(afterEquals, "\"") {
		// Quoted value
		closeQuotePos := findMatchingQuote(afterEquals, 0, '"')
		if closeQuotePos > 0 {
			return equalsPos + 1 + (len(input[equalsPos+1:]) - len(afterEquals)) + closeQuotePos + 1
		}
	} else if strings.HasPrefix(afterEquals, "'") {
		// Single quoted value
		closeQuotePos := findMatchingQuote(afterEquals, 0, '\'')
		if closeQuotePos > 0 {
			return equalsPos + 1 + (len(input[equalsPos+1:]) - len(afterEquals)) + closeQuotePos + 1
		}
	}

	// Couldn't find end, skip to next space
	spacePos := strings.IndexAny(input, " \t\n\r")
	if spacePos < 0 {
		return len(input)
	}
	return spacePos
}

// parseSpreadProp parses {...expression} syntax
// Returns: expression (without braces and dots), remaining input, error
//
// Examples:
//
//	{...component.fields} → "component.fields"
//	{...$store.settings.theme} → "$store.settings.theme"
func parseSpreadProp(input string) (string, string, error) {
	if !strings.HasPrefix(input, "{...") {
		return "", input, fmt.Errorf("not a spread prop")
	}

	// Find matching closing brace
	closeBracePos := findMatchingCloseBrace(input, 0)
	if closeBracePos <= 0 {
		return "", input, fmt.Errorf("no matching closing brace for spread")
	}

	// Extract expression between {...}
	// Skip the "{..." prefix (4 characters: '{', '.', '.', '.')
	spreadExpr := strings.TrimSpace(input[4:closeBracePos])

	if spreadExpr == "" {
		return "", input, fmt.Errorf("spread expression is empty")
	}

	remaining := input[closeBracePos+1:]

	return spreadExpr, remaining, nil
}

// parseRegularProp parses a regular prop: name="value" or name={expr}
// Returns: ComponentProp, remaining input, error
func parseRegularProp(input string) (ast.ComponentProp, string, error) {
	trimmed := strings.TrimSpace(input)

	// Find prop name (everything before =)
	equalsPos := strings.Index(trimmed, "=")
	if equalsPos <= 0 {
		// Check if this is a boolean prop (no value)
		spacePos := strings.IndexAny(trimmed, " \t\n\r")
		if spacePos < 0 {
			// Last prop, boolean
			propName := trimmed
			return ast.ComponentProp{
				Name:      propName,
				Value:     "true",
				IsDynamic: false,
			}, "", nil
		}

		// Boolean prop with more props after
		propName := trimmed[:spacePos]
		return ast.ComponentProp{
			Name:      propName,
			Value:     "true",
			IsDynamic: false,
		}, trimmed[spacePos:], nil
	}

	propName := strings.TrimSpace(trimmed[:equalsPos])
	afterEquals := strings.TrimLeft(trimmed[equalsPos+1:], " \t\n\r")

	if len(afterEquals) == 0 {
		return ast.ComponentProp{}, input, fmt.Errorf("prop %s has no value", propName)
	}

	// Parse value
	var propValue string
	var isDynamic bool
	var remaining string

	if strings.HasPrefix(afterEquals, "{") {
		// Dynamic expression: prop={expr}
		closeBracePos := findMatchingCloseBrace(afterEquals, 0)
		if closeBracePos <= 0 {
			return ast.ComponentProp{}, input, fmt.Errorf("prop %s: no matching closing brace", propName)
		}

		propValue = strings.TrimSpace(afterEquals[1:closeBracePos])
		isDynamic = true
		remaining = afterEquals[closeBracePos+1:]

	} else if strings.HasPrefix(afterEquals, "\"") {
		// Static string: prop="value"
		closeQuotePos := findMatchingQuote(afterEquals, 0, '"')
		if closeQuotePos <= 0 {
			return ast.ComponentProp{}, input, fmt.Errorf("prop %s: no matching closing quote", propName)
		}

		propValue = afterEquals[1:closeQuotePos]
		isDynamic = false
		remaining = afterEquals[closeQuotePos+1:]

	} else if strings.HasPrefix(afterEquals, "'") {
		// Static string: prop='value'
		closeQuotePos := findMatchingQuote(afterEquals, 0, '\'')
		if closeQuotePos <= 0 {
			return ast.ComponentProp{}, input, fmt.Errorf("prop %s: no matching closing quote", propName)
		}

		propValue = afterEquals[1:closeQuotePos]
		isDynamic = false
		remaining = afterEquals[closeQuotePos+1:]

	} else {
		// Unquoted value (boolean, number, etc.)
		spacePos := strings.IndexAny(afterEquals, " \t\n\r")
		if spacePos < 0 {
			// Last value
			propValue = afterEquals
			isDynamic = false
			remaining = ""
		} else {
			propValue = afterEquals[:spacePos]
			isDynamic = false
			remaining = afterEquals[spacePos:]
		}
	}

	return ast.ComponentProp{
		Name:      propName,
		Value:     propValue,
		IsDynamic: isDynamic,
	}, remaining, nil
}
